// Copyright (c) 2025 John Dewey

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	natsclient "github.com/osapi-io/nats-client/pkg/client"

	"github.com/retr0h/osapi/internal/job"
)

// WriteStatusEvent writes an append-only status event for a job.
// This eliminates race conditions by never updating existing keys.
func (c *Client) WriteStatusEvent(
	_ context.Context,
	jobID, event, hostname string,
	data map[string]interface{},
) error {
	// Capture time once for consistency across key and payload
	now := time.Now()

	// Create unique key with nanosecond timestamp to ensure uniqueness
	statusKey := fmt.Sprintf("status.%s.%s.%s.%d",
		jobID, event, sanitizeKeyForNATS(hostname), now.UnixNano())

	// Build event data
	eventData := map[string]interface{}{
		"job_id":    jobID,
		"event":     event,
		"hostname":  hostname,
		"timestamp": now.Format(time.RFC3339Nano),
		"unix_nano": now.UnixNano(),
	}

	// Add any additional data
	if data != nil {
		eventData["data"] = data
	}

	// Marshal event data
	eventJSON, err := json.Marshal(eventData)
	if err != nil {
		return fmt.Errorf("failed to marshal status event: %w", err)
	}

	// Write to KV - this always succeeds due to unique key
	err = c.natsClient.KVPut(c.kv.Bucket(), statusKey, eventJSON)
	if err != nil {
		return fmt.Errorf("failed to write status event: %w", err)
	}

	c.logger.Debug("wrote status event",
		slog.String("job_id", jobID),
		slog.String("event", event),
		slog.String("hostname", hostname),
		slog.String("key", statusKey),
	)

	return nil
}

// WriteJobResponse stores agent response data for a job.
func (c *Client) WriteJobResponse(
	_ context.Context,
	jobID, hostname string,
	responseData []byte,
	status string,
	errorMsg string,
	changed *bool,
) error {
	// Create response key with timestamp to ensure uniqueness
	responseKey := fmt.Sprintf("responses.%s.%s.%d",
		jobID, sanitizeKeyForNATS(hostname), time.Now().UnixNano())

	// Build response structure
	response := job.Response{
		Status:    job.Status(status),
		Data:      responseData,
		Error:     errorMsg,
		Changed:   changed,
		Hostname:  hostname,
		Timestamp: time.Now(),
	}

	// Marshal response (Response fields are always marshalable)
	responseJSON, _ := json.Marshal(response)

	// Store in KV
	err := c.natsClient.KVPut(c.kv.Bucket(), responseKey, responseJSON)
	if err != nil {
		return fmt.Errorf("failed to store job response: %w", err)
	}

	c.logger.Debug("wrote job response",
		slog.String("job_id", jobID),
		slog.String("hostname", hostname),
		slog.String("status", status),
		slog.String("key", responseKey),
	)

	return nil
}

// ConsumeJobs sets up message consumption for job processing.
func (c *Client) ConsumeJobs(
	ctx context.Context,
	streamName, consumerName string,
	handler func(jetstream.Msg) error,
	opts *natsclient.ConsumeOptions,
) error {
	// Use the natsClient to consume messages with the provided consumer name
	return c.natsClient.ConsumeMessages(ctx, streamName, consumerName,
		func(msg jetstream.Msg) error {
			return handler(msg)
		}, opts)
}

// GetJobData retrieves raw job data from the KV store.
func (c *Client) GetJobData(
	ctx context.Context,
	jobKey string,
) ([]byte, error) {
	c.logger.Debug("kv.get",
		slog.String("key", jobKey),
	)
	entry, err := c.kv.Get(ctx, jobKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get job data for key %s: %w", jobKey, err)
	}
	return entry.Value(), nil
}

// CreateOrUpdateConsumer creates or updates a JetStream consumer.
func (c *Client) CreateOrUpdateConsumer(
	ctx context.Context,
	streamName string,
	consumerConfig jetstream.ConsumerConfig,
) error {
	return c.natsClient.CreateOrUpdateConsumerWithConfig(ctx, streamName, consumerConfig)
}

// sanitizeKeyForNATS sanitizes a string for use as a NATS key.
func sanitizeKeyForNATS(
	input string,
) string {
	// Replace invalid characters with underscores
	reg := regexp.MustCompile(`[^a-zA-Z0-9_\-]`)
	return reg.ReplaceAllString(input, "_")
}

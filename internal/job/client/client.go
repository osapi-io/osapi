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

// Package client provides job client operations for NATS JetStream.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/messaging"
)

// Client provides methods for publishing job requests and retrieving responses.
type Client struct {
	logger     *slog.Logger
	natsClient messaging.NATSClient
	kv         jetstream.KeyValue
	registryKV jetstream.KeyValue
	factsKV    jetstream.KeyValue
	stateKV    jetstream.KeyValue
	timeout    time.Duration
	streamName string
}

// Options configures the jobs client.
type Options struct {
	// Timeout for waiting for job responses (default: 30s)
	Timeout time.Duration
	// KVBucket for job storage (required)
	KVBucket jetstream.KeyValue
	// RegistryKV is the KV bucket for agent registry (optional).
	RegistryKV jetstream.KeyValue
	// FactsKV is the KV bucket for agent facts (optional).
	FactsKV jetstream.KeyValue
	// StateKV is the KV bucket for persistent agent state (drain flags, timeline).
	StateKV jetstream.KeyValue
	// StreamName is the JetStream stream name (used to derive DLQ name).
	StreamName string
}

// New creates a new jobs client using an existing NATS client.
func New(
	logger *slog.Logger,
	natsClient messaging.NATSClient,
	opts *Options,
) (*Client, error) {
	if opts == nil {
		return nil, fmt.Errorf("options cannot be nil")
	}
	if opts.KVBucket == nil {
		return nil, fmt.Errorf("kvBucket cannot be nil")
	}

	return &Client{
		logger:     logger,
		natsClient: natsClient,
		kv:         opts.KVBucket,
		registryKV: opts.RegistryKV,
		factsKV:    opts.FactsKV,
		stateKV:    opts.StateKV,
		streamName: opts.StreamName,
		timeout:    opts.Timeout,
	}, nil
}

// publishAndWait stores a job in KV, publishes a notification, and waits for the agent response.
// Returns the job ID, response, and any error.
func (c *Client) publishAndWait(
	ctx context.Context,
	subject string,
	req *job.Request,
) (string, *job.Response, error) {
	// Generate job ID if not provided
	if req.JobID == "" {
		req.JobID = uuid.New().String()
	}
	req.Timestamp = time.Now()

	jobID := req.JobID
	createdTime := req.Timestamp.Format(time.RFC3339)

	// Build operation type from category and operation
	operationType := req.Category + "." + req.Operation
	operationData := map[string]interface{}{
		"type": operationType,
		"data": req.Data,
	}

	// Store immutable job data in KV (same structure as CreateJob)
	jobData := map[string]interface{}{
		"id":        jobID,
		"status":    "unprocessed",
		"created":   createdTime,
		"subject":   subject,
		"operation": operationData,
	}

	jobJSON, _ := json.Marshal(jobData)
	kvKey := "jobs." + jobID

	c.logger.DebugContext(ctx, "kv.put",
		slog.String("key", kvKey),
		slog.String("job_id", jobID),
	)

	if _, err := c.kv.Put(ctx, kvKey, jobJSON); err != nil {
		return "", nil, fmt.Errorf("failed to store job in KV: %w", err)
	}

	c.logger.InfoContext(ctx, "publishing job request",
		slog.String("job_id", jobID),
		slog.String("subject", subject),
		slog.String("type", string(req.Type)),
	)

	// Publish just the job ID to the stream as a notification.
	// Trace context is propagated via NATS message headers automatically by nats-client.
	if err := c.natsClient.Publish(ctx, subject, []byte(jobID)); err != nil {
		return "", nil, fmt.Errorf("failed to publish notification: %w", err)
	}

	// Watch for agent response in KV
	responsePattern := "responses." + jobID + ".>"
	watcher, err := c.kv.Watch(ctx, responsePattern)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create response watcher: %w", err)
	}
	defer func() {
		_ = watcher.Stop()
	}()

	timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			return "", nil, fmt.Errorf("timeout waiting for job response: %w", timeoutCtx.Err())
		case entry := <-watcher.Updates():
			if entry == nil {
				continue
			}

			var response job.Response
			if err := json.Unmarshal(entry.Value(), &response); err != nil {
				return "", nil, fmt.Errorf("failed to unmarshal response: %w", err)
			}

			c.logger.InfoContext(ctx, "received job response",
				slog.String("job_id", jobID),
				slog.String("status", string(response.Status)),
			)

			return jobID, &response, nil
		}
	}
}

// publishAndCollect stores a job in KV, publishes a notification, and collects
// agent responses. It uses the agent registry to determine how many agents are
// expected to respond and returns immediately once all have replied. The overall
// timeout acts as a safety net if an agent doesn't respond.
// Returns the job ID, collected responses keyed by hostname, and any error.
func (c *Client) publishAndCollect(
	ctx context.Context,
	subject string,
	target string,
	req *job.Request,
) (string, map[string]*job.Response, error) {
	// Generate job ID if not provided
	if req.JobID == "" {
		req.JobID = uuid.New().String()
	}
	req.Timestamp = time.Now()

	jobID := req.JobID
	createdTime := req.Timestamp.Format(time.RFC3339)

	// Build operation type from category and operation
	operationType := req.Category + "." + req.Operation
	operationData := map[string]interface{}{
		"type": operationType,
		"data": req.Data,
	}

	// Store immutable job data in KV
	jobData := map[string]interface{}{
		"id":        jobID,
		"status":    "unprocessed",
		"created":   createdTime,
		"subject":   subject,
		"operation": operationData,
	}

	jobJSON, _ := json.Marshal(jobData)
	kvKey := "jobs." + jobID

	c.logger.DebugContext(ctx, "kv.put",
		slog.String("key", kvKey),
		slog.String("job_id", jobID),
	)

	if _, err := c.kv.Put(ctx, kvKey, jobJSON); err != nil {
		return "", nil, fmt.Errorf("failed to store job in KV: %w", err)
	}

	// Determine expected agent count from registry for early completion.
	// If ListAgents fails, fall back to waiting for the full timeout.
	expectedCount := 0
	agents, err := c.ListAgents(ctx)
	if err != nil {
		c.logger.WarnContext(ctx, "failed to list agents for broadcast count, using full timeout",
			slog.String("error", err.Error()),
		)
	} else {
		expectedCount = job.CountExpectedAgents(agents, target)
		c.logger.DebugContext(ctx, "broadcast expected agent count",
			slog.String("target", target),
			slog.Int("expected_count", expectedCount),
		)
	}

	c.logger.InfoContext(ctx, "publishing broadcast job request",
		slog.String("job_id", jobID),
		slog.String("subject", subject),
		slog.String("type", string(req.Type)),
	)

	// Publish just the job ID to the stream as a notification.
	// Trace context is propagated via NATS message headers automatically by nats-client.
	if err := c.natsClient.Publish(ctx, subject, []byte(jobID)); err != nil {
		return "", nil, fmt.Errorf("failed to publish notification: %w", err)
	}

	// Watch for agent responses in KV
	responsePattern := "responses." + jobID + ".>"
	watcher, err := c.kv.Watch(ctx, responsePattern)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create response watcher: %w", err)
	}
	defer func() {
		_ = watcher.Stop()
	}()

	responses := make(map[string]*job.Response)

	// Overall timeout as safety net
	timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			if len(responses) == 0 {
				return "", nil, fmt.Errorf(
					"timeout waiting for broadcast responses: no agents responded",
				)
			}
			return jobID, responses, nil
		case entry := <-watcher.Updates():
			if entry == nil {
				continue
			}

			var response job.Response
			if err := json.Unmarshal(entry.Value(), &response); err != nil {
				c.logger.WarnContext(ctx, "failed to unmarshal broadcast response",
					slog.String("job_id", jobID),
					slog.String("error", err.Error()),
				)
				continue
			}

			hostname := response.Hostname
			if hostname == "" {
				hostname = "unknown"
			}

			c.logger.InfoContext(ctx, "received broadcast response",
				slog.String("job_id", jobID),
				slog.String("hostname", hostname),
				slog.String("status", string(response.Status)),
			)

			responses[hostname] = &response

			// Return early when all expected agents have responded.
			if expectedCount > 0 && len(responses) >= expectedCount {
				return jobID, responses, nil
			}
		}
	}
}

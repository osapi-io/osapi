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
	"github.com/nats-io/nats.go"

	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/messaging"
)

// Client provides methods for publishing job requests and retrieving responses.
type Client struct {
	logger             *slog.Logger
	natsClient         messaging.NATSClient
	kv                 nats.KeyValue
	timeout            time.Duration
	broadcastQuietTime time.Duration
}

// Options configures the jobs client.
type Options struct {
	// Timeout for waiting for job responses (default: 30s)
	Timeout time.Duration
	// BroadcastQuietPeriod is the silence window after the last broadcast
	// response before collection stops (default: 3s)
	BroadcastQuietPeriod time.Duration
	// KVBucket for job storage (required)
	KVBucket nats.KeyValue
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
		return nil, fmt.Errorf("KVBucket cannot be nil")
	}

	quietPeriod := opts.BroadcastQuietPeriod
	if quietPeriod == 0 {
		quietPeriod = broadcastQuietPeriod
	}

	return &Client{
		logger:             logger,
		natsClient:         natsClient,
		kv:                 opts.KVBucket,
		timeout:            opts.Timeout,
		broadcastQuietTime: quietPeriod,
	}, nil
}

// publishAndWait stores a job in KV, publishes a notification, and waits for the worker response.
func (c *Client) publishAndWait(
	ctx context.Context,
	subject string,
	req *job.Request,
) (*job.Response, error) {
	// Generate request ID if not provided
	if req.RequestID == "" {
		req.RequestID = uuid.New().String()
	}
	req.Timestamp = time.Now()

	jobID := req.RequestID
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

	c.logger.Debug("kv.put",
		slog.String("key", kvKey),
		slog.String("request_id", jobID),
	)

	if _, err := c.kv.Put(kvKey, jobJSON); err != nil {
		return nil, fmt.Errorf("failed to store job in KV: %w", err)
	}

	c.logger.Info("publishing job request",
		slog.String("request_id", jobID),
		slog.String("subject", subject),
		slog.String("type", string(req.Type)),
	)

	// Publish just the job ID to the stream as a notification
	if err := c.natsClient.Publish(ctx, subject, []byte(jobID)); err != nil {
		return nil, fmt.Errorf("failed to publish notification: %w", err)
	}

	// Watch for worker response in KV
	responsePattern := "responses." + jobID + ".>"
	watcher, err := c.kv.Watch(responsePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to create response watcher: %w", err)
	}
	defer func() {
		_ = watcher.Stop()
	}()

	timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			return nil, fmt.Errorf("timeout waiting for job response: %w", timeoutCtx.Err())
		case entry := <-watcher.Updates():
			if entry == nil {
				continue
			}

			var response job.Response
			if err := json.Unmarshal(entry.Value(), &response); err != nil {
				return nil, fmt.Errorf("failed to unmarshal response: %w", err)
			}

			c.logger.Info("received job response",
				slog.String("request_id", jobID),
				slog.String("status", string(response.Status)),
			)

			return &response, nil
		}
	}
}

// broadcastQuietPeriod is the duration of silence after the last response
// before we consider the broadcast complete. If no new responses arrive
// within this window, we return whatever we've collected.
const broadcastQuietPeriod = 3 * time.Second

// publishAndCollect stores a job in KV, publishes a notification, and collects
// worker responses using a quiet period strategy. After each response, a short
// timer resets. When the timer expires with no new responses, the collected
// results are returned. The overall timeout acts as a safety net.
func (c *Client) publishAndCollect(
	ctx context.Context,
	subject string,
	req *job.Request,
) (map[string]*job.Response, error) {
	// Generate request ID if not provided
	if req.RequestID == "" {
		req.RequestID = uuid.New().String()
	}
	req.Timestamp = time.Now()

	jobID := req.RequestID
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

	c.logger.Debug("kv.put",
		slog.String("key", kvKey),
		slog.String("request_id", jobID),
	)

	if _, err := c.kv.Put(kvKey, jobJSON); err != nil {
		return nil, fmt.Errorf("failed to store job in KV: %w", err)
	}

	c.logger.Info("publishing broadcast job request",
		slog.String("request_id", jobID),
		slog.String("subject", subject),
		slog.String("type", string(req.Type)),
	)

	// Publish just the job ID to the stream as a notification
	if err := c.natsClient.Publish(ctx, subject, []byte(jobID)); err != nil {
		return nil, fmt.Errorf("failed to publish notification: %w", err)
	}

	// Watch for worker responses in KV
	responsePattern := "responses." + jobID + ".>"
	watcher, err := c.kv.Watch(responsePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to create response watcher: %w", err)
	}
	defer func() {
		_ = watcher.Stop()
	}()

	responses := make(map[string]*job.Response)

	// Overall timeout as safety net
	timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Quiet period timer — starts at the full timeout (waiting for first response),
	// then resets to the short quiet period after each response arrives.
	quietTimer := time.NewTimer(c.timeout)
	defer quietTimer.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
		case <-quietTimer.C:
		case entry := <-watcher.Updates():
			if entry == nil {
				continue
			}

			var response job.Response
			if err := json.Unmarshal(entry.Value(), &response); err != nil {
				c.logger.Warn("failed to unmarshal broadcast response",
					slog.String("request_id", jobID),
					slog.String("error", err.Error()),
				)
				continue
			}

			hostname := response.Hostname
			if hostname == "" {
				hostname = "unknown"
			}

			c.logger.Info("received broadcast response",
				slog.String("request_id", jobID),
				slog.String("hostname", hostname),
				slog.String("status", string(response.Status)),
			)

			responses[hostname] = &response

			// Reset quiet period — if no more responses arrive within
			// this window, we're done collecting.
			quietTimer.Reset(c.broadcastQuietTime)
			continue
		}

		// Reached from either timeoutCtx.Done() or quietTimer.C
		if len(responses) == 0 {
			return nil, fmt.Errorf(
				"timeout waiting for broadcast responses: no workers responded",
			)
		}
		return responses, nil
	}
}

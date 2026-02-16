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
	natsclient "github.com/osapi-io/nats-client/pkg/client"

	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/messaging"
)

// Client provides methods for publishing job requests and retrieving responses.
type Client struct {
	logger     *slog.Logger
	natsClient messaging.NATSClient
	kv         nats.KeyValue
	timeout    time.Duration
}

// Options configures the jobs client.
type Options struct {
	// Timeout for waiting for job responses (default: 30s)
	Timeout time.Duration
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

	return &Client{
		logger:     logger,
		natsClient: natsClient,
		kv:         opts.KVBucket,
		timeout:    opts.Timeout,
	}, nil
}

// publishAndWait publishes a job request and waits for the response.
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

	// Marshal request (Request fields are always marshalable)
	data, _ := json.Marshal(req)

	c.logger.Info("publishing job request",
		slog.String("request_id", req.RequestID),
		slog.String("subject", subject),
		slog.String("type", string(req.Type)),
	)

	// Use nats-client's PublishAndWaitKV
	opts := &natsclient.RequestReplyOptions{
		RequestID: req.RequestID,
		Timeout:   c.timeout,
	}

	responseData, err := c.natsClient.PublishAndWaitKV(ctx, subject, data, c.kv, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to publish and wait: %w", err)
	}

	// Unmarshal response
	var response job.Response
	if err := json.Unmarshal(responseData, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	c.logger.Info("received job response",
		slog.String("request_id", req.RequestID),
		slog.String("status", string(response.Status)),
	)

	return &response, nil
}

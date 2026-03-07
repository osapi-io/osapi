// Copyright (c) 2026 John Dewey

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

// Package osapi provides a Go SDK for the OSAPI REST API.
//
// Create a client with New() and use the domain-specific services
// to interact with the API:
//
//	client := osapi.New("http://localhost:8080", "your-jwt-token")
//
//	// Get hostname
//	resp, err := client.Node.Hostname(ctx, "_any")
//
//	// Execute a command
//	resp, err := client.Node.Exec(ctx, osapi.ExecRequest{
//	    Command: "uptime",
//	    Target:  "_all",
//	})
package osapi

import (
	"log/slog"
	"net/http"

	"github.com/retr0h/osapi/pkg/sdk/osapi/gen"
)

// Client is the top-level OSAPI SDK client. Use New() to create one.
type Client struct {
	// Agent provides agent discovery and details operations.
	Agent *AgentService

	// Node provides node management operations (hostname, status, disk,
	// memory, load, OS, uptime, network DNS/ping, command exec/shell).
	Node *NodeService

	// Job provides job queue operations (create, get, list, delete, retry).
	Job *JobService

	// Health provides health check operations (liveness, readiness, status).
	Health *HealthService

	// Audit provides audit log operations (list, get, export).
	Audit *AuditService

	// Metrics provides Prometheus metrics access.
	Metrics *MetricsService

	// File provides file management operations (upload, list, get, delete).
	File *FileService

	httpClient    *gen.ClientWithResponses
	baseURL       string
	logger        *slog.Logger
	baseTransport http.RoundTripper
}

// Option configures the Client.
type Option func(*Client)

// WithLogger sets a custom logger. Defaults to slog.Default().
func WithLogger(
	logger *slog.Logger,
) Option {
	return func(c *Client) {
		c.logger = logger
	}
}

// WithHTTPTransport sets a custom base HTTP transport.
func WithHTTPTransport(
	transport http.RoundTripper,
) Option {
	return func(c *Client) {
		c.baseTransport = transport
	}
}

// New creates an OSAPI SDK client.
func New(
	baseURL string,
	bearerToken string,
	opts ...Option,
) *Client {
	c := &Client{
		baseURL:       baseURL,
		logger:        slog.Default(),
		baseTransport: http.DefaultTransport,
	}

	for _, opt := range opts {
		opt(c)
	}

	transport := &authTransport{
		base:       c.baseTransport,
		authHeader: "Bearer " + bearerToken,
		logger:     c.logger,
	}

	hc := &http.Client{
		Transport: transport,
	}

	// Error is unreachable: the only ClientOption passed (WithHTTPClient) cannot
	// fail, and NewClientWithResponses only errors when a ClientOption does.
	// Invalid URLs are caught later at HTTP call time with a clear parse error.
	httpClient, _ := gen.NewClientWithResponses(baseURL, gen.WithHTTPClient(hc))

	c.httpClient = httpClient
	c.Agent = &AgentService{client: httpClient}
	c.Node = &NodeService{client: httpClient}
	c.Job = &JobService{client: httpClient}
	c.Health = &HealthService{client: httpClient}
	c.Audit = &AuditService{client: httpClient}
	c.Metrics = &MetricsService{
		client:  httpClient,
		baseURL: baseURL,
	}
	c.File = &FileService{client: httpClient}

	return c
}

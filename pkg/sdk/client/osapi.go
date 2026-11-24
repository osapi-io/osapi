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

// Package client is the Go SDK of the OSAPI REST API.
//
// Create a client with New(), then call the typed services, one for each
// group of the API. internal/sdk/gen writes the services and their types
// from the API's spec:
//
//	c := client.New("http://localhost:8080", "your-jwt-token")
//
//	// The hostname of each host.
//	resp, err := c.Hostname.Get(ctx, "_all")
//
//	// Run a command on one host.
//	resp, err := c.Command.Exec(ctx, "web-01", client.CommandExecRequest{
//	    Command: "uptime",
//	})
package client

import (
	"log/slog"
	"net/http"
)

// Client is the top-level OSAPI SDK client. Use New() to create one. Each
// typed service is a field of the embedded services, such as c.DNS.
type Client struct {
	services

	hc            *http.Client
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

// WithHTTPTransport sets a custom base HTTP transport. Without it, the
// client sends through its own clone of http.DefaultTransport, so a caller
// that closes the idle connections of http.DefaultTransport (for example
// httptest.Server.Close) does not break a request of this client.
func WithHTTPTransport(
	transport http.RoundTripper,
) Option {
	return func(c *Client) {
		c.baseTransport = transport
	}
}

// cloneTransport returns a private copy of base when it is an
// *http.Transport. Any other RoundTripper is returned as is, because it
// cannot be cloned.
func cloneTransport(
	base http.RoundTripper,
) http.RoundTripper {
	if t, ok := base.(*http.Transport); ok {
		return t.Clone()
	}

	return base
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
		baseTransport: cloneTransport(http.DefaultTransport),
	}

	for _, opt := range opts {
		opt(c)
	}

	transport := &authTransport{
		base:       c.baseTransport,
		authHeader: "Bearer " + bearerToken,
		logger:     c.logger,
	}

	c.hc = &http.Client{
		Transport: transport,
	}
	c.wire(c)

	return c
}

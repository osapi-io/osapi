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
package client_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/pkg/sdk/client"
)

type ServicePublicTestSuite struct {
	suite.Suite

	ctx context.Context
}

func (suite *ServicePublicTestSuite) SetupTest() {
	suite.ctx = context.Background()
}

func (suite *ServicePublicTestSuite) calls() map[string]call {
	return map[string]call{
		"list": func(c *client.Client) (any, error) {
			return c.Service.List(suite.ctx, "_any")
		},
		"get": func(c *client.Client) (any, error) {
			return c.Service.Get(suite.ctx, "_any", "ssh")
		},
		"create": func(c *client.Client) (any, error) {
			return c.Service.Create(suite.ctx, "_any",
				client.ServiceCreateOpts{Name: "app", Object: "app.service"})
		},
		"update": func(c *client.Client) (any, error) {
			return c.Service.Update(suite.ctx, "_any", "app",
				client.ServiceUpdateOpts{Object: "app.service"})
		},
		"delete": func(c *client.Client) (any, error) {
			return c.Service.Delete(suite.ctx, "_any", "app")
		},
		"start": func(c *client.Client) (any, error) {
			return c.Service.Start(suite.ctx, "_any", "app")
		},
		"stop": func(c *client.Client) (any, error) {
			return c.Service.Stop(suite.ctx, "_any", "app")
		},
		"restart": func(c *client.Client) (any, error) {
			return c.Service.Restart(suite.ctx, "_any", "app")
		},
		"enable": func(c *client.Client) (any, error) {
			return c.Service.Enable(suite.ctx, "_any", "app")
		},
		"disable": func(c *client.Client) (any, error) {
			return c.Service.Disable(suite.ctx, "_any", "app")
		},
	}
}

const serviceRow = `{"job_id":"00000000-0000-0000-0000-000000000001","results":[{"hostname":"web-01",` +
	`"status":"ok","name":"ssh","state":"running","enabled":true,` +
	`"description":"OpenBSD Secure Shell server","changed":true}]}`

func (suite *ServicePublicTestSuite) TestService() {
	tests := []struct {
		name      string
		handler   http.HandlerFunc
		serverURL string
		wantErr   string
	}{
		{
			name: "when the API answers",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(serviceRow))
			},
		},
		{
			name: "when the API refuses",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
			wantErr: "forbidden",
		},
		{
			name: "when the body is empty",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			wantErr: "nil response body",
		},
		{
			name:      "when the server is unreachable",
			serverURL: "http://127.0.0.1:0",
			wantErr:   "service",
		},
	}
	for _, tc := range tests {
		for method, fn := range suite.calls() {
			suite.Run(tc.name+": "+method, func() {
				url := tc.serverURL
				if url == "" {
					server := httptest.NewServer(tc.handler)
					defer server.Close()
					url = server.URL
				}
				c := client.New(url, "test-token", client.WithLogger(slog.Default()))

				resp, err := fn(c)
				if tc.wantErr != "" {
					suite.ErrorContains(err, tc.wantErr)

					return
				}
				suite.Require().NoError(err)
				suite.NotNil(resp)
			})
		}
	}
}

func (suite *ServicePublicTestSuite) TestFields() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(serviceRow))
	}))
	defer server.Close()
	c := client.New(server.URL, "test-token", client.WithLogger(slog.Default()))

	list, err := c.Service.List(suite.ctx, "_any")
	suite.Require().NoError(err)
	suite.Equal(client.ServiceInfo{
		Hostname: "web-01", Status: "ok", Name: "ssh", State: "running",
		Enabled: true, Description: "OpenBSD Secure Shell server",
	}, list.Data.Results[0])
	suite.Equal("00000000-0000-0000-0000-000000000001", list.Data.JobID)

	started, err := c.Service.Start(suite.ctx, "_any", "app")
	suite.Require().NoError(err)
	suite.Equal(client.ServiceMutationResult{
		Hostname: "web-01", Status: "ok", Name: "ssh", Changed: true,
	}, started.Data.Results[0])
}

func TestServicePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ServicePublicTestSuite))
}

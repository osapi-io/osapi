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

	"github.com/retr0h/osapi/internal/client"
	"github.com/retr0h/osapi/internal/client/gen"
	"github.com/retr0h/osapi/internal/config"
)

type ClientPublicTestSuite struct {
	suite.Suite

	server    *httptest.Server
	appConfig config.Config
	genClient *gen.ClientWithResponses
	sut       *client.Client
}

func (s *ClientPublicTestSuite) SetupTest() {
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))

	s.appConfig = config.Config{
		API: config.API{
			Client: config.Client{
				URL: s.server.URL,
				Security: config.ClientSecurity{
					BearerToken: "test-token",
				},
			},
		},
	}

	var err error
	s.genClient, err = client.NewClientWithResponses(slog.Default(), s.appConfig)
	s.Require().NoError(err)

	s.sut = client.New(slog.Default(), s.appConfig, s.genClient)
}

func (s *ClientPublicTestSuite) TearDownTest() {
	s.server.Close()
}

func (s *ClientPublicTestSuite) TestNew() {
	tests := []struct {
		name string
	}{
		{
			name: "creates client with config and logger",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			c := client.New(slog.Default(), config.Config{}, nil)

			s.NotNil(c)
		})
	}
}

func (s *ClientPublicTestSuite) TestNewClientWithResponses() {
	tests := []struct {
		name        string
		config      config.Config
		expectError bool
	}{
		{
			name: "creates client with valid config URL",
			config: config.Config{
				API: config.API{
					Client: config.Client{
						URL: "http://localhost:8080",
						Security: config.ClientSecurity{
							BearerToken: "test-token",
						},
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			c, err := client.NewClientWithResponses(slog.Default(), tt.config)

			if tt.expectError {
				s.Error(err)
				s.Nil(c)
			} else {
				s.NoError(err)
				s.NotNil(c)
			}
		})
	}
}

func (s *ClientPublicTestSuite) TestRoundTrip() {
	tests := []struct {
		name           string
		bearerToken    string
		serverURL      string
		useRealServer  bool
		expectedHeader string
		expectError    bool
	}{
		{
			name:           "injects authorization header",
			bearerToken:    "my-secret-token",
			useRealServer:  true,
			expectedHeader: "Bearer my-secret-token",
		},
		{
			name:        "logs error when request fails",
			bearerToken: "my-secret-token",
			serverURL:   "http://127.0.0.1:0",
			expectError: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var receivedAuth string
			serverURL := tt.serverURL

			if tt.useRealServer {
				server := httptest.NewServer(
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						receivedAuth = r.Header.Get("Authorization")
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
						_, _ = w.Write([]byte(`{}`))
					}),
				)
				defer server.Close()
				serverURL = server.URL
			}

			appConfig := config.Config{
				API: config.API{
					Client: config.Client{
						URL: serverURL,
						Security: config.ClientSecurity{
							BearerToken: tt.bearerToken,
						},
					},
				},
			}

			genClient, err := client.NewClientWithResponses(slog.Default(), appConfig)
			s.Require().NoError(err)

			c := client.New(slog.Default(), appConfig, genClient)
			s.NotNil(c)

			_, err = genClient.GetSystemHostnameWithResponse(
				context.Background(),
				&gen.GetSystemHostnameParams{},
			)

			if tt.expectError {
				s.Error(err)
			} else {
				s.NoError(err)
				s.Equal(tt.expectedHeader, receivedAuth)
			}
		})
	}
}

func (s *ClientPublicTestSuite) TestGetSystemHostname() {
	tests := []struct {
		name string
	}{
		{
			name: "returns hostname response",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.Background()

			resp, err := s.sut.GetSystemHostname(ctx, "_any")

			s.NoError(err)
			s.NotNil(resp)
		})
	}
}

func (s *ClientPublicTestSuite) TestGetSystemStatus() {
	tests := []struct {
		name string
	}{
		{
			name: "returns system status response",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.Background()

			resp, err := s.sut.GetSystemStatus(ctx, "_any")

			s.NoError(err)
			s.NotNil(resp)
		})
	}
}

func (s *ClientPublicTestSuite) TestGetNetworkDNSByInterface() {
	tests := []struct {
		name          string
		interfaceName string
	}{
		{
			name:          "returns DNS config for interface",
			interfaceName: "eth0",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.Background()

			resp, err := s.sut.GetNetworkDNSByInterface(ctx, "_any", tt.interfaceName)

			s.NoError(err)
			s.NotNil(resp)
		})
	}
}

func (s *ClientPublicTestSuite) TestPutNetworkDNS() {
	tests := []struct {
		name          string
		servers       []string
		searchDomains []string
		interfaceName string
	}{
		{
			name:          "with servers and search domains",
			servers:       []string{"8.8.8.8"},
			searchDomains: []string{"example.com"},
			interfaceName: "eth0",
		},
		{
			name:          "with empty servers and domains",
			servers:       nil,
			searchDomains: nil,
			interfaceName: "eth0",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.Background()

			resp, err := s.sut.PutNetworkDNS(
				ctx,
				"_any",
				tt.servers,
				tt.searchDomains,
				tt.interfaceName,
			)

			s.NoError(err)
			s.NotNil(resp)
		})
	}
}

func (s *ClientPublicTestSuite) TestPostNetworkPing() {
	tests := []struct {
		name   string
		target string
	}{
		{
			name:   "returns ping response",
			target: "google.com",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.Background()

			resp, err := s.sut.PostNetworkPing(ctx, "_any", tt.target)

			s.NoError(err)
			s.NotNil(resp)
		})
	}
}

func (s *ClientPublicTestSuite) TestPostJob() {
	tests := []struct {
		name           string
		operation      map[string]interface{}
		targetHostname string
	}{
		{
			name:           "creates job with operation and target",
			operation:      map[string]interface{}{"type": "system.hostname.get"},
			targetHostname: "_any",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.Background()

			resp, err := s.sut.PostJob(ctx, tt.operation, tt.targetHostname)

			s.NoError(err)
			s.NotNil(resp)
		})
	}
}

func (s *ClientPublicTestSuite) TestGetJobByID() {
	tests := []struct {
		name    string
		id      string
		wantErr string
	}{
		{
			name: "returns job detail response",
			id:   "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:    "returns error for invalid uuid",
			id:      "not-a-uuid",
			wantErr: "invalid job ID",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.Background()

			resp, err := s.sut.GetJobByID(ctx, tt.id)

			if tt.wantErr != "" {
				s.ErrorContains(err, tt.wantErr)
				s.Nil(resp)
			} else {
				s.NoError(err)
				s.NotNil(resp)
			}
		})
	}
}

func (s *ClientPublicTestSuite) TestDeleteJobByID() {
	tests := []struct {
		name    string
		id      string
		wantErr string
	}{
		{
			name: "returns delete response",
			id:   "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:    "returns error for invalid uuid",
			id:      "not-a-uuid",
			wantErr: "invalid job ID",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.Background()

			resp, err := s.sut.DeleteJobByID(ctx, tt.id)

			if tt.wantErr != "" {
				s.ErrorContains(err, tt.wantErr)
				s.Nil(resp)
			} else {
				s.NoError(err)
				s.NotNil(resp)
			}
		})
	}
}

func (s *ClientPublicTestSuite) TestGetJobs() {
	tests := []struct {
		name   string
		status string
		limit  int
		offset int
	}{
		{
			name:   "returns jobs without filter",
			status: "",
		},
		{
			name:   "returns jobs with status filter",
			status: "completed",
		},
		{
			name:   "returns jobs with limit and offset",
			status: "",
			limit:  5,
			offset: 10,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.Background()

			resp, err := s.sut.GetJobs(ctx, tt.status, tt.limit, tt.offset)

			s.NoError(err)
			s.NotNil(resp)
		})
	}
}

func (s *ClientPublicTestSuite) TestRetryJobByID() {
	tests := []struct {
		name           string
		id             string
		targetHostname string
		wantErr        string
	}{
		{
			name:           "returns retry response",
			id:             "550e8400-e29b-41d4-a716-446655440000",
			targetHostname: "_any",
		},
		{
			name:    "returns error for invalid uuid",
			id:      "not-a-uuid",
			wantErr: "invalid job ID",
		},
		{
			name:           "returns retry response with empty target",
			id:             "550e8400-e29b-41d4-a716-446655440000",
			targetHostname: "",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.Background()

			resp, err := s.sut.RetryJobByID(ctx, tt.id, tt.targetHostname)

			if tt.wantErr != "" {
				s.ErrorContains(err, tt.wantErr)
				s.Nil(resp)
			} else {
				s.NoError(err)
				s.NotNil(resp)
			}
		})
	}
}

func (s *ClientPublicTestSuite) TestGetJobQueueStats() {
	tests := []struct {
		name string
	}{
		{
			name: "returns queue stats response",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.Background()

			resp, err := s.sut.GetJobQueueStats(ctx)

			s.NoError(err)
			s.NotNil(resp)
		})
	}
}

func (s *ClientPublicTestSuite) TestGetJobWorkers() {
	tests := []struct {
		name string
	}{
		{
			name: "returns workers response",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.Background()

			resp, err := s.sut.GetJobWorkers(ctx)

			s.NoError(err)
			s.NotNil(resp)
		})
	}
}

func (s *ClientPublicTestSuite) TestGetMetrics() {
	tests := []struct {
		name           string
		url            string
		serverBody     string
		serverStatus   int
		useServer      bool
		wantErr        string
		expectedResult string
	}{
		{
			name:           "returns metrics text",
			useServer:      true,
			serverBody:     "# HELP go_goroutines Number of goroutines.\n# TYPE go_goroutines gauge\ngo_goroutines 42\n",
			serverStatus:   http.StatusOK,
			expectedResult: "# HELP go_goroutines Number of goroutines.\n# TYPE go_goroutines gauge\ngo_goroutines 42\n",
		},
		{
			name:         "returns error on non-200 status",
			useServer:    true,
			serverBody:   "not found",
			serverStatus: http.StatusNotFound,
			wantErr:      "metrics endpoint returned status",
		},
		{
			name:    "returns error when request creation fails",
			url:     "://invalid-url",
			wantErr: "creating metrics request",
		},
		{
			name:    "returns error when server is unreachable",
			url:     "http://127.0.0.1:0",
			wantErr: "fetching metrics",
		},
		{
			name:         "returns error when response body read fails",
			useServer:    true,
			serverStatus: http.StatusOK,
			wantErr:      "reading metrics response",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			url := tt.url

			if tt.useServer {
				server := httptest.NewServer(
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.Header().Set("Content-Type", "text/plain")
						if tt.wantErr == "reading metrics response" {
							w.Header().Set("Content-Length", "9999")
							w.WriteHeader(tt.serverStatus)
							// Write partial data then let the handler return,
							// causing the connection to close prematurely.
							_, _ = w.Write([]byte("partial"))
							if f, ok := w.(http.Flusher); ok {
								f.Flush()
							}
							return
						}
						w.WriteHeader(tt.serverStatus)
						_, _ = w.Write([]byte(tt.serverBody))
					}),
				)
				defer server.Close()
				url = server.URL
			}

			appConfig := config.Config{
				API: config.API{
					Client: config.Client{
						URL: url,
					},
				},
			}
			c := client.New(slog.Default(), appConfig, nil)

			result, err := c.GetMetrics(context.Background())

			if tt.wantErr != "" {
				s.ErrorContains(err, tt.wantErr)
				s.Empty(result)
			} else {
				s.NoError(err)
				s.Equal(tt.expectedResult, result)
			}
		})
	}
}

func (s *ClientPublicTestSuite) TestGetHealth() {
	tests := []struct {
		name string
	}{
		{
			name: "returns health response",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.Background()

			resp, err := s.sut.GetHealth(ctx)

			s.NoError(err)
			s.NotNil(resp)
		})
	}
}

func (s *ClientPublicTestSuite) TestGetHealthReady() {
	tests := []struct {
		name string
	}{
		{
			name: "returns health ready response",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.Background()

			resp, err := s.sut.GetHealthReady(ctx)

			s.NoError(err)
			s.NotNil(resp)
		})
	}
}

func (s *ClientPublicTestSuite) TestGetHealthStatus() {
	tests := []struct {
		name string
	}{
		{
			name: "returns health status response",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.Background()

			resp, err := s.sut.GetHealthStatus(ctx)

			s.NoError(err)
			s.NotNil(resp)
		})
	}
}

func (s *ClientPublicTestSuite) TestGetAuditLogs() {
	tests := []struct {
		name string
	}{
		{
			name: "returns audit logs response",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.Background()

			resp, err := s.sut.GetAuditLogs(ctx, 20, 0)

			s.NoError(err)
			s.NotNil(resp)
		})
	}
}

func (s *ClientPublicTestSuite) TestGetAuditLogByID() {
	tests := []struct {
		name string
	}{
		{
			name: "returns audit log entry response",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.Background()

			resp, err := s.sut.GetAuditLogByID(ctx, "550e8400-e29b-41d4-a716-446655440000")

			s.NoError(err)
			s.NotNil(resp)
		})
	}
}

func TestClientPublicTestSuite(t *testing.T) {
	suite.Run(t, new(ClientPublicTestSuite))
}

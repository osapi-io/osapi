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
	s.genClient, err = client.NewClientWithResponses(s.appConfig)
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
			c, err := client.NewClientWithResponses(tt.config)

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
		expectedHeader string
	}{
		{
			name:           "injects authorization header",
			bearerToken:    "my-secret-token",
			expectedHeader: "Bearer my-secret-token",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			var receivedAuth string
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					receivedAuth = r.Header.Get("Authorization")
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{}`))
				}),
			)
			defer server.Close()

			appConfig := config.Config{
				API: config.API{
					Client: config.Client{
						URL: server.URL,
						Security: config.ClientSecurity{
							BearerToken: tt.bearerToken,
						},
					},
				},
			}

			genClient, err := client.NewClientWithResponses(appConfig)
			s.Require().NoError(err)

			c := client.New(slog.Default(), appConfig, genClient)
			s.NotNil(c)

			_, _ = genClient.GetSystemHostnameWithResponse(context.Background())

			s.Equal(tt.expectedHeader, receivedAuth)
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

			resp, err := s.sut.GetSystemHostname(ctx)

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

			resp, err := s.sut.GetSystemStatus(ctx)

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

			resp, err := s.sut.GetNetworkDNSByInterface(ctx, tt.interfaceName)

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

			resp, err := s.sut.PutNetworkDNS(ctx, tt.servers, tt.searchDomains, tt.interfaceName)

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

			resp, err := s.sut.PostNetworkPing(ctx, tt.target)

			s.NoError(err)
			s.NotNil(resp)
		})
	}
}

func TestClientPublicTestSuite(t *testing.T) {
	suite.Run(t, new(ClientPublicTestSuite))
}

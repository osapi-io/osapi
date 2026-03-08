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
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/pkg/sdk/client"
)

type NodePublicTestSuite struct {
	suite.Suite

	ctx context.Context
}

func (suite *NodePublicTestSuite) SetupTest() {
	suite.ctx = context.Background()
}

func (suite *NodePublicTestSuite) TestHostname() {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		serverURL    string
		target       string
		validateFunc func(*client.Response[client.Collection[client.HostnameResult]], error)
	}{
		{
			name:   "when requesting hostname returns results",
			target: "_any",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(
					[]byte(
						`{"job_id":"00000000-0000-0000-0000-000000000001","results":[{"hostname":"test-host"}]}`,
					),
				)
			},
			validateFunc: func(resp *client.Response[client.Collection[client.HostnameResult]], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal("00000000-0000-0000-0000-000000000001", resp.Data.JobID)
				suite.Len(resp.Data.Results, 1)
				suite.Equal("test-host", resp.Data.Results[0].Hostname)
			},
		},
		{
			name:   "when server returns 403 returns AuthError",
			target: "_any",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
			validateFunc: func(resp *client.Response[client.Collection[client.HostnameResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusForbidden, target.StatusCode)
			},
		},
		{
			name:      "when client HTTP call fails returns error",
			target:    "_any",
			serverURL: "http://127.0.0.1:0",
			validateFunc: func(resp *client.Response[client.Collection[client.HostnameResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "get hostname")
			},
		},
		{
			name:   "when server returns 200 with no JSON body returns UnexpectedStatusError",
			target: "_any",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			validateFunc: func(resp *client.Response[client.Collection[client.HostnameResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusOK, target.StatusCode)
				suite.Equal("nil response body", target.Message)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var (
				serverURL string
				cleanup   func()
			)

			if tc.serverURL != "" {
				serverURL = tc.serverURL
				cleanup = func() {}
			} else {
				server := httptest.NewServer(tc.handler)
				serverURL = server.URL
				cleanup = server.Close
			}
			defer cleanup()

			sut := client.New(
				serverURL,
				"test-token",
				client.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.Hostname(suite.ctx, tc.target)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestStatus() {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		serverURL    string
		target       string
		validateFunc func(*client.Response[client.Collection[client.NodeStatus]], error)
	}{
		{
			name:   "when requesting status returns results",
			target: "_any",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"results":[{"hostname":"web-01"}]}`))
			},
			validateFunc: func(resp *client.Response[client.Collection[client.NodeStatus]], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Len(resp.Data.Results, 1)
				suite.Equal("web-01", resp.Data.Results[0].Hostname)
			},
		},
		{
			name:   "when server returns 403 returns AuthError",
			target: "_any",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
			validateFunc: func(resp *client.Response[client.Collection[client.NodeStatus]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusForbidden, target.StatusCode)
			},
		},
		{
			name:      "when client HTTP call fails returns error",
			target:    "_any",
			serverURL: "http://127.0.0.1:0",
			validateFunc: func(resp *client.Response[client.Collection[client.NodeStatus]], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "get status")
			},
		},
		{
			name:   "when server returns 200 with no JSON body returns UnexpectedStatusError",
			target: "_any",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			validateFunc: func(resp *client.Response[client.Collection[client.NodeStatus]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusOK, target.StatusCode)
				suite.Equal("nil response body", target.Message)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var (
				serverURL string
				cleanup   func()
			)

			if tc.serverURL != "" {
				serverURL = tc.serverURL
				cleanup = func() {}
			} else {
				server := httptest.NewServer(tc.handler)
				serverURL = server.URL
				cleanup = server.Close
			}
			defer cleanup()

			sut := client.New(
				serverURL,
				"test-token",
				client.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.Status(suite.ctx, tc.target)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestDisk() {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		serverURL    string
		target       string
		validateFunc func(*client.Response[client.Collection[client.DiskResult]], error)
	}{
		{
			name:   "when requesting disk returns results",
			target: "_any",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"results":[{"hostname":"disk-host"}]}`))
			},
			validateFunc: func(resp *client.Response[client.Collection[client.DiskResult]], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Len(resp.Data.Results, 1)
				suite.Equal("disk-host", resp.Data.Results[0].Hostname)
			},
		},
		{
			name:   "when server returns 403 returns AuthError",
			target: "_any",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
			validateFunc: func(resp *client.Response[client.Collection[client.DiskResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusForbidden, target.StatusCode)
			},
		},
		{
			name:      "when client HTTP call fails returns error",
			target:    "_any",
			serverURL: "http://127.0.0.1:0",
			validateFunc: func(resp *client.Response[client.Collection[client.DiskResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "get disk")
			},
		},
		{
			name:   "when server returns 200 with no JSON body returns UnexpectedStatusError",
			target: "_any",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			validateFunc: func(resp *client.Response[client.Collection[client.DiskResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusOK, target.StatusCode)
				suite.Equal("nil response body", target.Message)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var (
				serverURL string
				cleanup   func()
			)

			if tc.serverURL != "" {
				serverURL = tc.serverURL
				cleanup = func() {}
			} else {
				server := httptest.NewServer(tc.handler)
				serverURL = server.URL
				cleanup = server.Close
			}
			defer cleanup()

			sut := client.New(
				serverURL,
				"test-token",
				client.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.Disk(suite.ctx, tc.target)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestMemory() {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		serverURL    string
		target       string
		validateFunc func(*client.Response[client.Collection[client.MemoryResult]], error)
	}{
		{
			name:   "when requesting memory returns results",
			target: "_any",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"results":[{"hostname":"mem-host"}]}`))
			},
			validateFunc: func(resp *client.Response[client.Collection[client.MemoryResult]], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Len(resp.Data.Results, 1)
				suite.Equal("mem-host", resp.Data.Results[0].Hostname)
			},
		},
		{
			name:   "when server returns 403 returns AuthError",
			target: "_any",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
			validateFunc: func(resp *client.Response[client.Collection[client.MemoryResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusForbidden, target.StatusCode)
			},
		},
		{
			name:      "when client HTTP call fails returns error",
			target:    "_any",
			serverURL: "http://127.0.0.1:0",
			validateFunc: func(resp *client.Response[client.Collection[client.MemoryResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "get memory")
			},
		},
		{
			name:   "when server returns 200 with no JSON body returns UnexpectedStatusError",
			target: "_any",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			validateFunc: func(resp *client.Response[client.Collection[client.MemoryResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusOK, target.StatusCode)
				suite.Equal("nil response body", target.Message)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var (
				serverURL string
				cleanup   func()
			)

			if tc.serverURL != "" {
				serverURL = tc.serverURL
				cleanup = func() {}
			} else {
				server := httptest.NewServer(tc.handler)
				serverURL = server.URL
				cleanup = server.Close
			}
			defer cleanup()

			sut := client.New(
				serverURL,
				"test-token",
				client.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.Memory(suite.ctx, tc.target)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestLoad() {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		serverURL    string
		target       string
		validateFunc func(*client.Response[client.Collection[client.LoadResult]], error)
	}{
		{
			name:   "when requesting load returns results",
			target: "_any",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"results":[{"hostname":"load-host"}]}`))
			},
			validateFunc: func(resp *client.Response[client.Collection[client.LoadResult]], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Len(resp.Data.Results, 1)
				suite.Equal("load-host", resp.Data.Results[0].Hostname)
			},
		},
		{
			name:   "when server returns 403 returns AuthError",
			target: "_any",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
			validateFunc: func(resp *client.Response[client.Collection[client.LoadResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusForbidden, target.StatusCode)
			},
		},
		{
			name:      "when client HTTP call fails returns error",
			target:    "_any",
			serverURL: "http://127.0.0.1:0",
			validateFunc: func(resp *client.Response[client.Collection[client.LoadResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "get load")
			},
		},
		{
			name:   "when server returns 200 with no JSON body returns UnexpectedStatusError",
			target: "_any",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			validateFunc: func(resp *client.Response[client.Collection[client.LoadResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusOK, target.StatusCode)
				suite.Equal("nil response body", target.Message)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var (
				serverURL string
				cleanup   func()
			)

			if tc.serverURL != "" {
				serverURL = tc.serverURL
				cleanup = func() {}
			} else {
				server := httptest.NewServer(tc.handler)
				serverURL = server.URL
				cleanup = server.Close
			}
			defer cleanup()

			sut := client.New(
				serverURL,
				"test-token",
				client.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.Load(suite.ctx, tc.target)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestOS() {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		serverURL    string
		target       string
		validateFunc func(*client.Response[client.Collection[client.OSInfoResult]], error)
	}{
		{
			name:   "when requesting OS info returns results",
			target: "_any",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"results":[{"hostname":"os-host"}]}`))
			},
			validateFunc: func(resp *client.Response[client.Collection[client.OSInfoResult]], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Len(resp.Data.Results, 1)
				suite.Equal("os-host", resp.Data.Results[0].Hostname)
			},
		},
		{
			name:   "when server returns 403 returns AuthError",
			target: "_any",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
			validateFunc: func(resp *client.Response[client.Collection[client.OSInfoResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusForbidden, target.StatusCode)
			},
		},
		{
			name:      "when client HTTP call fails returns error",
			target:    "_any",
			serverURL: "http://127.0.0.1:0",
			validateFunc: func(resp *client.Response[client.Collection[client.OSInfoResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "get os")
			},
		},
		{
			name:   "when server returns 200 with no JSON body returns UnexpectedStatusError",
			target: "_any",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			validateFunc: func(resp *client.Response[client.Collection[client.OSInfoResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusOK, target.StatusCode)
				suite.Equal("nil response body", target.Message)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var (
				serverURL string
				cleanup   func()
			)

			if tc.serverURL != "" {
				serverURL = tc.serverURL
				cleanup = func() {}
			} else {
				server := httptest.NewServer(tc.handler)
				serverURL = server.URL
				cleanup = server.Close
			}
			defer cleanup()

			sut := client.New(
				serverURL,
				"test-token",
				client.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.OS(suite.ctx, tc.target)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestUptime() {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		serverURL    string
		target       string
		validateFunc func(*client.Response[client.Collection[client.UptimeResult]], error)
	}{
		{
			name:   "when requesting uptime returns results",
			target: "_any",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(
					[]byte(`{"results":[{"hostname":"uptime-host","uptime":"2d3h"}]}`),
				)
			},
			validateFunc: func(resp *client.Response[client.Collection[client.UptimeResult]], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Len(resp.Data.Results, 1)
				suite.Equal("uptime-host", resp.Data.Results[0].Hostname)
				suite.Equal("2d3h", resp.Data.Results[0].Uptime)
			},
		},
		{
			name:   "when server returns 403 returns AuthError",
			target: "_any",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
			validateFunc: func(resp *client.Response[client.Collection[client.UptimeResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusForbidden, target.StatusCode)
			},
		},
		{
			name:      "when client HTTP call fails returns error",
			target:    "_any",
			serverURL: "http://127.0.0.1:0",
			validateFunc: func(resp *client.Response[client.Collection[client.UptimeResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "get uptime")
			},
		},
		{
			name:   "when server returns 200 with no JSON body returns UnexpectedStatusError",
			target: "_any",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			validateFunc: func(resp *client.Response[client.Collection[client.UptimeResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusOK, target.StatusCode)
				suite.Equal("nil response body", target.Message)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var (
				serverURL string
				cleanup   func()
			)

			if tc.serverURL != "" {
				serverURL = tc.serverURL
				cleanup = func() {}
			} else {
				server := httptest.NewServer(tc.handler)
				serverURL = server.URL
				cleanup = server.Close
			}
			defer cleanup()

			sut := client.New(
				serverURL,
				"test-token",
				client.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.Uptime(suite.ctx, tc.target)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestGetDNS() {
	tests := []struct {
		name          string
		handler       http.HandlerFunc
		serverURL     string
		target        string
		interfaceName string
		validateFunc  func(*client.Response[client.Collection[client.DNSConfig]], error)
	}{
		{
			name:          "when requesting DNS returns results",
			target:        "_any",
			interfaceName: "eth0",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(
					[]byte(`{"results":[{"hostname":"dns-host","servers":["8.8.8.8"]}]}`),
				)
			},
			validateFunc: func(resp *client.Response[client.Collection[client.DNSConfig]], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Len(resp.Data.Results, 1)
				suite.Equal("dns-host", resp.Data.Results[0].Hostname)
				suite.Equal([]string{"8.8.8.8"}, resp.Data.Results[0].Servers)
			},
		},
		{
			name:          "when server returns 403 returns AuthError",
			target:        "_any",
			interfaceName: "eth0",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
			validateFunc: func(resp *client.Response[client.Collection[client.DNSConfig]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusForbidden, target.StatusCode)
			},
		},
		{
			name:          "when client HTTP call fails returns error",
			target:        "_any",
			interfaceName: "eth0",
			serverURL:     "http://127.0.0.1:0",
			validateFunc: func(resp *client.Response[client.Collection[client.DNSConfig]], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "get dns")
			},
		},
		{
			name:          "when server returns 200 with no JSON body returns UnexpectedStatusError",
			target:        "_any",
			interfaceName: "eth0",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			validateFunc: func(resp *client.Response[client.Collection[client.DNSConfig]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusOK, target.StatusCode)
				suite.Equal("nil response body", target.Message)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var (
				serverURL string
				cleanup   func()
			)

			if tc.serverURL != "" {
				serverURL = tc.serverURL
				cleanup = func() {}
			} else {
				server := httptest.NewServer(tc.handler)
				serverURL = server.URL
				cleanup = server.Close
			}
			defer cleanup()

			sut := client.New(
				serverURL,
				"test-token",
				client.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.GetDNS(suite.ctx, tc.target, tc.interfaceName)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestUpdateDNS() {
	tests := []struct {
		name          string
		handler       http.HandlerFunc
		serverURL     string
		target        string
		iface         string
		servers       []string
		searchDomains []string
		validateFunc  func(*client.Response[client.Collection[client.DNSUpdateResult]], error)
	}{
		{
			name:          "when servers only provided sets servers",
			target:        "_any",
			iface:         "eth0",
			servers:       []string{"8.8.8.8", "8.8.4.4"},
			searchDomains: nil,
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusAccepted)
				_, _ = w.Write(
					[]byte(
						`{"results":[{"hostname":"dns-host","status":"completed","changed":true}]}`,
					),
				)
			},
			validateFunc: func(resp *client.Response[client.Collection[client.DNSUpdateResult]], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Len(resp.Data.Results, 1)
				suite.Equal("dns-host", resp.Data.Results[0].Hostname)
				suite.Equal("completed", resp.Data.Results[0].Status)
				suite.True(resp.Data.Results[0].Changed)
			},
		},
		{
			name:          "when search domains only provided sets search domains",
			target:        "_any",
			iface:         "eth0",
			servers:       nil,
			searchDomains: []string{"example.com"},
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusAccepted)
				_, _ = w.Write(
					[]byte(
						`{"results":[{"hostname":"dns-host","status":"completed","changed":true}]}`,
					),
				)
			},
			validateFunc: func(resp *client.Response[client.Collection[client.DNSUpdateResult]], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
			},
		},
		{
			name:          "when both provided sets servers and search domains",
			target:        "_any",
			iface:         "eth0",
			servers:       []string{"8.8.8.8"},
			searchDomains: []string{"example.com"},
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusAccepted)
				_, _ = w.Write(
					[]byte(
						`{"results":[{"hostname":"dns-host","status":"completed","changed":true}]}`,
					),
				)
			},
			validateFunc: func(resp *client.Response[client.Collection[client.DNSUpdateResult]], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
			},
		},
		{
			name:          "when neither provided sends empty body",
			target:        "_any",
			iface:         "eth0",
			servers:       nil,
			searchDomains: nil,
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusAccepted)
				_, _ = w.Write(
					[]byte(
						`{"results":[{"hostname":"dns-host","status":"completed","changed":true}]}`,
					),
				)
			},
			validateFunc: func(resp *client.Response[client.Collection[client.DNSUpdateResult]], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
			},
		},
		{
			name:    "when server returns 403 returns AuthError",
			target:  "_any",
			iface:   "eth0",
			servers: []string{"8.8.8.8"},
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
			validateFunc: func(resp *client.Response[client.Collection[client.DNSUpdateResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusForbidden, target.StatusCode)
			},
		},
		{
			name:      "when client HTTP call fails returns error",
			target:    "_any",
			iface:     "eth0",
			servers:   []string{"8.8.8.8"},
			serverURL: "http://127.0.0.1:0",
			validateFunc: func(resp *client.Response[client.Collection[client.DNSUpdateResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "update dns")
			},
		},
		{
			name:    "when server returns 202 with no JSON body returns UnexpectedStatusError",
			target:  "_any",
			iface:   "eth0",
			servers: []string{"8.8.8.8"},
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusAccepted)
			},
			validateFunc: func(resp *client.Response[client.Collection[client.DNSUpdateResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusAccepted, target.StatusCode)
				suite.Equal("nil response body", target.Message)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var (
				serverURL string
				cleanup   func()
			)

			if tc.serverURL != "" {
				serverURL = tc.serverURL
				cleanup = func() {}
			} else {
				server := httptest.NewServer(tc.handler)
				serverURL = server.URL
				cleanup = server.Close
			}
			defer cleanup()

			sut := client.New(
				serverURL,
				"test-token",
				client.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.UpdateDNS(
				suite.ctx,
				tc.target,
				tc.iface,
				tc.servers,
				tc.searchDomains,
			)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestPing() {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		serverURL    string
		target       string
		address      string
		validateFunc func(*client.Response[client.Collection[client.PingResult]], error)
	}{
		{
			name:    "when pinging address returns results",
			target:  "_any",
			address: "8.8.8.8",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(
					[]byte(
						`{"results":[{"hostname":"ping-host","packets_sent":4,"packets_received":4,"packet_loss":0.0}]}`,
					),
				)
			},
			validateFunc: func(resp *client.Response[client.Collection[client.PingResult]], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Len(resp.Data.Results, 1)
				suite.Equal("ping-host", resp.Data.Results[0].Hostname)
				suite.Equal(4, resp.Data.Results[0].PacketsSent)
			},
		},
		{
			name:    "when server returns 403 returns AuthError",
			target:  "_any",
			address: "8.8.8.8",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
			validateFunc: func(resp *client.Response[client.Collection[client.PingResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusForbidden, target.StatusCode)
			},
		},
		{
			name:      "when client HTTP call fails returns error",
			target:    "_any",
			address:   "8.8.8.8",
			serverURL: "http://127.0.0.1:0",
			validateFunc: func(resp *client.Response[client.Collection[client.PingResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "ping")
			},
		},
		{
			name:    "when server returns 200 with no JSON body returns UnexpectedStatusError",
			target:  "_any",
			address: "8.8.8.8",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			validateFunc: func(resp *client.Response[client.Collection[client.PingResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusOK, target.StatusCode)
				suite.Equal("nil response body", target.Message)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var (
				serverURL string
				cleanup   func()
			)

			if tc.serverURL != "" {
				serverURL = tc.serverURL
				cleanup = func() {}
			} else {
				server := httptest.NewServer(tc.handler)
				serverURL = server.URL
				cleanup = server.Close
			}
			defer cleanup()

			sut := client.New(
				serverURL,
				"test-token",
				client.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.Ping(suite.ctx, tc.target, tc.address)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestExec() {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		serverURL    string
		req          client.ExecRequest
		validateFunc func(*client.Response[client.Collection[client.CommandResult]], error)
	}{
		{
			name: "when basic command returns results",
			req: client.ExecRequest{
				Command: "whoami",
				Target:  "_any",
			},
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusAccepted)
				_, _ = w.Write(
					[]byte(
						`{"results":[{"hostname":"exec-host","stdout":"root\n","exit_code":0,"changed":true}]}`,
					),
				)
			},
			validateFunc: func(resp *client.Response[client.Collection[client.CommandResult]], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Len(resp.Data.Results, 1)
				suite.Equal("exec-host", resp.Data.Results[0].Hostname)
				suite.Equal("root\n", resp.Data.Results[0].Stdout)
				suite.Equal(0, resp.Data.Results[0].ExitCode)
			},
		},
		{
			name: "when all options provided returns results",
			req: client.ExecRequest{
				Command: "ls",
				Args:    []string{"-la", "/tmp"},
				Cwd:     "/tmp",
				Timeout: 10,
				Target:  "_any",
			},
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusAccepted)
				_, _ = w.Write(
					[]byte(
						`{"results":[{"hostname":"exec-host","stdout":"root\n","exit_code":0,"changed":true}]}`,
					),
				)
			},
			validateFunc: func(resp *client.Response[client.Collection[client.CommandResult]], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
			},
		},
		{
			name: "when server returns 400 returns ValidationError",
			req: client.ExecRequest{
				Target: "_any",
			},
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"command is required"}`))
			},
			validateFunc: func(resp *client.Response[client.Collection[client.CommandResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.ValidationError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusBadRequest, target.StatusCode)
			},
		},
		{
			name:      "when client HTTP call fails returns error",
			serverURL: "http://127.0.0.1:0",
			req: client.ExecRequest{
				Command: "whoami",
				Target:  "_any",
			},
			validateFunc: func(resp *client.Response[client.Collection[client.CommandResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "exec command")
			},
		},
		{
			name: "when server returns 202 with no JSON body returns UnexpectedStatusError",
			req: client.ExecRequest{
				Command: "whoami",
				Target:  "_any",
			},
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusAccepted)
			},
			validateFunc: func(resp *client.Response[client.Collection[client.CommandResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusAccepted, target.StatusCode)
				suite.Equal("nil response body", target.Message)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var (
				serverURL string
				cleanup   func()
			)

			if tc.serverURL != "" {
				serverURL = tc.serverURL
				cleanup = func() {}
			} else {
				server := httptest.NewServer(tc.handler)
				serverURL = server.URL
				cleanup = server.Close
			}
			defer cleanup()

			sut := client.New(
				serverURL,
				"test-token",
				client.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.Exec(suite.ctx, tc.req)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestShell() {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		serverURL    string
		req          client.ShellRequest
		validateFunc func(*client.Response[client.Collection[client.CommandResult]], error)
	}{
		{
			name: "when basic command returns results",
			req: client.ShellRequest{
				Command: "uname -a",
				Target:  "_any",
			},
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusAccepted)
				_, _ = w.Write(
					[]byte(
						`{"results":[{"hostname":"shell-host","exit_code":0,"changed":false}]}`,
					),
				)
			},
			validateFunc: func(resp *client.Response[client.Collection[client.CommandResult]], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Len(resp.Data.Results, 1)
				suite.Equal("shell-host", resp.Data.Results[0].Hostname)
			},
		},
		{
			name: "when cwd and timeout provided returns results",
			req: client.ShellRequest{
				Command: "ls -la",
				Cwd:     "/var/log",
				Timeout: 15,
				Target:  "_any",
			},
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusAccepted)
				_, _ = w.Write(
					[]byte(
						`{"results":[{"hostname":"shell-host","exit_code":0,"changed":false}]}`,
					),
				)
			},
			validateFunc: func(resp *client.Response[client.Collection[client.CommandResult]], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
			},
		},
		{
			name: "when server returns 400 returns ValidationError",
			req: client.ShellRequest{
				Target: "_any",
			},
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"command is required"}`))
			},
			validateFunc: func(resp *client.Response[client.Collection[client.CommandResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.ValidationError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusBadRequest, target.StatusCode)
			},
		},
		{
			name:      "when client HTTP call fails returns error",
			serverURL: "http://127.0.0.1:0",
			req: client.ShellRequest{
				Command: "uname -a",
				Target:  "_any",
			},
			validateFunc: func(resp *client.Response[client.Collection[client.CommandResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "shell command")
			},
		},
		{
			name: "when server returns 202 with no JSON body returns UnexpectedStatusError",
			req: client.ShellRequest{
				Command: "uname -a",
				Target:  "_any",
			},
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusAccepted)
			},
			validateFunc: func(resp *client.Response[client.Collection[client.CommandResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusAccepted, target.StatusCode)
				suite.Equal("nil response body", target.Message)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var (
				serverURL string
				cleanup   func()
			)

			if tc.serverURL != "" {
				serverURL = tc.serverURL
				cleanup = func() {}
			} else {
				server := httptest.NewServer(tc.handler)
				serverURL = server.URL
				cleanup = server.Close
			}
			defer cleanup()

			sut := client.New(
				serverURL,
				"test-token",
				client.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.Shell(suite.ctx, tc.req)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestFileDeploy() {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		serverURL    string
		req          client.FileDeployOpts
		validateFunc func(*client.Response[client.FileDeployResult], error)
	}{
		{
			name: "when deploying file returns result",
			req: client.FileDeployOpts{
				ObjectName:  "nginx.conf",
				Path:        "/etc/nginx/nginx.conf",
				ContentType: "raw",
				Target:      "_any",
			},
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusAccepted)
				_, _ = w.Write(
					[]byte(
						`{"job_id":"job-123","hostname":"web-01","changed":true}`,
					),
				)
			},
			validateFunc: func(resp *client.Response[client.FileDeployResult], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal("job-123", resp.Data.JobID)
				suite.Equal("web-01", resp.Data.Hostname)
				suite.True(resp.Data.Changed)
			},
		},
		{
			name: "when all options provided returns results",
			req: client.FileDeployOpts{
				ObjectName:  "app.conf.tmpl",
				Path:        "/etc/app/app.conf",
				ContentType: "template",
				Mode:        "0644",
				Owner:       "root",
				Group:       "root",
				Vars:        map[string]any{"port": 8080},
				Target:      "web-01",
			},
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusAccepted)
				_, _ = w.Write(
					[]byte(
						`{"job_id":"job-456","hostname":"web-01","changed":true}`,
					),
				)
			},
			validateFunc: func(resp *client.Response[client.FileDeployResult], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
			},
		},
		{
			name: "when server returns 400 returns ValidationError",
			req: client.FileDeployOpts{
				Target: "_any",
			},
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"object_name is required"}`))
			},
			validateFunc: func(resp *client.Response[client.FileDeployResult], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.ValidationError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusBadRequest, target.StatusCode)
			},
		},
		{
			name: "when server returns 403 returns AuthError",
			req: client.FileDeployOpts{
				ObjectName:  "nginx.conf",
				Path:        "/etc/nginx/nginx.conf",
				ContentType: "raw",
				Target:      "_any",
			},
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
			validateFunc: func(resp *client.Response[client.FileDeployResult], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusForbidden, target.StatusCode)
			},
		},
		{
			name:      "when client HTTP call fails returns error",
			serverURL: "http://127.0.0.1:0",
			req: client.FileDeployOpts{
				ObjectName:  "nginx.conf",
				Path:        "/etc/nginx/nginx.conf",
				ContentType: "raw",
				Target:      "_any",
			},
			validateFunc: func(resp *client.Response[client.FileDeployResult], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "file deploy")
			},
		},
		{
			name: "when server returns 202 with no JSON body returns UnexpectedStatusError",
			req: client.FileDeployOpts{
				ObjectName:  "nginx.conf",
				Path:        "/etc/nginx/nginx.conf",
				ContentType: "raw",
				Target:      "_any",
			},
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusAccepted)
			},
			validateFunc: func(resp *client.Response[client.FileDeployResult], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusAccepted, target.StatusCode)
				suite.Equal("nil response body", target.Message)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var (
				serverURL string
				cleanup   func()
			)

			if tc.serverURL != "" {
				serverURL = tc.serverURL
				cleanup = func() {}
			} else {
				server := httptest.NewServer(tc.handler)
				serverURL = server.URL
				cleanup = server.Close
			}
			defer cleanup()

			sut := client.New(
				serverURL,
				"test-token",
				client.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.FileDeploy(suite.ctx, tc.req)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestFileStatus() {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		serverURL    string
		target       string
		path         string
		validateFunc func(*client.Response[client.FileStatusResult], error)
	}{
		{
			name:   "when checking file status returns result",
			target: "_any",
			path:   "/etc/nginx/nginx.conf",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(
					[]byte(
						`{"job_id":"job-789","hostname":"web-01","path":"/etc/nginx/nginx.conf","status":"in-sync","sha256":"abc123"}`,
					),
				)
			},
			validateFunc: func(resp *client.Response[client.FileStatusResult], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal("job-789", resp.Data.JobID)
				suite.Equal("web-01", resp.Data.Hostname)
				suite.Equal("/etc/nginx/nginx.conf", resp.Data.Path)
				suite.Equal("in-sync", resp.Data.Status)
				suite.Equal("abc123", resp.Data.SHA256)
			},
		},
		{
			name:   "when server returns 400 returns ValidationError",
			target: "_any",
			path:   "",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"path is required"}`))
			},
			validateFunc: func(resp *client.Response[client.FileStatusResult], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.ValidationError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusBadRequest, target.StatusCode)
			},
		},
		{
			name:   "when server returns 403 returns AuthError",
			target: "_any",
			path:   "/etc/nginx/nginx.conf",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
			validateFunc: func(resp *client.Response[client.FileStatusResult], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusForbidden, target.StatusCode)
			},
		},
		{
			name:      "when client HTTP call fails returns error",
			target:    "_any",
			path:      "/etc/nginx/nginx.conf",
			serverURL: "http://127.0.0.1:0",
			validateFunc: func(resp *client.Response[client.FileStatusResult], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "file status")
			},
		},
		{
			name:   "when server returns 200 with no JSON body returns UnexpectedStatusError",
			target: "_any",
			path:   "/etc/nginx/nginx.conf",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			validateFunc: func(resp *client.Response[client.FileStatusResult], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *client.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusOK, target.StatusCode)
				suite.Equal("nil response body", target.Message)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var (
				serverURL string
				cleanup   func()
			)

			if tc.serverURL != "" {
				serverURL = tc.serverURL
				cleanup = func() {}
			} else {
				server := httptest.NewServer(tc.handler)
				serverURL = server.URL
				cleanup = server.Close
			}
			defer cleanup()

			sut := client.New(
				serverURL,
				"test-token",
				client.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.FileStatus(suite.ctx, tc.target, tc.path)
			tc.validateFunc(resp, err)
		})
	}
}

func TestNodePublicTestSuite(t *testing.T) {
	suite.Run(t, new(NodePublicTestSuite))
}

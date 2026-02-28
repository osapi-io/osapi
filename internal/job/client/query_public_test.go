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

package client_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/job/client"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
)

type QueryPublicTestSuite struct {
	suite.Suite

	mockCtrl       *gomock.Controller
	mockNATSClient *jobmocks.MockNATSClient
	mockKV         *jobmocks.MockKeyValue
	jobsClient     *client.Client
	ctx            context.Context
}

func (s *QueryPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockNATSClient = jobmocks.NewMockNATSClient(s.mockCtrl)
	s.mockKV = jobmocks.NewMockKeyValue(s.mockCtrl)
	s.ctx = context.Background()

	opts := &client.Options{
		Timeout:  30 * time.Second,
		KVBucket: s.mockKV,
	}
	var err error
	s.jobsClient, err = client.New(slog.Default(), s.mockNATSClient, opts)
	s.Require().NoError(err)
}

func (s *QueryPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *QueryPublicTestSuite) TestQueryNodeStatus() {
	tests := []struct {
		name          string
		hostname      string
		responseData  string
		mockError     error
		expectError   bool
		errorContains string
	}{
		{
			name:     "success",
			hostname: "server1",
			responseData: `{
				"status": "completed",
				"data": {
					"hostname": "server1",
					"uptime": 3600000000000,
					"os_info": {"name": "Linux", "version": "5.4.0"},
					"load_averages": {"load1": 0.5, "load5": 0.3, "load15": 0.1},
					"memory_stats": {"total": 8589934592, "available": 4294967296},
					"disk_usage": [{"filesystem": "/dev/sda1", "used": 50, "available": 50}]
				}
			}`,
		},
		{
			name:     "job failed",
			hostname: "server1",
			responseData: `{
				"status": "failed",
				"error": "unable to gather system info",
				"data": {}
			}`,
			expectError:   true,
			errorContains: "job failed: unable to gather system info",
		},
		{
			name:     "unmarshal error",
			hostname: "server1",
			responseData: `{
				"status": "completed",
				"data": "invalid_data_format"
			}`,
			expectError:   true,
			errorContains: "failed to unmarshal status response",
		},
		{
			name:     "partial data",
			hostname: "server1",
			responseData: `{
				"status": "completed",
				"data": {
					"hostname": "server1",
					"uptime": 3600000000000
				}
			}`,
		},
		{
			name:          "publish error",
			hostname:      "server1",
			mockError:     errors.New("connection timeout"),
			expectError:   true,
			errorContains: "failed to publish and wait",
		},
		{
			name:     "empty hostname",
			hostname: "",
			responseData: `{
				"status": "completed",
				"data": {
					"hostname": "default-server",
					"uptime": 1000000000000
				}
			}`,
		},
		{
			name:          "invalid JSON response",
			hostname:      "server1",
			responseData:  `{invalid json}`,
			expectError:   true,
			errorContains: "failed to unmarshal response",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, tt.hostname)

			setupPublishAndWaitMocks(
				s.mockCtrl,
				s.mockKV,
				s.mockNATSClient,
				subject,
				tt.responseData,
				tt.mockError,
			)

			_, result, err := s.jobsClient.QueryNodeStatus(s.ctx, tt.hostname)

			if tt.expectError {
				s.Error(err)
				s.Nil(result)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
				s.NotNil(result)
			}
		})
	}
}

func (s *QueryPublicTestSuite) TestQueryNodeHostname() {
	tests := []struct {
		name          string
		hostname      string
		responseData  string
		mockError     error
		expectError   bool
		errorContains string
		validateFunc  func(result string, agent *job.AgentInfo)
	}{
		{
			name:     "success",
			hostname: "server1",
			responseData: `{
				"status": "completed",
				"data": {"hostname": "server1.example.com"}
			}`,
		},
		{
			name:     "success with labels",
			hostname: "server1",
			responseData: `{
				"status": "completed",
				"hostname": "agent1",
				"data": {"hostname": "server1.example.com", "labels": {"group": "web", "env": "prod"}}
			}`,
			validateFunc: func(result string, agent *job.AgentInfo) {
				s.Equal("server1.example.com", result)
				s.Require().NotNil(agent)
				s.Equal("agent1", agent.Hostname)
				s.Equal(map[string]string{"group": "web", "env": "prod"}, agent.Labels)
			},
		},
		{
			name:     "job failed",
			hostname: "server1",
			responseData: `{
				"status": "failed",
				"error": "hostname resolution failed",
				"data": {}
			}`,
			expectError:   true,
			errorContains: "job failed: hostname resolution failed",
		},
		{
			name:     "unmarshal error",
			hostname: "server1",
			responseData: `{
				"status": "completed",
				"data": "invalid_hostname_format"
			}`,
			expectError:   true,
			errorContains: "failed to unmarshal hostname response",
		},
		{
			name:          "publish error",
			hostname:      "server1",
			mockError:     errors.New("network unavailable"),
			expectError:   true,
			errorContains: "failed to publish and wait",
		},
		{
			name:     "empty hostname in response",
			hostname: "server1",
			responseData: `{
				"status": "completed",
				"data": {"hostname": ""}
			}`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, tt.hostname)

			setupPublishAndWaitMocks(
				s.mockCtrl,
				s.mockKV,
				s.mockNATSClient,
				subject,
				tt.responseData,
				tt.mockError,
			)

			_, result, agent, err := s.jobsClient.QueryNodeHostname(s.ctx, tt.hostname)

			if tt.expectError {
				s.Error(err)
				s.Empty(result)
				s.Nil(agent)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
			}

			if tt.validateFunc != nil {
				tt.validateFunc(result, agent)
			}
		})
	}
}

func (s *QueryPublicTestSuite) TestQueryNetworkDNS() {
	tests := []struct {
		name          string
		hostname      string
		iface         string
		responseData  string
		mockError     error
		expectError   bool
		errorContains string
	}{
		{
			name:     "success",
			hostname: "server1",
			iface:    "eth0",
			responseData: `{
				"status": "completed",
				"data": {
					"DNSServers": ["8.8.8.8", "1.1.1.1"],
					"SearchDomains": ["example.com", "local"]
				}
			}`,
		},
		{
			name:     "job failed",
			hostname: "server1",
			iface:    "eth0",
			responseData: `{
				"status": "failed",
				"error": "interface not found",
				"data": {}
			}`,
			expectError:   true,
			errorContains: "job failed: interface not found",
		},
		{
			name:          "publish error",
			hostname:      "server1",
			iface:         "eth0",
			mockError:     errors.New("connection failed"),
			expectError:   true,
			errorContains: "failed to publish and wait",
		},
		{
			name:     "unmarshal error",
			hostname: "server1",
			iface:    "eth0",
			responseData: `{
				"status": "completed",
				"data": "invalid_dns_format"
			}`,
			expectError:   true,
			errorContains: "failed to unmarshal DNS response",
		},
		{
			name:     "complete data",
			hostname: "server1",
			iface:    "eth0",
			responseData: `{
				"status": "completed",
				"data": {
					"DNSServers": ["8.8.8.8", "1.1.1.1", "9.9.9.9"],
					"SearchDomains": ["example.com", "local", "internal"]
				}
			}`,
		},
		{
			name:     "empty interface",
			hostname: "server1",
			iface:    "",
			responseData: `{
				"status": "completed",
				"data": {
					"DNSServers": ["8.8.8.8"],
					"SearchDomains": []
				}
			}`,
		},
		{
			name:     "empty response arrays",
			hostname: "server1",
			iface:    "eth0",
			responseData: `{
				"status": "completed",
				"data": {
					"DNSServers": [],
					"SearchDomains": []
				}
			}`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, tt.hostname)

			setupPublishAndWaitMocks(
				s.mockCtrl,
				s.mockKV,
				s.mockNATSClient,
				subject,
				tt.responseData,
				tt.mockError,
			)

			_, result, _, err := s.jobsClient.QueryNetworkDNS(s.ctx, tt.hostname, tt.iface)

			if tt.expectError {
				s.Error(err)
				s.Nil(result)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
				s.NotNil(result)
			}
		})
	}
}

func (s *QueryPublicTestSuite) TestQueryNetworkPing() {
	tests := []struct {
		name          string
		hostname      string
		address       string
		responseData  string
		mockError     error
		expectError   bool
		errorContains string
	}{
		{
			name:     "success",
			hostname: "server1",
			address:  "google.com",
			responseData: `{
				"status": "completed",
				"data": {
					"PacketsSent": 4,
					"PacketsReceived": 4,
					"PacketLoss": 0.0,
					"MinRTT": 10500000,
					"AvgRTT": 12800000,
					"MaxRTT": 15200000
				}
			}`,
		},
		{
			name:     "job failed",
			hostname: "server1",
			address:  "unreachable.host",
			responseData: `{
				"status": "failed",
				"error": "host unreachable",
				"data": {}
			}`,
			expectError:   true,
			errorContains: "job failed: host unreachable",
		},
		{
			name:          "publish error",
			hostname:      "server1",
			address:       "google.com",
			mockError:     errors.New("timeout"),
			expectError:   true,
			errorContains: "failed to publish and wait",
		},
		{
			name:     "unmarshal error",
			hostname: "server1",
			address:  "google.com",
			responseData: `{
				"status": "completed",
				"data": "invalid_ping_format"
			}`,
			expectError:   true,
			errorContains: "failed to unmarshal ping response",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, tt.hostname)

			setupPublishAndWaitMocks(
				s.mockCtrl,
				s.mockKV,
				s.mockNATSClient,
				subject,
				tt.responseData,
				tt.mockError,
			)

			_, result, _, err := s.jobsClient.QueryNetworkPing(s.ctx, tt.hostname, tt.address)

			if tt.expectError {
				s.Error(err)
				s.Nil(result)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
				s.NotNil(result)
			}
		})
	}
}

func (s *QueryPublicTestSuite) TestQueryNetworkPingAny() {
	tests := []struct {
		name          string
		address       string
		responseData  string
		mockError     error
		expectError   bool
		errorContains string
	}{
		{
			name:    "success",
			address: "google.com",
			responseData: `{
				"status": "completed",
				"data": {
					"PacketsSent": 4,
					"PacketsReceived": 4,
					"PacketLoss": 0.0
				}
			}`,
		},
		{
			name:          "publish error",
			address:       "unreachable.host",
			mockError:     errors.New("no agents available"),
			expectError:   true,
			errorContains: "failed to publish and wait",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			setupPublishAndWaitMocks(
				s.mockCtrl,
				s.mockKV,
				s.mockNATSClient,
				"jobs.query._any",
				tt.responseData,
				tt.mockError,
			)

			_, result, _, err := s.jobsClient.QueryNetworkPingAny(s.ctx, tt.address)

			if tt.expectError {
				s.Error(err)
				s.Nil(result)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
				s.NotNil(result)
			}
		})
	}
}

func (s *QueryPublicTestSuite) TestQueryNodeStatusAny() {
	tests := []struct {
		name          string
		responseData  string
		mockError     error
		expectError   bool
		errorContains string
	}{
		{
			name: "success",
			responseData: `{
				"status": "completed",
				"data": {
					"hostname": "any-server",
					"uptime": 3600000000000
				}
			}`,
		},
		{
			name:          "publish error",
			mockError:     errors.New("no agents available"),
			expectError:   true,
			errorContains: "failed to publish and wait",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			setupPublishAndWaitMocks(
				s.mockCtrl,
				s.mockKV,
				s.mockNATSClient,
				"jobs.query._any",
				tt.responseData,
				tt.mockError,
			)

			_, result, err := s.jobsClient.QueryNodeStatusAny(s.ctx)

			if tt.expectError {
				s.Error(err)
				s.Nil(result)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
				s.NotNil(result)
			}
		})
	}
}

func (s *QueryPublicTestSuite) TestPublishAndWaitErrorPaths() {
	tests := []struct {
		name          string
		opts          *publishAndWaitMockOpts
		timeout       time.Duration
		expectError   bool
		errorContains string
	}{
		{
			name: "publish notification error",
			opts: &publishAndWaitMockOpts{
				mockError: errors.New("stream unavailable"),
				errorMode: errorOnPublish,
			},
			expectError:   true,
			errorContains: "failed to publish notification",
		},
		{
			name: "watch error",
			opts: &publishAndWaitMockOpts{
				mockError: errors.New("watch not supported"),
				errorMode: errorOnWatch,
			},
			expectError:   true,
			errorContains: "failed to create response watcher",
		},
		{
			name:    "timeout waiting for response",
			timeout: 10 * time.Millisecond,
			opts: &publishAndWaitMockOpts{
				mockError: errors.New("unused"),
				errorMode: errorOnTimeout,
			},
			expectError:   true,
			errorContains: "timeout waiting for job response",
		},
		{
			name: "nil entry skipped before real entry",
			opts: &publishAndWaitMockOpts{
				responseData: `{
					"status": "completed",
					"data": {"hostname": "server1.example.com"}
				}`,
				sendNilFirst: true,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			jobsClient := s.jobsClient
			if tt.timeout > 0 {
				opts := &client.Options{
					Timeout:  tt.timeout,
					KVBucket: s.mockKV,
				}
				var err error
				jobsClient, err = client.New(slog.Default(), s.mockNATSClient, opts)
				s.Require().NoError(err)
			}

			setupPublishAndWaitMocksWithOpts(
				s.mockCtrl,
				s.mockKV,
				s.mockNATSClient,
				"jobs.query.host.server1",
				tt.opts,
			)

			_, result, agent, err := jobsClient.QueryNodeHostname(s.ctx, "server1")

			if tt.expectError {
				s.Error(err)
				s.Empty(result)
				s.Nil(agent)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *QueryPublicTestSuite) TestQueryNodeStatusAll() {
	tests := []struct {
		name          string
		timeout       time.Duration
		opts          *publishAndCollectMockOpts
		expectError   bool
		errorContains string
		expectedCount int
	}{
		{
			name:    "multiple hosts respond",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":{"hostname":"server1","uptime":3600000000000}}`,
					`{"status":"completed","hostname":"server2","data":{"hostname":"server2","uptime":7200000000000}}`,
				},
			},
			expectedCount: 2,
		},
		{
			name:    "failed responses collected in errors",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":{"hostname":"server1","uptime":3600000000000}}`,
					`{"status":"failed","hostname":"server2","error":"disk full"}`,
				},
			},
			expectedCount: 1,
		},
		{
			name:    "no agents respond",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				mockError: errors.New("unused"),
				errorMode: errorOnTimeout,
			},
			expectError:   true,
			errorContains: "no agents responded",
		},
		{
			name: "KV put fails",
			opts: &publishAndCollectMockOpts{
				mockError: errors.New("kv unavailable"),
				errorMode: errorOnKVPut,
			},
			expectError:   true,
			errorContains: "failed to store job in KV",
		},
		{
			name: "watch fails",
			opts: &publishAndCollectMockOpts{
				mockError: errors.New("watch error"),
				errorMode: errorOnWatch,
			},
			expectError:   true,
			errorContains: "failed to create response watcher",
		},
		{
			name: "publish fails",
			opts: &publishAndCollectMockOpts{
				mockError: errors.New("publish error"),
				errorMode: errorOnPublish,
			},
			expectError:   true,
			errorContains: "failed to publish notification",
		},
		{
			name:    "unmarshal error skipped",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":{"hostname":"server1","uptime":3600000000000}}`,
					`{"status":"completed","hostname":"server2","data":"invalid_not_object"}`,
				},
			},
			expectedCount: 1,
		},
		{
			name:    "nil watcher entry skipped",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					"",
					`{"status":"completed","hostname":"server1","data":{"hostname":"server1","uptime":3600000000000}}`,
				},
				sendNilFirst: true,
			},
			expectedCount: 1,
		},
		{
			name:    "bad JSON from watcher skipped",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{invalid json}`,
					`{"status":"completed","hostname":"server1","data":{"hostname":"server1","uptime":3600000000000}}`,
				},
			},
			expectedCount: 1,
		},
		{
			name:    "empty hostname falls back to unknown",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"","data":{"hostname":"","uptime":3600000000000}}`,
				},
			},
			expectedCount: 1,
		},
		{
			name:    "response hostname empty uses map key",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":{"uptime":3600000000000}}`,
				},
			},
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			timeout := tt.timeout
			if timeout == 0 {
				timeout = 30 * time.Second
			}

			opts := &client.Options{
				Timeout:  timeout,
				KVBucket: s.mockKV,
			}
			jobsClient, err := client.New(slog.Default(), s.mockNATSClient, opts)
			s.Require().NoError(err)

			setupPublishAndCollectMocks(
				s.mockCtrl,
				s.mockKV,
				s.mockNATSClient,
				"jobs.query._all",
				tt.opts,
			)

			_, result, _, err := jobsClient.QueryNodeStatusAll(s.ctx)

			if tt.expectError {
				s.Error(err)
				s.Nil(result)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
				s.Len(result, tt.expectedCount)
			}
		})
	}
}

func (s *QueryPublicTestSuite) TestQueryNodeHostnameAll() {
	tests := []struct {
		name          string
		timeout       time.Duration
		opts          *publishAndCollectMockOpts
		expectError   bool
		errorContains string
		expectedCount int
	}{
		{
			name:    "multiple hosts respond",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":{"hostname":"host1.example.com"}}`,
					`{"status":"completed","hostname":"server2","data":{"hostname":"host2.example.com"}}`,
				},
			},
			expectedCount: 2,
		},
		{
			name:    "failed responses collected in errors",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":{"hostname":"host1.example.com"}}`,
					`{"status":"failed","hostname":"server2","error":"unreachable"}`,
				},
			},
			expectedCount: 1,
		},
		{
			name:    "unmarshal error skipped",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":{"hostname":"host1.example.com"}}`,
					`{"status":"completed","hostname":"server2","data":"invalid_not_object"}`,
				},
			},
			expectedCount: 1,
		},
		{
			name: "publish error",
			opts: &publishAndCollectMockOpts{
				mockError: errors.New("publish error"),
				errorMode: errorOnPublish,
			},
			expectError:   true,
			errorContains: "failed to collect broadcast responses",
		},
		{
			name:    "no agents respond",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				mockError: errors.New("unused"),
				errorMode: errorOnTimeout,
			},
			expectError:   true,
			errorContains: "no agents responded",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			timeout := tt.timeout
			if timeout == 0 {
				timeout = 30 * time.Second
			}

			opts := &client.Options{
				Timeout:  timeout,
				KVBucket: s.mockKV,
			}
			jobsClient, err := client.New(slog.Default(), s.mockNATSClient, opts)
			s.Require().NoError(err)

			setupPublishAndCollectMocks(
				s.mockCtrl,
				s.mockKV,
				s.mockNATSClient,
				"jobs.query._all",
				tt.opts,
			)

			_, result, _, err := jobsClient.QueryNodeHostnameAll(s.ctx)

			if tt.expectError {
				s.Error(err)
				s.Nil(result)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
				s.Len(result, tt.expectedCount)
			}
		})
	}
}

func (s *QueryPublicTestSuite) TestQueryNetworkDNSAll() {
	tests := []struct {
		name          string
		timeout       time.Duration
		opts          *publishAndCollectMockOpts
		expectError   bool
		errorContains string
		expectedCount int
	}{
		{
			name:    "multiple hosts respond",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":{"DNSServers":["8.8.8.8"],"SearchDomains":["example.com"]}}`,
					`{"status":"completed","hostname":"server2","data":{"DNSServers":["1.1.1.1"],"SearchDomains":["local"]}}`,
				},
			},
			expectedCount: 2,
		},
		{
			name:    "failed responses skipped",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":{"DNSServers":["8.8.8.8"],"SearchDomains":[]}}`,
					`{"status":"failed","hostname":"server2","error":"interface not found"}`,
				},
			},
			expectedCount: 1,
		},
		{
			name:    "unmarshal error skipped",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":{"DNSServers":["8.8.8.8"],"SearchDomains":[]}}`,
					`{"status":"completed","hostname":"server2","data":"not_an_object"}`,
				},
			},
			expectedCount: 1,
		},
		{
			name: "publish error",
			opts: &publishAndCollectMockOpts{
				mockError: errors.New("publish error"),
				errorMode: errorOnPublish,
			},
			expectError:   true,
			errorContains: "failed to collect broadcast responses",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			timeout := tt.timeout
			if timeout == 0 {
				timeout = 30 * time.Second
			}

			opts := &client.Options{
				Timeout:  timeout,
				KVBucket: s.mockKV,
			}
			jobsClient, err := client.New(slog.Default(), s.mockNATSClient, opts)
			s.Require().NoError(err)

			setupPublishAndCollectMocks(
				s.mockCtrl,
				s.mockKV,
				s.mockNATSClient,
				"jobs.query._all",
				tt.opts,
			)

			_, result, _, err := jobsClient.QueryNetworkDNSAll(s.ctx, "eth0")

			if tt.expectError {
				s.Error(err)
				s.Nil(result)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
				s.Len(result, tt.expectedCount)
			}
		})
	}
}

func (s *QueryPublicTestSuite) TestQueryNetworkPingAll() {
	tests := []struct {
		name          string
		timeout       time.Duration
		opts          *publishAndCollectMockOpts
		expectError   bool
		errorContains string
		expectedCount int
	}{
		{
			name:    "multiple hosts respond",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":{"PacketsSent":4,"PacketsReceived":4,"PacketLoss":0.0}}`,
					`{"status":"completed","hostname":"server2","data":{"PacketsSent":4,"PacketsReceived":3,"PacketLoss":25.0}}`,
				},
			},
			expectedCount: 2,
		},
		{
			name:    "failed responses skipped",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":{"PacketsSent":4,"PacketsReceived":4,"PacketLoss":0.0}}`,
					`{"status":"failed","hostname":"server2","error":"host unreachable"}`,
				},
			},
			expectedCount: 1,
		},
		{
			name:    "unmarshal error skipped",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":{"PacketsSent":4,"PacketsReceived":4,"PacketLoss":0.0}}`,
					`{"status":"completed","hostname":"server2","data":"not_an_object"}`,
				},
			},
			expectedCount: 1,
		},
		{
			name: "publish error",
			opts: &publishAndCollectMockOpts{
				mockError: errors.New("publish error"),
				errorMode: errorOnPublish,
			},
			expectError:   true,
			errorContains: "failed to collect broadcast responses",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			timeout := tt.timeout
			if timeout == 0 {
				timeout = 30 * time.Second
			}

			opts := &client.Options{
				Timeout:  timeout,
				KVBucket: s.mockKV,
			}
			jobsClient, err := client.New(slog.Default(), s.mockNATSClient, opts)
			s.Require().NoError(err)

			setupPublishAndCollectMocks(
				s.mockCtrl,
				s.mockKV,
				s.mockNATSClient,
				"jobs.query._all",
				tt.opts,
			)

			_, result, _, err := jobsClient.QueryNetworkPingAll(s.ctx, "1.1.1.1")

			if tt.expectError {
				s.Error(err)
				s.Nil(result)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
				s.Len(result, tt.expectedCount)
			}
		})
	}
}

func (s *QueryPublicTestSuite) TestListAgents() {
	tests := []struct {
		name          string
		setupMockKV   func(*jobmocks.MockKeyValue)
		useRegistryKV bool
		expectError   bool
		errorContains string
		expectedCount int
		validateFunc  func([]job.AgentInfo)
	}{
		{
			name:          "when registryKV is nil returns error",
			useRegistryKV: false,
			expectError:   true,
			errorContains: "agent registry not configured",
		},
		{
			name:          "when bucket is empty returns empty list",
			useRegistryKV: true,
			setupMockKV: func(kv *jobmocks.MockKeyValue) {
				kv.EXPECT().
					Keys(gomock.Any()).
					Return(nil, errors.New("nats: no keys found"))
			},
			expectedCount: 0,
		},
		{
			name:          "when agents exist returns agent list with enriched fields",
			useRegistryKV: true,
			setupMockKV: func(kv *jobmocks.MockKeyValue) {
				kv.EXPECT().
					Keys(gomock.Any()).
					Return([]string{"agents.server1", "agents.server2"}, nil)

				entry1 := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				entry1.EXPECT().Value().Return(
					[]byte(
						`{"hostname":"server1","labels":{"group":"web"},"registered_at":"2026-01-01T00:00:00Z","started_at":"2025-12-31T00:00:00Z","os_info":{"Distribution":"Ubuntu","Version":"24.04"},"uptime":18000000000000,"load_averages":{"Load1":0.5,"Load5":0.3,"Load15":0.2},"memory_stats":{"Total":8388608,"Free":4194304,"Cached":2097152}}`,
					),
				)
				kv.EXPECT().
					Get(gomock.Any(), "agents.server1").
					Return(entry1, nil)

				entry2 := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				entry2.EXPECT().Value().Return(
					[]byte(
						`{"hostname":"server2","labels":{"group":"db"},"registered_at":"2026-01-01T00:00:00Z"}`,
					),
				)
				kv.EXPECT().
					Get(gomock.Any(), "agents.server2").
					Return(entry2, nil)
			},
			expectedCount: 2,
			validateFunc: func(agents []job.AgentInfo) {
				s.Equal("server1", agents[0].Hostname)
				s.NotZero(agents[0].RegisteredAt)
				s.NotZero(agents[0].StartedAt)
				s.NotNil(agents[0].OSInfo)
				s.Equal("Ubuntu", agents[0].OSInfo.Distribution)
				s.NotNil(agents[0].LoadAverages)
				s.NotNil(agents[0].MemoryStats)
				s.NotZero(agents[0].Uptime)
			},
		},
		{
			name:          "when Keys fails returns error",
			useRegistryKV: true,
			setupMockKV: func(kv *jobmocks.MockKeyValue) {
				kv.EXPECT().
					Keys(gomock.Any()).
					Return(nil, errors.New("connection failed"))
			},
			expectError:   true,
			errorContains: "failed to list registry keys",
		},
		{
			name:          "when Get fails for a key skips it",
			useRegistryKV: true,
			setupMockKV: func(kv *jobmocks.MockKeyValue) {
				kv.EXPECT().
					Keys(gomock.Any()).
					Return([]string{"agents.server1", "agents.server2"}, nil)

				entry1 := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				entry1.EXPECT().Value().Return(
					[]byte(`{"hostname":"server1","registered_at":"2026-01-01T00:00:00Z"}`),
				)
				kv.EXPECT().
					Get(gomock.Any(), "agents.server1").
					Return(entry1, nil)

				kv.EXPECT().
					Get(gomock.Any(), "agents.server2").
					Return(nil, errors.New("key not found"))
			},
			expectedCount: 1,
		},
		{
			name:          "when unmarshal fails for a key skips it",
			useRegistryKV: true,
			setupMockKV: func(kv *jobmocks.MockKeyValue) {
				kv.EXPECT().
					Keys(gomock.Any()).
					Return([]string{"agents.server1", "agents.server2"}, nil)

				entry1 := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				entry1.EXPECT().Value().Return(
					[]byte(`{"hostname":"server1","registered_at":"2026-01-01T00:00:00Z"}`),
				)
				kv.EXPECT().
					Get(gomock.Any(), "agents.server1").
					Return(entry1, nil)

				entry2 := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				entry2.EXPECT().Value().Return([]byte(`invalid json`))
				kv.EXPECT().
					Get(gomock.Any(), "agents.server2").
					Return(entry2, nil)
			},
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			registryKV := jobmocks.NewMockKeyValue(s.mockCtrl)
			if tt.setupMockKV != nil {
				tt.setupMockKV(registryKV)
			}

			opts := &client.Options{
				Timeout:  30 * time.Second,
				KVBucket: s.mockKV,
			}
			if tt.useRegistryKV {
				opts.RegistryKV = registryKV
			}

			jobsClient, err := client.New(slog.Default(), s.mockNATSClient, opts)
			s.Require().NoError(err)

			result, err := jobsClient.ListAgents(s.ctx)

			if tt.expectError {
				s.Error(err)
				s.Nil(result)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
				s.Len(result, tt.expectedCount)
				if tt.validateFunc != nil {
					tt.validateFunc(result)
				}
			}
		})
	}
}

func (s *QueryPublicTestSuite) TestGetAgent() {
	tests := []struct {
		name          string
		hostname      string
		setupMockKV   func(*jobmocks.MockKeyValue)
		useRegistryKV bool
		expectError   bool
		errorContains string
		validateFunc  func(*job.AgentInfo)
	}{
		{
			name:          "when registryKV is nil returns error",
			hostname:      "server1",
			useRegistryKV: false,
			expectError:   true,
			errorContains: "agent registry not configured",
		},
		{
			name:          "when agent found returns agent info",
			hostname:      "server1",
			useRegistryKV: true,
			setupMockKV: func(kv *jobmocks.MockKeyValue) {
				entry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				entry.EXPECT().Value().Return(
					[]byte(
						`{"hostname":"server1","labels":{"group":"web"},"registered_at":"2026-01-01T00:00:00Z","started_at":"2025-12-31T00:00:00Z","os_info":{"Distribution":"Ubuntu","Version":"24.04"}}`,
					),
				)
				kv.EXPECT().
					Get(gomock.Any(), "agents.server1").
					Return(entry, nil)
			},
			validateFunc: func(info *job.AgentInfo) {
				s.Equal("server1", info.Hostname)
				s.Equal(map[string]string{"group": "web"}, info.Labels)
				s.NotZero(info.RegisteredAt)
				s.NotZero(info.StartedAt)
				s.NotNil(info.OSInfo)
				s.Equal("Ubuntu", info.OSInfo.Distribution)
			},
		},
		{
			name:          "when agent not found returns error",
			hostname:      "unknown",
			useRegistryKV: true,
			setupMockKV: func(kv *jobmocks.MockKeyValue) {
				kv.EXPECT().
					Get(gomock.Any(), "agents.unknown").
					Return(nil, errors.New("key not found"))
			},
			expectError:   true,
			errorContains: "agent not found",
		},
		{
			name:          "when unmarshal fails returns error",
			hostname:      "server1",
			useRegistryKV: true,
			setupMockKV: func(kv *jobmocks.MockKeyValue) {
				entry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				entry.EXPECT().Value().Return([]byte(`invalid json`))
				kv.EXPECT().
					Get(gomock.Any(), "agents.server1").
					Return(entry, nil)
			},
			expectError:   true,
			errorContains: "failed to unmarshal",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			registryKV := jobmocks.NewMockKeyValue(s.mockCtrl)
			if tt.setupMockKV != nil {
				tt.setupMockKV(registryKV)
			}

			opts := &client.Options{
				Timeout:  30 * time.Second,
				KVBucket: s.mockKV,
			}
			if tt.useRegistryKV {
				opts.RegistryKV = registryKV
			}

			jobsClient, err := client.New(slog.Default(), s.mockNATSClient, opts)
			s.Require().NoError(err)

			result, err := jobsClient.GetAgent(s.ctx, tt.hostname)

			if tt.expectError {
				s.Error(err)
				s.Nil(result)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
				s.NotNil(result)
				if tt.validateFunc != nil {
					tt.validateFunc(result)
				}
			}
		})
	}
}

func (s *QueryPublicTestSuite) TestQueryNodeDisk() {
	tests := []struct {
		name          string
		hostname      string
		responseData  string
		mockError     error
		expectError   bool
		errorContains string
	}{
		{
			name:     "success",
			hostname: "server1",
			responseData: `{
				"status": "completed",
				"data": {"disks":[{"Name":"/dev/sda1","Total":100000,"Used":50000,"Free":50000}]}
			}`,
		},
		{
			name:     "job failed",
			hostname: "server1",
			responseData: `{
				"status": "failed",
				"error": "permission denied",
				"data": {}
			}`,
			expectError:   true,
			errorContains: "job failed: permission denied",
		},
		{
			name:     "unmarshal error",
			hostname: "server1",
			responseData: `{
				"status": "completed",
				"data": "invalid_data_format"
			}`,
			expectError:   true,
			errorContains: "failed to unmarshal disk response",
		},
		{
			name:          "publish error",
			hostname:      "server1",
			mockError:     errors.New("connection timeout"),
			expectError:   true,
			errorContains: "failed to publish and wait",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, tt.hostname)

			setupPublishAndWaitMocks(
				s.mockCtrl,
				s.mockKV,
				s.mockNATSClient,
				subject,
				tt.responseData,
				tt.mockError,
			)

			_, result, _, err := s.jobsClient.QueryNodeDisk(s.ctx, tt.hostname)

			if tt.expectError {
				s.Error(err)
				s.Nil(result)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
				s.NotNil(result)
			}
		})
	}
}

func (s *QueryPublicTestSuite) TestQueryNodeDiskBroadcast() {
	tests := []struct {
		name          string
		timeout       time.Duration
		opts          *publishAndCollectMockOpts
		expectError   bool
		errorContains string
		expectedCount int
	}{
		{
			name:    "multiple hosts respond",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":{"disks":[{"Name":"/dev/sda1","Total":100000,"Used":50000,"Free":50000}]}}`,
					`{"status":"completed","hostname":"server2","data":{"disks":[{"Name":"/dev/sdb1","Total":200000,"Used":100000,"Free":100000}]}}`,
				},
			},
			expectedCount: 2,
		},
		{
			name:    "failed responses collected in errors",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":{"disks":[{"Name":"/dev/sda1","Total":100000,"Used":50000,"Free":50000}]}}`,
					`{"status":"failed","hostname":"server2","error":"disk read error"}`,
				},
			},
			expectedCount: 1,
		},
		{
			name:    "unmarshal error skipped",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":{"disks":[{"Name":"/dev/sda1","Total":100000,"Used":50000,"Free":50000}]}}`,
					`{"status":"completed","hostname":"server2","data":"invalid_data_format"}`,
				},
			},
			expectedCount: 1,
		},
		{
			name: "publish error",
			opts: &publishAndCollectMockOpts{
				mockError: errors.New("publish error"),
				errorMode: errorOnPublish,
			},
			expectError:   true,
			errorContains: "failed to collect broadcast responses",
		},
		{
			name:    "no agents respond",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				mockError: errors.New("unused"),
				errorMode: errorOnTimeout,
			},
			expectError:   true,
			errorContains: "no agents responded",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			timeout := tt.timeout
			if timeout == 0 {
				timeout = 30 * time.Second
			}

			opts := &client.Options{
				Timeout:  timeout,
				KVBucket: s.mockKV,
			}
			jobsClient, err := client.New(slog.Default(), s.mockNATSClient, opts)
			s.Require().NoError(err)

			setupPublishAndCollectMocks(
				s.mockCtrl,
				s.mockKV,
				s.mockNATSClient,
				"jobs.query._all",
				tt.opts,
			)

			_, result, _, err := jobsClient.QueryNodeDiskBroadcast(s.ctx, "_all")

			if tt.expectError {
				s.Error(err)
				s.Nil(result)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
				s.Len(result, tt.expectedCount)
			}
		})
	}
}

func (s *QueryPublicTestSuite) TestQueryNodeMemory() {
	tests := []struct {
		name          string
		hostname      string
		responseData  string
		mockError     error
		expectError   bool
		errorContains string
	}{
		{
			name:     "success",
			hostname: "server1",
			responseData: `{
				"status": "completed",
				"data": {"Total":8589934592,"Free":4294967296,"Cached":1073741824}
			}`,
		},
		{
			name:     "job failed",
			hostname: "server1",
			responseData: `{
				"status": "failed",
				"error": "memory read error",
				"data": {}
			}`,
			expectError:   true,
			errorContains: "job failed: memory read error",
		},
		{
			name:     "unmarshal error",
			hostname: "server1",
			responseData: `{
				"status": "completed",
				"data": "invalid_data_format"
			}`,
			expectError:   true,
			errorContains: "failed to unmarshal memory response",
		},
		{
			name:          "publish error",
			hostname:      "server1",
			mockError:     errors.New("connection timeout"),
			expectError:   true,
			errorContains: "failed to publish and wait",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, tt.hostname)

			setupPublishAndWaitMocks(
				s.mockCtrl,
				s.mockKV,
				s.mockNATSClient,
				subject,
				tt.responseData,
				tt.mockError,
			)

			_, result, _, err := s.jobsClient.QueryNodeMemory(s.ctx, tt.hostname)

			if tt.expectError {
				s.Error(err)
				s.Nil(result)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
				s.NotNil(result)
			}
		})
	}
}

func (s *QueryPublicTestSuite) TestQueryNodeMemoryBroadcast() {
	tests := []struct {
		name          string
		timeout       time.Duration
		opts          *publishAndCollectMockOpts
		expectError   bool
		errorContains string
		expectedCount int
	}{
		{
			name:    "multiple hosts respond",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":{"Total":8589934592,"Free":4294967296,"Cached":1073741824}}`,
					`{"status":"completed","hostname":"server2","data":{"Total":16777216000,"Free":8388608000,"Cached":2147483648}}`,
				},
			},
			expectedCount: 2,
		},
		{
			name:    "failed responses collected in errors",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":{"Total":8589934592,"Free":4294967296,"Cached":1073741824}}`,
					`{"status":"failed","hostname":"server2","error":"memory read error"}`,
				},
			},
			expectedCount: 1,
		},
		{
			name:    "unmarshal error skipped",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":{"Total":8589934592,"Free":4294967296,"Cached":1073741824}}`,
					`{"status":"completed","hostname":"server2","data":"invalid_data_format"}`,
				},
			},
			expectedCount: 1,
		},
		{
			name: "publish error",
			opts: &publishAndCollectMockOpts{
				mockError: errors.New("publish error"),
				errorMode: errorOnPublish,
			},
			expectError:   true,
			errorContains: "failed to collect broadcast responses",
		},
		{
			name:    "no agents respond",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				mockError: errors.New("unused"),
				errorMode: errorOnTimeout,
			},
			expectError:   true,
			errorContains: "no agents responded",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			timeout := tt.timeout
			if timeout == 0 {
				timeout = 30 * time.Second
			}

			opts := &client.Options{
				Timeout:  timeout,
				KVBucket: s.mockKV,
			}
			jobsClient, err := client.New(slog.Default(), s.mockNATSClient, opts)
			s.Require().NoError(err)

			setupPublishAndCollectMocks(
				s.mockCtrl,
				s.mockKV,
				s.mockNATSClient,
				"jobs.query._all",
				tt.opts,
			)

			_, result, _, err := jobsClient.QueryNodeMemoryBroadcast(s.ctx, "_all")

			if tt.expectError {
				s.Error(err)
				s.Nil(result)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
				s.Len(result, tt.expectedCount)
			}
		})
	}
}

func (s *QueryPublicTestSuite) TestQueryNodeLoad() {
	tests := []struct {
		name          string
		hostname      string
		responseData  string
		mockError     error
		expectError   bool
		errorContains string
	}{
		{
			name:     "success",
			hostname: "server1",
			responseData: `{
				"status": "completed",
				"data": {"Load1":0.5,"Load5":0.3,"Load15":0.2}
			}`,
		},
		{
			name:     "job failed",
			hostname: "server1",
			responseData: `{
				"status": "failed",
				"error": "load read error",
				"data": {}
			}`,
			expectError:   true,
			errorContains: "job failed: load read error",
		},
		{
			name:     "unmarshal error",
			hostname: "server1",
			responseData: `{
				"status": "completed",
				"data": "invalid_data_format"
			}`,
			expectError:   true,
			errorContains: "failed to unmarshal load response",
		},
		{
			name:          "publish error",
			hostname:      "server1",
			mockError:     errors.New("connection timeout"),
			expectError:   true,
			errorContains: "failed to publish and wait",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, tt.hostname)

			setupPublishAndWaitMocks(
				s.mockCtrl,
				s.mockKV,
				s.mockNATSClient,
				subject,
				tt.responseData,
				tt.mockError,
			)

			_, result, _, err := s.jobsClient.QueryNodeLoad(s.ctx, tt.hostname)

			if tt.expectError {
				s.Error(err)
				s.Nil(result)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
				s.NotNil(result)
			}
		})
	}
}

func (s *QueryPublicTestSuite) TestQueryNodeLoadBroadcast() {
	tests := []struct {
		name          string
		timeout       time.Duration
		opts          *publishAndCollectMockOpts
		expectError   bool
		errorContains string
		expectedCount int
	}{
		{
			name:    "multiple hosts respond",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":{"Load1":0.5,"Load5":0.3,"Load15":0.2}}`,
					`{"status":"completed","hostname":"server2","data":{"Load1":1.2,"Load5":0.8,"Load15":0.5}}`,
				},
			},
			expectedCount: 2,
		},
		{
			name:    "failed responses collected in errors",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":{"Load1":0.5,"Load5":0.3,"Load15":0.2}}`,
					`{"status":"failed","hostname":"server2","error":"load read error"}`,
				},
			},
			expectedCount: 1,
		},
		{
			name:    "unmarshal error skipped",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":{"Load1":0.5,"Load5":0.3,"Load15":0.2}}`,
					`{"status":"completed","hostname":"server2","data":"invalid_data_format"}`,
				},
			},
			expectedCount: 1,
		},
		{
			name: "publish error",
			opts: &publishAndCollectMockOpts{
				mockError: errors.New("publish error"),
				errorMode: errorOnPublish,
			},
			expectError:   true,
			errorContains: "failed to collect broadcast responses",
		},
		{
			name:    "no agents respond",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				mockError: errors.New("unused"),
				errorMode: errorOnTimeout,
			},
			expectError:   true,
			errorContains: "no agents responded",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			timeout := tt.timeout
			if timeout == 0 {
				timeout = 30 * time.Second
			}

			opts := &client.Options{
				Timeout:  timeout,
				KVBucket: s.mockKV,
			}
			jobsClient, err := client.New(slog.Default(), s.mockNATSClient, opts)
			s.Require().NoError(err)

			setupPublishAndCollectMocks(
				s.mockCtrl,
				s.mockKV,
				s.mockNATSClient,
				"jobs.query._all",
				tt.opts,
			)

			_, result, _, err := jobsClient.QueryNodeLoadBroadcast(s.ctx, "_all")

			if tt.expectError {
				s.Error(err)
				s.Nil(result)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
				s.Len(result, tt.expectedCount)
			}
		})
	}
}

func (s *QueryPublicTestSuite) TestQueryNodeOS() {
	tests := []struct {
		name          string
		hostname      string
		responseData  string
		mockError     error
		expectError   bool
		errorContains string
	}{
		{
			name:     "success",
			hostname: "server1",
			responseData: `{
				"status": "completed",
				"data": {"Distribution":"Ubuntu","Version":"24.04"}
			}`,
		},
		{
			name:     "job failed",
			hostname: "server1",
			responseData: `{
				"status": "failed",
				"error": "os info unavailable",
				"data": {}
			}`,
			expectError:   true,
			errorContains: "job failed: os info unavailable",
		},
		{
			name:     "unmarshal error",
			hostname: "server1",
			responseData: `{
				"status": "completed",
				"data": "invalid_data_format"
			}`,
			expectError:   true,
			errorContains: "failed to unmarshal OS info response",
		},
		{
			name:          "publish error",
			hostname:      "server1",
			mockError:     errors.New("connection timeout"),
			expectError:   true,
			errorContains: "failed to publish and wait",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, tt.hostname)

			setupPublishAndWaitMocks(
				s.mockCtrl,
				s.mockKV,
				s.mockNATSClient,
				subject,
				tt.responseData,
				tt.mockError,
			)

			_, result, _, err := s.jobsClient.QueryNodeOS(s.ctx, tt.hostname)

			if tt.expectError {
				s.Error(err)
				s.Nil(result)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
				s.NotNil(result)
			}
		})
	}
}

func (s *QueryPublicTestSuite) TestQueryNodeOSBroadcast() {
	tests := []struct {
		name          string
		timeout       time.Duration
		opts          *publishAndCollectMockOpts
		expectError   bool
		errorContains string
		expectedCount int
	}{
		{
			name:    "multiple hosts respond",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":{"Distribution":"Ubuntu","Version":"24.04"}}`,
					`{"status":"completed","hostname":"server2","data":{"Distribution":"CentOS","Version":"9"}}`,
				},
			},
			expectedCount: 2,
		},
		{
			name:    "failed responses collected in errors",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":{"Distribution":"Ubuntu","Version":"24.04"}}`,
					`{"status":"failed","hostname":"server2","error":"os info unavailable"}`,
				},
			},
			expectedCount: 1,
		},
		{
			name:    "unmarshal error skipped",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":{"Distribution":"Ubuntu","Version":"24.04"}}`,
					`{"status":"completed","hostname":"server2","data":"invalid_data_format"}`,
				},
			},
			expectedCount: 1,
		},
		{
			name: "publish error",
			opts: &publishAndCollectMockOpts{
				mockError: errors.New("publish error"),
				errorMode: errorOnPublish,
			},
			expectError:   true,
			errorContains: "failed to collect broadcast responses",
		},
		{
			name:    "no agents respond",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				mockError: errors.New("unused"),
				errorMode: errorOnTimeout,
			},
			expectError:   true,
			errorContains: "no agents responded",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			timeout := tt.timeout
			if timeout == 0 {
				timeout = 30 * time.Second
			}

			opts := &client.Options{
				Timeout:  timeout,
				KVBucket: s.mockKV,
			}
			jobsClient, err := client.New(slog.Default(), s.mockNATSClient, opts)
			s.Require().NoError(err)

			setupPublishAndCollectMocks(
				s.mockCtrl,
				s.mockKV,
				s.mockNATSClient,
				"jobs.query._all",
				tt.opts,
			)

			_, result, _, err := jobsClient.QueryNodeOSBroadcast(s.ctx, "_all")

			if tt.expectError {
				s.Error(err)
				s.Nil(result)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
				s.Len(result, tt.expectedCount)
			}
		})
	}
}

func (s *QueryPublicTestSuite) TestQueryNodeUptime() {
	tests := []struct {
		name          string
		hostname      string
		responseData  string
		mockError     error
		expectError   bool
		errorContains string
	}{
		{
			name:     "success",
			hostname: "server1",
			responseData: `{
				"status": "completed",
				"data": {"uptime_seconds":3600,"uptime":"1h0m0s"}
			}`,
		},
		{
			name:     "job failed",
			hostname: "server1",
			responseData: `{
				"status": "failed",
				"error": "uptime read error",
				"data": {}
			}`,
			expectError:   true,
			errorContains: "job failed: uptime read error",
		},
		{
			name:     "unmarshal error",
			hostname: "server1",
			responseData: `{
				"status": "completed",
				"data": "invalid_data_format"
			}`,
			expectError:   true,
			errorContains: "failed to unmarshal uptime response",
		},
		{
			name:          "publish error",
			hostname:      "server1",
			mockError:     errors.New("connection timeout"),
			expectError:   true,
			errorContains: "failed to publish and wait",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, tt.hostname)

			setupPublishAndWaitMocks(
				s.mockCtrl,
				s.mockKV,
				s.mockNATSClient,
				subject,
				tt.responseData,
				tt.mockError,
			)

			_, result, _, err := s.jobsClient.QueryNodeUptime(s.ctx, tt.hostname)

			if tt.expectError {
				s.Error(err)
				s.Nil(result)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
				s.NotNil(result)
			}
		})
	}
}

func (s *QueryPublicTestSuite) TestQueryNodeUptimeBroadcast() {
	tests := []struct {
		name          string
		timeout       time.Duration
		opts          *publishAndCollectMockOpts
		expectError   bool
		errorContains string
		expectedCount int
	}{
		{
			name:    "multiple hosts respond",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":{"uptime_seconds":3600,"uptime":"1h0m0s"}}`,
					`{"status":"completed","hostname":"server2","data":{"uptime_seconds":7200,"uptime":"2h0m0s"}}`,
				},
			},
			expectedCount: 2,
		},
		{
			name:    "failed responses collected in errors",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":{"uptime_seconds":3600,"uptime":"1h0m0s"}}`,
					`{"status":"failed","hostname":"server2","error":"uptime read error"}`,
				},
			},
			expectedCount: 1,
		},
		{
			name:    "unmarshal error skipped",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":{"uptime_seconds":3600,"uptime":"1h0m0s"}}`,
					`{"status":"completed","hostname":"server2","data":"invalid_data_format"}`,
				},
			},
			expectedCount: 1,
		},
		{
			name: "publish error",
			opts: &publishAndCollectMockOpts{
				mockError: errors.New("publish error"),
				errorMode: errorOnPublish,
			},
			expectError:   true,
			errorContains: "failed to collect broadcast responses",
		},
		{
			name:    "no agents respond",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				mockError: errors.New("unused"),
				errorMode: errorOnTimeout,
			},
			expectError:   true,
			errorContains: "no agents responded",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			timeout := tt.timeout
			if timeout == 0 {
				timeout = 30 * time.Second
			}

			opts := &client.Options{
				Timeout:  timeout,
				KVBucket: s.mockKV,
			}
			jobsClient, err := client.New(slog.Default(), s.mockNATSClient, opts)
			s.Require().NoError(err)

			setupPublishAndCollectMocks(
				s.mockCtrl,
				s.mockKV,
				s.mockNATSClient,
				"jobs.query._all",
				tt.opts,
			)

			_, result, _, err := jobsClient.QueryNodeUptimeBroadcast(s.ctx, "_all")

			if tt.expectError {
				s.Error(err)
				s.Nil(result)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
				s.Len(result, tt.expectedCount)
			}
		})
	}
}

func TestQueryPublicTestSuite(t *testing.T) {
	suite.Run(t, new(QueryPublicTestSuite))
}

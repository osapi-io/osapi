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

func (s *QueryPublicTestSuite) TestQuerySystemStatus() {
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
			subject := "jobs.query." + tt.hostname

			if tt.mockError != nil {
				s.mockNATSClient.EXPECT().
					PublishAndWaitKV(gomock.Any(), subject, gomock.Any(), s.mockKV, gomock.Any()).
					Return(nil, tt.mockError)
			} else {
				s.mockNATSClient.EXPECT().
					PublishAndWaitKV(gomock.Any(), subject, gomock.Any(), s.mockKV, gomock.Any()).
					Return([]byte(tt.responseData), nil)
			}

			result, err := s.jobsClient.QuerySystemStatus(s.ctx, tt.hostname)

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

func (s *QueryPublicTestSuite) TestQuerySystemHostname() {
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
				"data": {"hostname": "server1.example.com"}
			}`,
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
			subject := "jobs.query." + tt.hostname

			if tt.mockError != nil {
				s.mockNATSClient.EXPECT().
					PublishAndWaitKV(gomock.Any(), subject, gomock.Any(), s.mockKV, gomock.Any()).
					Return(nil, tt.mockError)
			} else {
				s.mockNATSClient.EXPECT().
					PublishAndWaitKV(gomock.Any(), subject, gomock.Any(), s.mockKV, gomock.Any()).
					Return([]byte(tt.responseData), nil)
			}

			result, err := s.jobsClient.QuerySystemHostname(s.ctx, tt.hostname)

			if tt.expectError {
				s.Error(err)
				s.Empty(result)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
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
			subject := "jobs.query." + tt.hostname

			if tt.mockError != nil {
				s.mockNATSClient.EXPECT().
					PublishAndWaitKV(gomock.Any(), subject, gomock.Any(), s.mockKV, gomock.Any()).
					Return(nil, tt.mockError)
			} else {
				s.mockNATSClient.EXPECT().
					PublishAndWaitKV(gomock.Any(), subject, gomock.Any(), s.mockKV, gomock.Any()).
					Return([]byte(tt.responseData), nil)
			}

			result, err := s.jobsClient.QueryNetworkDNS(s.ctx, tt.hostname, tt.iface)

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
			subject := "jobs.query." + tt.hostname

			if tt.mockError != nil {
				s.mockNATSClient.EXPECT().
					PublishAndWaitKV(gomock.Any(), subject, gomock.Any(), s.mockKV, gomock.Any()).
					Return(nil, tt.mockError)
			} else {
				s.mockNATSClient.EXPECT().
					PublishAndWaitKV(gomock.Any(), subject, gomock.Any(), s.mockKV, gomock.Any()).
					Return([]byte(tt.responseData), nil)
			}

			result, err := s.jobsClient.QueryNetworkPing(s.ctx, tt.hostname, tt.address)

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
			mockError:     errors.New("no workers available"),
			expectError:   true,
			errorContains: "failed to publish and wait",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.mockError != nil {
				s.mockNATSClient.EXPECT().
					PublishAndWaitKV(gomock.Any(), "jobs.query._any", gomock.Any(), s.mockKV, gomock.Any()).
					Return(nil, tt.mockError)
			} else {
				s.mockNATSClient.EXPECT().
					PublishAndWaitKV(gomock.Any(), "jobs.query._any", gomock.Any(), s.mockKV, gomock.Any()).
					Return([]byte(tt.responseData), nil)
			}

			result, err := s.jobsClient.QueryNetworkPingAny(s.ctx, tt.address)

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

func (s *QueryPublicTestSuite) TestQuerySystemStatusAny() {
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
			mockError:     errors.New("no workers available"),
			expectError:   true,
			errorContains: "failed to publish and wait",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.mockError != nil {
				s.mockNATSClient.EXPECT().
					PublishAndWaitKV(gomock.Any(), "jobs.query._any", gomock.Any(), s.mockKV, gomock.Any()).
					Return(nil, tt.mockError)
			} else {
				s.mockNATSClient.EXPECT().
					PublishAndWaitKV(gomock.Any(), "jobs.query._any", gomock.Any(), s.mockKV, gomock.Any()).
					Return([]byte(tt.responseData), nil)
			}

			result, err := s.jobsClient.QuerySystemStatusAny(s.ctx)

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

func (s *QueryPublicTestSuite) TestQuerySystemStatusAll() {
	tests := []struct {
		name          string
		expectError   bool
		errorContains string
	}{
		{
			name:          "not implemented",
			expectError:   true,
			errorContains: "broadcast queries not yet implemented",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result, err := s.jobsClient.QuerySystemStatusAll(s.ctx)

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

func TestQueryPublicTestSuite(t *testing.T) {
	suite.Run(t, new(QueryPublicTestSuite))
}

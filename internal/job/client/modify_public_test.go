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

type ModifyPublicTestSuite struct {
	suite.Suite

	mockCtrl       *gomock.Controller
	mockNATSClient *jobmocks.MockNATSClient
	mockKV         *jobmocks.MockKeyValue
	jobsClient     *client.Client
	ctx            context.Context
}

func (s *ModifyPublicTestSuite) SetupTest() {
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

func (s *ModifyPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *ModifyPublicTestSuite) TestModifyNetworkDNS() {
	tests := []struct {
		name          string
		hostname      string
		servers       []string
		searchDomains []string
		iface         string
		responseData  string
		mockError     error
		expectError   bool
		errorContains string
	}{
		{
			name:          "success",
			hostname:      "server1",
			servers:       []string{"8.8.8.8", "1.1.1.1"},
			searchDomains: []string{"example.com", "local"},
			iface:         "eth0",
			responseData: `{
				"status": "completed",
				"data": {"updated": true}
			}`,
			expectError: false,
		},
		{
			name:          "job failed",
			hostname:      "server1",
			servers:       []string{"invalid.ip"},
			searchDomains: []string{"example.com"},
			iface:         "eth0",
			responseData: `{
				"status": "failed",
				"error": "invalid DNS server address",
				"data": {}
			}`,
			expectError:   true,
			errorContains: "job failed: invalid DNS server address",
		},
		{
			name:          "publish error",
			hostname:      "server1",
			servers:       []string{"8.8.8.8"},
			searchDomains: []string{"example.com"},
			iface:         "eth0",
			mockError:     errors.New("connection failed"),
			expectError:   true,
			errorContains: "failed to publish and wait",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.mockError != nil {
				s.mockNATSClient.EXPECT().
					PublishAndWaitKV(gomock.Any(), "jobs.modify.server1", gomock.Any(), s.mockKV, gomock.Any()).
					Return(nil, tt.mockError)
			} else {
				s.mockNATSClient.EXPECT().
					PublishAndWaitKV(gomock.Any(), "jobs.modify.server1", gomock.Any(), s.mockKV, gomock.Any()).
					Return([]byte(tt.responseData), nil)
			}

			err := s.jobsClient.ModifyNetworkDNS(
				s.ctx,
				tt.hostname,
				tt.servers,
				tt.searchDomains,
				tt.iface,
			)

			if tt.expectError {
				s.Error(err)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *ModifyPublicTestSuite) TestModifyNetworkPing() {
	tests := []struct {
		name          string
		hostname      string
		address       string
		responseData  string
		mockError     error
		expectError   bool
		errorContains string
		validateFunc  func(result interface{}, err error)
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
			expectError: false,
			validateFunc: func(result interface{}, err error) {
				s.NoError(err)
				s.NotNil(result)
				// Note: result is of type *ping.Result, but we can't access fields without type assertion
			},
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
			validateFunc: func(result interface{}, err error) {
				s.Error(err)
				s.Nil(result)
			},
		},
		{
			name:          "publish error",
			hostname:      "server1",
			address:       "google.com",
			mockError:     errors.New("timeout"),
			expectError:   true,
			errorContains: "failed to publish and wait",
			validateFunc: func(result interface{}, err error) {
				s.Error(err)
				s.Nil(result)
			},
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
			validateFunc: func(result interface{}, err error) {
				s.Error(err)
				s.Nil(result)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.mockError != nil {
				s.mockNATSClient.EXPECT().
					PublishAndWaitKV(gomock.Any(), "jobs.modify.server1", gomock.Any(), s.mockKV, gomock.Any()).
					Return(nil, tt.mockError)
			} else {
				s.mockNATSClient.EXPECT().
					PublishAndWaitKV(gomock.Any(), "jobs.modify.server1", gomock.Any(), s.mockKV, gomock.Any()).
					Return([]byte(tt.responseData), nil)
			}

			result, err := s.jobsClient.ModifyNetworkPing(s.ctx, tt.hostname, tt.address)

			if tt.validateFunc != nil {
				tt.validateFunc(result, err)
			}

			if tt.expectError {
				s.Error(err)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *ModifyPublicTestSuite) TestModifyNetworkDNSAny() {
	tests := []struct {
		name          string
		servers       []string
		searchDomains []string
		iface         string
		responseData  string
		mockError     error
		expectError   bool
		errorContains string
	}{
		{
			name:          "success",
			servers:       []string{"8.8.8.8", "1.1.1.1"},
			searchDomains: []string{"example.com"},
			iface:         "eth0",
			responseData: `{
				"status": "completed",
				"data": {"updated": true}
			}`,
			expectError: false,
		},
		{
			name:          "error",
			servers:       []string{"invalid.ip"},
			searchDomains: []string{"example.com"},
			iface:         "eth0",
			mockError:     errors.New("no workers available"),
			expectError:   true,
			errorContains: "failed to publish and wait",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.mockError != nil {
				s.mockNATSClient.EXPECT().
					PublishAndWaitKV(gomock.Any(), "jobs.modify._any", gomock.Any(), s.mockKV, gomock.Any()).
					Return(nil, tt.mockError)
			} else {
				s.mockNATSClient.EXPECT().
					PublishAndWaitKV(gomock.Any(), "jobs.modify._any", gomock.Any(), s.mockKV, gomock.Any()).
					Return([]byte(tt.responseData), nil)
			}

			err := s.jobsClient.ModifyNetworkDNSAny(s.ctx, tt.servers, tt.searchDomains, tt.iface)

			if tt.expectError {
				s.Error(err)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *ModifyPublicTestSuite) TestModifyNetworkPingAny() {
	tests := []struct {
		name          string
		address       string
		responseData  string
		mockError     error
		expectError   bool
		errorContains string
		validateFunc  func(result interface{}, err error)
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
			expectError: false,
			validateFunc: func(result interface{}, err error) {
				s.NoError(err)
				s.NotNil(result)
			},
		},
		{
			name:          "error",
			address:       "unreachable.host",
			mockError:     errors.New("no workers available"),
			expectError:   true,
			errorContains: "failed to publish and wait",
			validateFunc: func(result interface{}, err error) {
				s.Error(err)
				s.Nil(result)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.mockError != nil {
				s.mockNATSClient.EXPECT().
					PublishAndWaitKV(gomock.Any(), "jobs.modify._any", gomock.Any(), s.mockKV, gomock.Any()).
					Return(nil, tt.mockError)
			} else {
				s.mockNATSClient.EXPECT().
					PublishAndWaitKV(gomock.Any(), "jobs.modify._any", gomock.Any(), s.mockKV, gomock.Any()).
					Return([]byte(tt.responseData), nil)
			}

			result, err := s.jobsClient.ModifyNetworkPingAny(s.ctx, tt.address)

			if tt.validateFunc != nil {
				tt.validateFunc(result, err)
			}

			if tt.expectError {
				s.Error(err)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
			}
		})
	}
}

func TestModifyPublicTestSuite(t *testing.T) {
	suite.Run(t, new(ModifyPublicTestSuite))
}

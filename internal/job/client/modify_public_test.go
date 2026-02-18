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
			setupPublishAndWaitMocks(
				s.mockCtrl,
				s.mockKV,
				s.mockNATSClient,
				"jobs.modify.server1",
				tt.responseData,
				tt.mockError,
			)

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
			setupPublishAndWaitMocks(
				s.mockCtrl,
				s.mockKV,
				s.mockNATSClient,
				"jobs.modify._any",
				tt.responseData,
				tt.mockError,
			)

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

func (s *ModifyPublicTestSuite) TestModifyNetworkDNSAll() {
	tests := []struct {
		name          string
		timeout       time.Duration
		opts          *publishAndCollectMockOpts
		expectError   bool
		errorContains string
		expectedCount int
		expectHostErr bool
	}{
		{
			name:    "all hosts succeed",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":{"updated":true}}`,
					`{"status":"completed","hostname":"server2","data":{"updated":true}}`,
				},
			},
			expectedCount: 2,
		},
		{
			name:    "partial failure",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":{"updated":true}}`,
					`{"status":"failed","hostname":"server2","error":"disk full"}`,
				},
			},
			expectedCount: 2,
			expectHostErr: true,
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
			name:    "no workers respond",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				mockError: errors.New("unused"),
				errorMode: errorOnTimeout,
			},
			expectError:   true,
			errorContains: "no workers responded",
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
				"jobs.modify._all",
				tt.opts,
			)

			result, err := jobsClient.ModifyNetworkDNSAll(
				s.ctx,
				[]string{"8.8.8.8"},
				[]string{"example.com"},
				"eth0",
			)

			if tt.expectError {
				s.Error(err)
				s.Nil(result)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
				s.Len(result, tt.expectedCount)
				if tt.expectHostErr {
					hasErr := false
					for _, hostErr := range result {
						if hostErr != nil {
							hasErr = true
							break
						}
					}
					s.True(hasErr)
				}
			}
		})
	}
}

func TestModifyPublicTestSuite(t *testing.T) {
	suite.Run(t, new(ModifyPublicTestSuite))
}

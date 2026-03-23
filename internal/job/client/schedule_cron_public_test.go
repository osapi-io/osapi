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
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/retr0h/osapi/internal/job/client"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
	"github.com/retr0h/osapi/internal/provider/scheduled/cron"
)

type ScheduleCronPublicTestSuite struct {
	suite.Suite

	mockCtrl       *gomock.Controller
	mockNATSClient *jobmocks.MockNATSClient
	mockKV         *jobmocks.MockKeyValue
	jobsClient     *client.Client
	ctx            context.Context
}

func (s *ScheduleCronPublicTestSuite) SetupTest() {
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

	s.jobsClient.SetMeterProvider(sdkmetric.NewMeterProvider())
}

func (s *ScheduleCronPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *ScheduleCronPublicTestSuite) TestQueryScheduleCronList() {
	tests := []struct {
		name          string
		target        string
		responseData  string
		mockError     error
		expectError   bool
		errorContains string
	}{
		{
			name:   "success",
			target: "server1",
			responseData: `{
				"status": "completed",
				"data": [{"name":"backup","schedule":"0 2 * * *","user":"root","command":"/usr/local/bin/backup.sh"}]
			}`,
			expectError: false,
		},
		{
			name:   "job failed",
			target: "server1",
			responseData: `{
				"status": "failed",
				"error": "permission denied"
			}`,
			expectError:   true,
			errorContains: "job failed: permission denied",
		},
		{
			name:          "publish error",
			target:        "server1",
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
				"jobs.query.host.server1",
				tt.responseData,
				tt.mockError,
			)

			resp, err := s.jobsClient.QueryScheduleCronList(
				s.ctx,
				tt.target,
			)

			if tt.expectError {
				s.Error(err)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
				s.NotNil(resp)
				s.Equal("completed", string(resp.Status))
			}
		})
	}
}

func (s *ScheduleCronPublicTestSuite) TestQueryScheduleCronListBroadcast() {
	tests := []struct {
		name          string
		timeout       time.Duration
		opts          *publishAndCollectMockOpts
		expectError   bool
		errorContains string
		expectedCount int
	}{
		{
			name:    "multiple agents respond",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":[{"name":"backup","schedule":"0 2 * * *","user":"root","command":"/usr/bin/backup.sh"}]}`,
					`{"status":"completed","hostname":"server2","data":[{"name":"cleanup","schedule":"0 3 * * *","user":"root","command":"/usr/bin/cleanup.sh"}]}`,
				},
			},
			expectedCount: 2,
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
			jobsClient.SetMeterProvider(sdkmetric.NewMeterProvider())

			setupPublishAndCollectMocks(
				s.mockCtrl,
				s.mockKV,
				s.mockNATSClient,
				"jobs.query._all",
				tt.opts,
			)

			_, result, err := jobsClient.QueryScheduleCronListBroadcast(s.ctx, "_all")

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

func (s *ScheduleCronPublicTestSuite) TestQueryScheduleCronGet() {
	tests := []struct {
		name          string
		target        string
		cronName      string
		responseData  string
		mockError     error
		expectError   bool
		errorContains string
	}{
		{
			name:     "success",
			target:   "server1",
			cronName: "backup",
			responseData: `{
				"status": "completed",
				"data": {"name":"backup","schedule":"0 2 * * *","user":"root","command":"/usr/local/bin/backup.sh"}
			}`,
			expectError: false,
		},
		{
			name:     "job failed",
			target:   "server1",
			cronName: "missing",
			responseData: `{
				"status": "failed",
				"error": "cron entry not found"
			}`,
			expectError:   true,
			errorContains: "job failed: cron entry not found",
		},
		{
			name:          "publish error",
			target:        "server1",
			cronName:      "backup",
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
				"jobs.query.host.server1",
				tt.responseData,
				tt.mockError,
			)

			resp, err := s.jobsClient.QueryScheduleCronGet(
				s.ctx,
				tt.target,
				tt.cronName,
			)

			if tt.expectError {
				s.Error(err)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
				s.NotNil(resp)
				s.Equal("completed", string(resp.Status))
			}
		})
	}
}

func (s *ScheduleCronPublicTestSuite) TestModifyScheduleCronCreate() {
	tests := []struct {
		name          string
		target        string
		entry         cron.Entry
		responseData  string
		mockError     error
		expectError   bool
		errorContains string
	}{
		{
			name:   "success",
			target: "server1",
			entry: cron.Entry{
				Name:     "logrotate",
				Schedule: "0 0 * * *",
				User:     "root",
				Command:  "/usr/sbin/logrotate",
			},
			responseData: `{
				"status": "completed",
				"data": {"name":"logrotate","changed":true}
			}`,
			expectError: false,
		},
		{
			name:   "job failed",
			target: "server1",
			entry: cron.Entry{
				Name:     "dup",
				Schedule: "* * * * *",
				User:     "root",
				Command:  "echo",
			},
			responseData: `{
				"status": "failed",
				"error": "entry already exists"
			}`,
			expectError:   true,
			errorContains: "job failed: entry already exists",
		},
		{
			name:   "publish error",
			target: "server1",
			entry: cron.Entry{
				Name:     "logrotate",
				Schedule: "0 0 * * *",
				User:     "root",
				Command:  "/usr/sbin/logrotate",
			},
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
				"jobs.modify.host.server1",
				tt.responseData,
				tt.mockError,
			)

			resp, err := s.jobsClient.ModifyScheduleCronCreate(
				s.ctx,
				tt.target,
				tt.entry,
			)

			if tt.expectError {
				s.Error(err)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
				s.NotNil(resp)
				s.Equal("completed", string(resp.Status))
			}
		})
	}
}

func (s *ScheduleCronPublicTestSuite) TestModifyScheduleCronUpdate() {
	tests := []struct {
		name          string
		target        string
		entry         cron.Entry
		responseData  string
		mockError     error
		expectError   bool
		errorContains string
	}{
		{
			name:   "success",
			target: "server1",
			entry: cron.Entry{
				Name:     "backup",
				Schedule: "0 3 * * *",
				User:     "root",
				Command:  "/usr/local/bin/backup.sh",
			},
			responseData: `{
				"status": "completed",
				"data": {"name":"backup","changed":true}
			}`,
			expectError: false,
		},
		{
			name:   "job failed",
			target: "server1",
			entry: cron.Entry{
				Name:     "missing",
				Schedule: "* * * * *",
				User:     "root",
				Command:  "echo",
			},
			responseData: `{
				"status": "failed",
				"error": "entry not found"
			}`,
			expectError:   true,
			errorContains: "job failed: entry not found",
		},
		{
			name:   "publish error",
			target: "server1",
			entry: cron.Entry{
				Name:     "backup",
				Schedule: "0 3 * * *",
				User:     "root",
				Command:  "/usr/local/bin/backup.sh",
			},
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
				"jobs.modify.host.server1",
				tt.responseData,
				tt.mockError,
			)

			resp, err := s.jobsClient.ModifyScheduleCronUpdate(
				s.ctx,
				tt.target,
				tt.entry,
			)

			if tt.expectError {
				s.Error(err)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
				s.NotNil(resp)
				s.Equal("completed", string(resp.Status))
			}
		})
	}
}

func (s *ScheduleCronPublicTestSuite) TestModifyScheduleCronDelete() {
	tests := []struct {
		name          string
		target        string
		cronName      string
		responseData  string
		mockError     error
		expectError   bool
		errorContains string
	}{
		{
			name:     "success",
			target:   "server1",
			cronName: "backup",
			responseData: `{
				"status": "completed",
				"data": {"name":"backup","changed":true}
			}`,
			expectError: false,
		},
		{
			name:     "job failed",
			target:   "server1",
			cronName: "missing",
			responseData: `{
				"status": "failed",
				"error": "entry not found"
			}`,
			expectError:   true,
			errorContains: "job failed: entry not found",
		},
		{
			name:          "publish error",
			target:        "server1",
			cronName:      "backup",
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
				"jobs.modify.host.server1",
				tt.responseData,
				tt.mockError,
			)

			resp, err := s.jobsClient.ModifyScheduleCronDelete(
				s.ctx,
				tt.target,
				tt.cronName,
			)

			if tt.expectError {
				s.Error(err)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
				s.NotNil(resp)
				s.Equal("completed", string(resp.Status))
			}
		})
	}
}

func TestScheduleCronPublicTestSuite(t *testing.T) {
	suite.Run(t, new(ScheduleCronPublicTestSuite))
}

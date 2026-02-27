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

	"github.com/retr0h/osapi/internal/job/client"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
)

type ModifyCommandPublicTestSuite struct {
	suite.Suite

	mockCtrl       *gomock.Controller
	mockNATSClient *jobmocks.MockNATSClient
	mockKV         *jobmocks.MockKeyValue
	jobsClient     *client.Client
	ctx            context.Context
}

func (s *ModifyCommandPublicTestSuite) SetupTest() {
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

func (s *ModifyCommandPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *ModifyCommandPublicTestSuite) TestModifyCommandExec() {
	tests := []struct {
		name          string
		hostname      string
		command       string
		args          []string
		cwd           string
		timeout       int
		responseData  string
		mockError     error
		expectError   bool
		errorContains string
	}{
		{
			name:     "success",
			hostname: "server1",
			command:  "ls",
			args:     []string{"-la"},
			cwd:      "/tmp",
			timeout:  30,
			responseData: `{
				"status": "completed",
				"data": {"stdout":"total 0\n","stderr":"","exit_code":0,"duration_ms":12}
			}`,
			expectError: false,
		},
		{
			name:     "job failed",
			hostname: "server1",
			command:  "bad-cmd",
			responseData: `{
				"status": "failed",
				"error": "command not found"
			}`,
			expectError:   true,
			errorContains: "job failed: command not found",
		},
		{
			name:          "publish error",
			hostname:      "server1",
			command:       "ls",
			mockError:     errors.New("connection failed"),
			expectError:   true,
			errorContains: "failed to publish and wait",
		},
		{
			name:     "unmarshal error",
			hostname: "server1",
			command:  "ls",
			responseData: `{
				"status": "completed",
				"data": "not valid json object"
			}`,
			expectError:   true,
			errorContains: "failed to unmarshal command result",
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

			_, result, _, err := s.jobsClient.ModifyCommandExec(
				s.ctx,
				tt.hostname,
				tt.command,
				tt.args,
				tt.cwd,
				tt.timeout,
			)

			if tt.expectError {
				s.Error(err)
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

func (s *ModifyCommandPublicTestSuite) TestModifyCommandShell() {
	tests := []struct {
		name          string
		hostname      string
		command       string
		cwd           string
		timeout       int
		responseData  string
		mockError     error
		expectError   bool
		errorContains string
	}{
		{
			name:     "success",
			hostname: "server1",
			command:  "echo hello | tr a-z A-Z",
			cwd:      "/tmp",
			timeout:  30,
			responseData: `{
				"status": "completed",
				"data": {"stdout":"HELLO\n","stderr":"","exit_code":0,"duration_ms":5}
			}`,
			expectError: false,
		},
		{
			name:     "job failed",
			hostname: "server1",
			command:  "bad command",
			responseData: `{
				"status": "failed",
				"error": "shell error"
			}`,
			expectError:   true,
			errorContains: "job failed: shell error",
		},
		{
			name:          "publish error",
			hostname:      "server1",
			command:       "echo",
			mockError:     errors.New("connection failed"),
			expectError:   true,
			errorContains: "failed to publish and wait",
		},
		{
			name:     "unmarshal error",
			hostname: "server1",
			command:  "echo",
			responseData: `{
				"status": "completed",
				"data": "not valid json object"
			}`,
			expectError:   true,
			errorContains: "failed to unmarshal command result",
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

			_, result, _, err := s.jobsClient.ModifyCommandShell(
				s.ctx,
				tt.hostname,
				tt.command,
				tt.cwd,
				tt.timeout,
			)

			if tt.expectError {
				s.Error(err)
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

func (s *ModifyCommandPublicTestSuite) TestModifyCommandExecBroadcast() {
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
					`{"status":"completed","hostname":"server1","data":{"stdout":"ok\n","stderr":"","exit_code":0,"duration_ms":10}}`,
					`{"status":"completed","hostname":"server2","data":{"stdout":"ok\n","stderr":"","exit_code":0,"duration_ms":15}}`,
				},
			},
			expectedCount: 2,
		},
		{
			name:    "partial failure",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":{"stdout":"ok\n","stderr":"","exit_code":0,"duration_ms":10}}`,
					`{"status":"failed","hostname":"server2","error":"command not found"}`,
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
			name:    "unmarshal error in broadcast response",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":"invalid json"}`,
				},
			},
			expectedCount: 1,
			expectHostErr: true,
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

			_, results, errs, err := jobsClient.ModifyCommandExecBroadcast(
				s.ctx,
				"_all",
				"ls",
				[]string{"-la"},
				"/tmp",
				30,
			)

			if tt.expectError {
				s.Error(err)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
				totalCount := len(results) + len(errs)
				s.Equal(tt.expectedCount, totalCount)
				if tt.expectHostErr {
					s.NotEmpty(errs)
				}
			}
		})
	}
}

func (s *ModifyCommandPublicTestSuite) TestModifyCommandShellBroadcast() {
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
					`{"status":"completed","hostname":"server1","data":{"stdout":"HELLO\n","stderr":"","exit_code":0,"duration_ms":8}}`,
					`{"status":"completed","hostname":"server2","data":{"stdout":"HELLO\n","stderr":"","exit_code":0,"duration_ms":12}}`,
				},
			},
			expectedCount: 2,
		},
		{
			name:    "partial failure",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":{"stdout":"ok\n","stderr":"","exit_code":0,"duration_ms":5}}`,
					`{"status":"failed","hostname":"server2","error":"shell error"}`,
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
			name:    "unmarshal error in broadcast response",
			timeout: 50 * time.Millisecond,
			opts: &publishAndCollectMockOpts{
				responseEntries: []string{
					`{"status":"completed","hostname":"server1","data":"invalid json"}`,
				},
			},
			expectedCount: 1,
			expectHostErr: true,
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

			_, results, errs, err := jobsClient.ModifyCommandShellBroadcast(
				s.ctx,
				"_all",
				"echo hello | tr a-z A-Z",
				"/tmp",
				30,
			)

			if tt.expectError {
				s.Error(err)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
			} else {
				s.NoError(err)
				totalCount := len(results) + len(errs)
				s.Equal(tt.expectedCount, totalCount)
				if tt.expectHostErr {
					s.NotEmpty(errs)
				}
			}
		})
	}
}

func TestModifyCommandPublicTestSuite(t *testing.T) {
	suite.Run(t, new(ModifyCommandPublicTestSuite))
}

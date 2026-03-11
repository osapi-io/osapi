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

	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/job/client"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
)

type ModifyContainerPublicTestSuite struct {
	suite.Suite

	mockCtrl       *gomock.Controller
	mockNATSClient *jobmocks.MockNATSClient
	mockKV         *jobmocks.MockKeyValue
	jobsClient     *client.Client
	ctx            context.Context
}

func (s *ModifyContainerPublicTestSuite) SetupTest() {
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

func (s *ModifyContainerPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *ModifyContainerPublicTestSuite) TestModifyContainerCreate() {
	tests := []struct {
		name          string
		target        string
		data          *job.ContainerCreateData
		responseData  string
		mockError     error
		expectError   bool
		errorContains string
	}{
		{
			name:   "success",
			target: "server1",
			data: &job.ContainerCreateData{
				Image: "nginx:latest",
				Name:  "test-nginx",
			},
			responseData: `{
				"status": "completed",
				"data": {"id":"abc123","name":"test-nginx","image":"nginx:latest","state":"created","created":"2026-03-11T10:00:00Z"}
			}`,
			expectError: false,
		},
		{
			name:   "job failed",
			target: "server1",
			data: &job.ContainerCreateData{
				Image: "invalid:image",
			},
			responseData: `{
				"status": "failed",
				"error": "image not found"
			}`,
			expectError:   true,
			errorContains: "job failed: image not found",
		},
		{
			name:   "publish error",
			target: "server1",
			data: &job.ContainerCreateData{
				Image: "nginx:latest",
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

			resp, err := s.jobsClient.ModifyContainerCreate(
				s.ctx,
				tt.target,
				tt.data,
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

func (s *ModifyContainerPublicTestSuite) TestModifyContainerStart() {
	tests := []struct {
		name          string
		target        string
		id            string
		responseData  string
		mockError     error
		expectError   bool
		errorContains string
	}{
		{
			name:   "success",
			target: "server1",
			id:     "abc123",
			responseData: `{
				"status": "completed",
				"data": {"message":"Container started successfully"}
			}`,
			expectError: false,
		},
		{
			name:   "job failed",
			target: "server1",
			id:     "abc123",
			responseData: `{
				"status": "failed",
				"error": "container not found"
			}`,
			expectError:   true,
			errorContains: "job failed: container not found",
		},
		{
			name:          "publish error",
			target:        "server1",
			id:            "abc123",
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

			resp, err := s.jobsClient.ModifyContainerStart(
				s.ctx,
				tt.target,
				tt.id,
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

func (s *ModifyContainerPublicTestSuite) TestModifyContainerStop() {
	timeout := 10
	tests := []struct {
		name          string
		target        string
		id            string
		data          *job.ContainerStopData
		responseData  string
		mockError     error
		expectError   bool
		errorContains string
	}{
		{
			name:   "success with timeout",
			target: "server1",
			id:     "abc123",
			data:   &job.ContainerStopData{Timeout: &timeout},
			responseData: `{
				"status": "completed",
				"data": {"message":"Container stopped successfully"}
			}`,
			expectError: false,
		},
		{
			name:   "success without timeout",
			target: "server1",
			id:     "abc123",
			data:   &job.ContainerStopData{},
			responseData: `{
				"status": "completed",
				"data": {"message":"Container stopped successfully"}
			}`,
			expectError: false,
		},
		{
			name:   "job failed",
			target: "server1",
			id:     "abc123",
			data:   &job.ContainerStopData{},
			responseData: `{
				"status": "failed",
				"error": "container not running"
			}`,
			expectError:   true,
			errorContains: "job failed: container not running",
		},
		{
			name:          "publish error",
			target:        "server1",
			id:            "abc123",
			data:          &job.ContainerStopData{},
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

			resp, err := s.jobsClient.ModifyContainerStop(
				s.ctx,
				tt.target,
				tt.id,
				tt.data,
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

func (s *ModifyContainerPublicTestSuite) TestModifyContainerRemove() {
	tests := []struct {
		name          string
		target        string
		id            string
		data          *job.ContainerRemoveData
		responseData  string
		mockError     error
		expectError   bool
		errorContains string
	}{
		{
			name:   "success with force",
			target: "server1",
			id:     "abc123",
			data:   &job.ContainerRemoveData{Force: true},
			responseData: `{
				"status": "completed",
				"data": {"message":"Container removed successfully"}
			}`,
			expectError: false,
		},
		{
			name:   "success without force",
			target: "server1",
			id:     "abc123",
			data:   &job.ContainerRemoveData{Force: false},
			responseData: `{
				"status": "completed",
				"data": {"message":"Container removed successfully"}
			}`,
			expectError: false,
		},
		{
			name:   "job failed",
			target: "server1",
			id:     "abc123",
			data:   &job.ContainerRemoveData{},
			responseData: `{
				"status": "failed",
				"error": "container is running"
			}`,
			expectError:   true,
			errorContains: "job failed: container is running",
		},
		{
			name:          "publish error",
			target:        "server1",
			id:            "abc123",
			data:          &job.ContainerRemoveData{},
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

			resp, err := s.jobsClient.ModifyContainerRemove(
				s.ctx,
				tt.target,
				tt.id,
				tt.data,
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

func (s *ModifyContainerPublicTestSuite) TestQueryContainerList() {
	tests := []struct {
		name          string
		target        string
		data          *job.ContainerListData
		responseData  string
		mockError     error
		expectError   bool
		errorContains string
	}{
		{
			name:   "success",
			target: "server1",
			data: &job.ContainerListData{
				State: "running",
				Limit: 10,
			},
			responseData: `{
				"status": "completed",
				"data": [{"id":"abc123","name":"web","state":"running"}]
			}`,
			expectError: false,
		},
		{
			name:   "job failed",
			target: "server1",
			data:   &job.ContainerListData{},
			responseData: `{
				"status": "failed",
				"error": "runtime not available"
			}`,
			expectError:   true,
			errorContains: "job failed: runtime not available",
		},
		{
			name:          "publish error",
			target:        "server1",
			data:          &job.ContainerListData{},
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

			resp, err := s.jobsClient.QueryContainerList(
				s.ctx,
				tt.target,
				tt.data,
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

func (s *ModifyContainerPublicTestSuite) TestQueryContainerInspect() {
	tests := []struct {
		name          string
		target        string
		id            string
		responseData  string
		mockError     error
		expectError   bool
		errorContains string
	}{
		{
			name:   "success",
			target: "server1",
			id:     "abc123",
			responseData: `{
				"status": "completed",
				"data": {"id":"abc123","name":"web","state":"running"}
			}`,
			expectError: false,
		},
		{
			name:   "job failed",
			target: "server1",
			id:     "abc123",
			responseData: `{
				"status": "failed",
				"error": "container not found"
			}`,
			expectError:   true,
			errorContains: "job failed: container not found",
		},
		{
			name:          "publish error",
			target:        "server1",
			id:            "abc123",
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

			resp, err := s.jobsClient.QueryContainerInspect(
				s.ctx,
				tt.target,
				tt.id,
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

func (s *ModifyContainerPublicTestSuite) TestModifyContainerExec() {
	tests := []struct {
		name          string
		target        string
		id            string
		data          *job.ContainerExecData
		responseData  string
		mockError     error
		expectError   bool
		errorContains string
	}{
		{
			name:   "success",
			target: "server1",
			id:     "abc123",
			data: &job.ContainerExecData{
				Command: []string{"ls", "-la"},
			},
			responseData: `{
				"status": "completed",
				"data": {"stdout":"output","stderr":"","exit_code":0}
			}`,
			expectError: false,
		},
		{
			name:   "job failed",
			target: "server1",
			id:     "abc123",
			data: &job.ContainerExecData{
				Command: []string{"bad-cmd"},
			},
			responseData: `{
				"status": "failed",
				"error": "command not found"
			}`,
			expectError:   true,
			errorContains: "job failed: command not found",
		},
		{
			name:   "publish error",
			target: "server1",
			id:     "abc123",
			data: &job.ContainerExecData{
				Command: []string{"ls"},
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

			resp, err := s.jobsClient.ModifyContainerExec(
				s.ctx,
				tt.target,
				tt.id,
				tt.data,
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

func (s *ModifyContainerPublicTestSuite) TestModifyContainerPull() {
	tests := []struct {
		name          string
		target        string
		data          *job.ContainerPullData
		responseData  string
		mockError     error
		expectError   bool
		errorContains string
	}{
		{
			name:   "success",
			target: "server1",
			data: &job.ContainerPullData{
				Image: "nginx:latest",
			},
			responseData: `{
				"status": "completed",
				"data": {"image_id":"sha256:abc","tag":"latest","size":2048}
			}`,
			expectError: false,
		},
		{
			name:   "job failed",
			target: "server1",
			data: &job.ContainerPullData{
				Image: "invalid:image",
			},
			responseData: `{
				"status": "failed",
				"error": "image not found"
			}`,
			expectError:   true,
			errorContains: "job failed: image not found",
		},
		{
			name:   "publish error",
			target: "server1",
			data: &job.ContainerPullData{
				Image: "nginx:latest",
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

			resp, err := s.jobsClient.ModifyContainerPull(
				s.ctx,
				tt.target,
				tt.data,
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

func TestModifyContainerPublicTestSuite(t *testing.T) {
	suite.Run(t, new(ModifyContainerPublicTestSuite))
}

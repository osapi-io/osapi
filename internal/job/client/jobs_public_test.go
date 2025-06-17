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

type JobsPublicTestSuite struct {
	suite.Suite

	mockCtrl       *gomock.Controller
	mockNATSClient *jobmocks.MockNATSClient
	mockKV         *jobmocks.MockKeyValue
	jobsClient     *client.Client
	ctx            context.Context
}

func (s *JobsPublicTestSuite) SetupTest() {
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

func (s *JobsPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *JobsPublicTestSuite) TestCreateJob() {
	tests := []struct {
		name          string
		operationData map[string]interface{}
		targetHost    string
		expectedErr   string
		setupMocks    func()
	}{
		{
			name: "successful job creation",
			operationData: map[string]interface{}{
				"type": "system.hostname.get",
				"data": map[string]string{"param": "value"},
			},
			targetHost: "server1",
			setupMocks: func() {
				// Mock KV operations for job storage
				s.mockKV.EXPECT().Put(gomock.Any(), gomock.Any()).Return(uint64(1), nil).Times(2)
				s.mockKV.EXPECT().Bucket().Return("test-bucket").AnyTimes()
				// Mock the publish notification
				s.mockNATSClient.EXPECT().
					Publish(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
		},
		{
			name: "missing operation type",
			operationData: map[string]interface{}{
				"data": map[string]string{"param": "value"},
			},
			targetHost:  "server1",
			expectedErr: "invalid operation format: missing type field",
			setupMocks:  func() {},
		},
		{
			name: "KV storage error",
			operationData: map[string]interface{}{
				"type": "system.hostname.get",
				"data": map[string]string{"param": "value"},
			},
			targetHost:  "server1",
			expectedErr: "failed to store job in KV",
			setupMocks: func() {
				s.mockKV.EXPECT().
					Put(gomock.Any(), gomock.Any()).
					Return(uint64(0), errors.New("kv error"))
				s.mockKV.EXPECT().Bucket().Return("test-bucket").AnyTimes()
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMocks()

			result, err := s.jobsClient.CreateJob(s.ctx, tt.operationData, tt.targetHost)

			if tt.expectedErr != "" {
				s.Error(err)
				s.Contains(err.Error(), tt.expectedErr)
			} else {
				s.NoError(err)
				s.NotEmpty(result.JobID)
				s.Equal(uint64(1), result.Revision)
				s.Equal("created", result.Status)
			}
		})
	}
}

func (s *JobsPublicTestSuite) TestGetQueueStats() {
	// Mock KV Keys returning job keys
	keys := []string{"jobs.job-1", "jobs.job-2"}
	s.mockKV.EXPECT().Keys().Return(keys, nil)

	// Mock Get calls for job data
	mockEntry1 := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
	mockEntry1.EXPECT().
		Value().
		Return([]byte(`{"id":"job-1","operation":{"type":"system.hostname.get"}}`)).
		AnyTimes()
	s.mockKV.EXPECT().Get("jobs.job-1").Return(mockEntry1, nil).AnyTimes()

	mockEntry2 := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
	mockEntry2.EXPECT().
		Value().
		Return([]byte(`{"id":"job-2","operation":{"type":"system.status.get"}}`)).
		AnyTimes()
	s.mockKV.EXPECT().Get("jobs.job-2").Return(mockEntry2, nil).AnyTimes()

	// Mock Keys calls for status events (called twice)
	s.mockKV.EXPECT().Keys().Return([]string{}, nil).Times(2)

	// Mock GetStreamInfo for DLQ count
	s.mockNATSClient.EXPECT().
		GetStreamInfo(gomock.Any(), "JOBS-DLQ").
		Return(nil, errors.New("stream not found"))

	stats, err := s.jobsClient.GetQueueStats(s.ctx)
	s.NoError(err)
	s.NotNil(stats)
	s.GreaterOrEqual(stats.TotalJobs, 0)
}

func (s *JobsPublicTestSuite) TestGetJobStatus() {
	tests := []struct {
		name        string
		jobID       string
		expectedErr string
		setupMocks  func()
	}{
		{
			name:  "successful job status retrieval",
			jobID: "job-123",
			setupMocks: func() {
				mockEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return([]byte(`{
					"id": "job-123",
					"status": "completed",
					"created": "2024-01-01T00:00:00Z",
					"subject": "jobs.query.server1",
					"operation": {"type": "system.hostname.get"}
				}`)).AnyTimes()
				s.mockKV.EXPECT().Get("jobs.job-123").Return(mockEntry, nil)
				s.mockKV.EXPECT().Keys().Return([]string{}, nil).AnyTimes()
			},
		},
		{
			name:        "job not found",
			jobID:       "nonexistent",
			expectedErr: "job not found: nonexistent",
			setupMocks: func() {
				s.mockKV.EXPECT().Get("jobs.nonexistent").Return(nil, errors.New("key not found"))
			},
		},
		{
			name:        "invalid job data",
			jobID:       "job-invalid",
			expectedErr: "failed to parse job data",
			setupMocks: func() {
				mockEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return([]byte(`invalid json`))
				s.mockKV.EXPECT().Get("jobs.job-invalid").Return(mockEntry, nil)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMocks()

			jobStatus, err := s.jobsClient.GetJobStatus(s.ctx, tt.jobID)

			if tt.expectedErr != "" {
				s.Error(err)
				s.Contains(err.Error(), tt.expectedErr)
				s.Nil(jobStatus)
			} else {
				s.NoError(err)
				s.NotNil(jobStatus)
				s.Equal(tt.jobID, jobStatus.ID)
			}
		})
	}
}

func (s *JobsPublicTestSuite) TestListJobs() {
	tests := []struct {
		name         string
		statusFilter string
		expectedErr  string
		setupMocks   func()
		expectedJobs int
	}{
		{
			name:         "no jobs found",
			statusFilter: "",
			setupMocks: func() {
				s.mockKV.EXPECT().Keys().Return(nil, errors.New("nats: no keys found"))
			},
			expectedJobs: 0,
		},
		{
			name:         "kv error",
			statusFilter: "",
			expectedErr:  "error fetching jobs",
			setupMocks: func() {
				s.mockKV.EXPECT().Keys().Return(nil, errors.New("connection failed"))
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMocks()

			jobs, err := s.jobsClient.ListJobs(s.ctx, tt.statusFilter)

			if tt.expectedErr != "" {
				s.Error(err)
				s.Contains(err.Error(), tt.expectedErr)
			} else {
				s.NoError(err)
				s.Len(jobs, tt.expectedJobs)
			}
		})
	}
}

func TestJobsPublicTestSuite(t *testing.T) {
	suite.Run(t, new(JobsPublicTestSuite))
}

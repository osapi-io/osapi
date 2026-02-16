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
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/nats-io/nats.go"
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

func (s *JobsPublicTestSuite) TestNew() {
	tests := []struct {
		name        string
		opts        *client.Options
		expectedErr string
	}{
		{
			name:        "nil options",
			opts:        nil,
			expectedErr: "options cannot be nil",
		},
		{
			name:        "nil KV bucket",
			opts:        &client.Options{},
			expectedErr: "KVBucket cannot be nil",
		},
		{
			name: "valid options",
			opts: &client.Options{
				KVBucket: s.mockKV,
				Timeout:  30 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			c, err := client.New(slog.Default(), s.mockNATSClient, tt.opts)

			if tt.expectedErr != "" {
				s.Error(err)
				s.Contains(err.Error(), tt.expectedErr)
				s.Nil(c)
			} else {
				s.NoError(err)
				s.NotNil(c)
			}
		})
	}
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
				s.mockKV.EXPECT().Put(gomock.Any(), gomock.Any()).Return(uint64(1), nil).Times(2)
				s.mockKV.EXPECT().Bucket().Return("test-bucket").AnyTimes()
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
		{
			name: "modify operation routes to modify subject",
			operationData: map[string]interface{}{
				"type": "network.dns.update",
				"data": map[string]string{"interface": "eth0"},
			},
			targetHost: "server1",
			setupMocks: func() {
				s.mockKV.EXPECT().Put(gomock.Any(), gomock.Any()).Return(uint64(1), nil).Times(2)
				s.mockKV.EXPECT().Bucket().Return("test-bucket").AnyTimes()
				s.mockNATSClient.EXPECT().
					Publish(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
		},
		{
			name: "empty target hostname defaults to any",
			operationData: map[string]interface{}{
				"type": "system.hostname.get",
				"data": map[string]string{},
			},
			targetHost: "",
			setupMocks: func() {
				s.mockKV.EXPECT().Put(gomock.Any(), gomock.Any()).Return(uint64(1), nil).Times(2)
				s.mockKV.EXPECT().Bucket().Return("test-bucket").AnyTimes()
				s.mockNATSClient.EXPECT().
					Publish(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
		},
		{
			name: "publish notification error",
			operationData: map[string]interface{}{
				"type": "system.hostname.get",
				"data": map[string]string{},
			},
			targetHost:  "server1",
			expectedErr: "failed to send notification",
			setupMocks: func() {
				s.mockKV.EXPECT().Put(gomock.Any(), gomock.Any()).Return(uint64(1), nil).Times(2)
				s.mockKV.EXPECT().Bucket().Return("test-bucket").AnyTimes()
				s.mockNATSClient.EXPECT().
					Publish(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("publish failed"))
			},
		},
		{
			name: "status event put error is logged not returned",
			operationData: map[string]interface{}{
				"type": "system.hostname.get",
				"data": map[string]string{},
			},
			targetHost: "server1",
			setupMocks: func() {
				s.mockKV.EXPECT().Put(gomock.Any(), gomock.Any()).Return(uint64(1), nil)
				s.mockKV.EXPECT().
					Put(gomock.Any(), gomock.Any()).
					Return(uint64(0), errors.New("status put failed"))
				s.mockKV.EXPECT().Bucket().Return("test-bucket").AnyTimes()
				s.mockNATSClient.EXPECT().
					Publish(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
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
				s.Equal("created", result.Status)
			}
		})
	}
}

func (s *JobsPublicTestSuite) TestGetJobStatus() {
	now := time.Now().Format(time.RFC3339)

	tests := []struct {
		name           string
		jobID          string
		expectedErr    string
		expectedStatus string
		expectedError  string
		workerCount    int
		responseCount  int
		setupMocks     func()
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
				}`))
				s.mockKV.EXPECT().Get("jobs.job-123").Return(mockEntry, nil)
				s.mockKV.EXPECT().Keys().Return([]string{}, nil)
			},
			expectedStatus: "submitted",
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
		{
			name:  "completed worker with response",
			jobID: "job-1",
			setupMocks: func() {
				mockJobEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockJobEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-1","operation":{"type":"system.hostname.get"},"created":"2024-01-01T00:00:00Z","subject":"jobs.query.server1"}`,
				))
				s.mockKV.EXPECT().Get("jobs.job-1").Return(mockJobEntry, nil)

				s.mockKV.EXPECT().Keys().Return([]string{
					"status.job-1.acknowledged.worker1.100",
					"status.job-1.started.worker1.200",
					"status.job-1.completed.worker1.300",
					"responses.job-1.worker1.400",
				}, nil)

				ackEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				ackEntry.EXPECT().Value().Return([]byte(fmt.Sprintf(
					`{"job_id":"job-1","event":"acknowledged","hostname":"worker1","timestamp":"%s"}`,
					now,
				)))
				s.mockKV.EXPECT().Get("status.job-1.acknowledged.worker1.100").Return(ackEntry, nil)

				startEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				startEntry.EXPECT().Value().Return([]byte(fmt.Sprintf(
					`{"job_id":"job-1","event":"started","hostname":"worker1","timestamp":"%s"}`,
					now,
				)))
				s.mockKV.EXPECT().Get("status.job-1.started.worker1.200").Return(startEntry, nil)

				compEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				compEntry.EXPECT().Value().Return([]byte(fmt.Sprintf(
					`{"job_id":"job-1","event":"completed","hostname":"worker1","timestamp":"%s"}`,
					now,
				)))
				s.mockKV.EXPECT().Get("status.job-1.completed.worker1.300").Return(compEntry, nil)

				respEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				respEntry.EXPECT().Value().Return([]byte(
					`{"status":"completed","data":"eyJ0ZXN0IjogdHJ1ZX0="}`,
				))
				s.mockKV.EXPECT().Get("responses.job-1.worker1.400").Return(respEntry, nil)
			},
			expectedStatus: "completed",
			workerCount:    1,
			responseCount:  1,
		},
		{
			name:  "failed worker with error",
			jobID: "job-1",
			setupMocks: func() {
				mockJobEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockJobEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-1","operation":{"type":"system.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get("jobs.job-1").Return(mockJobEntry, nil)

				s.mockKV.EXPECT().Keys().Return([]string{
					"status.job-1.started.worker1.100",
					"status.job-1.failed.worker1.200",
				}, nil)

				startEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				startEntry.EXPECT().Value().Return([]byte(fmt.Sprintf(
					`{"job_id":"job-1","event":"started","hostname":"worker1","timestamp":"%s"}`,
					now,
				)))
				s.mockKV.EXPECT().Get("status.job-1.started.worker1.100").Return(startEntry, nil)

				failEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				failEntry.EXPECT().Value().Return([]byte(fmt.Sprintf(
					`{"job_id":"job-1","event":"failed","hostname":"worker1","timestamp":"%s","data":{"error":"disk full"}}`,
					now,
				)))
				s.mockKV.EXPECT().Get("status.job-1.failed.worker1.200").Return(failEntry, nil)
			},
			expectedStatus: "failed",
			expectedError:  "disk full",
			workerCount:    1,
		},
		{
			name:  "partial failure with multiple workers",
			jobID: "job-1",
			setupMocks: func() {
				mockJobEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockJobEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-1","operation":{"type":"system.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get("jobs.job-1").Return(mockJobEntry, nil)

				s.mockKV.EXPECT().Keys().Return([]string{
					"status.job-1.completed.worker1.100",
					"status.job-1.failed.worker2.200",
				}, nil)

				compEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				compEntry.EXPECT().Value().Return([]byte(fmt.Sprintf(
					`{"job_id":"job-1","event":"completed","hostname":"worker1","timestamp":"%s"}`,
					now,
				)))
				s.mockKV.EXPECT().Get("status.job-1.completed.worker1.100").Return(compEntry, nil)

				failEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				failEntry.EXPECT().Value().Return([]byte(fmt.Sprintf(
					`{"job_id":"job-1","event":"failed","hostname":"worker2","timestamp":"%s","data":{"error":"timeout"}}`,
					now,
				)))
				s.mockKV.EXPECT().Get("status.job-1.failed.worker2.200").Return(failEntry, nil)
			},
			expectedStatus: "partial_failure",
			workerCount:    2,
		},
		{
			name:  "processing state",
			jobID: "job-1",
			setupMocks: func() {
				mockJobEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockJobEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-1","operation":{"type":"system.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get("jobs.job-1").Return(mockJobEntry, nil)

				s.mockKV.EXPECT().Keys().Return([]string{
					"status.job-1.acknowledged.worker1.100",
					"status.job-1.started.worker1.200",
				}, nil)

				ackEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				ackEntry.EXPECT().Value().Return([]byte(fmt.Sprintf(
					`{"job_id":"job-1","event":"acknowledged","hostname":"worker1","timestamp":"%s"}`,
					now,
				)))
				s.mockKV.EXPECT().Get("status.job-1.acknowledged.worker1.100").Return(ackEntry, nil)

				startEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				startEntry.EXPECT().Value().Return([]byte(fmt.Sprintf(
					`{"job_id":"job-1","event":"started","hostname":"worker1","timestamp":"%s"}`,
					now,
				)))
				s.mockKV.EXPECT().Get("status.job-1.started.worker1.200").Return(startEntry, nil)
			},
			expectedStatus: "processing",
			workerCount:    1,
		},
		{
			name:  "event get error skipped gracefully",
			jobID: "job-1",
			setupMocks: func() {
				mockJobEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockJobEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-1","operation":{"type":"system.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get("jobs.job-1").Return(mockJobEntry, nil)

				s.mockKV.EXPECT().Keys().Return([]string{
					"status.job-1.completed.worker1.100",
				}, nil)

				s.mockKV.EXPECT().
					Get("status.job-1.completed.worker1.100").
					Return(nil, errors.New("kv error"))
			},
			expectedStatus: "submitted",
		},
		{
			name:  "invalid event JSON skipped gracefully",
			jobID: "job-1",
			setupMocks: func() {
				mockJobEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockJobEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-1","operation":{"type":"system.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get("jobs.job-1").Return(mockJobEntry, nil)

				s.mockKV.EXPECT().Keys().Return([]string{
					"status.job-1.completed.worker1.100",
				}, nil)

				invalidEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				invalidEntry.EXPECT().Value().Return([]byte(`invalid json`))
				s.mockKV.EXPECT().
					Get("status.job-1.completed.worker1.100").
					Return(invalidEntry, nil)
			},
			expectedStatus: "submitted",
		},
		{
			name:        "keys error returns error",
			jobID:       "job-1",
			expectedErr: "failed to get status events",
			setupMocks: func() {
				mockJobEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockJobEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-1","operation":{"type":"system.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get("jobs.job-1").Return(mockJobEntry, nil)

				s.mockKV.EXPECT().Keys().Return(nil, errors.New("connection failed"))
			},
		},
		{
			name:  "response parse error skipped",
			jobID: "job-1",
			setupMocks: func() {
				mockJobEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockJobEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-1","operation":{"type":"system.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get("jobs.job-1").Return(mockJobEntry, nil)

				s.mockKV.EXPECT().Keys().Return([]string{
					"responses.job-1.worker1.100",
				}, nil)

				respEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				respEntry.EXPECT().Value().Return([]byte(`not json`))
				s.mockKV.EXPECT().Get("responses.job-1.worker1.100").Return(respEntry, nil)
			},
			expectedStatus: "submitted",
			responseCount:  0,
		},
		{
			name:  "response get error skipped",
			jobID: "job-1",
			setupMocks: func() {
				mockJobEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockJobEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-1","operation":{"type":"system.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get("jobs.job-1").Return(mockJobEntry, nil)

				s.mockKV.EXPECT().Keys().Return([]string{
					"responses.job-1.worker1.100",
				}, nil)

				s.mockKV.EXPECT().
					Get("responses.job-1.worker1.100").
					Return(nil, errors.New("kv error"))
			},
			expectedStatus: "submitted",
			responseCount:  0,
		},
		{
			name:  "invalid timestamp skipped",
			jobID: "job-1",
			setupMocks: func() {
				mockJobEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockJobEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-1","operation":{"type":"system.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get("jobs.job-1").Return(mockJobEntry, nil)

				s.mockKV.EXPECT().Keys().Return([]string{
					"status.job-1.completed.worker1.100",
				}, nil)

				badEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				badEntry.EXPECT().Value().Return([]byte(
					`{"job_id":"job-1","event":"completed","hostname":"worker1","timestamp":"not-a-date"}`,
				))
				s.mockKV.EXPECT().Get("status.job-1.completed.worker1.100").Return(badEntry, nil)
			},
			expectedStatus: "submitted",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMocks()

			jobStatus, err := s.jobsClient.GetJobStatus(s.ctx, tt.jobID)

			if tt.expectedErr != "" {
				s.Error(err)
				s.Contains(err.Error(), tt.expectedErr)
			} else {
				s.NoError(err)
				s.NotNil(jobStatus)
				if tt.expectedStatus != "" {
					s.Equal(tt.expectedStatus, jobStatus.Status)
				}
				if tt.expectedError != "" {
					s.Equal(tt.expectedError, jobStatus.Error)
				}
				if tt.workerCount > 0 {
					s.Len(jobStatus.WorkerStates, tt.workerCount)
				}
				if tt.responseCount > 0 {
					s.Len(jobStatus.Responses, tt.responseCount)
				}
			}
		})
	}
}

func (s *JobsPublicTestSuite) TestGetQueueStats() {
	tests := []struct {
		name         string
		expectedErr  string
		setupMocks   func()
		expectedJobs int
		expectedDLQ  int
	}{
		{
			name: "no keys found",
			setupMocks: func() {
				s.mockKV.EXPECT().Keys().Return(nil, errors.New("nats: no keys found"))
			},
			expectedJobs: 0,
			expectedDLQ:  0,
		},
		{
			name:        "keys error",
			expectedErr: "error fetching jobs",
			setupMocks: func() {
				s.mockKV.EXPECT().Keys().Return(nil, errors.New("connection failed"))
			},
		},
		{
			name: "get job error skipped",
			setupMocks: func() {
				s.mockKV.EXPECT().Keys().Return([]string{"jobs.job-1"}, nil)
				s.mockKV.EXPECT().Get("jobs.job-1").Return(nil, errors.New("kv error"))
				s.mockNATSClient.EXPECT().
					GetStreamInfo(gomock.Any(), "JOBS-DLQ").
					Return(nil, errors.New("no stream"))
			},
			expectedJobs: 0,
		},
		{
			name: "invalid job JSON skipped",
			setupMocks: func() {
				s.mockKV.EXPECT().Keys().Return([]string{"jobs.job-1"}, nil)

				mockEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return([]byte(`not json`))
				s.mockKV.EXPECT().Get("jobs.job-1").Return(mockEntry, nil)

				s.mockNATSClient.EXPECT().
					GetStreamInfo(gomock.Any(), "JOBS-DLQ").
					Return(nil, errors.New("no stream"))
			},
			expectedJobs: 0,
		},
		{
			name: "with DLQ info",
			setupMocks: func() {
				s.mockKV.EXPECT().Keys().Return([]string{"jobs.job-1"}, nil)

				mockEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-1","operation":{"type":"system.hostname.get"}}`,
				))
				s.mockKV.EXPECT().Get("jobs.job-1").Return(mockEntry, nil)
				s.mockKV.EXPECT().Keys().Return([]string{}, nil)

				s.mockNATSClient.EXPECT().
					GetStreamInfo(gomock.Any(), "JOBS-DLQ").
					Return(&nats.StreamInfo{State: nats.StreamState{Msgs: 5}}, nil)
			},
			expectedJobs: 1,
			expectedDLQ:  5,
		},
		{
			name: "with jobs and DLQ error",
			setupMocks: func() {
				keys := []string{"jobs.job-1", "jobs.job-2"}
				s.mockKV.EXPECT().Keys().Return(keys, nil)

				mockEntry1 := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry1.EXPECT().
					Value().
					Return([]byte(`{"id":"job-1","operation":{"type":"system.hostname.get"}}`))
				s.mockKV.EXPECT().Get("jobs.job-1").Return(mockEntry1, nil)

				mockEntry2 := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry2.EXPECT().
					Value().
					Return([]byte(`{"id":"job-2","operation":{"type":"system.status.get"}}`))
				s.mockKV.EXPECT().Get("jobs.job-2").Return(mockEntry2, nil)

				s.mockKV.EXPECT().Keys().Return([]string{}, nil).Times(2)

				s.mockNATSClient.EXPECT().
					GetStreamInfo(gomock.Any(), "JOBS-DLQ").
					Return(nil, errors.New("stream not found"))
			},
			expectedJobs: 2,
			expectedDLQ:  0,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMocks()

			stats, err := s.jobsClient.GetQueueStats(s.ctx)

			if tt.expectedErr != "" {
				s.Error(err)
				s.Contains(err.Error(), tt.expectedErr)
			} else {
				s.NoError(err)
				s.NotNil(stats)
				s.Equal(tt.expectedJobs, stats.TotalJobs)
				if tt.expectedDLQ > 0 {
					s.Equal(tt.expectedDLQ, stats.DLQCount)
				}
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
		{
			name:         "returns all jobs",
			statusFilter: "",
			setupMocks: func() {
				s.mockKV.EXPECT().
					Keys().
					Return([]string{"jobs.job-1", "jobs.job-2"}, nil).
					Times(3)

				mockEntry1 := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry1.EXPECT().Value().Return([]byte(
					`{"id":"job-1","operation":{"type":"system.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get("jobs.job-1").Return(mockEntry1, nil)

				mockEntry2 := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry2.EXPECT().Value().Return([]byte(
					`{"id":"job-2","operation":{"type":"system.status.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get("jobs.job-2").Return(mockEntry2, nil)
			},
			expectedJobs: 2,
		},
		{
			name:         "filters out non matching status",
			statusFilter: "completed",
			setupMocks: func() {
				s.mockKV.EXPECT().Keys().Return([]string{"jobs.job-1"}, nil).Times(2)

				mockEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-1","operation":{"type":"system.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get("jobs.job-1").Return(mockEntry, nil)
			},
			expectedJobs: 0,
		},
		{
			name:         "only processes jobs prefix keys",
			statusFilter: "",
			setupMocks: func() {
				s.mockKV.EXPECT().Keys().Return(
					[]string{
						"status.job-1.submitted._api.123",
						"jobs.job-1",
						"responses.job-1.host.123",
					},
					nil,
				).Times(2)

				mockJobEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockJobEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-1","operation":{"type":"system.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get("jobs.job-1").Return(mockJobEntry, nil)

				mockStatusEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockStatusEntry.EXPECT().Value().Return([]byte(
					`{"job_id":"job-1","event":"submitted","hostname":"_api","timestamp":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().
					Get("status.job-1.submitted._api.123").
					Return(mockStatusEntry, nil)

				mockRespEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockRespEntry.EXPECT().Value().Return([]byte(
					`{"status":"completed","data":"{}"}`,
				))
				s.mockKV.EXPECT().
					Get("responses.job-1.host.123").
					Return(mockRespEntry, nil)
			},
			expectedJobs: 1,
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

func (s *JobsPublicTestSuite) TestDeleteJob() {
	tests := []struct {
		name        string
		jobID       string
		expectedErr string
		setupMocks  func()
	}{
		{
			name:  "successful deletion",
			jobID: "job-123",
			setupMocks: func() {
				mockEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return([]byte(`{"id":"job-123"}`)).AnyTimes()
				s.mockKV.EXPECT().Get("jobs.job-123").Return(mockEntry, nil)
				s.mockKV.EXPECT().Delete("jobs.job-123").Return(nil)
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
			name:        "delete error",
			jobID:       "job-456",
			expectedErr: "failed to delete job",
			setupMocks: func() {
				mockEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return([]byte(`{"id":"job-456"}`)).AnyTimes()
				s.mockKV.EXPECT().Get("jobs.job-456").Return(mockEntry, nil)
				s.mockKV.EXPECT().Delete("jobs.job-456").Return(errors.New("storage failure"))
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMocks()

			err := s.jobsClient.DeleteJob(s.ctx, tt.jobID)

			if tt.expectedErr != "" {
				s.Error(err)
				s.Contains(err.Error(), tt.expectedErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func TestJobsPublicTestSuite(t *testing.T) {
	suite.Run(t, new(JobsPublicTestSuite))
}

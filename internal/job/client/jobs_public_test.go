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
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/job"
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
		Timeout:    30 * time.Second,
		KVBucket:   s.mockKV,
		StreamName: "JOBS",
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
				"type": "node.hostname.get",
				"data": map[string]string{"param": "value"},
			},
			targetHost: "server1",
			setupMocks: func() {
				s.mockKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(1), nil).
					Times(2)
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
				"type": "node.hostname.get",
				"data": map[string]string{"param": "value"},
			},
			targetHost:  "server1",
			expectedErr: "failed to store job in KV",
			setupMocks: func() {
				s.mockKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
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
				s.mockKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(1), nil).
					Times(2)
				s.mockKV.EXPECT().Bucket().Return("test-bucket").AnyTimes()
				s.mockNATSClient.EXPECT().
					Publish(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
		},
		{
			name: "empty target hostname defaults to any",
			operationData: map[string]interface{}{
				"type": "node.hostname.get",
				"data": map[string]string{},
			},
			targetHost: "",
			setupMocks: func() {
				s.mockKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(1), nil).
					Times(2)
				s.mockKV.EXPECT().Bucket().Return("test-bucket").AnyTimes()
				s.mockNATSClient.EXPECT().
					Publish(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
		},
		{
			name: "publish notification error",
			operationData: map[string]interface{}{
				"type": "node.hostname.get",
				"data": map[string]string{},
			},
			targetHost:  "server1",
			expectedErr: "failed to send notification",
			setupMocks: func() {
				s.mockKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(1), nil).
					Times(2)
				s.mockKV.EXPECT().Bucket().Return("test-bucket").AnyTimes()
				s.mockNATSClient.EXPECT().
					Publish(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("publish failed"))
			},
		},
		{
			name: "unmarshalable operation data",
			operationData: map[string]interface{}{
				"type": "node.hostname.get",
				"data": make(chan int),
			},
			targetHost:  "server1",
			expectedErr: "failed to marshal job with status",
			setupMocks:  func() {},
		},
		{
			name: "status event put error is logged not returned",
			operationData: map[string]interface{}{
				"type": "node.hostname.get",
				"data": map[string]string{},
			},
			targetHost: "server1",
			setupMocks: func() {
				s.mockKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(1), nil)
				s.mockKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
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
		validateFunc   func(qj *job.QueuedJob)
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
					"operation": {"type": "node.hostname.get"}
				}`))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-123").Return(mockEntry, nil)
				s.mockKV.EXPECT().Keys(gomock.Any()).Return([]string{}, nil)
			},
			expectedStatus: "submitted",
		},
		{
			name:        "job not found",
			jobID:       "nonexistent",
			expectedErr: "job not found: nonexistent",
			setupMocks: func() {
				s.mockKV.EXPECT().
					Get(gomock.Any(), "jobs.nonexistent").
					Return(nil, errors.New("key not found"))
			},
		},
		{
			name:        "invalid job data",
			jobID:       "job-invalid",
			expectedErr: "failed to parse job data",
			setupMocks: func() {
				mockEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return([]byte(`invalid json`))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-invalid").Return(mockEntry, nil)
			},
		},
		{
			name:  "completed worker with response",
			jobID: "job-1",
			setupMocks: func() {
				mockJobEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockJobEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-1","operation":{"type":"node.hostname.get"},"created":"2024-01-01T00:00:00Z","subject":"jobs.query.server1"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-1").Return(mockJobEntry, nil)

				s.mockKV.EXPECT().Keys(gomock.Any()).Return([]string{
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
				s.mockKV.EXPECT().
					Get(gomock.Any(), "status.job-1.acknowledged.worker1.100").
					Return(ackEntry, nil)

				startEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				startEntry.EXPECT().Value().Return([]byte(fmt.Sprintf(
					`{"job_id":"job-1","event":"started","hostname":"worker1","timestamp":"%s"}`,
					now,
				)))
				s.mockKV.EXPECT().
					Get(gomock.Any(), "status.job-1.started.worker1.200").
					Return(startEntry, nil)

				compEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				compEntry.EXPECT().Value().Return([]byte(fmt.Sprintf(
					`{"job_id":"job-1","event":"completed","hostname":"worker1","timestamp":"%s"}`,
					now,
				)))
				s.mockKV.EXPECT().
					Get(gomock.Any(), "status.job-1.completed.worker1.300").
					Return(compEntry, nil)

				respEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				respEntry.EXPECT().Value().Return([]byte(
					`{"status":"completed","data":"eyJ0ZXN0IjogdHJ1ZX0="}`,
				))
				s.mockKV.EXPECT().
					Get(gomock.Any(), "responses.job-1.worker1.400").
					Return(respEntry, nil)
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
					`{"id":"job-1","operation":{"type":"node.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-1").Return(mockJobEntry, nil)

				s.mockKV.EXPECT().Keys(gomock.Any()).Return([]string{
					"status.job-1.started.worker1.100",
					"status.job-1.failed.worker1.200",
				}, nil)

				startEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				startEntry.EXPECT().Value().Return([]byte(fmt.Sprintf(
					`{"job_id":"job-1","event":"started","hostname":"worker1","timestamp":"%s"}`,
					now,
				)))
				s.mockKV.EXPECT().
					Get(gomock.Any(), "status.job-1.started.worker1.100").
					Return(startEntry, nil)

				failEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				failEntry.EXPECT().Value().Return([]byte(fmt.Sprintf(
					`{"job_id":"job-1","event":"failed","hostname":"worker1","timestamp":"%s","data":{"error":"disk full"}}`,
					now,
				)))
				s.mockKV.EXPECT().
					Get(gomock.Any(), "status.job-1.failed.worker1.200").
					Return(failEntry, nil)
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
					`{"id":"job-1","operation":{"type":"node.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-1").Return(mockJobEntry, nil)

				s.mockKV.EXPECT().Keys(gomock.Any()).Return([]string{
					"status.job-1.completed.worker1.100",
					"status.job-1.failed.worker2.200",
				}, nil)

				compEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				compEntry.EXPECT().Value().Return([]byte(fmt.Sprintf(
					`{"job_id":"job-1","event":"completed","hostname":"worker1","timestamp":"%s"}`,
					now,
				)))
				s.mockKV.EXPECT().
					Get(gomock.Any(), "status.job-1.completed.worker1.100").
					Return(compEntry, nil)

				failEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				failEntry.EXPECT().Value().Return([]byte(fmt.Sprintf(
					`{"job_id":"job-1","event":"failed","hostname":"worker2","timestamp":"%s","data":{"error":"timeout"}}`,
					now,
				)))
				s.mockKV.EXPECT().
					Get(gomock.Any(), "status.job-1.failed.worker2.200").
					Return(failEntry, nil)
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
					`{"id":"job-1","operation":{"type":"node.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-1").Return(mockJobEntry, nil)

				s.mockKV.EXPECT().Keys(gomock.Any()).Return([]string{
					"status.job-1.acknowledged.worker1.100",
					"status.job-1.started.worker1.200",
				}, nil)

				ackEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				ackEntry.EXPECT().Value().Return([]byte(fmt.Sprintf(
					`{"job_id":"job-1","event":"acknowledged","hostname":"worker1","timestamp":"%s"}`,
					now,
				)))
				s.mockKV.EXPECT().
					Get(gomock.Any(), "status.job-1.acknowledged.worker1.100").
					Return(ackEntry, nil)

				startEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				startEntry.EXPECT().Value().Return([]byte(fmt.Sprintf(
					`{"job_id":"job-1","event":"started","hostname":"worker1","timestamp":"%s"}`,
					now,
				)))
				s.mockKV.EXPECT().
					Get(gomock.Any(), "status.job-1.started.worker1.200").
					Return(startEntry, nil)
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
					`{"id":"job-1","operation":{"type":"node.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-1").Return(mockJobEntry, nil)

				s.mockKV.EXPECT().Keys(gomock.Any()).Return([]string{
					"status.job-1.completed.worker1.100",
				}, nil)

				s.mockKV.EXPECT().
					Get(gomock.Any(), "status.job-1.completed.worker1.100").
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
					`{"id":"job-1","operation":{"type":"node.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-1").Return(mockJobEntry, nil)

				s.mockKV.EXPECT().Keys(gomock.Any()).Return([]string{
					"status.job-1.completed.worker1.100",
				}, nil)

				invalidEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				invalidEntry.EXPECT().Value().Return([]byte(`invalid json`))
				s.mockKV.EXPECT().
					Get(gomock.Any(), "status.job-1.completed.worker1.100").
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
					`{"id":"job-1","operation":{"type":"node.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-1").Return(mockJobEntry, nil)

				s.mockKV.EXPECT().Keys(gomock.Any()).Return(nil, errors.New("connection failed"))
			},
		},
		{
			name:  "response parse error skipped",
			jobID: "job-1",
			setupMocks: func() {
				mockJobEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockJobEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-1","operation":{"type":"node.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-1").Return(mockJobEntry, nil)

				s.mockKV.EXPECT().Keys(gomock.Any()).Return([]string{
					"responses.job-1.worker1.100",
				}, nil)

				respEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				respEntry.EXPECT().Value().Return([]byte(`not json`))
				s.mockKV.EXPECT().
					Get(gomock.Any(), "responses.job-1.worker1.100").
					Return(respEntry, nil)
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
					`{"id":"job-1","operation":{"type":"node.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-1").Return(mockJobEntry, nil)

				s.mockKV.EXPECT().Keys(gomock.Any()).Return([]string{
					"responses.job-1.worker1.100",
				}, nil)

				s.mockKV.EXPECT().
					Get(gomock.Any(), "responses.job-1.worker1.100").
					Return(nil, errors.New("kv error"))
			},
			expectedStatus: "submitted",
			responseCount:  0,
		},
		{
			name:  "out-of-order timestamps triggers sort",
			jobID: "job-1",
			setupMocks: func() {
				mockJobEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockJobEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-1","operation":{"type":"node.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-1").Return(mockJobEntry, nil)

				s.mockKV.EXPECT().Keys(gomock.Any()).Return([]string{
					"status.job-1.completed.worker1.300",
					"status.job-1.started.worker1.100",
				}, nil)

				compEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				compEntry.EXPECT().Value().Return([]byte(
					`{"job_id":"job-1","event":"completed","hostname":"worker1","timestamp":"2024-01-01T00:00:05Z"}`,
				))
				s.mockKV.EXPECT().
					Get(gomock.Any(), "status.job-1.completed.worker1.300").
					Return(compEntry, nil)

				startEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				startEntry.EXPECT().Value().Return([]byte(
					`{"job_id":"job-1","event":"started","hostname":"worker1","timestamp":"2024-01-01T00:00:01Z"}`,
				))
				s.mockKV.EXPECT().
					Get(gomock.Any(), "status.job-1.started.worker1.100").
					Return(startEntry, nil)
			},
			expectedStatus: "processing",
			workerCount:    1,
		},
		{
			name:  "acknowledged only worker shows processing",
			jobID: "job-1",
			setupMocks: func() {
				mockJobEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockJobEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-1","operation":{"type":"node.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-1").Return(mockJobEntry, nil)

				s.mockKV.EXPECT().Keys(gomock.Any()).Return([]string{
					"status.job-1.acknowledged.worker1.100",
				}, nil)

				ackEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				ackEntry.EXPECT().Value().Return([]byte(fmt.Sprintf(
					`{"job_id":"job-1","event":"acknowledged","hostname":"worker1","timestamp":"%s"}`,
					now,
				)))
				s.mockKV.EXPECT().
					Get(gomock.Any(), "status.job-1.acknowledged.worker1.100").
					Return(ackEntry, nil)
			},
			expectedStatus: "processing",
			workerCount:    1,
		},
		{
			name:  "redelivered job has positive duration",
			jobID: "job-1",
			setupMocks: func() {
				mockJobEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockJobEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-1","operation":{"type":"node.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-1").Return(mockJobEntry, nil)

				// Simulate NATS redelivery: two started/failed cycles
				s.mockKV.EXPECT().Keys(gomock.Any()).Return([]string{
					"status.job-1.started.worker1.100",
					"status.job-1.failed.worker1.200",
					"status.job-1.started.worker1.300",
					"status.job-1.failed.worker1.400",
				}, nil)

				// First attempt: started at T+0, failed at T+1s
				start1 := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				start1.EXPECT().Value().Return([]byte(
					`{"job_id":"job-1","event":"started","hostname":"worker1","timestamp":"2024-01-01T00:00:01Z"}`,
				))
				s.mockKV.EXPECT().
					Get(gomock.Any(), "status.job-1.started.worker1.100").
					Return(start1, nil)

				fail1 := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				fail1.EXPECT().Value().Return([]byte(
					`{"job_id":"job-1","event":"failed","hostname":"worker1","timestamp":"2024-01-01T00:00:02Z","data":{"error":"attempt 1"}}`,
				))
				s.mockKV.EXPECT().
					Get(gomock.Any(), "status.job-1.failed.worker1.200").
					Return(fail1, nil)

				// Second attempt (redelivery): started at T+60s, failed at T+61s
				start2 := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				start2.EXPECT().Value().Return([]byte(
					`{"job_id":"job-1","event":"started","hostname":"worker1","timestamp":"2024-01-01T00:01:00Z"}`,
				))
				s.mockKV.EXPECT().
					Get(gomock.Any(), "status.job-1.started.worker1.300").
					Return(start2, nil)

				fail2 := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				fail2.EXPECT().Value().Return([]byte(
					`{"job_id":"job-1","event":"failed","hostname":"worker1","timestamp":"2024-01-01T00:01:01Z","data":{"error":"attempt 2"}}`,
				))
				s.mockKV.EXPECT().
					Get(gomock.Any(), "status.job-1.failed.worker1.400").
					Return(fail2, nil)
			},
			expectedStatus: "failed",
			expectedError:  "attempt 2",
			workerCount:    1,
			validateFunc: func(qj *job.QueuedJob) {
				ws := qj.WorkerStates["worker1"]
				// Duration should span from first start to last failure (60s)
				// and must be positive (not negative like the old bug)
				s.Equal("1m0s", ws.Duration)
				s.False(ws.StartTime.IsZero())
				s.False(ws.EndTime.IsZero())
				s.True(ws.EndTime.After(ws.StartTime))
			},
		},
		{
			name:  "completed job has sub-second duration with RFC3339Nano",
			jobID: "job-1",
			setupMocks: func() {
				mockJobEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockJobEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-1","operation":{"type":"node.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-1").Return(mockJobEntry, nil)

				s.mockKV.EXPECT().Keys(gomock.Any()).Return([]string{
					"status.job-1.started.worker1.100",
					"status.job-1.completed.worker1.200",
				}, nil)

				startEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				startEntry.EXPECT().Value().Return([]byte(
					`{"job_id":"job-1","event":"started","hostname":"worker1","timestamp":"2024-01-01T00:00:01.000000000Z"}`,
				))
				s.mockKV.EXPECT().
					Get(gomock.Any(), "status.job-1.started.worker1.100").
					Return(startEntry, nil)

				compEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				compEntry.EXPECT().Value().Return([]byte(
					`{"job_id":"job-1","event":"completed","hostname":"worker1","timestamp":"2024-01-01T00:00:01.045000000Z"}`,
				))
				s.mockKV.EXPECT().
					Get(gomock.Any(), "status.job-1.completed.worker1.200").
					Return(compEntry, nil)
			},
			expectedStatus: "completed",
			workerCount:    1,
			validateFunc: func(qj *job.QueuedJob) {
				ws := qj.WorkerStates["worker1"]
				s.Equal("45ms", ws.Duration)
			},
		},
		{
			name:  "invalid timestamp skipped",
			jobID: "job-1",
			setupMocks: func() {
				mockJobEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockJobEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-1","operation":{"type":"node.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-1").Return(mockJobEntry, nil)

				s.mockKV.EXPECT().Keys(gomock.Any()).Return([]string{
					"status.job-1.completed.worker1.100",
				}, nil)

				badEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				badEntry.EXPECT().Value().Return([]byte(
					`{"job_id":"job-1","event":"completed","hostname":"worker1","timestamp":"not-a-date"}`,
				))
				s.mockKV.EXPECT().
					Get(gomock.Any(), "status.job-1.completed.worker1.100").
					Return(badEntry, nil)
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
				if tt.validateFunc != nil {
					tt.validateFunc(jobStatus)
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
				s.mockKV.EXPECT().Keys(gomock.Any()).Return(nil, jetstream.ErrNoKeysFound)
			},
			expectedJobs: 0,
			expectedDLQ:  0,
		},
		{
			name:        "keys error",
			expectedErr: "error fetching jobs",
			setupMocks: func() {
				s.mockKV.EXPECT().Keys(gomock.Any()).Return(nil, errors.New("connection failed"))
			},
		},
		{
			name: "get job error skipped",
			setupMocks: func() {
				s.mockKV.EXPECT().Keys(gomock.Any()).Return([]string{"jobs.job-1"}, nil)
				s.mockKV.EXPECT().
					Get(gomock.Any(), "jobs.job-1").
					Return(nil, errors.New("kv error"))
				s.mockNATSClient.EXPECT().
					GetStreamInfo(gomock.Any(), "JOBS-DLQ").
					Return(nil, errors.New("no stream"))
			},
			expectedJobs: 0,
		},
		{
			name: "invalid job JSON skipped",
			setupMocks: func() {
				s.mockKV.EXPECT().Keys(gomock.Any()).Return([]string{"jobs.job-1"}, nil)

				mockEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return([]byte(`not json`))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-1").Return(mockEntry, nil)

				s.mockNATSClient.EXPECT().
					GetStreamInfo(gomock.Any(), "JOBS-DLQ").
					Return(nil, errors.New("no stream"))
			},
			expectedJobs: 0,
		},
		{
			name: "with DLQ info",
			setupMocks: func() {
				s.mockKV.EXPECT().Keys(gomock.Any()).Return([]string{"jobs.job-1"}, nil)

				mockEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-1","operation":{"type":"node.hostname.get"}}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-1").Return(mockEntry, nil)

				s.mockNATSClient.EXPECT().
					GetStreamInfo(gomock.Any(), "JOBS-DLQ").
					Return(&jetstream.StreamInfo{State: jetstream.StreamState{Msgs: 5}}, nil)
			},
			expectedJobs: 1,
			expectedDLQ:  5,
		},
		{
			name: "operation without type field",
			setupMocks: func() {
				s.mockKV.EXPECT().Keys(gomock.Any()).Return([]string{"jobs.job-1"}, nil)

				mockEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-1","operation":{"data":"some value"}}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-1").Return(mockEntry, nil)

				s.mockNATSClient.EXPECT().
					GetStreamInfo(gomock.Any(), "JOBS-DLQ").
					Return(nil, errors.New("no stream"))
			},
			expectedJobs: 1,
		},
		{
			name: "operation as non-map value",
			setupMocks: func() {
				s.mockKV.EXPECT().Keys(gomock.Any()).Return([]string{"jobs.job-1"}, nil)

				mockEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-1","operation":"string-value"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-1").Return(mockEntry, nil)

				s.mockNATSClient.EXPECT().
					GetStreamInfo(gomock.Any(), "JOBS-DLQ").
					Return(nil, errors.New("no stream"))
			},
			expectedJobs: 1,
		},
		{
			name: "non-jobs prefix keys skipped",
			setupMocks: func() {
				s.mockKV.EXPECT().Keys(gomock.Any()).Return([]string{
					"status.job-1.submitted._api.100",
					"responses.job-1.worker1.200",
				}, nil)

				s.mockNATSClient.EXPECT().
					GetStreamInfo(gomock.Any(), "JOBS-DLQ").
					Return(nil, errors.New("no stream"))
			},
			expectedJobs: 0,
		},
		{
			name: "with jobs and DLQ error",
			setupMocks: func() {
				keys := []string{"jobs.job-1", "jobs.job-2"}
				s.mockKV.EXPECT().Keys(gomock.Any()).Return(keys, nil)

				mockEntry1 := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry1.EXPECT().
					Value().
					Return([]byte(`{"id":"job-1","operation":{"type":"node.hostname.get"}}`))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-1").Return(mockEntry1, nil)

				mockEntry2 := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry2.EXPECT().
					Value().
					Return([]byte(`{"id":"job-2","operation":{"type":"node.status.get"}}`))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-2").Return(mockEntry2, nil)

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

func (s *JobsPublicTestSuite) TestGetQueueStatsDLQNameDerivedFromStreamName() {
	tests := []struct {
		name         string
		streamName   string
		expectedDLQ  string
		setupMocks   func(dlqName string)
		expectedMsgs int
	}{
		{
			name:        "when stream name is JOBS",
			streamName:  "JOBS",
			expectedDLQ: "JOBS-DLQ",
			setupMocks: func(dlqName string) {
				s.mockKV.EXPECT().Keys(gomock.Any()).Return([]string{}, nil)
				s.mockNATSClient.EXPECT().
					GetStreamInfo(gomock.Any(), dlqName).
					Return(&jetstream.StreamInfo{State: jetstream.StreamState{Msgs: 5}}, nil)
			},
			expectedMsgs: 5,
		},
		{
			name:        "when stream name is namespaced",
			streamName:  "osapi-JOBS",
			expectedDLQ: "osapi-JOBS-DLQ",
			setupMocks: func(dlqName string) {
				s.mockKV.EXPECT().Keys(gomock.Any()).Return([]string{}, nil)
				s.mockNATSClient.EXPECT().
					GetStreamInfo(gomock.Any(), dlqName).
					Return(&jetstream.StreamInfo{State: jetstream.StreamState{Msgs: 3}}, nil)
			},
			expectedMsgs: 3,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMocks(tt.expectedDLQ)

			opts := &client.Options{
				Timeout:    30 * time.Second,
				KVBucket:   s.mockKV,
				StreamName: tt.streamName,
			}
			c, err := client.New(slog.Default(), s.mockNATSClient, opts)
			s.Require().NoError(err)

			stats, err := c.GetQueueStats(s.ctx)
			s.NoError(err)
			s.NotNil(stats)
			s.Equal(tt.expectedMsgs, stats.DLQCount)
		})
	}
}

func (s *JobsPublicTestSuite) TestListJobs() {
	tests := []struct {
		name               string
		statusFilter       string
		limit              int
		offset             int
		expectedErr        string
		setupMocks         func()
		expectedJobs       int
		expectedTotalCount int
	}{
		{
			name:         "no jobs found",
			statusFilter: "",
			setupMocks: func() {
				s.mockKV.EXPECT().Keys(gomock.Any()).Return(nil, jetstream.ErrNoKeysFound)
			},
			expectedJobs:       0,
			expectedTotalCount: 0,
		},
		{
			name:        "kv error",
			expectedErr: "error fetching jobs",
			setupMocks: func() {
				s.mockKV.EXPECT().Keys(gomock.Any()).Return(nil, errors.New("connection failed"))
			},
		},
		{
			name: "returns all jobs no limit",
			setupMocks: func() {
				s.mockKV.EXPECT().
					Keys(gomock.Any()).
					Return([]string{"jobs.job-1", "jobs.job-2"}, nil)

				mockEntry1 := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry1.EXPECT().Value().Return([]byte(
					`{"id":"job-1","operation":{"type":"node.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-1").Return(mockEntry1, nil)

				mockEntry2 := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry2.EXPECT().Value().Return([]byte(
					`{"id":"job-2","operation":{"type":"node.status.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-2").Return(mockEntry2, nil)
			},
			expectedJobs:       2,
			expectedTotalCount: 2,
		},
		{
			name:         "filters out non matching status",
			statusFilter: "completed",
			setupMocks: func() {
				s.mockKV.EXPECT().Keys(gomock.Any()).Return([]string{"jobs.job-1"}, nil)

				mockEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-1","operation":{"type":"node.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-1").Return(mockEntry, nil)
			},
			expectedJobs:       0,
			expectedTotalCount: 0,
		},
		{
			name: "empty job ID after trim skipped",
			setupMocks: func() {
				s.mockKV.EXPECT().Keys(gomock.Any()).Return([]string{"jobs."}, nil)
			},
			expectedJobs:       0,
			expectedTotalCount: 0,
		},
		{
			name: "getJobStatusFromKeys error skipped",
			setupMocks: func() {
				s.mockKV.EXPECT().
					Keys(gomock.Any()).
					Return([]string{"jobs.job-bad"}, nil)

				s.mockKV.EXPECT().
					Get(gomock.Any(), "jobs.job-bad").
					Return(nil, errors.New("kv error"))
			},
			expectedJobs:       0,
			expectedTotalCount: 1,
		},
		{
			name: "only processes jobs prefix keys",
			setupMocks: func() {
				s.mockKV.EXPECT().Keys(gomock.Any()).Return(
					[]string{
						"status.job-1.submitted._api.123",
						"jobs.job-1",
						"responses.job-1.host.123",
					},
					nil,
				)

				mockJobEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockJobEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-1","operation":{"type":"node.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-1").Return(mockJobEntry, nil)

				mockStatusEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockStatusEntry.EXPECT().Value().Return([]byte(
					`{"job_id":"job-1","event":"submitted","hostname":"_api","timestamp":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().
					Get(gomock.Any(), "status.job-1.submitted._api.123").
					Return(mockStatusEntry, nil)

				mockRespEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockRespEntry.EXPECT().Value().Return([]byte(
					`{"status":"completed","data":"{}"}`,
				))
				s.mockKV.EXPECT().
					Get(gomock.Any(), "responses.job-1.host.123").
					Return(mockRespEntry, nil)
			},
			expectedJobs:       1,
			expectedTotalCount: 1,
		},
		{
			name:  "limit restricts returned jobs",
			limit: 1,
			setupMocks: func() {
				s.mockKV.EXPECT().
					Keys(gomock.Any()).
					Return([]string{"jobs.job-1", "jobs.job-2"}, nil)

				// Only job-2 is fetched (newest-first, limit 1)
				mockEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-2","operation":{"type":"node.status.get"},"created":"2024-01-02T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-2").Return(mockEntry, nil)
			},
			expectedJobs:       1,
			expectedTotalCount: 2,
		},
		{
			name:   "offset skips jobs",
			offset: 1,
			setupMocks: func() {
				s.mockKV.EXPECT().
					Keys(gomock.Any()).
					Return([]string{"jobs.job-1", "jobs.job-2"}, nil)

				// After reversing: [job-2, job-1], offset 1 skips job-2, returns job-1
				mockEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-1","operation":{"type":"node.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-1").Return(mockEntry, nil)
			},
			expectedJobs:       1,
			expectedTotalCount: 2,
		},
		{
			name:   "offset beyond total returns empty",
			offset: 10,
			setupMocks: func() {
				s.mockKV.EXPECT().
					Keys(gomock.Any()).
					Return([]string{"jobs.job-1", "jobs.job-2"}, nil)
			},
			expectedJobs:       0,
			expectedTotalCount: 2,
		},
		{
			name:         "filter with offset skips matching jobs",
			statusFilter: "submitted",
			offset:       1,
			setupMocks: func() {
				s.mockKV.EXPECT().
					Keys(gomock.Any()).
					Return([]string{"jobs.job-1", "jobs.job-2", "jobs.job-3"}, nil)

				// Reversed: job-3, job-2, job-1 — all submitted
				mockEntry3 := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry3.EXPECT().Value().Return([]byte(
					`{"id":"job-3","operation":{"type":"node.status.get"},"created":"2024-01-03T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-3").Return(mockEntry3, nil)

				mockEntry2 := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry2.EXPECT().Value().Return([]byte(
					`{"id":"job-2","operation":{"type":"node.status.get"},"created":"2024-01-02T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-2").Return(mockEntry2, nil)

				mockEntry1 := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry1.EXPECT().Value().Return([]byte(
					`{"id":"job-1","operation":{"type":"node.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-1").Return(mockEntry1, nil)
			},
			expectedJobs:       2,
			expectedTotalCount: 3,
		},
		{
			name:         "filter with limit restricts results",
			statusFilter: "submitted",
			limit:        1,
			setupMocks: func() {
				s.mockKV.EXPECT().
					Keys(gomock.Any()).
					Return([]string{"jobs.job-1", "jobs.job-2"}, nil)

				// Reversed: job-2, job-1 — both submitted
				mockEntry2 := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry2.EXPECT().Value().Return([]byte(
					`{"id":"job-2","operation":{"type":"node.status.get"},"created":"2024-01-02T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-2").Return(mockEntry2, nil)

				mockEntry1 := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry1.EXPECT().Value().Return([]byte(
					`{"id":"job-1","operation":{"type":"node.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-1").Return(mockEntry1, nil)
			},
			expectedJobs:       1,
			expectedTotalCount: 2,
		},
		{
			name:         "filter skips jobs with get error",
			statusFilter: "submitted",
			setupMocks: func() {
				s.mockKV.EXPECT().
					Keys(gomock.Any()).
					Return([]string{"jobs.job-bad", "jobs.job-good"}, nil)

				// Reversed: job-good, job-bad
				s.mockKV.EXPECT().
					Get(gomock.Any(), "jobs.job-good").
					Return(nil, errors.New("kv error"))

				mockEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-bad","operation":{"type":"node.status.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-bad").Return(mockEntry, nil)
			},
			expectedJobs:       1,
			expectedTotalCount: 1,
		},
		{
			name: "getJobStatusFromKeys with invalid JSON",
			setupMocks: func() {
				s.mockKV.EXPECT().
					Keys(gomock.Any()).
					Return([]string{"jobs.job-1"}, nil)

				mockEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return([]byte(`not valid json`))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-1").Return(mockEntry, nil)
			},
			expectedJobs:       0,
			expectedTotalCount: 1,
		},
		{
			name: "newest first ordering",
			setupMocks: func() {
				s.mockKV.EXPECT().
					Keys(gomock.Any()).
					Return([]string{"jobs.job-1", "jobs.job-2", "jobs.job-3"}, nil)

				// Reversed: job-3, job-2, job-1
				mockEntry3 := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry3.EXPECT().Value().Return([]byte(
					`{"id":"job-3","operation":{"type":"node.status.get"},"created":"2024-01-03T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-3").Return(mockEntry3, nil)

				mockEntry2 := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry2.EXPECT().Value().Return([]byte(
					`{"id":"job-2","operation":{"type":"node.status.get"},"created":"2024-01-02T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-2").Return(mockEntry2, nil)

				mockEntry1 := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry1.EXPECT().Value().Return([]byte(
					`{"id":"job-1","operation":{"type":"node.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-1").Return(mockEntry1, nil)
			},
			expectedJobs:       3,
			expectedTotalCount: 3,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMocks()

			result, err := s.jobsClient.ListJobs(
				s.ctx,
				tt.statusFilter,
				tt.limit,
				tt.offset,
			)

			if tt.expectedErr != "" {
				s.Error(err)
				s.Contains(err.Error(), tt.expectedErr)
			} else {
				s.NoError(err)
				s.NotNil(result)
				s.Len(result.Jobs, tt.expectedJobs)
				s.Equal(tt.expectedTotalCount, result.TotalCount)
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
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-123").Return(mockEntry, nil)
				s.mockKV.EXPECT().Delete(gomock.Any(), "jobs.job-123").Return(nil)
			},
		},
		{
			name:        "job not found",
			jobID:       "nonexistent",
			expectedErr: "job not found: nonexistent",
			setupMocks: func() {
				s.mockKV.EXPECT().
					Get(gomock.Any(), "jobs.nonexistent").
					Return(nil, errors.New("key not found"))
			},
		},
		{
			name:        "delete error",
			jobID:       "job-456",
			expectedErr: "failed to delete job",
			setupMocks: func() {
				mockEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return([]byte(`{"id":"job-456"}`)).AnyTimes()
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-456").Return(mockEntry, nil)
				s.mockKV.EXPECT().
					Delete(gomock.Any(), "jobs.job-456").
					Return(errors.New("storage failure"))
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

func (s *JobsPublicTestSuite) TestRetriedEventInTimeline() {
	now := time.Now().Format(time.RFC3339)

	tests := []struct {
		name         string
		jobID        string
		setupMocks   func()
		validateFunc func(qj *job.QueuedJob)
	}{
		{
			name:  "retried event appears in timeline with new job ID",
			jobID: "job-original",
			setupMocks: func() {
				mockJobEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockJobEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-original","operation":{"type":"node.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-original").Return(mockJobEntry, nil)

				s.mockKV.EXPECT().Keys(gomock.Any()).Return([]string{
					"status.job-original.submitted._api.100",
					"status.job-original.failed.worker1.200",
					"status.job-original.retried._api.300",
				}, nil)

				submitEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				submitEntry.EXPECT().Value().Return([]byte(fmt.Sprintf(
					`{"job_id":"job-original","event":"submitted","hostname":"_api","timestamp":"%s"}`,
					now,
				)))
				s.mockKV.EXPECT().
					Get(gomock.Any(), "status.job-original.submitted._api.100").
					Return(submitEntry, nil)

				failEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				failEntry.EXPECT().Value().Return([]byte(fmt.Sprintf(
					`{"job_id":"job-original","event":"failed","hostname":"worker1","timestamp":"%s","data":{"error":"timeout"}}`,
					now,
				)))
				s.mockKV.EXPECT().
					Get(gomock.Any(), "status.job-original.failed.worker1.200").
					Return(failEntry, nil)

				retriedEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				retriedEntry.EXPECT().Value().Return([]byte(fmt.Sprintf(
					`{"job_id":"job-original","event":"retried","hostname":"_api","timestamp":"%s","data":{"new_job_id":"job-new-123","target_hostname":"_any"}}`,
					now,
				)))
				s.mockKV.EXPECT().
					Get(gomock.Any(), "status.job-original.retried._api.300").
					Return(retriedEntry, nil)
			},
			validateFunc: func(qj *job.QueuedJob) {
				s.Equal("failed", qj.Status)
				s.Len(qj.Timeline, 3)

				// Find retried event in timeline
				var found bool
				for _, te := range qj.Timeline {
					if te.Event == "retried" {
						s.Contains(te.Message, "job-new-123")
						found = true
					}
				}
				s.True(found, "retried event should appear in timeline")
			},
		},
		{
			name:  "retried event without new_job_id shows fallback message",
			jobID: "job-original-2",
			setupMocks: func() {
				mockJobEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockJobEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-original-2","operation":{"type":"node.hostname.get"},"created":"2024-01-01T00:00:00Z"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-original-2").Return(mockJobEntry, nil)

				s.mockKV.EXPECT().Keys(gomock.Any()).Return([]string{
					"status.job-original-2.retried._api.100",
				}, nil)

				retriedEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				retriedEntry.EXPECT().Value().Return([]byte(fmt.Sprintf(
					`{"job_id":"job-original-2","event":"retried","hostname":"_api","timestamp":"%s","data":{}}`,
					now,
				)))
				s.mockKV.EXPECT().
					Get(gomock.Any(), "status.job-original-2.retried._api.100").
					Return(retriedEntry, nil)
			},
			validateFunc: func(qj *job.QueuedJob) {
				s.Len(qj.Timeline, 1)
				s.Equal("retried", qj.Timeline[0].Event)
				s.Equal("Job retried", qj.Timeline[0].Message)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMocks()

			qj, err := s.jobsClient.GetJobStatus(s.ctx, tt.jobID)
			s.NoError(err)
			s.NotNil(qj)
			tt.validateFunc(qj)
		})
	}
}

func (s *JobsPublicTestSuite) TestRetryJob() {
	tests := []struct {
		name        string
		jobID       string
		target      string
		expectedErr string
		setupMocks  func()
	}{
		{
			name:   "successful retry",
			jobID:  "job-123",
			target: "_any",
			setupMocks: func() {
				// Read original job
				mockEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-123","operation":{"type":"node.hostname.get","data":{}},"subject":"jobs.query._any"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-123").Return(mockEntry, nil)

				// CreateJob: store new job + status event + publish
				s.mockKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(1), nil).
					Times(2)
				s.mockKV.EXPECT().Bucket().Return("test-bucket").AnyTimes()
				s.mockNATSClient.EXPECT().
					Publish(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)

				// Write retried event on original job
				s.mockKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(2), nil)
			},
		},
		{
			name:        "job not found",
			jobID:       "nonexistent",
			target:      "_any",
			expectedErr: "job not found: nonexistent",
			setupMocks: func() {
				s.mockKV.EXPECT().
					Get(gomock.Any(), "jobs.nonexistent").
					Return(nil, errors.New("key not found"))
			},
		},
		{
			name:        "invalid job JSON",
			jobID:       "job-bad",
			target:      "_any",
			expectedErr: "failed to parse job data",
			setupMocks: func() {
				mockEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return([]byte(`not json`))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-bad").Return(mockEntry, nil)
			},
		},
		{
			name:        "missing operation field",
			jobID:       "job-no-op",
			target:      "_any",
			expectedErr: "job has no operation data",
			setupMocks: func() {
				mockEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-no-op","subject":"jobs.query._any"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-no-op").Return(mockEntry, nil)
			},
		},
		{
			name:        "create job fails",
			jobID:       "job-456",
			target:      "_any",
			expectedErr: "failed to create retry job",
			setupMocks: func() {
				mockEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-456","operation":{"type":"node.hostname.get","data":{}},"subject":"jobs.query._any"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-456").Return(mockEntry, nil)

				// CreateJob fails on KV put
				s.mockKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(0), errors.New("kv error"))
				s.mockKV.EXPECT().Bucket().Return("test-bucket").AnyTimes()
			},
		},
		{
			name:   "retried event put error is logged not returned",
			jobID:  "job-789",
			target: "_any",
			setupMocks: func() {
				mockEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-789","operation":{"type":"node.hostname.get","data":{}},"subject":"jobs.query._any"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-789").Return(mockEntry, nil)

				// CreateJob succeeds
				s.mockKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(1), nil).
					Times(2)
				s.mockKV.EXPECT().Bucket().Return("test-bucket").AnyTimes()
				s.mockNATSClient.EXPECT().
					Publish(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)

				// Retried event put fails (should be logged, not returned)
				s.mockKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(0), errors.New("event put failed"))
			},
		},
		{
			name:   "empty target defaults to any",
			jobID:  "job-empty-target",
			target: "",
			setupMocks: func() {
				mockEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return([]byte(
					`{"id":"job-empty-target","operation":{"type":"node.hostname.get","data":{}},"subject":"jobs.query._any"}`,
				))
				s.mockKV.EXPECT().Get(gomock.Any(), "jobs.job-empty-target").Return(mockEntry, nil)

				s.mockKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(1), nil).
					Times(2)
				s.mockKV.EXPECT().Bucket().Return("test-bucket").AnyTimes()
				s.mockNATSClient.EXPECT().
					Publish(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)

				// Retried event
				s.mockKV.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uint64(2), nil)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMocks()

			result, err := s.jobsClient.RetryJob(s.ctx, tt.jobID, tt.target)

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

func TestJobsPublicTestSuite(t *testing.T) {
	suite.Run(t, new(JobsPublicTestSuite))
}

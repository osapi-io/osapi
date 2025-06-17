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
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/nats-io/nats.go/jetstream"
	natsclient "github.com/osapi-io/nats-client/pkg/client"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/job/client"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
)

type WorkerPublicTestSuite struct {
	suite.Suite

	mockCtrl       *gomock.Controller
	mockNATSClient *jobmocks.MockNATSClient
	mockKV         *jobmocks.MockKeyValue
	jobsClient     *client.Client
	ctx            context.Context
}

func (s *WorkerPublicTestSuite) SetupTest() {
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

func (s *WorkerPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *WorkerPublicTestSuite) TestWriteStatusEvent() {
	tests := []struct {
		name         string
		jobID        string
		event        string
		hostname     string
		data         map[string]interface{}
		bucket       string
		kvError      error
		expectError  bool
		errorMsg     string
		validateKey  func(string) bool
		validateData func([]byte) bool
	}{
		{
			name:     "successful status event with data",
			jobID:    "job-123",
			event:    "started",
			hostname: "worker-1",
			data:     map[string]interface{}{"key": "value", "count": 42},
			bucket:   "test-bucket",
			validateKey: func(key string) bool {
				return len(key) > 0 && key[:6] == "status"
			},
			validateData: func(data []byte) bool {
				var eventData map[string]interface{}
				err := json.Unmarshal(data, &eventData)
				if err != nil {
					return false
				}
				return eventData["job_id"] == "job-123" &&
					eventData["event"] == "started" &&
					eventData["hostname"] == "worker-1"
			},
		},
		{
			name:     "successful status event without data",
			jobID:    "job-456",
			event:    "completed",
			hostname: "worker-2",
			data:     nil,
			bucket:   "test-bucket",
			validateKey: func(key string) bool {
				return len(key) > 0 && key[:6] == "status"
			},
			validateData: func(data []byte) bool {
				var eventData map[string]interface{}
				err := json.Unmarshal(data, &eventData)
				if err != nil {
					return false
				}
				_, hasData := eventData["data"]
				return eventData["job_id"] == "job-456" && !hasData
			},
		},
		{
			name:     "hostname with special characters",
			jobID:    "job-789",
			event:    "failed",
			hostname: "worker.host-name@domain.com",
			data:     map[string]interface{}{"error": "timeout"},
			bucket:   "test-bucket",
			validateKey: func(key string) bool {
				// Should sanitize hostname in key
				return len(key) > 0 && key[:6] == "status"
			},
			validateData: func(data []byte) bool {
				var eventData map[string]interface{}
				err := json.Unmarshal(data, &eventData)
				if err != nil {
					return false
				}
				return eventData["hostname"] == "worker.host-name@domain.com"
			},
		},
		{
			name:        "KV put error",
			jobID:       "job-error",
			event:       "started",
			hostname:    "worker-1",
			data:        map[string]interface{}{"key": "value"},
			bucket:      "test-bucket",
			kvError:     errors.New("kv connection failed"),
			expectError: true,
			errorMsg:    "failed to write status event",
		},
		{
			name:     "empty job ID",
			jobID:    "",
			event:    "started",
			hostname: "worker-1",
			data:     map[string]interface{}{"key": "value"},
			bucket:   "test-bucket",
			validateKey: func(key string) bool {
				return len(key) > 0 && key[:6] == "status"
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.mockKV.EXPECT().Bucket().Return(tt.bucket)

			if tt.expectError {
				s.mockNATSClient.EXPECT().
					KVPut(tt.bucket, gomock.Any(), gomock.Any()).
					Return(tt.kvError)
			} else {
				s.mockNATSClient.EXPECT().
					KVPut(tt.bucket, gomock.Any(), gomock.Any()).
					Do(func(_, key string, data []byte) {
						if tt.validateKey != nil {
							s.True(tt.validateKey(key), "Key validation failed for: %s", key)
						}
						if tt.validateData != nil {
							s.True(tt.validateData(data), "Data validation failed")
						}
					}).
					Return(nil)
			}

			err := s.jobsClient.WriteStatusEvent(s.ctx, tt.jobID, tt.event, tt.hostname, tt.data)

			if tt.expectError {
				s.Error(err)
				s.Contains(err.Error(), tt.errorMsg)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *WorkerPublicTestSuite) TestWriteJobResponse() {
	tests := []struct {
		name         string
		jobID        string
		hostname     string
		responseData []byte
		status       string
		errorMsg     string
		bucket       string
		kvError      error
		expectError  bool
		errorText    string
		validateKey  func(string) bool
		validateData func([]byte) bool
	}{
		{
			name:         "successful job response completed",
			jobID:        "job-123",
			hostname:     "worker-1",
			responseData: []byte(`{"result": "success", "count": 42}`),
			status:       "completed",
			errorMsg:     "",
			bucket:       "test-bucket",
			validateKey: func(key string) bool {
				return len(key) > 0 && key[:9] == "responses"
			},
			validateData: func(data []byte) bool {
				var response job.Response
				err := json.Unmarshal(data, &response)
				if err != nil {
					return false
				}
				return string(response.Status) == "completed" &&
					response.Error == "" &&
					!response.Timestamp.IsZero()
			},
		},
		{
			name:         "successful job response with error",
			jobID:        "job-456",
			hostname:     "worker-2",
			responseData: []byte(`{"error": "processing failed"}`),
			status:       "failed",
			errorMsg:     "task execution failed",
			bucket:       "test-bucket",
			validateKey: func(key string) bool {
				return len(key) > 0 && key[:9] == "responses"
			},
			validateData: func(data []byte) bool {
				var response job.Response
				err := json.Unmarshal(data, &response)
				if err != nil {
					return false
				}
				return string(response.Status) == "failed" &&
					response.Error == "task execution failed"
			},
		},
		{
			name:         "empty response data",
			jobID:        "job-789",
			hostname:     "worker-3",
			responseData: []byte{},
			status:       "completed",
			errorMsg:     "",
			bucket:       "test-bucket",
			validateKey: func(key string) bool {
				return len(key) > 0 && key[:9] == "responses"
			},
			validateData: func(data []byte) bool {
				var response job.Response
				err := json.Unmarshal(data, &response)
				if err != nil {
					return false
				}
				return len(response.Data) == 0
			},
		},
		{
			name:         "hostname with special characters",
			jobID:        "job-special",
			hostname:     "worker.host-name@domain.com",
			responseData: []byte(`{"data": "test"}`),
			status:       "completed",
			errorMsg:     "",
			bucket:       "test-bucket",
			validateKey: func(key string) bool {
				// Should sanitize hostname in key
				return len(key) > 0 && key[:9] == "responses"
			},
		},
		{
			name:         "KV put error",
			jobID:        "job-error",
			hostname:     "worker-1",
			responseData: []byte(`{"result": "success"}`),
			status:       "completed",
			errorMsg:     "",
			bucket:       "test-bucket",
			kvError:      errors.New("storage failure"),
			expectError:  true,
			errorText:    "failed to store job response",
		},
		{
			name:     "large response data",
			jobID:    "job-large",
			hostname: "worker-1",
			responseData: []byte(
				`{"data": "large_data_payload_with_repeated_content_` + strings.Repeat(
					"x",
					500,
				) + `"}`,
			),
			status:   "completed",
			errorMsg: "",
			bucket:   "test-bucket",
			validateKey: func(key string) bool {
				return len(key) > 0 && key[:9] == "responses"
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.mockKV.EXPECT().Bucket().Return(tt.bucket)

			if tt.expectError {
				s.mockNATSClient.EXPECT().
					KVPut(tt.bucket, gomock.Any(), gomock.Any()).
					Return(tt.kvError)
			} else {
				s.mockNATSClient.EXPECT().
					KVPut(tt.bucket, gomock.Any(), gomock.Any()).
					Do(func(_, key string, data []byte) {
						if tt.validateKey != nil {
							s.True(tt.validateKey(key), "Key validation failed for: %s", key)
						}
						if tt.validateData != nil {
							s.True(tt.validateData(data), "Data validation failed")
						}
					}).
					Return(nil)
			}

			err := s.jobsClient.WriteJobResponse(
				s.ctx,
				tt.jobID,
				tt.hostname,
				tt.responseData,
				tt.status,
				tt.errorMsg,
			)

			if tt.expectError {
				s.Error(err)
				s.Contains(err.Error(), tt.errorText)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *WorkerPublicTestSuite) TestConsumeJobs() {
	tests := []struct {
		name         string
		streamName   string
		consumerName string
		handler      func(jetstream.Msg) error
		opts         *natsclient.ConsumeOptions
		consumeError error
		expectError  bool
		errorMsg     string
	}{
		{
			name:         "successful job consumption",
			streamName:   "test-stream",
			consumerName: "test-consumer",
			handler: func(_ jetstream.Msg) error {
				return nil
			},
			opts: nil,
		},
		{
			name:         "successful job consumption with options",
			streamName:   "jobs-stream",
			consumerName: "worker-consumer",
			handler: func(_ jetstream.Msg) error {
				return nil
			},
			opts: &natsclient.ConsumeOptions{
				QueueGroup:  "test-queue",
				MaxInFlight: 5,
			},
		},
		{
			name:         "handler that returns error",
			streamName:   "test-stream",
			consumerName: "test-consumer",
			handler: func(_ jetstream.Msg) error {
				return errors.New("handler processing error")
			},
			opts: nil,
		},
		{
			name:         "NATS consume error",
			streamName:   "test-stream",
			consumerName: "test-consumer",
			handler: func(_ jetstream.Msg) error {
				return nil
			},
			opts:         nil,
			consumeError: errors.New("stream not found"),
			expectError:  true,
			errorMsg:     "stream not found",
		},
		{
			name:         "empty stream name",
			streamName:   "",
			consumerName: "test-consumer",
			handler: func(_ jetstream.Msg) error {
				return nil
			},
			opts: nil,
		},
		{
			name:         "empty consumer name",
			streamName:   "test-stream",
			consumerName: "",
			handler: func(_ jetstream.Msg) error {
				return nil
			},
			opts: nil,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.mockNATSClient.EXPECT().
				ConsumeMessages(gomock.Any(), tt.streamName, tt.consumerName, gomock.Any(), tt.opts).
				Return(tt.consumeError)

			err := s.jobsClient.ConsumeJobs(
				s.ctx,
				tt.streamName,
				tt.consumerName,
				tt.handler,
				tt.opts,
			)

			if tt.expectError {
				s.Error(err)
				s.Contains(err.Error(), tt.errorMsg)
			} else {
				if tt.consumeError != nil {
					s.Error(err)
				} else {
					s.NoError(err)
				}
			}
		})
	}
}

func (s *WorkerPublicTestSuite) TestGetJobData() {
	tests := []struct {
		name         string
		jobKey       string
		expectedErr  string
		setupMocks   func()
		expectedData []byte
	}{
		{
			name:   "successful get job data",
			jobKey: "jobs.job-123",
			setupMocks: func() {
				mockEntry := jobmocks.NewMockKeyValueEntry(s.mockCtrl)
				mockEntry.EXPECT().Value().Return([]byte(`{"test": "data"}`))
				s.mockKV.EXPECT().Get("jobs.job-123").Return(mockEntry, nil)
			},
			expectedData: []byte(`{"test": "data"}`),
		},
		{
			name:        "job not found error",
			jobKey:      "jobs.nonexistent",
			expectedErr: "failed to get job data for key jobs.nonexistent",
			setupMocks: func() {
				s.mockKV.EXPECT().Get("jobs.nonexistent").Return(nil, errors.New("key not found"))
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMocks()

			data, err := s.jobsClient.GetJobData(s.ctx, tt.jobKey)

			if tt.expectedErr != "" {
				s.Error(err)
				s.Contains(err.Error(), tt.expectedErr)
			} else {
				s.NoError(err)
				s.Equal(tt.expectedData, data)
			}
		})
	}
}

func (s *WorkerPublicTestSuite) TestCreateOrUpdateConsumer() {
	tests := []struct {
		name           string
		streamName     string
		consumerConfig jetstream.ConsumerConfig
		expectedErr    string
		setupMocks     func()
	}{
		{
			name:           "successful consumer creation",
			streamName:     "test-stream",
			consumerConfig: jetstream.ConsumerConfig{Name: "test-consumer"},
			setupMocks: func() {
				s.mockNATSClient.EXPECT().
					CreateOrUpdateConsumerWithConfig(gomock.Any(), "test-stream", jetstream.ConsumerConfig{Name: "test-consumer"}).
					Return(nil)
			},
		},
		{
			name:           "consumer creation error",
			streamName:     "test-stream",
			consumerConfig: jetstream.ConsumerConfig{Name: "test-consumer"},
			expectedErr:    "consumer creation failed",
			setupMocks: func() {
				s.mockNATSClient.EXPECT().
					CreateOrUpdateConsumerWithConfig(gomock.Any(), "test-stream", jetstream.ConsumerConfig{Name: "test-consumer"}).
					Return(errors.New("consumer creation failed"))
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMocks()

			err := s.jobsClient.CreateOrUpdateConsumer(s.ctx, tt.streamName, tt.consumerConfig)

			if tt.expectedErr != "" {
				s.Error(err)
				s.Contains(err.Error(), tt.expectedErr)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *WorkerPublicTestSuite) TestSanitizeKeyForNATS() {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid characters only",
			input:    "validKey123",
			expected: "validKey123",
		},
		{
			name:     "alphanumeric with underscores and hyphens",
			input:    "valid_key-123",
			expected: "valid_key-123",
		},
		{
			name:     "hostname with dots",
			input:    "server.example.com",
			expected: "server_example_com",
		},
		{
			name:     "hostname with special characters",
			input:    "worker.host-name@domain.com",
			expected: "worker_host-name_domain_com",
		},
		{
			name:     "email-like string",
			input:    "user@domain.com",
			expected: "user_domain_com",
		},
		{
			name:     "string with spaces",
			input:    "worker node 1",
			expected: "worker_node_1",
		},
		{
			name:     "string with mixed special characters",
			input:    "worker#1!@#$%^&*()",
			expected: "worker_1__________",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only special characters",
			input:    "!@#$%^&*()",
			expected: "_________",
		},
		{
			name:     "path-like string",
			input:    "/path/to/resource",
			expected: "_path_to_resource",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// We need to access the private function through reflection or create a test helper
			// Since it's a private function, we'll test it indirectly through the public methods
			// that use it (WriteStatusEvent and WriteJobResponse)

			// For this test, we'll create a simple wrapper or test the behavior through
			// the public methods that call sanitizeKeyForNATS
			s.mockKV.EXPECT().Bucket().Return("test-bucket")
			s.mockNATSClient.EXPECT().
				KVPut("test-bucket", gomock.Any(), gomock.Any()).
				Do(func(_, key string, _ []byte) {
					// Verify that the key contains the sanitized version
					s.Contains(key, tt.expected)
				}).
				Return(nil)

			// Test through WriteStatusEvent which uses sanitizeKeyForNATS
			err := s.jobsClient.WriteStatusEvent(s.ctx, "test-job", "started", tt.input, nil)
			s.NoError(err)
		})
	}
}

func TestWorkerPublicTestSuite(t *testing.T) {
	suite.Run(t, new(WorkerPublicTestSuite))
}

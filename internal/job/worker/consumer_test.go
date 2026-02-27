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

package worker

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/job/mocks"
	commandMocks "github.com/retr0h/osapi/internal/provider/command/mocks"
	dnsMocks "github.com/retr0h/osapi/internal/provider/network/dns/mocks"
	pingMocks "github.com/retr0h/osapi/internal/provider/network/ping/mocks"
	diskMocks "github.com/retr0h/osapi/internal/provider/system/disk/mocks"
	hostMocks "github.com/retr0h/osapi/internal/provider/system/host/mocks"
	loadMocks "github.com/retr0h/osapi/internal/provider/system/load/mocks"
	memMocks "github.com/retr0h/osapi/internal/provider/system/mem/mocks"
)

type ConsumerTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *mocks.MockJobClient
	worker        *Worker
}

func (s *ConsumerTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = mocks.NewMockJobClient(s.mockCtrl)

	appFs := afero.NewMemMapFs()
	appConfig := config.Config{
		NATS: config.NATS{
			Stream: config.NATSStream{
				Name: "test-stream",
			},
		},
		Job: config.Job{
			Worker: config.JobWorker{
				Hostname:   "test-worker",
				QueueGroup: "test-queue",
				MaxJobs:    5,
				Consumer: config.JobWorkerConsumer{
					AckWait:       "30s",
					BackOff:       []string{"1s", "2s"},
					MaxDeliver:    3,
					MaxAckPending: 10,
					ReplayPolicy:  "instant",
				},
			},
		},
	}

	// Create mock providers
	hostMock := hostMocks.NewDefaultMockProvider(s.mockCtrl)
	diskMock := diskMocks.NewDefaultMockProvider(s.mockCtrl)
	memMock := memMocks.NewDefaultMockProvider(s.mockCtrl)
	loadMock := loadMocks.NewDefaultMockProvider(s.mockCtrl)
	dnsMock := dnsMocks.NewDefaultMockProvider(s.mockCtrl)
	pingMock := pingMocks.NewDefaultMockProvider(s.mockCtrl)
	commandMock := commandMocks.NewDefaultMockProvider(s.mockCtrl)

	s.worker = New(
		appFs,
		appConfig,
		slog.Default(),
		s.mockJobClient,
		"test-stream",
		hostMock,
		diskMock,
		memMock,
		loadMock,
		dnsMock,
		pingMock,
		commandMock,
		nil,
	)
}

func (s *ConsumerTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *ConsumerTestSuite) TestConsumeQueryJobs() {
	tests := []struct {
		name       string
		hostname   string
		setupMocks func()
		expectErr  bool
	}{
		{
			name:     "successful query job consumption",
			hostname: "test-worker",
			setupMocks: func() {
				// Expect consumer creation for all 3 query patterns
				s.mockJobClient.EXPECT().
					CreateOrUpdateConsumer(gomock.Any(), "test-stream", gomock.Any()).
					Return(nil).
					Times(3)

				// Expect job consumption for all consumers
				s.mockJobClient.EXPECT().
					ConsumeJobs(gomock.Any(), "test-stream", gomock.Any(), gomock.Any(), gomock.Any()).
					Return(context.Canceled).
					Times(3)
			},
			expectErr: false,
		},
		{
			name:     "consumer creation failure",
			hostname: "test-worker",
			setupMocks: func() {
				// Fail all consumer creations
				s.mockJobClient.EXPECT().
					CreateOrUpdateConsumer(gomock.Any(), "test-stream", gomock.Any()).
					Return(errors.New("consumer creation failed")).
					Times(3)

				// No consumption should happen since creation failed
			},
			expectErr: false, // Should not return error, just log and continue
		},
		{
			name:     "partial consumer creation failure",
			hostname: "test-worker",
			setupMocks: func() {
				// First consumer succeeds
				s.mockJobClient.EXPECT().
					CreateOrUpdateConsumer(gomock.Any(), "test-stream", gomock.Any()).
					Return(nil).
					Times(1)

				// Second and third consumers fail
				s.mockJobClient.EXPECT().
					CreateOrUpdateConsumer(gomock.Any(), "test-stream", gomock.Any()).
					Return(errors.New("consumer creation failed")).
					Times(2)

				// Only one consumption should happen
				s.mockJobClient.EXPECT().
					ConsumeJobs(gomock.Any(), "test-stream", gomock.Any(), gomock.Any(), gomock.Any()).
					Return(context.Canceled).
					Times(1)
			},
			expectErr: false,
		},
		{
			name:     "empty hostname",
			hostname: "",
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					CreateOrUpdateConsumer(gomock.Any(), "test-stream", gomock.Any()).
					Return(nil).
					Times(3)

				s.mockJobClient.EXPECT().
					ConsumeJobs(gomock.Any(), "test-stream", gomock.Any(), gomock.Any(), gomock.Any()).
					Return(context.Canceled).
					Times(3)
			},
			expectErr: false,
		},
		{
			name:     "consume error logged",
			hostname: "test-worker",
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					CreateOrUpdateConsumer(gomock.Any(), "test-stream", gomock.Any()).
					Return(nil).
					Times(3)

				s.mockJobClient.EXPECT().
					ConsumeJobs(gomock.Any(), "test-stream", gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("connection lost")).
					Times(3)
			},
			expectErr: false,
		},
		{
			name:     "with labels creates extra consumers",
			hostname: "test-worker",
			setupMocks: func() {
				// 3 base + 3 prefix levels for group:web.dev.us-east = 6 total
				s.mockJobClient.EXPECT().
					CreateOrUpdateConsumer(gomock.Any(), "test-stream", gomock.Any()).
					Return(nil).
					Times(6)

				s.mockJobClient.EXPECT().
					ConsumeJobs(gomock.Any(), "test-stream", gomock.Any(), gomock.Any(), gomock.Any()).
					Return(context.Canceled).
					Times(6)
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Set labels for the label-specific test case
			if tt.name == "with labels creates extra consumers" {
				s.worker.appConfig.Job.Worker.Labels = map[string]string{
					"group": "web.dev.us-east",
				}
			} else {
				s.worker.appConfig.Job.Worker.Labels = nil
			}

			tt.setupMocks()

			err := s.worker.consumeQueryJobs(ctx, tt.hostname)

			if tt.expectErr {
				s.Error(err)
			} else {
				s.NoError(err)
			}

			// Allow goroutines to execute before cleanup
			time.Sleep(10 * time.Millisecond)
		})
	}
}

func (s *ConsumerTestSuite) TestConsumeModifyJobs() {
	tests := []struct {
		name       string
		hostname   string
		setupMocks func()
		expectErr  bool
	}{
		{
			name:     "successful modify job consumption",
			hostname: "test-worker",
			setupMocks: func() {
				// Expect consumer creation for all 3 modify patterns
				s.mockJobClient.EXPECT().
					CreateOrUpdateConsumer(gomock.Any(), "test-stream", gomock.Any()).
					Return(nil).
					Times(3)

				// Expect job consumption for all consumers
				s.mockJobClient.EXPECT().
					ConsumeJobs(gomock.Any(), "test-stream", gomock.Any(), gomock.Any(), gomock.Any()).
					Return(context.Canceled).
					Times(3)
			},
			expectErr: false,
		},
		{
			name:     "consumer creation failure",
			hostname: "test-worker",
			setupMocks: func() {
				// Fail all consumer creations
				s.mockJobClient.EXPECT().
					CreateOrUpdateConsumer(gomock.Any(), "test-stream", gomock.Any()).
					Return(errors.New("consumer creation failed")).
					Times(3)
			},
			expectErr: false, // Should not return error, just log and continue
		},
		{
			name:     "hostname with special characters",
			hostname: "test-worker.domain.com",
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					CreateOrUpdateConsumer(gomock.Any(), "test-stream", gomock.Any()).
					Return(nil).
					Times(3)

				s.mockJobClient.EXPECT().
					ConsumeJobs(gomock.Any(), "test-stream", gomock.Any(), gomock.Any(), gomock.Any()).
					Return(context.Canceled).
					Times(3)
			},
			expectErr: false,
		},
		{
			name:     "consume error logged",
			hostname: "test-worker",
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					CreateOrUpdateConsumer(gomock.Any(), "test-stream", gomock.Any()).
					Return(nil).
					Times(3)

				s.mockJobClient.EXPECT().
					ConsumeJobs(gomock.Any(), "test-stream", gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("connection lost")).
					Times(3)
			},
			expectErr: false,
		},
		{
			name:     "with labels creates extra consumers",
			hostname: "test-worker",
			setupMocks: func() {
				// 3 base + 3 prefix levels for group:web.dev.us-east = 6 total
				s.mockJobClient.EXPECT().
					CreateOrUpdateConsumer(gomock.Any(), "test-stream", gomock.Any()).
					Return(nil).
					Times(6)

				s.mockJobClient.EXPECT().
					ConsumeJobs(gomock.Any(), "test-stream", gomock.Any(), gomock.Any(), gomock.Any()).
					Return(context.Canceled).
					Times(6)
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Set labels for the label-specific test case
			if tt.name == "with labels creates extra consumers" {
				s.worker.appConfig.Job.Worker.Labels = map[string]string{
					"group": "web.dev.us-east",
				}
			} else {
				s.worker.appConfig.Job.Worker.Labels = nil
			}

			tt.setupMocks()

			err := s.worker.consumeModifyJobs(ctx, tt.hostname)

			if tt.expectErr {
				s.Error(err)
			} else {
				s.NoError(err)
			}

			// Allow goroutines to execute before cleanup
			time.Sleep(10 * time.Millisecond)
		})
	}
}

func (s *ConsumerTestSuite) TestCreateConsumer() {
	tests := []struct {
		name          string
		streamName    string
		consumerName  string
		filterSubject string
		config        config.JobWorkerConsumer
		setupMocks    func()
		expectErr     bool
		errorMsg      string
	}{
		{
			name:          "successful consumer creation with instant replay",
			streamName:    "test-stream",
			consumerName:  "test-consumer",
			filterSubject: "jobs.query._any",
			config: config.JobWorkerConsumer{
				AckWait:       "30s",
				BackOff:       []string{"1s", "2s", "5s"},
				MaxDeliver:    3,
				MaxAckPending: 10,
				ReplayPolicy:  "instant",
			},
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					CreateOrUpdateConsumer(gomock.Any(), "test-stream", gomock.Any()).
					Do(func(_ context.Context, _ string, config jetstream.ConsumerConfig) {
						s.Equal("test-consumer", config.Durable)
						s.Equal("jobs.query._any", config.FilterSubject)
						s.Equal(jetstream.AckExplicitPolicy, config.AckPolicy)
						s.Equal(jetstream.DeliverAllPolicy, config.DeliverPolicy)
						s.Equal(3, config.MaxDeliver)
						s.Equal(30*time.Second, config.AckWait)
						s.Len(config.BackOff, 3)
						s.Equal(10, config.MaxAckPending)
						s.Equal(jetstream.ReplayInstantPolicy, config.ReplayPolicy)
					}).
					Return(nil)
			},
			expectErr: false,
		},
		{
			name:          "successful consumer creation with original replay",
			streamName:    "test-stream",
			consumerName:  "test-consumer-orig",
			filterSubject: "jobs.modify._any",
			config: config.JobWorkerConsumer{
				AckWait:       "60s",
				BackOff:       []string{"2s"},
				MaxDeliver:    5,
				MaxAckPending: 20,
				ReplayPolicy:  "original",
			},
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					CreateOrUpdateConsumer(gomock.Any(), "test-stream", gomock.Any()).
					Do(func(_ context.Context, _ string, config jetstream.ConsumerConfig) {
						s.Equal("test-consumer-orig", config.Durable)
						s.Equal("jobs.modify._any", config.FilterSubject)
						s.Equal(60*time.Second, config.AckWait)
						s.Len(config.BackOff, 1)
						s.Equal(5, config.MaxDeliver)
						s.Equal(20, config.MaxAckPending)
						s.Equal(jetstream.ReplayOriginalPolicy, config.ReplayPolicy)
					}).
					Return(nil)
			},
			expectErr: false,
		},
		{
			name:          "consumer creation failure",
			streamName:    "test-stream",
			consumerName:  "test-consumer",
			filterSubject: "jobs.query._any",
			config: config.JobWorkerConsumer{
				AckWait:       "30s",
				BackOff:       []string{"1s"},
				MaxDeliver:    3,
				MaxAckPending: 10,
				ReplayPolicy:  "instant",
			},
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					CreateOrUpdateConsumer(gomock.Any(), "test-stream", gomock.Any()).
					Return(errors.New("stream not found"))
			},
			expectErr: true,
			errorMsg:  "stream not found",
		},
		{
			name:          "invalid duration in config",
			streamName:    "test-stream",
			consumerName:  "test-consumer",
			filterSubject: "jobs.query._any",
			config: config.JobWorkerConsumer{
				AckWait:       "invalid-duration",
				BackOff:       []string{"invalid", "2s"},
				MaxDeliver:    3,
				MaxAckPending: 10,
				ReplayPolicy:  "instant",
			},
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					CreateOrUpdateConsumer(gomock.Any(), "test-stream", gomock.Any()).
					Do(func(_ context.Context, _ string, config jetstream.ConsumerConfig) {
						// Should have default AckWait (0) and only valid BackOff durations
						s.Equal(time.Duration(0), config.AckWait)
						s.Len(config.BackOff, 1) // Only valid "2s" should be included
					}).
					Return(nil)
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Update worker config for this test
			s.worker.appConfig.Job.Worker.Consumer = tt.config

			tt.setupMocks()

			err := s.worker.createConsumer(
				context.Background(),
				tt.streamName,
				tt.consumerName,
				tt.filterSubject,
			)

			if tt.expectErr {
				s.Error(err)
				s.Contains(err.Error(), tt.errorMsg)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *ConsumerTestSuite) TestHandleJobMessageJS() {
	tests := []struct {
		name       string
		msgData    []byte
		msgSubject string
		setupMocks func()
		expectErr  bool
		errorMsg   string
	}{
		{
			name:       "successful message handling",
			msgData:    []byte("test-job-key"),
			msgSubject: "jobs.query.test-worker",
			setupMocks: func() {
				// Mock successful job data retrieval and processing
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.test-job-key").
					Return([]byte(`{
						"id": "test-job-123",
						"operation": {
							"type": "node.hostname.get",
							"data": {}
						}
					}`), nil)

				// Mock status event writes
				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "test-job-key", "acknowledged", gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "test-job-key", "started", gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "test-job-key", "completed", gomock.Any(), gomock.Any()).
					Return(nil)

				// Mock response write
				s.mockJobClient.EXPECT().
					WriteJobResponse(gomock.Any(), "test-job-key", gomock.Any(), gomock.Any(), "completed", "", gomock.Any()).
					Return(nil)
			},
			expectErr: false,
		},
		{
			name:       "job processing failure",
			msgData:    []byte("failed-job-key"),
			msgSubject: "jobs.query.test-worker",
			setupMocks: func() {
				// Mock job data retrieval failure
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.failed-job-key").
					Return(nil, errors.New("job not found"))
			},
			expectErr: true,
			errorMsg:  "job not found",
		},
		{
			name:       "invalid subject",
			msgData:    []byte("test-job-key"),
			msgSubject: "invalid.subject",
			setupMocks: func() {
				// No mocks needed as it should fail early
			},
			expectErr: true,
			errorMsg:  "failed to parse subject",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMocks()

			// Create mock JetStream message
			mockMsg := &mockJetStreamMsg{
				subject: tt.msgSubject,
				data:    tt.msgData,
			}

			err := s.worker.handleJobMessageJS(mockMsg)

			if tt.expectErr {
				s.Error(err)
				s.Contains(err.Error(), tt.errorMsg)
			} else {
				s.NoError(err)
			}
		})
	}
}

// mockJetStreamMsg implements jetstream.Msg interface for testing
type mockJetStreamMsg struct {
	subject string
	data    []byte
}

func (m *mockJetStreamMsg) Subject() string                           { return m.subject }
func (m *mockJetStreamMsg) Data() []byte                              { return m.data }
func (m *mockJetStreamMsg) Headers() nats.Header                      { return nil }
func (m *mockJetStreamMsg) Reply() string                             { return "" }
func (m *mockJetStreamMsg) Metadata() (*jetstream.MsgMetadata, error) { return nil, nil }
func (m *mockJetStreamMsg) Ack() error                                { return nil }
func (m *mockJetStreamMsg) DoubleAck(_ context.Context) error         { return nil }
func (m *mockJetStreamMsg) Nak() error                                { return nil }
func (m *mockJetStreamMsg) NakWithDelay(time.Duration) error          { return nil }
func (m *mockJetStreamMsg) Term() error                               { return nil }
func (m *mockJetStreamMsg) TermWithReason(string) error               { return nil }
func (m *mockJetStreamMsg) InProgress() error                         { return nil }

func TestConsumerTestSuite(t *testing.T) {
	suite.Run(t, new(ConsumerTestSuite))
}

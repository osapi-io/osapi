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

package agent_test

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/nats-io/nats.go/jetstream"
	natsclient "github.com/osapi-io/nats-client/pkg/client"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/agent"
	agentmocks "github.com/osapi-io/osapi/internal/agent/mocks"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/job/mocks"
	diskMocks "github.com/osapi-io/osapi/internal/provider/node/disk/mocks"
	hostMocks "github.com/osapi-io/osapi/internal/provider/node/host/mocks"
	loadMocks "github.com/osapi-io/osapi/internal/provider/node/load/mocks"
	memoryMocks "github.com/osapi-io/osapi/internal/provider/node/memory/mocks"
	processMocks "github.com/osapi-io/osapi/internal/telemetry/process/mocks"
)

type ConsumerPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *mocks.MockJobClient
	testAgent     *agent.Agent
}

func (s *ConsumerPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = mocks.NewMockJobClient(s.mockCtrl)
	s.testAgent = s.newAgent(s.mockCtrl, s.mockJobClient)
}

// newAgent is an agent of the suite's configuration on jobClient.
func (s *ConsumerPublicTestSuite) newAgent(
	ctrl *gomock.Controller,
	jobClient *mocks.MockJobClient,
) *agent.Agent {
	appConfig := config.Config{
		NATS: config.NATS{
			Stream: config.NATSStream{
				Name: "test-stream",
			},
		},
		Agent: config.AgentConfig{
			Hostname:   "test-agent",
			QueueGroup: "test-queue",
			MaxJobs:    5,
			Consumer: config.AgentConsumer{
				AckWait:       "30s",
				BackOff:       []string{"1s", "2s"},
				MaxDeliver:    3,
				MaxAckPending: 10,
				ReplayPolicy:  "instant",
			},
		},
	}

	return newTestAgent(newTestAgentParams{
		appConfig:       appConfig,
		jobClient:       jobClient,
		streamName:      "test-stream",
		hostProvider:    hostMocks.NewDefaultMockProvider(ctrl),
		diskProvider:    diskMocks.NewDefaultMockProvider(ctrl),
		memoryProvider:  memoryMocks.NewDefaultMockProvider(ctrl),
		loadProvider:    loadMocks.NewDefaultMockProvider(ctrl),
		processProvider: processMocks.NewDefaultMockProvider(ctrl),
	})
}

func (s *ConsumerPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *ConsumerPublicTestSuite) TestConsumerNamePrefix() {
	tests := []struct {
		name         string
		consumerName string
		validateFunc func(result string)
	}{
		{
			name:         "when consumer name is configured returns name with underscore",
			consumerName: "my-consumer",
			validateFunc: func(result string) {
				s.Equal("my-consumer_", result)
			},
		},
		{
			name:         "when consumer name is empty returns empty string",
			consumerName: "",
			validateFunc: func(result string) {
				s.Equal("", result)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			cfg := agent.GetAgentAppConfig(s.testAgent)
			cfg.Agent.Consumer.Name = tt.consumerName
			agent.SetAgentAppConfig(s.testAgent, cfg)

			result := agent.ExportConsumerNamePrefix(s.testAgent)
			tt.validateFunc(result)
		})
	}
}

func (s *ConsumerPublicTestSuite) TestConsumeQueryJobs() {
	tests := []struct {
		name         string
		hostname     string
		labels       map[string]string
		setupMocks   func()
		validateFunc func(error)
	}{
		{
			name:     "successful query job consumption",
			hostname: "test-agent",
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
			validateFunc: func(err error) {
				if false {
					s.Error(err)
				} else {
					s.NoError(err)
				}

				// Allow goroutines to execute before cleanup
				time.Sleep(10 * time.Millisecond)
			},
		},
		{
			name:     "consumer creation failure",
			hostname: "test-agent",
			setupMocks: func() {
				// Fail all consumer creations
				s.mockJobClient.EXPECT().
					CreateOrUpdateConsumer(gomock.Any(), "test-stream", gomock.Any()).
					Return(errors.New("consumer creation failed")).
					Times(3)

				// No consumption should happen since creation failed
			},
			validateFunc: func(err error) {
				if false {
					s.Error(err)
				} else {
					s.NoError(err)
				}

				// Allow goroutines to execute before cleanup
				time.Sleep(10 * time.Millisecond)
			},
		},
		{
			name:     "partial consumer creation failure",
			hostname: "test-agent",
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
			validateFunc: func(err error) {
				if false {
					s.Error(err)
				} else {
					s.NoError(err)
				}

				// Allow goroutines to execute before cleanup
				time.Sleep(10 * time.Millisecond)
			},
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
			validateFunc: func(err error) {
				if false {
					s.Error(err)
				} else {
					s.NoError(err)
				}

				// Allow goroutines to execute before cleanup
				time.Sleep(10 * time.Millisecond)
			},
		},
		{
			name:     "consume error logged",
			hostname: "test-agent",
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
			validateFunc: func(err error) {
				if false {
					s.Error(err)
				} else {
					s.NoError(err)
				}

				// Allow goroutines to execute before cleanup
				time.Sleep(10 * time.Millisecond)
			},
		},
		{
			name:     "with labels creates extra consumers",
			hostname: "test-agent",
			labels: map[string]string{
				"group": "web.dev.us-east",
			},
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
			validateFunc: func(err error) {
				if false {
					s.Error(err)
				} else {
					s.NoError(err)
				}

				// Allow goroutines to execute before cleanup
				time.Sleep(10 * time.Millisecond)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			cfg := agent.GetAgentAppConfig(s.testAgent)
			cfg.Agent.Labels = tt.labels
			agent.SetAgentAppConfig(s.testAgent, cfg)

			tt.setupMocks()

			tt.validateFunc(agent.ExportConsumeQueryJobs(ctx, s.testAgent, tt.hostname))
		})
	}
}

func (s *ConsumerPublicTestSuite) TestConsumeModifyJobs() {
	tests := []struct {
		name         string
		hostname     string
		labels       map[string]string
		setupMocks   func()
		validateFunc func(error)
	}{
		{
			name:     "successful modify job consumption",
			hostname: "test-agent",
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
			validateFunc: func(err error) {
				if false {
					s.Error(err)
				} else {
					s.NoError(err)
				}

				// Allow goroutines to execute before cleanup
				time.Sleep(10 * time.Millisecond)
			},
		},
		{
			name:     "consumer creation failure",
			hostname: "test-agent",
			setupMocks: func() {
				// Fail all consumer creations
				s.mockJobClient.EXPECT().
					CreateOrUpdateConsumer(gomock.Any(), "test-stream", gomock.Any()).
					Return(errors.New("consumer creation failed")).
					Times(3)
			},
			validateFunc: func(err error) {
				if false {
					s.Error(err)
				} else {
					s.NoError(err)
				}

				// Allow goroutines to execute before cleanup
				time.Sleep(10 * time.Millisecond)
			},
		},
		{
			name:     "hostname with special characters",
			hostname: "test-agent.domain.com",
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
			validateFunc: func(err error) {
				if false {
					s.Error(err)
				} else {
					s.NoError(err)
				}

				// Allow goroutines to execute before cleanup
				time.Sleep(10 * time.Millisecond)
			},
		},
		{
			name:     "consume error logged",
			hostname: "test-agent",
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
			validateFunc: func(err error) {
				if false {
					s.Error(err)
				} else {
					s.NoError(err)
				}

				// Allow goroutines to execute before cleanup
				time.Sleep(10 * time.Millisecond)
			},
		},
		{
			name:     "with labels creates extra consumers",
			hostname: "test-agent",
			labels: map[string]string{
				"group": "web.dev.us-east",
			},
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
			validateFunc: func(err error) {
				if false {
					s.Error(err)
				} else {
					s.NoError(err)
				}

				// Allow goroutines to execute before cleanup
				time.Sleep(10 * time.Millisecond)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			cfg := agent.GetAgentAppConfig(s.testAgent)
			cfg.Agent.Labels = tt.labels
			agent.SetAgentAppConfig(s.testAgent, cfg)

			tt.setupMocks()

			tt.validateFunc(agent.ExportConsumeModifyJobs(ctx, s.testAgent, tt.hostname))
		})
	}
}

func (s *ConsumerPublicTestSuite) TestCreateConsumer() {
	tests := []struct {
		name          string
		streamName    string
		consumerName  string
		filterSubject string
		agentConsumer config.AgentConsumer
		setupMocks    func()
		validateFunc  func(error)
	}{
		{
			name:          "successful consumer creation with instant replay",
			streamName:    "test-stream",
			consumerName:  "test-consumer",
			filterSubject: "jobs.query._any",
			agentConsumer: config.AgentConsumer{
				AckWait:           "30s",
				BackOff:           []string{"1s", "2s", "5s"},
				MaxDeliver:        3,
				MaxAckPending:     10,
				ReplayPolicy:      "instant",
				InactiveThreshold: "24h",
			},
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					CreateOrUpdateConsumer(gomock.Any(), "test-stream", gomock.Any()).
					Do(func(_ context.Context, _ string, cfg jetstream.ConsumerConfig) {
						s.Equal(24*time.Hour, cfg.InactiveThreshold)
						s.Equal("test-consumer", cfg.Durable)
						s.Equal("jobs.query._any", cfg.FilterSubject)
						s.Equal(jetstream.AckExplicitPolicy, cfg.AckPolicy)
						s.Equal(jetstream.DeliverNewPolicy, cfg.DeliverPolicy)
						s.Equal(3, cfg.MaxDeliver)
						s.Equal(30*time.Second, cfg.AckWait)
						s.Len(cfg.BackOff, 3)
						s.Equal(10, cfg.MaxAckPending)
						s.Equal(jetstream.ReplayInstantPolicy, cfg.ReplayPolicy)
					}).
					Return(nil)
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name:          "successful consumer creation with original replay",
			streamName:    "test-stream",
			consumerName:  "test-consumer-orig",
			filterSubject: "jobs.modify._any",
			agentConsumer: config.AgentConsumer{
				AckWait:       "60s",
				BackOff:       []string{"2s"},
				MaxDeliver:    5,
				MaxAckPending: 20,
				ReplayPolicy:  "original",
			},
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					CreateOrUpdateConsumer(gomock.Any(), "test-stream", gomock.Any()).
					Do(func(_ context.Context, _ string, cfg jetstream.ConsumerConfig) {
						s.Equal("test-consumer-orig", cfg.Durable)
						s.Equal("jobs.modify._any", cfg.FilterSubject)
						s.Equal(60*time.Second, cfg.AckWait)
						s.Len(cfg.BackOff, 1)
						s.Equal(5, cfg.MaxDeliver)
						s.Equal(20, cfg.MaxAckPending)
						s.Equal(jetstream.ReplayOriginalPolicy, cfg.ReplayPolicy)
					}).
					Return(nil)
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name:          "consumer creation failure",
			streamName:    "test-stream",
			consumerName:  "test-consumer",
			filterSubject: "jobs.query._any",
			agentConsumer: config.AgentConsumer{
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
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "stream not found")
			},
		},
		{
			name:          "when inactive_threshold is empty the consumer has no threshold",
			streamName:    "test-stream",
			consumerName:  "test-consumer",
			filterSubject: "jobs.query._all",
			agentConsumer: config.AgentConsumer{
				AckWait:      "30s",
				ReplayPolicy: "instant",
			},
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					CreateOrUpdateConsumer(gomock.Any(), "test-stream", gomock.Any()).
					Do(func(_ context.Context, _ string, cfg jetstream.ConsumerConfig) {
						s.Equal(time.Duration(0), cfg.InactiveThreshold)
					}).
					Return(nil)
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name:          "invalid duration in config",
			streamName:    "test-stream",
			consumerName:  "test-consumer",
			filterSubject: "jobs.query._any",
			agentConsumer: config.AgentConsumer{
				AckWait:           "invalid-duration",
				BackOff:           []string{"invalid", "2s"},
				MaxDeliver:        3,
				MaxAckPending:     10,
				ReplayPolicy:      "instant",
				InactiveThreshold: "invalid",
			},
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					CreateOrUpdateConsumer(gomock.Any(), "test-stream", gomock.Any()).
					Do(func(_ context.Context, _ string, cfg jetstream.ConsumerConfig) {
						// Should have default AckWait (0) and only valid BackOff durations
						s.Equal(time.Duration(0), cfg.AckWait)
						s.Equal(time.Duration(0), cfg.InactiveThreshold)
						s.Len(cfg.BackOff, 1) // Only valid "2s" should be included
					}).
					Return(nil)
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Update agent config for this test
			cfg := agent.GetAgentAppConfig(s.testAgent)
			cfg.Agent.Consumer = tt.agentConsumer
			agent.SetAgentAppConfig(s.testAgent, cfg)

			tt.setupMocks()

			tt.validateFunc(agent.ExportCreateConsumer(
				context.Background(),
				s.testAgent,
				tt.streamName,
				tt.consumerName,
				tt.filterSubject,
			))
		})
	}
}

func (s *ConsumerPublicTestSuite) TestHandleJobMessageJS() {
	tests := []struct {
		name         string
		msgData      []byte
		msgSubject   string
		setupMocks   func()
		validateFunc func(error)
	}{
		{
			name:       "successful message handling",
			msgData:    []byte("test-job-key"),
			msgSubject: "jobs.query.test-agent",
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					HasJobResponse(gomock.Any(), "test-job-key", gomock.Any()).
					Return(false, nil)

				// Mock successful job data retrieval and processing
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.test-job-key").
					Return(signedJob(`{
						"id": "test-job-123",
						"operation": {
							"type": "network.ping.do",
							"data": {"address": "1.1.1.1"}
						}
					}`), nil)

				// Mock status event writes
				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "test-job-key", types.AgentJobStatusAcknowledged, gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "test-job-key", types.AgentJobStatusStarted, gomock.Any(), gomock.Any()).
					Return(nil)

				s.mockJobClient.EXPECT().
					WriteStatusEvent(gomock.Any(), "test-job-key", types.AgentJobStatusCompleted, gomock.Any(), gomock.Any()).
					Return(nil)

				// Mock response write
				s.mockJobClient.EXPECT().
					WriteJobResponse(gomock.Any(), "test-job-key", gomock.Any(), gomock.Any(), types.AgentJobStatusCompleted, "", gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			validateFunc: func(err error) {
				s.NoError(err)
			},
		},
		{
			name:       "job processing failure",
			msgData:    []byte("failed-job-key"),
			msgSubject: "jobs.query.test-agent",
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					HasJobResponse(gomock.Any(), "failed-job-key", gomock.Any()).
					Return(false, nil)

				// Mock job data retrieval failure
				s.mockJobClient.EXPECT().
					GetJobData(gomock.Any(), "jobs.failed-job-key").
					Return(nil, errors.New("job not found"))
			},
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "job not found")
			},
		},
		{
			name:       "invalid subject",
			msgData:    []byte("test-job-key"),
			msgSubject: "invalid.subject",
			setupMocks: func() {
				s.mockJobClient.EXPECT().
					HasJobResponse(gomock.Any(), "test-job-key", gomock.Any()).
					Return(false, nil)
				// No further mocks needed: an unparsable subject is
				// terminated (msg.TermWithReason, stubbed below) rather
				// than retried.
			},
			validateFunc: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "failed to parse subject")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMocks()

			mockMsg := agentmocks.NewMockMsg(s.mockCtrl)
			mockMsg.EXPECT().Subject().Return(tt.msgSubject).AnyTimes()
			mockMsg.EXPECT().Data().Return(tt.msgData).AnyTimes()
			mockMsg.EXPECT().Headers().Return(nil).AnyTimes()
			mockMsg.EXPECT().InProgress().Return(nil).AnyTimes()
			mockMsg.EXPECT().TermWithReason(gomock.Any()).Return(nil).AnyTimes()

			tt.validateFunc(agent.ExportHandleJobMessageJS(s.testAgent, mockMsg))
		})
	}
}

// TestConsumersCondition checks that an agent whose job consumers cannot be
// created or bound reports the ConsumersUnavailable condition, tries again,
// and clears the condition once each consumer runs.
func (s *ConsumerPublicTestSuite) TestConsumersCondition() {
	createFailed := errors.New(
		"nats: API error: code=400 err_code=10026 description=maximum consumers limit reached",
	)
	bindFailed := errors.New("failed to get consumer: nats: consumer not found")

	tests := []struct {
		name string
		// create and consume answer the n-th call, from 0, to the mocks.
		create func(n int32) error
		// consume blocks until ctx ends when it returns nil.
		consume      func(n int32) error
		wantHolds    bool
		wantReason   []string
		wantRecovers bool
	}{
		{
			name:    "when each consumer is created and bound the condition does not hold",
			create:  func(int32) error { return nil },
			consume: func(int32) error { return nil },
		},
		{
			name:       "when a consumer cannot be created the condition holds with the error",
			create:     func(int32) error { return createFailed },
			consume:    func(int32) error { return nil },
			wantHolds:  true,
			wantReason: []string{"3 job consumers do not run", "maximum consumers limit reached"},
		},
		{
			name:       "when a consumer cannot be bound the condition holds with the error",
			create:     func(int32) error { return nil },
			consume:    func(int32) error { return bindFailed },
			wantHolds:  true,
			wantReason: []string{"3 job consumers do not run", "consumer not found"},
		},
		{
			name: "when a failed consumer is created on a retry the condition clears",
			create: func(n int32) error {
				if n < 3 {
					return createFailed
				}

				return nil
			},
			consume:      func(int32) error { return nil },
			wantRecovers: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			defer agent.SetConsumerRetryInterval(20 * time.Millisecond)()
			ctx, cancel := context.WithCancel(context.Background())

			// Each case has its own mocks: the AnyTimes of one case would
			// answer the next.
			ctrl := gomock.NewController(s.T())
			jobClient := mocks.NewMockJobClient(ctrl)
			a := s.newAgent(ctrl, jobClient)

			var creates, consumes, running atomic.Int32
			jobClient.EXPECT().
				CreateOrUpdateConsumer(gomock.Any(), "test-stream", gomock.Any()).
				DoAndReturn(func(context.Context, string, jetstream.ConsumerConfig) error {
					return tt.create(creates.Add(1) - 1)
				}).
				AnyTimes()
			jobClient.EXPECT().
				ConsumeJobs(gomock.Any(), "test-stream", gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(
					ctx context.Context,
					_, _ string,
					_ natsclient.JetStreamMessageHandler,
					_ *natsclient.ConsumeOptions,
				) error {
					if err := tt.consume(consumes.Add(1) - 1); err != nil {
						return err
					}
					running.Add(1)
					<-ctx.Done()

					return ctx.Err()
				}).
				AnyTimes()

			s.NoError(agent.ExportConsumeQueryJobs(ctx, a, "test-agent"))

			switch {
			case tt.wantRecovers:
				s.Eventually(func() bool { return running.Load() == 3 },
					2*time.Second, 10*time.Millisecond, "each consumer runs after a retry")
				s.False(agent.ExportConsumersCondition(a).Status)
			case tt.wantHolds:
				// The goroutines record a failed bind one by one, so the
				// reason settles once each has.
				s.Eventually(func() bool {
					c := agent.ExportConsumersCondition(a)
					if !c.Status || c.Type != types.ConditionConsumersUnavailable {
						return false
					}
					for _, want := range tt.wantReason {
						if !strings.Contains(c.Reason, want) {
							return false
						}
					}

					return true
				}, 2*time.Second, 10*time.Millisecond, "the condition holds with the error")
				// The agent tries again.
				before := creates.Load()
				s.Eventually(func() bool { return creates.Load() > before },
					2*time.Second, 10*time.Millisecond, "the agent tries again")
			default:
				s.Eventually(func() bool { return running.Load() == 3 },
					2*time.Second, 10*time.Millisecond)
				c := agent.ExportConsumersCondition(a)
				s.False(c.Status)
				s.Empty(c.Reason)
			}

			// A cancel, such as a drain, stops the retries and clears the
			// condition: a stopped consumer is not a failed one.
			cancel()
			s.Eventually(func() bool {
				return !agent.ExportConsumersCondition(a).Status
			}, 2*time.Second, 10*time.Millisecond, "the condition clears on cancel")
			agent.ExportWaitConsumers(a)
		})
	}
}

func TestConsumerPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ConsumerPublicTestSuite))
}

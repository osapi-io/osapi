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

package agent_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	agentMocks "github.com/osapi-io/osapi/internal/agent/mocks"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/agent"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/job/mocks"
	commandMocks "github.com/osapi-io/osapi/internal/provider/command/mocks"
	pingMocks "github.com/osapi-io/osapi/internal/provider/network/ping/mocks"
	diskMocks "github.com/osapi-io/osapi/internal/provider/node/disk/mocks"
	hostMocks "github.com/osapi-io/osapi/internal/provider/node/host/mocks"
	loadMocks "github.com/osapi-io/osapi/internal/provider/node/load/mocks"
	memoryMocks "github.com/osapi-io/osapi/internal/provider/node/memory/mocks"
	processMocks "github.com/osapi-io/osapi/internal/telemetry/process/mocks"
)

type DrainPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *mocks.MockJobClient
	mockKV        *agentMocks.MockNATSClient
	mockEntry     *mocks.MockKeyValueEntry
	testAgent     *agent.Agent
	// cordoned closes when the cordoned timeline event is written.
	cordoned chan struct{}
}

func (s *DrainPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = mocks.NewMockJobClient(s.mockCtrl)
	s.mockKV = agentMocks.NewMockNATSClient(s.mockCtrl)
	s.mockEntry = mocks.NewMockKeyValueEntry(s.mockCtrl)
	s.cordoned = make(chan struct{})

	appConfig := config.Config{
		Agent: config.AgentConfig{
			Labels: map[string]string{"group": "web"},
		},
	}

	s.testAgent = newTestAgent(newTestAgentParams{
		appConfig:       appConfig,
		jobClient:       s.mockJobClient,
		streamName:      "test-stream",
		hostProvider:    hostMocks.NewDefaultMockProvider(s.mockCtrl),
		diskProvider:    diskMocks.NewDefaultMockProvider(s.mockCtrl),
		memoryProvider:  memoryMocks.NewDefaultMockProvider(s.mockCtrl),
		loadProvider:    loadMocks.NewDefaultMockProvider(s.mockCtrl),
		pingProvider:    pingMocks.NewDefaultMockProvider(s.mockCtrl),
		commandProvider: commandMocks.NewDefaultMockProvider(s.mockCtrl),
		processProvider: processMocks.NewDefaultMockProvider(s.mockCtrl),
		registryBucket:  "registry",
		natsClient:      s.mockKV,
	})
	agent.SetAgentState(s.testAgent, job.AgentStateReady)
	agent.SetAgentMachineID(s.testAgent, "test-machine-id")
	ctx, cancel := context.WithCancel(context.Background())
	consumerCtx, consumerCancel := context.WithCancel(ctx)
	agent.SetAgentLifecycle(ctx, consumerCtx, s.testAgent, cancel, consumerCancel)
}

func (s *DrainPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *DrainPublicTestSuite) TestCheckDrainFlag() {
	tests := []struct {
		name         string
		setupMock    func()
		validateFunc func(bool)
	}{
		{
			name: "when drain key exists returns true",
			setupMock: func() {
				s.mockJobClient.EXPECT().
					CheckDrainFlag(gomock.Any(), "test-machine-id").
					Return(true)
			},
			validateFunc: func(result bool) {
				s.True(result)
			},
		},
		{
			name: "when drain key missing returns false",
			setupMock: func() {
				s.mockJobClient.EXPECT().
					CheckDrainFlag(gomock.Any(), "test-machine-id").
					Return(false)
			},
			validateFunc: func(result bool) {
				s.False(result)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()
			result := agent.ExportCheckDrainFlag(
				context.Background(),
				s.testAgent,
				"test-machine-id",
			)
			tt.validateFunc(result)
		})
	}
}

// TestDrainAtStartupThenUndrainThenDrain replays the hang of #876. An
// agent starts with its drain flag set: the heartbeat finds it before
// startup starts the consumers. Startup must then start none, so that the
// undrain starts the only set and the next drain stops it. Before the fix,
// startup started a set while cordoned, the undrain started a second one
// over it, and the drain waited for the first set forever.
func (s *DrainPublicTestSuite) TestDrainAtStartupThenUndrainThenDrain() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// No consumers yet: startup has not started them.
	agent.SetAgentLifecycle(ctx, nil, s.testAgent, cancel, nil)
	agent.SetAgentState(s.testAgent, job.AgentStateReady)

	var started atomic.Int32
	s.mockJobClient.EXPECT().
		CreateOrUpdateConsumer(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()
	// A consumer runs until its context is cancelled, as nats-client's does.
	s.mockJobClient.EXPECT().
		ConsumeJobs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, _, _ string, _ any, _ any) error {
			started.Add(1)
			<-ctx.Done()
			return ctx.Err()
		}).
		AnyTimes()
	s.mockJobClient.EXPECT().
		WriteAgentTimelineEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()

	// The heartbeat finds the drain flag first.
	s.mockJobClient.EXPECT().CheckDrainFlag(gomock.Any(), "test-machine-id").Return(true)
	agent.ExportHandleDrainDetection(ctx, s.testAgent, "test-machine-id", "test-agent")
	s.Eventually(func() bool { return s.testAgent.State() == job.AgentStateCordoned },
		time.Second, 10*time.Millisecond)

	// Startup then starts no consumers, because the agent is cordoned.
	agent.ExportStartConsumers(s.testAgent)
	s.Never(func() bool { return started.Load() > 0 }, 200*time.Millisecond, 20*time.Millisecond,
		"a cordoned agent takes no jobs")

	// The undrain starts the one set.
	s.mockJobClient.EXPECT().CheckDrainFlag(gomock.Any(), "test-machine-id").Return(false)
	agent.ExportHandleDrainDetection(ctx, s.testAgent, "test-machine-id", "test-agent")
	s.Equal(job.AgentStateReady, s.testAgent.State())
	s.Eventually(func() bool { return started.Load() > 0 }, time.Second, 10*time.Millisecond)

	// A set starts each consumer on its own goroutine: wait until all of
	// them have started, so the count is the whole set.
	n := settled(&started)

	// A second start while it runs adds no second set.
	agent.ExportStartConsumers(s.testAgent)
	s.Never(func() bool { return started.Load() > n }, 200*time.Millisecond, 20*time.Millisecond)

	// The next drain stops it, and returns.
	s.mockJobClient.EXPECT().CheckDrainFlag(gomock.Any(), "test-machine-id").Return(true)
	done := make(chan struct{})
	go func() {
		agent.ExportHandleDrainDetection(ctx, s.testAgent, "test-machine-id", "test-agent")
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		s.FailNow("the drain did not return: a consumer set was left running")
	}
	s.Eventually(func() bool { return s.testAgent.State() == job.AgentStateCordoned },
		3*time.Second, 10*time.Millisecond, "a consumer set was left running")
}

// TestDrainWithAStuckConsumer checks that a consumer that does not stop
// never holds the heartbeat: the drain returns at once in Draining, the
// agent reports DrainStuck once the limit passes, and it reaches Cordoned
// and clears the condition when the consumer stops at last.
func (s *DrainPublicTestSuite) TestDrainWithAStuckConsumer() {
	defer agent.SetDrainStuckAfter(100 * time.Millisecond)()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	agent.SetAgentLifecycle(ctx, nil, s.testAgent, cancel, nil)
	agent.SetAgentState(s.testAgent, job.AgentStateReady)

	release := make(chan struct{})
	var started atomic.Int32
	s.mockJobClient.EXPECT().
		CreateOrUpdateConsumer(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()
	// A consumer that ignores its cancel until released.
	s.mockJobClient.EXPECT().
		ConsumeJobs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(context.Context, string, string, any, any) error {
			started.Add(1)
			<-release
			return nil
		}).
		AnyTimes()
	s.mockJobClient.EXPECT().
		WriteAgentTimelineEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()

	agent.ExportStartConsumers(s.testAgent)
	s.Eventually(func() bool { return started.Load() > 0 }, time.Second, 10*time.Millisecond)
	settled(&started)

	s.mockJobClient.EXPECT().CheckDrainFlag(gomock.Any(), "test-machine-id").Return(true)
	done := make(chan struct{})
	go func() {
		agent.ExportHandleDrainDetection(ctx, s.testAgent, "test-machine-id", "test-agent")
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		s.FailNow("the drain held the heartbeat while a consumer would not stop")
	}
	s.Equal(job.AgentStateDraining, s.testAgent.State())
	s.False(agent.ExportDrainCondition(s.testAgent).Status)

	s.Eventually(func() bool { return agent.ExportDrainCondition(s.testAgent).Status },
		2*time.Second, 20*time.Millisecond, "the agent reports DrainStuck")
	c := agent.ExportDrainCondition(s.testAgent)
	s.Equal(job.ConditionDrainStuck, c.Type)
	s.NotEmpty(c.Reason)
	s.Equal(job.AgentStateDraining, s.testAgent.State())

	close(release)
	s.Eventually(func() bool { return s.testAgent.State() == job.AgentStateCordoned },
		2*time.Second, 10*time.Millisecond)
	s.False(agent.ExportDrainCondition(s.testAgent).Status, "the condition clears")
}

func (s *DrainPublicTestSuite) TestHandleDrainDetection() {
	tests := []struct {
		name         string
		initialState string
		setupMock    func()
		validateFunc func(string)
	}{
		{
			name:         "when drain flag set and agent is Ready moves to Draining, then Cordoned",
			initialState: job.AgentStateReady,
			setupMock: func() {
				s.mockJobClient.EXPECT().
					CheckDrainFlag(gomock.Any(), "test-machine-id").
					Return(true)
				s.mockJobClient.EXPECT().
					WriteAgentTimelineEvent(
						gomock.Any(),
						"test-agent",
						"drain",
						"Drain initiated",
					).
					Return(nil)
				s.mockJobClient.EXPECT().
					WriteAgentTimelineEvent(
						gomock.Any(),
						"test-agent",
						"cordoned",
						"All jobs completed",
					).
					DoAndReturn(func(context.Context, string, string, string) error {
						close(s.cordoned)
						return nil
					})
			},
			validateFunc: func(got string) {
				// The heartbeat does not wait for the consumers.
				s.Equal(job.AgentStateDraining, got)
				select {
				case <-s.cordoned:
				case <-time.After(3 * time.Second):
					s.FailNow("the drain did not finish")
				}
				s.Equal(job.AgentStateCordoned, s.testAgent.State())
			},
		},
		{
			name:         "when drain flag removed and agent is still Draining waits for the drain",
			initialState: job.AgentStateDraining,
			setupMock: func() {
				s.mockJobClient.EXPECT().
					CheckDrainFlag(gomock.Any(), "test-machine-id").
					Return(false)
			},
			validateFunc: func(got string) {
				// The undrain applies once the drain has reached Cordoned.
				s.Equal(job.AgentStateDraining, got)
			},
		},
		{
			name:         "when drain flag removed and agent is Cordoned transitions to Ready",
			initialState: job.AgentStateCordoned,
			setupMock: func() {
				s.mockJobClient.EXPECT().
					CheckDrainFlag(gomock.Any(), "test-machine-id").
					Return(false)
				s.mockJobClient.EXPECT().
					WriteAgentTimelineEvent(
						gomock.Any(),
						"test-agent",
						"undrain",
						"Resumed accepting jobs",
					).
					Return(nil)
				// startConsumers re-creates consumers
				s.mockJobClient.EXPECT().
					CreateOrUpdateConsumer(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil).
					AnyTimes()
				s.mockJobClient.EXPECT().
					ConsumeJobs(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(context.Canceled).
					AnyTimes()
			},
			validateFunc: func(got string) {
				s.Equal(job.AgentStateReady, got)
			},
		},
		{
			name:         "when drain flag still set and agent is already Draining stays Draining",
			initialState: job.AgentStateDraining,
			setupMock: func() {
				s.mockJobClient.EXPECT().
					CheckDrainFlag(gomock.Any(), "test-machine-id").
					Return(true)
			},
			validateFunc: func(got string) {
				s.Equal(job.AgentStateDraining, got)
			},
		},
		{
			name:         "when no drain flag and agent is Ready stays Ready",
			initialState: job.AgentStateReady,
			setupMock: func() {
				s.mockJobClient.EXPECT().
					CheckDrainFlag(gomock.Any(), "test-machine-id").
					Return(false)
			},
			validateFunc: func(got string) {
				s.Equal(job.AgentStateReady, got)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			agent.SetAgentState(s.testAgent, tt.initialState)
			tt.setupMock()
			agent.ExportHandleDrainDetection(
				context.Background(),
				s.testAgent,
				"test-machine-id",
				"test-agent",
			)
			tt.validateFunc(agent.GetAgentState(s.testAgent))
		})
	}
}

func TestDrainPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(DrainPublicTestSuite))
}

// settled is the value of n once it has not changed for 100ms.
func settled(
	n *atomic.Int32,
) int32 {
	last := n.Load()
	for {
		time.Sleep(100 * time.Millisecond)
		cur := n.Load()
		if cur == last {
			return cur
		}
		last = cur
	}
}

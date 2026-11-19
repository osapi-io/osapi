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
	netinfoMocks "github.com/osapi-io/osapi/internal/provider/network/netinfo/mocks"
	dnsMocks "github.com/osapi-io/osapi/internal/provider/network/netplan/dns/mocks"
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
}

func (s *DrainPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = mocks.NewMockJobClient(s.mockCtrl)
	s.mockKV = agentMocks.NewMockNATSClient(s.mockCtrl)
	s.mockEntry = mocks.NewMockKeyValueEntry(s.mockCtrl)

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
		dnsProvider:     dnsMocks.NewDefaultMockProvider(s.mockCtrl),
		pingProvider:    pingMocks.NewDefaultMockProvider(s.mockCtrl),
		netinfoProvider: netinfoMocks.NewDefaultMockProvider(s.mockCtrl),
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
	s.Equal(job.AgentStateCordoned, s.testAgent.State())

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
	s.Equal(job.AgentStateCordoned, s.testAgent.State())
}

func (s *DrainPublicTestSuite) TestHandleDrainDetection() {
	tests := []struct {
		name         string
		initialState string
		setupMock    func()
		validateFunc func(string)
	}{
		{
			name:         "when drain flag set and agent is Ready transitions to Cordoned",
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
					Return(nil)
			},
			validateFunc: func(got string) {
				s.Equal(job.AgentStateCordoned, got)
			},
		},
		{
			name:         "when drain flag removed and agent is Draining transitions to Ready",
			initialState: job.AgentStateDraining,
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

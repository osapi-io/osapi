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
	"sync"
	"testing"

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

// LifecyclePublicTestSuite covers the agent's shared lifecycle state under
// concurrent access. These tests are about the absence of a data race rather
// than about a return value, so they are meaningful only with -race, which CI
// runs.
type LifecyclePublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *mocks.MockJobClient
	mockKV        *mocks.MockKeyValue
	testAgent     *agent.Agent
}

func (s *LifecyclePublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = mocks.NewMockJobClient(s.mockCtrl)
	s.mockKV = mocks.NewMockKeyValue(s.mockCtrl)

	s.testAgent = newTestAgent(newTestAgentParams{
		appConfig:       config.Config{},
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
		registryKV:      s.mockKV,
	})
	agent.SetAgentState(s.testAgent, job.AgentStateReady)
	agent.SetAgentMachineID(s.testAgent, "test-machine-id")

	ctx, cancel := context.WithCancel(context.Background())
	consumerCtx, consumerCancel := context.WithCancel(ctx)
	agent.SetAgentLifecycle(ctx, consumerCtx, s.testAgent, cancel, consumerCancel)
}

func (s *LifecyclePublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

// TestStateUnderConcurrentAccess runs the drain check against a state another
// goroutine is writing, which is what the heartbeat goroutine and an enrollment
// acceptance do to each other.
func (s *LifecyclePublicTestSuite) TestStateUnderConcurrentAccess() {
	s.mockJobClient.EXPECT().
		CheckDrainFlag(gomock.Any(), "test-machine-id").
		Return(false).
		AnyTimes()

	const iterations = 50

	var wg sync.WaitGroup

	wg.Add(3)

	go func() {
		defer wg.Done()

		for range iterations {
			agent.ExportHandleDrainDetection(
				context.Background(),
				s.testAgent,
				"test-machine-id",
				"test-agent",
			)
		}
	}()

	go func() {
		defer wg.Done()

		for range iterations {
			agent.SetAgentState(s.testAgent, job.AgentStateReady)
		}
	}()

	go func() {
		defer wg.Done()

		for range iterations {
			s.NotEmpty(agent.GetAgentState(s.testAgent))
		}
	}()

	wg.Wait()

	s.Equal(job.AgentStateReady, agent.GetAgentState(s.testAgent))
}

// TestCachedFactsUnderConcurrentAccess replaces the cached facts while another
// goroutine reads them, which is what the facts refresh does to a job in flight.
func (s *LifecyclePublicTestSuite) TestCachedFactsUnderConcurrentAccess() {
	const iterations = 50

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()

		for i := range iterations {
			agent.SetAgentCachedFacts(s.testAgent, &job.FactsRegistration{
				FQDN:     "test-agent",
				CPUCount: i,
			})
		}
	}()

	go func() {
		defer wg.Done()

		for range iterations {
			// Nil until the first write lands, which is itself a valid answer.
			_ = s.testAgent.GetFacts()
		}
	}()

	wg.Wait()

	s.NotNil(agent.GetAgentCachedFacts(s.testAgent))
}

// In order for `go test` to run this suite, we need to create
// a normal test function and pass our suite to suite.Run.
func TestLifecyclePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(LifecyclePublicTestSuite))
}

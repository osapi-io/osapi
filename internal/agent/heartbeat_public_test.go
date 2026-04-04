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
	"errors"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/agent"
	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/job/mocks"
	commandMocks "github.com/retr0h/osapi/internal/provider/command/mocks"
	netinfoMocks "github.com/retr0h/osapi/internal/provider/network/netinfo/mocks"
	dnsMocks "github.com/retr0h/osapi/internal/provider/network/netplan/dns/mocks"
	pingMocks "github.com/retr0h/osapi/internal/provider/network/ping/mocks"
	diskMocks "github.com/retr0h/osapi/internal/provider/node/disk/mocks"
	hostMocks "github.com/retr0h/osapi/internal/provider/node/host/mocks"
	loadMocks "github.com/retr0h/osapi/internal/provider/node/load/mocks"
	memMocks "github.com/retr0h/osapi/internal/provider/node/mem/mocks"
	processMocks "github.com/retr0h/osapi/internal/telemetry/process/mocks"
)

type HeartbeatPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *mocks.MockJobClient
	mockKV        *mocks.MockKeyValue
	appFs         avfs.VFS
	appConfig     config.Config
	logger        *slog.Logger
}

func (s *HeartbeatPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = mocks.NewMockJobClient(s.mockCtrl)
	s.mockKV = mocks.NewMockKeyValue(s.mockCtrl)
	s.appFs = memfs.New()
	s.logger = slog.Default()

	s.appConfig = config.Config{
		NATS: config.NATS{
			Stream: config.NATSStream{Name: "test-stream"},
			Registry: config.NATSRegistry{
				Bucket:   "agent-registry",
				TTL:      "30s",
				Storage:  "file",
				Replicas: 1,
			},
		},
		Agent: config.AgentConfig{
			Hostname:   "test-agent",
			QueueGroup: "test-queue",
			MaxJobs:    5,
			Labels:     map[string]string{"group": "web"},
			Consumer: config.AgentConsumer{
				AckWait:       "30s",
				BackOff:       []string{"1s", "2s", "5s"},
				MaxDeliver:    3,
				MaxAckPending: 10,
				ReplayPolicy:  "instant",
			},
		},
	}
}

func (s *HeartbeatPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *HeartbeatPublicTestSuite) TestStartWithHeartbeat() {
	tests := []struct {
		name      string
		setupFunc func() *agent.Agent
		stopFunc  func(a *agent.Agent)
	}{
		{
			name: "when registryKV is set registers and deregisters",
			setupFunc: func() *agent.Agent {
				// Drain check on each heartbeat tick (no drain flag present)
				s.mockJobClient.EXPECT().
					CheckDrainFlag(gomock.Any(), "test-agent").
					Return(false).
					AnyTimes()

				// Heartbeat initial write
				s.mockKV.EXPECT().
					Put(gomock.Any(), "agents.test_agent", gomock.Any()).
					Return(uint64(1), nil).
					MinTimes(1)

				// Deregister on stop
				s.mockKV.EXPECT().
					Delete(gomock.Any(), "agents.test_agent").
					Return(nil).
					Times(1)

					// 3 base + 1 label = 4 consumers per job type, x2 (query+modify) = 8
				s.mockJobClient.EXPECT().
					CreateOrUpdateConsumer(gomock.Any(), "test-stream", gomock.Any()).
					Return(nil).
					Times(8)

				s.mockJobClient.EXPECT().
					ConsumeJobs(gomock.Any(), "test-stream", gomock.Any(), gomock.Any(), gomock.Any()).
					Return(context.Canceled).
					Times(8)

				return newTestAgent(newTestAgentParams{
					appFs:           s.appFs,
					appConfig:       s.appConfig,
					logger:          s.logger,
					jobClient:       s.mockJobClient,
					streamName:      "test-stream",
					hostProvider:    hostMocks.NewDefaultMockProvider(s.mockCtrl),
					diskProvider:    diskMocks.NewDefaultMockProvider(s.mockCtrl),
					memProvider:     memMocks.NewDefaultMockProvider(s.mockCtrl),
					loadProvider:    loadMocks.NewDefaultMockProvider(s.mockCtrl),
					dnsProvider:     dnsMocks.NewDefaultMockProvider(s.mockCtrl),
					pingProvider:    pingMocks.NewDefaultMockProvider(s.mockCtrl),
					netinfoProvider: netinfoMocks.NewDefaultMockProvider(s.mockCtrl),
					commandProvider: commandMocks.NewDefaultMockProvider(s.mockCtrl),
					processProvider: processMocks.NewDefaultMockProvider(s.mockCtrl),
					registryKV:      s.mockKV,
				})
			},
			stopFunc: func(a *agent.Agent) {
				stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				a.Stop(stopCtx)
			},
		},
		{
			name: "when registryKV is nil skips heartbeat",
			setupFunc: func() *agent.Agent {
				// 3 base + 1 label = 4 consumers per job type, x2 (query+modify) = 8
				s.mockJobClient.EXPECT().
					CreateOrUpdateConsumer(gomock.Any(), "test-stream", gomock.Any()).
					Return(nil).
					Times(8)

				s.mockJobClient.EXPECT().
					ConsumeJobs(gomock.Any(), "test-stream", gomock.Any(), gomock.Any(), gomock.Any()).
					Return(context.Canceled).
					Times(8)

				return newTestAgent(newTestAgentParams{
					appFs:           s.appFs,
					appConfig:       s.appConfig,
					logger:          s.logger,
					jobClient:       s.mockJobClient,
					streamName:      "test-stream",
					hostProvider:    hostMocks.NewDefaultMockProvider(s.mockCtrl),
					diskProvider:    diskMocks.NewDefaultMockProvider(s.mockCtrl),
					memProvider:     memMocks.NewDefaultMockProvider(s.mockCtrl),
					loadProvider:    loadMocks.NewDefaultMockProvider(s.mockCtrl),
					dnsProvider:     dnsMocks.NewDefaultMockProvider(s.mockCtrl),
					pingProvider:    pingMocks.NewDefaultMockProvider(s.mockCtrl),
					netinfoProvider: netinfoMocks.NewDefaultMockProvider(s.mockCtrl),
					commandProvider: commandMocks.NewDefaultMockProvider(s.mockCtrl),
				})
			},
			stopFunc: func(a *agent.Agent) {
				stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				a.Stop(stopCtx)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			a := tt.setupFunc()
			a.Start()
			tt.stopFunc(a)
		})
	}
}

func TestHeartbeatPublicTestSuite(t *testing.T) {
	suite.Run(t, new(HeartbeatPublicTestSuite))
}

// HeartbeatLowLevelPublicTestSuite tests the lower-level heartbeat methods.
type HeartbeatLowLevelPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *mocks.MockJobClient
	mockKV        *mocks.MockKeyValue
	testAgent     *agent.Agent
}

func (s *HeartbeatLowLevelPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = mocks.NewMockJobClient(s.mockCtrl)
	s.mockKV = mocks.NewMockKeyValue(s.mockCtrl)

	appConfig := config.Config{
		Agent: config.AgentConfig{
			Labels: map[string]string{"group": "web"},
		},
	}

	// Use DefaultMockProviders so provider calls during writeRegistration are satisfied.
	s.testAgent = newTestAgent(newTestAgentParams{
		appConfig:       appConfig,
		jobClient:       s.mockJobClient,
		streamName:      "test-stream",
		hostProvider:    hostMocks.NewDefaultMockProvider(s.mockCtrl),
		diskProvider:    diskMocks.NewDefaultMockProvider(s.mockCtrl),
		memProvider:     memMocks.NewDefaultMockProvider(s.mockCtrl),
		loadProvider:    loadMocks.NewDefaultMockProvider(s.mockCtrl),
		dnsProvider:     dnsMocks.NewDefaultMockProvider(s.mockCtrl),
		pingProvider:    pingMocks.NewDefaultMockProvider(s.mockCtrl),
		netinfoProvider: netinfoMocks.NewDefaultMockProvider(s.mockCtrl),
		commandProvider: commandMocks.NewDefaultMockProvider(s.mockCtrl),
		processProvider: processMocks.NewDefaultMockProvider(s.mockCtrl),
		registryKV:      s.mockKV,
	})
	agent.SetAgentState(s.testAgent, job.AgentStateReady)

	// writeRegistration now calls handleDrainDetection which checks drain flag.
	// Default: no drain flag present.
	s.mockJobClient.EXPECT().
		CheckDrainFlag(gomock.Any(), "test-agent").
		Return(false).
		AnyTimes()
}

func (s *HeartbeatLowLevelPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
	agent.ResetMarshalJSON()
	agent.ResetHeartbeatInterval()
}

func (s *HeartbeatLowLevelPublicTestSuite) TestWriteRegistration() {
	tests := []struct {
		name         string
		setupMock    func()
		teardownMock func()
	}{
		{
			name: "when marshal fails logs warning",
			setupMock: func() {
				agent.SetMarshalJSON(func(_ interface{}) ([]byte, error) {
					return nil, fmt.Errorf("marshal failure")
				})
			},
			teardownMock: func() {
				agent.ResetMarshalJSON()
			},
		},
		{
			name: "when Put fails logs warning",
			setupMock: func() {
				s.mockKV.EXPECT().
					Put(gomock.Any(), "agents.test_agent", gomock.Any()).
					Return(uint64(0), errors.New("put failed"))
			},
		},
		{
			name: "when Put succeeds writes registration",
			setupMock: func() {
				s.mockKV.EXPECT().
					Put(gomock.Any(), "agents.test_agent", gomock.Any()).
					Return(uint64(1), nil)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()
			if tt.teardownMock != nil {
				defer tt.teardownMock()
			}
			agent.ExportWriteRegistration(context.Background(), s.testAgent, "test-agent")
		})
	}
}

func (s *HeartbeatLowLevelPublicTestSuite) TestWriteRegistrationStoresHeartbeatTime() {
	tests := []struct {
		name string
	}{
		{
			name: "when Put succeeds stores last heartbeat time",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.mockKV.EXPECT().
				Put(gomock.Any(), "agents.test_agent", gomock.Any()).
				Return(uint64(1), nil)

			before := time.Now()
			agent.ExportWriteRegistration(context.Background(), s.testAgent, "test-agent")
			after := time.Now()

			got := s.testAgent.LastHeartbeatTime()
			s.False(got.IsZero(), "expected non-zero heartbeat time after successful Put")
			s.True(
				!got.Before(before) && !got.After(after),
				"heartbeat time should be between before and after write",
			)
		})
	}
}

func (s *HeartbeatLowLevelPublicTestSuite) TestDeregister() {
	tests := []struct {
		name      string
		setupMock func()
	}{
		{
			name: "when Delete fails logs warning",
			setupMock: func() {
				s.mockKV.EXPECT().
					Delete(gomock.Any(), "agents.test_agent").
					Return(errors.New("delete failed"))
			},
		},
		{
			name: "when Delete succeeds logs deregistration",
			setupMock: func() {
				s.mockKV.EXPECT().
					Delete(gomock.Any(), "agents.test_agent").
					Return(nil)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()
			agent.ExportDeregister(s.testAgent, "test-agent")
		})
	}
}

func (s *HeartbeatLowLevelPublicTestSuite) TestStartHeartbeatRefresh() {
	tests := []struct {
		name      string
		setupMock func()
	}{
		{
			name: "ticker fires and refreshes registration",
			setupMock: func() {
				// Initial write + at least 1 ticker refresh
				s.mockKV.EXPECT().
					Put(gomock.Any(), "agents.test_agent", gomock.Any()).
					Return(uint64(1), nil).
					MinTimes(2)

				// Deregister on cancel
				s.mockKV.EXPECT().
					Delete(gomock.Any(), "agents.test_agent").
					Return(nil).
					Times(1)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()

			agent.SetHeartbeatInterval(10 * time.Millisecond)

			ctx, cancel := context.WithCancel(context.Background())
			agent.ExportStartHeartbeat(ctx, s.testAgent, "test-agent")

			// Wait for at least one ticker refresh
			time.Sleep(50 * time.Millisecond)
			cancel()

			// Wait for goroutine to finish
			agent.WaitAgentWG(s.testAgent)
		})
	}
}

func (s *HeartbeatLowLevelPublicTestSuite) TestRegistryKey() {
	tests := []struct {
		name     string
		hostname string
		expected string
	}{
		{
			name:     "simple hostname",
			hostname: "web-01",
			expected: "agents.web_01",
		},
		{
			name:     "hostname with dots",
			hostname: "Johns-MacBook-Pro.local",
			expected: "agents.Johns_MacBook_Pro_local",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := agent.ExportRegistryKey(tt.hostname)
			s.Equal(tt.expected, result)
		})
	}
}

func TestHeartbeatLowLevelPublicTestSuite(t *testing.T) {
	suite.Run(t, new(HeartbeatLowLevelPublicTestSuite))
}

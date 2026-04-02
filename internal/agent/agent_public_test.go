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
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/agent"
	"github.com/retr0h/osapi/internal/config"
	execmocks "github.com/retr0h/osapi/internal/exec/mocks"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/job/mocks"
	commandMocks "github.com/retr0h/osapi/internal/provider/command/mocks"
	dnsMocks "github.com/retr0h/osapi/internal/provider/network/dns/mocks"
	netinfoMocks "github.com/retr0h/osapi/internal/provider/network/netinfo/mocks"
	pingMocks "github.com/retr0h/osapi/internal/provider/network/ping/mocks"
	diskMocks "github.com/retr0h/osapi/internal/provider/node/disk/mocks"
	hostMocks "github.com/retr0h/osapi/internal/provider/node/host/mocks"
	loadMocks "github.com/retr0h/osapi/internal/provider/node/load/mocks"
	memMocks "github.com/retr0h/osapi/internal/provider/node/mem/mocks"
	"github.com/retr0h/osapi/internal/telemetry/metrics"
	processMocks "github.com/retr0h/osapi/internal/telemetry/process/mocks"
)

type AgentPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *mocks.MockJobClient
	appFs         avfs.VFS
	appConfig     config.Config
	logger        *slog.Logger
}

func (s *AgentPublicTestSuite) getFreePort() int {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	s.Require().NoError(err)
	defer func() { _ = l.Close() }()

	return l.Addr().(*net.TCPAddr).Port
}

func (s *AgentPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = mocks.NewMockJobClient(s.mockCtrl)
	s.appFs = memfs.New()
	s.logger = slog.Default()

	s.appConfig = config.Config{
		NATS: config.NATS{
			Stream: config.NATSStream{Name: "test-stream"},
		},
		Agent: config.AgentConfig{
			Hostname:   "test-agent",
			QueueGroup: "test-queue",
			MaxJobs:    5,
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

func (s *AgentPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *AgentPublicTestSuite) buildAgent() *agent.Agent {
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
	})
}

func (s *AgentPublicTestSuite) TestNew() {
	tests := []struct {
		name string
	}{
		{
			name: "creates agent with all providers",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			a := s.buildAgent()

			s.NotNil(a)

			a.SetSubComponents(map[string]job.SubComponentInfo{
				"agent.heartbeat": {Status: "ok"},
			})
		})
	}
}

func (s *AgentPublicTestSuite) TestStart() {
	tests := []struct {
		name      string
		setupFunc func() *agent.Agent
		stopFunc  func(a *agent.Agent)
	}{
		{
			name: "starts and stops gracefully",
			setupFunc: func() *agent.Agent {
				s.mockJobClient.EXPECT().
					CreateOrUpdateConsumer(gomock.Any(), "test-stream", gomock.Any()).
					Return(nil).
					Times(6)

				s.mockJobClient.EXPECT().
					ConsumeJobs(gomock.Any(), "test-stream", gomock.Any(), gomock.Any(), gomock.Any()).
					Return(context.Canceled).
					Times(6)

				return s.buildAgent()
			},
			stopFunc: func(a *agent.Agent) {
				stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				a.Stop(stopCtx)
			},
		},
		{
			name: "returns early when preflight fails",
			setupFunc: func() *agent.Agent {
				mockExecMgr := execmocks.NewMockManager(s.mockCtrl)
				mockExecMgr.EXPECT().
					RunCmd("sudo", gomock.Any()).
					Return("", fmt.Errorf("sudo: a password is required")).
					AnyTimes()

				cfg := s.appConfig
				cfg.Agent.PrivilegeEscalation = config.PrivilegeEscalation{
					Enabled: true,
				}

				return newTestAgent(newTestAgentParams{
					appFs:           s.appFs,
					appConfig:       cfg,
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
					execManager:     mockExecMgr,
				})
			},
			stopFunc: func(a *agent.Agent) {
				// Agent should not have started consumers, so IsReady fails.
				err := a.IsReady()
				s.Error(err)
			},
		},
		{
			name: "starts when preflight passes",
			setupFunc: func() *agent.Agent {
				mockExecMgr := execmocks.NewMockManager(s.mockCtrl)
				mockExecMgr.EXPECT().
					RunCmd("sudo", gomock.Any()).
					Return("/usr/bin/something", nil).
					AnyTimes()

				// Write a fake proc status file with all capabilities set.
				tmpDir := s.T().TempDir()
				path := filepath.Join(tmpDir, "status")
				err := os.WriteFile(path, []byte("Name:\tosapi\nCapEff:\t000000000000003f\n"), 0o644)
				s.Require().NoError(err)
				agent.SetProcStatusPath(path)
				s.T().Cleanup(func() {
					agent.ResetProcStatusPath()
				})

				cfg := s.appConfig
				cfg.Agent.PrivilegeEscalation = config.PrivilegeEscalation{
					Enabled: true,
				}

				s.mockJobClient.EXPECT().
					CreateOrUpdateConsumer(gomock.Any(), "test-stream", gomock.Any()).
					Return(nil).
					Times(6)

				s.mockJobClient.EXPECT().
					ConsumeJobs(gomock.Any(), "test-stream", gomock.Any(), gomock.Any(), gomock.Any()).
					Return(context.Canceled).
					Times(6)

				return newTestAgent(newTestAgentParams{
					appFs:           s.appFs,
					appConfig:       cfg,
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
					execManager:     mockExecMgr,
				})
			},
			stopFunc: func(a *agent.Agent) {
				stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				a.Stop(stopCtx)
			},
		},
		{
			name: "stop times out when agents are slow to finish",
			setupFunc: func() *agent.Agent {
				blockCh := make(chan struct{})

				s.mockJobClient.EXPECT().
					CreateOrUpdateConsumer(gomock.Any(), "test-stream", gomock.Any()).
					Return(nil).
					Times(6)

				s.mockJobClient.EXPECT().
					ConsumeJobs(gomock.Any(), "test-stream", gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						_ context.Context,
						_ string,
						_ string,
						_ interface{},
						_ interface{},
					) error {
						<-blockCh
						return nil
					}).
					Times(6)

				a := s.buildAgent()

				// Schedule cleanup after Stop returns
				s.T().Cleanup(func() {
					close(blockCh)
					time.Sleep(10 * time.Millisecond)
				})

				return a
			},
			stopFunc: func(a *agent.Agent) {
				stopCtx, cancel := context.WithCancel(context.Background())
				cancel()

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

func (s *AgentPublicTestSuite) TestIsReady() {
	tests := []struct {
		name      string
		setupFunc func() *agent.Agent
		wantErr   bool
		errMsg    string
	}{
		{
			name: "returns error when agent not started",
			setupFunc: func() *agent.Agent {
				return s.buildAgent()
			},
			wantErr: true,
			errMsg:  "agent not started",
		},
		{
			name: "returns nil when agent is started",
			setupFunc: func() *agent.Agent {
				s.mockJobClient.EXPECT().
					CreateOrUpdateConsumer(gomock.Any(), "test-stream", gomock.Any()).
					Return(nil).
					Times(6)

				s.mockJobClient.EXPECT().
					ConsumeJobs(gomock.Any(), "test-stream", gomock.Any(), gomock.Any(), gomock.Any()).
					Return(context.Canceled).
					Times(6)

				a := s.buildAgent()
				a.Start()
				s.T().Cleanup(func() {
					stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					a.Stop(stopCtx)
				})

				return a
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			a := tt.setupFunc()
			err := a.IsReady()

			if tt.wantErr {
				s.Error(err)
				s.Contains(err.Error(), tt.errMsg)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *AgentPublicTestSuite) TestSetMeterProvider() {
	tests := []struct {
		name string
	}{
		{
			name: "creates OTEL instruments without panic",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			a := s.buildAgent()

			port := s.getFreePort()
			srv := metrics.New("127.0.0.1", port, slog.Default())
			s.Require().NotNil(srv)

			s.NotPanics(func() {
				a.SetMeterProvider(srv.MeterProvider())
			})

			ctx, cancel := context.WithTimeout(
				context.Background(),
				5*time.Second,
			)
			defer cancel()

			srv.Stop(ctx)
		})
	}
}

func (s *AgentPublicTestSuite) TestLastHeartbeatTime() {
	tests := []struct {
		name     string
		wantZero bool
	}{
		{
			name:     "returns zero time before any heartbeat",
			wantZero: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			a := s.buildAgent()

			got := a.LastHeartbeatTime()
			if tt.wantZero {
				s.True(got.IsZero())
			}
		})
	}
}

func TestAgentPublicTestSuite(t *testing.T) {
	suite.Run(t, new(AgentPublicTestSuite))
}

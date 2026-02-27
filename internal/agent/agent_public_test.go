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
	"log/slog"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/agent"
	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/job/mocks"
	commandMocks "github.com/retr0h/osapi/internal/provider/command/mocks"
	dnsMocks "github.com/retr0h/osapi/internal/provider/network/dns/mocks"
	pingMocks "github.com/retr0h/osapi/internal/provider/network/ping/mocks"
	diskMocks "github.com/retr0h/osapi/internal/provider/node/disk/mocks"
	hostMocks "github.com/retr0h/osapi/internal/provider/node/host/mocks"
	loadMocks "github.com/retr0h/osapi/internal/provider/node/load/mocks"
	memMocks "github.com/retr0h/osapi/internal/provider/node/mem/mocks"
)

type AgentPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *mocks.MockJobClient
	appFs         afero.Fs
	appConfig     config.Config
	logger        *slog.Logger
}

func (s *AgentPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = mocks.NewMockJobClient(s.mockCtrl)
	s.appFs = afero.NewMemMapFs()
	s.logger = slog.Default()

	s.appConfig = config.Config{
		NATS: config.NATS{
			Stream: config.NATSStream{Name: "test-stream"},
		},
		Node: config.Node{
			Agent: config.NodeAgent{
				Hostname:   "test-agent",
				QueueGroup: "test-queue",
				MaxJobs:    5,
				Consumer: config.NodeAgentConsumer{
					AckWait:       "30s",
					BackOff:       []string{"1s", "2s", "5s"},
					MaxDeliver:    3,
					MaxAckPending: 10,
					ReplayPolicy:  "instant",
				},
			},
		},
	}
}

func (s *AgentPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
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
			a := agent.New(
				s.appFs,
				s.appConfig,
				s.logger,
				s.mockJobClient,
				"test-stream",
				hostMocks.NewDefaultMockProvider(s.mockCtrl),
				diskMocks.NewDefaultMockProvider(s.mockCtrl),
				memMocks.NewDefaultMockProvider(s.mockCtrl),
				loadMocks.NewDefaultMockProvider(s.mockCtrl),
				dnsMocks.NewDefaultMockProvider(s.mockCtrl),
				pingMocks.NewDefaultMockProvider(s.mockCtrl),
				commandMocks.NewDefaultMockProvider(s.mockCtrl),
				nil,
			)

			s.NotNil(a)
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

				return agent.New(
					s.appFs,
					s.appConfig,
					s.logger,
					s.mockJobClient,
					"test-stream",
					hostMocks.NewDefaultMockProvider(s.mockCtrl),
					diskMocks.NewDefaultMockProvider(s.mockCtrl),
					memMocks.NewDefaultMockProvider(s.mockCtrl),
					loadMocks.NewDefaultMockProvider(s.mockCtrl),
					dnsMocks.NewDefaultMockProvider(s.mockCtrl),
					pingMocks.NewDefaultMockProvider(s.mockCtrl),
					commandMocks.NewDefaultMockProvider(s.mockCtrl),
					nil,
				)
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

				a := agent.New(
					s.appFs,
					s.appConfig,
					s.logger,
					s.mockJobClient,
					"test-stream",
					hostMocks.NewDefaultMockProvider(s.mockCtrl),
					diskMocks.NewDefaultMockProvider(s.mockCtrl),
					memMocks.NewDefaultMockProvider(s.mockCtrl),
					loadMocks.NewDefaultMockProvider(s.mockCtrl),
					dnsMocks.NewDefaultMockProvider(s.mockCtrl),
					pingMocks.NewDefaultMockProvider(s.mockCtrl),
					commandMocks.NewDefaultMockProvider(s.mockCtrl),
					nil,
				)

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

func TestAgentPublicTestSuite(t *testing.T) {
	suite.Run(t, new(AgentPublicTestSuite))
}

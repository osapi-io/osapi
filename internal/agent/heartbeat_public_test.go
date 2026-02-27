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
	"log/slog"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/job/mocks"
	"github.com/retr0h/osapi/internal/agent"
	commandMocks "github.com/retr0h/osapi/internal/provider/command/mocks"
	dnsMocks "github.com/retr0h/osapi/internal/provider/network/dns/mocks"
	pingMocks "github.com/retr0h/osapi/internal/provider/network/ping/mocks"
	diskMocks "github.com/retr0h/osapi/internal/provider/node/disk/mocks"
	hostMocks "github.com/retr0h/osapi/internal/provider/node/host/mocks"
	loadMocks "github.com/retr0h/osapi/internal/provider/node/load/mocks"
	memMocks "github.com/retr0h/osapi/internal/provider/node/mem/mocks"
)

type HeartbeatPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *mocks.MockJobClient
	mockKV        *mocks.MockKeyValue
	appFs         afero.Fs
	appConfig     config.Config
	logger        *slog.Logger
}

func (s *HeartbeatPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = mocks.NewMockJobClient(s.mockCtrl)
	s.mockKV = mocks.NewMockKeyValue(s.mockCtrl)
	s.appFs = afero.NewMemMapFs()
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
		Node: config.Node{
			Agent: config.NodeAgent{
				Hostname:   "test-worker",
				QueueGroup: "test-queue",
				MaxJobs:    5,
				Labels:     map[string]string{"group": "web"},
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
				// Heartbeat initial write
				s.mockKV.EXPECT().
					Put(gomock.Any(), "workers.test_worker", gomock.Any()).
					Return(uint64(1), nil).
					MinTimes(1)

				// Deregister on stop
				s.mockKV.EXPECT().
					Delete(gomock.Any(), "workers.test_worker").
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
					s.mockKV,
				)
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

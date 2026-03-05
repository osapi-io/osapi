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

package agent

import (
	"context"
	"log/slog"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/config"
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
)

type DrainTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *mocks.MockJobClient
	mockKV        *mocks.MockKeyValue
	mockEntry     *mocks.MockKeyValueEntry
	agent         *Agent
}

func (s *DrainTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = mocks.NewMockJobClient(s.mockCtrl)
	s.mockKV = mocks.NewMockKeyValue(s.mockCtrl)
	s.mockEntry = mocks.NewMockKeyValueEntry(s.mockCtrl)

	appConfig := config.Config{
		Agent: config.AgentConfig{
			Labels: map[string]string{"group": "web"},
		},
	}

	s.agent = New(
		afero.NewMemMapFs(),
		appConfig,
		slog.Default(),
		s.mockJobClient,
		"test-stream",
		hostMocks.NewDefaultMockProvider(s.mockCtrl),
		diskMocks.NewDefaultMockProvider(s.mockCtrl),
		memMocks.NewDefaultMockProvider(s.mockCtrl),
		loadMocks.NewDefaultMockProvider(s.mockCtrl),
		dnsMocks.NewDefaultMockProvider(s.mockCtrl),
		pingMocks.NewDefaultMockProvider(s.mockCtrl),
		netinfoMocks.NewDefaultMockProvider(s.mockCtrl),
		commandMocks.NewDefaultMockProvider(s.mockCtrl),
		s.mockKV,
		nil,
	)
	s.agent.state = job.AgentStateReady
	s.agent.ctx, s.agent.cancel = context.WithCancel(context.Background())
	s.agent.consumerCtx, s.agent.consumerCancel = context.WithCancel(s.agent.ctx)
}

func (s *DrainTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *DrainTestSuite) TestCheckDrainFlag() {
	tests := []struct {
		name         string
		setupMock    func()
		validateFunc func(bool)
	}{
		{
			name: "when drain key exists returns true",
			setupMock: func() {
				s.mockKV.EXPECT().
					Get(gomock.Any(), "drain.test_agent").
					Return(s.mockEntry, nil)
			},
			validateFunc: func(result bool) {
				s.True(result)
			},
		},
		{
			name: "when drain key missing returns false",
			setupMock: func() {
				s.mockKV.EXPECT().
					Get(gomock.Any(), "drain.test_agent").
					Return(nil, jetstream.ErrKeyNotFound)
			},
			validateFunc: func(result bool) {
				s.False(result)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()
			result := s.agent.checkDrainFlag(context.Background(), "test-agent")
			tt.validateFunc(result)
		})
	}
}

func (s *DrainTestSuite) TestHandleDrainDetection() {
	tests := []struct {
		name          string
		initialState  string
		setupMock     func()
		expectedState string
	}{
		{
			name:         "when drain flag set and agent is Ready transitions to Cordoned",
			initialState: job.AgentStateReady,
			setupMock: func() {
				s.mockKV.EXPECT().
					Get(gomock.Any(), "drain.test_agent").
					Return(s.mockEntry, nil)
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
			expectedState: job.AgentStateCordoned,
		},
		{
			name:         "when drain flag removed and agent is Draining transitions to Ready",
			initialState: job.AgentStateDraining,
			setupMock: func() {
				s.mockKV.EXPECT().
					Get(gomock.Any(), "drain.test_agent").
					Return(nil, jetstream.ErrKeyNotFound)
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
			expectedState: job.AgentStateReady,
		},
		{
			name:         "when drain flag removed and agent is Cordoned transitions to Ready",
			initialState: job.AgentStateCordoned,
			setupMock: func() {
				s.mockKV.EXPECT().
					Get(gomock.Any(), "drain.test_agent").
					Return(nil, jetstream.ErrKeyNotFound)
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
			expectedState: job.AgentStateReady,
		},
		{
			name:         "when drain flag still set and agent is already Draining stays Draining",
			initialState: job.AgentStateDraining,
			setupMock: func() {
				s.mockKV.EXPECT().
					Get(gomock.Any(), "drain.test_agent").
					Return(s.mockEntry, nil)
			},
			expectedState: job.AgentStateDraining,
		},
		{
			name:         "when no drain flag and agent is Ready stays Ready",
			initialState: job.AgentStateReady,
			setupMock: func() {
				s.mockKV.EXPECT().
					Get(gomock.Any(), "drain.test_agent").
					Return(nil, jetstream.ErrKeyNotFound)
			},
			expectedState: job.AgentStateReady,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.agent.state = tt.initialState
			tt.setupMock()
			s.agent.handleDrainDetection(context.Background(), "test-agent")
			s.Equal(tt.expectedState, s.agent.state)
		})
	}
}

func TestDrainTestSuite(t *testing.T) {
	suite.Run(t, new(DrainTestSuite))
}

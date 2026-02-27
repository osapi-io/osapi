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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	apiagent "github.com/retr0h/osapi/internal/api/agent"
	"github.com/retr0h/osapi/internal/api/agent/gen"
	jobtypes "github.com/retr0h/osapi/internal/job"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
	"github.com/retr0h/osapi/internal/provider/node/host"
	"github.com/retr0h/osapi/internal/provider/node/load"
	"github.com/retr0h/osapi/internal/provider/node/mem"
)

type AgentListPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apiagent.Agent
	ctx           context.Context
}

func (s *AgentListPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apiagent.New(slog.Default(), s.mockJobClient)
	s.ctx = context.Background()
}

func (s *AgentListPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *AgentListPublicTestSuite) TestGetAgent() {
	tests := []struct {
		name         string
		mockAgents   []jobtypes.AgentInfo
		mockError    error
		validateFunc func(resp gen.GetAgentResponseObject)
	}{
		{
			name: "success with agents",
			mockAgents: []jobtypes.AgentInfo{
				{
					Hostname:     "server1",
					Labels:       map[string]string{"group": "web"},
					RegisteredAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
					StartedAt:    time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
					OSInfo:       &host.OSInfo{Distribution: "Ubuntu", Version: "24.04"},
					Uptime:       5 * time.Hour,
					LoadAverages: &load.AverageStats{Load1: 0.5, Load5: 0.3, Load15: 0.2},
					MemoryStats:  &mem.Stats{Total: 8388608, Free: 4194304, Cached: 2097152},
				},
				{Hostname: "server2"},
			},
			validateFunc: func(resp gen.GetAgentResponseObject) {
				r, ok := resp.(gen.GetAgent200JSONResponse)
				s.True(ok)
				s.Equal(2, r.Total)
				s.Len(r.Agents, 2)
				s.Equal("server1", r.Agents[0].Hostname)
				s.Equal(gen.Ready, r.Agents[0].Status)
				s.NotNil(r.Agents[0].Labels)
				s.NotNil(r.Agents[0].RegisteredAt)
				s.NotNil(r.Agents[0].StartedAt)
				s.NotNil(r.Agents[0].OsInfo)
				s.Equal("Ubuntu", r.Agents[0].OsInfo.Distribution)
				s.NotNil(r.Agents[0].LoadAverage)
				s.NotNil(r.Agents[0].Memory)
				s.NotNil(r.Agents[0].Uptime)
				s.Equal("server2", r.Agents[1].Hostname)
				s.Equal(gen.Ready, r.Agents[1].Status)
			},
		},
		{
			name:       "success with no agents",
			mockAgents: []jobtypes.AgentInfo{},
			validateFunc: func(resp gen.GetAgentResponseObject) {
				r, ok := resp.(gen.GetAgent200JSONResponse)
				s.True(ok)
				s.Equal(0, r.Total)
				s.Empty(r.Agents)
			},
		},
		{
			name:      "job client error",
			mockError: assert.AnError,
			validateFunc: func(resp gen.GetAgentResponseObject) {
				_, ok := resp.(gen.GetAgent500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.mockJobClient.EXPECT().
				ListAgents(gomock.Any()).
				Return(tt.mockAgents, tt.mockError)

			resp, err := s.handler.GetAgent(s.ctx, gen.GetAgentRequestObject{})
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func TestAgentListPublicTestSuite(t *testing.T) {
	suite.Run(t, new(AgentListPublicTestSuite))
}

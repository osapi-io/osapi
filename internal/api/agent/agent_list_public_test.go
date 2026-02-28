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
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/api"
	apiagent "github.com/retr0h/osapi/internal/api/agent"
	"github.com/retr0h/osapi/internal/api/agent/gen"
	"github.com/retr0h/osapi/internal/authtoken"
	"github.com/retr0h/osapi/internal/config"
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
	appConfig     config.Config
	logger        *slog.Logger
}

func (s *AgentListPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apiagent.New(slog.Default(), s.mockJobClient)
	s.ctx = context.Background()
	s.appConfig = config.Config{}
	s.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
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

func (s *AgentListPublicTestSuite) TestGetAgentValidationHTTP() {
	tests := []struct {
		name         string
		setupJobMock func() *jobmocks.MockJobClient
		wantCode     int
		wantContains []string
	}{
		{
			name: "when agents exist returns agent list",
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(s.mockCtrl)
				mock.EXPECT().
					ListAgents(gomock.Any()).
					Return([]jobtypes.AgentInfo{
						{Hostname: "server1"},
						{Hostname: "server2"},
					}, nil)
				return mock
			},
			wantCode:     http.StatusOK,
			wantContains: []string{`"total":2`, `"server1"`, `"server2"`, `"status":"Ready"`},
		},
		{
			name: "when no agents returns empty list",
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(s.mockCtrl)
				mock.EXPECT().
					ListAgents(gomock.Any()).
					Return([]jobtypes.AgentInfo{}, nil)
				return mock
			},
			wantCode:     http.StatusOK,
			wantContains: []string{`"total":0`},
		},
		{
			name: "when job client errors returns 500",
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(s.mockCtrl)
				mock.EXPECT().
					ListAgents(gomock.Any()).
					Return(nil, assert.AnError)
				return mock
			},
			wantCode:     http.StatusInternalServerError,
			wantContains: []string{`"error"`},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			agentHandler := apiagent.New(s.logger, jobMock)
			strictHandler := gen.NewStrictHandler(agentHandler, nil)

			a := api.New(s.appConfig, s.logger)
			gen.RegisterHandlers(a.Echo, strictHandler)

			req := httptest.NewRequest(http.MethodGet, "/agent", nil)
			rec := httptest.NewRecorder()

			a.Echo.ServeHTTP(rec, req)

			s.Equal(tc.wantCode, rec.Code)
			for _, str := range tc.wantContains {
				s.Contains(rec.Body.String(), str)
			}
		})
	}
}

const rbacAgentListTestSigningKey = "test-signing-key-for-rbac-agent-list"

func (s *AgentListPublicTestSuite) TestGetAgentRBACHTTP() {
	tokenManager := authtoken.New(s.logger)

	tests := []struct {
		name         string
		setupAuth    func(req *http.Request)
		setupJobMock func() *jobmocks.MockJobClient
		wantCode     int
		wantContains []string
	}{
		{
			name: "when no token returns 401",
			setupAuth: func(_ *http.Request) {
				// No auth header set
			},
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(s.mockCtrl)
			},
			wantCode:     http.StatusUnauthorized,
			wantContains: []string{"Bearer token required"},
		},
		{
			name: "when insufficient permissions returns 403",
			setupAuth: func(req *http.Request) {
				token, err := tokenManager.Generate(
					rbacAgentListTestSigningKey,
					[]string{"read"},
					"test-user",
					[]string{"network:read"},
				)
				s.Require().NoError(err)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			},
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(s.mockCtrl)
			},
			wantCode:     http.StatusForbidden,
			wantContains: []string{"Insufficient permissions"},
		},
		{
			name: "when valid token with agent:read returns 200",
			setupAuth: func(req *http.Request) {
				token, err := tokenManager.Generate(
					rbacAgentListTestSigningKey,
					[]string{"admin"},
					"test-user",
					nil,
				)
				s.Require().NoError(err)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			},
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(s.mockCtrl)
				mock.EXPECT().
					ListAgents(gomock.Any()).
					Return([]jobtypes.AgentInfo{
						{Hostname: "server1"},
						{Hostname: "server2"},
					}, nil)
				return mock
			},
			wantCode:     http.StatusOK,
			wantContains: []string{`"total":2`, `"server1"`},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			appConfig := config.Config{
				API: config.API{
					Server: config.Server{
						Security: config.ServerSecurity{
							SigningKey: rbacAgentListTestSigningKey,
						},
					},
				},
			}

			server := api.New(appConfig, s.logger)
			handlers := server.GetAgentHandler(jobMock)
			server.RegisterHandlers(handlers)

			req := httptest.NewRequest(
				http.MethodGet,
				"/agent",
				nil,
			)
			tc.setupAuth(req)
			rec := httptest.NewRecorder()

			server.Echo.ServeHTTP(rec, req)

			s.Equal(tc.wantCode, rec.Code)
			for _, str := range tc.wantContains {
				s.Contains(rec.Body.String(), str)
			}
		})
	}
}

func TestAgentListPublicTestSuite(t *testing.T) {
	suite.Run(t, new(AgentListPublicTestSuite))
}

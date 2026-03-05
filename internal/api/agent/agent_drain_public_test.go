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

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/api"
	apiagent "github.com/retr0h/osapi/internal/api/agent"
	"github.com/retr0h/osapi/internal/api/agent/gen"
	"github.com/retr0h/osapi/internal/authtoken"
	"github.com/retr0h/osapi/internal/config"
	jobtypes "github.com/retr0h/osapi/internal/job"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
)

type AgentDrainPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apiagent.Agent
	ctx           context.Context
	appConfig     config.Config
	logger        *slog.Logger
}

func (s *AgentDrainPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apiagent.New(slog.Default(), s.mockJobClient)
	s.ctx = context.Background()
	s.appConfig = config.Config{}
	s.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (s *AgentDrainPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *AgentDrainPublicTestSuite) TestDrainAgent() {
	tests := []struct {
		name         string
		hostname     string
		mockAgent    *jobtypes.AgentInfo
		mockGetErr   error
		mockWriteErr error
		skipWrite    bool
		validateFunc func(resp gen.DrainAgentResponseObject)
	}{
		{
			name:     "success drains agent",
			hostname: "server1",
			mockAgent: &jobtypes.AgentInfo{
				Hostname: "server1",
				State:    jobtypes.AgentStateReady,
			},
			validateFunc: func(resp gen.DrainAgentResponseObject) {
				r, ok := resp.(gen.DrainAgent200JSONResponse)
				s.True(ok)
				s.Contains(r.Message, "drain initiated for agent server1")
			},
		},
		{
			name:       "agent not found returns 404",
			hostname:   "unknown",
			mockGetErr: fmt.Errorf("agent not found: unknown"),
			skipWrite:  true,
			validateFunc: func(resp gen.DrainAgentResponseObject) {
				_, ok := resp.(gen.DrainAgent404JSONResponse)
				s.True(ok)
			},
		},
		{
			name:     "agent already draining returns 409",
			hostname: "server1",
			mockAgent: &jobtypes.AgentInfo{
				Hostname: "server1",
				State:    jobtypes.AgentStateDraining,
			},
			skipWrite: true,
			validateFunc: func(resp gen.DrainAgentResponseObject) {
				_, ok := resp.(gen.DrainAgent409JSONResponse)
				s.True(ok)
			},
		},
		{
			name:     "agent already cordoned returns 409",
			hostname: "server1",
			mockAgent: &jobtypes.AgentInfo{
				Hostname: "server1",
				State:    jobtypes.AgentStateCordoned,
			},
			skipWrite: true,
			validateFunc: func(resp gen.DrainAgentResponseObject) {
				_, ok := resp.(gen.DrainAgent409JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.mockJobClient.EXPECT().
				GetAgent(gomock.Any(), tt.hostname).
				Return(tt.mockAgent, tt.mockGetErr)

			if !tt.skipWrite {
				s.mockJobClient.EXPECT().
					WriteAgentTimelineEvent(gomock.Any(), tt.hostname, "drain", "Drain initiated via API").
					Return(tt.mockWriteErr)
			}

			resp, err := s.handler.DrainAgent(s.ctx, gen.DrainAgentRequestObject{
				Hostname: tt.hostname,
			})
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func (s *AgentDrainPublicTestSuite) TestDrainAgentValidationHTTP() {
	tests := []struct {
		name         string
		hostname     string
		setupJobMock func() *jobmocks.MockJobClient
		wantCode     int
		wantContains []string
	}{
		{
			name:     "when agent exists returns 200",
			hostname: "server1",
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(s.mockCtrl)
				mock.EXPECT().
					GetAgent(gomock.Any(), "server1").
					Return(&jobtypes.AgentInfo{
						Hostname: "server1",
						State:    jobtypes.AgentStateReady,
					}, nil)
				mock.EXPECT().
					WriteAgentTimelineEvent(gomock.Any(), "server1", "drain", "Drain initiated via API").
					Return(nil)
				return mock
			},
			wantCode:     http.StatusOK,
			wantContains: []string{`"message"`, `drain initiated`},
		},
		{
			name:     "when agent not found returns 404",
			hostname: "unknown",
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(s.mockCtrl)
				mock.EXPECT().
					GetAgent(gomock.Any(), "unknown").
					Return(nil, fmt.Errorf("agent not found: unknown"))
				return mock
			},
			wantCode:     http.StatusNotFound,
			wantContains: []string{`"error"`},
		},
		{
			name:     "when agent already draining returns 409",
			hostname: "server1",
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(s.mockCtrl)
				mock.EXPECT().
					GetAgent(gomock.Any(), "server1").
					Return(&jobtypes.AgentInfo{
						Hostname: "server1",
						State:    jobtypes.AgentStateDraining,
					}, nil)
				return mock
			},
			wantCode:     http.StatusConflict,
			wantContains: []string{`"error"`, `already in Draining`},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			agentHandler := apiagent.New(s.logger, jobMock)
			strictHandler := gen.NewStrictHandler(agentHandler, nil)

			a := api.New(s.appConfig, s.logger)
			gen.RegisterHandlers(a.Echo, strictHandler)

			req := httptest.NewRequest(
				http.MethodPost,
				fmt.Sprintf("/agent/%s/drain", tc.hostname),
				nil,
			)
			rec := httptest.NewRecorder()

			a.Echo.ServeHTTP(rec, req)

			s.Equal(tc.wantCode, rec.Code)
			for _, str := range tc.wantContains {
				s.Contains(rec.Body.String(), str)
			}
		})
	}
}

const rbacAgentDrainTestSigningKey = "test-signing-key-for-rbac-agent-drain"

func (s *AgentDrainPublicTestSuite) TestDrainAgentRBACHTTP() {
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
					rbacAgentDrainTestSigningKey,
					[]string{"read"},
					"test-user",
					[]string{"agent:read"},
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
			name: "when valid token with agent:write returns 200",
			setupAuth: func(req *http.Request) {
				token, err := tokenManager.Generate(
					rbacAgentDrainTestSigningKey,
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
					GetAgent(gomock.Any(), "server1").
					Return(&jobtypes.AgentInfo{
						Hostname: "server1",
						State:    jobtypes.AgentStateReady,
					}, nil)
				mock.EXPECT().
					WriteAgentTimelineEvent(gomock.Any(), "server1", "drain", "Drain initiated via API").
					Return(nil)
				return mock
			},
			wantCode:     http.StatusOK,
			wantContains: []string{`"message"`, `drain initiated`},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			appConfig := config.Config{
				API: config.API{
					Server: config.Server{
						Security: config.ServerSecurity{
							SigningKey: rbacAgentDrainTestSigningKey,
						},
					},
				},
			}

			server := api.New(appConfig, s.logger)
			handlers := server.GetAgentHandler(jobMock)
			server.RegisterHandlers(handlers)

			req := httptest.NewRequest(
				http.MethodPost,
				"/agent/server1/drain",
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

func TestAgentDrainPublicTestSuite(t *testing.T) {
	suite.Run(t, new(AgentDrainPublicTestSuite))
}

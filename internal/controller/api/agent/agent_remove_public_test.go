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
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller/api"
	apiagent "github.com/osapi-io/osapi/internal/controller/api/agent"
	"github.com/osapi-io/osapi/internal/controller/api/agent/gen"
	agentmocks "github.com/osapi-io/osapi/internal/controller/api/agent/mocks"
	"github.com/osapi-io/osapi/internal/controller/enrollment"
	jobmocks "github.com/osapi-io/osapi/internal/job/mocks"
)

type AgentRemovePublicTestSuite struct {
	suite.Suite

	mockCtrl       *gomock.Controller
	mockJobClient  *jobmocks.MockJobClient
	mockEnrollment *agentmocks.MockEnrollmentManager
	handler        *apiagent.Agent
	ctx            context.Context
	logger         *slog.Logger
}

func (s *AgentRemovePublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.mockEnrollment = agentmocks.NewMockEnrollmentManager(s.mockCtrl)
	s.handler = apiagent.New(slog.Default(), s.mockJobClient, s.mockEnrollment)
	s.ctx = context.Background()
	s.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (s *AgentRemovePublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *AgentRemovePublicTestSuite) TestRemoveAgent() {
	ptr := func(v string) *string { return &v }

	tests := []struct {
		name         string
		hostname     string
		machineID    *string
		useNilEnroll bool
		setupMock    func()
		validateFunc func(resp gen.RemoveAgentResponseObject)
	}{
		{
			name:     "returns 400 when hostname is empty",
			hostname: "",
			validateFunc: func(resp gen.RemoveAgentResponseObject) {
				s.IsType(gen.RemoveAgent400JSONResponse{}, resp)
			},
		},
		{
			name:      "returns 400 when machine ID is too long",
			hostname:  "web-01",
			machineID: ptr(strings.Repeat("a", 256)),
			validateFunc: func(resp gen.RemoveAgentResponseObject) {
				s.IsType(gen.RemoveAgent400JSONResponse{}, resp)
			},
		},
		{
			name:         "returns 500 when enrollment not enabled",
			hostname:     "web-01",
			useNilEnroll: true,
			validateFunc: func(resp gen.RemoveAgentResponseObject) {
				r, ok := resp.(gen.RemoveAgent500JSONResponse)
				s.Require().True(ok)
				s.Equal("enrollment not enabled", *r.Error)
			},
		},
		{
			name:     "removes the agent and names its machine ID",
			hostname: "web-01",
			setupMock: func() {
				s.mockEnrollment.EXPECT().
					RemoveByHostname(gomock.Any(), "web-01", "").
					Return("machine-001", nil)
			},
			validateFunc: func(resp gen.RemoveAgentResponseObject) {
				r, ok := resp.(gen.RemoveAgent200JSONResponse)
				s.Require().True(ok)
				s.Equal("agent web-01 (machine-001) removed", r.Message)
			},
		},
		{
			name:      "passes the machine ID through",
			hostname:  "web-01",
			machineID: ptr("machine-002"),
			setupMock: func() {
				s.mockEnrollment.EXPECT().
					RemoveByHostname(gomock.Any(), "web-01", "machine-002").
					Return("machine-002", nil)
			},
			validateFunc: func(resp gen.RemoveAgentResponseObject) {
				s.IsType(gen.RemoveAgent200JSONResponse{}, resp)
			},
		},
		{
			name:     "returns 404 when no accepted agent has the hostname",
			hostname: "web-01",
			setupMock: func() {
				s.mockEnrollment.EXPECT().
					RemoveByHostname(gomock.Any(), "web-01", "").
					Return("", fmt.Errorf("x: %w", enrollment.ErrNoAcceptedAgent))
			},
			validateFunc: func(resp gen.RemoveAgentResponseObject) {
				r, ok := resp.(gen.RemoveAgent404JSONResponse)
				s.Require().True(ok)
				s.Equal(`no accepted agent with hostname "web-01"`, *r.Error)
			},
		},
		{
			name:      "returns 404 naming the machine ID when one was given",
			hostname:  "web-01",
			machineID: ptr("machine-009"),
			setupMock: func() {
				s.mockEnrollment.EXPECT().
					RemoveByHostname(gomock.Any(), "web-01", "machine-009").
					Return("", fmt.Errorf("x: %w", enrollment.ErrNoAcceptedAgent))
			},
			validateFunc: func(resp gen.RemoveAgentResponseObject) {
				r, ok := resp.(gen.RemoveAgent404JSONResponse)
				s.Require().True(ok)
				s.Contains(*r.Error, `and machine ID "machine-009"`)
			},
		},
		{
			name:     "returns 409 when two accepted agents share the hostname",
			hostname: "web-01",
			setupMock: func() {
				s.mockEnrollment.EXPECT().
					RemoveByHostname(gomock.Any(), "web-01", "").
					Return("", fmt.Errorf("machine IDs a, b: %w", enrollment.ErrAmbiguousHostname))
			},
			validateFunc: func(resp gen.RemoveAgentResponseObject) {
				r, ok := resp.(gen.RemoveAgent409JSONResponse)
				s.Require().True(ok)
				s.Contains(*r.Error, "machine IDs a, b")
				s.Contains(*r.Error, "name one with machine_id")
			},
		},
		{
			name:     "returns 500 on another error",
			hostname: "web-01",
			setupMock: func() {
				s.mockEnrollment.EXPECT().
					RemoveByHostname(gomock.Any(), "web-01", "").
					Return("", fmt.Errorf("kv down"))
			},
			validateFunc: func(resp gen.RemoveAgentResponseObject) {
				r, ok := resp.(gen.RemoveAgent500JSONResponse)
				s.Require().True(ok)
				s.Equal("kv down", *r.Error)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			handler := s.handler
			if tt.useNilEnroll {
				handler = apiagent.New(slog.Default(), s.mockJobClient, nil)
			}
			if tt.setupMock != nil {
				tt.setupMock()
			}

			resp, err := handler.RemoveAgent(s.ctx, gen.RemoveAgentRequestObject{
				Hostname: tt.hostname,
				Params:   gen.RemoveAgentParams{MachineId: tt.machineID},
			})
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

const rbacAgentRemoveTestSigningKey = "test-signing-key-for-rbac-agent-remove"

func (s *AgentRemovePublicTestSuite) TestRemoveAgentRBACHTTP() {
	tokenManager := authtoken.New(s.logger)

	tests := []struct {
		name         string
		path         string
		roles        []string
		permissions  []string
		noToken      bool
		setupMocks   func(em *agentmocks.MockEnrollmentManager)
		validateFunc func(*httptest.ResponseRecorder)
	}{
		{
			name:    "when no token returns 401",
			path:    "/api/agent/web-01",
			noToken: true,
			validateFunc: func(rec *httptest.ResponseRecorder) {
				s.Equal(http.StatusUnauthorized, rec.Code)
			},
		},
		{
			name:        "when the token has only agent:read returns 403",
			path:        "/api/agent/web-01",
			roles:       []string{"read"},
			permissions: []string{"agent:read"},
			validateFunc: func(rec *httptest.ResponseRecorder) {
				s.Equal(http.StatusForbidden, rec.Code)
			},
		},
		{
			name:  "when an admin removes it returns 200",
			path:  "/api/agent/web-01?machine_id=machine-001",
			roles: []string{"admin"},
			setupMocks: func(em *agentmocks.MockEnrollmentManager) {
				em.EXPECT().
					RemoveByHostname(gomock.Any(), "web-01", "machine-001").
					Return("machine-001", nil)
			},
			validateFunc: func(rec *httptest.ResponseRecorder) {
				s.Equal(http.StatusOK, rec.Code)
				s.Contains(rec.Body.String(), "machine-001")
			},
		},
		{
			name:  "when two share the hostname returns 409",
			path:  "/api/agent/web-01",
			roles: []string{"admin"},
			setupMocks: func(em *agentmocks.MockEnrollmentManager) {
				em.EXPECT().
					RemoveByHostname(gomock.Any(), "web-01", "").
					Return("", fmt.Errorf("x: %w", enrollment.ErrAmbiguousHostname))
			},
			validateFunc: func(rec *httptest.ResponseRecorder) {
				s.Equal(http.StatusConflict, rec.Code)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			jm := jobmocks.NewMockJobClient(s.mockCtrl)
			em := agentmocks.NewMockEnrollmentManager(s.mockCtrl)
			if tc.setupMocks != nil {
				tc.setupMocks(em)
			}

			appConfig := config.Config{
				Controller: config.Controller{
					API: config.APIServer{
						Security: config.ServerSecurity{
							SigningKey: rbacAgentRemoveTestSigningKey,
						},
					},
				},
			}

			server := api.New(appConfig, s.logger)
			server.RegisterHandlers(apiagent.Handler(
				s.logger,
				jm,
				appConfig.Controller.API.Security.SigningKey,
				nil,
				em,
			))

			req := httptest.NewRequest(http.MethodDelete, tc.path, nil)
			if !tc.noToken {
				token, err := tokenManager.Generate(
					rbacAgentRemoveTestSigningKey, tc.roles, "test-user", tc.permissions,
				)
				s.Require().NoError(err)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			}
			rec := httptest.NewRecorder()

			server.Echo.ServeHTTP(rec, req)

			tc.validateFunc(rec)
		})
	}
}

func TestAgentRemovePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(AgentRemovePublicTestSuite))
}

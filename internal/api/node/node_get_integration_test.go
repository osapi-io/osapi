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

package node_test

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/api"
	"github.com/retr0h/osapi/internal/api/node"
	nodeGen "github.com/retr0h/osapi/internal/api/node/gen"
	"github.com/retr0h/osapi/internal/authtoken"
	"github.com/retr0h/osapi/internal/config"
	jobtypes "github.com/retr0h/osapi/internal/job"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
	"github.com/retr0h/osapi/internal/provider/node/host"
	"github.com/retr0h/osapi/internal/provider/node/load"
	"github.com/retr0h/osapi/internal/provider/node/mem"
)

type NodeGetIntegrationTestSuite struct {
	suite.Suite
	ctrl *gomock.Controller

	appConfig config.Config
	logger    *slog.Logger
}

func (suite *NodeGetIntegrationTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())

	suite.appConfig = config.Config{}
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (suite *NodeGetIntegrationTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *NodeGetIntegrationTestSuite) TestGetNodeDetailsValidation() {
	tests := []struct {
		name         string
		hostname     string
		setupJobMock func() *jobmocks.MockJobClient
		wantCode     int
		wantContains []string
	}{
		{
			name:     "when agent exists returns details",
			hostname: "server1",
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(suite.ctrl)
				mock.EXPECT().
					GetAgent(gomock.Any(), "server1").
					Return(&jobtypes.AgentInfo{
						Hostname:     "server1",
						Labels:       map[string]string{"group": "web"},
						RegisteredAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
						StartedAt:    time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
						OSInfo:       &host.OSInfo{Distribution: "Ubuntu", Version: "24.04"},
						Uptime:       5 * time.Hour,
						LoadAverages: &load.AverageStats{Load1: 0.5, Load5: 0.3, Load15: 0.2},
						MemoryStats:  &mem.Stats{Total: 8388608, Free: 4194304, Cached: 2097152},
					}, nil)
				return mock
			},
			wantCode:     http.StatusOK,
			wantContains: []string{`"server1"`, `"Ready"`, `"Ubuntu"`},
		},
		{
			name:     "when agent not found returns 404",
			hostname: "unknown",
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(suite.ctrl)
				mock.EXPECT().
					GetAgent(gomock.Any(), "unknown").
					Return(nil, fmt.Errorf("agent not found: unknown"))
				return mock
			},
			wantCode:     http.StatusNotFound,
			wantContains: []string{`"error"`},
		},
		{
			name:     "when client error returns 500",
			hostname: "server1",
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(suite.ctrl)
				mock.EXPECT().
					GetAgent(gomock.Any(), "server1").
					Return(nil, fmt.Errorf("connection failed"))
				return mock
			},
			wantCode:     http.StatusInternalServerError,
			wantContains: []string{`"error"`},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			nodeHandler := node.New(suite.logger, jobMock)
			strictHandler := nodeGen.NewStrictHandler(nodeHandler, nil)

			a := api.New(suite.appConfig, suite.logger)
			nodeGen.RegisterHandlers(a.Echo, strictHandler)

			req := httptest.NewRequest(
				http.MethodGet,
				fmt.Sprintf("/node/%s", tc.hostname),
				nil,
			)
			rec := httptest.NewRecorder()

			a.Echo.ServeHTTP(rec, req)

			suite.Equal(tc.wantCode, rec.Code)
			for _, s := range tc.wantContains {
				suite.Contains(rec.Body.String(), s)
			}
		})
	}
}

const rbacNodeGetTestSigningKey = "test-signing-key-for-rbac-node-get"

func (suite *NodeGetIntegrationTestSuite) TestGetNodeDetailsRBAC() {
	tokenManager := authtoken.New(suite.logger)

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
				return jobmocks.NewMockJobClient(suite.ctrl)
			},
			wantCode:     http.StatusUnauthorized,
			wantContains: []string{"Bearer token required"},
		},
		{
			name: "when insufficient permissions returns 403",
			setupAuth: func(req *http.Request) {
				token, err := tokenManager.Generate(
					rbacNodeGetTestSigningKey,
					[]string{"read"},
					"test-user",
					[]string{"network:read"},
				)
				suite.Require().NoError(err)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			},
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(suite.ctrl)
			},
			wantCode:     http.StatusForbidden,
			wantContains: []string{"Insufficient permissions"},
		},
		{
			name: "when valid token with node:read returns 200",
			setupAuth: func(req *http.Request) {
				token, err := tokenManager.Generate(
					rbacNodeGetTestSigningKey,
					[]string{"admin"},
					"test-user",
					nil,
				)
				suite.Require().NoError(err)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			},
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(suite.ctrl)
				mock.EXPECT().
					GetAgent(gomock.Any(), "server1").
					Return(&jobtypes.AgentInfo{
						Hostname: "server1",
					}, nil)
				return mock
			},
			wantCode:     http.StatusOK,
			wantContains: []string{`"server1"`, `"Ready"`},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			appConfig := config.Config{
				API: config.API{
					Server: config.Server{
						Security: config.ServerSecurity{
							SigningKey: rbacNodeGetTestSigningKey,
						},
					},
				},
			}

			server := api.New(appConfig, suite.logger)
			handlers := server.GetNodeHandler(jobMock)
			server.RegisterHandlers(handlers)

			req := httptest.NewRequest(
				http.MethodGet,
				"/node/server1",
				nil,
			)
			tc.setupAuth(req)
			rec := httptest.NewRecorder()

			server.Echo.ServeHTTP(rec, req)

			suite.Equal(tc.wantCode, rec.Code)
			for _, s := range tc.wantContains {
				suite.Contains(rec.Body.String(), s)
			}
		})
	}
}

func TestNodeGetIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(NodeGetIntegrationTestSuite))
}

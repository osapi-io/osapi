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

package command_test

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/api"
	apicommand "github.com/retr0h/osapi/internal/api/command"
	commandGen "github.com/retr0h/osapi/internal/api/command/gen"
	"github.com/retr0h/osapi/internal/authtoken"
	"github.com/retr0h/osapi/internal/config"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
	"github.com/retr0h/osapi/internal/provider/command"
)

type CommandExecPostIntegrationTestSuite struct {
	suite.Suite
	ctrl *gomock.Controller

	appConfig config.Config
	logger    *slog.Logger
}

func (suite *CommandExecPostIntegrationTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())

	suite.appConfig = config.Config{}
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (suite *CommandExecPostIntegrationTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *CommandExecPostIntegrationTestSuite) TestPostCommandExecValidation() {
	tests := []struct {
		name         string
		path         string
		body         string
		setupJobMock func() *jobmocks.MockJobClient
		wantCode     int
		wantContains []string
	}{
		{
			name: "when valid request",
			path: "/command/exec",
			body: `{"command":"ls","args":["-la"]}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(suite.ctrl)
				mock.EXPECT().
					ModifyCommandExec(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("550e8400-e29b-41d4-a716-446655440000", &command.Result{
						Stdout:     "file1\nfile2",
						Stderr:     "",
						ExitCode:   0,
						DurationMs: 42,
						Changed:    true,
					}, "worker1", nil)
				return mock
			},
			wantCode:     http.StatusAccepted,
			wantContains: []string{`"results"`, `"worker1"`, `"changed":true`},
		},
		{
			name: "when missing command",
			path: "/command/exec",
			body: `{}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(suite.ctrl)
			},
			wantCode:     http.StatusBadRequest,
			wantContains: []string{`"error"`, "Command", "required"},
		},
		{
			name: "when invalid timeout",
			path: "/command/exec",
			body: `{"command":"ls","timeout":999}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(suite.ctrl)
			},
			wantCode:     http.StatusBadRequest,
			wantContains: []string{`"error"`, "Timeout", "max"},
		},
		{
			name: "when empty target_hostname",
			path: "/command/exec?target_hostname=",
			body: `{"command":"ls"}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(suite.ctrl)
			},
			wantCode:     http.StatusBadRequest,
			wantContains: []string{`"error"`},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			commandHandler := apicommand.New(suite.logger, jobMock)
			strictHandler := commandGen.NewStrictHandler(commandHandler, nil)

			a := api.New(suite.appConfig, suite.logger)
			commandGen.RegisterHandlers(a.Echo, strictHandler)

			req := httptest.NewRequest(
				http.MethodPost,
				tc.path,
				strings.NewReader(tc.body),
			)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			a.Echo.ServeHTTP(rec, req)

			suite.Equal(tc.wantCode, rec.Code)
			for _, s := range tc.wantContains {
				suite.Contains(rec.Body.String(), s)
			}
		})
	}
}

const rbacExecTestSigningKey = "test-signing-key-for-exec-rbac"

func (suite *CommandExecPostIntegrationTestSuite) TestPostCommandExecRBAC() {
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
					rbacExecTestSigningKey,
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
			name: "when valid token with command:execute returns 202",
			setupAuth: func(req *http.Request) {
				token, err := tokenManager.Generate(
					rbacExecTestSigningKey,
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
					ModifyCommandExec(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(
						"550e8400-e29b-41d4-a716-446655440000",
						&command.Result{
							Stdout:     "output",
							Stderr:     "",
							ExitCode:   0,
							DurationMs: 10,
							Changed:    true,
						},
						"worker1",
						nil,
					)
				return mock
			},
			wantCode:     http.StatusAccepted,
			wantContains: []string{`"results"`, `"changed":true`},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			appConfig := config.Config{
				API: config.API{
					Server: config.Server{
						Security: config.ServerSecurity{
							SigningKey: rbacExecTestSigningKey,
						},
					},
				},
			}

			server := api.New(appConfig, suite.logger)
			handlers := server.GetCommandHandler(jobMock)
			server.RegisterHandlers(handlers)

			req := httptest.NewRequest(
				http.MethodPost,
				"/command/exec",
				strings.NewReader(`{"command":"ls","args":["-la"]}`),
			)
			req.Header.Set("Content-Type", "application/json")
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

func TestCommandExecPostIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CommandExecPostIntegrationTestSuite))
}

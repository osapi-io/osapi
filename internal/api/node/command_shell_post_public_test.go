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
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/api"
	apinode "github.com/retr0h/osapi/internal/api/node"
	"github.com/retr0h/osapi/internal/api/node/gen"
	"github.com/retr0h/osapi/internal/authtoken"
	"github.com/retr0h/osapi/internal/config"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
	"github.com/retr0h/osapi/internal/provider/command"
	"github.com/retr0h/osapi/internal/validation"
)

type CommandShellPostPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apinode.Node
	ctx           context.Context
	appConfig     config.Config
	logger        *slog.Logger
}

func (s *CommandShellPostPublicTestSuite) SetupSuite() {
	validation.RegisterTargetValidator(func(_ context.Context) ([]validation.AgentTarget, error) {
		return []validation.AgentTarget{
			{Hostname: "server1", Labels: map[string]string{"group": "web"}},
			{Hostname: "server2"},
		}, nil
	})
}

func (s *CommandShellPostPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apinode.New(slog.Default(), s.mockJobClient)
	s.ctx = context.Background()
	s.appConfig = config.Config{}
	s.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (s *CommandShellPostPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *CommandShellPostPublicTestSuite) TestPostNodeCommandShell() {
	tests := []struct {
		name         string
		request      gen.PostNodeCommandShellRequestObject
		setupMock    func()
		validateFunc func(resp gen.PostNodeCommandShellResponseObject)
	}{
		{
			name: "success",
			request: gen.PostNodeCommandShellRequestObject{
				Hostname: "_any",
				Body: &gen.PostNodeCommandShellJSONRequestBody{
					Command: "echo hello",
					Timeout: intPtr(30),
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyCommandShell(
						gomock.Any(),
						"_any",
						"echo hello",
						"",
						30,
					).
					Return(
						"550e8400-e29b-41d4-a716-446655440000",
						&command.Result{
							Stdout:     "hello",
							Stderr:     "",
							ExitCode:   0,
							DurationMs: 5,
							Changed:    false,
						},
						"agent1",
						nil,
					)
			},
			validateFunc: func(resp gen.PostNodeCommandShellResponseObject) {
				r, ok := resp.(gen.PostNodeCommandShell202JSONResponse)
				s.True(ok)
				s.Require().Len(r.Results, 1)
				s.Equal("agent1", r.Results[0].Hostname)
				s.Require().NotNil(r.Results[0].Stdout)
				s.Equal("hello", *r.Results[0].Stdout)
				s.Require().NotNil(r.Results[0].ExitCode)
				s.Equal(0, *r.Results[0].ExitCode)
				s.Require().NotNil(r.Results[0].Changed)
				s.False(*r.Results[0].Changed)
			},
		},
		{
			name: "success with all optional fields",
			request: gen.PostNodeCommandShellRequestObject{
				Hostname: "_any",
				Body: &gen.PostNodeCommandShellJSONRequestBody{
					Command: "echo hello",
					Cwd:     strPtr("/tmp"),
					Timeout: intPtr(30),
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyCommandShell(
						gomock.Any(),
						"_any",
						"echo hello",
						"/tmp",
						30,
					).
					Return(
						"550e8400-e29b-41d4-a716-446655440000",
						&command.Result{
							Stdout:     "hello",
							ExitCode:   0,
							DurationMs: 5,
						},
						"agent1",
						nil,
					)
			},
			validateFunc: func(resp gen.PostNodeCommandShellResponseObject) {
				r, ok := resp.(gen.PostNodeCommandShell202JSONResponse)
				s.True(ok)
				s.Require().Len(r.Results, 1)
				s.Equal("agent1", r.Results[0].Hostname)
			},
		},
		{
			name: "validation error empty hostname",
			request: gen.PostNodeCommandShellRequestObject{
				Hostname: "",
				Body: &gen.PostNodeCommandShellJSONRequestBody{
					Command: "echo hello",
				},
			},
			setupMock: func() {},
			validateFunc: func(resp gen.PostNodeCommandShellResponseObject) {
				r, ok := resp.(gen.PostNodeCommandShell400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "required")
			},
		},
		{
			name: "body validation error empty command",
			request: gen.PostNodeCommandShellRequestObject{
				Hostname: "_any",
				Body: &gen.PostNodeCommandShellJSONRequestBody{
					Command: "",
				},
			},
			setupMock: func() {},
			validateFunc: func(resp gen.PostNodeCommandShellResponseObject) {
				r, ok := resp.(gen.PostNodeCommandShell400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
			},
		},
		{
			name: "job client error",
			request: gen.PostNodeCommandShellRequestObject{
				Hostname: "_any",
				Body: &gen.PostNodeCommandShellJSONRequestBody{
					Command: "echo hello",
					Timeout: intPtr(30),
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyCommandShell(
						gomock.Any(),
						"_any",
						"echo hello",
						"",
						30,
					).
					Return("", nil, "", assert.AnError)
			},
			validateFunc: func(resp gen.PostNodeCommandShellResponseObject) {
				_, ok := resp.(gen.PostNodeCommandShell500JSONResponse)
				s.True(ok)
			},
		},
		{
			name: "broadcast all success",
			request: gen.PostNodeCommandShellRequestObject{
				Hostname: "_all",
				Body: &gen.PostNodeCommandShellJSONRequestBody{
					Command: "echo hello",
					Timeout: intPtr(30),
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyCommandShellBroadcast(
						gomock.Any(),
						"_all",
						"echo hello",
						"",
						30,
					).
					Return(
						"550e8400-e29b-41d4-a716-446655440000",
						map[string]*command.Result{
							"server1": {
								Stdout:   "hello",
								ExitCode: 0,
							},
							"server2": {
								Stdout:   "hello",
								ExitCode: 0,
							},
						},
						map[string]string{},
						nil,
					)
			},
			validateFunc: func(resp gen.PostNodeCommandShellResponseObject) {
				s.NotNil(resp)
			},
		},
		{
			name: "broadcast all with errors",
			request: gen.PostNodeCommandShellRequestObject{
				Hostname: "_all",
				Body: &gen.PostNodeCommandShellJSONRequestBody{
					Command: "echo hello",
					Timeout: intPtr(30),
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyCommandShellBroadcast(
						gomock.Any(),
						"_all",
						"echo hello",
						"",
						30,
					).
					Return(
						"550e8400-e29b-41d4-a716-446655440000",
						map[string]*command.Result{
							"server1": {
								Stdout:   "hello",
								ExitCode: 0,
							},
						},
						map[string]string{
							"server2": "shell not available",
						},
						nil,
					)
			},
			validateFunc: func(resp gen.PostNodeCommandShellResponseObject) {
				r, ok := resp.(gen.PostNodeCommandShell202JSONResponse)
				s.True(ok)
				s.Len(r.Results, 2)
				var foundError bool
				for _, item := range r.Results {
					if item.Error != nil {
						foundError = true
						s.Equal("server2", item.Hostname)
						s.Equal("shell not available", *item.Error)
					}
				}
				s.True(foundError)
			},
		},
		{
			name: "broadcast all error",
			request: gen.PostNodeCommandShellRequestObject{
				Hostname: "_all",
				Body: &gen.PostNodeCommandShellJSONRequestBody{
					Command: "echo hello",
					Timeout: intPtr(30),
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyCommandShellBroadcast(
						gomock.Any(),
						"_all",
						"echo hello",
						"",
						30,
					).
					Return("", nil, nil, assert.AnError)
			},
			validateFunc: func(resp gen.PostNodeCommandShellResponseObject) {
				_, ok := resp.(gen.PostNodeCommandShell500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()

			resp, err := s.handler.PostNodeCommandShell(s.ctx, tt.request)
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func (s *CommandShellPostPublicTestSuite) TestPostCommandShellHTTP() {
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
			path: "/node/server1/command/shell",
			body: `{"command":"echo hello"}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(s.mockCtrl)
				mock.EXPECT().
					ModifyCommandShell(gomock.Any(), "server1", gomock.Any(), gomock.Any(), gomock.Any()).
					Return("550e8400-e29b-41d4-a716-446655440000", &command.Result{
						Stdout:     "hello",
						Stderr:     "",
						ExitCode:   0,
						DurationMs: 15,
						Changed:    true,
					}, "agent1", nil)
				return mock
			},
			wantCode:     http.StatusAccepted,
			wantContains: []string{`"results"`, `"agent1"`, `"changed":true`},
		},
		{
			name: "when missing command",
			path: "/node/server1/command/shell",
			body: `{}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(s.mockCtrl)
			},
			wantCode:     http.StatusBadRequest,
			wantContains: []string{`"error"`, "Command", "required"},
		},
		{
			name: "when invalid timeout",
			path: "/node/server1/command/shell",
			body: `{"command":"echo hello","timeout":999}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(s.mockCtrl)
			},
			wantCode:     http.StatusBadRequest,
			wantContains: []string{`"error"`, "Timeout", "max"},
		},
		{
			name: "when target agent not found",
			path: "/node/nonexistent/command/shell",
			body: `{"command":"echo hello"}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(s.mockCtrl)
			},
			wantCode:     http.StatusBadRequest,
			wantContains: []string{`"error"`, "valid_target", "not found"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			nodeHandler := apinode.New(s.logger, jobMock)
			strictHandler := gen.NewStrictHandler(nodeHandler, nil)

			a := api.New(s.appConfig, s.logger)
			gen.RegisterHandlers(a.Echo, strictHandler)

			req := httptest.NewRequest(
				http.MethodPost,
				tc.path,
				strings.NewReader(tc.body),
			)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			a.Echo.ServeHTTP(rec, req)

			s.Equal(tc.wantCode, rec.Code)
			for _, str := range tc.wantContains {
				s.Contains(rec.Body.String(), str)
			}
		})
	}
}

const rbacShellTestSigningKey = "test-signing-key-for-shell-rbac"

func (s *CommandShellPostPublicTestSuite) TestPostCommandShellRBACHTTP() {
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
					rbacShellTestSigningKey,
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
			name: "when valid token with command:execute returns 202",
			setupAuth: func(req *http.Request) {
				token, err := tokenManager.Generate(
					rbacShellTestSigningKey,
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
					ModifyCommandShell(gomock.Any(), "server1", gomock.Any(), gomock.Any(), gomock.Any()).
					Return(
						"550e8400-e29b-41d4-a716-446655440000",
						&command.Result{
							Stdout:     "hello",
							Stderr:     "",
							ExitCode:   0,
							DurationMs: 10,
							Changed:    true,
						},
						"agent1",
						nil,
					)
				return mock
			},
			wantCode:     http.StatusAccepted,
			wantContains: []string{`"results"`, `"changed":true`},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			appConfig := config.Config{
				API: config.API{
					Server: config.Server{
						Security: config.ServerSecurity{
							SigningKey: rbacShellTestSigningKey,
						},
					},
				},
			}

			server := api.New(appConfig, s.logger)
			handlers := server.GetNodeHandler(jobMock)
			server.RegisterHandlers(handlers)

			req := httptest.NewRequest(
				http.MethodPost,
				"/node/server1/command/shell",
				strings.NewReader(`{"command":"echo hello"}`),
			)
			req.Header.Set("Content-Type", "application/json")
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

func TestCommandShellPostPublicTestSuite(t *testing.T) {
	suite.Run(t, new(CommandShellPostPublicTestSuite))
}

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

	"github.com/retr0h/osapi/internal/authtoken"
	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/controller/api"
	apinode "github.com/retr0h/osapi/internal/controller/api/node"
	"github.com/retr0h/osapi/internal/controller/api/node/gen"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
	"github.com/retr0h/osapi/internal/validation"
)

type FileUndeployPostPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apinode.Node
	ctx           context.Context
	appConfig     config.Config
	logger        *slog.Logger
}

func (s *FileUndeployPostPublicTestSuite) SetupSuite() {
	validation.RegisterTargetValidator(func(_ context.Context) ([]validation.AgentTarget, error) {
		return []validation.AgentTarget{
			{Hostname: "server1", Labels: map[string]string{"group": "web"}},
			{Hostname: "server2"},
		}, nil
	})
}

func (s *FileUndeployPostPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apinode.New(slog.Default(), s.mockJobClient)
	s.ctx = context.Background()
	s.appConfig = config.Config{}
	s.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (s *FileUndeployPostPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *FileUndeployPostPublicTestSuite) TestPostNodeFileUndeploy() {
	tests := []struct {
		name         string
		request      gen.PostNodeFileUndeployRequestObject
		setupMock    func()
		validateFunc func(resp gen.PostNodeFileUndeployResponseObject)
	}{
		{
			name: "when undeploy succeeds",
			request: gen.PostNodeFileUndeployRequestObject{
				Hostname: "_any",
				Body: &gen.PostNodeFileUndeployJSONRequestBody{
					Path: "/etc/cron.d/backup",
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyFileUndeploy(
						gomock.Any(),
						"_any",
						"/etc/cron.d/backup",
					).
					Return(
						"550e8400-e29b-41d4-a716-446655440000",
						"agent1",
						true,
						nil,
					)
			},
			validateFunc: func(resp gen.PostNodeFileUndeployResponseObject) {
				r, ok := resp.(gen.PostNodeFileUndeploy202JSONResponse)
				s.True(ok)
				s.Equal("550e8400-e29b-41d4-a716-446655440000", r.JobId)
				s.Equal("agent1", r.Hostname)
				s.True(r.Changed)
			},
		},
		{
			name: "when invalid hostname",
			request: gen.PostNodeFileUndeployRequestObject{
				Hostname: "",
				Body: &gen.PostNodeFileUndeployJSONRequestBody{
					Path: "/etc/cron.d/backup",
				},
			},
			setupMock: func() {},
			validateFunc: func(resp gen.PostNodeFileUndeployResponseObject) {
				r, ok := resp.(gen.PostNodeFileUndeploy400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "required")
			},
		},
		{
			name: "when validation fails empty path",
			request: gen.PostNodeFileUndeployRequestObject{
				Hostname: "_any",
				Body: &gen.PostNodeFileUndeployJSONRequestBody{
					Path: "",
				},
			},
			setupMock: func() {},
			validateFunc: func(resp gen.PostNodeFileUndeployResponseObject) {
				r, ok := resp.(gen.PostNodeFileUndeploy400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "Path")
			},
		},
		{
			name: "when undeploy fails",
			request: gen.PostNodeFileUndeployRequestObject{
				Hostname: "_any",
				Body: &gen.PostNodeFileUndeployJSONRequestBody{
					Path: "/etc/cron.d/backup",
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyFileUndeploy(
						gomock.Any(),
						"_any",
						"/etc/cron.d/backup",
					).
					Return("", "", false, assert.AnError)
			},
			validateFunc: func(resp gen.PostNodeFileUndeployResponseObject) {
				_, ok := resp.(gen.PostNodeFileUndeploy500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()

			resp, err := s.handler.PostNodeFileUndeploy(s.ctx, tt.request)
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func (s *FileUndeployPostPublicTestSuite) TestPostNodeFileUndeployValidationHTTP() {
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
			path: "/node/server1/file/undeploy",
			body: `{"path":"/etc/cron.d/backup"}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(s.mockCtrl)
				mock.EXPECT().
					ModifyFileUndeploy(gomock.Any(), "server1", "/etc/cron.d/backup").
					Return("550e8400-e29b-41d4-a716-446655440000", "agent1", true, nil)
				return mock
			},
			wantCode:     http.StatusAccepted,
			wantContains: []string{`"job_id"`, `"agent1"`, `"changed":true`},
		},
		{
			name: "when missing path",
			path: "/node/server1/file/undeploy",
			body: `{}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(s.mockCtrl)
			},
			wantCode:     http.StatusBadRequest,
			wantContains: []string{`"error"`, "Path", "required"},
		},
		{
			name: "when server error",
			path: "/node/server1/file/undeploy",
			body: `{"path":"/etc/cron.d/backup"}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(s.mockCtrl)
				mock.EXPECT().
					ModifyFileUndeploy(gomock.Any(), "server1", "/etc/cron.d/backup").
					Return("", "", false, assert.AnError)
				return mock
			},
			wantCode:     http.StatusInternalServerError,
			wantContains: []string{`"error"`},
		},
		{
			name: "when target agent not found",
			path: "/node/nonexistent/file/undeploy",
			body: `{"path":"/etc/cron.d/backup"}`,
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

const rbacFileUndeployTestSigningKey = "test-signing-key-for-file-undeploy-rbac"

func (s *FileUndeployPostPublicTestSuite) TestPostNodeFileUndeployRBACHTTP() {
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
					rbacFileUndeployTestSigningKey,
					[]string{"read"},
					"test-user",
					[]string{"node:read"},
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
			name: "when valid token with file:write returns 202",
			setupAuth: func(req *http.Request) {
				token, err := tokenManager.Generate(
					rbacFileUndeployTestSigningKey,
					[]string{"admin"},
					"test-user",
					[]string{"file:write"},
				)
				s.Require().NoError(err)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			},
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(s.mockCtrl)
				mock.EXPECT().
					ModifyFileUndeploy(gomock.Any(), "server1", "/etc/cron.d/backup").
					Return(
						"550e8400-e29b-41d4-a716-446655440000",
						"agent1",
						true,
						nil,
					)
				return mock
			},
			wantCode:     http.StatusAccepted,
			wantContains: []string{`"job_id"`, `"changed":true`},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			appConfig := config.Config{
				Controller: config.Controller{
					API: config.APIServer{
						Security: config.ServerSecurity{
							SigningKey: rbacFileUndeployTestSigningKey,
						},
					},
				},
			}

			server := api.New(appConfig, s.logger)
			handlers := server.GetNodeHandler(jobMock)
			server.RegisterHandlers(handlers)

			req := httptest.NewRequest(
				http.MethodPost,
				"/node/server1/file/undeploy",
				strings.NewReader(`{"path":"/etc/cron.d/backup"}`),
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

func TestFileUndeployPostPublicTestSuite(t *testing.T) {
	suite.Run(t, new(FileUndeployPostPublicTestSuite))
}

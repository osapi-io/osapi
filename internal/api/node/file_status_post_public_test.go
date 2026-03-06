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
	"github.com/retr0h/osapi/internal/provider/file"
	"github.com/retr0h/osapi/internal/validation"
)

type FileStatusPostPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apinode.Node
	ctx           context.Context
	appConfig     config.Config
	logger        *slog.Logger
}

func (s *FileStatusPostPublicTestSuite) SetupSuite() {
	validation.RegisterTargetValidator(func(_ context.Context) ([]validation.AgentTarget, error) {
		return []validation.AgentTarget{
			{Hostname: "server1", Labels: map[string]string{"group": "web"}},
			{Hostname: "server2"},
		}, nil
	})
}

func (s *FileStatusPostPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apinode.New(slog.Default(), s.mockJobClient)
	s.ctx = context.Background()
	s.appConfig = config.Config{}
	s.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (s *FileStatusPostPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *FileStatusPostPublicTestSuite) TestPostNodeFileStatus() {
	tests := []struct {
		name         string
		request      gen.PostNodeFileStatusRequestObject
		setupMock    func()
		validateFunc func(resp gen.PostNodeFileStatusResponseObject)
	}{
		{
			name: "when success with sha256",
			request: gen.PostNodeFileStatusRequestObject{
				Hostname: "_any",
				Body: &gen.PostNodeFileStatusJSONRequestBody{
					Path: "/etc/nginx/nginx.conf",
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryFileStatus(
						gomock.Any(),
						"_any",
						"/etc/nginx/nginx.conf",
					).
					Return(
						"550e8400-e29b-41d4-a716-446655440000",
						&file.StatusResult{
							Path:   "/etc/nginx/nginx.conf",
							Status: "in-sync",
							SHA256: "abc123def456",
						},
						"agent1",
						nil,
					)
			},
			validateFunc: func(resp gen.PostNodeFileStatusResponseObject) {
				r, ok := resp.(gen.PostNodeFileStatus200JSONResponse)
				s.True(ok)
				s.Equal("550e8400-e29b-41d4-a716-446655440000", r.JobId)
				s.Equal("agent1", r.Hostname)
				s.Equal("/etc/nginx/nginx.conf", r.Path)
				s.Equal("in-sync", r.Status)
				s.Require().NotNil(r.Sha256)
				s.Equal("abc123def456", *r.Sha256)
			},
		},
		{
			name: "when success missing file no sha256",
			request: gen.PostNodeFileStatusRequestObject{
				Hostname: "_any",
				Body: &gen.PostNodeFileStatusJSONRequestBody{
					Path: "/etc/missing.conf",
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryFileStatus(
						gomock.Any(),
						"_any",
						"/etc/missing.conf",
					).
					Return(
						"550e8400-e29b-41d4-a716-446655440000",
						&file.StatusResult{
							Path:   "/etc/missing.conf",
							Status: "missing",
						},
						"agent1",
						nil,
					)
			},
			validateFunc: func(resp gen.PostNodeFileStatusResponseObject) {
				r, ok := resp.(gen.PostNodeFileStatus200JSONResponse)
				s.True(ok)
				s.Equal("missing", r.Status)
				s.Nil(r.Sha256)
			},
		},
		{
			name: "when validation error empty hostname",
			request: gen.PostNodeFileStatusRequestObject{
				Hostname: "",
				Body: &gen.PostNodeFileStatusJSONRequestBody{
					Path: "/etc/nginx/nginx.conf",
				},
			},
			setupMock: func() {},
			validateFunc: func(resp gen.PostNodeFileStatusResponseObject) {
				r, ok := resp.(gen.PostNodeFileStatus400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "required")
			},
		},
		{
			name: "when validation error missing path",
			request: gen.PostNodeFileStatusRequestObject{
				Hostname: "_any",
				Body: &gen.PostNodeFileStatusJSONRequestBody{
					Path: "",
				},
			},
			setupMock: func() {},
			validateFunc: func(resp gen.PostNodeFileStatusResponseObject) {
				r, ok := resp.(gen.PostNodeFileStatus400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "Path")
			},
		},
		{
			name: "when job client error",
			request: gen.PostNodeFileStatusRequestObject{
				Hostname: "_any",
				Body: &gen.PostNodeFileStatusJSONRequestBody{
					Path: "/etc/nginx/nginx.conf",
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryFileStatus(
						gomock.Any(),
						"_any",
						"/etc/nginx/nginx.conf",
					).
					Return("", nil, "", assert.AnError)
			},
			validateFunc: func(resp gen.PostNodeFileStatusResponseObject) {
				_, ok := resp.(gen.PostNodeFileStatus500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()

			resp, err := s.handler.PostNodeFileStatus(s.ctx, tt.request)
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func (s *FileStatusPostPublicTestSuite) TestPostNodeFileStatusHTTP() {
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
			path: "/node/server1/file/status",
			body: `{"path":"/etc/nginx/nginx.conf"}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(s.mockCtrl)
				mock.EXPECT().
					QueryFileStatus(gomock.Any(), "server1", "/etc/nginx/nginx.conf").
					Return("550e8400-e29b-41d4-a716-446655440000", &file.StatusResult{
						Path:   "/etc/nginx/nginx.conf",
						Status: "in-sync",
						SHA256: "abc123",
					}, "agent1", nil)
				return mock
			},
			wantCode:     http.StatusOK,
			wantContains: []string{`"job_id"`, `"agent1"`, `"in-sync"`, `"sha256"`},
		},
		{
			name: "when missing path",
			path: "/node/server1/file/status",
			body: `{}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(s.mockCtrl)
			},
			wantCode:     http.StatusBadRequest,
			wantContains: []string{`"error"`, "Path", "required"},
		},
		{
			name: "when target agent not found",
			path: "/node/nonexistent/file/status",
			body: `{"path":"/etc/nginx/nginx.conf"}`,
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

const rbacFileStatusTestSigningKey = "test-signing-key-for-file-status-rbac"

func (s *FileStatusPostPublicTestSuite) TestPostNodeFileStatusRBACHTTP() {
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
					rbacFileStatusTestSigningKey,
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
			name: "when valid token with file:read returns 200",
			setupAuth: func(req *http.Request) {
				token, err := tokenManager.Generate(
					rbacFileStatusTestSigningKey,
					[]string{"admin"},
					"test-user",
					[]string{"file:read"},
				)
				s.Require().NoError(err)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			},
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(s.mockCtrl)
				mock.EXPECT().
					QueryFileStatus(gomock.Any(), "server1", "/etc/nginx/nginx.conf").
					Return(
						"550e8400-e29b-41d4-a716-446655440000",
						&file.StatusResult{
							Path:   "/etc/nginx/nginx.conf",
							Status: "in-sync",
							SHA256: "abc123",
						},
						"agent1",
						nil,
					)
				return mock
			},
			wantCode:     http.StatusOK,
			wantContains: []string{`"job_id"`, `"in-sync"`},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			appConfig := config.Config{
				API: config.API{
					Server: config.Server{
						Security: config.ServerSecurity{
							SigningKey: rbacFileStatusTestSigningKey,
						},
					},
				},
			}

			server := api.New(appConfig, s.logger)
			handlers := server.GetNodeHandler(jobMock)
			server.RegisterHandlers(handlers)

			req := httptest.NewRequest(
				http.MethodPost,
				"/node/server1/file/status",
				strings.NewReader(`{"path":"/etc/nginx/nginx.conf"}`),
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

func TestFileStatusPostPublicTestSuite(t *testing.T) {
	suite.Run(t, new(FileStatusPostPublicTestSuite))
}

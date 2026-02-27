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

package job_test

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/api"
	apijob "github.com/retr0h/osapi/internal/api/job"
	jobGen "github.com/retr0h/osapi/internal/api/job/gen"
	"github.com/retr0h/osapi/internal/authtoken"
	"github.com/retr0h/osapi/internal/config"
	jobtypes "github.com/retr0h/osapi/internal/job"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
)

type JobStatusIntegrationTestSuite struct {
	suite.Suite
	ctrl *gomock.Controller

	appConfig config.Config
	logger    *slog.Logger
}

func (suite *JobStatusIntegrationTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())

	suite.appConfig = config.Config{}
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (suite *JobStatusIntegrationTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *JobStatusIntegrationTestSuite) TestGetJobStatusValidation() {
	tests := []struct {
		name         string
		setupJobMock func() *jobmocks.MockJobClient
		wantCode     int
		wantContains []string
	}{
		{
			name: "when valid request returns queue stats",
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(suite.ctrl)
				mock.EXPECT().
					GetQueueStats(gomock.Any()).
					Return(&jobtypes.QueueStats{
						TotalJobs: 42,
						StatusCounts: map[string]int{
							"completed": 30,
							"failed":    5,
						},
						OperationCounts: map[string]int{
							"node.hostname.get": 15,
						},
						DLQCount: 2,
					}, nil)
				return mock
			},
			wantCode:     http.StatusOK,
			wantContains: []string{`"total_jobs":42`, `"dlq_count":2`},
		},
		{
			name: "when job client errors returns 500",
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(suite.ctrl)
				mock.EXPECT().
					GetQueueStats(gomock.Any()).
					Return(nil, assert.AnError)
				return mock
			},
			wantCode:     http.StatusInternalServerError,
			wantContains: []string{`"error"`},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			jobHandler := apijob.New(suite.logger, jobMock)
			strictHandler := jobGen.NewStrictHandler(jobHandler, nil)

			a := api.New(suite.appConfig, suite.logger)
			jobGen.RegisterHandlers(a.Echo, strictHandler)

			req := httptest.NewRequest(http.MethodGet, "/job/status", nil)
			rec := httptest.NewRecorder()

			a.Echo.ServeHTTP(rec, req)

			suite.Equal(tc.wantCode, rec.Code)
			for _, s := range tc.wantContains {
				suite.Contains(rec.Body.String(), s)
			}
		})
	}
}

const rbacJobStatusTestSigningKey = "test-signing-key-for-rbac-integration"

func (suite *JobStatusIntegrationTestSuite) TestGetJobStatusRBAC() {
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
					rbacJobStatusTestSigningKey,
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
			name: "when valid token with job:read returns 200",
			setupAuth: func(req *http.Request) {
				token, err := tokenManager.Generate(
					rbacJobStatusTestSigningKey,
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
					GetQueueStats(gomock.Any()).
					Return(&jobtypes.QueueStats{
						TotalJobs: 42,
						StatusCounts: map[string]int{
							"completed": 30,
							"failed":    5,
						},
						OperationCounts: map[string]int{
							"node.hostname.get": 15,
						},
						DLQCount: 2,
					}, nil)
				return mock
			},
			wantCode:     http.StatusOK,
			wantContains: []string{`"total_jobs":42`},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			appConfig := config.Config{
				API: config.API{
					Server: config.Server{
						Security: config.ServerSecurity{
							SigningKey: rbacJobStatusTestSigningKey,
						},
					},
				},
			}

			server := api.New(appConfig, suite.logger)
			handlers := server.GetJobHandler(jobMock)
			server.RegisterHandlers(handlers)

			req := httptest.NewRequest(
				http.MethodGet,
				"/job/status",
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

func TestJobStatusIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(JobStatusIntegrationTestSuite))
}

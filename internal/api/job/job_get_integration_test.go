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
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/api"
	apijob "github.com/retr0h/osapi/internal/api/job"
	jobGen "github.com/retr0h/osapi/internal/api/job/gen"
	"github.com/retr0h/osapi/internal/authtoken"
	"github.com/retr0h/osapi/internal/config"
	jobtypes "github.com/retr0h/osapi/internal/job"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
)

type JobGetIntegrationTestSuite struct {
	suite.Suite
	ctrl *gomock.Controller

	appConfig config.Config
	logger    *slog.Logger
}

func (suite *JobGetIntegrationTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())

	suite.appConfig = config.Config{}
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (suite *JobGetIntegrationTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *JobGetIntegrationTestSuite) TestGetJobByIDValidation() {
	tests := []struct {
		name         string
		jobID        string
		setupJobMock func() *jobmocks.MockJobClient
		wantCode     int
		wantContains []string
	}{
		{
			name:  "when valid uuid",
			jobID: "550e8400-e29b-41d4-a716-446655440000",
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(suite.ctrl)
				mock.EXPECT().
					GetJobStatus(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").
					Return(&jobtypes.QueuedJob{
						ID:      "550e8400-e29b-41d4-a716-446655440000",
						Status:  "completed",
						Created: "2026-02-19T00:00:00Z",
					}, nil)
				return mock
			},
			wantCode: http.StatusOK,
			wantContains: []string{
				`"id":"550e8400-e29b-41d4-a716-446655440000"`,
				`"status":"completed"`,
			},
		},
		{
			name:  "when invalid uuid",
			jobID: "not-a-uuid",
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(suite.ctrl)
			},
			wantCode:     http.StatusBadRequest,
			wantContains: []string{`"message"`, "Invalid format for parameter id"},
		},
		{
			name:  "when job has timeline events",
			jobID: "660e8400-e29b-41d4-a716-446655440000",
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(suite.ctrl)
				mock.EXPECT().
					GetJobStatus(gomock.Any(), "660e8400-e29b-41d4-a716-446655440000").
					Return(&jobtypes.QueuedJob{
						ID:      "660e8400-e29b-41d4-a716-446655440000",
						Status:  "failed",
						Created: "2026-02-19T10:00:00Z",
						Timeline: []jobtypes.TimelineEvent{
							{
								Timestamp: time.Date(2026, 2, 19, 10, 0, 0, 0, time.UTC),
								Event:     "submitted",
								Hostname:  "_api",
								Message:   "Job submitted to queue",
							},
							{
								Timestamp: time.Date(2026, 2, 19, 10, 0, 3, 0, time.UTC),
								Event:     "failed",
								Hostname:  "worker-1",
								Message:   "Job failed on worker-1",
								Error:     "timeout",
							},
						},
					}, nil)
				return mock
			},
			wantCode: http.StatusOK,
			wantContains: []string{
				`"timeline"`,
				`"submitted"`,
				`"failed"`,
				`"Job submitted to queue"`,
				`"timeout"`,
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			jobHandler := apijob.New(suite.logger, jobMock)
			strictHandler := jobGen.NewStrictHandler(jobHandler, nil)

			a := api.New(suite.appConfig, suite.logger)
			jobGen.RegisterHandlers(a.Echo, strictHandler)

			req := httptest.NewRequest(
				http.MethodGet,
				"/job/"+tc.jobID,
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

const rbacJobGetTestSigningKey = "test-signing-key-for-rbac-integration"

func (suite *JobGetIntegrationTestSuite) TestGetJobByIDRBAC() {
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
					rbacJobGetTestSigningKey,
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
					rbacJobGetTestSigningKey,
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
					GetJobStatus(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000").
					Return(&jobtypes.QueuedJob{
						ID:      "550e8400-e29b-41d4-a716-446655440000",
						Status:  "completed",
						Created: "2026-02-19T00:00:00Z",
					}, nil)
				return mock
			},
			wantCode:     http.StatusOK,
			wantContains: []string{`"id":"550e8400-e29b-41d4-a716-446655440000"`},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			appConfig := config.Config{
				API: config.API{
					Server: config.Server{
						Security: config.ServerSecurity{
							SigningKey: rbacJobGetTestSigningKey,
						},
					},
				},
			}

			server := api.New(appConfig, suite.logger)
			handlers := server.GetJobHandler(jobMock)
			server.RegisterHandlers(handlers)

			req := httptest.NewRequest(
				http.MethodGet,
				"/job/550e8400-e29b-41d4-a716-446655440000",
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

func TestJobGetIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(JobGetIntegrationTestSuite))
}

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
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/api"
	apijob "github.com/retr0h/osapi/internal/api/job"
	jobGen "github.com/retr0h/osapi/internal/api/job/gen"
	"github.com/retr0h/osapi/internal/authtoken"
	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/job/client"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
)

type JobRetryIntegrationTestSuite struct {
	suite.Suite
	ctrl *gomock.Controller

	appConfig config.Config
	logger    *slog.Logger
}

func (suite *JobRetryIntegrationTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())

	suite.appConfig = config.Config{}
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (suite *JobRetryIntegrationTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *JobRetryIntegrationTestSuite) TestRetryJobByIDValidation() {
	tests := []struct {
		name         string
		jobID        string
		body         string
		setupJobMock func() *jobmocks.MockJobClient
		wantCode     int
		wantContains []string
	}{
		{
			name:  "when valid request with target",
			jobID: "550e8400-e29b-41d4-a716-446655440000",
			body:  `{"target_hostname":"_any"}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(suite.ctrl)
				mock.EXPECT().
					RetryJob(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000", "_any").
					Return(&client.CreateJobResult{
						JobID:     "660e8400-e29b-41d4-a716-446655440000",
						Status:    "created",
						Revision:  1,
						Timestamp: "2026-02-19T00:00:00Z",
					}, nil)
				return mock
			},
			wantCode: http.StatusCreated,
			wantContains: []string{
				`"job_id":"660e8400-e29b-41d4-a716-446655440000"`,
				`"status":"created"`,
			},
		},
		{
			name:  "when valid request without body",
			jobID: "550e8400-e29b-41d4-a716-446655440000",
			body:  `{}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(suite.ctrl)
				mock.EXPECT().
					RetryJob(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000", "").
					Return(&client.CreateJobResult{
						JobID:     "770e8400-e29b-41d4-a716-446655440000",
						Status:    "created",
						Revision:  1,
						Timestamp: "2026-02-19T00:00:00Z",
					}, nil)
				return mock
			},
			wantCode:     http.StatusCreated,
			wantContains: []string{`"job_id":"770e8400-e29b-41d4-a716-446655440000"`},
		},
		{
			name:  "when invalid uuid",
			jobID: "not-a-uuid",
			body:  `{}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(suite.ctrl)
			},
			wantCode:     http.StatusBadRequest,
			wantContains: []string{`"message"`, "Invalid format for parameter id"},
		},
		{
			name:  "when empty target hostname in body",
			jobID: "550e8400-e29b-41d4-a716-446655440000",
			body:  `{"target_hostname":""}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(suite.ctrl)
			},
			wantCode:     http.StatusBadRequest,
			wantContains: []string{`"error"`, "TargetHostname"},
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
				http.MethodPost,
				"/job/"+tc.jobID+"/retry",
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

const rbacJobRetryTestSigningKey = "test-signing-key-for-rbac-integration"

func (suite *JobRetryIntegrationTestSuite) TestRetryJobByIDRBAC() {
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
					rbacJobRetryTestSigningKey,
					[]string{"read"},
					"test-user",
					[]string{"job:read"},
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
			name: "when valid token with job:write returns 201",
			setupAuth: func(req *http.Request) {
				token, err := tokenManager.Generate(
					rbacJobRetryTestSigningKey,
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
					RetryJob(gomock.Any(), "550e8400-e29b-41d4-a716-446655440000", "_any").
					Return(&client.CreateJobResult{
						JobID:     "660e8400-e29b-41d4-a716-446655440000",
						Status:    "created",
						Revision:  1,
						Timestamp: "2026-02-19T00:00:00Z",
					}, nil)
				return mock
			},
			wantCode:     http.StatusCreated,
			wantContains: []string{`"job_id":"660e8400-e29b-41d4-a716-446655440000"`},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			appConfig := config.Config{
				API: config.API{
					Server: config.Server{
						Security: config.ServerSecurity{
							SigningKey: rbacJobRetryTestSigningKey,
						},
					},
				},
			}

			server := api.New(appConfig, suite.logger)
			handlers := server.GetJobHandler(jobMock)
			server.RegisterHandlers(handlers)

			req := httptest.NewRequest(
				http.MethodPost,
				"/job/550e8400-e29b-41d4-a716-446655440000/retry",
				strings.NewReader(`{"target_hostname":"_any"}`),
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

func TestJobRetryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(JobRetryIntegrationTestSuite))
}

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
	"context"
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
	"github.com/retr0h/osapi/internal/api/job/gen"
	"github.com/retr0h/osapi/internal/authtoken"
	"github.com/retr0h/osapi/internal/config"
	jobtypes "github.com/retr0h/osapi/internal/job"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
)

type JobStatusPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apijob.Job
	ctx           context.Context
	appConfig     config.Config
	logger        *slog.Logger
}

func (s *JobStatusPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apijob.New(slog.Default(), s.mockJobClient)
	s.ctx = context.Background()
	s.appConfig = config.Config{}
	s.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (s *JobStatusPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *JobStatusPublicTestSuite) TestGetJobStatus() {
	tests := []struct {
		name         string
		mockStats    *jobtypes.QueueStats
		mockError    error
		validateFunc func(resp gen.GetJobStatusResponseObject)
	}{
		{
			name: "success",
			mockStats: &jobtypes.QueueStats{
				TotalJobs: 42,
				StatusCounts: map[string]int{
					"completed": 30,
					"failed":    5,
				},
				OperationCounts: map[string]int{
					"node.hostname.get": 15,
				},
				DLQCount: 2,
			},
			validateFunc: func(resp gen.GetJobStatusResponseObject) {
				r, ok := resp.(gen.GetJobStatus200JSONResponse)
				s.True(ok)
				s.Equal(42, *r.TotalJobs)
				s.Equal(2, *r.DlqCount)
			},
		},
		{
			name:      "job client error",
			mockError: assert.AnError,
			validateFunc: func(resp gen.GetJobStatusResponseObject) {
				_, ok := resp.(gen.GetJobStatus500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.mockJobClient.EXPECT().
				GetQueueStats(gomock.Any()).
				Return(tt.mockStats, tt.mockError)

			resp, err := s.handler.GetJobStatus(s.ctx, gen.GetJobStatusRequestObject{})
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func (s *JobStatusPublicTestSuite) TestGetJobStatusHTTP() {
	tests := []struct {
		name         string
		setupJobMock func() *jobmocks.MockJobClient
		wantCode     int
		wantContains []string
	}{
		{
			name: "when valid request returns queue stats",
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(s.mockCtrl)
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
				mock := jobmocks.NewMockJobClient(s.mockCtrl)
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
		s.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			jobHandler := apijob.New(s.logger, jobMock)
			strictHandler := gen.NewStrictHandler(jobHandler, nil)

			a := api.New(s.appConfig, s.logger)
			gen.RegisterHandlers(a.Echo, strictHandler)

			req := httptest.NewRequest(http.MethodGet, "/job/status", nil)
			rec := httptest.NewRecorder()

			a.Echo.ServeHTTP(rec, req)

			s.Equal(tc.wantCode, rec.Code)
			for _, str := range tc.wantContains {
				s.Contains(rec.Body.String(), str)
			}
		})
	}
}

const rbacJobStatusTestSigningKey = "test-signing-key-for-rbac-integration"

func (s *JobStatusPublicTestSuite) TestGetJobStatusRBACHTTP() {
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
					rbacJobStatusTestSigningKey,
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
			name: "when valid token with job:read returns 200",
			setupAuth: func(req *http.Request) {
				token, err := tokenManager.Generate(
					rbacJobStatusTestSigningKey,
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
		s.Run(tc.name, func() {
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

			server := api.New(appConfig, s.logger)
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

			s.Equal(tc.wantCode, rec.Code)
			for _, str := range tc.wantContains {
				s.Contains(rec.Body.String(), str)
			}
		})
	}
}

func TestJobStatusPublicTestSuite(t *testing.T) {
	suite.Run(t, new(JobStatusPublicTestSuite))
}

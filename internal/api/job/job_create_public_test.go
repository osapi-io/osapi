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
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/api"
	apijob "github.com/retr0h/osapi/internal/api/job"
	"github.com/retr0h/osapi/internal/api/job/gen"
	"github.com/retr0h/osapi/internal/authtoken"
	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/job/client"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
)

type JobCreatePublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apijob.Job
	ctx           context.Context
	appConfig     config.Config
	logger        *slog.Logger
}

func (s *JobCreatePublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apijob.New(slog.Default(), s.mockJobClient)
	s.ctx = context.Background()
	s.appConfig = config.Config{}
	s.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (s *JobCreatePublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *JobCreatePublicTestSuite) TestPostJob() {
	tests := []struct {
		name         string
		request      gen.PostJobRequestObject
		mockResult   *client.CreateJobResult
		mockError    error
		expectMock   bool
		validateFunc func(resp gen.PostJobResponseObject)
	}{
		{
			name: "success",
			request: gen.PostJobRequestObject{
				Body: &gen.CreateJobRequest{
					Operation:      map[string]interface{}{"type": "node.hostname.get"},
					TargetHostname: "_any",
				},
			},
			mockResult: &client.CreateJobResult{
				JobID:     "550e8400-e29b-41d4-a716-446655440000",
				Status:    "created",
				Revision:  1,
				Timestamp: "2025-06-14T10:00:00Z",
			},
			expectMock: true,
			validateFunc: func(resp gen.PostJobResponseObject) {
				r, ok := resp.(gen.PostJob201JSONResponse)
				s.True(ok)
				s.Equal("550e8400-e29b-41d4-a716-446655440000", r.JobId.String())
				s.Equal("created", r.Status)
			},
		},
		{
			name: "validation error missing operation",
			request: gen.PostJobRequestObject{
				Body: &gen.PostJobJSONRequestBody{
					TargetHostname: "_any",
				},
			},
			expectMock: false,
			validateFunc: func(resp gen.PostJobResponseObject) {
				r, ok := resp.(gen.PostJob400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "Operation")
				s.Contains(*r.Error, "required")
			},
		},
		{
			name: "validation error empty target hostname",
			request: gen.PostJobRequestObject{
				Body: &gen.PostJobJSONRequestBody{
					Operation: map[string]interface{}{"type": "test"},
				},
			},
			expectMock: false,
			validateFunc: func(resp gen.PostJobResponseObject) {
				r, ok := resp.(gen.PostJob400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "TargetHostname")
				s.Contains(*r.Error, "required")
			},
		},
		{
			name: "job client error",
			request: gen.PostJobRequestObject{
				Body: &gen.CreateJobRequest{
					Operation:      map[string]interface{}{"type": "invalid"},
					TargetHostname: "_any",
				},
			},
			mockError:  assert.AnError,
			expectMock: true,
			validateFunc: func(resp gen.PostJobResponseObject) {
				_, ok := resp.(gen.PostJob500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.expectMock {
				s.mockJobClient.EXPECT().
					CreateJob(gomock.Any(), tt.request.Body.Operation, tt.request.Body.TargetHostname).
					Return(tt.mockResult, tt.mockError)
			}

			resp, err := s.handler.PostJob(s.ctx, tt.request)
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func (s *JobCreatePublicTestSuite) TestPostJobHTTP() {
	tests := []struct {
		name         string
		body         string
		setupJobMock func() *jobmocks.MockJobClient
		wantCode     int
		wantContains []string
	}{
		{
			name: "when valid request",
			body: `{"operation":{"type":"node.hostname.get"},"target_hostname":"_any"}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(s.mockCtrl)
				mock.EXPECT().
					CreateJob(gomock.Any(), gomock.Any(), "_any").
					Return(&client.CreateJobResult{
						JobID:     "550e8400-e29b-41d4-a716-446655440000",
						Status:    "created",
						Revision:  1,
						Timestamp: "2025-06-14T10:00:00Z",
					}, nil)
				return mock
			},
			wantCode: http.StatusCreated,
			wantContains: []string{
				`"job_id":"550e8400-e29b-41d4-a716-446655440000"`,
				`"status":"created"`,
			},
		},
		{
			name: "when missing operation",
			body: `{"target_hostname":"_any"}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(s.mockCtrl)
			},
			wantCode:     http.StatusBadRequest,
			wantContains: []string{`"error"`, "Operation", "required"},
		},
		{
			name: "when empty target hostname",
			body: `{"operation":{"type":"test"}}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(s.mockCtrl)
			},
			wantCode:     http.StatusBadRequest,
			wantContains: []string{`"error"`, "TargetHostname", "required"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			jobHandler := apijob.New(s.logger, jobMock)
			strictHandler := gen.NewStrictHandler(jobHandler, nil)

			a := api.New(s.appConfig, s.logger)
			gen.RegisterHandlers(a.Echo, strictHandler)

			req := httptest.NewRequest(
				http.MethodPost,
				"/job",
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

const rbacJobCreateTestSigningKey = "test-signing-key-for-rbac-integration"

func (s *JobCreatePublicTestSuite) TestPostJobRBACHTTP() {
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
					rbacJobCreateTestSigningKey,
					[]string{"read"},
					"test-user",
					[]string{"job:read"},
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
			name: "when valid token with job:write returns 201",
			setupAuth: func(req *http.Request) {
				token, err := tokenManager.Generate(
					rbacJobCreateTestSigningKey,
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
					CreateJob(gomock.Any(), gomock.Any(), "_any").
					Return(&client.CreateJobResult{
						JobID:     "550e8400-e29b-41d4-a716-446655440000",
						Status:    "created",
						Revision:  1,
						Timestamp: "2025-06-14T10:00:00Z",
					}, nil)
				return mock
			},
			wantCode:     http.StatusCreated,
			wantContains: []string{`"job_id":"550e8400-e29b-41d4-a716-446655440000"`},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			appConfig := config.Config{
				API: config.API{
					Server: config.Server{
						Security: config.ServerSecurity{
							SigningKey: rbacJobCreateTestSigningKey,
						},
					},
				},
			}

			server := api.New(appConfig, s.logger)
			handlers := server.GetJobHandler(jobMock)
			server.RegisterHandlers(handlers)

			req := httptest.NewRequest(
				http.MethodPost,
				"/job",
				strings.NewReader(`{"operation":{"type":"test"},"target_hostname":"_any"}`),
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

func TestJobCreatePublicTestSuite(t *testing.T) {
	suite.Run(t, new(JobCreatePublicTestSuite))
}

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

package schedule_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"k8s.io/utils/ptr"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller/api"
	apischedule "github.com/osapi-io/osapi/internal/controller/api/node/schedule"
	"github.com/osapi-io/osapi/internal/controller/api/node/schedule/gen"
	"github.com/osapi-io/osapi/internal/job"
	jobmocks "github.com/osapi-io/osapi/internal/job/mocks"
	"github.com/osapi-io/osapi/internal/validation"
)

type ScheduleUpdatePublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apischedule.Schedule
	ctx           context.Context
	appConfig     config.Config
	logger        *slog.Logger
}

func (s *ScheduleUpdatePublicTestSuite) SetupSuite() {
	validation.RegisterTargetValidator(func(_ context.Context) ([]validation.AgentTarget, error) {
		return []validation.AgentTarget{
			{Hostname: "server1", Labels: map[string]string{"group": "web"}},
			{Hostname: "server2"},
		}, nil
	})
}

func (s *ScheduleUpdatePublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apischedule.New(slog.Default(), s.mockJobClient)
	s.ctx = context.Background()
	s.appConfig = config.Config{}
	s.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (s *ScheduleUpdatePublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *ScheduleUpdatePublicTestSuite) TestPutNodeSchedule() {
	tests := []struct {
		name         string
		request      gen.PutNodeScheduleRequestObject
		setupMock    func()
		validateFunc func(resp gen.PutNodeScheduleResponseObject)
	}{
		{
			name: "success with all fields",
			request: gen.PutNodeScheduleRequestObject{
				Hostname: "server1",
				Name:     "backup",
				Body: &gen.PutNodeScheduleJSONRequestBody{
					Schedule:    ptr.To("0 3 * * *"),
					Object:      ptr.To("backup-v2-script"),
					User:        ptr.To("admin"),
					ContentType: (*gen.ScheduleUpdateRequestContentType)(ptr.To("template")),
					Vars:        &map[string]interface{}{"region": "us-east"},
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					Modify(
						gomock.Any(),
						"server1",
						"schedule",
						job.OperationCronUpdate,
						gomock.Any(),
					).
					Return(
						"550e8400-e29b-41d4-a716-446655440000",
						&job.Response{
							JobID:    "550e8400-e29b-41d4-a716-446655440000",
							Hostname: "agent1",
							Changed:  ptr.To(true),
							Data:     json.RawMessage(`{"name":"backup","changed":true}`),
						},
						nil,
					)
			},
			validateFunc: func(resp gen.PutNodeScheduleResponseObject) {
				r, ok := resp.(gen.PutNodeSchedule200JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.JobId)
				s.Require().Len(r.Results, 1)
				s.Equal("agent1", r.Results[0].Hostname)
				s.Require().NotNil(r.Results[0].Changed)
				s.True(*r.Results[0].Changed)
				s.Equal("backup", *r.Results[0].Name)
			},
		},
		{
			name: "when empty body returns 400",
			request: gen.PutNodeScheduleRequestObject{
				Hostname: "server1",
				Name:     "backup",
				Body:     &gen.PutNodeScheduleJSONRequestBody{},
			},
			setupMock: func() {},
			validateFunc: func(resp gen.PutNodeScheduleResponseObject) {
				r, ok := resp.(gen.PutNodeSchedule400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "at least one field")
			},
		},
		{
			name: "success with nil response data",
			request: gen.PutNodeScheduleRequestObject{
				Hostname: "server1",
				Name:     "backup",
				Body: &gen.PutNodeScheduleJSONRequestBody{
					Schedule: ptr.To("0 3 * * *"),
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					Modify(
						gomock.Any(),
						"server1",
						"schedule",
						job.OperationCronUpdate,
						gomock.Any(),
					).
					Return(
						"550e8400-e29b-41d4-a716-446655440000",
						&job.Response{
							JobID:    "550e8400-e29b-41d4-a716-446655440000",
							Hostname: "agent1",
							Changed:  ptr.To(true),
							Data:     nil,
						},
						nil,
					)
			},
			validateFunc: func(resp gen.PutNodeScheduleResponseObject) {
				r, ok := resp.(gen.PutNodeSchedule200JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.JobId)
				s.Require().Len(r.Results, 1)
				s.Equal("", *r.Results[0].Name)
			},
		},
		{
			name: "broadcast success",
			request: gen.PutNodeScheduleRequestObject{
				Hostname: "_all",
				Name:     "backup",
				Body: &gen.PutNodeScheduleJSONRequestBody{
					Schedule: ptr.To("0 3 * * *"),
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyBroadcast(
						gomock.Any(),
						"_all",
						"schedule",
						job.OperationCronUpdate,
						gomock.Any(),
					).
					Return("550e8400-e29b-41d4-a716-446655440000", map[string]*job.Response{
						"server1": {
							JobID:    "550e8400-e29b-41d4-a716-446655440000",
							Hostname: "server1",
							Changed:  ptr.To(true),
							Data:     json.RawMessage(`{"name":"backup","changed":true}`),
						},
						"server2": {
							JobID:    "550e8400-e29b-41d4-a716-446655440000",
							Hostname: "server2",
							Changed:  ptr.To(true),
							Data:     json.RawMessage(`{"name":"backup","changed":true}`),
						},
					}, nil)
			},
			validateFunc: func(resp gen.PutNodeScheduleResponseObject) {
				r, ok := resp.(gen.PutNodeSchedule200JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.JobId)
				s.Len(r.Results, 2)
			},
		},
		{
			name: "broadcast with error entries",
			request: gen.PutNodeScheduleRequestObject{
				Hostname: "_all",
				Name:     "backup",
				Body: &gen.PutNodeScheduleJSONRequestBody{
					Schedule: ptr.To("0 3 * * *"),
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyBroadcast(
						gomock.Any(),
						"_all",
						"schedule",
						job.OperationCronUpdate,
						gomock.Any(),
					).
					Return("550e8400-e29b-41d4-a716-446655440000", map[string]*job.Response{
						"server1": {
							JobID:    "550e8400-e29b-41d4-a716-446655440000",
							Hostname: "server1",
							Changed:  ptr.To(true),
							Data:     json.RawMessage(`{"name":"backup","changed":true}`),
						},
						"server2": {
							Status:    job.StatusFailed,
							Error:     "cron entry not found",
							ErrorCode: job.ErrorCodeNotFound,
							Hostname:  "server2",
						},
					}, nil)
			},
			validateFunc: func(resp gen.PutNodeScheduleResponseObject) {
				r, ok := resp.(gen.PutNodeSchedule200JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.JobId)
				s.Len(r.Results, 2)
				errCount := 0
				for _, res := range r.Results {
					if res.Error != nil {
						errCount++
						s.Equal("cron entry not found", *res.Error)
					}
				}
				s.Equal(1, errCount)
			},
		},
		{
			name: "broadcast with skipped host",
			request: gen.PutNodeScheduleRequestObject{
				Hostname: "_all",
				Name:     "backup",
				Body: &gen.PutNodeScheduleJSONRequestBody{
					Schedule: ptr.To("0 3 * * *"),
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyBroadcast(
						gomock.Any(),
						"_all",
						"schedule",
						job.OperationCronUpdate,
						gomock.Any(),
					).
					Return("550e8400-e29b-41d4-a716-446655440000", map[string]*job.Response{
						"server1": {
							Status:   job.StatusSkipped,
							Error:    "cron: operation not supported on this OS family",
							Hostname: "server1",
						},
					}, nil)
			},
			validateFunc: func(resp gen.PutNodeScheduleResponseObject) {
				r, ok := resp.(gen.PutNodeSchedule200JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.JobId)
				s.Require().Len(r.Results, 1)
				s.Equal(gen.ScheduleMutationResultStatusSkipped, r.Results[0].Status)
				s.Require().NotNil(r.Results[0].Error)
				s.Contains(*r.Results[0].Error, "not supported")
			},
		},
		{
			name: "broadcast with a host that never answered",
			request: gen.PutNodeScheduleRequestObject{
				Hostname: "_all",
				Name:     "backup",
				Body: &gen.PutNodeScheduleJSONRequestBody{
					Schedule: ptr.To("0 3 * * *"),
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyBroadcast(
						gomock.Any(),
						"_all",
						"schedule",
						job.OperationCronUpdate,
						gomock.Any(),
					).
					Return("550e8400-e29b-41d4-a716-446655440000", map[string]*job.Response{
						"server1": {
							Status:   job.StatusTimeout,
							Error:    "timeout: agent did not respond",
							Hostname: "server1",
						},
					}, nil)
			},
			validateFunc: func(resp gen.PutNodeScheduleResponseObject) {
				r, ok := resp.(gen.PutNodeSchedule200JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.JobId)
				s.Require().Len(r.Results, 1)
				s.Equal(gen.ScheduleMutationResultStatusTimeout, r.Results[0].Status)
				s.Require().NotNil(r.Results[0].Error)
				s.Contains(*r.Results[0].Error, "timeout: agent did not respond")
			},
		},
		{
			name: "broadcast error collecting responses",
			request: gen.PutNodeScheduleRequestObject{
				Hostname: "_all",
				Name:     "backup",
				Body: &gen.PutNodeScheduleJSONRequestBody{
					Schedule: ptr.To("0 3 * * *"),
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyBroadcast(
						gomock.Any(),
						"_all",
						"schedule",
						job.OperationCronUpdate,
						gomock.Any(),
					).
					Return("", nil, assert.AnError)
			},
			validateFunc: func(resp gen.PutNodeScheduleResponseObject) {
				_, ok := resp.(gen.PutNodeSchedule500JSONResponse)
				s.True(ok)
			},
		},
		{
			name: "body validation error schedule too short",
			request: gen.PutNodeScheduleRequestObject{
				Hostname: "server1",
				Name:     "backup",
				Body: &gen.PutNodeScheduleJSONRequestBody{
					Schedule: ptr.To("short"),
				},
			},
			setupMock: func() {},
			validateFunc: func(resp gen.PutNodeScheduleResponseObject) {
				r, ok := resp.(gen.PutNodeSchedule400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "Schedule")
			},
		},
		{
			name: "validation error empty hostname",
			request: gen.PutNodeScheduleRequestObject{
				Hostname: "",
				Name:     "backup",
				Body: &gen.PutNodeScheduleJSONRequestBody{
					Schedule: ptr.To("0 3 * * *"),
				},
			},
			setupMock: func() {},
			validateFunc: func(resp gen.PutNodeScheduleResponseObject) {
				r, ok := resp.(gen.PutNodeSchedule400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "required")
			},
		},
		{
			name: "not found error",
			request: gen.PutNodeScheduleRequestObject{
				Hostname: "server1",
				Name:     "nonexistent",
				Body: &gen.PutNodeScheduleJSONRequestBody{
					Schedule: ptr.To("0 3 * * *"),
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					Modify(
						gomock.Any(),
						"server1",
						"schedule",
						job.OperationCronUpdate,
						gomock.Any(),
					).
					Return("", nil, fmt.Errorf("cron entry not found: %w", job.ErrNotFound))
			},
			validateFunc: func(resp gen.PutNodeScheduleResponseObject) {
				r, ok := resp.(gen.PutNodeSchedule404JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "not found")
			},
		},
		{
			name: "does not exist error",
			request: gen.PutNodeScheduleRequestObject{
				Hostname: "server1",
				Name:     "missing",
				Body: &gen.PutNodeScheduleJSONRequestBody{
					Schedule: ptr.To("0 3 * * *"),
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					Modify(
						gomock.Any(),
						"server1",
						"schedule",
						job.OperationCronUpdate,
						gomock.Any(),
					).
					Return("", nil, fmt.Errorf("cron entry does not exist: %w", job.ErrNotFound))
			},
			validateFunc: func(resp gen.PutNodeScheduleResponseObject) {
				r, ok := resp.(gen.PutNodeSchedule404JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "does not exist")
			},
		},
		{
			name: "when job skipped",
			request: gen.PutNodeScheduleRequestObject{
				Hostname: "server1",
				Name:     "backup",
				Body: &gen.PutNodeScheduleJSONRequestBody{
					Schedule: ptr.To("0 3 * * *"),
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					Modify(
						gomock.Any(),
						"server1",
						"schedule",
						job.OperationCronUpdate,
						gomock.Any(),
					).
					Return(
						"550e8400-e29b-41d4-a716-446655440000",
						&job.Response{
							Status:   job.StatusSkipped,
							Hostname: "server1",
							Error:    "cron: operation not supported on this OS family",
						},
						nil,
					)
			},
			validateFunc: func(resp gen.PutNodeScheduleResponseObject) {
				r, ok := resp.(gen.PutNodeSchedule200JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.JobId)
				s.Require().Len(r.Results, 1)
				s.Equal("server1", r.Results[0].Hostname)
				s.Equal(gen.ScheduleMutationResultStatusSkipped, r.Results[0].Status)
				s.Require().NotNil(r.Results[0].Error)
				s.Contains(*r.Results[0].Error, "not supported")
			},
		},
		{
			name: "job client error",
			request: gen.PutNodeScheduleRequestObject{
				Hostname: "server1",
				Name:     "backup",
				Body: &gen.PutNodeScheduleJSONRequestBody{
					Schedule: ptr.To("0 3 * * *"),
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					Modify(
						gomock.Any(),
						"server1",
						"schedule",
						job.OperationCronUpdate,
						gomock.Any(),
					).
					Return("", nil, assert.AnError)
			},
			validateFunc: func(resp gen.PutNodeScheduleResponseObject) {
				_, ok := resp.(gen.PutNodeSchedule500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()

			resp, err := s.handler.PutNodeSchedule(s.ctx, tt.request)
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func (s *ScheduleUpdatePublicTestSuite) TestPutNodeScheduleValidationHTTP() {
	tests := []struct {
		name         string
		path         string
		body         string
		setupJobMock func() *jobmocks.MockJobClient
		validateFunc func(*httptest.ResponseRecorder)
	}{
		{
			name: "when valid request",
			path: "/api/node/server1/schedule/backup",
			body: `{"schedule":"0 3 * * *"}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(s.mockCtrl)
				mock.EXPECT().
					Modify(gomock.Any(), "server1", "schedule", job.OperationCronUpdate, gomock.Any()).
					Return(
						"550e8400-e29b-41d4-a716-446655440000",
						&job.Response{
							JobID:    "550e8400-e29b-41d4-a716-446655440000",
							Hostname: "agent1",
							Changed:  ptr.To(true),
							Data:     json.RawMessage(`{"name":"backup","changed":true}`),
						},
						nil,
					)
				return mock
			},
			validateFunc: func(rec *httptest.ResponseRecorder) {
				s.Equal(http.StatusOK, rec.Code)
				s.Contains(rec.Body.String(), `"job_id"`)
				s.Contains(rec.Body.String(), `"results"`)
			},
		},
		{
			name: "when empty body returns 400",
			path: "/api/node/server1/schedule/backup",
			body: `{}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(s.mockCtrl)
			},
			validateFunc: func(rec *httptest.ResponseRecorder) {
				s.Equal(http.StatusBadRequest, rec.Code)
				s.Contains(rec.Body.String(), `"error"`)
				s.Contains(rec.Body.String(), "at least one field")
			},
		},
		{
			name: "when schedule is not a cron expression returns 400",
			path: "/api/node/server1/schedule/backup",
			body: `{"schedule":"not-a-cron"}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(s.mockCtrl)
			},
			validateFunc: func(rec *httptest.ResponseRecorder) {
				s.Equal(http.StatusBadRequest, rec.Code)
				s.Contains(rec.Body.String(), `"error"`)
				s.Contains(rec.Body.String(), "Schedule")
				s.Contains(rec.Body.String(), "cron_schedule")
				s.Contains(rec.Body.String(), "not a valid cron expression")
			},
		},
		{
			name: "when content type is neither raw nor template returns 400",
			path: "/api/node/server1/schedule/backup",
			body: `{"content_type":"jinja"}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(s.mockCtrl)
			},
			validateFunc: func(rec *httptest.ResponseRecorder) {
				s.Equal(http.StatusBadRequest, rec.Code)
				s.Contains(rec.Body.String(), "ContentType")
				s.Contains(rec.Body.String(), "oneof")
			},
		},
		{
			name: "when target agent not found",
			path: "/api/node/nonexistent/schedule/backup",
			body: `{"schedule":"0 3 * * *"}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(s.mockCtrl)
			},
			validateFunc: func(rec *httptest.ResponseRecorder) {
				s.Equal(http.StatusBadRequest, rec.Code)
				s.Contains(rec.Body.String(), `"error"`)
				s.Contains(rec.Body.String(), "valid_target")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			scheduleHandler := apischedule.New(s.logger, jobMock)
			strictHandler := gen.NewStrictHandler(scheduleHandler, nil)

			a := api.New(s.appConfig, s.logger)
			gen.RegisterHandlers(a.Echo, strictHandler)

			req := httptest.NewRequest(
				http.MethodPut,
				tc.path,
				strings.NewReader(tc.body),
			)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			a.Echo.ServeHTTP(rec, req)

			tc.validateFunc(rec)
		})
	}
}

const rbacCronUpdateTestSigningKey = "test-signing-key-for-rbac-cron-update"

func (s *ScheduleUpdatePublicTestSuite) TestPutNodeScheduleRBACHTTP() {
	tokenManager := authtoken.New(s.logger)

	tests := []struct {
		name         string
		setupAuth    func(req *http.Request)
		setupJobMock func() *jobmocks.MockJobClient
		validateFunc func(*httptest.ResponseRecorder)
	}{
		{
			name: "when no token returns 401",
			setupAuth: func(_ *http.Request) {
				// No auth header set
			},
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(s.mockCtrl)
			},
			validateFunc: func(rec *httptest.ResponseRecorder) {
				s.Equal(http.StatusUnauthorized, rec.Code)
				s.Contains(rec.Body.String(), "Bearer token required")
			},
		},
		{
			name: "when insufficient permissions returns 403",
			setupAuth: func(req *http.Request) {
				token, err := tokenManager.Generate(
					rbacCronUpdateTestSigningKey,
					[]string{"read"},
					"test-user",
					[]string{"schedule:read"},
				)
				s.Require().NoError(err)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			},
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(s.mockCtrl)
			},
			validateFunc: func(rec *httptest.ResponseRecorder) {
				s.Equal(http.StatusForbidden, rec.Code)
				s.Contains(rec.Body.String(), "Insufficient permissions")
			},
		},
		{
			name: "when valid admin token returns 200",
			setupAuth: func(req *http.Request) {
				token, err := tokenManager.Generate(
					rbacCronUpdateTestSigningKey,
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
					Modify(gomock.Any(), "server1", "schedule", job.OperationCronUpdate, gomock.Any()).
					Return(
						"550e8400-e29b-41d4-a716-446655440000",
						&job.Response{
							JobID:    "550e8400-e29b-41d4-a716-446655440000",
							Hostname: "agent1",
							Changed:  ptr.To(true),
							Data:     json.RawMessage(`{"name":"backup","changed":true}`),
						},
						nil,
					)
				return mock
			},
			validateFunc: func(rec *httptest.ResponseRecorder) {
				s.Equal(http.StatusOK, rec.Code)
				s.Contains(rec.Body.String(), `"job_id"`)
				s.Contains(rec.Body.String(), `"results"`)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			appConfig := config.Config{
				Controller: config.Controller{
					API: config.APIServer{
						Security: config.ServerSecurity{
							SigningKey: rbacCronUpdateTestSigningKey,
						},
					},
				},
			}

			server := api.New(appConfig, s.logger)
			handlers := apischedule.Handler(
				s.logger,
				jobMock,
				appConfig.Controller.API.Security.SigningKey,
				nil,
			)
			server.RegisterHandlers(handlers)

			req := httptest.NewRequest(
				http.MethodPut,
				"/api/node/server1/schedule/backup",
				strings.NewReader(`{"schedule":"0 3 * * *"}`),
			)
			req.Header.Set("Content-Type", "application/json")
			tc.setupAuth(req)
			rec := httptest.NewRecorder()

			server.Echo.ServeHTTP(rec, req)

			tc.validateFunc(rec)
		})
	}
}

func TestScheduleUpdatePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ScheduleUpdatePublicTestSuite))
}

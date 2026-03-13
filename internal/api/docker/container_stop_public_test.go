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

package container_test

import (
	"context"
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
	apicontainer "github.com/retr0h/osapi/internal/api/container"
	"github.com/retr0h/osapi/internal/api/container/gen"
	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/job"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
	"github.com/retr0h/osapi/internal/validation"
)

type ContainerStopPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apicontainer.Container
	ctx           context.Context
	appConfig     config.Config
	logger        *slog.Logger
}

func (s *ContainerStopPublicTestSuite) SetupSuite() {
	validation.RegisterTargetValidator(func(_ context.Context) ([]validation.AgentTarget, error) {
		return []validation.AgentTarget{
			{Hostname: "server1", Labels: map[string]string{"group": "web"}},
			{Hostname: "server2"},
		}, nil
	})
}

func (s *ContainerStopPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apicontainer.New(slog.Default(), s.mockJobClient)
	s.ctx = context.Background()
	s.appConfig = config.Config{}
	s.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (s *ContainerStopPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *ContainerStopPublicTestSuite) TestPostNodeContainerStop() {
	tests := []struct {
		name         string
		request      gen.PostNodeContainerStopRequestObject
		setupMock    func()
		validateFunc func(resp gen.PostNodeContainerStopResponseObject)
	}{
		{
			name: "success without body",
			request: gen.PostNodeContainerStopRequestObject{
				Hostname: "server1",
				Id:       "abc123",
				Body:     nil,
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyContainerStop(
						gomock.Any(),
						"server1",
						"abc123",
						gomock.Any(),
					).
					Return(&job.Response{
						JobID:    "550e8400-e29b-41d4-a716-446655440000",
						Hostname: "agent1",
						Changed:  boolPtr(true),
					}, nil)
			},
			validateFunc: func(resp gen.PostNodeContainerStopResponseObject) {
				r, ok := resp.(gen.PostNodeContainerStop202JSONResponse)
				s.True(ok)
				s.Require().Len(r.Results, 1)
				s.Equal("agent1", r.Results[0].Hostname)
				s.Require().NotNil(r.Results[0].Id)
				s.Equal("abc123", *r.Results[0].Id)
				s.Require().NotNil(r.Results[0].Changed)
				s.True(*r.Results[0].Changed)
				s.Require().NotNil(r.Results[0].Message)
				s.Equal("container stopped", *r.Results[0].Message)
			},
		},
		{
			name: "success with timeout",
			request: gen.PostNodeContainerStopRequestObject{
				Hostname: "server1",
				Id:       "abc123",
				Body: &gen.PostNodeContainerStopJSONRequestBody{
					Timeout: intPtr(30),
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyContainerStop(
						gomock.Any(),
						"server1",
						"abc123",
						gomock.Any(),
					).
					Return(&job.Response{
						JobID:    "550e8400-e29b-41d4-a716-446655440000",
						Hostname: "agent1",
						Changed:  boolPtr(true),
					}, nil)
			},
			validateFunc: func(resp gen.PostNodeContainerStopResponseObject) {
				r, ok := resp.(gen.PostNodeContainerStop202JSONResponse)
				s.True(ok)
				s.Require().Len(r.Results, 1)
				s.Equal("agent1", r.Results[0].Hostname)
			},
		},
		{
			name: "body validation error invalid timeout",
			request: gen.PostNodeContainerStopRequestObject{
				Hostname: "server1",
				Id:       "abc123",
				Body: &gen.PostNodeContainerStopJSONRequestBody{
					Timeout: intPtr(999),
				},
			},
			setupMock: func() {},
			validateFunc: func(resp gen.PostNodeContainerStopResponseObject) {
				r, ok := resp.(gen.PostNodeContainerStop400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "Timeout")
			},
		},
		{
			name: "validation error empty hostname",
			request: gen.PostNodeContainerStopRequestObject{
				Hostname: "",
				Id:       "abc123",
			},
			setupMock: func() {},
			validateFunc: func(resp gen.PostNodeContainerStopResponseObject) {
				r, ok := resp.(gen.PostNodeContainerStop400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "required")
			},
		},
		{
			name: "job client error",
			request: gen.PostNodeContainerStopRequestObject{
				Hostname: "server1",
				Id:       "abc123",
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyContainerStop(
						gomock.Any(),
						"server1",
						"abc123",
						gomock.Any(),
					).
					Return(nil, assert.AnError)
			},
			validateFunc: func(resp gen.PostNodeContainerStopResponseObject) {
				_, ok := resp.(gen.PostNodeContainerStop500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()

			resp, err := s.handler.PostNodeContainerStop(s.ctx, tt.request)
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func (s *ContainerStopPublicTestSuite) TestPostNodeContainerStopValidationHTTP() {
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
			path: "/node/server1/container/abc123/stop",
			body: `{"timeout":10}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(s.mockCtrl)
				mock.EXPECT().
					ModifyContainerStop(gomock.Any(), "server1", "abc123", gomock.Any()).
					Return(&job.Response{
						JobID:    "550e8400-e29b-41d4-a716-446655440000",
						Hostname: "agent1",
						Changed:  boolPtr(true),
					}, nil)
				return mock
			},
			wantCode:     http.StatusAccepted,
			wantContains: []string{`"job_id"`, `"results"`, `"container stopped"`},
		},
		{
			name: "when invalid timeout",
			path: "/node/server1/container/abc123/stop",
			body: `{"timeout":999}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(s.mockCtrl)
			},
			wantCode:     http.StatusBadRequest,
			wantContains: []string{`"error"`, "Timeout"},
		},
		{
			name: "when target agent not found",
			path: "/node/nonexistent/container/abc123/stop",
			body: `{}`,
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

			containerHandler := apicontainer.New(s.logger, jobMock)
			strictHandler := gen.NewStrictHandler(containerHandler, nil)

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

func TestContainerStopPublicTestSuite(t *testing.T) {
	suite.Run(t, new(ContainerStopPublicTestSuite))
}

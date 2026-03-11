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
	"encoding/json"
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

type ContainerExecPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apicontainer.Container
	ctx           context.Context
	appConfig     config.Config
	logger        *slog.Logger
}

func (s *ContainerExecPublicTestSuite) SetupSuite() {
	validation.RegisterTargetValidator(func(_ context.Context) ([]validation.AgentTarget, error) {
		return []validation.AgentTarget{
			{Hostname: "server1", Labels: map[string]string{"group": "web"}},
			{Hostname: "server2"},
		}, nil
	})
}

func (s *ContainerExecPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apicontainer.New(slog.Default(), s.mockJobClient)
	s.ctx = context.Background()
	s.appConfig = config.Config{}
	s.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (s *ContainerExecPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *ContainerExecPublicTestSuite) TestPostNodeContainerExec() {
	tests := []struct {
		name         string
		request      gen.PostNodeContainerExecRequestObject
		setupMock    func()
		validateFunc func(resp gen.PostNodeContainerExecResponseObject)
	}{
		{
			name: "success",
			request: gen.PostNodeContainerExecRequestObject{
				Hostname: "server1",
				Id:       "abc123",
				Body: &gen.PostNodeContainerExecJSONRequestBody{
					Command: []string{"ls", "-la"},
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyContainerExec(
						gomock.Any(),
						"server1",
						"abc123",
						gomock.Any(),
					).
					Return(&job.Response{
						JobID:    "550e8400-e29b-41d4-a716-446655440000",
						Hostname: "agent1",
						Changed:  boolPtr(true),
						Data:     json.RawMessage(`{"stdout":"file1\nfile2","stderr":"","exit_code":0}`),
					}, nil)
			},
			validateFunc: func(resp gen.PostNodeContainerExecResponseObject) {
				r, ok := resp.(gen.PostNodeContainerExec202JSONResponse)
				s.True(ok)
				s.Require().Len(r.Results, 1)
				s.Equal("agent1", r.Results[0].Hostname)
				s.Require().NotNil(r.Results[0].Stdout)
				s.Equal("file1\nfile2", *r.Results[0].Stdout)
				s.Require().NotNil(r.Results[0].ExitCode)
				s.Equal(0, *r.Results[0].ExitCode)
				s.Require().NotNil(r.Results[0].Changed)
				s.True(*r.Results[0].Changed)
			},
		},
		{
			name: "validation error empty hostname",
			request: gen.PostNodeContainerExecRequestObject{
				Hostname: "",
				Id:       "abc123",
				Body: &gen.PostNodeContainerExecJSONRequestBody{
					Command: []string{"ls"},
				},
			},
			setupMock: func() {},
			validateFunc: func(resp gen.PostNodeContainerExecResponseObject) {
				r, ok := resp.(gen.PostNodeContainerExec400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "required")
			},
		},
		{
			name: "body validation error empty command",
			request: gen.PostNodeContainerExecRequestObject{
				Hostname: "server1",
				Id:       "abc123",
				Body: &gen.PostNodeContainerExecJSONRequestBody{
					Command: []string{},
				},
			},
			setupMock: func() {},
			validateFunc: func(resp gen.PostNodeContainerExecResponseObject) {
				r, ok := resp.(gen.PostNodeContainerExec400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
			},
		},
		{
			name: "job client error",
			request: gen.PostNodeContainerExecRequestObject{
				Hostname: "server1",
				Id:       "abc123",
				Body: &gen.PostNodeContainerExecJSONRequestBody{
					Command: []string{"ls"},
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyContainerExec(
						gomock.Any(),
						"server1",
						"abc123",
						gomock.Any(),
					).
					Return(nil, assert.AnError)
			},
			validateFunc: func(resp gen.PostNodeContainerExecResponseObject) {
				_, ok := resp.(gen.PostNodeContainerExec500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()

			resp, err := s.handler.PostNodeContainerExec(s.ctx, tt.request)
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func (s *ContainerExecPublicTestSuite) TestPostNodeContainerExecValidationHTTP() {
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
			path: "/node/server1/container/abc123/exec",
			body: `{"command":["ls","-la"]}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(s.mockCtrl)
				mock.EXPECT().
					ModifyContainerExec(gomock.Any(), "server1", "abc123", gomock.Any()).
					Return(&job.Response{
						JobID:    "550e8400-e29b-41d4-a716-446655440000",
						Hostname: "agent1",
						Changed:  boolPtr(true),
						Data:     json.RawMessage(`{"stdout":"output","stderr":"","exit_code":0}`),
					}, nil)
				return mock
			},
			wantCode:     http.StatusAccepted,
			wantContains: []string{`"results"`, `"stdout"`},
		},
		{
			name: "when missing command",
			path: "/node/server1/container/abc123/exec",
			body: `{}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(s.mockCtrl)
			},
			wantCode:     http.StatusBadRequest,
			wantContains: []string{`"error"`, "Command", "required"},
		},
		{
			name: "when target agent not found",
			path: "/node/nonexistent/container/abc123/exec",
			body: `{"command":["ls"]}`,
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

func TestContainerExecPublicTestSuite(t *testing.T) {
	suite.Run(t, new(ContainerExecPublicTestSuite))
}

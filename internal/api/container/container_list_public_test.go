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

type ContainerListPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apicontainer.Container
	ctx           context.Context
	appConfig     config.Config
	logger        *slog.Logger
}

func (s *ContainerListPublicTestSuite) SetupSuite() {
	validation.RegisterTargetValidator(func(_ context.Context) ([]validation.AgentTarget, error) {
		return []validation.AgentTarget{
			{Hostname: "server1", Labels: map[string]string{"group": "web"}},
			{Hostname: "server2"},
		}, nil
	})
}

func (s *ContainerListPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apicontainer.New(slog.Default(), s.mockJobClient)
	s.ctx = context.Background()
	s.appConfig = config.Config{}
	s.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (s *ContainerListPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *ContainerListPublicTestSuite) TestGetNodeContainer() {
	stateAll := gen.All

	tests := []struct {
		name         string
		request      gen.GetNodeContainerRequestObject
		setupMock    func()
		validateFunc func(resp gen.GetNodeContainerResponseObject)
	}{
		{
			name: "success",
			request: gen.GetNodeContainerRequestObject{
				Hostname: "server1",
				Params: gen.GetNodeContainerParams{
					State: &stateAll,
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryContainerList(
						gomock.Any(),
						"server1",
						gomock.Any(),
					).
					Return(&job.Response{
						JobID:    "550e8400-e29b-41d4-a716-446655440000",
						Hostname: "agent1",
						Data: json.RawMessage(`[{"id":"abc123","name":"my-nginx","image":"nginx:latest","state":"running"}]`),
					}, nil)
			},
			validateFunc: func(resp gen.GetNodeContainerResponseObject) {
				r, ok := resp.(gen.GetNodeContainer200JSONResponse)
				s.True(ok)
				s.Require().Len(r.Results, 1)
				s.Equal("agent1", r.Results[0].Hostname)
				s.Require().NotNil(r.Results[0].Containers)
				s.Len(*r.Results[0].Containers, 1)
				c := (*r.Results[0].Containers)[0]
				s.Equal("abc123", *c.Id)
				s.Equal("my-nginx", *c.Name)
			},
		},
		{
			name: "validation error empty hostname",
			request: gen.GetNodeContainerRequestObject{
				Hostname: "",
				Params:   gen.GetNodeContainerParams{},
			},
			setupMock: func() {},
			validateFunc: func(resp gen.GetNodeContainerResponseObject) {
				r, ok := resp.(gen.GetNodeContainer400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "required")
			},
		},
		{
			name: "job client error",
			request: gen.GetNodeContainerRequestObject{
				Hostname: "server1",
				Params:   gen.GetNodeContainerParams{},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryContainerList(
						gomock.Any(),
						"server1",
						gomock.Any(),
					).
					Return(nil, assert.AnError)
			},
			validateFunc: func(resp gen.GetNodeContainerResponseObject) {
				_, ok := resp.(gen.GetNodeContainer500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()

			resp, err := s.handler.GetNodeContainer(s.ctx, tt.request)
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func (s *ContainerListPublicTestSuite) TestGetNodeContainerValidationHTTP() {
	tests := []struct {
		name         string
		path         string
		setupJobMock func() *jobmocks.MockJobClient
		wantCode     int
		wantContains []string
	}{
		{
			name: "when valid request",
			path: "/node/server1/container",
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(s.mockCtrl)
				mock.EXPECT().
					QueryContainerList(gomock.Any(), "server1", gomock.Any()).
					Return(&job.Response{
						JobID:    "550e8400-e29b-41d4-a716-446655440000",
						Hostname: "agent1",
						Data:     json.RawMessage(`[]`),
					}, nil)
				return mock
			},
			wantCode:     http.StatusOK,
			wantContains: []string{`"results"`},
		},
		{
			name: "when target agent not found",
			path: "/node/nonexistent/container",
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

			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()

			a.Echo.ServeHTTP(rec, req)

			s.Equal(tc.wantCode, rec.Code)
			for _, str := range tc.wantContains {
				s.Contains(rec.Body.String(), str)
			}
		})
	}
}

func TestContainerListPublicTestSuite(t *testing.T) {
	suite.Run(t, new(ContainerListPublicTestSuite))
}

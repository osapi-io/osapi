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
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/api"
	apinode "github.com/retr0h/osapi/internal/api/node"
	"github.com/retr0h/osapi/internal/api/node/gen"
	"github.com/retr0h/osapi/internal/authtoken"
	"github.com/retr0h/osapi/internal/config"
	jobtypes "github.com/retr0h/osapi/internal/job"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
	"github.com/retr0h/osapi/internal/provider/node/disk"
	"github.com/retr0h/osapi/internal/provider/node/host"
	"github.com/retr0h/osapi/internal/provider/node/load"
	"github.com/retr0h/osapi/internal/provider/node/mem"
	"github.com/retr0h/osapi/internal/validation"
)

type NodeStatusGetPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apinode.Node
	ctx           context.Context
	appConfig     config.Config
	logger        *slog.Logger
}

func (s *NodeStatusGetPublicTestSuite) SetupSuite() {
	validation.RegisterTargetValidator(func(_ context.Context) ([]validation.AgentTarget, error) {
		return []validation.AgentTarget{
			{Hostname: "server1", Labels: map[string]string{"group": "web"}},
			{Hostname: "server2"},
		}, nil
	})
}

func (s *NodeStatusGetPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apinode.New(slog.Default(), s.mockJobClient)
	s.ctx = context.Background()
	s.appConfig = config.Config{}
	s.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (s *NodeStatusGetPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *NodeStatusGetPublicTestSuite) TestGetNodeStatus() {
	tests := []struct {
		name         string
		request      gen.GetNodeStatusRequestObject
		setupMock    func()
		validateFunc func(resp gen.GetNodeStatusResponseObject)
	}{
		{
			name:    "success",
			request: gen.GetNodeStatusRequestObject{Hostname: "_any"},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNodeStatus(gomock.Any(), "_any").
					Return("550e8400-e29b-41d4-a716-446655440000", &jobtypes.NodeStatusResponse{
						Hostname: "test-host",
						Uptime:   time.Hour,
					}, nil)
			},
			validateFunc: func(resp gen.GetNodeStatusResponseObject) {
				_, ok := resp.(gen.GetNodeStatus200JSONResponse)
				s.True(ok)
			},
		},
		{
			name:      "validation error empty hostname",
			request:   gen.GetNodeStatusRequestObject{Hostname: ""},
			setupMock: func() {},
			validateFunc: func(resp gen.GetNodeStatusResponseObject) {
				r, ok := resp.(gen.GetNodeStatus400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "required")
			},
		},
		{
			name:    "job client error",
			request: gen.GetNodeStatusRequestObject{Hostname: "_any"},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNodeStatus(gomock.Any(), "_any").
					Return("", nil, assert.AnError)
			},
			validateFunc: func(resp gen.GetNodeStatusResponseObject) {
				_, ok := resp.(gen.GetNodeStatus500JSONResponse)
				s.True(ok)
			},
		},
		{
			name:    "broadcast all success",
			request: gen.GetNodeStatusRequestObject{Hostname: "_all"},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNodeStatusBroadcast(gomock.Any(), "_all").
					Return("550e8400-e29b-41d4-a716-446655440000", []*jobtypes.NodeStatusResponse{
						{Hostname: "server1", Uptime: time.Hour},
						{Hostname: "server2", Uptime: 2 * time.Hour},
					}, map[string]string{}, nil)
			},
			validateFunc: func(resp gen.GetNodeStatusResponseObject) {
				s.NotNil(resp)
			},
		},
		{
			name:    "broadcast all with errors",
			request: gen.GetNodeStatusRequestObject{Hostname: "_all"},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNodeStatusBroadcast(gomock.Any(), "_all").
					Return("550e8400-e29b-41d4-a716-446655440000", []*jobtypes.NodeStatusResponse{
						{Hostname: "server1", Uptime: time.Hour},
					}, map[string]string{
						"server2": "disk full",
					}, nil)
			},
			validateFunc: func(resp gen.GetNodeStatusResponseObject) {
				r, ok := resp.(gen.GetNodeStatus200JSONResponse)
				s.True(ok)
				s.Len(r.Results, 2)
				var foundError bool
				for _, res := range r.Results {
					if res.Error != nil {
						foundError = true
						s.Equal("server2", res.Hostname)
						s.Equal("disk full", *res.Error)
					}
				}
				s.True(foundError)
			},
		},
		{
			name:    "broadcast all error",
			request: gen.GetNodeStatusRequestObject{Hostname: "_all"},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNodeStatusBroadcast(gomock.Any(), "_all").
					Return("", nil, nil, assert.AnError)
			},
			validateFunc: func(resp gen.GetNodeStatusResponseObject) {
				_, ok := resp.(gen.GetNodeStatus500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()

			resp, err := s.handler.GetNodeStatus(s.ctx, tt.request)
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func (s *NodeStatusGetPublicTestSuite) TestGetNodeStatusHTTP() {
	tests := []struct {
		name         string
		path         string
		setupJobMock func() *jobmocks.MockJobClient
		wantCode     int
		wantBody     string
		wantContains []string
	}{
		{
			name: "when get Ok",
			path: "/node/server1",
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(s.mockCtrl)
				mock.EXPECT().
					QueryNodeStatus(gomock.Any(), "server1").
					Return("550e8400-e29b-41d4-a716-446655440000", &jobtypes.NodeStatusResponse{
						Hostname: "default-hostname",
						Uptime:   5 * time.Hour,
						OSInfo: &host.OSInfo{
							Distribution: "Ubuntu",
							Version:      "24.04",
						},
						LoadAverages: &load.AverageStats{
							Load1:  1,
							Load5:  0.5,
							Load15: 0.2,
						},
						MemoryStats: &mem.Stats{
							Total:  8388608,
							Free:   4194304,
							Cached: 2097152,
						},
						DiskUsage: []disk.UsageStats{
							{
								Name:  "/dev/disk1",
								Total: 500000000000,
								Used:  250000000000,
								Free:  250000000000,
							},
						},
					}, nil)
				return mock
			},
			wantCode: http.StatusOK,
			wantBody: `
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "results": [
    {
      "disks": [
        {
          "free": 250000000000,
          "name": "/dev/disk1",
          "total": 500000000000,
          "used": 250000000000
        }
      ],
      "hostname": "default-hostname",
      "load_average": {
        "1min": 1,
        "5min": 0.5,
        "15min": 0.2
      },
      "memory": {
        "free": 4194304,
        "total": 8388608,
        "used": 2097152
      },
      "os_info": {
        "distribution": "Ubuntu",
        "version": "24.04"
      },
      "uptime": "0 days, 5 hours, 0 minutes"
    }
  ]
}
`,
		},
		{
			name: "when job client errors",
			path: "/node/server1",
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(s.mockCtrl)
				mock.EXPECT().
					QueryNodeStatus(gomock.Any(), "server1").
					Return("", nil, assert.AnError)
				return mock
			},
			wantCode: http.StatusInternalServerError,
			wantBody: `{"error":"assert.AnError general error for testing"}`,
		},
		{
			name: "when broadcast all",
			path: "/node/_all",
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(s.mockCtrl)
				mock.EXPECT().
					QueryNodeStatusBroadcast(gomock.Any(), "_all").
					Return("550e8400-e29b-41d4-a716-446655440000", []*jobtypes.NodeStatusResponse{
						{
							Hostname: "server1",
							Uptime:   time.Hour,
						},
						{
							Hostname: "server2",
							Uptime:   2 * time.Hour,
						},
					}, map[string]string{}, nil)
				return mock
			},
			wantCode:     http.StatusOK,
			wantContains: []string{`"results"`, `"server1"`, `"server2"`},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			nodeHandler := apinode.New(s.logger, jobMock)
			strictHandler := gen.NewStrictHandler(nodeHandler, nil)

			a := api.New(s.appConfig, s.logger)
			gen.RegisterHandlers(a.Echo, strictHandler)

			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()

			a.Echo.ServeHTTP(rec, req)

			s.Equal(tc.wantCode, rec.Code)
			if tc.wantBody != "" {
				s.JSONEq(tc.wantBody, rec.Body.String())
			}
			for _, str := range tc.wantContains {
				s.Contains(rec.Body.String(), str)
			}
		})
	}
}

const rbacStatusTestSigningKey = "test-signing-key-for-rbac-integration"

func (s *NodeStatusGetPublicTestSuite) TestGetNodeStatusRBACHTTP() {
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
					rbacStatusTestSigningKey,
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
			name: "when valid token with node:read returns 200",
			setupAuth: func(req *http.Request) {
				token, err := tokenManager.Generate(
					rbacStatusTestSigningKey,
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
					QueryNodeStatus(gomock.Any(), "server1").
					Return(
						"550e8400-e29b-41d4-a716-446655440000",
						&jobtypes.NodeStatusResponse{
							Hostname: "default-hostname",
							Uptime:   5 * time.Hour,
							OSInfo: &host.OSInfo{
								Distribution: "Ubuntu",
								Version:      "24.04",
							},
							LoadAverages: &load.AverageStats{
								Load1:  1,
								Load5:  0.5,
								Load15: 0.2,
							},
							MemoryStats: &mem.Stats{
								Total:  8388608,
								Free:   4194304,
								Cached: 2097152,
							},
							DiskUsage: []disk.UsageStats{
								{
									Name:  "/dev/disk1",
									Total: 500000000000,
									Used:  250000000000,
									Free:  250000000000,
								},
							},
						},
						nil,
					)
				return mock
			},
			wantCode:     http.StatusOK,
			wantContains: []string{`"hostname":"default-hostname"`, `"job_id"`},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			appConfig := config.Config{
				API: config.API{
					Server: config.Server{
						Security: config.ServerSecurity{
							SigningKey: rbacStatusTestSigningKey,
						},
					},
				},
			}

			server := api.New(appConfig, s.logger)
			handlers := server.GetNodeHandler(jobMock)
			server.RegisterHandlers(handlers)

			req := httptest.NewRequest(http.MethodGet, "/node/server1", nil)
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

func TestNodeStatusGetPublicTestSuite(t *testing.T) {
	suite.Run(t, new(NodeStatusGetPublicTestSuite))
}

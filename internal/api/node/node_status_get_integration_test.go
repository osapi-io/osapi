// Copyright (c) 2024 John Dewey

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
	"github.com/retr0h/osapi/internal/api/node"
	nodeGen "github.com/retr0h/osapi/internal/api/node/gen"
	"github.com/retr0h/osapi/internal/authtoken"
	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/job"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
	"github.com/retr0h/osapi/internal/provider/system/disk"
	"github.com/retr0h/osapi/internal/provider/system/host"
	"github.com/retr0h/osapi/internal/provider/system/load"
	"github.com/retr0h/osapi/internal/provider/system/mem"
	"github.com/retr0h/osapi/internal/validation"
)

type NodeStatusGetIntegrationTestSuite struct {
	suite.Suite
	ctrl *gomock.Controller

	appConfig config.Config
	logger    *slog.Logger
}

func (suite *NodeStatusGetIntegrationTestSuite) SetupSuite() {
	validation.RegisterTargetValidator(func(_ context.Context) ([]validation.WorkerTarget, error) {
		return []validation.WorkerTarget{
			{Hostname: "server1", Labels: map[string]string{"group": "web"}},
			{Hostname: "server2"},
		}, nil
	})
}

func (suite *NodeStatusGetIntegrationTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())

	suite.appConfig = config.Config{}
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (suite *NodeStatusGetIntegrationTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *NodeStatusGetIntegrationTestSuite) TestGetNodeStatusValidation() {
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
			path: "/node/status",
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(suite.ctrl)
				mock.EXPECT().
					QueryNodeStatus(gomock.Any(), job.AnyHost).
					Return("550e8400-e29b-41d4-a716-446655440000", &job.SystemStatusResponse{
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
			path: "/node/status",
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(suite.ctrl)
				mock.EXPECT().
					QueryNodeStatus(gomock.Any(), job.AnyHost).
					Return("", nil, assert.AnError)
				return mock
			},
			wantCode: http.StatusInternalServerError,
			wantBody: `{"error":"assert.AnError general error for testing"}`,
		},
		{
			name: "when empty target_hostname returns 400",
			path: "/node/status?target_hostname=",
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(suite.ctrl)
			},
			wantCode:     http.StatusBadRequest,
			wantContains: []string{`"error"`},
		},
		{
			name: "when broadcast all",
			path: "/node/status?target_hostname=_all",
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(suite.ctrl)
				mock.EXPECT().
					QueryNodeStatusBroadcast(gomock.Any(), gomock.Any()).
					Return("550e8400-e29b-41d4-a716-446655440000", []*job.SystemStatusResponse{
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
		suite.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			nodeHandler := node.New(suite.logger, jobMock)
			strictHandler := nodeGen.NewStrictHandler(nodeHandler, nil)

			a := api.New(suite.appConfig, suite.logger)
			nodeGen.RegisterHandlers(a.Echo, strictHandler)

			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()

			a.Echo.ServeHTTP(rec, req)

			suite.Equal(tc.wantCode, rec.Code)
			if tc.wantBody != "" {
				suite.JSONEq(tc.wantBody, rec.Body.String())
			}
			for _, s := range tc.wantContains {
				suite.Contains(rec.Body.String(), s)
			}
		})
	}
}

const rbacStatusTestSigningKey = "test-signing-key-for-rbac-integration"

func (suite *NodeStatusGetIntegrationTestSuite) TestGetNodeStatusRBAC() {
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
					rbacStatusTestSigningKey,
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
			name: "when valid token with node:read returns 200",
			setupAuth: func(req *http.Request) {
				token, err := tokenManager.Generate(
					rbacStatusTestSigningKey,
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
					QueryNodeStatus(gomock.Any(), job.AnyHost).
					Return(
						"550e8400-e29b-41d4-a716-446655440000",
						&job.SystemStatusResponse{
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
		suite.Run(tc.name, func() {
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

			server := api.New(appConfig, suite.logger)
			handlers := server.GetNodeHandler(jobMock)
			server.RegisterHandlers(handlers)

			req := httptest.NewRequest(http.MethodGet, "/node/status", nil)
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

// In order for `go test` to run this suite, we need to create
// a normal test function and pass our suite to suite.Run.
func TestNodeStatusGetIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(NodeStatusGetIntegrationTestSuite))
}

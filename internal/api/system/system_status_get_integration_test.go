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

package system_test

import (
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
	"github.com/retr0h/osapi/internal/api/system"
	systemGen "github.com/retr0h/osapi/internal/api/system/gen"
	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/job"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
	"github.com/retr0h/osapi/internal/provider/system/disk"
	"github.com/retr0h/osapi/internal/provider/system/host"
	"github.com/retr0h/osapi/internal/provider/system/load"
	"github.com/retr0h/osapi/internal/provider/system/mem"
)

type SystemStatusGetIntegrationTestSuite struct {
	suite.Suite
	ctrl *gomock.Controller

	appConfig config.Config
	logger    *slog.Logger
}

func (suite *SystemStatusGetIntegrationTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())

	suite.appConfig = config.Config{}
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (suite *SystemStatusGetIntegrationTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *SystemStatusGetIntegrationTestSuite) TestGetSystemStatus() {
	tests := []struct {
		name         string
		path         string
		setupJobMock func() *jobmocks.MockJobClient
		wantCode     int
		wantBody     string
	}{
		{
			name: "when get Ok",
			path: "/system/status",
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(suite.ctrl)
				mock.EXPECT().
					QuerySystemStatus(gomock.Any(), job.AnyHost).
					Return(&job.SystemStatusResponse{
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
`,
		},
		{
			name: "when job client errors",
			path: "/system/status",
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(suite.ctrl)
				mock.EXPECT().
					QuerySystemStatus(gomock.Any(), job.AnyHost).
					Return(nil, assert.AnError)
				return mock
			},
			wantCode: http.StatusInternalServerError,
			wantBody: `{"error":"assert.AnError general error for testing"}`,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			systemHandler := system.New(jobMock)
			strictHandler := systemGen.NewStrictHandler(systemHandler, nil)

			a := api.New(suite.appConfig, suite.logger)
			systemGen.RegisterHandlers(a.Echo, strictHandler)

			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()

			a.Echo.ServeHTTP(rec, req)

			suite.Equal(tc.wantCode, rec.Code)
			suite.JSONEq(tc.wantBody, rec.Body.String())
		})
	}
}

// In order for `go test` to run this suite, we need to create
// a normal test function and pass our suite to suite.Run.
func TestSystemStatusGetIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(SystemStatusGetIntegrationTestSuite))
}

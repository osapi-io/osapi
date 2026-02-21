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
	jobGen "github.com/retr0h/osapi/internal/api/job/gen"
	"github.com/retr0h/osapi/internal/config"
	jobtypes "github.com/retr0h/osapi/internal/job"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
)

type JobWorkersGetIntegrationTestSuite struct {
	suite.Suite
	ctrl *gomock.Controller

	appConfig config.Config
	logger    *slog.Logger
}

func (suite *JobWorkersGetIntegrationTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())

	suite.appConfig = config.Config{}
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (suite *JobWorkersGetIntegrationTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *JobWorkersGetIntegrationTestSuite) TestGetJobWorkers() {
	tests := []struct {
		name         string
		setupJobMock func() *jobmocks.MockJobClient
		wantCode     int
		wantContains []string
	}{
		{
			name: "when workers exist returns worker list",
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(suite.ctrl)
				mock.EXPECT().
					ListWorkers(gomock.Any()).
					Return("", []jobtypes.WorkerInfo{
						{Hostname: "server1"},
						{Hostname: "server2"},
					}, nil)
				return mock
			},
			wantCode:     http.StatusOK,
			wantContains: []string{`"total":2`, `"server1"`, `"server2"`},
		},
		{
			name: "when no workers returns empty list",
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(suite.ctrl)
				mock.EXPECT().
					ListWorkers(gomock.Any()).
					Return("", []jobtypes.WorkerInfo{}, nil)
				return mock
			},
			wantCode:     http.StatusOK,
			wantContains: []string{`"total":0`},
		},
		{
			name: "when job client errors returns 500",
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(suite.ctrl)
				mock.EXPECT().
					ListWorkers(gomock.Any()).
					Return("", nil, assert.AnError)
				return mock
			},
			wantCode:     http.StatusInternalServerError,
			wantContains: []string{`"error"`},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			jobHandler := apijob.New(suite.logger, jobMock)
			strictHandler := jobGen.NewStrictHandler(jobHandler, nil)

			a := api.New(suite.appConfig, suite.logger)
			jobGen.RegisterHandlers(a.Echo, strictHandler)

			req := httptest.NewRequest(http.MethodGet, "/job/workers", nil)
			rec := httptest.NewRecorder()

			a.Echo.ServeHTTP(rec, req)

			suite.Equal(tc.wantCode, rec.Code)
			for _, s := range tc.wantContains {
				suite.Contains(rec.Body.String(), s)
			}
		})
	}
}

func TestJobWorkersGetIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(JobWorkersGetIntegrationTestSuite))
}

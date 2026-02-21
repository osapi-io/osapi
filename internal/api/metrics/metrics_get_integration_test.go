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

package metrics_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/api"
	"github.com/retr0h/osapi/internal/config"
)

type MetricsGetIntegrationTestSuite struct {
	suite.Suite

	appConfig config.Config
	logger    *slog.Logger
}

func (suite *MetricsGetIntegrationTestSuite) SetupTest() {
	suite.appConfig = config.Config{}
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (suite *MetricsGetIntegrationTestSuite) TestGetMetrics() {
	tests := []struct {
		name         string
		path         string
		wantCode     int
		wantContains string
	}{
		{
			name:         "when metrics endpoint is wired returns prometheus text",
			path:         "/metrics",
			wantCode:     http.StatusOK,
			wantContains: "test_metric 42",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			metricsHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(
					[]byte(
						"# HELP test_metric A test metric.\n# TYPE test_metric gauge\ntest_metric 42\n",
					),
				)
			})

			a := api.New(suite.appConfig, suite.logger)
			handlers := a.GetMetricsHandler(metricsHandler, tc.path)
			a.RegisterHandlers(handlers)

			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()

			a.Echo.ServeHTTP(rec, req)

			suite.Equal(tc.wantCode, rec.Code)
			suite.Contains(rec.Body.String(), tc.wantContains)
		})
	}
}

func TestMetricsGetIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(MetricsGetIntegrationTestSuite))
}

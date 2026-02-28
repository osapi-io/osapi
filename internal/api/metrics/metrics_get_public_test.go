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

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/api"
	"github.com/retr0h/osapi/internal/api/metrics"
	"github.com/retr0h/osapi/internal/config"
)

type MetricsGetPublicTestSuite struct {
	suite.Suite

	appConfig config.Config
	logger    *slog.Logger
}

func (s *MetricsGetPublicTestSuite) SetupTest() {
	s.appConfig = config.Config{}
	s.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (s *MetricsGetPublicTestSuite) TestRegisterHandler() {
	tests := []struct {
		name     string
		path     string
		validate func(e *echo.Echo)
	}{
		{
			name: "registers GET route at provided path",
			path: "/metrics",
			validate: func(e *echo.Echo) {
				routes := e.Routes()
				s.NotEmpty(routes)

				found := false
				for _, r := range routes {
					if r.Path == "/metrics" && r.Method == http.MethodGet {
						found = true
						break
					}
				}
				s.True(found, "expected GET /metrics route to be registered")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			m := metrics.New(handler, tt.path)
			register := m.RegisterHandler()

			e := echo.New()
			register(e)

			tt.validate(e)
		})
	}
}

func (s *MetricsGetPublicTestSuite) TestGetMetricsHTTP() {
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
		s.Run(tc.name, func() {
			metricsHandler := http.HandlerFunc(func(
				w http.ResponseWriter,
				_ *http.Request,
			) {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(
					[]byte(
						"# HELP test_metric A test metric.\n# TYPE test_metric gauge\ntest_metric 42\n",
					),
				)
			})

			a := api.New(s.appConfig, s.logger)
			handlers := a.GetMetricsHandler(metricsHandler, tc.path)
			a.RegisterHandlers(handlers)

			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()

			a.Echo.ServeHTTP(rec, req)

			s.Equal(tc.wantCode, rec.Code)
			s.Contains(rec.Body.String(), tc.wantContains)
		})
	}
}

func TestMetricsGetPublicTestSuite(t *testing.T) {
	suite.Run(t, new(MetricsGetPublicTestSuite))
}

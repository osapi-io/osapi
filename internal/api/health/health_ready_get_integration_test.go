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

package health_test

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/api"
	"github.com/retr0h/osapi/internal/api/health"
	healthGen "github.com/retr0h/osapi/internal/api/health/gen"
	"github.com/retr0h/osapi/internal/config"
)

type HealthReadyGetIntegrationTestSuite struct {
	suite.Suite

	appConfig config.Config
	logger    *slog.Logger
}

func (suite *HealthReadyGetIntegrationTestSuite) SetupTest() {
	suite.appConfig = config.Config{}
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (suite *HealthReadyGetIntegrationTestSuite) TestGetHealthReady() {
	tests := []struct {
		name         string
		checker      *health.NATSChecker
		wantCode     int
		wantContains []string
	}{
		{
			name: "when all checks pass returns ready",
			checker: &health.NATSChecker{
				NATSCheck: func() error { return nil },
				KVCheck:   func() error { return nil },
			},
			wantCode:     http.StatusOK,
			wantContains: []string{`"status":"ready"`},
		},
		{
			name: "when NATS check fails returns not ready",
			checker: &health.NATSChecker{
				NATSCheck: func() error { return fmt.Errorf("NATS not connected") },
				KVCheck:   func() error { return nil },
			},
			wantCode:     http.StatusServiceUnavailable,
			wantContains: []string{`"status":"not_ready"`, `"error"`},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			healthHandler := health.New(suite.logger, tc.checker, time.Now(), "0.1.0", nil)
			strictHandler := healthGen.NewStrictHandler(healthHandler, nil)

			a := api.New(suite.appConfig, suite.logger)
			healthGen.RegisterHandlers(a.Echo, strictHandler)

			req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
			rec := httptest.NewRecorder()

			a.Echo.ServeHTTP(rec, req)

			suite.Equal(tc.wantCode, rec.Code)
			for _, s := range tc.wantContains {
				suite.Contains(rec.Body.String(), s)
			}
		})
	}
}

func TestHealthReadyGetIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(HealthReadyGetIntegrationTestSuite))
}

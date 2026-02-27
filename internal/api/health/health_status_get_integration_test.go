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
	"context"
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
	"github.com/retr0h/osapi/internal/authtoken"
	"github.com/retr0h/osapi/internal/config"
)

type HealthStatusGetIntegrationTestSuite struct {
	suite.Suite

	appConfig config.Config
	logger    *slog.Logger
}

func (suite *HealthStatusGetIntegrationTestSuite) SetupTest() {
	suite.appConfig = config.Config{}
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (suite *HealthStatusGetIntegrationTestSuite) TestGetHealthStatusValidation() {
	tests := []struct {
		name         string
		checker      *health.NATSChecker
		metrics      health.MetricsProvider
		wantCode     int
		wantContains []string
	}{
		{
			name: "when all components healthy returns status with metrics",
			checker: &health.NATSChecker{
				NATSCheck: func() error { return nil },
				KVCheck:   func() error { return nil },
			},
			metrics: &health.ClosureMetricsProvider{
				NATSInfoFn: func(_ context.Context) (*health.NATSMetrics, error) {
					return &health.NATSMetrics{URL: "nats://localhost:4222", Version: "2.10.0"}, nil
				},
				StreamInfoFn: func(_ context.Context) ([]health.StreamMetrics, error) {
					return []health.StreamMetrics{
						{Name: "JOBS", Messages: 42, Bytes: 1024, Consumers: 1},
					}, nil
				},
				KVInfoFn: func(_ context.Context) ([]health.KVMetrics, error) {
					return []health.KVMetrics{
						{Name: "job-queue", Keys: 10, Bytes: 2048},
					}, nil
				},
				ConsumerStatsFn: func(_ context.Context) (*health.ConsumerMetrics, error) {
					return &health.ConsumerMetrics{Total: 2}, nil
				},
				JobStatsFn: func(_ context.Context) (*health.JobMetrics, error) {
					return &health.JobMetrics{
						Total: 100, Unprocessed: 5, Processing: 2,
						Completed: 90, Failed: 3, DLQ: 0,
					}, nil
				},
				AgentStatsFn: func(_ context.Context) (*health.AgentMetrics, error) {
					return &health.AgentMetrics{Total: 3, Ready: 3}, nil
				},
			},
			wantCode: http.StatusOK,
			wantContains: []string{
				`"status":"ok"`,
				`"version":"0.1.0"`,
				`"uptime"`,
				`"nats"`,
				`"streams"`,
				`"kv_buckets"`,
				`"consumers"`,
				`"jobs"`,
				`"agents"`,
			},
		},
		{
			name: "when nil metrics omits metrics fields",
			checker: &health.NATSChecker{
				NATSCheck: func() error { return nil },
				KVCheck:   func() error { return nil },
			},
			metrics:  nil,
			wantCode: http.StatusOK,
			wantContains: []string{
				`"status":"ok"`,
				`"version":"0.1.0"`,
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			healthHandler := health.New(
				suite.logger, tc.checker, time.Now(), "0.1.0", tc.metrics,
			)
			strictHandler := healthGen.NewStrictHandler(healthHandler, nil)

			a := api.New(suite.appConfig, suite.logger)
			healthGen.RegisterHandlers(a.Echo, strictHandler)

			req := httptest.NewRequest(http.MethodGet, "/health/status", nil)
			rec := httptest.NewRecorder()

			a.Echo.ServeHTTP(rec, req)

			suite.Equal(tc.wantCode, rec.Code)
			for _, s := range tc.wantContains {
				suite.Contains(rec.Body.String(), s)
			}
		})
	}
}

const rbacHealthStatusTestSigningKey = "test-signing-key-for-rbac-integration"

func (suite *HealthStatusGetIntegrationTestSuite) TestGetHealthStatusRBAC() {
	tokenManager := authtoken.New(suite.logger)

	tests := []struct {
		name         string
		setupAuth    func(req *http.Request)
		wantCode     int
		wantContains []string
	}{
		{
			name: "when no token returns 401",
			setupAuth: func(_ *http.Request) {
				// No auth header set
			},
			wantCode:     http.StatusUnauthorized,
			wantContains: []string{"Bearer token required"},
		},
		{
			name: "when insufficient permissions returns 403",
			setupAuth: func(req *http.Request) {
				token, err := tokenManager.Generate(
					rbacHealthStatusTestSigningKey,
					[]string{"read"},
					"test-user",
					[]string{"job:read"},
				)
				suite.Require().NoError(err)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			},
			wantCode:     http.StatusForbidden,
			wantContains: []string{"Insufficient permissions"},
		},
		{
			name: "when valid token with health:read returns 200",
			setupAuth: func(req *http.Request) {
				token, err := tokenManager.Generate(
					rbacHealthStatusTestSigningKey,
					[]string{"admin"},
					"test-user",
					nil,
				)
				suite.Require().NoError(err)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			},
			wantCode:     http.StatusOK,
			wantContains: []string{`"status":"ok"`, `"version"`},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			checker := &health.NATSChecker{
				NATSCheck: func() error { return nil },
				KVCheck:   func() error { return nil },
			}
			metrics := &health.ClosureMetricsProvider{
				NATSInfoFn: func(_ context.Context) (*health.NATSMetrics, error) {
					return &health.NATSMetrics{URL: "nats://localhost:4222", Version: "2.10.0"}, nil
				},
				StreamInfoFn: func(_ context.Context) ([]health.StreamMetrics, error) {
					return []health.StreamMetrics{}, nil
				},
				KVInfoFn: func(_ context.Context) ([]health.KVMetrics, error) {
					return []health.KVMetrics{}, nil
				},
				JobStatsFn: func(_ context.Context) (*health.JobMetrics, error) {
					return &health.JobMetrics{}, nil
				},
				ConsumerStatsFn: func(_ context.Context) (*health.ConsumerMetrics, error) {
					return &health.ConsumerMetrics{}, nil
				},
				AgentStatsFn: func(_ context.Context) (*health.AgentMetrics, error) {
					return &health.AgentMetrics{}, nil
				},
			}

			appConfig := config.Config{
				API: config.API{
					Server: config.Server{
						Security: config.ServerSecurity{
							SigningKey: rbacHealthStatusTestSigningKey,
						},
					},
				},
			}

			server := api.New(appConfig, suite.logger)
			handlers := server.GetHealthHandler(checker, time.Now(), "0.1.0", metrics)
			server.RegisterHandlers(handlers)

			req := httptest.NewRequest(http.MethodGet, "/health/status", nil)
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

func TestHealthStatusGetIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(HealthStatusGetIntegrationTestSuite))
}

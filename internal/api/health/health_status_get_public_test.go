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
	"github.com/retr0h/osapi/internal/api/health/gen"
	"github.com/retr0h/osapi/internal/authtoken"
	"github.com/retr0h/osapi/internal/config"
)

type stubChecker struct{}

func (s *stubChecker) CheckHealth(
	_ context.Context,
) error {
	return nil
}

const rbacHealthStatusTestSigningKey = "test-signing-key-for-rbac-integration"

type HealthStatusGetPublicTestSuite struct {
	suite.Suite

	ctx       context.Context
	appConfig config.Config
	logger    *slog.Logger
}

func (s *HealthStatusGetPublicTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.appConfig = config.Config{}
	s.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (s *HealthStatusGetPublicTestSuite) TestGetHealthStatus() {
	tests := []struct {
		name         string
		checker      health.Checker
		metrics      health.MetricsProvider
		validateFunc func(resp gen.GetHealthStatusResponseObject)
	}{
		{
			name: "all components healthy",
			checker: &health.NATSChecker{
				NATSCheck: func() error { return nil },
				KVCheck:   func() error { return nil },
			},
			validateFunc: func(resp gen.GetHealthStatusResponseObject) {
				r, ok := resp.(gen.GetHealthStatus200JSONResponse)
				s.True(ok)
				s.Equal("ok", r.Status)
				s.Equal("ok", r.Components["nats"].Status)
				s.Equal("ok", r.Components["kv"].Status)
				s.Equal("0.1.0", r.Version)
				s.NotEmpty(r.Uptime)
			},
		},
		{
			name: "NATS unhealthy returns 503",
			checker: &health.NATSChecker{
				NATSCheck: func() error { return fmt.Errorf("nats not connected") },
				KVCheck:   func() error { return nil },
			},
			validateFunc: func(resp gen.GetHealthStatusResponseObject) {
				r, ok := resp.(gen.GetHealthStatus503JSONResponse)
				s.True(ok)
				s.Equal("degraded", r.Status)
				s.Equal("error", r.Components["nats"].Status)
				s.Require().NotNil(r.Components["nats"].Error)
				s.Contains(*r.Components["nats"].Error, "nats not connected")
				s.Equal("ok", r.Components["kv"].Status)
			},
		},
		{
			name: "KV unhealthy returns 503",
			checker: &health.NATSChecker{
				NATSCheck: func() error { return nil },
				KVCheck:   func() error { return fmt.Errorf("kv bucket not accessible") },
			},
			validateFunc: func(resp gen.GetHealthStatusResponseObject) {
				r, ok := resp.(gen.GetHealthStatus503JSONResponse)
				s.True(ok)
				s.Equal("degraded", r.Status)
				s.Equal("ok", r.Components["nats"].Status)
				s.Equal("error", r.Components["kv"].Status)
				s.Require().NotNil(r.Components["kv"].Error)
				s.Contains(*r.Components["kv"].Error, "kv bucket not accessible")
			},
		},
		{
			name: "both unhealthy returns 503",
			checker: &health.NATSChecker{
				NATSCheck: func() error { return fmt.Errorf("nats down") },
				KVCheck:   func() error { return fmt.Errorf("kv down") },
			},
			validateFunc: func(resp gen.GetHealthStatusResponseObject) {
				r, ok := resp.(gen.GetHealthStatus503JSONResponse)
				s.True(ok)
				s.Equal("degraded", r.Status)
				s.Equal("error", r.Components["nats"].Status)
				s.Equal("error", r.Components["kv"].Status)
			},
		},
		{
			name: "includes version and uptime",
			checker: &health.NATSChecker{
				NATSCheck: func() error { return nil },
				KVCheck:   func() error { return nil },
			},
			validateFunc: func(resp gen.GetHealthStatusResponseObject) {
				r, ok := resp.(gen.GetHealthStatus200JSONResponse)
				s.True(ok)
				s.Equal("0.1.0", r.Version)
				s.NotEmpty(r.Uptime)
			},
		},
		{
			name:    "non-NATSChecker returns ok with nil components",
			checker: &stubChecker{},
			validateFunc: func(resp gen.GetHealthStatusResponseObject) {
				r, ok := resp.(gen.GetHealthStatus200JSONResponse)
				s.True(ok)
				s.Equal("ok", r.Status)
				s.Equal("ok", r.Components["nats"].Status)
				s.Equal("ok", r.Components["kv"].Status)
			},
		},
		{
			name: "nil MetricsProvider omits metrics",
			checker: &health.NATSChecker{
				NATSCheck: func() error { return nil },
				KVCheck:   func() error { return nil },
			},
			metrics: nil,
			validateFunc: func(resp gen.GetHealthStatusResponseObject) {
				r, ok := resp.(gen.GetHealthStatus200JSONResponse)
				s.True(ok)
				s.Nil(r.Nats)
				s.Nil(r.Streams)
				s.Nil(r.KvBuckets)
				s.Nil(r.Jobs)
			},
		},
		{
			name: "successful metrics populated",
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
					return &health.ConsumerMetrics{
						Total: 2,
						Consumers: []health.ConsumerDetail{
							{Name: "query_any_web_01", Pending: 0, AckPending: 3, Redelivered: 0},
							{Name: "modify_any_web_01", Pending: 1, AckPending: 0, Redelivered: 1},
						},
					}, nil
				},
				JobStatsFn: func(_ context.Context) (*health.JobMetrics, error) {
					return &health.JobMetrics{
						Total: 100, Unprocessed: 5, Processing: 2,
						Completed: 90, Failed: 3, DLQ: 0,
					}, nil
				},
				AgentStatsFn: func(_ context.Context) (*health.AgentMetrics, error) {
					return &health.AgentMetrics{
						Total: 3,
						Ready: 3,
						Agents: []health.AgentDetail{
							{Hostname: "web-01", Labels: "group=web.prod", Registered: "15s ago"},
							{Hostname: "web-02", Labels: "group=web.prod", Registered: "8s ago"},
							{Hostname: "db-01", Labels: "", Registered: "2m ago"},
						},
					}, nil
				},
			},
			validateFunc: func(resp gen.GetHealthStatusResponseObject) {
				r, ok := resp.(gen.GetHealthStatus200JSONResponse)
				s.True(ok)

				s.Require().NotNil(r.Nats)
				s.Equal("nats://localhost:4222", r.Nats.Url)
				s.Equal("2.10.0", r.Nats.Version)

				s.Require().NotNil(r.Streams)
				s.Len(*r.Streams, 1)
				s.Equal("JOBS", (*r.Streams)[0].Name)
				s.Equal(42, (*r.Streams)[0].Messages)

				s.Require().NotNil(r.KvBuckets)
				s.Len(*r.KvBuckets, 1)
				s.Equal("job-queue", (*r.KvBuckets)[0].Name)
				s.Equal(10, (*r.KvBuckets)[0].Keys)

				s.Require().NotNil(r.Consumers)
				s.Equal(2, r.Consumers.Total)
				s.Require().NotNil(r.Consumers.Consumers)
				s.Len(*r.Consumers.Consumers, 2)
				s.Equal("query_any_web_01", (*r.Consumers.Consumers)[0].Name)
				s.Equal(0, (*r.Consumers.Consumers)[0].Pending)
				s.Equal(3, (*r.Consumers.Consumers)[0].AckPending)
				s.Equal(1, (*r.Consumers.Consumers)[1].Pending)
				s.Equal(1, (*r.Consumers.Consumers)[1].Redelivered)

				s.Require().NotNil(r.Jobs)
				s.Equal(100, r.Jobs.Total)
				s.Equal(5, r.Jobs.Unprocessed)
				s.Equal(90, r.Jobs.Completed)

				s.Require().NotNil(r.Agents)
				s.Equal(3, r.Agents.Total)
				s.Equal(3, r.Agents.Ready)
				s.Require().NotNil(r.Agents.Agents)
				s.Len(*r.Agents.Agents, 3)
				s.Equal("web-01", (*r.Agents.Agents)[0].Hostname)
				s.Require().NotNil((*r.Agents.Agents)[0].Labels)
				s.Equal("group=web.prod", *(*r.Agents.Agents)[0].Labels)
				s.Equal("15s ago", (*r.Agents.Agents)[0].Registered)
				s.Equal("db-01", (*r.Agents.Agents)[2].Hostname)
				s.Nil((*r.Agents.Agents)[2].Labels)
				s.Equal("2m ago", (*r.Agents.Agents)[2].Registered)
			},
		},
		{
			name: "partial metric failures degrade gracefully",
			checker: &health.NATSChecker{
				NATSCheck: func() error { return nil },
				KVCheck:   func() error { return nil },
			},
			metrics: &health.ClosureMetricsProvider{
				NATSInfoFn: func(_ context.Context) (*health.NATSMetrics, error) {
					return nil, fmt.Errorf("nats info unavailable")
				},
				StreamInfoFn: func(_ context.Context) ([]health.StreamMetrics, error) {
					return nil, fmt.Errorf("stream info unavailable")
				},
				KVInfoFn: func(_ context.Context) ([]health.KVMetrics, error) {
					return []health.KVMetrics{
						{Name: "job-queue", Keys: 5, Bytes: 512},
					}, nil
				},
				ConsumerStatsFn: func(_ context.Context) (*health.ConsumerMetrics, error) {
					return nil, fmt.Errorf("consumer stats unavailable")
				},
				JobStatsFn: func(_ context.Context) (*health.JobMetrics, error) {
					return nil, fmt.Errorf("job stats unavailable")
				},
				AgentStatsFn: func(_ context.Context) (*health.AgentMetrics, error) {
					return nil, fmt.Errorf("agent stats unavailable")
				},
			},
			validateFunc: func(resp gen.GetHealthStatusResponseObject) {
				r, ok := resp.(gen.GetHealthStatus200JSONResponse)
				s.True(ok)
				s.Equal("ok", r.Status)
				s.Nil(r.Nats)
				s.Nil(r.Streams)
				s.Require().NotNil(r.KvBuckets)
				s.Len(*r.KvBuckets, 1)
				s.Nil(r.Consumers)
				s.Nil(r.Jobs)
				s.Nil(r.Agents)
			},
		},
		{
			name: "all metric failures degrade gracefully",
			checker: &health.NATSChecker{
				NATSCheck: func() error { return nil },
				KVCheck:   func() error { return nil },
			},
			metrics: &health.ClosureMetricsProvider{
				NATSInfoFn: func(_ context.Context) (*health.NATSMetrics, error) {
					return nil, fmt.Errorf("nats info unavailable")
				},
				StreamInfoFn: func(_ context.Context) ([]health.StreamMetrics, error) {
					return nil, fmt.Errorf("stream info unavailable")
				},
				KVInfoFn: func(_ context.Context) ([]health.KVMetrics, error) {
					return nil, fmt.Errorf("kv info unavailable")
				},
				ConsumerStatsFn: func(_ context.Context) (*health.ConsumerMetrics, error) {
					return nil, fmt.Errorf("consumer stats unavailable")
				},
				JobStatsFn: func(_ context.Context) (*health.JobMetrics, error) {
					return nil, fmt.Errorf("job stats unavailable")
				},
				AgentStatsFn: func(_ context.Context) (*health.AgentMetrics, error) {
					return nil, fmt.Errorf("agent stats unavailable")
				},
			},
			validateFunc: func(resp gen.GetHealthStatusResponseObject) {
				r, ok := resp.(gen.GetHealthStatus200JSONResponse)
				s.True(ok)
				s.Equal("ok", r.Status)
				s.Nil(r.Nats)
				s.Nil(r.Streams)
				s.Nil(r.KvBuckets)
				s.Nil(r.Consumers)
				s.Nil(r.Jobs)
				s.Nil(r.Agents)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			handler := health.New(slog.Default(), tt.checker, time.Now(), "0.1.0", tt.metrics)

			resp, err := handler.GetHealthStatus(s.ctx, gen.GetHealthStatusRequestObject{})
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func (s *HealthStatusGetPublicTestSuite) TestGetHealthStatusHTTP() {
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
				NATSInfoFn: func(
					_ context.Context,
				) (*health.NATSMetrics, error) {
					return &health.NATSMetrics{
						URL:     "nats://localhost:4222",
						Version: "2.10.0",
					}, nil
				},
				StreamInfoFn: func(
					_ context.Context,
				) ([]health.StreamMetrics, error) {
					return []health.StreamMetrics{
						{Name: "JOBS", Messages: 42, Bytes: 1024, Consumers: 1},
					}, nil
				},
				KVInfoFn: func(
					_ context.Context,
				) ([]health.KVMetrics, error) {
					return []health.KVMetrics{
						{Name: "job-queue", Keys: 10, Bytes: 2048},
					}, nil
				},
				ConsumerStatsFn: func(
					_ context.Context,
				) (*health.ConsumerMetrics, error) {
					return &health.ConsumerMetrics{
						Total: 2,
						Consumers: []health.ConsumerDetail{
							{Name: "query_any_web_01", Pending: 0, AckPending: 3, Redelivered: 0},
						},
					}, nil
				},
				JobStatsFn: func(
					_ context.Context,
				) (*health.JobMetrics, error) {
					return &health.JobMetrics{
						Total: 100, Unprocessed: 5, Processing: 2,
						Completed: 90, Failed: 3, DLQ: 0,
					}, nil
				},
				AgentStatsFn: func(
					_ context.Context,
				) (*health.AgentMetrics, error) {
					return &health.AgentMetrics{
						Total: 3,
						Ready: 3,
						Agents: []health.AgentDetail{
							{Hostname: "web-01", Labels: "group=web.prod", Registered: "15s ago"},
						},
					}, nil
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
				`"web-01"`,
				`"group=web.prod"`,
				`"query_any_web_01"`,
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
		s.Run(tc.name, func() {
			healthHandler := health.New(
				s.logger,
				tc.checker,
				time.Now(),
				"0.1.0",
				tc.metrics,
			)
			strictHandler := gen.NewStrictHandler(healthHandler, nil)

			a := api.New(s.appConfig, s.logger)
			gen.RegisterHandlers(a.Echo, strictHandler)

			req := httptest.NewRequest(http.MethodGet, "/health/status", nil)
			rec := httptest.NewRecorder()

			a.Echo.ServeHTTP(rec, req)

			s.Equal(tc.wantCode, rec.Code)
			for _, want := range tc.wantContains {
				s.Contains(rec.Body.String(), want)
			}
		})
	}
}

func (s *HealthStatusGetPublicTestSuite) TestGetHealthStatusRBACHTTP() {
	tokenManager := authtoken.New(s.logger)

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
				s.Require().NoError(err)
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
				s.Require().NoError(err)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			},
			wantCode:     http.StatusOK,
			wantContains: []string{`"status":"ok"`, `"version"`},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			checker := &health.NATSChecker{
				NATSCheck: func() error { return nil },
				KVCheck:   func() error { return nil },
			}
			metrics := &health.ClosureMetricsProvider{
				NATSInfoFn: func(
					_ context.Context,
				) (*health.NATSMetrics, error) {
					return &health.NATSMetrics{
						URL:     "nats://localhost:4222",
						Version: "2.10.0",
					}, nil
				},
				StreamInfoFn: func(
					_ context.Context,
				) ([]health.StreamMetrics, error) {
					return []health.StreamMetrics{}, nil
				},
				KVInfoFn: func(
					_ context.Context,
				) ([]health.KVMetrics, error) {
					return []health.KVMetrics{}, nil
				},
				JobStatsFn: func(
					_ context.Context,
				) (*health.JobMetrics, error) {
					return &health.JobMetrics{}, nil
				},
				ConsumerStatsFn: func(
					_ context.Context,
				) (*health.ConsumerMetrics, error) {
					return &health.ConsumerMetrics{}, nil
				},
				AgentStatsFn: func(
					_ context.Context,
				) (*health.AgentMetrics, error) {
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

			server := api.New(appConfig, s.logger)
			handlers := server.GetHealthHandler(
				checker,
				time.Now(),
				"0.1.0",
				metrics,
			)
			server.RegisterHandlers(handlers)

			req := httptest.NewRequest(http.MethodGet, "/health/status", nil)
			tc.setupAuth(req)
			rec := httptest.NewRecorder()

			server.Echo.ServeHTTP(rec, req)

			s.Equal(tc.wantCode, rec.Code)
			for _, want := range tc.wantContains {
				s.Contains(rec.Body.String(), want)
			}
		})
	}
}

func TestHealthStatusGetPublicTestSuite(t *testing.T) {
	suite.Run(t, new(HealthStatusGetPublicTestSuite))
}

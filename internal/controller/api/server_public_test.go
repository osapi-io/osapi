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

package api_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/suite"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"

	auditmocks "github.com/osapi-io/osapi/internal/audit/mocks"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller/api"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

type ServerPublicTestSuite struct {
	suite.Suite

	mockCtrl *gomock.Controller
}

func (s *ServerPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
}

func (s *ServerPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *ServerPublicTestSuite) TestNew() {
	tests := []struct {
		name         string
		appConfig    config.Config
		opts         []api.Option
		validateFunc func(*api.Server)
	}{
		{
			name: "creates server with default config",
			appConfig: config.Config{
				Controller: config.Controller{
					API: config.APIServer{
						Security: config.ServerSecurity{
							SigningKeys: []config.SigningKey{{ID: "test", Key: "test-key"}},
						},
					},
				},
			},
			validateFunc: func(server *api.Server) {
				s.NotNil(server)
				s.NotNil(server.Echo)
			},
		},
		{
			name: "creates server with custom roles",
			appConfig: config.Config{
				Controller: config.Controller{
					API: config.APIServer{
						Security: config.ServerSecurity{
							SigningKeys: []config.SigningKey{{ID: "test", Key: "test-key"}},
							Roles: map[types.Role]config.CustomRole{
								"ops": {
									Permissions: []types.Permission{"status:read", "health:read"},
								},
							},
						},
					},
				},
			},
			validateFunc: func(server *api.Server) {
				s.NotNil(server)
				s.NotNil(server.Echo)
			},
		},
		{
			name: "creates server with audit store option",
			appConfig: config.Config{
				Controller: config.Controller{
					API: config.APIServer{
						Security: config.ServerSecurity{
							SigningKeys: []config.SigningKey{{ID: "test", Key: "test-key"}},
						},
					},
				},
			},
			opts: []api.Option{
				api.WithAuditStore(auditmocks.NewMockStore(s.mockCtrl)),
			},
			validateFunc: func(server *api.Server) {
				s.NotNil(server)
				s.NotNil(server.Echo)
			},
		},
		{
			name: "creates server with meter provider option",
			appConfig: config.Config{
				Controller: config.Controller{
					API: config.APIServer{
						Security: config.ServerSecurity{
							SigningKeys: []config.SigningKey{{ID: "test", Key: "test-key"}},
						},
					},
				},
			},
			opts: []api.Option{
				api.WithMeterProvider(sdkmetric.NewMeterProvider()),
			},
			validateFunc: func(server *api.Server) {
				s.NotNil(server)
				s.NotNil(server.Echo)
			},
		},
		{
			name: "creates server with CORS origins",
			appConfig: config.Config{
				Controller: config.Controller{
					API: config.APIServer{
						Security: config.ServerSecurity{
							SigningKeys: []config.SigningKey{{ID: "test", Key: "test-key"}},
							CORS: config.CORS{
								AllowOrigins: []string{
									"http://localhost:3000",
									"https://example.com",
								},
							},
						},
					},
				},
			},
			validateFunc: func(server *api.Server) {
				s.NotNil(server)
				s.Equal(
					"http://localhost:3000",
					s.corsHeaderFor(server),
					"a configured origin is answered",
				)
			},
		},
		{
			// The v4 default was "*", which this package no longer inherits:
			// v5 panics on an empty AllowOrigins and nothing here panics, so
			// the middleware is left off and the browser applies same-origin
			// policy. Statement coverage cannot see this, because the branch
			// that is skipped has no statements.
			name: "sends no CORS header when no origins are configured",
			appConfig: config.Config{
				Controller: config.Controller{
					API: config.APIServer{
						Security: config.ServerSecurity{
							SigningKeys: []config.SigningKey{{ID: "test", Key: "test-key"}},
						},
					},
				},
			},
			validateFunc: func(server *api.Server) {
				s.NotNil(server)
				s.Empty(
					s.corsHeaderFor(server),
					"an unconfigured server allows no origin",
				)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.validateFunc(api.New(tt.appConfig, slog.Default(), tt.opts...))
		})
	}
}

// corsHeaderFor issues a cross-origin GET through the server and returns the
// Access-Control-Allow-Origin header it answered with, empty when absent.
func (s *ServerPublicTestSuite) corsHeaderFor(
	server *api.Server,
) string {
	s.T().Helper()

	server.Echo.GET("/cors-probe", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/cors-probe", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	rec := httptest.NewRecorder()
	server.Echo.ServeHTTP(rec, req)

	return rec.Header().Get("Access-Control-Allow-Origin")
}

func (s *ServerPublicTestSuite) TestStartAndStop() {
	tests := []struct {
		name  string
		start bool
	}{
		{
			name:  "starts and stops gracefully",
			start: true,
		},
		{
			// Stop is reachable without Start when startup fails earlier in
			// the boot sequence. v5 cancels a context to shut down, so there
			// is nothing to cancel here and Stop must not panic.
			name:  "stops without having started",
			start: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			appConfig := config.Config{
				Controller: config.Controller{
					API: config.APIServer{
						Port: 0,
						Security: config.ServerSecurity{
							SigningKeys: []config.SigningKey{{ID: "test", Key: "test-key"}},
						},
					},
				},
			}

			server := api.New(appConfig, slog.Default())

			if tt.start {
				s.Require().NoError(server.Start())
				time.Sleep(50 * time.Millisecond)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			s.NotPanics(func() { server.Stop(ctx) })

			// Stopping twice is idempotent: the second call takes the
			// nil-stop path the first one left behind.
			s.NotPanics(func() { server.Stop(ctx) })
		})
	}
}

func (s *ServerPublicTestSuite) TestStartErrorPath() {
	tests := []struct {
		name string
	}{
		{
			name: "a port in use is the error of Start",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ln, err := net.Listen("tcp", ":0")
			s.Require().NoError(err)
			defer func() { _ = ln.Close() }()

			port := ln.Addr().(*net.TCPAddr).Port

			appConfig := config.Config{
				Controller: config.Controller{
					API: config.APIServer{
						Port: port,
						Security: config.ServerSecurity{
							SigningKeys: []config.SigningKey{{ID: "test", Key: "test-key"}},
						},
					},
				},
			}

			server := api.New(appConfig, slog.Default())
			s.ErrorContains(server.Start(), "listen on ")
		})
	}
}

func (s *ServerPublicTestSuite) TestStopErrorPath() {
	tests := []struct {
		name string
	}{
		{
			name: "logs error when shutdown context expired with active connections",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ln, err := net.Listen("tcp", ":0")
			s.Require().NoError(err)
			port := ln.Addr().(*net.TCPAddr).Port
			_ = ln.Close()

			appConfig := config.Config{
				Controller: config.Controller{
					API: config.APIServer{
						Port: port,
						Security: config.ServerSecurity{
							SigningKeys: []config.SigningKey{{ID: "test", Key: "test-key"}},
						},
					},
				},
			}

			server := api.New(appConfig, slog.Default())

			server.Echo.GET("/slow", func(c *echo.Context) error {
				time.Sleep(10 * time.Second)
				return c.String(http.StatusOK, "done")
			})

			s.Require().NoError(server.Start())
			time.Sleep(50 * time.Millisecond)

			go http.Get(fmt.Sprintf("http://localhost:%d/slow", port)) //nolint:errcheck
			time.Sleep(50 * time.Millisecond)

			ctx, cancel := context.WithDeadline(
				context.Background(),
				time.Now().Add(-time.Second),
			)
			defer cancel()

			server.Stop(ctx)
		})
	}
}

func (s *ServerPublicTestSuite) TestUntraced() {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{name: "when a job request is traced", path: "/api/node/web-01/package", want: false},
		{name: "when the agent list is traced", path: "/api/agent", want: false},
		{name: "when the health status poll is skipped", path: "/api/health/status", want: true},
		{name: "when the liveness probe is skipped", path: "/api/health", want: true},
		{name: "when a file of the UI is skipped", path: "/assets/index.js", want: true},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)

			s.Equal(tc.want, api.ExportUntraced(e.NewContext(req, httptest.NewRecorder())))
		})
	}
}

func (s *ServerPublicTestSuite) TestTraceResponse() {
	traced := trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(
		trace.SpanContextConfig{
			TraceID: trace.TraceID{
				0x26,
				0xe4,
				0x48,
				0xc9,
				1,
				2,
				3,
				4,
				5,
				6,
				7,
				8,
				9,
				10,
				11,
				12,
			},
			SpanID:     trace.SpanID{0x1a, 0x2b, 3, 4, 5, 6, 7, 8},
			TraceFlags: trace.FlagsSampled,
		},
	))
	tests := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{
			name: "when the request has a span returns its trace",
			ctx:  traced,
			want: "00-26e448c90102030405060708090a0b0c-1a2b030405060708-01",
		},
		{
			name: "when the request has no span returns no header",
			ctx:  context.Background(),
			want: "",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/node/web-01", nil).WithContext(tc.ctx)
			rec := httptest.NewRecorder()
			called := false

			err := api.ExportTraceResponse(func(_ *echo.Context) error {
				called = true

				return nil
			})(e.NewContext(req, rec))

			s.Require().NoError(err)
			s.True(called)
			s.Equal(tc.want, rec.Header().Get("traceresponse"))
		})
	}
}

func TestServerPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ServerPublicTestSuite))
}

// lockedBuffer is a bytes.Buffer that the goroutine of the server and the
// test can use at the same time.
type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(
	p []byte,
) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.String()
}

func (s *ServerPublicTestSuite) TestStartLogsThroughLogger() {
	tests := []struct {
		name string
		want []string
	}{
		{
			name: "writes the server started line through the logger of the server",
			want: []string{
				"msg=\"http(s) server started\"",
				"subsystem=controller.api",
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			out := &lockedBuffer{}
			logger := slog.New(slog.NewTextHandler(out, nil))

			appConfig := config.Config{
				Controller: config.Controller{
					API: config.APIServer{
						Host: "127.0.0.1",
						Port: 0,
						Security: config.ServerSecurity{
							SigningKeys: []config.SigningKey{{ID: "test", Key: "test-key"}},
						},
					},
				},
			}

			server := api.New(appConfig, logger)
			s.Require().NoError(server.Start())

			s.Eventually(func() bool {
				return strings.Contains(out.String(), "http(s) server started")
			}, 2*time.Second, 10*time.Millisecond)

			for _, w := range tt.want {
				s.Contains(out.String(), w)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			server.Stop(ctx)
		})
	}
}

func (s *ServerPublicTestSuite) TestStartServes() {
	appConfig := config.Config{
		Controller: config.Controller{
			API: config.APIServer{
				Security: config.ServerSecurity{
					SigningKeys: []config.SigningKey{{ID: "test", Key: "test-key"}},
				},
			},
		},
	}
	tests := []struct {
		name    string
		opts    []api.Option
		closeLn bool
		wantLog string
	}{
		{
			name: "with TLS, it serves through a TLS listener",
			opts: []api.Option{api.WithTLS(&tls.Config{MinVersion: tls.VersionTLS12})},
		},
		{
			name:    "a listener that fails while it serves is logged",
			closeLn: true,
			wantLog: "server stopped with an error",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			var ln net.Listener
			api.SetListenFn(func(ctx context.Context, network, _ string) (net.Listener, error) {
				l, err := (&net.ListenConfig{}).Listen(ctx, network, "127.0.0.1:0")
				ln = l

				return l, err
			})
			defer api.ResetListenFn()

			var buf lockedBuffer
			log := slog.New(slog.NewTextHandler(&buf, nil))
			server := api.New(appConfig, log, tt.opts...)
			s.Require().NoError(server.Start())
			if tt.closeLn {
				_ = ln.Close()
			}
			s.Eventually(func() bool {
				return strings.Contains(buf.String(), "starting server")
			}, 2*time.Second, 10*time.Millisecond)
			if tt.wantLog != "" {
				s.Eventually(func() bool {
					return strings.Contains(buf.String(), tt.wantLog)
				}, 2*time.Second, 10*time.Millisecond)
			}
			server.Stop(context.Background())
		})
	}
}

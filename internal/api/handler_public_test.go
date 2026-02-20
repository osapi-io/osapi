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
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/api"
	"github.com/retr0h/osapi/internal/api/health"
	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/job/mocks"
)

type HandlerPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *mocks.MockJobClient
	server        *api.Server
}

func (s *HandlerPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = mocks.NewMockJobClient(s.mockCtrl)

	appConfig := config.Config{
		API: config.API{
			Server: config.Server{
				Security: config.ServerSecurity{
					SigningKey: "test-signing-key",
				},
			},
		},
	}

	s.server = api.New(appConfig, slog.Default())
}

func (s *HandlerPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *HandlerPublicTestSuite) TestGetSystemHandler() {
	tests := []struct {
		name     string
		validate func([]func(e *echo.Echo))
	}{
		{
			name: "returns system handler functions",
			validate: func(handlers []func(e *echo.Echo)) {
				s.NotEmpty(handlers)
			},
		},
		{
			name: "closure registers routes and middleware executes",
			validate: func(handlers []func(e *echo.Echo)) {
				e := echo.New()
				for _, h := range handlers {
					h(e)
				}
				s.NotEmpty(e.Routes())

				req := httptest.NewRequest(http.MethodGet, "/system/hostname", nil)
				rec := httptest.NewRecorder()
				e.ServeHTTP(rec, req)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			handlers := s.server.GetSystemHandler(s.mockJobClient)

			tt.validate(handlers)
		})
	}
}

func (s *HandlerPublicTestSuite) TestGetNetworkHandler() {
	tests := []struct {
		name     string
		validate func([]func(e *echo.Echo))
	}{
		{
			name: "returns network handler functions",
			validate: func(handlers []func(e *echo.Echo)) {
				s.NotEmpty(handlers)
			},
		},
		{
			name: "closure registers routes and middleware executes",
			validate: func(handlers []func(e *echo.Echo)) {
				e := echo.New()
				for _, h := range handlers {
					h(e)
				}
				s.NotEmpty(e.Routes())

				req := httptest.NewRequest(http.MethodGet, "/network/dns/eth0", nil)
				rec := httptest.NewRecorder()
				e.ServeHTTP(rec, req)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			handlers := s.server.GetNetworkHandler(s.mockJobClient)

			tt.validate(handlers)
		})
	}
}

func (s *HandlerPublicTestSuite) TestGetJobHandler() {
	tests := []struct {
		name     string
		validate func([]func(e *echo.Echo))
	}{
		{
			name: "returns job handler functions",
			validate: func(handlers []func(e *echo.Echo)) {
				s.NotEmpty(handlers)
			},
		},
		{
			name: "closure registers routes and middleware executes",
			validate: func(handlers []func(e *echo.Echo)) {
				e := echo.New()
				for _, h := range handlers {
					h(e)
				}
				s.NotEmpty(e.Routes())

				req := httptest.NewRequest(http.MethodGet, "/job", nil)
				rec := httptest.NewRecorder()
				e.ServeHTTP(rec, req)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			handlers := s.server.GetJobHandler(s.mockJobClient)

			tt.validate(handlers)
		})
	}
}

func (s *HandlerPublicTestSuite) TestGetHealthHandler() {
	tests := []struct {
		name     string
		validate func([]func(e *echo.Echo))
	}{
		{
			name: "returns health handler functions",
			validate: func(handlers []func(e *echo.Echo)) {
				s.NotEmpty(handlers)
			},
		},
		{
			name: "closure registers routes and middleware executes for unauthenticated",
			validate: func(handlers []func(e *echo.Echo)) {
				e := echo.New()
				for _, h := range handlers {
					h(e)
				}
				s.NotEmpty(e.Routes())

				req := httptest.NewRequest(http.MethodGet, "/health", nil)
				rec := httptest.NewRecorder()
				e.ServeHTTP(rec, req)
			},
		},
		{
			name: "closure registers routes and middleware executes for authenticated",
			validate: func(handlers []func(e *echo.Echo)) {
				e := echo.New()
				for _, h := range handlers {
					h(e)
				}
				s.NotEmpty(e.Routes())

				req := httptest.NewRequest(http.MethodGet, "/health/status", nil)
				rec := httptest.NewRecorder()
				e.ServeHTTP(rec, req)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			appConfig := config.Config{
				API: config.API{
					Server: config.Server{
						Security: config.ServerSecurity{
							SigningKey: "test-signing-key",
						},
					},
				},
			}

			checker := &health.NATSChecker{}
			healthHandler := health.New(slog.Default(), checker, time.Now(), "0.1.0", nil)
			serverWithHealth := api.New(
				appConfig,
				slog.Default(),
				api.WithHealthHandler(healthHandler),
			)
			handlers := serverWithHealth.GetHealthHandler()

			tt.validate(handlers)
		})
	}
}

func (s *HandlerPublicTestSuite) TestCreateHandlers() {
	tests := []struct {
		name        string
		withHealth  bool
		expectedLen int
	}{
		{
			name:        "returns handler functions without health",
			withHealth:  false,
			expectedLen: 3,
		},
		{
			name:        "returns handler functions with health",
			withHealth:  true,
			expectedLen: 4,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			appConfig := config.Config{
				API: config.API{
					Server: config.Server{
						Security: config.ServerSecurity{
							SigningKey: "test-signing-key",
						},
					},
				},
			}

			var opts []api.Option
			if tt.withHealth {
				checker := &health.NATSChecker{}
				healthHandler := health.New(slog.Default(), checker, time.Now(), "0.1.0", nil)
				opts = append(opts, api.WithHealthHandler(healthHandler))
			}

			server := api.New(appConfig, slog.Default(), opts...)
			handlers := server.CreateHandlers(s.mockJobClient)

			s.Len(handlers, tt.expectedLen)
		})
	}
}

func (s *HandlerPublicTestSuite) TestRegisterHandlers() {
	tests := []struct {
		name string
	}{
		{
			name: "registers handlers with Echo",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			handlers := s.server.CreateHandlers(s.mockJobClient)

			routesBefore := len(s.server.Echo.Routes())
			s.server.RegisterHandlers(handlers)
			routesAfter := len(s.server.Echo.Routes())

			s.Greater(routesAfter, routesBefore)
		})
	}
}

func TestHandlerPublicTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerPublicTestSuite))
}

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
			checker := &health.NATSChecker{}
			handlers := s.server.GetHealthHandler(checker, time.Now(), "0.1.0", nil)

			tt.validate(handlers)
		})
	}
}

func (s *HandlerPublicTestSuite) TestGetMetricsHandler() {
	tests := []struct {
		name     string
		validate func([]func(e *echo.Echo))
	}{
		{
			name: "returns metrics handler functions",
			validate: func(handlers []func(e *echo.Echo)) {
				s.NotEmpty(handlers)
			},
		},
		{
			name: "closure registers route at provided path",
			validate: func(handlers []func(e *echo.Echo)) {
				e := echo.New()
				for _, h := range handlers {
					h(e)
				}
				s.NotEmpty(e.Routes())

				req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
				rec := httptest.NewRecorder()
				e.ServeHTTP(rec, req)

				s.Equal(http.StatusOK, rec.Code)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			metricsHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			handlers := s.server.GetMetricsHandler(metricsHandler, "/metrics")

			tt.validate(handlers)
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
			checker := &health.NATSChecker{}

			handlers := make([]func(e *echo.Echo), 0, 4)
			handlers = append(handlers, s.server.GetSystemHandler(s.mockJobClient)...)
			handlers = append(handlers, s.server.GetNetworkHandler(s.mockJobClient)...)
			handlers = append(handlers, s.server.GetJobHandler(s.mockJobClient)...)
			handlers = append(
				handlers,
				s.server.GetHealthHandler(checker, time.Now(), "0.1.0", nil)...)

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

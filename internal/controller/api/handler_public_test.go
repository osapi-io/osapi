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

	auditmocks "github.com/retr0h/osapi/internal/audit/mocks"
	"github.com/retr0h/osapi/internal/config"
	"github.com/retr0h/osapi/internal/controller/api"
	fileMocks "github.com/retr0h/osapi/internal/controller/api/file/mocks"
	"github.com/retr0h/osapi/internal/controller/api/health"
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
		Controller: config.Controller{
			API: config.APIServer{
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

func (s *HandlerPublicTestSuite) TestGetAgentHandler() {
	tests := []struct {
		name     string
		validate func([]func(e *echo.Echo))
	}{
		{
			name: "returns handler functions",
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

				req := httptest.NewRequest(http.MethodGet, "/agent", nil)
				rec := httptest.NewRecorder()
				e.ServeHTTP(rec, req)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			handlers := s.server.GetAgentHandler(s.mockJobClient)

			tt.validate(handlers)
		})
	}
}

func (s *HandlerPublicTestSuite) TestGetNodeHandler() {
	tests := []struct {
		name     string
		validate func([]func(e *echo.Echo))
	}{
		{
			name: "returns handler functions",
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

				req := httptest.NewRequest(http.MethodGet, "/node/hostname", nil)
				rec := httptest.NewRecorder()
				e.ServeHTTP(rec, req)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			handlers := s.server.GetNodeHandler(s.mockJobClient)

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
			handlers := s.server.GetHealthHandler(checker, time.Now(), "0.1.0", nil, nil)

			tt.validate(handlers)
		})
	}
}

func (s *HandlerPublicTestSuite) TestGetAuditHandler() {
	tests := []struct {
		name     string
		validate func([]func(e *echo.Echo))
	}{
		{
			name: "returns audit handler functions",
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

				req := httptest.NewRequest(http.MethodGet, "/audit", nil)
				rec := httptest.NewRecorder()
				e.ServeHTTP(rec, req)
			},
		},
		{
			name: "closure registers export route and middleware executes",
			validate: func(handlers []func(e *echo.Echo)) {
				e := echo.New()
				for _, h := range handlers {
					h(e)
				}
				s.NotEmpty(e.Routes())

				req := httptest.NewRequest(http.MethodGet, "/audit/export", nil)
				rec := httptest.NewRecorder()
				e.ServeHTTP(rec, req)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctrl := gomock.NewController(s.T())
			defer ctrl.Finish()

			store := auditmocks.NewMockStore(ctrl)
			handlers := s.server.GetAuditHandler(store)

			tt.validate(handlers)
		})
	}
}

func (s *HandlerPublicTestSuite) TestGetFileHandler() {
	tests := []struct {
		name     string
		validate func([]func(e *echo.Echo))
	}{
		{
			name: "returns handler functions",
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

				req := httptest.NewRequest(http.MethodGet, "/file", nil)
				rec := httptest.NewRecorder()
				e.ServeHTTP(rec, req)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctrl := gomock.NewController(s.T())
			defer ctrl.Finish()
			mockObjStore := fileMocks.NewMockObjectStoreManager(ctrl)

			handlers := s.server.GetFileHandler(mockObjStore)

			tt.validate(handlers)
		})
	}
}

func (s *HandlerPublicTestSuite) TestGetDockerHandler() {
	tests := []struct {
		name     string
		validate func([]func(e *echo.Echo))
	}{
		{
			name: "returns docker handler functions",
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

				req := httptest.NewRequest(http.MethodGet, "/node/hostname/container/docker", nil)
				rec := httptest.NewRecorder()
				e.ServeHTTP(rec, req)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			handlers := s.server.GetDockerHandler(s.mockJobClient)

			tt.validate(handlers)
		})
	}
}

func (s *HandlerPublicTestSuite) TestGetNodeScheduleHandler() {
	tests := []struct {
		name     string
		validate func([]func(e *echo.Echo))
	}{
		{
			name: "returns schedule handler functions",
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

				req := httptest.NewRequest(http.MethodGet, "/node/hostname/schedule/cron", nil)
				rec := httptest.NewRecorder()
				e.ServeHTTP(rec, req)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			handlers := s.server.GetNodeScheduleHandler(s.mockJobClient)

			tt.validate(handlers)
		})
	}
}

func (s *HandlerPublicTestSuite) TestGetNodeSysctlHandler() {
	tests := []struct {
		name     string
		validate func([]func(e *echo.Echo))
	}{
		{
			name: "returns sysctl handler functions",
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

				req := httptest.NewRequest(http.MethodGet, "/node/hostname/sysctl", nil)
				rec := httptest.NewRecorder()
				e.ServeHTTP(rec, req)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			handlers := s.server.GetNodeSysctlHandler(s.mockJobClient)

			tt.validate(handlers)
		})
	}
}

func (s *HandlerPublicTestSuite) TestGetFactsHandler() {
	tests := []struct {
		name     string
		validate func([]func(e *echo.Echo))
	}{
		{
			name: "returns facts handler functions",
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

				req := httptest.NewRequest(http.MethodGet, "/facts/keys", nil)
				rec := httptest.NewRecorder()
				e.ServeHTTP(rec, req)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			handlers := s.server.GetFactsHandler()

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

			handlers := make([]func(e *echo.Echo), 0, 5)
			handlers = append(handlers, s.server.GetAgentHandler(s.mockJobClient)...)
			handlers = append(handlers, s.server.GetNodeHandler(s.mockJobClient)...)
			handlers = append(handlers, s.server.GetJobHandler(s.mockJobClient)...)
			handlers = append(
				handlers,
				s.server.GetHealthHandler(checker, time.Now(), "0.1.0", nil, nil)...)

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

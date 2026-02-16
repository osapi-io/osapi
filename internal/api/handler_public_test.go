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
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/api"
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
		name string
	}{
		{
			name: "returns system handler functions",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			handlers := s.server.GetSystemHandler(s.mockJobClient)

			s.NotEmpty(handlers)
		})
	}
}

func (s *HandlerPublicTestSuite) TestGetNetworkHandler() {
	tests := []struct {
		name string
	}{
		{
			name: "returns network handler functions",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			handlers := s.server.GetNetworkHandler(s.mockJobClient)

			s.NotEmpty(handlers)
		})
	}
}

func (s *HandlerPublicTestSuite) TestGetJobHandler() {
	tests := []struct {
		name string
	}{
		{
			name: "returns job handler functions",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			handlers := s.server.GetJobHandler(s.mockJobClient)

			s.NotEmpty(handlers)
		})
	}
}

func (s *HandlerPublicTestSuite) TestCreateHandlers() {
	tests := []struct {
		name string
	}{
		{
			name: "returns all handler functions",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			handlers := s.server.CreateHandlers(s.mockJobClient)

			s.Len(handlers, 3)
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

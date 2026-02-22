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

package network_test

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/api"
	apinetwork "github.com/retr0h/osapi/internal/api/network"
	networkGen "github.com/retr0h/osapi/internal/api/network/gen"
	"github.com/retr0h/osapi/internal/authtoken"
	"github.com/retr0h/osapi/internal/config"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
)

type NetworkDNSPutByInterfaceIntegrationTestSuite struct {
	suite.Suite
	ctrl *gomock.Controller

	appConfig config.Config
	logger    *slog.Logger
}

func (suite *NetworkDNSPutByInterfaceIntegrationTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())

	suite.appConfig = config.Config{}
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (suite *NetworkDNSPutByInterfaceIntegrationTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *NetworkDNSPutByInterfaceIntegrationTestSuite) TestPutNetworkDNSValidation() {
	tests := []struct {
		name         string
		path         string
		body         string
		setupJobMock func() *jobmocks.MockJobClient
		wantCode     int
		wantContains []string
	}{
		{
			name: "when valid request",
			path: "/network/dns",
			body: `{"servers":["1.1.1.1","8.8.8.8"],"search_domains":["foo.bar"],"interface_name":"eth0"}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(suite.ctrl)
				mock.EXPECT().
					ModifyNetworkDNS(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("550e8400-e29b-41d4-a716-446655440000", "worker1", nil)
				return mock
			},
			wantCode:     http.StatusAccepted,
			wantContains: []string{`"results"`, `"worker1"`, `"ok"`},
		},
		{
			name: "when missing interface name",
			path: "/network/dns",
			body: `{"servers":["1.1.1.1"]}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(suite.ctrl)
			},
			wantCode:     http.StatusBadRequest,
			wantContains: []string{`"error"`, "InterfaceName", "required"},
		},
		{
			name: "when non-alphanum interface name",
			path: "/network/dns",
			body: `{"servers":["1.1.1.1"],"interface_name":"eth-0!"}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(suite.ctrl)
			},
			wantCode:     http.StatusBadRequest,
			wantContains: []string{`"error"`, "InterfaceName", "alphanum"},
		},
		{
			name: "when invalid server IP",
			path: "/network/dns",
			body: `{"servers":["not-an-ip"],"interface_name":"eth0"}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(suite.ctrl)
			},
			wantCode:     http.StatusBadRequest,
			wantContains: []string{`"error"`, "Servers", "ip"},
		},
		{
			name: "when invalid search domain",
			path: "/network/dns",
			body: `{"search_domains":["not a valid hostname!"],"interface_name":"eth0"}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(suite.ctrl)
			},
			wantCode:     http.StatusBadRequest,
			wantContains: []string{`"error"`, "SearchDomains", "hostname"},
		},
		{
			name: "when empty target_hostname returns 400",
			path: "/network/dns?target_hostname=",
			body: `{"servers":["1.1.1.1"],"interface_name":"eth0"}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(suite.ctrl)
			},
			wantCode:     http.StatusBadRequest,
			wantContains: []string{`"error"`},
		},
		{
			name: "when broadcast all",
			path: "/network/dns?target_hostname=_all",
			body: `{"servers":["1.1.1.1"],"search_domains":["foo.bar"],"interface_name":"eth0"}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(suite.ctrl)
				mock.EXPECT().
					ModifyNetworkDNSBroadcast(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("550e8400-e29b-41d4-a716-446655440000", map[string]error{
						"server1": nil,
					}, nil)
				return mock
			},
			wantCode:     http.StatusAccepted,
			wantContains: []string{`"results"`, `"server1"`},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			networkHandler := apinetwork.New(suite.logger, jobMock)
			strictHandler := networkGen.NewStrictHandler(networkHandler, nil)

			a := api.New(suite.appConfig, suite.logger)
			networkGen.RegisterHandlers(a.Echo, strictHandler)

			req := httptest.NewRequest(
				http.MethodPut,
				tc.path,
				strings.NewReader(tc.body),
			)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			a.Echo.ServeHTTP(rec, req)

			suite.Equal(tc.wantCode, rec.Code)
			for _, s := range tc.wantContains {
				suite.Contains(rec.Body.String(), s)
			}
		})
	}
}

const rbacDNSPutTestSigningKey = "test-signing-key-for-rbac-integration"

func (suite *NetworkDNSPutByInterfaceIntegrationTestSuite) TestPutNetworkDNSRBAC() {
	tokenManager := authtoken.New(suite.logger)

	tests := []struct {
		name         string
		setupAuth    func(req *http.Request)
		setupJobMock func() *jobmocks.MockJobClient
		wantCode     int
		wantContains []string
	}{
		{
			name: "when no token returns 401",
			setupAuth: func(_ *http.Request) {
				// No auth header set
			},
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(suite.ctrl)
			},
			wantCode:     http.StatusUnauthorized,
			wantContains: []string{"Bearer token required"},
		},
		{
			name: "when insufficient permissions returns 403",
			setupAuth: func(req *http.Request) {
				token, err := tokenManager.Generate(
					rbacDNSPutTestSigningKey,
					[]string{"read"},
					"test-user",
					[]string{"network:read"},
				)
				suite.Require().NoError(err)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			},
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(suite.ctrl)
			},
			wantCode:     http.StatusForbidden,
			wantContains: []string{"Insufficient permissions"},
		},
		{
			name: "when valid token with network:write returns 200",
			setupAuth: func(req *http.Request) {
				token, err := tokenManager.Generate(
					rbacDNSPutTestSigningKey,
					[]string{"admin"},
					"test-user",
					nil,
				)
				suite.Require().NoError(err)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			},
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(suite.ctrl)
				mock.EXPECT().
					ModifyNetworkDNS(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(
						"550e8400-e29b-41d4-a716-446655440000",
						"worker1",
						nil,
					)
				return mock
			},
			wantCode:     http.StatusAccepted,
			wantContains: []string{`"results"`},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			appConfig := config.Config{
				API: config.API{
					Server: config.Server{
						Security: config.ServerSecurity{
							SigningKey: rbacDNSPutTestSigningKey,
						},
					},
				},
			}

			server := api.New(appConfig, suite.logger)
			handlers := server.GetNetworkHandler(jobMock)
			server.RegisterHandlers(handlers)

			req := httptest.NewRequest(
				http.MethodPut,
				"/network/dns",
				strings.NewReader(`{"servers":["8.8.8.8"],"interface_name":"eth0"}`),
			)
			req.Header.Set("Content-Type", "application/json")
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

func TestNetworkDNSPutByInterfaceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(NetworkDNSPutByInterfaceIntegrationTestSuite))
}

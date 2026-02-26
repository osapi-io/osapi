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
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/api"
	apinetwork "github.com/retr0h/osapi/internal/api/network"
	networkGen "github.com/retr0h/osapi/internal/api/network/gen"
	"github.com/retr0h/osapi/internal/authtoken"
	"github.com/retr0h/osapi/internal/config"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
	"github.com/retr0h/osapi/internal/provider/network/dns"
	"github.com/retr0h/osapi/internal/validation"
)

type NetworkDNSGetByInterfaceIntegrationTestSuite struct {
	suite.Suite
	ctrl *gomock.Controller

	appConfig config.Config
	logger    *slog.Logger
}

func (suite *NetworkDNSGetByInterfaceIntegrationTestSuite) SetupSuite() {
	validation.RegisterTargetValidator(func(_ context.Context) ([]validation.WorkerTarget, error) {
		return []validation.WorkerTarget{
			{Hostname: "server1", Labels: map[string]string{"group": "web"}},
			{Hostname: "server2"},
		}, nil
	})
}

func (suite *NetworkDNSGetByInterfaceIntegrationTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())

	suite.appConfig = config.Config{}
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (suite *NetworkDNSGetByInterfaceIntegrationTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *NetworkDNSGetByInterfaceIntegrationTestSuite) TestGetNetworkDNSByInterfaceValidation() {
	tests := []struct {
		name         string
		path         string
		setupJobMock func() *jobmocks.MockJobClient
		wantCode     int
		wantContains []string
	}{
		{
			name: "when valid request",
			path: "/network/dns/eth0",
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(suite.ctrl)
				mock.EXPECT().
					QueryNetworkDNS(gomock.Any(), gomock.Any(), "eth0").
					Return("550e8400-e29b-41d4-a716-446655440000", &dns.Config{
						DNSServers:    []string{"8.8.8.8"},
						SearchDomains: []string{"example.com"},
					}, "worker1", nil)
				return mock
			},
			wantCode: http.StatusOK,
			wantContains: []string{
				`"results"`,
				`"servers"`,
				`"8.8.8.8"`,
				`"search_domains"`,
				`"example.com"`,
			},
		},
		{
			name: "when non-alphanum interface name",
			path: "/network/dns/eth-0!",
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(suite.ctrl)
			},
			wantCode:     http.StatusBadRequest,
			wantContains: []string{`"error"`, "InterfaceName", "alphanum"},
		},
		{
			name: "when empty target_hostname returns 400",
			path: "/network/dns/eth0?target_hostname=",
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(suite.ctrl)
			},
			wantCode:     http.StatusBadRequest,
			wantContains: []string{`"error"`},
		},
		{
			name: "when broadcast all",
			path: "/network/dns/eth0?target_hostname=_all",
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(suite.ctrl)
				mock.EXPECT().
					QueryNetworkDNSBroadcast(gomock.Any(), gomock.Any(), "eth0").
					Return("550e8400-e29b-41d4-a716-446655440000", map[string]*dns.Config{
						"server1": {
							DNSServers:    []string{"8.8.8.8"},
							SearchDomains: []string{"example.com"},
						},
					}, map[string]string{}, nil)
				return mock
			},
			wantCode:     http.StatusOK,
			wantContains: []string{`"results"`, `"8.8.8.8"`},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			networkHandler := apinetwork.New(suite.logger, jobMock)
			strictHandler := networkGen.NewStrictHandler(networkHandler, nil)

			a := api.New(suite.appConfig, suite.logger)
			networkGen.RegisterHandlers(a.Echo, strictHandler)

			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()

			a.Echo.ServeHTTP(rec, req)

			suite.Equal(tc.wantCode, rec.Code)
			for _, s := range tc.wantContains {
				suite.Contains(rec.Body.String(), s)
			}
		})
	}
}

const rbacDNSGetTestSigningKey = "test-signing-key-for-rbac-integration"

func (suite *NetworkDNSGetByInterfaceIntegrationTestSuite) TestGetNetworkDNSByInterfaceRBAC() {
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
					rbacDNSGetTestSigningKey,
					[]string{"read"},
					"test-user",
					[]string{"job:read"},
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
			name: "when valid token with network:read returns 200",
			setupAuth: func(req *http.Request) {
				token, err := tokenManager.Generate(
					rbacDNSGetTestSigningKey,
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
					QueryNetworkDNS(gomock.Any(), gomock.Any(), "eth0").
					Return(
						"550e8400-e29b-41d4-a716-446655440000",
						&dns.Config{
							DNSServers:    []string{"8.8.8.8"},
							SearchDomains: []string{"example.com"},
						},
						"worker1",
						nil,
					)
				return mock
			},
			wantCode:     http.StatusOK,
			wantContains: []string{`"results"`, `"8.8.8.8"`},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			appConfig := config.Config{
				API: config.API{
					Server: config.Server{
						Security: config.ServerSecurity{
							SigningKey: rbacDNSGetTestSigningKey,
						},
					},
				},
			}

			server := api.New(appConfig, suite.logger)
			handlers := server.GetNetworkHandler(jobMock)
			server.RegisterHandlers(handlers)

			req := httptest.NewRequest(http.MethodGet, "/network/dns/eth0", nil)
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

func TestNetworkDNSGetByInterfaceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(NetworkDNSGetByInterfaceIntegrationTestSuite))
}

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
	"github.com/retr0h/osapi/internal/config"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
	"github.com/retr0h/osapi/internal/provider/network/dns"
)

type NetworkDNSGetByInterfaceIntegrationTestSuite struct {
	suite.Suite
	ctrl *gomock.Controller

	appConfig config.Config
	logger    *slog.Logger
}

func (suite *NetworkDNSGetByInterfaceIntegrationTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())

	suite.appConfig = config.Config{}
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (suite *NetworkDNSGetByInterfaceIntegrationTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *NetworkDNSGetByInterfaceIntegrationTestSuite) TestGetNetworkDNSByInterface() {
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
					Return(&dns.Config{
						DNSServers:    []string{"8.8.8.8"},
						SearchDomains: []string{"example.com"},
					}, nil)
				return mock
			},
			wantCode:     http.StatusOK,
			wantContains: []string{`"servers"`, `"8.8.8.8"`, `"search_domains"`, `"example.com"`},
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
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			networkHandler := apinetwork.New(jobMock)
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

func TestNetworkDNSGetByInterfaceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(NetworkDNSGetByInterfaceIntegrationTestSuite))
}

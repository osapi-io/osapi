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

package node_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/api"
	apinode "github.com/retr0h/osapi/internal/api/node"
	"github.com/retr0h/osapi/internal/api/node/gen"
	"github.com/retr0h/osapi/internal/authtoken"
	"github.com/retr0h/osapi/internal/config"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
	"github.com/retr0h/osapi/internal/validation"
)

type NetworkDNSPutByInterfacePublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apinode.Node
	ctx           context.Context
	appConfig     config.Config
	logger        *slog.Logger
}

func (s *NetworkDNSPutByInterfacePublicTestSuite) SetupSuite() {
	validation.RegisterTargetValidator(func(_ context.Context) ([]validation.AgentTarget, error) {
		return []validation.AgentTarget{
			{Hostname: "server1", Labels: map[string]string{"group": "web"}},
			{Hostname: "server2"},
		}, nil
	})
}

func (s *NetworkDNSPutByInterfacePublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apinode.New(slog.Default(), s.mockJobClient)
	s.ctx = context.Background()
	s.appConfig = config.Config{}
	s.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (s *NetworkDNSPutByInterfacePublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *NetworkDNSPutByInterfacePublicTestSuite) TestPutNodeNetworkDNS() {
	tests := []struct {
		name         string
		request      gen.PutNodeNetworkDNSRequestObject
		setupMock    func()
		validateFunc func(resp gen.PutNodeNetworkDNSResponseObject)
	}{
		{
			name: "success",
			request: gen.PutNodeNetworkDNSRequestObject{
				Hostname: "_any",
				Body: &gen.PutNodeNetworkDNSJSONRequestBody{
					InterfaceName: "eth0",
					Servers:       &[]string{"1.1.1.1"},
					SearchDomains: &[]string{"example.com"},
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyNetworkDNS(
						gomock.Any(),
						"_any",
						[]string{"1.1.1.1"},
						[]string{"example.com"},
						"eth0",
					).
					Return(
						"550e8400-e29b-41d4-a716-446655440000",
						"agent1",
						true,
						nil,
					)
			},
			validateFunc: func(resp gen.PutNodeNetworkDNSResponseObject) {
				r, ok := resp.(gen.PutNodeNetworkDNS202JSONResponse)
				s.True(ok)
				s.Require().Len(r.Results, 1)
				s.Equal("agent1", r.Results[0].Hostname)
				s.Equal(gen.Ok, r.Results[0].Status)
				s.Require().NotNil(r.Results[0].Changed)
				s.True(*r.Results[0].Changed)
			},
		},
		{
			name: "validation error empty hostname",
			request: gen.PutNodeNetworkDNSRequestObject{
				Hostname: "",
				Body: &gen.PutNodeNetworkDNSJSONRequestBody{
					InterfaceName: "eth0",
					Servers:       &[]string{"1.1.1.1"},
				},
			},
			setupMock: func() {},
			validateFunc: func(resp gen.PutNodeNetworkDNSResponseObject) {
				r, ok := resp.(gen.PutNodeNetworkDNS400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "required")
			},
		},
		{
			name: "body validation error empty interface name",
			request: gen.PutNodeNetworkDNSRequestObject{
				Hostname: "_any",
				Body: &gen.PutNodeNetworkDNSJSONRequestBody{
					InterfaceName: "",
					Servers:       &[]string{"1.1.1.1"},
				},
			},
			setupMock: func() {},
			validateFunc: func(resp gen.PutNodeNetworkDNSResponseObject) {
				r, ok := resp.(gen.PutNodeNetworkDNS400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
			},
		},
		{
			name: "job client error",
			request: gen.PutNodeNetworkDNSRequestObject{
				Hostname: "_any",
				Body: &gen.PutNodeNetworkDNSJSONRequestBody{
					InterfaceName: "eth0",
					Servers:       &[]string{"1.1.1.1"},
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyNetworkDNS(
						gomock.Any(),
						"_any",
						[]string{"1.1.1.1"},
						[]string(nil),
						"eth0",
					).
					Return("", "", false, assert.AnError)
			},
			validateFunc: func(resp gen.PutNodeNetworkDNSResponseObject) {
				_, ok := resp.(gen.PutNodeNetworkDNS500JSONResponse)
				s.True(ok)
			},
		},
		{
			name: "broadcast all success",
			request: gen.PutNodeNetworkDNSRequestObject{
				Hostname: "_all",
				Body: &gen.PutNodeNetworkDNSJSONRequestBody{
					InterfaceName: "eth0",
					Servers:       &[]string{"1.1.1.1"},
					SearchDomains: &[]string{"example.com"},
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyNetworkDNSBroadcast(
						gomock.Any(),
						"_all",
						[]string{"1.1.1.1"},
						[]string{"example.com"},
						"eth0",
					).
					Return(
						"550e8400-e29b-41d4-a716-446655440000",
						map[string]error{
							"server1": nil,
							"server2": nil,
						},
						map[string]bool{
							"server1": true,
							"server2": false,
						},
						nil,
					)
			},
			validateFunc: func(resp gen.PutNodeNetworkDNSResponseObject) {
				s.NotNil(resp)
			},
		},
		{
			name: "broadcast all with errors",
			request: gen.PutNodeNetworkDNSRequestObject{
				Hostname: "_all",
				Body: &gen.PutNodeNetworkDNSJSONRequestBody{
					InterfaceName: "eth0",
					Servers:       &[]string{"1.1.1.1"},
					SearchDomains: &[]string{"example.com"},
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyNetworkDNSBroadcast(
						gomock.Any(),
						"_all",
						[]string{"1.1.1.1"},
						[]string{"example.com"},
						"eth0",
					).
					Return(
						"550e8400-e29b-41d4-a716-446655440000",
						map[string]error{
							"server1": nil,
							"server2": errors.New("permission denied"),
						},
						map[string]bool{
							"server1": true,
							"server2": false,
						},
						nil,
					)
			},
			validateFunc: func(resp gen.PutNodeNetworkDNSResponseObject) {
				r, ok := resp.(gen.PutNodeNetworkDNS202JSONResponse)
				s.True(ok)
				s.Len(r.Results, 2)
				var foundError bool
				for _, item := range r.Results {
					if item.Error != nil {
						foundError = true
						s.Equal("server2", item.Hostname)
						s.Equal(gen.Failed, item.Status)
					}
				}
				s.True(foundError)
			},
		},
		{
			name: "broadcast all error",
			request: gen.PutNodeNetworkDNSRequestObject{
				Hostname: "_all",
				Body: &gen.PutNodeNetworkDNSJSONRequestBody{
					InterfaceName: "eth0",
					Servers:       &[]string{"1.1.1.1"},
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					ModifyNetworkDNSBroadcast(
						gomock.Any(),
						"_all",
						[]string{"1.1.1.1"},
						[]string(nil),
						"eth0",
					).
					Return("", nil, nil, assert.AnError)
			},
			validateFunc: func(resp gen.PutNodeNetworkDNSResponseObject) {
				_, ok := resp.(gen.PutNodeNetworkDNS500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()

			resp, err := s.handler.PutNodeNetworkDNS(s.ctx, tt.request)
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func (s *NetworkDNSPutByInterfacePublicTestSuite) TestPutNetworkDNSHTTP() {
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
			path: "/node/server1/network/dns",
			body: `{"servers":["1.1.1.1","8.8.8.8"],"search_domains":["foo.bar"],"interface_name":"eth0"}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(s.mockCtrl)
				mock.EXPECT().
					ModifyNetworkDNS(gomock.Any(), "server1", gomock.Any(), gomock.Any(), gomock.Any()).
					Return("550e8400-e29b-41d4-a716-446655440000", "agent1", true, nil)
				return mock
			},
			wantCode:     http.StatusAccepted,
			wantContains: []string{`"results"`, `"agent1"`, `"ok"`, `"changed":true`},
		},
		{
			name: "when missing interface name",
			path: "/node/server1/network/dns",
			body: `{"servers":["1.1.1.1"]}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(s.mockCtrl)
			},
			wantCode:     http.StatusBadRequest,
			wantContains: []string{`"error"`, "InterfaceName", "required"},
		},
		{
			name: "when non-alphanum interface name",
			path: "/node/server1/network/dns",
			body: `{"servers":["1.1.1.1"],"interface_name":"eth-0!"}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(s.mockCtrl)
			},
			wantCode:     http.StatusBadRequest,
			wantContains: []string{`"error"`, "InterfaceName", "alphanum"},
		},
		{
			name: "when invalid server IP",
			path: "/node/server1/network/dns",
			body: `{"servers":["not-an-ip"],"interface_name":"eth0"}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(s.mockCtrl)
			},
			wantCode:     http.StatusBadRequest,
			wantContains: []string{`"error"`, "Servers", "ip"},
		},
		{
			name: "when invalid search domain",
			path: "/node/server1/network/dns",
			body: `{"search_domains":["not a valid hostname!"],"interface_name":"eth0"}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(s.mockCtrl)
			},
			wantCode:     http.StatusBadRequest,
			wantContains: []string{`"error"`, "SearchDomains", "hostname"},
		},
		{
			name: "when target agent not found",
			path: "/node/nonexistent/network/dns",
			body: `{"servers":["1.1.1.1"],"interface_name":"eth0"}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(s.mockCtrl)
			},
			wantCode:     http.StatusBadRequest,
			wantContains: []string{`"error"`, "valid_target", "not found"},
		},
		{
			name: "when broadcast all",
			path: "/node/_all/network/dns",
			body: `{"servers":["1.1.1.1"],"search_domains":["foo.bar"],"interface_name":"eth0"}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(s.mockCtrl)
				mock.EXPECT().
					ModifyNetworkDNSBroadcast(gomock.Any(), "_all", gomock.Any(), gomock.Any(), gomock.Any()).
					Return("550e8400-e29b-41d4-a716-446655440000", map[string]error{
						"server1": nil,
					}, map[string]bool{
						"server1": true,
					}, nil)
				return mock
			},
			wantCode:     http.StatusAccepted,
			wantContains: []string{`"results"`, `"server1"`, `"changed":true`},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			nodeHandler := apinode.New(s.logger, jobMock)
			strictHandler := gen.NewStrictHandler(nodeHandler, nil)

			a := api.New(s.appConfig, s.logger)
			gen.RegisterHandlers(a.Echo, strictHandler)

			req := httptest.NewRequest(
				http.MethodPut,
				tc.path,
				strings.NewReader(tc.body),
			)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			a.Echo.ServeHTTP(rec, req)

			s.Equal(tc.wantCode, rec.Code)
			for _, str := range tc.wantContains {
				s.Contains(rec.Body.String(), str)
			}
		})
	}
}

const rbacDNSPutTestSigningKey = "test-signing-key-for-dns-put-rbac"

func (s *NetworkDNSPutByInterfacePublicTestSuite) TestPutNetworkDNSRBACHTTP() {
	tokenManager := authtoken.New(s.logger)

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
				return jobmocks.NewMockJobClient(s.mockCtrl)
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
				s.Require().NoError(err)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			},
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(s.mockCtrl)
			},
			wantCode:     http.StatusForbidden,
			wantContains: []string{"Insufficient permissions"},
		},
		{
			name: "when valid token with network:write returns 202",
			setupAuth: func(req *http.Request) {
				token, err := tokenManager.Generate(
					rbacDNSPutTestSigningKey,
					[]string{"admin"},
					"test-user",
					nil,
				)
				s.Require().NoError(err)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			},
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(s.mockCtrl)
				mock.EXPECT().
					ModifyNetworkDNS(gomock.Any(), "server1", gomock.Any(), gomock.Any(), gomock.Any()).
					Return(
						"550e8400-e29b-41d4-a716-446655440000",
						"agent1",
						true,
						nil,
					)
				return mock
			},
			wantCode:     http.StatusAccepted,
			wantContains: []string{`"results"`, `"changed":true`},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
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

			server := api.New(appConfig, s.logger)
			handlers := server.GetNodeHandler(jobMock)
			server.RegisterHandlers(handlers)

			req := httptest.NewRequest(
				http.MethodPut,
				"/node/server1/network/dns",
				strings.NewReader(`{"servers":["8.8.8.8"],"interface_name":"eth0"}`),
			)
			req.Header.Set("Content-Type", "application/json")
			tc.setupAuth(req)
			rec := httptest.NewRecorder()

			server.Echo.ServeHTTP(rec, req)

			s.Equal(tc.wantCode, rec.Code)
			for _, str := range tc.wantContains {
				s.Contains(rec.Body.String(), str)
			}
		})
	}
}

func TestNetworkDNSPutByInterfacePublicTestSuite(t *testing.T) {
	suite.Run(t, new(NetworkDNSPutByInterfacePublicTestSuite))
}

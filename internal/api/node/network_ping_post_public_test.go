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
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/api"
	apinode "github.com/retr0h/osapi/internal/api/node"
	"github.com/retr0h/osapi/internal/api/node/gen"
	"github.com/retr0h/osapi/internal/authtoken"
	"github.com/retr0h/osapi/internal/config"
	jobmocks "github.com/retr0h/osapi/internal/job/mocks"
	"github.com/retr0h/osapi/internal/provider/network/ping"
	"github.com/retr0h/osapi/internal/validation"
)

type NetworkPingPostPublicTestSuite struct {
	suite.Suite

	mockCtrl      *gomock.Controller
	mockJobClient *jobmocks.MockJobClient
	handler       *apinode.Node
	ctx           context.Context
	appConfig     config.Config
	logger        *slog.Logger
}

func (s *NetworkPingPostPublicTestSuite) SetupSuite() {
	validation.RegisterTargetValidator(func(_ context.Context) ([]validation.AgentTarget, error) {
		return []validation.AgentTarget{
			{Hostname: "server1", Labels: map[string]string{"group": "web"}},
			{Hostname: "server2"},
		}, nil
	})
}

func (s *NetworkPingPostPublicTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.mockJobClient = jobmocks.NewMockJobClient(s.mockCtrl)
	s.handler = apinode.New(slog.Default(), s.mockJobClient)
	s.ctx = context.Background()
	s.appConfig = config.Config{}
	s.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (s *NetworkPingPostPublicTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *NetworkPingPostPublicTestSuite) TestPostNodeNetworkPing() {
	tests := []struct {
		name         string
		request      gen.PostNodeNetworkPingRequestObject
		setupMock    func()
		validateFunc func(resp gen.PostNodeNetworkPingResponseObject)
	}{
		{
			name: "success",
			request: gen.PostNodeNetworkPingRequestObject{
				Hostname: "_any",
				Body: &gen.PostNodeNetworkPingJSONRequestBody{
					Address: "8.8.8.8",
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNetworkPing(gomock.Any(), "_any", "8.8.8.8").
					Return(
						"550e8400-e29b-41d4-a716-446655440000",
						&ping.Result{
							AvgRTT:          20 * time.Millisecond,
							MaxRTT:          25 * time.Millisecond,
							MinRTT:          15 * time.Millisecond,
							PacketLoss:      0,
							PacketsReceived: 3,
							PacketsSent:     3,
						},
						"agent1",
						nil,
					)
			},
			validateFunc: func(resp gen.PostNodeNetworkPingResponseObject) {
				r, ok := resp.(gen.PostNodeNetworkPing200JSONResponse)
				s.True(ok)
				s.Require().Len(r.Results, 1)
				s.Equal("agent1", r.Results[0].Hostname)
				s.Require().NotNil(r.Results[0].PacketsSent)
				s.Equal(3, *r.Results[0].PacketsSent)
				s.Require().NotNil(r.Results[0].PacketsReceived)
				s.Equal(3, *r.Results[0].PacketsReceived)
			},
		},
		{
			name: "validation error empty hostname",
			request: gen.PostNodeNetworkPingRequestObject{
				Hostname: "",
				Body: &gen.PostNodeNetworkPingJSONRequestBody{
					Address: "8.8.8.8",
				},
			},
			setupMock: func() {},
			validateFunc: func(resp gen.PostNodeNetworkPingResponseObject) {
				r, ok := resp.(gen.PostNodeNetworkPing400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
				s.Contains(*r.Error, "required")
			},
		},
		{
			name: "body validation error empty address",
			request: gen.PostNodeNetworkPingRequestObject{
				Hostname: "_any",
				Body: &gen.PostNodeNetworkPingJSONRequestBody{
					Address: "",
				},
			},
			setupMock: func() {},
			validateFunc: func(resp gen.PostNodeNetworkPingResponseObject) {
				r, ok := resp.(gen.PostNodeNetworkPing400JSONResponse)
				s.True(ok)
				s.Require().NotNil(r.Error)
			},
		},
		{
			name: "job client error",
			request: gen.PostNodeNetworkPingRequestObject{
				Hostname: "_any",
				Body: &gen.PostNodeNetworkPingJSONRequestBody{
					Address: "8.8.8.8",
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNetworkPing(gomock.Any(), "_any", "8.8.8.8").
					Return("", nil, "", assert.AnError)
			},
			validateFunc: func(resp gen.PostNodeNetworkPingResponseObject) {
				_, ok := resp.(gen.PostNodeNetworkPing500JSONResponse)
				s.True(ok)
			},
		},
		{
			name: "broadcast all success",
			request: gen.PostNodeNetworkPingRequestObject{
				Hostname: "_all",
				Body: &gen.PostNodeNetworkPingJSONRequestBody{
					Address: "8.8.8.8",
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNetworkPingBroadcast(gomock.Any(), "_all", "8.8.8.8").
					Return(
						"550e8400-e29b-41d4-a716-446655440000",
						map[string]*ping.Result{
							"server1": {
								AvgRTT:          20 * time.Millisecond,
								PacketsSent:     3,
								PacketsReceived: 3,
							},
							"server2": {
								AvgRTT:          30 * time.Millisecond,
								PacketsSent:     3,
								PacketsReceived: 3,
							},
						},
						map[string]string{},
						nil,
					)
			},
			validateFunc: func(resp gen.PostNodeNetworkPingResponseObject) {
				s.NotNil(resp)
			},
		},
		{
			name: "broadcast all with errors",
			request: gen.PostNodeNetworkPingRequestObject{
				Hostname: "_all",
				Body: &gen.PostNodeNetworkPingJSONRequestBody{
					Address: "8.8.8.8",
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNetworkPingBroadcast(gomock.Any(), "_all", "8.8.8.8").
					Return(
						"550e8400-e29b-41d4-a716-446655440000",
						map[string]*ping.Result{
							"server1": {
								AvgRTT:          20 * time.Millisecond,
								PacketsSent:     3,
								PacketsReceived: 3,
							},
						},
						map[string]string{
							"server2": "host unreachable",
						},
						nil,
					)
			},
			validateFunc: func(resp gen.PostNodeNetworkPingResponseObject) {
				r, ok := resp.(gen.PostNodeNetworkPing200JSONResponse)
				s.True(ok)
				s.Len(r.Results, 2)
				var foundError bool
				for _, h := range r.Results {
					if h.Error != nil {
						foundError = true
						s.Equal("server2", h.Hostname)
						s.Equal("host unreachable", *h.Error)
					}
				}
				s.True(foundError)
			},
		},
		{
			name: "broadcast all error",
			request: gen.PostNodeNetworkPingRequestObject{
				Hostname: "_all",
				Body: &gen.PostNodeNetworkPingJSONRequestBody{
					Address: "8.8.8.8",
				},
			},
			setupMock: func() {
				s.mockJobClient.EXPECT().
					QueryNetworkPingBroadcast(gomock.Any(), "_all", "8.8.8.8").
					Return("", nil, nil, assert.AnError)
			},
			validateFunc: func(resp gen.PostNodeNetworkPingResponseObject) {
				_, ok := resp.(gen.PostNodeNetworkPing500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setupMock()

			resp, err := s.handler.PostNodeNetworkPing(s.ctx, tt.request)
			s.NoError(err)
			tt.validateFunc(resp)
		})
	}
}

func (s *NetworkPingPostPublicTestSuite) TestPostNetworkPingHTTP() {
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
			path: "/node/server1/network/ping",
			body: `{"address":"1.1.1.1"}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(s.mockCtrl)
				mock.EXPECT().
					QueryNetworkPing(gomock.Any(), "server1", "1.1.1.1").
					Return("550e8400-e29b-41d4-a716-446655440000", &ping.Result{
						PacketsSent:     3,
						PacketsReceived: 3,
						PacketLoss:      0,
						MinRTT:          10 * time.Millisecond,
						AvgRTT:          15 * time.Millisecond,
						MaxRTT:          20 * time.Millisecond,
					}, "agent1", nil)
				return mock
			},
			wantCode:     http.StatusOK,
			wantContains: []string{`"results"`, `"packets_sent":3`, `"packets_received":3`},
		},
		{
			name: "when missing address",
			path: "/node/server1/network/ping",
			body: `{}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(s.mockCtrl)
			},
			wantCode:     http.StatusBadRequest,
			wantContains: []string{`"error"`, "Address", "required"},
		},
		{
			name: "when invalid address format",
			path: "/node/server1/network/ping",
			body: `{"address":"not-an-ip"}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				return jobmocks.NewMockJobClient(s.mockCtrl)
			},
			wantCode:     http.StatusBadRequest,
			wantContains: []string{`"error"`, "Address", "ip"},
		},
		{
			name: "when broadcast all",
			path: "/node/_all/network/ping",
			body: `{"address":"1.1.1.1"}`,
			setupJobMock: func() *jobmocks.MockJobClient {
				mock := jobmocks.NewMockJobClient(s.mockCtrl)
				mock.EXPECT().
					QueryNetworkPingBroadcast(gomock.Any(), "_all", "1.1.1.1").
					Return("550e8400-e29b-41d4-a716-446655440000", map[string]*ping.Result{
						"server1": {
							PacketsSent:     3,
							PacketsReceived: 3,
							PacketLoss:      0,
							MinRTT:          10 * time.Millisecond,
							AvgRTT:          15 * time.Millisecond,
							MaxRTT:          20 * time.Millisecond,
						},
					}, map[string]string{}, nil)
				return mock
			},
			wantCode:     http.StatusOK,
			wantContains: []string{`"results"`, `"packets_sent":3`},
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
				http.MethodPost,
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

const rbacPingTestSigningKey = "test-signing-key-for-ping-rbac"

func (s *NetworkPingPostPublicTestSuite) TestPostNetworkPingRBACHTTP() {
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
					rbacPingTestSigningKey,
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
			name: "when valid token with network:write returns 200",
			setupAuth: func(req *http.Request) {
				token, err := tokenManager.Generate(
					rbacPingTestSigningKey,
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
					QueryNetworkPing(gomock.Any(), "server1", "8.8.8.8").
					Return(
						"550e8400-e29b-41d4-a716-446655440000",
						&ping.Result{
							PacketsSent:     3,
							PacketsReceived: 3,
							PacketLoss:      0,
							MinRTT:          10 * time.Millisecond,
							AvgRTT:          15 * time.Millisecond,
							MaxRTT:          20 * time.Millisecond,
						},
						"agent1",
						nil,
					)
				return mock
			},
			wantCode:     http.StatusOK,
			wantContains: []string{`"results"`, `"packets_sent":3`},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			jobMock := tc.setupJobMock()

			appConfig := config.Config{
				API: config.API{
					Server: config.Server{
						Security: config.ServerSecurity{
							SigningKey: rbacPingTestSigningKey,
						},
					},
				},
			}

			server := api.New(appConfig, s.logger)
			handlers := server.GetNodeHandler(jobMock)
			server.RegisterHandlers(handlers)

			req := httptest.NewRequest(
				http.MethodPost,
				"/node/server1/network/ping",
				strings.NewReader(`{"address":"8.8.8.8"}`),
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

func TestNetworkPingPostPublicTestSuite(t *testing.T) {
	suite.Run(t, new(NetworkPingPostPublicTestSuite))
}

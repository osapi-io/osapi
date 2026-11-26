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
package health_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/internal/controller/endpoint/server"
	"github.com/osapi-io/osapi/internal/controller/endpoints/health"
	"github.com/osapi-io/osapi/internal/controller/endpoints/health/gen"
	"github.com/osapi-io/osapi/internal/controller/endpoints/health/mocks"
)

var signingKey = authtoken.Key{ID: "test", Secret: "test-signing-key-for-the-health-group"}

var degraded = health.StatusResponse{
	Status:  "degraded",
	Checks:  []health.Check{{Name: "nats", Status: "error", Error: "nats down"}},
	Version: "0.1.0",
	Uptime:  "1m0s",
}

type MethodsPublicTestSuite struct {
	suite.Suite

	ctrl     *gomock.Controller
	reporter *mocks.MockReporter
}

func (s *MethodsPublicTestSuite) SetupSubTest() {
	s.ctrl = gomock.NewController(s.T())
	s.reporter = mocks.NewMockReporter(s.ctrl)
}

func (s *MethodsPublicTestSuite) TearDownSubTest() {
	s.ctrl.Finish()
}

func (s *MethodsPublicTestSuite) TestStatusComponent() {
	tests := []struct {
		name string
		kind types.ComponentType
		want []string
	}{
		{
			name: "when components have the type, the registry has only those",
			kind: "agent",
			want: []string{"web-01", "web-02"},
		},
		{
			name: "when none has the type, the registry is empty, not null",
			kind: "nothing",
			want: []string{},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.reporter.EXPECT().
				Status(gomock.Any()).
				Return(health.StatusResponse{Registry: []health.ComponentEntry{
					{
						Type:       "controller",
						Hostname:   "hub",
						Components: []health.ComponentEntry{{Type: "api", Status: "ok"}},
					},
					{Type: "nats", Hostname: "hub"},
					{Type: "agent", Hostname: "web-01"},
					{Type: "agent", Hostname: "web-02"},
				}})

			got, err := health.Status(context.Background(),
				health.Deps{Health: s.reporter}, health.StatusInput{Component: tc.kind})
			s.Require().NoError(err)

			s.NotNil(got.Registry)
			hosts := []string{}
			for _, e := range got.Registry {
				hosts = append(hosts, e.Hostname)
			}
			s.Equal(tc.want, hosts)
		})
	}
}

func (s *MethodsPublicTestSuite) TestLiveness() {
	got, err := health.Liveness(context.Background(), health.Deps{}, health.LivenessInput{})
	s.Require().NoError(err)
	s.Equal(health.Response{Status: "ok"}, got)
}

func (s *MethodsPublicTestSuite) TestReady() {
	tests := []struct {
		name         string
		err          error
		validateFunc func(health.ReadyResponse, error)
	}{
		{
			name: "when the dependencies answer, it is ready",
			validateFunc: func(got health.ReadyResponse, err error) {
				s.Require().NoError(err)
				s.Equal(health.ReadyResponse{Status: "ready"}, got)
			},
		},
		{
			name: "when a dependency fails, it answers 503 with the reason",
			err:  errors.New("nats down"),
			validateFunc: func(got health.ReadyResponse, err error) {
				var answer *endpoint.AnswerError
				s.Require().ErrorAs(err, &answer)
				s.Equal(http.StatusServiceUnavailable, answer.Code)
				s.Equal(health.ReadyResponse{Status: "not_ready", Error: "nats down"}, got)
			},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.reporter.EXPECT().Ready(gomock.Any()).Return(tc.err)
			tc.validateFunc(health.Ready(context.Background(),
				health.Deps{Health: s.reporter}, health.ReadyInput{}))
		})
	}
}

func (s *MethodsPublicTestSuite) TestStatus() {
	tests := []struct {
		name string
		resp health.StatusResponse
	}{
		{
			name: "when each component is healthy, it answers the status",
			resp: health.StatusResponse{Status: "ok", Version: "0.1.0", Uptime: "1m0s"},
		},
		{
			name: "when a component is not healthy, it answers the degraded status, not an error",
			resp: degraded,
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.reporter.EXPECT().Status(gomock.Any()).Return(tc.resp)
			got, err := health.Status(context.Background(),
				health.Deps{Health: s.reporter}, health.StatusInput{})
			s.Require().NoError(err)
			s.Equal(tc.resp, got)
		})
	}
}

func token(
	perms ...string,
) string {
	t, _ := authtoken.New(slog.Default()).
		Generate(signingKey, []types.Role{types.RoleRead}, "test", perms, nil)

	return t
}

// TestRoutes runs requests through the generated routes, with the auth and
// the status mapping of the endpoint runtime.
func (s *MethodsPublicTestSuite) TestRoutes() {
	tests := []struct {
		name     string
		path     string
		token    string
		setup    func()
		wantCode int
		wantBody string
	}{
		{
			name:     "when the liveness probe has no token, it is a 200",
			path:     "/api/health",
			setup:    func() {},
			wantCode: http.StatusOK,
			wantBody: `{"status":"ok"}`,
		},
		{
			name:     "when the readiness probe has no token and it is ready, it is a 200",
			path:     "/api/health/ready",
			setup:    func() { s.reporter.EXPECT().Ready(gomock.Any()).Return(nil) },
			wantCode: http.StatusOK,
			wantBody: `{"status":"ready"}`,
		},
		{
			name: "when it is not ready, the probe is a 503 with the reason",
			path: "/api/health/ready",
			setup: func() {
				s.reporter.EXPECT().Ready(gomock.Any()).Return(errors.New("nats down"))
			},
			wantCode: http.StatusServiceUnavailable,
			wantBody: `{"status":"not_ready","error":"nats down"}`,
		},
		{
			name:     "when status has no token, it is a 401",
			path:     "/api/health/status",
			setup:    func() {},
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "when the token lacks health:read, status is a 403",
			path:     "/api/health/status",
			token:    token("agent:read"),
			setup:    func() {},
			wantCode: http.StatusForbidden,
		},
		{
			name:  "when the token has health:read and each component is healthy, it is a 200",
			path:  "/api/health/status",
			token: token("health:read"),
			setup: func() {
				s.reporter.EXPECT().Status(gomock.Any()).Return(health.StatusResponse{
					Status: "ok", Version: "0.1.0", Uptime: "1m0s",
				})
			},
			wantCode: http.StatusOK,
			wantBody: `{"status":"ok","version":"0.1.0","uptime":"1m0s","checks":null,"consumers":0}`,
		},
		{
			name:  "when a component is not healthy, status is a 200 with the degraded status",
			path:  "/api/health/status",
			token: token("health:read"),
			setup: func() {
				s.reporter.EXPECT().Status(gomock.Any()).Return(degraded)
			},
			wantCode: http.StatusOK,
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()
			srv := server.New(
				slog.Default(),
				authtoken.Keys{signingKey},
				nil,
				health.Deps{Health: s.reporter},
			)
			e := echo.New()
			srv.Mount(e, gen.Routes(srv))
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			s.Equal(tc.wantCode, rec.Code, rec.Body.String())
			if tc.wantBody != "" {
				s.JSONEq(tc.wantBody, rec.Body.String())
			}
		})
	}
}

// TestSpecConditionType: the spec's ConditionType has each condition, so a
// client checks a value when it compiles.
func (s *MethodsPublicTestSuite) TestSpecConditionType() {
	doc, err := endpoint.OpenAPI(health.Group)
	s.Require().NoError(err)

	conditions := make([]any, 0, len(types.ConditionType("").Values()))
	for _, v := range types.ConditionType("").Values() {
		conditions = append(conditions, string(v))
	}
	s.Equal(conditions, doc.Components.Schemas["ConditionType"].Value.Enum)
}

func TestMethodsPublicTestSuite(t *testing.T) {
	suite.Run(t, new(MethodsPublicTestSuite))
}

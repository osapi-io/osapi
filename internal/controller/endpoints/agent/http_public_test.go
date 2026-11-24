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
package agent_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/controller/endpoint/server"
	"github.com/osapi-io/osapi/internal/controller/endpoints/agent"
	"github.com/osapi-io/osapi/internal/controller/endpoints/agent/gen"
	"github.com/osapi-io/osapi/internal/controller/endpoints/agent/mocks"
	"github.com/osapi-io/osapi/internal/job"
)

var signingKey = authtoken.Key{ID: "test", Secret: "test-signing-key-for-the-agent-group"}

// HTTPPublicTestSuite runs requests through the agent group's generated
// routes, with the auth, the decoding and the status mapping of the
// endpoint runtime, as the controller serves them.
type HTTPPublicTestSuite struct {
	suite.Suite

	ctrl *gomock.Controller
	jobs *mocks.MockJobs
}

func (s *HTTPPublicTestSuite) SetupSubTest() {
	s.ctrl = gomock.NewController(s.T())
	s.jobs = mocks.NewMockJobs(s.ctrl)
}

func (s *HTTPPublicTestSuite) TearDownSubTest() {
	s.ctrl.Finish()
}

func token(
	perms ...string,
) string {
	t, _ := authtoken.New(slog.Default()).Generate(signingKey, []string{"read"}, "test", perms, nil)

	return t
}

func (s *HTTPPublicTestSuite) serve(
	method string,
	path string,
	tok string,
) *httptest.ResponseRecorder {
	srv := server.New(
		slog.Default(),
		authtoken.Keys{signingKey},
		nil,
		agent.Deps{Jobs: s.jobs},
	)
	e := echo.New()
	srv.Mount(e, gen.Routes(srv))

	req := httptest.NewRequest(method, path, nil)
	if tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	return rec
}

func (s *HTTPPublicTestSuite) TestRoutes() {
	read, write := token("agent:read"), token("agent:write")
	long := strings.Repeat("a", 256)

	tests := []struct {
		name     string
		method   string
		path     string
		token    string
		setup    func()
		wantCode int
		wantBody string
	}{
		{
			name:     "when list has a token with agent:read, it is a 200",
			method:   http.MethodGet,
			path:     "/api/agent",
			token:    read,
			setup:    func() { s.jobs.EXPECT().ListAgents(gomock.Any()).Return(nil, nil) },
			wantCode: http.StatusOK,
			wantBody: `{"items":[],"total_items":0}`,
		},
		{
			name:     "when there is no token, it is a 401",
			method:   http.MethodGet,
			path:     "/api/agent",
			setup:    func() {},
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "when the token lacks agent:read, it is a 403",
			method:   http.MethodGet,
			path:     "/api/agent",
			token:    token("job:read"),
			setup:    func() {},
			wantCode: http.StatusForbidden,
		},
		{
			name:     "when the hostname is longer than 255, it is a 400",
			method:   http.MethodGet,
			path:     "/api/agent/" + long,
			token:    read,
			setup:    func() {},
			wantCode: http.StatusBadRequest,
		},
		{
			name:   "when get finds the agent, its details are the body",
			method: http.MethodGet,
			path:   "/api/agent/server1",
			token:  read,
			setup: func() {
				s.jobs.EXPECT().GetAgent(gomock.Any(), "server1").
					Return(&job.AgentInfo{Hostname: "server1"}, nil)
			},
			wantCode: http.StatusOK,
			wantBody: `{"hostname":"server1","state":"Ready","key_stored":false,"verified":false}`,
		},
		{
			name:   "when get finds no agent, it is a 404",
			method: http.MethodGet,
			path:   "/api/agent/server9",
			token:  read,
			setup: func() {
				s.jobs.EXPECT().GetAgent(gomock.Any(), "server9").Return(nil, job.ErrNotFound)
			},
			wantCode: http.StatusNotFound,
		},
		{
			name:   "when drain finds the agent already draining, it is a 409",
			method: http.MethodPost,
			path:   "/api/agent/server1/drain",
			token:  write,
			setup: func() {
				s.jobs.EXPECT().GetAgent(gomock.Any(), "server1").
					Return(&job.AgentInfo{Hostname: "server1", State: types.AgentDraining}, nil)
			},
			wantCode: http.StatusConflict,
		},
		{
			name:   "when drain cannot write its flag, it is a 500",
			method: http.MethodPost,
			path:   "/api/agent/server1/drain",
			token:  write,
			setup: func() {
				s.jobs.EXPECT().GetAgent(gomock.Any(), "server1").
					Return(&job.AgentInfo{Hostname: "server1", MachineID: "m-1"}, nil)
				s.jobs.EXPECT().SetDrainFlag(gomock.Any(), "m-1").Return(errBoom)
			},
			wantCode: http.StatusInternalServerError,
		},
		{
			name:   "when undrain succeeds, it is a 204",
			method: http.MethodPost,
			path:   "/api/agent/server1/undrain",
			token:  write,
			setup: func() {
				s.jobs.EXPECT().GetAgent(gomock.Any(), "server1").
					Return(&job.AgentInfo{Hostname: "server1", MachineID: "m-1", State: types.AgentDraining}, nil)
				s.jobs.EXPECT().DeleteDrainFlag(gomock.Any(), "m-1").Return(nil)
				s.jobs.EXPECT().WriteAgentTimelineEvent(gomock.Any(), "server1", "undrain",
					"Undrain initiated via API").Return(nil)
			},
			wantCode: http.StatusNoContent,
		},
		{
			name:     "when a write has only agent:read, it is a 403",
			method:   http.MethodPost,
			path:     "/api/agent/server1/drain",
			token:    read,
			setup:    func() {},
			wantCode: http.StatusForbidden,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()

			rec := s.serve(tc.method, tc.path, tc.token)

			s.Equal(tc.wantCode, rec.Code, rec.Body.String())
			if tc.wantBody != "" {
				s.JSONEq(tc.wantBody, rec.Body.String())
			}
		})
	}
}

func TestHTTPPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(HTTPPublicTestSuite))
}

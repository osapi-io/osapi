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

package behavior_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller/api"
	apibehavior "github.com/osapi-io/osapi/internal/controller/api/behavior"
	jobmocks "github.com/osapi-io/osapi/internal/job/mocks"
)

const signingKey = "test-signing-key-for-the-behavior-handler"

type kvProvider struct{}

type kvIn struct {
	Key   string `json:"key"   validate:"required"`
	Value string `json:"value"`
	Count int    `json:"count"`
}

type none struct{}

type kvEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type kvResult struct {
	Key     string `json:"key"`
	Changed bool   `json:"changed"`
}

func kvList(_ context.Context, _ kvProvider, _ none) ([]kvEntry, error)  { return nil, nil }
func kvGet(_ context.Context, _ kvProvider, _ kvIn) (kvEntry, error)     { return kvEntry{}, nil }
func kvCreate(_ context.Context, _ kvProvider, _ kvIn) (kvResult, error) { return kvResult{}, nil }
func kvTouch(_ context.Context, _ kvProvider, _ kvIn) (kvEntry, error)   { return kvEntry{}, nil }

// kv is a behavior the tests serve. It is defined but not registered, so
// nothing else sees it.
var kv = behavior.Define("kv", "Key values", behavior.Node, behavior.API,
	behavior.Read("list", kvList),
	behavior.Read("get", kvGet).Key("key"),
	behavior.Write("create", kvCreate),
	behavior.Write("update", kvTouch).Key("key"),
)

// token signs a token with perms.
func token(
	perms ...string,
) string {
	t, _ := authtoken.New(slog.Default()).Generate(signingKey, []string{"read"}, "test", perms)

	return t
}

// send sends one request to e.
func send(
	e *echo.Echo,
	method string,
	path string,
	tok string,
) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, http.NoBody)
	if tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	return rec
}

type HandlerPublicTestSuite struct {
	suite.Suite
}

func (s *HandlerPublicTestSuite) TearDownSubTest() {
	apibehavior.ResetRouters()
}

func (s *HandlerPublicTestSuite) TestHandler() {
	// ok answers every request it is given, so a test sees whether the
	// token check let the request through.
	ok := func(c *echo.Context) error { return c.JSON(http.StatusOK, map[string]string{}) }
	routes := func(_ *apibehavior.Server) []apibehavior.Route {
		return []apibehavior.Route{{
			Method:     http.MethodGet,
			Path:       "/api/next/node/:hostname/kv",
			Permission: "kv:read",
			Handler:    ok,
		}}
	}

	tests := []struct {
		name     string
		routers  []func(s *apibehavior.Server) []apibehavior.Route
		path     string
		token    string
		wantCode int
	}{
		{
			name:     "when the token grants the route's permission, the handler runs",
			routers:  []func(s *apibehavior.Server) []apibehavior.Route{routes},
			path:     "/api/next/node/server1/kv",
			token:    token("kv:read"),
			wantCode: http.StatusOK,
		},
		{
			name:     "when there is no token, it is a 401",
			routers:  []func(s *apibehavior.Server) []apibehavior.Route{routes},
			path:     "/api/next/node/server1/kv",
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "when the token lacks the route's permission, it is a 403",
			routers:  []func(s *apibehavior.Server) []apibehavior.Route{routes},
			path:     "/api/next/node/server1/kv",
			token:    token("kv:write"),
			wantCode: http.StatusForbidden,
		},
		{
			name:     "when no behavior registered the route, it is not served",
			path:     "/api/next/node/server1/kv",
			token:    token("kv:read"),
			wantCode: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			apibehavior.SetRouters(tc.routers)
			jobs := jobmocks.NewMockJobClient(gomock.NewController(s.T()))
			a := api.New(config.Config{}, slog.Default())
			for _, register := range apibehavior.Handler(slog.Default(), jobs, signingKey, nil) {
				register(a.Echo)
			}

			rec := send(a.Echo, http.MethodGet, tc.path, tc.token)

			s.Equal(tc.wantCode, rec.Code, rec.Body.String())
			if tc.wantCode == http.StatusUnauthorized {
				var body map[string]string
				s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &body))
				s.NotEmpty(body["error"])
			}
		})
	}
}

func (s *HandlerPublicTestSuite) TestRegister() {
	tests := []struct {
		name string
	}{
		{name: "when a gen package registers its routes, Handler mounts them"},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			apibehavior.SetRouters(nil)
			apibehavior.Register(func(_ *apibehavior.Server) []apibehavior.Route {
				return []apibehavior.Route{{
					Method:     http.MethodGet,
					Path:       "/api/next/node/:hostname/kv",
					Permission: "kv:read",
					Handler:    func(c *echo.Context) error { return c.NoContent(http.StatusNoContent) },
				}}
			})
			a := api.New(config.Config{}, slog.Default())
			for _, register := range apibehavior.Handler(slog.Default(), nil, signingKey, nil) {
				register(a.Echo)
			}

			rec := send(a.Echo, http.MethodGet, "/api/next/node/server1/kv", token("kv:read"))

			s.Equal(http.StatusNoContent, rec.Code)
		})
	}
}

func (s *HandlerPublicTestSuite) TestPath() {
	tests := []struct {
		name   string
		method string
		want   string
	}{
		{
			name:   "when the method lists, the path has the target only",
			method: "list",
			want:   "/api/next/node/:hostname/kv",
		},
		{
			name:   "when the method is keyed, the key is a path parameter",
			method: "get",
			want:   "/api/next/node/:hostname/kv/:key",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, apibehavior.Path(kv, kv.Method(tc.method)))
		})
	}
}

func TestHandlerPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(HandlerPublicTestSuite))
}

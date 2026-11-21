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

package server_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/internal/controller/endpoint/server"
)

const signingKey = "test-signing-key-for-the-endpoint-server"

var (
	errMissing = errors.New("no such thing")
	errBusy    = errors.New("busy")
)

type deps struct {
	result string
	err    error
}

type getInput struct {
	Name string `json:"name" in:"path" validate:"required,max=5"`
}

type thing struct {
	Name string `json:"name"`
}

func get(
	_ context.Context,
	d deps,
	in getInput,
) (thing, error) {
	return thing{Name: in.Name + d.result}, d.err
}

func remove(
	_ context.Context,
	d deps,
	_ getInput,
) (struct{}, error) {
	return struct{}{}, d.err
}

func decodeGet(
	c *echo.Context,
) (getInput, string, bool) {
	var in getInput
	if msg, ok := server.Param(c, "name", &in.Name); !ok {
		return in, msg, false
	}
	if in.Name == "bad" {
		return in, "bad name", false
	}

	return in, "", true
}

var group = endpoint.Define("thing", "Things",
	endpoint.Uses[deps](),
	endpoint.Get[getInput, thing]("get").
		Handle(get).
		Errors(
			endpoint.NotFound(errMissing, "No such thing."),
			endpoint.Conflict(errBusy, "Busy."),
		),
	endpoint.Delete[getInput, struct{}]("remove").Handle(remove).Status(http.StatusNoContent),
	endpoint.Post[getInput, thing]("make").Sub("make").Handle(get).Public().
		Status(http.StatusCreated).Location("name"),
	endpoint.Get[getInput, thing]("open").Sub("open").Handle(get).Public().
		Answers(http.StatusServiceUnavailable, "Not ready."),
)

func token(
	perms ...string,
) string {
	t, _ := authtoken.New(slog.Default()).Generate(signingKey, []string{"read"}, "test", perms)

	return t
}

type ServerPublicTestSuite struct {
	suite.Suite
}

// serve mounts the group's routes on a new echo with d, and runs one
// request.
func (s *ServerPublicTestSuite) serve(
	d deps,
	method string,
	path string,
	tok string,
) *httptest.ResponseRecorder {
	srv := server.New(slog.Default(), signingKey, nil, d)
	e := echo.New()
	routes := make([]server.Route, 0, len(group.Methods))
	for _, m := range group.Methods {
		r := server.Route{
			Method:     m.Verb,
			Path:       server.Path(group, m),
			Permission: group.Permission(m),
		}
		switch m.Name {
		case "remove":
			r.Handler = server.Serve(srv, group, m.Name, remove, decodeGet)
		default:
			r.Handler = server.Serve(srv, group, m.Name, get, decodeGet)
		}
		routes = append(routes, r)
	}
	srv.Mount(e, routes)

	req := httptest.NewRequest(method, path, nil)
	if tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	return rec
}

func (s *ServerPublicTestSuite) TestServe() {
	tests := []struct {
		name     string
		deps     deps
		method   string
		path     string
		token    string
		wantCode int
		wantBody string
		wantLoc  string
	}{
		{
			name:     "when a method creates a resource, the Location header names it",
			deps:     deps{result: "7"},
			method:   http.MethodPost,
			path:     "/api/thing/make/web",
			wantCode: http.StatusCreated,
			wantBody: `{"name":"web7"}`,
			wantLoc:  "/api/thing/web7",
		},
		{
			name:     "when the token grants the permission, the handler's output is the body",
			deps:     deps{result: "!"},
			method:   http.MethodGet,
			path:     "/api/thing/web",
			token:    token("thing:read"),
			wantCode: http.StatusOK,
			wantBody: `{"name":"web!"}`,
		},
		{
			name:     "when there is no token, it is a 401",
			method:   http.MethodGet,
			path:     "/api/thing/web",
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "when the token lacks the permission, it is a 403",
			method:   http.MethodGet,
			path:     "/api/thing/web",
			token:    token("thing:write"),
			wantCode: http.StatusForbidden,
		},
		{
			name:     "when the input does not decode, it is a 400 with the reason",
			method:   http.MethodGet,
			path:     "/api/thing/bad",
			token:    token("thing:read"),
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":"bad name","reason":"invalid_request"}`,
		},
		{
			name:     "when the handler answers a status with its output, the output is the body",
			deps:     deps{result: "?", err: endpoint.Answer(http.StatusServiceUnavailable)},
			method:   http.MethodGet,
			path:     "/api/thing/open/web",
			wantCode: http.StatusServiceUnavailable,
			wantBody: `{"name":"web?"}`,
		},
		{
			name:     "when the handler returns a mapped error, it is that status",
			deps:     deps{err: errMissing},
			method:   http.MethodGet,
			path:     "/api/thing/web",
			token:    token("thing:read"),
			wantCode: http.StatusNotFound,
			wantBody: `{"error":"no such thing","reason":"not_found"}`,
		},
		{
			name:     "when the handler wraps a mapped error, errors.Is finds it",
			deps:     deps{err: errors.Join(errors.New("lookup"), errBusy)},
			method:   http.MethodGet,
			path:     "/api/thing/web",
			token:    token("thing:read"),
			wantCode: http.StatusConflict,
		},
		{
			name:     "when the handler returns any other error, it is a 500",
			deps:     deps{err: errors.New("nats: timeout")},
			method:   http.MethodGet,
			path:     "/api/thing/web",
			token:    token("thing:read"),
			wantCode: http.StatusInternalServerError,
			wantBody: `{"error":"nats: timeout","reason":"server_error"}`,
		},
		{
			name:     "when the handler returns an Internal error, the caller gets only its message",
			deps:     deps{err: endpoint.Internal("could not read", errors.New("dial 10.0.0.9"))},
			method:   http.MethodGet,
			path:     "/api/thing/web",
			token:    token("thing:read"),
			wantCode: http.StatusInternalServerError,
			wantBody: `{"error":"could not read","reason":"server_error"}`,
		},
		{
			name:     "when an Internal error has no cause, its message is logged",
			deps:     deps{err: endpoint.Internal("could not read", nil)},
			method:   http.MethodGet,
			path:     "/api/thing/web",
			token:    token("thing:read"),
			wantCode: http.StatusInternalServerError,
			wantBody: `{"error":"could not read","reason":"server_error"}`,
		},
		{
			name:     "when the success status is 204, there is no body",
			method:   http.MethodDelete,
			path:     "/api/thing/web",
			token:    token("thing:write"),
			wantCode: http.StatusNoContent,
		},
		{
			name:     "when the method is public, no token is needed",
			method:   http.MethodGet,
			path:     "/api/thing/open/web",
			wantCode: http.StatusOK,
			wantBody: `{"name":"web"}`,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			rec := s.serve(tc.deps, tc.method, tc.path, tc.token)

			s.Equal(tc.wantCode, rec.Code, rec.Body.String())
			if tc.wantBody != "" {
				s.JSONEq(tc.wantBody, rec.Body.String())
			}
			if tc.wantCode == http.StatusNoContent {
				s.Empty(rec.Body.String())
			}
			s.Equal(tc.wantLoc, rec.Header().Get("Location"))
		})
	}
}

func (s *ServerPublicTestSuite) TestDeps() {
	srv := server.New(slog.Default(), signingKey, nil, deps{result: "x"})

	s.Equal("x", server.Deps[deps](srv).result)
	s.PanicsWithValue("endpoint: no string given to the server", func() {
		server.Deps[string](srv)
	})
	s.PanicsWithValue("endpoint: no server_test.deps given to the server", func() {
		server.Deps[deps](server.New(slog.Default(), signingKey, nil))
	})
	s.Equal(struct{}{}, server.Deps[struct{}](srv), "an empty struct needs no wiring")
}

func (s *ServerPublicTestSuite) TestHandler() {
	called := 0
	server.Register(func(_ *server.Server) []server.Route {
		called++

		return []server.Route{{
			Method:  http.MethodGet,
			Path:    "/api/zzhandler",
			Handler: func(c *echo.Context) error { return c.String(http.StatusOK, "ok") },
		}}
	})

	e := echo.New()
	for _, register := range server.Handler(slog.Default(), signingKey, nil) {
		register(e)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/zzhandler", nil))

	s.Equal(1, called)
	s.Equal(http.StatusOK, rec.Code)
	s.True(strings.HasPrefix(rec.Body.String(), "ok"))
}

func TestServerPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ServerPublicTestSuite))
}

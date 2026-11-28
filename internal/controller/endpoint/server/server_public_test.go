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

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/internal/controller/endpoint/server"
	"github.com/osapi-io/osapi/internal/secret"
	"github.com/osapi-io/osapi/internal/secret/vault"
	"github.com/osapi-io/osapi/internal/wiring"
)

var signingKey = authtoken.Key{ID: "test", Secret: "test-signing-key-for-the-endpoint-server"}

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

// pair is an output whose ID is two fields, such as a vault and a key.
type pair struct {
	Vault string `json:"vault"`
	Key   string `json:"key"`
}

func makePair(
	_ context.Context,
	d deps,
	in getInput,
) (pair, error) {
	return pair{Vault: "default", Key: in.Name + d.result}, d.err
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
	endpoint.Post[getInput, pair]("pair").Sub("pair").Handle(makePair).Public().
		Status(http.StatusCreated).Location("vault", "key"),
	endpoint.Get[getInput, thing]("open").Sub("open").Handle(get).Public().
		Answers(http.StatusServiceUnavailable, "Not ready."),
)

func token(
	perms ...string,
) string {
	t, _ := authtoken.New().
		Generate(signingKey, []types.Role{types.RoleRead}, "test", perms, nil)

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
	srv := server.New(slog.Default(), authtoken.Keys{signingKey}, nil, d)
	e := echo.New()
	routes := make([]server.Route, 0, len(group.Methods))
	for _, m := range group.Methods {
		r := server.Route{
			Method:     m.HTTPMethod(),
			Path:       server.Path(group, m),
			Permission: string(group.Permission(m)),
		}
		switch m.Name {
		case "remove":
			r.Handler = server.Serve(srv, group, m.Name, remove, decodeGet)
		case "pair":
			r.Handler = server.Serve(srv, group, m.Name, makePair, decodeGet)
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
			name:     "when the ID of a resource is two fields, the Location header joins them, escaped",
			deps:     deps{result: " x"},
			method:   http.MethodPost,
			path:     "/api/thing/pair/web",
			wantCode: http.StatusCreated,
			wantBody: `{"vault":"default","key":"web x"}`,
			wantLoc:  "/api/thing/default/web%20x",
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

type loginInput struct {
	Hostname string                  `json:"-"`
	Password secret.Value            `json:"password"      validate:"required,no_linebreak"`
	Env      map[string]secret.Value `json:"env,omitempty"`
}

// fakeVault is a vault of keys, with an error for each Get when err is
// set.
type fakeVault struct {
	name string
	keys map[string]string
	err  error
}

func (v fakeVault) Name() string { return v.name }

func (v fakeVault) Get(
	_ context.Context,
	key string,
) ([]byte, vault.Meta, error) {
	if v.err != nil {
		return nil, vault.Meta{}, v.err
	}
	value, ok := v.keys[key]
	if !ok {
		return nil, vault.Meta{}, vault.ErrNotFound
	}

	return []byte(value), vault.Meta{Key: key, Version: 1}, nil
}

func (fakeVault) Put(context.Context, string, []byte) (vault.Meta, error) {
	return vault.Meta{}, nil
}

func (fakeVault) Delete(context.Context, string) error { return nil }

func (fakeVault) List(context.Context) ([]vault.Meta, error) { return nil, nil }

// fakeVaults are vaults by name, each with its spec.
type fakeVaults map[string]struct {
	spec  vault.Spec
	vault fakeVault
}

func (f fakeVaults) Get(
	name string,
) (vault.Vault, bool) {
	v, ok := f[name]

	return v.vault, ok
}

func (f fakeVaults) Spec(
	name string,
) (vault.Spec, bool) {
	v, ok := f[name]

	return v.spec, ok
}

// TestServeResolvesAReference: the route resolves each reference from its
// vault, after the limits of use, and gives the handler the value. Each
// refusal names the reference, never a value.
func (s *ServerPublicTestSuite) TestServeResolvesAReference() {
	g := endpoint.Define("login", "Logins",
		endpoint.Uses[deps](),
		endpoint.Post[loginInput, thing]("make").Public().
			Handle(func(context.Context, deps, loginInput) (thing, error) { return thing{}, nil }),
	)
	vaults := fakeVaults{
		"default": {
			spec: vault.Spec{Name: "default", Type: "nats"},
			vault: fakeVault{
				name: "default",
				keys: map[string]string{"pw": "hunter2", "nl": "a\nb"},
			},
		},
		"prod": {
			spec: vault.Spec{
				Name:    "prod",
				Type:    "nats",
				Targets: []string{"web-01", "group:prod"},
			},
			vault: fakeVault{name: "prod", keys: map[string]string{"pw": "prod-pw"}},
		},
		"down": {
			spec:  vault.Spec{Name: "down", Type: "nats"},
			vault: fakeVault{name: "down", err: errors.New("nats: timeout")},
		},
	}
	user := token("login:write", "secret:use")
	tests := []struct {
		name     string
		vaults   server.Vaults
		host     string
		token    string
		password secret.Value
		env      map[string]secret.Value
		wantCode int
		wantBody string
		wantGot  loginInput
	}{
		{
			name:     "when the secret is a reference, the handler gets its value",
			vaults:   vaults,
			host:     "web-01",
			token:    user,
			password: "secret://default/pw",
			env:      map[string]secret.Value{"A": "secret://default/pw", "B": "literal"},
			wantCode: http.StatusOK,
			wantGot: loginInput{
				Hostname: "web-01", Password: "hunter2",
				Env: map[string]secret.Value{"A": "hunter2", "B": "literal"},
			},
		},
		{
			name:     "when the secret is a value, the handler gets it",
			host:     "web-01",
			token:    token("login:write"),
			password: "hunter2",
			wantCode: http.StatusOK,
			wantGot:  loginInput{Hostname: "web-01", Password: "hunter2"},
		},
		{
			name:     "when the controller has no vault, it is a 400 that names the vault",
			host:     "web-01",
			token:    user,
			password: "secret://default/pw",
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":"secret://default/pw: the vault \"default\" does not exist","reason":"invalid_request"}`,
		},
		{
			name:     "when the vault does not exist, it is a 400 that names it",
			vaults:   vaults,
			host:     "web-01",
			token:    user,
			password: "secret://other/pw",
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":"secret://other/pw: the vault \"other\" does not exist","reason":"invalid_request"}`,
		},
		{
			name:     "when the reference has no key, it is a 400 that gives the form",
			vaults:   vaults,
			host:     "web-01",
			token:    user,
			password: "secret://default",
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":"secret://default: a reference is secret://<vault>/<key>","reason":"invalid_request"}`,
		},
		{
			name:     "when the key is not a name that a vault takes, it is a 400",
			vaults:   vaults,
			host:     "web-01",
			token:    user,
			password: "secret://default/a.b",
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":"secret://default/a.b: key \"a.b\": a key is 1 to 128 letters, digits, - or _, such as registry-token","reason":"invalid_request"}`,
		},
		{
			name:     "when the token does not have secret:use, it is a 403",
			vaults:   vaults,
			host:     "web-01",
			token:    token("login:write"),
			password: "secret://default/pw",
			wantCode: http.StatusForbidden,
			wantBody: `{"error":"secret://default/pw: the token does not have the permission secret:use","reason":"forbidden"}`,
		},
		{
			name:     "when the token does not have secret:use, a key that is not there is a 403 too",
			vaults:   vaults,
			host:     "web-01",
			token:    token("login:write"),
			password: "secret://default/none",
			wantCode: http.StatusForbidden,
			wantBody: `{"error":"secret://default/none: the token does not have the permission secret:use","reason":"forbidden"}`,
		},
		{
			name:     "when the host is in the targets of the vault, the handler gets the value",
			vaults:   vaults,
			host:     "web-01",
			token:    user,
			password: "secret://prod/pw",
			wantCode: http.StatusOK,
			wantGot:  loginInput{Hostname: "web-01", Password: "prod-pw"},
		},
		{
			name:     "when the host is outside the targets of the vault, it is a 403",
			vaults:   vaults,
			host:     "db-01",
			token:    user,
			password: "secret://prod/pw",
			wantCode: http.StatusForbidden,
			wantBody: `{"error":"secret://prod/pw: the vault \"prod\" gives its secrets only to web-01, group:prod, not to \"db-01\"","reason":"forbidden"}`,
		},
		{
			name:     "when the target is _all and the vault has targets, it is a 403",
			vaults:   vaults,
			host:     "_all",
			token:    user,
			password: "secret://prod/pw",
			wantCode: http.StatusForbidden,
			wantBody: `{"error":"secret://prod/pw: the vault \"prod\" gives its secrets only to web-01, group:prod, not to \"_all\"","reason":"forbidden"}`,
		},
		{
			name:     "when the request names no host and the vault has targets, it is a 403",
			vaults:   vaults,
			token:    user,
			password: "secret://prod/pw",
			wantCode: http.StatusForbidden,
			wantBody: `{"error":"secret://prod/pw: the vault \"prod\" gives its secrets only to web-01, group:prod, and the request names no host","reason":"forbidden"}`,
		},
		{
			name:     "when the key is not there, it is a 400 that names it",
			vaults:   vaults,
			host:     "web-01",
			token:    user,
			password: "secret://default/none",
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":"secret://default/none: the vault \"default\" has no key \"none\"","reason":"invalid_request"}`,
		},
		{
			name:     "when the vault does not answer, it is a 500 that does not give the cause",
			vaults:   vaults,
			host:     "web-01",
			token:    user,
			password: "secret://down/pw",
			wantCode: http.StatusInternalServerError,
			wantBody: `{"error":"secret://down/pw: the vault could not be read","reason":"server_error"}`,
		},
		{
			name:     "when the value fails a validate tag of its field, it is a 400 without the value",
			vaults:   vaults,
			host:     "web-01",
			token:    user,
			password: "secret://default/nl",
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			given := []any{deps{}}
			if tc.vaults != nil {
				given = append(given, server.Secrets{Vaults: tc.vaults})
			}
			srv := server.New(slog.Default(), authtoken.Keys{signingKey}, nil, given...)
			e := echo.New()
			m := g.Method("make")
			route, path := "/api/login/:hostname", "/api/login/"+tc.host
			if tc.host == "" {
				route, path = "/api/login", "/api/login"
			}
			var got loginInput
			srv.Mount(e, []server.Route{{
				Method:     m.HTTPMethod(),
				Path:       route,
				Permission: "login:write",
				Handler: server.Serve(srv, g, "make",
					func(_ context.Context, _ deps, in loginInput) (thing, error) {
						got = in

						return thing{}, nil
					},
					func(c *echo.Context) (loginInput, string, bool) {
						return loginInput{
							Hostname: c.Param("hostname"), Password: tc.password, Env: tc.env,
						}, "", true
					}),
			}})
			req := httptest.NewRequest(m.HTTPMethod(), path, nil)
			req.Header.Set("Authorization", "Bearer "+tc.token)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			s.Equal(tc.wantCode, rec.Code, rec.Body.String())
			if tc.wantBody != "" {
				s.JSONEq(tc.wantBody, rec.Body.String())
			}
			s.NotContains(rec.Body.String(), "hunter2", "a refusal never gives a value")
			s.NotContains(rec.Body.String(), "a\\nb", "a refusal never gives a value")
			s.Equal(tc.wantGot, got)
		})
	}
}

func (s *ServerPublicTestSuite) TestRefusal() {
	s.Equal("secret://v/k: gone", server.ExportRefusal(errors.New("secret://v/k: gone")))
}

func (s *ServerPublicTestSuite) TestDeps() {
	srv := server.New(slog.Default(), authtoken.Keys{signingKey}, nil, deps{result: "x"})

	s.Equal("x", server.Deps[deps](srv).result)
	_ = wiring.Take()
	s.Empty(server.Deps[string](srv))
	s.EqualError(wiring.Take(), "endpoint: no string given to the server")
	server.Deps[deps](server.New(slog.Default(), authtoken.Keys{signingKey}, nil))
	s.EqualError(wiring.Take(), "endpoint: no server_test.deps given to the server")
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
	for _, register := range server.Handler(slog.Default(), authtoken.Keys{signingKey}, nil) {
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

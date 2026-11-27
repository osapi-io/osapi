//go:build integration

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

package integration_test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// AuthorizationSuite sends each route of the API contract twice: with no
// token it must be 401, and with a token that lacks the route's permission
// 403, with the error body. Neither request changes anything, so each write
// route runs too, with no write flag. It runs against the suite's server,
// or a fixture's against a deployment, as SecurityDefaultsSuite does.
type AuthorizationSuite struct {
	suite.Suite

	base string
	dir  string
}

func (s *AuthorizationSuite) SetupSuite() {
	s.dir = tempDir
	port := apiPort
	if deployment() {
		f := startFixture(s.T(), true)
		s.dir, port = f.dir, f.apiPort
	}
	s.base = fmt.Sprintf("https://127.0.0.1:%d", port)
}

// route is one operation of the contract and the permission it needs.
type route struct {
	method     string
	path       string
	permission string
}

// opLine is an operation of internal/apicontract/testdata/joined.golden.
var opLine = regexp.MustCompile(`^op (\S+) (\S+) .* auth=BearerAuth:(\S+) `)

// routes is each operation of the contract that needs a permission.
func (s *AuthorizationSuite) routes() []route {
	f, err := os.Open(
		filepath.Join(repoRoot(), "internal", "apicontract", "testdata", "joined.golden"),
	)
	s.Require().NoError(err)
	defer f.Close()
	var out []route
	scan := bufio.NewScanner(f)
	scan.Buffer(make([]byte, 0, 1<<16), 1<<20)
	for scan.Scan() {
		if m := opLine.FindStringSubmatch(scan.Text()); m != nil {
			out = append(out, route{method: m[1], path: m[2], permission: m[3]})
		}
	}
	s.Require().NoError(scan.Err())

	return out
}

// pathParam fills each path parameter of a route with a value that names
// nothing: a refused request must not reach it.
var pathParam = regexp.MustCompile(`\{([a-z_]+)\}`)

func (s *AuthorizationSuite) send(
	r route,
	token string,
) (int, []byte) {
	path := pathParam.ReplaceAllStringFunc(r.path, func(m string) string {
		if m == "{id}" || strings.HasSuffix(m, "_id}") {
			return "00000000-0000-4000-8000-000000000000"
		}

		return "integration-nosuch"
	})
	body := ""
	if r.method != http.MethodGet && r.method != http.MethodDelete {
		body = "{}"
	}

	return s.sendBody(route{method: r.method, path: path}, token, body)
}

// sendBody sends r with token and body, "" for none, and returns the
// status and the body of the answer.
func (s *AuthorizationSuite) sendBody(
	r route,
	token string,
	body string,
) (int, []byte) {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req, err := http.NewRequest(r.method, s.base+r.path, reader)
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := apiClientFor(filepath.Join(s.dir, "pki", "controller", "api.crt")).Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	return resp.StatusCode, data
}

// TestEachRouteRefuses: each route needs a token, and the permission that
// the contract names, and its refusal is an ErrorResponse.
func (s *AuthorizationSuite) TestEachRouteRefuses() {
	routes := s.routes()
	s.Require().Greater(len(routes), 50, "the contract has the routes")
	for _, r := range routes {
		// A permission that each built-in role grants, such as catalog:read,
		// has no role to refuse: its route is checked for the 401 only.
		role, refusable := roleWithout(authtoken.Permission(r.permission))
		token, err := authtoken.New(nil).
			Generate(signingKey, []types.Role{role}, "authz@test", nil, nil)
		if refusable {
			s.Require().NoError(err)
		}
		s.Run(r.method+" "+r.path, func() {
			code, _ := s.send(r, "")
			s.Equal(http.StatusUnauthorized, code, "no token")
			if !refusable {
				return
			}
			code, data := s.send(r, token)
			s.Equal(http.StatusForbidden, code, "a token without %s: %s", r.permission, data)
			var e struct {
				Error string `json:"error"`
			}
			s.NoError(json.Unmarshal(data, &e), "the refusal is JSON: %s", data)
			s.NotEmpty(e.Error, "the refusal says why")
		})
	}
}

// roleWithout is a built-in role that does not grant p.
func roleWithout(
	p authtoken.Permission,
) (types.Role, bool) {
	for _, role := range []types.Role{
		types.RoleObserver, types.RoleAuditor, types.RoleRead, types.RoleAutomation, types.RoleOperator,
	} {
		perms, ok := authtoken.RolePermissions(role)
		if ok && !slices.Contains(perms, p) {
			return role, true
		}
	}

	return "", false
}

func TestAuthorizationSuite(t *testing.T) {
	suite.Run(t, new(AuthorizationSuite))
}

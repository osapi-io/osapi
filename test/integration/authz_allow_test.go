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
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// roleWith is the built-in role with the fewest permissions that grants p.
func roleWith(
	p authtoken.Permission,
) (types.Role, bool) {
	for _, role := range []types.Role{
		types.RoleObserver, types.RoleAuditor, types.RoleRead,
		types.RoleAutomation, types.RoleOperator, types.RoleAdmin,
	} {
		if perms, ok := authtoken.RolePermissions(role); ok && slices.Contains(perms, p) {
			return role, true
		}
	}

	return "", false
}

// TestEachReadAllows: each GET route with no path parameter answers with no
// refusal to a token of the smallest built-in role that grants its
// permission. A route with a path parameter needs a thing that the suite has
// not made, so the read-backs of the examples cover it.
func (s *AuthorizationSuite) TestEachReadAllows() {
	count := 0
	for _, r := range s.routes() {
		if r.method != http.MethodGet || strings.Contains(r.path, "{") {
			continue
		}
		role, ok := roleWith(authtoken.Permission(r.permission))
		s.Require().True(ok, "%s %s: no built-in role grants %s", r.method, r.path, r.permission)
		token, err := authtoken.New(slog.New(slog.DiscardHandler)).
			Generate(signingKey, []types.Role{role}, "authz-allow@test", nil, nil)
		s.Require().NoError(err)
		count++
		s.Run(r.method+" "+r.path, func() {
			code, data := s.send(r, token)
			s.Less(code, 300, "a %s token, which grants %s: %d %s", role, r.permission, code, data)
		})
	}
	s.Greater(count, 10, "the contract has the reads")
}

// token signs claims with secret and the kid, as Generate does: a test
// makes a token that the controller must refuse.
func (s *AuthorizationSuite) token(
	kid string,
	secret string,
	expires time.Time,
) string {
	claims := authtoken.CustomClaims{
		Roles: []types.Role{types.RoleAdmin},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "osapi", Subject: "authz-bad@test",
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			ExpiresAt: jwt.NewNumericDate(expires),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	if kid != "" {
		t.Header["kid"] = kid
	}
	out, err := t.SignedString([]byte(secret))
	s.Require().NoError(err)

	return out
}

// TestBadTokens: the controller refuses each token that is not a good
// token of its keys with a 401, also one with the admin role.
func (s *AuthorizationSuite) TestBadTokens() {
	read := route{method: http.MethodGet, path: "/api/agent", permission: "agent:read"}
	later := time.Now().Add(time.Hour)
	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "an expired token",
			token: s.token(signingKey.ID, signingKey.Secret, time.Now().Add(-time.Minute)),
		},
		{name: "a token that is not a JWT", token: "not-a-jwt"},
		{
			name:  "a token whose kid names no key of the config",
			token: s.token("integration-nosuch", signingKey.Secret, later),
		},
		{name: "a token with no kid", token: s.token("", signingKey.Secret, later)},
		{
			name:  "a token with the config's kid that another key signed",
			token: s.token(signingKey.ID, strings.Repeat("0", 64), later),
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			code, data := s.send(read, tt.token)
			s.Equal(http.StatusUnauthorized, code, "%s", data)
		})
	}
	s.Run("a good admin token, the control", func() {
		code, data := s.send(read, s.token(signingKey.ID, signingKey.Secret, later))
		s.Equal(http.StatusOK, code, "%s", data)
	})
}

// TestPermissionsToken: a token's permissions, as an identity provider
// sets them, replace the permissions of its roles: a token of the admin
// role with agent:read grants agent:read and no other.
func (s *AuthorizationSuite) TestPermissionsToken() {
	token, err := authtoken.New(slog.New(slog.DiscardHandler)).Generate(
		signingKey, []types.Role{types.RoleAdmin}, "authz-perms@test", []string{"agent:read"}, nil,
	)
	s.Require().NoError(err)
	tests := []struct {
		name string
		r    route
		want int
	}{
		{
			name: "the permission that it has",
			r:    route{method: http.MethodGet, path: "/api/agent"},
			want: http.StatusOK,
		},
		{
			name: "a permission that it has not",
			r:    route{method: http.MethodGet, path: "/api/job"},
			want: http.StatusForbidden,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			code, data := s.send(tt.r, token)
			s.Equal(tt.want, code, "%s", data)
		})
	}
}

// TestAuditOfAWrite: a write is in the audit log with the user and the
// route, and the log has no secret value that the write sent.
func (s *AuthorizationSuite) TestAuditOfAWrite() {
	if !runWrites {
		s.T().Skip("a write: set OSAPI_INTEGRATION_WRITES=1")
	}
	user := fmt.Sprintf("authz-audit-%d@test", time.Now().UnixNano())
	admin, err := authtoken.New(slog.New(slog.DiscardHandler)).
		Generate(signingKey, []types.Role{types.RoleAdmin}, user, nil, nil)
	s.Require().NoError(err)
	key := fmt.Sprintf("authz-%d", time.Now().UnixNano())
	value := "authz-value-" + key
	put := route{method: http.MethodPost, path: "/api/secret/default"}
	code, data := s.sendBody(put, admin, `{"key":"`+key+`","value":"`+value+`"}`)
	s.Require().Less(code, 300, "%s", data)
	defer func() {
		_, _ = s.sendBody(
			route{method: http.MethodDelete, path: put.path + "/" + key},
			admin,
			"",
		)
	}()

	var log struct {
		Items []map[string]any `json:"items"`
	}
	list := route{method: http.MethodGet, path: "/api/audit?user=" + user}
	deadline := time.Now().Add(15 * time.Second)
	for {
		code, data = s.send(list, admin)
		s.Require().Equal(http.StatusOK, code, "%s", data)
		s.Require().NoError(json.Unmarshal(data, &log), "%s", data)
		if len(log.Items) > 0 || time.Now().After(deadline) {
			break
		}
		time.Sleep(time.Second)
	}
	s.Require().NotEmpty(log.Items, "the write is in the audit log")
	found := false
	for _, e := range log.Items {
		if e["method"] == http.MethodPost && e["path"] == put.path {
			found = true
			s.Equal(user, e["user"])
		}
	}
	s.True(found, "an entry has POST %s", put.path)
	s.NotContains(string(data), value, "the audit log has no secret value")
}

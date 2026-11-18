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
package facts_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/controller/endpoint/server"
	"github.com/osapi-io/osapi/internal/controller/endpoints/facts"
	"github.com/osapi-io/osapi/internal/controller/endpoints/facts/gen"
	factskeys "github.com/osapi-io/osapi/internal/facts"
)

const signingKey = "test-signing-key-for-the-facts-group"

type MethodsPublicTestSuite struct {
	suite.Suite
}

func (s *MethodsPublicTestSuite) TestKeys() {
	got, err := facts.Keys(context.Background(), facts.Deps{}, facts.KeysInput{})
	s.Require().NoError(err)

	s.Len(got.Items, len(factskeys.BuiltInKeys()))
	names := make([]string, 0, len(got.Items))
	for _, k := range got.Items {
		names = append(names, k.Key)
		s.True(k.Builtin, k.Key)
		s.NotEmpty(k.Description, k.Key)
	}
	s.Contains(names, factskeys.KeyInterfacePrimary)
	s.Contains(names, factskeys.KeyHostname)
}

func token(
	roles []string,
	perms ...string,
) string {
	t, _ := authtoken.New(slog.Default()).Generate(signingKey, roles, "test", perms)

	return t
}

// TestRoutes runs requests through the generated route, with the auth of
// the endpoint runtime.
func (s *MethodsPublicTestSuite) TestRoutes() {
	tests := []struct {
		name     string
		token    string
		wantCode int
	}{
		{
			name:     "when the token has agent:read, it is a 200",
			token:    token([]string{"read"}, "agent:read"),
			wantCode: http.StatusOK,
		},
		{
			name:     "when the role is admin, it is a 200",
			token:    token([]string{"admin"}),
			wantCode: http.StatusOK,
		},
		{
			name:     "when there is no token, it is a 401",
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "when the token lacks agent:read, it is a 403",
			token:    token([]string{"read"}, "job:read"),
			wantCode: http.StatusForbidden,
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			srv := server.New(slog.Default(), signingKey, nil, facts.Deps{})
			e := echo.New()
			srv.Mount(e, gen.Routes(srv))
			req := httptest.NewRequest(http.MethodGet, "/api/facts/keys", nil)
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			s.Equal(tc.wantCode, rec.Code, rec.Body.String())
		})
	}
}

func TestMethodsPublicTestSuite(t *testing.T) {
	suite.Run(t, new(MethodsPublicTestSuite))
}

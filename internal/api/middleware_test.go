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

package api

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	strictecho "github.com/oapi-codegen/runtime/strictmiddleware/echo"
	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/authtoken"
)

const testSigningKey = "test-signing-key-for-middleware"

type MiddlewareTestSuite struct {
	suite.Suite

	tokenManager *authtoken.Token
}

func (s *MiddlewareTestSuite) SetupSuite() {
	logger := slog.Default()
	s.tokenManager = authtoken.New(logger)
}

func (s *MiddlewareTestSuite) generateToken(roles []string) string {
	token, err := s.tokenManager.Generate(testSigningKey, roles, "test-subject")
	s.Require().NoError(err)

	return token
}

func (s *MiddlewareTestSuite) TestHasScope() {
	tests := []struct {
		name          string
		roles         []string
		requiredScope string
		expected      bool
	}{
		{
			name:          "admin has read scope",
			roles:         []string{"admin"},
			requiredScope: "read",
			expected:      true,
		},
		{
			name:          "admin has write scope",
			roles:         []string{"admin"},
			requiredScope: "write",
			expected:      true,
		},
		{
			name:          "admin has admin scope",
			roles:         []string{"admin"},
			requiredScope: "admin",
			expected:      true,
		},
		{
			name:          "write role has read scope",
			roles:         []string{"write"},
			requiredScope: "read",
			expected:      true,
		},
		{
			name:          "write role has write scope",
			roles:         []string{"write"},
			requiredScope: "write",
			expected:      true,
		},
		{
			name:          "write role does not have admin scope",
			roles:         []string{"write"},
			requiredScope: "admin",
			expected:      false,
		},
		{
			name:          "read role has read scope",
			roles:         []string{"read"},
			requiredScope: "read",
			expected:      true,
		},
		{
			name:          "read role does not have write scope",
			roles:         []string{"read"},
			requiredScope: "write",
			expected:      false,
		},
		{
			name:          "unknown role has no scopes",
			roles:         []string{"unknown"},
			requiredScope: "read",
			expected:      false,
		},
		{
			name:          "empty roles",
			roles:         []string{},
			requiredScope: "read",
			expected:      false,
		},
		{
			name:          "nil roles",
			roles:         nil,
			requiredScope: "read",
			expected:      false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := hasScope(tt.roles, tt.requiredScope)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *MiddlewareTestSuite) TestScopeMiddleware() {
	handlerCalled := false
	testHandler := strictecho.StrictEchoHandlerFunc(
		func(_ echo.Context, _ interface{}) (interface{}, error) {
			handlerCalled = true
			return "ok", nil
		},
	)

	contextKey := "BearerAuthScopes"

	tests := []struct {
		name            string
		authHeader      string
		requiredScopes  []string
		expectedStatus  int
		expectCalled    bool
		setupContextKey bool
	}{
		{
			name:            "no auth header returns 401",
			authHeader:      "",
			requiredScopes:  []string{"read"},
			expectedStatus:  http.StatusUnauthorized,
			expectCalled:    false,
			setupContextKey: true,
		},
		{
			name:            "non-bearer auth header returns 401",
			authHeader:      "Basic dXNlcjpwYXNz",
			requiredScopes:  []string{"read"},
			expectedStatus:  http.StatusUnauthorized,
			expectCalled:    false,
			setupContextKey: true,
		},
		{
			name:            "invalid token returns 401",
			authHeader:      "Bearer invalid-token-string",
			requiredScopes:  []string{"read"},
			expectedStatus:  http.StatusUnauthorized,
			expectCalled:    false,
			setupContextKey: true,
		},
		{
			name:            "valid token with sufficient scope calls handler",
			authHeader:      "", // set dynamically
			requiredScopes:  []string{"read"},
			expectedStatus:  http.StatusOK,
			expectCalled:    true,
			setupContextKey: true,
		},
		{
			name:            "valid token with insufficient scope returns 403",
			authHeader:      "", // set dynamically
			requiredScopes:  []string{"admin"},
			expectedStatus:  http.StatusForbidden,
			expectCalled:    false,
			setupContextKey: true,
		},
		{
			name:            "valid token with no required scopes calls handler",
			authHeader:      "", // set dynamically
			requiredScopes:  nil,
			expectedStatus:  http.StatusOK,
			expectCalled:    true,
			setupContextKey: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			handlerCalled = false

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			// Set auth header
			authHeader := tt.authHeader
			if authHeader == "" && tt.expectedStatus != http.StatusUnauthorized {
				// Generate a valid token with "read" role
				token := s.generateToken([]string{"read"})
				authHeader = "Bearer " + token
			}
			if authHeader != "" {
				req.Header.Set("Authorization", authHeader)
			}

			ctx := e.NewContext(req, rec)

			if tt.setupContextKey && tt.requiredScopes != nil {
				ctx.Set(contextKey, tt.requiredScopes)
			}

			wrapped := scopeMiddleware(testHandler, s.tokenManager, testSigningKey, contextKey)
			_, _ = wrapped(ctx, nil)

			s.Equal(tt.expectCalled, handlerCalled)
			if !tt.expectCalled {
				s.Equal(tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(MiddlewareTestSuite))
}

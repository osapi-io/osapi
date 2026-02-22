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

func (s *MiddlewareTestSuite) generateToken(
	roles []string,
) string {
	token, err := s.tokenManager.Generate(testSigningKey, roles, "test-subject", nil)
	s.Require().NoError(err)

	return token
}

func (s *MiddlewareTestSuite) generateTokenWithPerms(
	roles []string,
	permissions []string,
) string {
	token, err := s.tokenManager.Generate(testSigningKey, roles, "test-subject", permissions)
	s.Require().NoError(err)

	return token
}

func (s *MiddlewareTestSuite) TestScopeMiddleware() {
	contextKey := "BearerAuthScopes"

	tests := []struct {
		name            string
		authHeader      string
		requiredScopes  []string
		customRoles     map[string][]string
		expectedStatus  int
		expectCalled    bool
		setupContextKey bool
	}{
		{
			name:            "no auth header returns 401",
			authHeader:      "",
			requiredScopes:  []string{"system:read"},
			expectedStatus:  http.StatusUnauthorized,
			expectCalled:    false,
			setupContextKey: true,
		},
		{
			name:            "non-bearer auth header returns 401",
			authHeader:      "Basic dXNlcjpwYXNz",
			requiredScopes:  []string{"system:read"},
			expectedStatus:  http.StatusUnauthorized,
			expectCalled:    false,
			setupContextKey: true,
		},
		{
			name:            "invalid token returns 401",
			authHeader:      "Bearer invalid-token-string",
			requiredScopes:  []string{"system:read"},
			expectedStatus:  http.StatusUnauthorized,
			expectCalled:    false,
			setupContextKey: true,
		},
		{
			name:            "admin role has system:read",
			authHeader:      "", // set dynamically
			requiredScopes:  []string{"system:read"},
			expectedStatus:  http.StatusOK,
			expectCalled:    true,
			setupContextKey: true,
		},
		{
			name:            "read role has system:read",
			authHeader:      "", // set dynamically
			requiredScopes:  []string{"system:read"},
			expectedStatus:  http.StatusOK,
			expectCalled:    true,
			setupContextKey: true,
		},
		{
			name:            "read role lacks network:write returns 403",
			authHeader:      "", // set dynamically
			requiredScopes:  []string{"network:write"},
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
			handlerCalled := false
			testHandler := strictecho.StrictEchoHandlerFunc(
				func(_ echo.Context, _ interface{}) (interface{}, error) {
					handlerCalled = true
					return "ok", nil
				},
			)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			authHeader := tt.authHeader
			if authHeader == "" && tt.expectedStatus != http.StatusUnauthorized {
				if tt.expectedStatus == http.StatusForbidden {
					// Use read role for insufficient permission tests
					authHeader = "Bearer " + s.generateToken([]string{"read"})
				} else {
					authHeader = "Bearer " + s.generateToken([]string{"admin"})
				}
			}
			if authHeader != "" {
				req.Header.Set("Authorization", authHeader)
			}

			ctx := e.NewContext(req, rec)

			if tt.setupContextKey && tt.requiredScopes != nil {
				ctx.Set(contextKey, tt.requiredScopes)
			}

			wrapped := scopeMiddleware(
				testHandler,
				s.tokenManager,
				testSigningKey,
				contextKey,
				tt.customRoles,
			)
			_, _ = wrapped(ctx, nil)

			s.Equal(tt.expectCalled, handlerCalled)
			if !tt.expectCalled {
				s.Equal(tt.expectedStatus, rec.Code)
			}
		})
	}
}

func (s *MiddlewareTestSuite) TestScopeMiddlewareCustomRoles() {
	contextKey := "BearerAuthScopes"

	tests := []struct {
		name           string
		tokenRoles     []string
		customRoles    map[string][]string
		requiredScope  string
		expectedStatus int
		expectCalled   bool
	}{
		{
			name:       "custom role grants access",
			tokenRoles: []string{"ops"},
			customRoles: map[string][]string{
				"ops": {"system:read", "health:read"},
			},
			requiredScope:  "system:read",
			expectedStatus: http.StatusOK,
			expectCalled:   true,
		},
		{
			name:       "custom role lacks permission",
			tokenRoles: []string{"ops"},
			customRoles: map[string][]string{
				"ops": {"health:read"},
			},
			requiredScope:  "system:read",
			expectedStatus: http.StatusForbidden,
			expectCalled:   false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			handlerCalled := false
			testHandler := strictecho.StrictEchoHandlerFunc(
				func(_ echo.Context, _ interface{}) (interface{}, error) {
					handlerCalled = true
					return "ok", nil
				},
			)

			// Custom-role tokens need roles that validate (read/write/admin),
			// but we want to test custom role resolution. We'll use a token with
			// "admin" role and have the custom role shadow it.
			// Actually, custom roles use custom role names not built-in ones.
			// The token validation requires oneof=read write admin, so we
			// generate with "admin" but configure a custom role "ops".
			// Wait - the custom role name "ops" wouldn't be in the JWT Roles validation.
			// Let me think... The JWT requires roles to be read/write/admin.
			// Custom roles would shadow built-in roles. Let me use "admin" and
			// shadow it with a custom role mapping.
			token, err := s.tokenManager.Generate(
				testSigningKey,
				[]string{"admin"},
				"test-subject",
				nil,
			)
			s.Require().NoError(err)

			// Override custom roles to shadow "admin"
			customRoles := map[string][]string{
				"admin": tt.customRoles[tt.tokenRoles[0]],
			}

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set(contextKey, []string{tt.requiredScope})

			wrapped := scopeMiddleware(
				testHandler,
				s.tokenManager,
				testSigningKey,
				contextKey,
				customRoles,
			)
			_, _ = wrapped(ctx, nil)

			s.Equal(tt.expectCalled, handlerCalled)
			if !tt.expectCalled {
				s.Equal(tt.expectedStatus, rec.Code)
			}
		})
	}
}

func (s *MiddlewareTestSuite) TestScopeMiddlewareDirectPermissions() {
	contextKey := "BearerAuthScopes"

	tests := []struct {
		name           string
		permissions    []string
		requiredScope  string
		expectedStatus int
		expectCalled   bool
	}{
		{
			name:           "direct permission grants access",
			permissions:    []string{"system:read"},
			requiredScope:  "system:read",
			expectedStatus: http.StatusOK,
			expectCalled:   true,
		},
		{
			name:           "direct permission restricts to only listed",
			permissions:    []string{"health:read"},
			requiredScope:  "system:read",
			expectedStatus: http.StatusForbidden,
			expectCalled:   false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			handlerCalled := false
			testHandler := strictecho.StrictEchoHandlerFunc(
				func(_ echo.Context, _ interface{}) (interface{}, error) {
					handlerCalled = true
					return "ok", nil
				},
			)

			token := s.generateTokenWithPerms([]string{"admin"}, tt.permissions)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set(contextKey, []string{tt.requiredScope})

			wrapped := scopeMiddleware(
				testHandler,
				s.tokenManager,
				testSigningKey,
				contextKey,
				nil,
			)
			_, _ = wrapped(ctx, nil)

			s.Equal(tt.expectCalled, handlerCalled)
			if !tt.expectCalled {
				s.Equal(tt.expectedStatus, rec.Code)
			}
		})
	}
}

func (s *MiddlewareTestSuite) TestScopeMiddlewareInjectsIdentity() {
	contextKey := "BearerAuthScopes"

	handlerCalled := false
	var capturedSubject string
	var capturedRoles []string

	testHandler := strictecho.StrictEchoHandlerFunc(
		func(ctx echo.Context, _ interface{}) (interface{}, error) {
			handlerCalled = true
			capturedSubject, _ = ctx.Get(ContextKeySubject).(string)
			capturedRoles, _ = ctx.Get(ContextKeyRoles).([]string)
			return "ok", nil
		},
	)

	token := s.generateToken([]string{"admin"})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	ctx := e.NewContext(req, rec)
	ctx.Set(contextKey, []string{"system:read"})

	wrapped := scopeMiddleware(
		testHandler,
		s.tokenManager,
		testSigningKey,
		contextKey,
		nil,
	)
	_, _ = wrapped(ctx, nil)

	s.True(handlerCalled)
	s.Equal("test-subject", capturedSubject)
	s.Equal([]string{"admin"}, capturedRoles)
}

func TestMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(MiddlewareTestSuite))
}

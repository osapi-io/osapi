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

package api_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/controller/api"
	"github.com/osapi-io/osapi/internal/validation"
)

var testSigningKey = authtoken.Key{ID: "test", Secret: "test-signing-key-for-middleware"}

type MiddlewarePublicTestSuite struct {
	suite.Suite

	tokenManager *authtoken.Token
}

func (s *MiddlewarePublicTestSuite) SetupSuite() {
	logger := slog.Default()
	s.tokenManager = authtoken.New(logger)
}

func (s *MiddlewarePublicTestSuite) generateToken(
	roles []string,
) string {
	token, err := s.tokenManager.Generate(testSigningKey, roles, "test-subject", nil, nil)
	s.Require().NoError(err)

	return token
}

func (s *MiddlewarePublicTestSuite) generateTokenWithPerms(
	roles []string,
	permissions []string,
) string {
	token, err := s.tokenManager.Generate(testSigningKey, roles, "test-subject", permissions, nil)
	s.Require().NoError(err)

	return token
}

func (s *MiddlewarePublicTestSuite) TestScopeMiddleware() {
	contextKey := "BearerAuthScopes"

	tests := []struct {
		name            string
		authHeader      string
		requiredScopes  []string
		customRoles     map[string][]string
		expectedStatus  int
		setupContextKey bool
		validateFunc    func(bool, *httptest.ResponseRecorder)
	}{
		{
			name:           "no auth header returns 401",
			authHeader:     "",
			requiredScopes: []string{"trace:read"},
			expectedStatus: http.StatusUnauthorized,
			validateFunc: func(called bool, rec *httptest.ResponseRecorder) {
				s.False(called)
				s.Equal(http.StatusUnauthorized, rec.Code)
			},
			setupContextKey: true,
		},
		{
			name:           "non-bearer auth header returns 401",
			authHeader:     "Basic dXNlcjpwYXNz",
			requiredScopes: []string{"trace:read"},
			expectedStatus: http.StatusUnauthorized,
			validateFunc: func(called bool, rec *httptest.ResponseRecorder) {
				s.False(called)
				s.Equal(http.StatusUnauthorized, rec.Code)
			},
			setupContextKey: true,
		},
		{
			name:           "invalid token returns 401",
			authHeader:     "Bearer invalid-token-string",
			requiredScopes: []string{"trace:read"},
			expectedStatus: http.StatusUnauthorized,
			validateFunc: func(called bool, rec *httptest.ResponseRecorder) {
				s.False(called)
				s.Equal(http.StatusUnauthorized, rec.Code)
			},
			setupContextKey: true,
		},
		{
			name: "token from a key that is not in the list returns 401",
			authHeader: func() string {
				t, _ := s.tokenManager.Generate(
					authtoken.Key{ID: "removed", Secret: "a-key-that-the-controller-dropped"},
					[]string{"admin"}, "test-subject", nil, nil,
				)
				return "Bearer " + t
			}(),
			requiredScopes: []string{"trace:read"},
			expectedStatus: http.StatusUnauthorized,
			validateFunc: func(called bool, rec *httptest.ResponseRecorder) {
				s.False(called)
				s.Equal(http.StatusUnauthorized, rec.Code)
				s.Contains(rec.Body.String(), `unknown kid \"removed\"`)
			},
			setupContextKey: true,
		},
		{
			name:           "admin role has trace:read",
			authHeader:     "", // set dynamically
			requiredScopes: []string{"trace:read"},
			expectedStatus: http.StatusOK,
			validateFunc: func(called bool, _ *httptest.ResponseRecorder) {
				s.True(called)
			},
			setupContextKey: true,
		},
		{
			name:           "read role has trace:read",
			authHeader:     "", // set dynamically
			requiredScopes: []string{"trace:read"},
			expectedStatus: http.StatusOK,
			validateFunc: func(called bool, _ *httptest.ResponseRecorder) {
				s.True(called)
			},
			setupContextKey: true,
		},
		{
			name:           "read role lacks dns:write returns 403",
			authHeader:     "", // set dynamically
			requiredScopes: []string{"dns:write"},
			expectedStatus: http.StatusForbidden,
			validateFunc: func(called bool, rec *httptest.ResponseRecorder) {
				s.False(called)
				s.Equal(http.StatusForbidden, rec.Code)
			},
			setupContextKey: true,
		},
		{
			name:           "valid token with no required scopes calls handler",
			authHeader:     "", // set dynamically
			requiredScopes: nil,
			expectedStatus: http.StatusOK,
			validateFunc: func(called bool, _ *httptest.ResponseRecorder) {
				s.True(called)
			},
			setupContextKey: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			handlerCalled := false
			testHandler := api.StrictHandlerFunc(
				func(_ *echo.Context, _ interface{}) (interface{}, error) {
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

			wrapped := api.ExportScopeMiddleware(
				testHandler,
				s.tokenManager,
				authtoken.Keys{testSigningKey},
				contextKey,
				customOf(tt.customRoles),
			)
			_, _ = wrapped(ctx, nil)

			tt.validateFunc(handlerCalled, rec)
		})
	}
}

func (s *MiddlewarePublicTestSuite) TestScopeMiddlewareCustomRoles() {
	contextKey := "BearerAuthScopes"

	tests := []struct {
		name           string
		tokenRoles     []string
		customRoles    map[string][]string
		requiredScope  string
		expectedStatus int
		validateFunc   func(bool, *httptest.ResponseRecorder)
	}{
		{
			name:       "custom role grants access",
			tokenRoles: []string{"ops"},
			customRoles: map[string][]string{
				"ops": {"trace:read", "health:read"},
			},
			requiredScope:  "trace:read",
			expectedStatus: http.StatusOK,
			validateFunc: func(called bool, _ *httptest.ResponseRecorder) {
				s.True(called)
			},
		},
		{
			name:       "custom role lacks permission",
			tokenRoles: []string{"ops"},
			customRoles: map[string][]string{
				"ops": {"health:read"},
			},
			requiredScope:  "trace:read",
			expectedStatus: http.StatusForbidden,
			validateFunc: func(called bool, rec *httptest.ResponseRecorder) {
				s.False(called)
				s.Equal(http.StatusForbidden, rec.Code)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			handlerCalled := false
			testHandler := api.StrictHandlerFunc(
				func(_ *echo.Context, _ interface{}) (interface{}, error) {
					handlerCalled = true
					return "ok", nil
				},
			)

			token, err := s.tokenManager.Generate(
				testSigningKey,
				[]string{"admin"},
				"test-subject",
				nil,
				nil,
			)
			s.Require().NoError(err)

			// Override custom roles to shadow "admin".
			customRoles := authtoken.CustomRoles{
				"admin": authtoken.PermissionsOf(tt.customRoles[tt.tokenRoles[0]]),
			}

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set(contextKey, []string{tt.requiredScope})

			wrapped := api.ExportScopeMiddleware(
				testHandler,
				s.tokenManager,
				authtoken.Keys{testSigningKey},
				contextKey,
				customRoles,
			)
			_, _ = wrapped(ctx, nil)

			tt.validateFunc(handlerCalled, rec)
		})
	}
}

func (s *MiddlewarePublicTestSuite) TestScopeMiddlewareDirectPermissions() {
	contextKey := "BearerAuthScopes"

	tests := []struct {
		name           string
		permissions    []string
		requiredScope  string
		expectedStatus int
		validateFunc   func(bool, *httptest.ResponseRecorder)
	}{
		{
			name:           "direct permission grants access",
			permissions:    []string{"trace:read"},
			requiredScope:  "trace:read",
			expectedStatus: http.StatusOK,
			validateFunc: func(called bool, _ *httptest.ResponseRecorder) {
				s.True(called)
			},
		},
		{
			name:           "direct permission restricts to only listed",
			permissions:    []string{"health:read"},
			requiredScope:  "trace:read",
			expectedStatus: http.StatusForbidden,
			validateFunc: func(called bool, rec *httptest.ResponseRecorder) {
				s.False(called)
				s.Equal(http.StatusForbidden, rec.Code)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			handlerCalled := false
			testHandler := api.StrictHandlerFunc(
				func(_ *echo.Context, _ interface{}) (interface{}, error) {
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

			wrapped := api.ExportScopeMiddleware(
				testHandler,
				s.tokenManager,
				authtoken.Keys{testSigningKey},
				contextKey,
				nil,
			)
			_, _ = wrapped(ctx, nil)

			tt.validateFunc(handlerCalled, rec)
		})
	}
}

func (s *MiddlewarePublicTestSuite) TestScopeMiddlewareTargetScope() {
	contextKey := "BearerAuthScopes"

	validation.RegisterTargetValidator(
		func(_ context.Context) ([]validation.AgentTarget, error) {
			return []validation.AgentTarget{
				{
					MachineID: "m-web1", Hostname: "web-01", Verified: true,
					Labels: map[string]string{"group": "web"},
				},
				{MachineID: "m-db1", Hostname: "db-01", Verified: true},
			}, nil
		},
	)

	tests := []struct {
		name         string
		targets      []string
		hostname     string
		validateFunc func(bool, []string, *httptest.ResponseRecorder)
	}{
		{
			name:     "when the token has no scope, each host runs",
			hostname: "_all",
			validateFunc: func(called bool, scope []string, _ *httptest.ResponseRecorder) {
				s.True(called)
				s.Nil(scope)
			},
		},
		{
			name:     "when the host is in the scope, the handler runs with the scope",
			targets:  []string{"group:web"},
			hostname: "web-01",
			validateFunc: func(called bool, scope []string, _ *httptest.ResponseRecorder) {
				s.True(called)
				s.Equal([]string{"group:web"}, scope)
			},
		},
		{
			name:     "when the label target is in the scope, the handler runs",
			targets:  []string{"group:web"},
			hostname: "group:web",
			validateFunc: func(called bool, _ []string, _ *httptest.ResponseRecorder) {
				s.True(called)
			},
		},
		{
			name:    "when the route has no host, the handler runs",
			targets: []string{"group:web"},
			validateFunc: func(called bool, _ []string, _ *httptest.ResponseRecorder) {
				s.True(called)
			},
		},
		{
			name:     "when the host is outside the scope, it is a 403",
			targets:  []string{"group:web"},
			hostname: "db-01",
			validateFunc: func(called bool, _ []string, rec *httptest.ResponseRecorder) {
				s.False(called)
				s.Equal(http.StatusForbidden, rec.Code)
				s.Contains(rec.Body.String(), `outside the token's target scope`)
			},
		},
		{
			name:     "when a scoped token targets _all, it is a 403",
			targets:  []string{"group:web"},
			hostname: "_all",
			validateFunc: func(called bool, _ []string, rec *httptest.ResponseRecorder) {
				s.False(called)
				s.Equal(http.StatusForbidden, rec.Code)
			},
		},
		{
			name:     "when a scoped token targets _any, it is a 403",
			targets:  []string{"web-01"},
			hostname: "_any",
			validateFunc: func(called bool, _ []string, rec *httptest.ResponseRecorder) {
				s.False(called)
				s.Equal(http.StatusForbidden, rec.Code)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			handlerCalled := false
			var scope []string
			testHandler := api.StrictHandlerFunc(
				func(ctx *echo.Context, _ interface{}) (interface{}, error) {
					handlerCalled = true
					scope = authtoken.TargetsFrom(ctx.Request().Context())
					s.True(authtoken.Granted(ctx.Request().Context(), "trace:read"),
						"the handler has the permissions of the token")
					return "ok", nil
				},
			)

			token, err := s.tokenManager.Generate(
				testSigningKey, []string{"admin"}, "test-subject", nil, tt.targets,
			)
			s.Require().NoError(err)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.Set(contextKey, []string{"trace:read"})
			if tt.hostname != "" {
				ctx.SetPathValues(echo.PathValues{{Name: "hostname", Value: tt.hostname}})
			}

			wrapped := api.ExportScopeMiddleware(
				testHandler,
				s.tokenManager,
				authtoken.Keys{testSigningKey},
				contextKey,
				nil,
			)
			_, _ = wrapped(ctx, nil)

			tt.validateFunc(handlerCalled, scope, rec)
		})
	}
}

func (s *MiddlewarePublicTestSuite) TestScopeMiddlewareInjectsIdentity() {
	contextKey := "BearerAuthScopes"

	tests := []struct {
		name           string
		tokenRoles     []string
		authHeader     string
		requiredScopes []string
		validateFunc   func(bool, string, []string)
	}{
		{
			name:           "admin token injects subject and roles",
			tokenRoles:     []string{"admin"},
			requiredScopes: []string{"trace:read"},
			validateFunc: func(called bool, subject string, roles []string) {
				s.True(called)
				s.Equal("test-subject", subject)
				s.Equal([]string{"admin"}, roles)
			},
		},
		{
			name:           "read token injects subject and read role",
			tokenRoles:     []string{"read"},
			requiredScopes: []string{"trace:read"},
			validateFunc: func(called bool, subject string, roles []string) {
				s.True(called)
				s.Equal("test-subject", subject)
				s.Equal([]string{"read"}, roles)
			},
		},
		{
			name:           "invalid token does not inject identity",
			authHeader:     "Bearer invalid-token-string",
			requiredScopes: []string{"trace:read"},
			validateFunc: func(called bool, _ string, _ []string) {
				s.False(called)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			handlerCalled := false
			var capturedSubject string
			var capturedRoles []string

			testHandler := api.StrictHandlerFunc(
				func(ctx *echo.Context, _ interface{}) (interface{}, error) {
					handlerCalled = true
					capturedSubject, _ = ctx.Get(api.ContextKeySubject).(string)
					capturedRoles, _ = ctx.Get(api.ContextKeyRoles).([]string)
					return "ok", nil
				},
			)

			authHeader := tt.authHeader
			if authHeader == "" && len(tt.tokenRoles) > 0 {
				authHeader = "Bearer " + s.generateToken(tt.tokenRoles)
			}

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if authHeader != "" {
				req.Header.Set("Authorization", authHeader)
			}
			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			if tt.requiredScopes != nil {
				ctx.Set(contextKey, tt.requiredScopes)
			}

			wrapped := api.ExportScopeMiddleware(
				testHandler,
				s.tokenManager,
				authtoken.Keys{testSigningKey},
				contextKey,
				nil,
			)
			_, _ = wrapped(ctx, nil)

			tt.validateFunc(handlerCalled, capturedSubject, capturedRoles)
		})
	}
}

func TestMiddlewarePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MiddlewarePublicTestSuite))
}

// customOf is m as the custom roles of the config.
func customOf(
	m map[string][]string,
) authtoken.CustomRoles {
	if m == nil {
		return nil
	}
	out := make(authtoken.CustomRoles, len(m))
	for role, perms := range m {
		out[authtoken.Role(role)] = authtoken.PermissionsOf(perms)
	}

	return out
}

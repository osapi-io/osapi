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

package authtoken_test

import (
	"encoding/base64"
	"log/slog"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/authtoken"
)

type AuthTokenPublicTestSuite struct {
	suite.Suite

	token      *authtoken.Token
	signingKey authtoken.Key
	oldKey     authtoken.Key
	keys       authtoken.Keys
}

func (s *AuthTokenPublicTestSuite) SetupTest() {
	s.token = authtoken.New(slog.Default())
	s.signingKey = authtoken.Key{ID: "2026-11", Secret: "test-signing-key-for-jwt-operations"}
	s.oldKey = authtoken.Key{ID: "2026-08", Secret: "test-old-signing-key-for-jwt-tests"}
	s.keys = authtoken.Keys{s.signingKey, s.oldKey}
}

func (s *AuthTokenPublicTestSuite) TestNew() {
	t := authtoken.New(slog.Default())
	s.NotNil(t)
}

func (s *AuthTokenPublicTestSuite) TestGenerate() {
	tokenString, err := s.token.Generate(s.signingKey, []string{"admin"}, "test-subject", nil, nil)

	s.NoError(err)
	s.NotEmpty(tokenString)
}

func (s *AuthTokenPublicTestSuite) TestValidate() {
	tests := []struct {
		name         string
		tokenFunc    func() string
		signingKeys  authtoken.Keys
		expectError  bool
		errContains  string
		validateFunc func(*authtoken.CustomClaims)
	}{
		{
			name: "valid token",
			tokenFunc: func() string {
				t, _ := s.token.Generate(s.signingKey, []string{"admin"}, "test-subject", nil, nil)
				return t
			},
			signingKeys: s.keys,
			expectError: false,
			validateFunc: func(claims *authtoken.CustomClaims) {
				s.Equal([]string{"admin"}, claims.Roles)
				s.Equal("test-subject", claims.Subject)
				s.Equal("osapi", claims.Issuer)
				s.Equal("2026-11", claims.KeyID)
			},
		},
		{
			name: "when an older key of the list signed the token, its key verifies it",
			tokenFunc: func() string {
				t, _ := s.token.Generate(s.oldKey, []string{"read"}, "test-subject", nil, nil)
				return t
			},
			signingKeys: s.keys,
			validateFunc: func(claims *authtoken.CustomClaims) {
				s.Equal("2026-08", claims.KeyID)
				s.Equal([]string{"read"}, claims.Roles)
			},
		},
		{
			name: "when the key of the token is no longer in the list, the token is refused",
			tokenFunc: func() string {
				t, _ := s.token.Generate(s.oldKey, []string{"read"}, "test-subject", nil, nil)
				return t
			},
			signingKeys: authtoken.Keys{s.signingKey},
			expectError: true,
			errContains: `unknown kid "2026-08"`,
		},
		{
			name: "when the token has no kid, it is refused",
			tokenFunc: func() string {
				claims := authtoken.CustomClaims{
					Roles: []string{"admin"},
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
					},
				}
				t, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
					SignedString([]byte(s.signingKey.Secret))
				return t
			},
			signingKeys: s.keys,
			expectError: true,
			errContains: "token has no kid",
		},
		{
			name: "when the token has a target scope, it is in the claims",
			tokenFunc: func() string {
				t, _ := s.token.Generate(
					s.signingKey, []string{"operator"}, "team-web", nil,
					[]string{"group:web", "web-01"},
				)
				return t
			},
			signingKeys: s.keys,
			validateFunc: func(claims *authtoken.CustomClaims) {
				s.Equal([]string{"group:web", "web-01"}, claims.Targets)
			},
		},
		{
			name: "when a target of the scope is _all, the token is refused",
			tokenFunc: func() string {
				t, _ := s.token.Generate(
					s.signingKey, []string{"operator"}, "team-web", nil, []string{"_all"},
				)
				return t
			},
			signingKeys: s.keys,
			expectError: true,
			errContains: "Targets",
		},
		{
			name: "wrong signing key",
			tokenFunc: func() string {
				t, _ := s.token.Generate(s.signingKey, []string{"read"}, "test-subject", nil, nil)
				return t
			},
			signingKeys: authtoken.Keys{{ID: "2026-11", Secret: "wrong-key"}},
			expectError: true,
			errContains: "signature is invalid",
		},
		{
			name: "malformed token",
			tokenFunc: func() string {
				return "not-a-valid-jwt-token"
			},
			signingKeys: s.keys,
			expectError: true,
			errContains: "invalid number of segments",
		},
		{
			name: "empty token",
			tokenFunc: func() string {
				return ""
			},
			signingKeys: s.keys,
			expectError: true,
			errContains: "invalid number of segments",
		},
		{
			name: "unexpected signing method",
			tokenFunc: func() string {
				header := base64.RawURLEncoding.EncodeToString(
					[]byte(`{"alg":"none","typ":"JWT"}`),
				)
				payload := base64.RawURLEncoding.EncodeToString(
					[]byte(`{"roles":["admin"]}`),
				)
				return header + "." + payload + "."
			},
			signingKeys: s.keys,
			expectError: true,
			errContains: "unexpected signing method",
		},
		{
			name: "claims fail struct validation",
			tokenFunc: func() string {
				claims := authtoken.CustomClaims{
					Roles: []string{""},
					RegisteredClaims: jwt.RegisteredClaims{
						Issuer:    "osapi",
						IssuedAt:  jwt.NewNumericDate(time.Now()),
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
						Subject:   "test",
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				token.Header["kid"] = s.signingKey.ID
				t, _ := token.SignedString([]byte(s.signingKey.Secret))
				return t
			},
			signingKeys: s.keys,
			expectError: true,
			errContains: "Roles",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tokenString := tt.tokenFunc()

			claims, err := s.token.Validate(tokenString, tt.signingKeys)

			if tt.expectError {
				s.Error(err)
				s.Nil(claims)
				if tt.errContains != "" {
					s.Contains(err.Error(), tt.errContains)
				}
			} else {
				s.NoError(err)
				s.NotNil(claims)
				tt.validateFunc(claims)
			}
		})
	}
}

func (s *AuthTokenPublicTestSuite) TestGenerateAndValidateRoundTrip() {
	tests := []struct {
		name         string
		roles        []string
		subject      string
		validateFunc func(*authtoken.CustomClaims, error)
	}{
		{
			name:    "admin role round trip",
			roles:   []string{"admin"},
			subject: "admin-user",
			validateFunc: func(claims *authtoken.CustomClaims, err error) {
				s.NoError(err)
				s.NotNil(claims)
				s.Equal([]string{"admin"}, claims.Roles)
				s.Equal("admin-user", claims.Subject)
			},
		},
		{
			name:    "multiple roles round trip",
			roles:   []string{"read", "automation"},
			subject: "rw-user",
			validateFunc: func(claims *authtoken.CustomClaims, err error) {
				s.NoError(err)
				s.NotNil(claims)
				s.Equal([]string{"read", "automation"}, claims.Roles)
				s.Equal("rw-user", claims.Subject)
			},
		},
		{
			name:    "read only round trip",
			roles:   []string{"read"},
			subject: "reader",
			validateFunc: func(claims *authtoken.CustomClaims, err error) {
				s.NoError(err)
				s.NotNil(claims)
				s.Equal([]string{"read"}, claims.Roles)
				s.Equal("reader", claims.Subject)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tokenString, err := s.token.Generate(s.signingKey, tt.roles, tt.subject, nil, nil)
			s.NoError(err)
			s.NotEmpty(tokenString)

			tt.validateFunc(s.token.Validate(tokenString, s.keys))
		})
	}
}

func (s *AuthTokenPublicTestSuite) TestSigner() {
	tests := []struct {
		name   string
		keys   authtoken.Keys
		wantID string
		wantOK bool
	}{
		{
			name:   "when there are keys, the first signs",
			keys:   s.keys,
			wantID: "2026-11",
			wantOK: true,
		},
		{
			name: "when there is no key, nothing signs",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			key, ok := tt.keys.Signer()

			s.Equal(tt.wantOK, ok)
			s.Equal(tt.wantID, key.ID)
		})
	}
}

func TestAuthTokenPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(AuthTokenPublicTestSuite))
}

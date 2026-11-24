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

package token_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/cli/commands/token"

	// The endpoint groups declare their permissions, as in the binary.
	"github.com/osapi-io/osapi/internal/cli/commands/token/mocks"
	_ "github.com/osapi-io/osapi/internal/controller/endpoints/all"
)

var (
	key  = authtoken.Key{ID: "2026-11", Secret: "0123456789abcdef0123456789abcdef"}
	keys = authtoken.Keys{key}
)

type MethodsPublicTestSuite struct {
	suite.Suite

	ctrl   *gomock.Controller
	tokens *mocks.MockTokens
	config *mocks.MockConfig
	deps   token.Deps
}

func (s *MethodsPublicTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.tokens = mocks.NewMockTokens(s.ctrl)
	s.config = mocks.NewMockConfig(s.ctrl)
	s.deps = token.Deps{Tokens: s.tokens, Config: s.config}
}

func (s *MethodsPublicTestSuite) TestGenerate() {
	tests := []struct {
		name         string
		in           token.GenerateInput
		setup        func()
		validateFunc func(token.GenerateOutput, error)
	}{
		{
			name: "when the token is signed, it is the answer with what it carries",
			in: token.GenerateInput{
				Roles: []string{"read"}, Subject: "ops", Permissions: []string{"trace:read"},
				Targets: []string{"group:web"},
			},
			setup: func() {
				s.config.EXPECT().CustomRoles().Return(nil, nil)
				s.config.EXPECT().SigningKeys(gomock.Any()).Return(keys, nil)
				s.tokens.EXPECT().
					Generate(key, []string{"read"}, "ops", []string{"trace:read"}, []string{"group:web"}).
					Return("signed", nil)
			},
			validateFunc: func(got token.GenerateOutput, err error) {
				s.Require().NoError(err)
				s.Equal(token.GenerateOutput{
					Token: "signed", KeyID: "2026-11", Subject: "ops",
					Roles: []string{"read"}, Permissions: []string{"trace:read"},
					Targets: []string{"group:web"},
				}, got)
			},
		},
		{
			name: "when the token cannot be signed, it is the error",
			in:   token.GenerateInput{Roles: []string{"read"}, Subject: "ops"},
			setup: func() {
				s.config.EXPECT().CustomRoles().Return(nil, nil)
				s.config.EXPECT().SigningKeys(gomock.Any()).Return(keys, nil)
				s.tokens.EXPECT().Generate(key, []string{"read"}, "ops", nil, nil).
					Return("", errors.New("no key"))
			},
			validateFunc: func(_ token.GenerateOutput, err error) {
				s.EqualError(err, "failed to generate token: no key")
			},
		},
		{
			name: "when the config has no signing key, it is refused",
			in:   token.GenerateInput{Roles: []string{"read"}, Subject: "ops"},
			setup: func() {
				s.config.EXPECT().CustomRoles().Return(nil, nil)
				s.config.EXPECT().SigningKeys(gomock.Any()).Return(nil, nil)
			},
			validateFunc: func(_ token.GenerateOutput, err error) {
				s.ErrorIs(err, token.ErrNoSigningKey)
			},
		},
		{
			name: "when a signing key does not resolve, it is the error",
			in:   token.GenerateInput{Roles: []string{"read"}, Subject: "ops"},
			setup: func() {
				s.config.EXPECT().CustomRoles().Return(nil, nil)
				s.config.EXPECT().SigningKeys(gomock.Any()).
					Return(nil, errors.New("secret://default/jwt: no such key"))
			},
			validateFunc: func(_ token.GenerateOutput, err error) {
				s.EqualError(err, "secret://default/jwt: no such key")
			},
		},
		{
			name: "when the role is a custom role of the config, it is signed",
			in:   token.GenerateInput{Roles: []string{"security"}, Subject: "ops"},
			setup: func() {
				s.config.EXPECT().CustomRoles().
					Return(authtoken.CustomRoles{"security": {"audit:read"}}, nil)
				s.config.EXPECT().SigningKeys(gomock.Any()).Return(keys, nil)
				s.tokens.EXPECT().Generate(key, []string{"security"}, "ops", nil, nil).
					Return("signed", nil)
			},
			validateFunc: func(got token.GenerateOutput, err error) {
				s.Require().NoError(err)
				s.Equal("signed", got.Token)
			},
		},
		{
			name: "when the role is neither built in nor of the config, it is refused",
			in:   token.GenerateInput{Roles: []string{"root"}, Subject: "ops"},
			setup: func() {
				s.config.EXPECT().CustomRoles().Return(nil, nil)
			},
			validateFunc: func(_ token.GenerateOutput, err error) {
				s.ErrorIs(err, token.ErrNoRole)
				s.ErrorContains(err, `role "root"`)
			},
		},
		{
			name: "when the config cannot be read, it is the error",
			in:   token.GenerateInput{Roles: []string{"read"}, Subject: "ops"},
			setup: func() {
				s.config.EXPECT().CustomRoles().Return(nil, errors.New("no file"))
			},
			validateFunc: func(_ token.GenerateOutput, err error) {
				s.EqualError(err, "no file")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()

			got, err := token.Generate(context.Background(), s.deps, tc.in)

			tc.validateFunc(got, err)
		})
	}
}

func (s *MethodsPublicTestSuite) TestValidate() {
	issued := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	expires := issued.Add(time.Hour)

	tests := []struct {
		name         string
		setup        func()
		validateFunc func(token.Claims, error)
	}{
		{
			name: "when the token is valid, it is its claims with what they grant",
			setup: func() {
				s.config.EXPECT().SigningKeys(gomock.Any()).Return(keys, nil)
				s.tokens.EXPECT().Validate("jwt", keys).Return(&authtoken.CustomClaims{
					Roles:       []string{"read"},
					Permissions: []string{"trace:read", "audit:read"},
					Targets:     []string{"group:web"},
					KeyID:       "2026-11",
					RegisteredClaims: jwt.RegisteredClaims{
						Subject:   "ops",
						Audience:  jwt.ClaimStrings{"osapi"},
						IssuedAt:  jwt.NewNumericDate(issued),
						ExpiresAt: jwt.NewNumericDate(expires),
					},
				}, nil)
			},
			validateFunc: func(got token.Claims, err error) {
				s.Require().NoError(err)
				s.Equal(token.Claims{
					KeyID:                "2026-11",
					Subject:              "ops",
					Roles:                []string{"read"},
					Permissions:          []string{"trace:read", "audit:read"},
					Targets:              []string{"group:web"},
					EffectivePermissions: []string{"audit:read", "trace:read"},
					Audience:             []string{"osapi"},
					IssuedAt:             issued,
					ExpiresAt:            expires,
				}, got)
			},
		},
		{
			name: "when the token has no times, they are zero",
			setup: func() {
				s.config.EXPECT().SigningKeys(gomock.Any()).Return(keys, nil)
				s.tokens.EXPECT().Validate("jwt", keys).Return(&authtoken.CustomClaims{
					Roles: []string{"read"},
				}, nil)
			},
			validateFunc: func(got token.Claims, err error) {
				s.Require().NoError(err)
				s.True(got.IssuedAt.IsZero())
				s.True(got.ExpiresAt.IsZero())
				s.NotEmpty(got.EffectivePermissions)
			},
		},
		{
			name: "when a signing key does not resolve, it is the error",
			setup: func() {
				s.config.EXPECT().SigningKeys(gomock.Any()).
					Return(nil, errors.New("secret://default/jwt: no such key"))
			},
			validateFunc: func(_ token.Claims, err error) {
				s.EqualError(err, "secret://default/jwt: no such key")
			},
		},
		{
			name: "when the token is not valid, it is the error",
			setup: func() {
				s.config.EXPECT().SigningKeys(gomock.Any()).Return(keys, nil)
				s.tokens.EXPECT().Validate("jwt", keys).Return(nil, errors.New("expired"))
			},
			validateFunc: func(_ token.Claims, err error) {
				s.EqualError(err, "failed to validate token: expired")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()

			got, err := token.Validate(
				context.Background(),
				s.deps,
				token.ValidateInput{Token: "jwt"},
			)

			tc.validateFunc(got, err)
		})
	}
}

// TestPermissionRule holds --permissions to the permissions of the binary.
func (s *MethodsPublicTestSuite) TestPermissionRule() {
	tests := []struct {
		name    string
		perms   []string
		wantErr string
	}{
		{
			name:  "when each permission is one of the binary's, it passes",
			perms: []string{"trace:read", "audit:read"},
		},
		{
			name:    "when a permission is not one, it is a usage error",
			perms:   []string{"trace:read", "root:all"},
			wantErr: `--permissions: "root:all" is not a permission`,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := command.Check(context.Background(), token.GenerateInput{
				Roles: []string{"read"}, Subject: "ops", Permissions: tc.perms,
			})

			if tc.wantErr == "" {
				s.NoError(err)

				return
			}
			s.ErrorContains(err, tc.wantErr)
		})
	}
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}

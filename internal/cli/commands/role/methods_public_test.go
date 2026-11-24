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

package role_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/cli/commands/role"
	"github.com/osapi-io/osapi/internal/cli/commands/role/mocks"
)

type MethodsPublicTestSuite struct {
	suite.Suite

	ctrl   *gomock.Controller
	config *mocks.MockConfig
	deps   role.Deps
}

func (s *MethodsPublicTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.config = mocks.NewMockConfig(s.ctrl)
	s.deps = role.Deps{Config: s.config}
}

func (s *MethodsPublicTestSuite) TestList() {
	tests := []struct {
		name         string
		setup        func()
		validateFunc func(role.ListOutput, error)
	}{
		{
			name: "when the config has custom roles, they follow the built-in ones by name",
			setup: func() {
				s.config.EXPECT().CustomRoles().Return(authtoken.CustomRoles{
					"ops":      {"trace:read"},
					"security": {"audit:read"},
				}, nil)
			},
			validateFunc: func(got role.ListOutput, err error) {
				s.Require().NoError(err)
				names := make([]string, 0, len(got.Roles))
				for _, r := range got.Roles {
					names = append(names, r.Name+"/"+r.Source)
				}
				s.Equal([]string{
					"admin/built-in", "operator/built-in", "automation/built-in",
					"auditor/built-in", "read/built-in", "observer/built-in",
					"ops/config", "security/config",
				}, names)
				s.Equal([]string{"trace:read"}, got.Roles[6].Permissions)
				perms, _ := authtoken.RolePermissions("read")
				s.Equal(authtoken.Strings(perms), got.Roles[4].Permissions)
			},
		},
		{
			name: "when the config cannot be read, it is the error",
			setup: func() {
				s.config.EXPECT().CustomRoles().Return(nil, errors.New("bad yaml"))
			},
			validateFunc: func(_ role.ListOutput, err error) {
				s.EqualError(err, "bad yaml")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()

			got, err := role.List(context.Background(), s.deps, role.ListInput{})

			tc.validateFunc(got, err)
		})
	}
}

func (s *MethodsPublicTestSuite) TestGet() {
	tests := []struct {
		name         string
		role         string
		setup        func()
		validateFunc func(role.Role, error)
	}{
		{
			name:  "when the role is built in, it is the role",
			role:  "read",
			setup: func() { s.config.EXPECT().CustomRoles().Return(nil, nil) },
			validateFunc: func(got role.Role, err error) {
				s.Require().NoError(err)
				s.Equal("read", got.Name)
				s.Equal("built-in", got.Source)
			},
		},
		{
			name: "when the role is of the config, it is the role",
			role: "ops",
			setup: func() {
				s.config.EXPECT().
					CustomRoles().
					Return(authtoken.CustomRoles{"ops": {"trace:read"}}, nil)
			},
			validateFunc: func(got role.Role, err error) {
				s.Require().NoError(err)
				s.Equal(
					role.Role{Name: "ops", Source: "config", Permissions: []string{"trace:read"}},
					got,
				)
			},
		},
		{
			name:  "when no role has the name, it is ErrNoRole",
			role:  "nope",
			setup: func() { s.config.EXPECT().CustomRoles().Return(nil, nil) },
			validateFunc: func(_ role.Role, err error) {
				s.ErrorIs(err, role.ErrNoRole)
				s.EqualError(err, `role "nope": no such role`)
			},
		},
		{
			name:  "when the config cannot be read, it is the error",
			role:  "read",
			setup: func() { s.config.EXPECT().CustomRoles().Return(nil, errors.New("bad yaml")) },
			validateFunc: func(_ role.Role, err error) {
				s.EqualError(err, "bad yaml")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()

			got, err := role.Get(context.Background(), s.deps, role.GetInput{Name: tc.role})

			tc.validateFunc(got, err)
		})
	}
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}

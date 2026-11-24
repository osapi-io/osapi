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
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/behavior"
	_ "github.com/osapi-io/osapi/internal/behaviors/all"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	_ "github.com/osapi-io/osapi/internal/controller/endpoints/all"
	"github.com/osapi-io/osapi/internal/validation"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

type motdProvider interface{ Message() string }

type motdIn struct {
	Text string `json:"text"`
}

type motdOut struct {
	Text    string `json:"text"`
	Changed bool   `json:"changed"`
}

func motd(
	_ context.Context,
	_ motdProvider,
	in motdIn,
) (motdOut, error) {
	return motdOut{Text: in.Text}, nil
}

// A behavior registered in this test binary only, so its permissions are
// not among the real behaviors'.
// hiddenperm is not served, so a token can never call it and it adds no
// permission.
var _ = behavior.Register(behavior.Define("hiddenperm", "Used by other behaviors", behavior.Node,
	behavior.Internal,
	behavior.Read("get", motd),
))

var _ = behavior.Register(behavior.Define("motdperm", "Message of the day", behavior.API,
	behavior.Node,
	behavior.Read("get", motd),
	behavior.Write("update", motd),
))

// keyperm's write is admin only.
var _ = behavior.Register(behavior.Define("keyperm", "Keys", behavior.API,
	behavior.Node,
	behavior.Read("get", motd),
	behavior.Write("update", motd).AdminOnly(),
))

// Endpoint groups registered in this test binary only: one whose method
// names a permission that nothing declares, and one whose two writes
// disagree on AdminOnly.
type epIn struct{}

var (
	_ = endpoint.Register(endpoint.Define("zztypo", "A typo",
		endpoint.Get[epIn, epIn]("status").Permission("helth:read"),
	))
	_ = endpoint.Register(endpoint.Define("zzsplit", "Two writes",
		endpoint.Post[epIn, epIn]("one").AdminOnly(),
		endpoint.Post[epIn, epIn]("two"),
		endpoint.Get[epIn, epIn]("open").Public(),
	))
)

type PermissionsPublicTestSuite struct {
	suite.Suite
}

func (s *PermissionsPublicTestSuite) TestResolvePermissions() {
	tests := []struct {
		name              string
		roles             []string
		directPermissions []string
		customRoles       map[string][]string
		validateFunc      func(map[string]bool)
	}{
		{
			name:  "admin role gets all permissions",
			roles: []string{"admin"},
			validateFunc: func(resolved map[string]bool) {
				for _, p := range authtoken.AllPermissions() {
					s.True(resolved[string(p)], "expected permission %s to be present", p)
				}
			},
		},
		{
			name:  "admin role gets command:execute and command:shell",
			roles: []string{"admin"},
			validateFunc: func(resolved map[string]bool) {
				for _, p := range []string{
					string(types.PermCommandExecute),
					string(types.PermCommandShell),
				} {
					s.True(resolved[string(p)], "expected permission %s to be present", p)
				}
			},
		},
		{
			name:  "automation role gets write permissions but not audit or command",
			roles: []string{"automation"},
			validateFunc: func(resolved map[string]bool) {
				for _, p := range []string{
					string(types.PermHostnameRead),
					string(types.PermDnsRead),
					string(types.PermDnsWrite),
					string(types.PermJobRead),
					string(types.PermJobWrite),
					string(types.PermHealthRead),
					string(types.PermFileRead),
					string(types.PermFileWrite),
				} {
					s.True(resolved[string(p)], "expected permission %s to be present", p)
				}
				for _, p := range []string{
					string(types.PermAuditRead),
					string(types.PermCommandExecute),
					string(types.PermCommandShell),
				} {
					s.False(resolved[string(p)], "expected permission %s to be absent", p)
				}
			},
		},
		{
			name:  "read role gets read-only permissions",
			roles: []string{"read"},
			validateFunc: func(resolved map[string]bool) {
				for _, p := range []string{
					string(types.PermHostnameRead),
					string(types.PermDnsRead),
					string(types.PermJobRead),
					string(types.PermHealthRead),
					string(types.PermFileRead),
				} {
					s.True(resolved[string(p)], "expected permission %s to be present", p)
				}
				for _, p := range []string{
					string(types.PermDnsWrite),
					string(types.PermJobWrite),
					string(types.PermAuditRead),
					string(types.PermFileWrite),
					string(types.PermCommandExecute),
					string(types.PermCommandShell),
				} {
					s.False(resolved[string(p)], "expected permission %s to be absent", p)
				}
			},
		},
		{
			name:  "unknown role gets no permissions",
			roles: []string{"unknown"},
			validateFunc: func(resolved map[string]bool) {
				for _, p := range authtoken.AllPermissions() {
					s.False(resolved[string(p)], "expected permission %s to be absent", p)
				}
			},
		},
		{
			name:  "empty roles gets no permissions",
			roles: []string{},
			validateFunc: func(resolved map[string]bool) {
				for _, p := range authtoken.AllPermissions() {
					s.False(resolved[string(p)], "expected permission %s to be absent", p)
				}
			},
		},
		{
			name:  "nil roles gets no permissions",
			roles: nil,
			validateFunc: func(resolved map[string]bool) {
				for _, p := range authtoken.AllPermissions() {
					s.False(resolved[string(p)], "expected permission %s to be absent", p)
				}
			},
		},
		{
			name:              "direct permissions override roles",
			roles:             []string{"admin"},
			directPermissions: []string{string(types.PermHostnameRead)},
			validateFunc: func(resolved map[string]bool) {
				for _, p := range []string{string(types.PermHostnameRead)} {
					s.True(resolved[string(p)], "expected permission %s to be present", p)
				}
				for _, p := range []string{
					string(types.PermDnsRead),
					string(types.PermDnsWrite),
					string(types.PermJobRead),
					string(types.PermJobWrite),
					string(types.PermHealthRead),
					string(types.PermAuditRead),
				} {
					s.False(resolved[string(p)], "expected permission %s to be absent", p)
				}
			},
		},
		{
			name:  "custom role overrides default",
			roles: []string{"ops"},
			customRoles: map[string][]string{
				"ops": {string(types.PermHostnameRead), string(types.PermHealthRead)},
			},
			validateFunc: func(resolved map[string]bool) {
				for _, p := range []string{
					string(types.PermHostnameRead),
					string(types.PermHealthRead),
				} {
					s.True(resolved[string(p)], "expected permission %s to be present", p)
				}
				for _, p := range []string{
					string(types.PermDnsRead),
					string(types.PermDnsWrite),
					string(types.PermJobRead),
					string(types.PermJobWrite),
				} {
					s.False(resolved[string(p)], "expected permission %s to be absent", p)
				}
			},
		},
		{
			name:  "custom role shadows built-in role",
			roles: []string{"read"},
			customRoles: map[string][]string{
				"read": {string(types.PermHealthRead)},
			},
			validateFunc: func(resolved map[string]bool) {
				for _, p := range []string{string(types.PermHealthRead)} {
					s.True(resolved[string(p)], "expected permission %s to be present", p)
				}
				for _, p := range []string{
					string(types.PermHostnameRead),
					string(types.PermDnsRead),
				} {
					s.False(resolved[string(p)], "expected permission %s to be absent", p)
				}
			},
		},
		{
			name:  "a built-in role grants a behavior's permissions by verb",
			roles: []string{"read"},
			validateFunc: func(resolved map[string]bool) {
				s.True(resolved["motdperm:read"])
				s.False(resolved["motdperm:write"])
			},
		},
		{
			name:  "multiple roles merge permissions",
			roles: []string{"read", "automation"},
			validateFunc: func(resolved map[string]bool) {
				for _, p := range []string{
					string(types.PermHostnameRead),
					string(types.PermDnsRead),
					string(types.PermDnsWrite),
					string(types.PermJobRead),
					string(types.PermJobWrite),
					string(types.PermHealthRead),
				} {
					s.True(resolved[string(p)], "expected permission %s to be present", p)
				}
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			resolved := authtoken.ResolvePermissions(
				authtoken.RolesOf(tt.roles),
				authtoken.PermissionsOf(tt.directPermissions),
				customOf(tt.customRoles),
			)
			set := make(map[string]bool, len(resolved))
			for p, ok := range resolved {
				set[string(p)] = ok
			}
			tt.validateFunc(set)
		})
	}
}

func (s *PermissionsPublicTestSuite) TestAllPermissions() {
	all := authtoken.AllPermissions()

	s.Contains(all, types.PermCommandShell, "a behavior's, from its methods")
	s.Contains(all, types.PermAuditRead, "an endpoint group's, from its methods")
	s.Contains(all, authtoken.Permission("zzsplit:write"))
	s.NotContains(
		all,
		authtoken.Permission("helth:read"),
		"a permission that a method only names is not one",
	)
	s.Contains(all, authtoken.Permission("motdperm:read"), "a behavior's, from its methods")
	s.Contains(all, authtoken.Permission("motdperm:write"))
	s.NotContains(
		all,
		authtoken.Permission("hiddenperm:read"),
		"a behavior with no API adds no permission",
	)
	seen := map[authtoken.Permission]bool{}
	for _, p := range all {
		s.False(seen[p], "%s is listed twice", p)
		seen[p] = true
	}
}

func (s *PermissionsPublicTestSuite) TestRolePermissions() {
	tests := []struct {
		name         string
		role         authtoken.Role
		validateFunc func(perms []authtoken.Permission, ok bool)
	}{
		{
			name: "when admin, it gets each permission",
			role: authtoken.RoleAdmin,
			validateFunc: func(perms []authtoken.Permission, ok bool) {
				s.True(ok)
				s.Equal(authtoken.AllPermissions(), perms)
				s.Contains(perms, types.PermCommandShell)
				s.Contains(perms, types.PermSecretReveal, "the value of a secret")
			},
		},
		{
			name: "when operator, it gets each read and each execute that is not admin only",
			role: authtoken.RoleOperator,
			validateFunc: func(perms []authtoken.Permission, ok bool) {
				s.True(ok)
				s.Contains(perms, authtoken.Permission("motdperm:read"))
				s.NotContains(
					perms,
					authtoken.Permission("motdperm:write"),
					"a write is not an action",
				)
				s.Contains(perms, types.PermPowerExecute)
				s.Contains(perms, types.PermProcessExecute)
				s.NotContains(
					perms,
					types.PermCommandExecute,
					"an arbitrary command is admin only",
				)
				s.Contains(
					perms,
					types.PermDockerExecute,
					"a container's start and stop are not admin only",
				)
				s.NotContains(
					perms,
					types.PermDockerExec,
					"a command in a container is admin only",
				)
				s.NotContains(perms, types.PermCommandShell)
				s.Contains(perms, types.PermSecretRead, "the names of the secrets")
				s.NotContains(perms, types.PermSecretWrite, "a value is a write")
				s.NotContains(perms, types.PermSecretUse, "no action takes a secret")
			},
		},
		{
			name: "when automation, it gets each read and each write that is not admin only",
			role: authtoken.RoleAutomation,
			validateFunc: func(perms []authtoken.Permission, ok bool) {
				s.True(ok)
				s.Contains(perms, authtoken.Permission("motdperm:read"))
				s.Contains(perms, authtoken.Permission("motdperm:write"))
				s.Contains(perms, authtoken.Permission("keyperm:read"))
				s.NotContains(
					perms,
					authtoken.Permission("keyperm:write"),
					"marked AdminOnly on a behavior",
				)
				s.NotContains(perms, types.PermAdmissionWrite, "marked AdminOnly on an endpoint")
				s.NotContains(perms, types.PermAuditRead, "a read can be admin only")
				s.NotContains(perms, types.PermPowerExecute, "an action is not a write")
				s.NotContains(perms, types.PermUserWrite, "an account reaches root")
				s.NotContains(perms, types.PermScheduleWrite, "a cron entry runs as any user")
				s.Contains(perms, types.PermFileWrite, "a file stays with automation")
				s.Contains(perms, types.PermJobWrite)
				s.Contains(perms, types.PermSecretRead)
				s.Contains(perms, types.PermSecretWrite)
				s.Contains(perms, types.PermSecretUse, "a secret goes into a write")
				s.NotContains(perms, types.PermSecretReveal, "a value is admin only")
			},
		},
		{
			name: "when read, it gets each read that is not admin only",
			role: authtoken.RoleRead,
			validateFunc: func(perms []authtoken.Permission, ok bool) {
				s.True(ok)
				s.Contains(perms, authtoken.Permission("motdperm:read"))
				s.NotContains(perms, authtoken.Permission("motdperm:write"))
				s.NotContains(perms, types.PermAuditRead)
				s.Contains(perms, types.PermAdmissionRead)
				s.Contains(perms, types.PermLogRead, "read keeps logs, files and processes")
				s.Contains(perms, types.PermSecretRead, "the names, and no value")
				s.NotContains(perms, types.PermSecretWrite)
				s.NotContains(perms, types.PermSecretUse)
			},
		},
		{
			name: "when auditor, it gets the audit log, the traces, the jobs, the health and the catalog",
			role: authtoken.RoleAuditor,
			validateFunc: func(perms []authtoken.Permission, ok bool) {
				s.True(ok)
				s.ElementsMatch([]authtoken.Permission{
					types.PermAuditRead,
					types.PermTraceRead,
					types.PermJobRead,
					types.PermHealthRead,
					types.PermCatalogRead,
				}, perms)
			},
		},
		{
			name: "when observer, it gets the health, the status, the agents and the catalog only",
			role: authtoken.RoleObserver,
			validateFunc: func(perms []authtoken.Permission, ok bool) {
				s.True(ok)
				s.ElementsMatch([]authtoken.Permission{
					types.PermHealthRead, types.PermStatusRead, types.PermAgentRead,
					types.PermCatalogRead,
				}, perms)
			},
		},
		{
			name: "when the role is not built in",
			role: "write",
			validateFunc: func(perms []authtoken.Permission, ok bool) {
				s.False(ok)
				s.Nil(perms)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.validateFunc(authtoken.RolePermissions(tt.role))
		})
	}
}

func (s *PermissionsPublicTestSuite) TestPermissionRule() {
	_, ok := validation.Var(context.Background(), "job:read", authtoken.PermissionRule)
	s.True(ok, "a permission that osapi knows")
	msg, ok := validation.Var(context.Background(), "helth:read", authtoken.PermissionRule)
	s.False(ok)
	s.Contains(msg, "is not a permission")
}

func (s *PermissionsPublicTestSuite) TestRoles() {
	s.Equal(
		[]authtoken.Role{"admin", "operator", "automation", "auditor", "read", "observer"},
		authtoken.Roles(),
	)
}

func (s *PermissionsPublicTestSuite) TestAdminOnly() {
	s.True(authtoken.AdminOnly(types.PermCommandShell), "by its verb")
	s.True(authtoken.AdminOnly(types.PermDockerExec), "by its verb")
	s.True(authtoken.AdminOnly(types.PermCommandExecute), "by its mark")
	s.False(authtoken.AdminOnly(types.PermPowerExecute), "an execute verb is not admin only")
	s.True(authtoken.AdminOnly(types.PermAgentWrite), "by its mark")
	s.False(authtoken.AdminOnly(types.PermAgentRead))
}

func (s *PermissionsPublicTestSuite) TestCheck() {
	errs := authtoken.Check()
	got := make([]string, 0, len(errs))
	for _, err := range errs {
		got = append(got, err.Error())
	}

	s.Equal([]string{
		"endpoint zztypo status names helth:read, which nothing declares: no role grants it",
		"zzsplit:write is AdminOnly on endpoint zzsplit one but not on endpoint zzsplit two",
	}, got, "only this binary's test groups are wrong; the real ones are not")
}

func (s *PermissionsPublicTestSuite) TestGo() {
	src := authtoken.Go()

	s.True(strings.HasPrefix(src, "// Code generated"), "it says it is generated")
	s.Contains(src, "package types\n")
	for _, p := range authtoken.AllPermissions() {
		s.Regexp(`\t`+authtoken.ConstName(p)+` +Permission = "`+string(p)+`"\n`, src,
			"a constant for %s", p)
	}
}

func (s *PermissionsPublicTestSuite) TestConstName() {
	s.Equal("PermDiskRead", authtoken.ConstName("disk:read"))
	s.Equal("PermUpdatepackagesWrite", authtoken.ConstName("updatepackages:write"))
	s.Equal("PermScheduleRead", authtoken.ConstName("schedule:read"),
		"the name is the permission's, not a domain's old name")
	s.Equal("Perm", authtoken.ConstName(""), "an empty permission has no parts")
}

func (s *PermissionsPublicTestSuite) TestRoleTable() {
	table := authtoken.RoleTable()

	s.True(strings.HasPrefix(table, "| Role | Permissions |\n| --- | --- |\n"))
	for _, role := range authtoken.Roles() {
		perms, _ := authtoken.RolePermissions(role)
		names := make([]string, 0, len(perms))
		for _, p := range perms {
			names = append(names, "`"+string(p)+"`")
		}
		s.Contains(table, "| `"+string(role)+"` | "+strings.Join(names, ", ")+" |",
			"a row with each permission of %s", role)
	}
	s.False(strings.HasSuffix(table, "\n"), "the caller adds the lines around it")
}

// TestRoleDescription says who each built-in role is for, and nothing for a
// role that is not built in.
func (s *PermissionsPublicTestSuite) TestRoleDescription() {
	for _, r := range authtoken.Roles() {
		s.NotEmpty(authtoken.RoleDescription(r), "%s has a description", r)
	}
	s.Empty(authtoken.RoleDescription("write"))
}

func (s *PermissionsPublicTestSuite) TestHasPermission() {
	tests := []struct {
		name         string
		resolved     map[authtoken.Permission]bool
		required     authtoken.Permission
		validateFunc func(bool)
	}{
		{
			name:     "present permission returns true",
			resolved: map[authtoken.Permission]bool{types.PermHostnameRead: true},
			required: types.PermHostnameRead,
			validateFunc: func(got bool) {
				s.Equal(true, got)
			},
		},
		{
			name:     "absent permission returns false",
			resolved: map[authtoken.Permission]bool{types.PermHostnameRead: true},
			required: types.PermJobWrite,
			validateFunc: func(got bool) {
				s.Equal(false, got)
			},
		},
		{
			name:     "empty resolved set returns false",
			resolved: map[authtoken.Permission]bool{},
			required: types.PermHostnameRead,
			validateFunc: func(got bool) {
				s.Equal(false, got)
			},
		},
		{
			name:     "nil resolved set returns false",
			resolved: nil,
			required: types.PermHostnameRead,
			validateFunc: func(got bool) {
				s.Equal(false, got)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := authtoken.HasPermission(tt.resolved, tt.required)
			tt.validateFunc(result)
		})
	}
}

func TestPermissionsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(PermissionsPublicTestSuite))
}

func (s *PermissionsPublicTestSuite) TestStrings() {
	tests := []struct {
		name  string
		perms []authtoken.Permission
		want  []string
	}{
		{
			name:  "when there are permissions, each is its text",
			perms: []authtoken.Permission{types.PermHostnameRead, types.PermJobWrite},
			want:  []string{"hostname:read", "job:write"},
		},
		{
			name:  "when there are none, it is empty",
			perms: nil,
			want:  []string{},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, authtoken.Strings(tc.perms))
		})
	}
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

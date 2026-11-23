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

package authtoken

import (
	"fmt"
	"slices"
	"strings"

	"github.com/osapi-io/osapi/internal/dsl"

	"github.com/osapi-io/osapi/internal/validation/rule"
	"github.com/osapi-io/osapi/pkg/sdk/client"
)

// Permission is a type alias for client.Permission.
type Permission = client.Permission

// Role is a type alias for client.Role: a built-in role or a custom role of
// the config.
type Role = client.Role

// CustomRoles are the custom roles of the config, each with its
// permissions.
type CustomRoles map[Role][]Permission

// RolesOf is roles, as a token's claims hold them, as Roles.
func RolesOf(
	roles []string,
) []Role {
	out := make([]Role, 0, len(roles))
	for _, r := range roles {
		out = append(out, Role(r))
	}

	return out
}

// PermissionsOf is perms, as a token's claims hold them, as Permissions.
func PermissionsOf(
	perms []string,
) []Permission {
	out := make([]Permission, 0, len(perms))
	for _, p := range perms {
		out = append(out, Permission(p))
	}

	return out
}

// builtinPermissions are the permissions of the hand-written domains under
// internal/controller/api, which go as each becomes a behavior, and
// health:read and trace:read, which several endpoint groups name with
// Permission. A hand-written domain lists each permission it needs, also one
// that a group or a behavior declares too, so it keeps it when they change. An endpoint group's own permissions and a behavior's
// come from their definitions.
var builtinPermissions = []Permission{
	"node:read",
	"file:read",
	"file:write",
	"health:read",
	"trace:read",
	"command:execute",
	"command:shell",
	"power:execute",
	"log:read",
}

// adminVerbs are the verbs of a permission that only the admin role gets:
// to open a shell.
var adminVerbs = []string{"shell"}

// builtinAdminOnly are the permissions of the hand-written domains that only
// the admin role gets, because each reaches root directly: an arbitrary
// command (command:execute). The user and sshkey behaviors
// mark their writes AdminOnly, for an account or an SSH key (user:write).
// A behavior or a group that takes one over marks its methods AdminOnly,
// and Check fails until it does.
var builtinAdminOnly = []Permission{
	"command:execute",
}

// declared is a permission that something needs: a hand-written domain, a
// method of an endpoint group, or a method of a behavior.
type declared struct {
	permission Permission
	// adminOnly is true when the method is marked AdminOnly.
	adminOnly bool
	// named is true for an endpoint method that names its permission, which
	// must be one that something else declares.
	named bool
	where string
}

// declarations are the permissions of the hand-written domains, then those
// of each registered DSL whose definitions serve the API, from dsl.All. They are
// made on each call, not once, because a group or a behavior registers in
// the init of its own package, which can run after this package's.
func declarations() []declared {
	out := make([]declared, 0, len(builtinPermissions))
	for _, p := range builtinPermissions {
		out = append(out, declared{
			permission: p,
			adminOnly:  slices.Contains(builtinAdminOnly, p),
			where:      "authtoken",
		})
	}
	for _, d := range dsl.All() {
		if !d.ServesAPI() {
			continue
		}
		for _, p := range d.Permissions() {
			out = append(out, declared{
				permission: p.Permission,
				adminOnly:  p.AdminOnly,
				named:      p.Named,
				where:      p.Where,
			})
		}
	}

	return out
}

// AllPermissions is every permission osapi knows: each one that the
// hand-written domains, an endpoint group or a behavior declares, once. A
// permission that an endpoint method only names is not one.
func AllPermissions() []Permission {
	var all []Permission
	for _, d := range declarations() {
		if !d.named && !slices.Contains(all, d.permission) {
			all = append(all, d.permission)
		}
	}

	return all
}

// PermissionRule is the validate rule of a permission: one that osapi
// knows, such as the permissions of a token or of a custom role.
const PermissionRule = "permission"

func init() {
	rule.Declare(rule.Custom{
		Tag:     PermissionRule,
		Check:   func(p string) bool { return slices.Contains(AllPermissions(), Permission(p)) },
		Schema:  rule.None,
		Message: "is not a permission: osapi token roles get --name admin lists each one",
	})
}

// The built-in roles, one for each kind of person or program that uses
// osapi. The SDK's Role constants are written from Roles.
const (
	// RoleAdmin is the owner of the fleet: each permission.
	RoleAdmin Role = "admin"
	// RoleOperator is the person on call: each read, and each action that
	// makes something happen (execute), such as a restart, a signal or a
	// reboot, but no change to the configuration.
	RoleOperator Role = "operator"
	// RoleAutomation is CI or a workflow that changes the configuration:
	// each read and each write, but no action.
	RoleAutomation Role = "automation"
	// RoleAuditor reads the audit log, the traces and the jobs, and the
	// health. It changes and runs nothing.
	RoleAuditor Role = "auditor"
	// RoleRead reads each thing that is not admin only. Logs, files and
	// process arguments can hold secrets, and read has them.
	RoleRead Role = "read"
	// RoleObserver is a dashboard or a monitor: the health, the status of
	// each host, and the agents with their facts. It reads no file, log or
	// command output.
	RoleObserver Role = "observer"
)

// Roles are the built-in roles.
func Roles() []Role {
	return []Role{RoleAdmin, RoleOperator, RoleAutomation, RoleAuditor, RoleRead, RoleObserver}
}

// auditorPermissions and observerPermissions are the roles that a list
// defines, not a verb: what an auditor and an observer read.
var (
	auditorPermissions  = []Permission{"audit:read", "trace:read", "job:read", "health:read"}
	observerPermissions = []Permission{"health:read", "status:read", "agent:read"}
)

// AdminOnly reports whether only the admin role grants p: its verb is
// shell, the hand-written domain that needs it says so, or a method that
// needs it is marked AdminOnly.
func AdminOnly(
	p Permission,
) bool {
	_, verb, _ := strings.Cut(string(p), ":")
	if slices.Contains(adminVerbs, verb) {
		return true
	}
	for _, d := range declarations() {
		if d.permission == p && d.adminOnly {
			return true
		}
	}

	return false
}

// RolePermissions is what a built-in role grants, and whether role is one.
// admin grants each permission. operator grants each read and each execute
// permission that is not admin only, and automation each read and each
// write that is not admin only. read grants each read that is not admin
// only. auditor and observer grant their lists, of the permissions that
// osapi knows. Thus a new behavior's permissions go to the roles of their
// verbs, and nobody adds them to a role by hand.
func RolePermissions(
	role Role,
) ([]Permission, bool) {
	var grants func(p Permission, verb string) bool
	switch role {
	case RoleAdmin:
		return AllPermissions(), true
	case RoleOperator:
		grants = func(p Permission, verb string) bool {
			return (verb == "read" || verb == "execute") && !AdminOnly(p)
		}
	case RoleAutomation:
		grants = func(p Permission, verb string) bool {
			return (verb == "read" || verb == "write") && !AdminOnly(p)
		}
	case RoleRead:
		grants = func(p Permission, verb string) bool { return verb == "read" && !AdminOnly(p) }
	case RoleAuditor:
		grants = func(p Permission, _ string) bool { return slices.Contains(auditorPermissions, p) }
	case RoleObserver:
		grants = func(p Permission, _ string) bool { return slices.Contains(observerPermissions, p) }
	default:
		return nil, false
	}

	var out []Permission
	for _, p := range AllPermissions() {
		_, verb, _ := strings.Cut(string(p), ":")
		if grants(p, verb) {
			out = append(out, p)
		}
	}

	return out, true
}

// Check is each problem with the declared permissions: a method that names a
// permission nothing declares, which no role grants, so each call gets 403;
// and a permission that one method marks AdminOnly and another does not.
func Check() []error {
	all := AllPermissions()
	decls := declarations()
	var errs []error
	for _, d := range decls {
		if d.named && !slices.Contains(all, d.permission) {
			errs = append(errs, fmt.Errorf(
				"%s names %s, which nothing declares: no role grants it", d.where, d.permission))
		}
	}
	for _, p := range all {
		var marked, unmarked []string
		for _, d := range decls {
			if d.permission != p || d.where == "authtoken" {
				continue
			}
			if d.adminOnly {
				marked = append(marked, d.where)
			} else {
				unmarked = append(unmarked, d.where)
			}
		}
		if len(marked) > 0 && len(unmarked) > 0 {
			errs = append(errs, fmt.Errorf(
				"%s is AdminOnly on %s but not on %s",
				p, strings.Join(marked, ", "), strings.Join(unmarked, ", ")))
		}
	}

	return errs
}

// ResolvePermissions computes the effective permission set for a token.
// If directPermissions is non-empty, it is returned directly (IdP override).
// Otherwise roles are expanded through customRoles first, then RolePermissions.
func ResolvePermissions(
	roles []Role,
	directPermissions []Permission,
	customRoles CustomRoles,
) map[Permission]bool {
	if len(directPermissions) > 0 {
		set := make(map[Permission]bool, len(directPermissions))
		for _, p := range directPermissions {
			set[p] = true
		}
		return set
	}

	set := make(map[Permission]bool)
	for _, role := range roles {
		// Try custom roles first
		if customRoles != nil {
			if perms, ok := customRoles[role]; ok {
				for _, p := range perms {
					set[p] = true
				}
				continue
			}
		}
		// Fall back to default role permissions
		if perms, ok := RolePermissions(role); ok {
			for _, p := range perms {
				set[p] = true
			}
		}
	}
	return set
}

// HasPermission checks whether the resolved set contains the required permission.
func HasPermission(
	resolved map[Permission]bool,
	required Permission,
) bool {
	return resolved[required]
}

// Strings is perms as strings, for a token's claims and the output of a
// command, which hold a permission as text.
func Strings(
	perms []Permission,
) []string {
	out := make([]string, 0, len(perms))
	for _, p := range perms {
		out = append(out, string(p))
	}

	return out
}

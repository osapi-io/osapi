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
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// Permission is a type alias for types.Permission.
type Permission = types.Permission

// Role is a type alias for types.Role: a built-in role or a custom role of
// the config.
type Role = types.Role

// CustomRoles are the custom roles of the config, each with its
// permissions.
type CustomRoles map[Role][]Permission

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

// builtinPermissions are the permissions that no method declares:
// health:read and trace:read, which endpoint groups name with Permission,
// secret:use, which a request needs for each secret://<vault>/<key>
// reference that it gives, and secret:reveal, which secret reveal names
// and marks AdminOnly: the one method that gives the value of a secret. An endpoint group's own permissions and a
// behavior's come from their definitions.
var builtinPermissions = []Permission{
	"health:read",
	"trace:read",
	"secret:use",
	"secret:reveal",
}

// declared is a permission that something needs: a built-in permission, a
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

// declarations are the built-in permissions, then those of each registered
// DSL whose definitions serve the API, from dsl.All. They are made on each
// call, not once, because a group or a behavior registers in the init of its
// own package, which can run after this package's.
func declarations() []declared {
	out := make([]declared, 0, len(builtinPermissions))
	for _, p := range builtinPermissions {
		out = append(out, declared{permission: p, where: "authtoken"})
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

// AllPermissions is every permission osapi knows: each built-in one, and
// each one that an endpoint group or a behavior declares, once. A
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
		Message: "is not a permission: osapi controller role get --name admin lists each one",
	})
}

// roleDescriptions are what each built-in role is for, in one line, as the
// UI shows it.
var roleDescriptions = map[Role]string{
	types.RoleAdmin:      "Everything, including admission, tokens, shells and the audit log",
	types.RoleOperator:   "Reads, and actions such as a restart, a signal or a reboot",
	types.RoleAutomation: "Reads, and changes to the configuration, with no actions",
	types.RoleAuditor:    "The audit log, traces, jobs and health, read only",
	types.RoleRead:       "Reads each thing that is not admin only",
	types.RoleObserver:   "Health, status and agents only, for dashboards",
}

// RoleDescription is what a built-in role is for, in one line, or "" for a
// custom role.
func RoleDescription(
	r Role,
) string {
	return roleDescriptions[r]
}

// Roles are the built-in roles.
func Roles() []Role {
	return []Role{
		types.RoleAdmin,
		types.RoleOperator,
		types.RoleAutomation,
		types.RoleAuditor,
		types.RoleRead,
		types.RoleObserver,
	}
}

// auditorPermissions and observerPermissions are the roles that a list
// defines, not a verb: what an auditor and an observer read.
var (
	auditorPermissions = []Permission{
		"audit:read", "trace:read", "job:read", "health:read", "catalog:read",
	}
	observerPermissions = []Permission{"health:read", "status:read", "agent:read", "catalog:read"}
)

// AdminOnly reports whether only the admin role grants p: its verb is
// shell or exec, or a method that needs it is marked AdminOnly.
func AdminOnly(
	p Permission,
) bool {
	_, verb, _ := strings.Cut(string(p), ":")
	if types.Verb(verb).AdminOnly() {
		return true
	}
	for _, d := range declarations() {
		if d.permission == p && d.adminOnly {
			return true
		}
	}

	return false
}

// operatorWrites are the writes that are actions of the operator, not
// changes to a host's configuration: a job that the operator ran and that
// failed is the operator's to retry or delete (osapi-io/osapi#1409).
var operatorWrites = []Permission{"job:write"}

// RolePermissions is what a built-in role grants, and whether role is one.
// admin grants each permission. operator grants each read and each execute
// permission that is not admin only, and operatorWrites, and automation each read and each
// write that is not admin only. read grants each read that is not admin
// only. auditor and observer grant their lists, of the permissions that
// osapi knows. Thus a new behavior's permissions go to the roles of their
// verbs, and nobody adds them to a role by hand.
func RolePermissions(
	role Role,
) ([]Permission, bool) {
	var grants func(p Permission, verb string) bool
	switch role {
	case types.RoleAdmin:
		return AllPermissions(), true
	case types.RoleOperator:
		grants = func(p Permission, verb string) bool {
			return (verb == "read" || verb == "execute" || slices.Contains(operatorWrites, p)) &&
				!AdminOnly(p)
		}
	case types.RoleAutomation:
		// A secret goes into a write, so the role that writes uses one.
		grants = func(p Permission, verb string) bool {
			return (verb == "read" || verb == "write" || verb == "use") && !AdminOnly(p)
		}
	case types.RoleRead:
		grants = func(p Permission, verb string) bool { return verb == "read" && !AdminOnly(p) }
	case types.RoleAuditor:
		grants = func(p Permission, _ string) bool { return slices.Contains(auditorPermissions, p) }
	case types.RoleObserver:
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

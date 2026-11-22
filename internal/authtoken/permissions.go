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
	"slices"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/pkg/sdk/client"
)

// Permission is a type alias for client.Permission.
type Permission = client.Permission

// builtinPermissions are the hand-written domains' permissions. They are the
// source of the SDK's constants, which go generate ./internal/behaviors/all
// writes from AllPermissions. A behavior's
// are not listed here: AllPermissions adds them from its definition.
var builtinPermissions = []Permission{
	"admission:read",
	"admission:write",
	"agent:read",
	"agent:write",
	"node:read",
	"network:read",
	"network:write",
	"job:read",
	"job:write",
	"health:read",
	"audit:read",
	"trace:read",
	"command:execute",
	"command:shell",
	"file:read",
	"file:write",
	"docker:read",
	"docker:write",
	"docker:execute",
	"schedule:read",
	"schedule:write",
	"sysctl:read",
	"sysctl:write",
	"ntp:read",
	"ntp:write",
	"timezone:read",
	"timezone:write",
	"power:execute",
	"process:read",
	"process:execute",
	"user:read",
	"user:write",
	"package:read",
	"package:write",
	"log:read",
	"certificate:read",
	"certificate:write",
	"service:read",
	"service:write",
}

// builtinRolePermissions maps built-in role names to the hand-written domains'
// permissions each grants. RolePermissions adds each behavior's.
var builtinRolePermissions = map[string][]Permission{
	client.RoleAdmin: {
		// Admitting a node is a trust decision: only admin may make it.
		"admission:read",
		"admission:write",
		"agent:read",
		"agent:write",
		"node:read",
		"network:read",
		"network:write",
		"job:read",
		"job:write",
		"health:read",
		"trace:read",
		"audit:read",
		"command:execute",
		"command:shell",
		"file:read",
		"file:write",
		"docker:read",
		"docker:write",
		"docker:execute",
		"schedule:read",
		"schedule:write",
		"sysctl:read",
		"sysctl:write",
		"ntp:read",
		"ntp:write",
		"timezone:read",
		"timezone:write",
		"power:execute",
		"process:read",
		"process:execute",
		"user:read",
		"user:write",
		"package:read",
		"package:write",
		"log:read",
		"certificate:read",
		"certificate:write",
		"service:read",
		"service:write",
	},
	client.RoleWrite: {
		"admission:read",
		"agent:read",
		"node:read",
		"network:read",
		"network:write",
		"job:read",
		"job:write",
		"health:read",
		"trace:read",
		"file:read",
		"file:write",
		"docker:read",
		"docker:write",
		"schedule:read",
		"schedule:write",
		"sysctl:read",
		"sysctl:write",
		"ntp:read",
		"ntp:write",
		"timezone:read",
		"timezone:write",
		"process:read",
		"user:read",
		"user:write",
		"package:read",
		"package:write",
		"log:read",
		"certificate:read",
		"certificate:write",
		"service:read",
		"service:write",
	},
	client.RoleRead: {
		"admission:read",
		"agent:read",
		"node:read",
		"network:read",
		"job:read",
		"health:read",
		"trace:read",
		"file:read",
		"docker:read",
		"schedule:read",
		"sysctl:read",
		"ntp:read",
		"timezone:read",
		"process:read",
		"user:read",
		"package:read",
		"log:read",
		"certificate:read",
		"service:read",
	},
}

// AllPermissions is every permission osapi knows: the hand-written domains',
// then each registered behavior's, one per verb its methods use. It is built
// on each call rather than once, because behaviors register in their own
// packages' init, which may run after this package's.
func AllPermissions() []Permission {
	all := slices.Clone(builtinPermissions)
	for _, p := range behaviorPermissions(func(behavior.Verb) bool { return true }) {
		if !slices.Contains(all, p) {
			all = append(all, p)
		}
	}

	return all
}

// RolePermissions is what a built-in role grants, and whether role is one.
// admin and write get every behavior permission; read gets each behavior's
// read permission.
func RolePermissions(
	role string,
) ([]Permission, bool) {
	perms, ok := builtinRolePermissions[role]
	if !ok {
		return nil, false
	}

	grants := func(behavior.Verb) bool { return true }
	if role == client.RoleRead {
		grants = func(v behavior.Verb) bool { return v == behavior.VerbRead }
	}

	out := slices.Clone(perms)
	for _, p := range behaviorPermissions(grants) {
		if !slices.Contains(out, p) {
			out = append(out, p)
		}
	}

	return out, true
}

// behaviorPermissions is the permission of every registered behavior's
// methods whose verb keep accepts, without repeats.
func behaviorPermissions(
	keep func(behavior.Verb) bool,
) []Permission {
	var out []Permission
	for _, b := range behavior.All() {
		// Only a behavior served over the API has methods a token can call.
		if !b.API {
			continue
		}
		for _, m := range b.Methods {
			p := b.Permission(m.Verb)
			if keep(m.Verb) && !slices.Contains(out, p) {
				out = append(out, p)
			}
		}
	}

	return out
}

// ResolvePermissions computes the effective permission set for a token.
// If directPermissions is non-empty, it is returned directly (IdP override).
// Otherwise roles are expanded through customRoles first, then RolePermissions.
func ResolvePermissions(
	roles []string,
	directPermissions []string,
	customRoles map[string][]string,
) map[string]bool {
	if len(directPermissions) > 0 {
		set := make(map[string]bool, len(directPermissions))
		for _, p := range directPermissions {
			set[p] = true
		}
		return set
	}

	set := make(map[string]bool)
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
	resolved map[string]bool,
	required string,
) bool {
	return resolved[required]
}

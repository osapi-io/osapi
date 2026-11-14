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

// Permission constants re-exported from the SDK.
const (
	PermAgentRead        = client.PermAgentRead
	PermAgentWrite       = client.PermAgentWrite
	PermNodeRead         = client.PermNodeRead
	PermNodeWrite        = client.PermNodeWrite
	PermNetworkRead      = client.PermNetworkRead
	PermNetworkWrite     = client.PermNetworkWrite
	PermJobRead          = client.PermJobRead
	PermJobWrite         = client.PermJobWrite
	PermHealthRead       = client.PermHealthRead
	PermAuditRead        = client.PermAuditRead
	PermCommandExecute   = client.PermCommandExecute
	PermCommandShell     = client.PermCommandShell
	PermFileRead         = client.PermFileRead
	PermFileWrite        = client.PermFileWrite
	PermDockerRead       = client.PermDockerRead
	PermDockerWrite      = client.PermDockerWrite
	PermDockerExecute    = client.PermDockerExecute
	PermCronRead         = client.PermCronRead
	PermCronWrite        = client.PermCronWrite
	PermSysctlRead       = client.PermSysctlRead
	PermSysctlWrite      = client.PermSysctlWrite
	PermNtpRead          = client.PermNtpRead
	PermNtpWrite         = client.PermNtpWrite
	PermTimezoneRead     = client.PermTimezoneRead
	PermTimezoneWrite    = client.PermTimezoneWrite
	PermPowerExecute     = client.PermPowerExecute
	PermProcessRead      = client.PermProcessRead
	PermProcessExecute   = client.PermProcessExecute
	PermUserRead         = client.PermUserRead
	PermUserWrite        = client.PermUserWrite
	PermPackageRead      = client.PermPackageRead
	PermPackageWrite     = client.PermPackageWrite
	PermLogRead          = client.PermLogRead
	PermCertificateRead  = client.PermCertificateRead
	PermCertificateWrite = client.PermCertificateWrite
	PermServiceRead      = client.PermServiceRead
	PermServiceWrite     = client.PermServiceWrite
)

// builtinPermissions are the hand-written domains' permissions. A behavior's
// are not listed here: AllPermissions adds them from its definition.
var builtinPermissions = []Permission{
	PermAgentRead,
	PermAgentWrite,
	PermNodeRead,
	PermNodeWrite,
	PermNetworkRead,
	PermNetworkWrite,
	PermJobRead,
	PermJobWrite,
	PermHealthRead,
	PermAuditRead,
	PermCommandExecute,
	PermCommandShell,
	PermFileRead,
	PermFileWrite,
	PermDockerRead,
	PermDockerWrite,
	PermDockerExecute,
	PermCronRead,
	PermCronWrite,
	PermSysctlRead,
	PermSysctlWrite,
	PermNtpRead,
	PermNtpWrite,
	PermTimezoneRead,
	PermTimezoneWrite,
	PermPowerExecute,
	PermProcessRead,
	PermProcessExecute,
	PermUserRead,
	PermUserWrite,
	PermPackageRead,
	PermPackageWrite,
	PermLogRead,
	PermCertificateRead,
	PermCertificateWrite,
	PermServiceRead,
	PermServiceWrite,
}

// builtinRolePermissions maps built-in role names to the hand-written domains'
// permissions each grants. RolePermissions adds each behavior's.
var builtinRolePermissions = map[string][]Permission{
	client.RoleAdmin: {
		PermAgentRead,
		PermAgentWrite,
		PermNodeRead,
		PermNodeWrite,
		PermNetworkRead,
		PermNetworkWrite,
		PermJobRead,
		PermJobWrite,
		PermHealthRead,
		PermAuditRead,
		PermCommandExecute,
		PermCommandShell,
		PermFileRead,
		PermFileWrite,
		PermDockerRead,
		PermDockerWrite,
		PermDockerExecute,
		PermCronRead,
		PermCronWrite,
		PermSysctlRead,
		PermSysctlWrite,
		PermNtpRead,
		PermNtpWrite,
		PermTimezoneRead,
		PermTimezoneWrite,
		PermPowerExecute,
		PermProcessRead,
		PermProcessExecute,
		PermUserRead,
		PermUserWrite,
		PermPackageRead,
		PermPackageWrite,
		PermLogRead,
		PermCertificateRead,
		PermCertificateWrite,
		PermServiceRead,
		PermServiceWrite,
	},
	client.RoleWrite: {
		PermAgentRead,
		PermNodeRead,
		PermNodeWrite,
		PermNetworkRead,
		PermNetworkWrite,
		PermJobRead,
		PermJobWrite,
		PermHealthRead,
		PermFileRead,
		PermFileWrite,
		PermDockerRead,
		PermDockerWrite,
		PermCronRead,
		PermCronWrite,
		PermSysctlRead,
		PermSysctlWrite,
		PermNtpRead,
		PermNtpWrite,
		PermTimezoneRead,
		PermTimezoneWrite,
		PermProcessRead,
		PermUserRead,
		PermUserWrite,
		PermPackageRead,
		PermPackageWrite,
		PermLogRead,
		PermCertificateRead,
		PermCertificateWrite,
		PermServiceRead,
		PermServiceWrite,
	},
	client.RoleRead: {
		PermAgentRead,
		PermNodeRead,
		PermNetworkRead,
		PermJobRead,
		PermHealthRead,
		PermFileRead,
		PermDockerRead,
		PermCronRead,
		PermSysctlRead,
		PermNtpRead,
		PermTimezoneRead,
		PermProcessRead,
		PermUserRead,
		PermPackageRead,
		PermLogRead,
		PermCertificateRead,
		PermServiceRead,
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

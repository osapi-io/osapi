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

import "github.com/retr0h/osapi/pkg/sdk/client"

// Permission is a type alias for client.Permission.
type Permission = client.Permission

// Permission constants re-exported from the SDK.
const (
	PermAgentRead      = client.PermAgentRead
	PermAgentWrite     = client.PermAgentWrite
	PermNodeRead       = client.PermNodeRead
	PermNodeWrite      = client.PermNodeWrite
	PermNetworkRead    = client.PermNetworkRead
	PermNetworkWrite   = client.PermNetworkWrite
	PermJobRead        = client.PermJobRead
	PermJobWrite       = client.PermJobWrite
	PermHealthRead     = client.PermHealthRead
	PermAuditRead      = client.PermAuditRead
	PermCommandExecute = client.PermCommandExecute
	PermFileRead       = client.PermFileRead
	PermFileWrite      = client.PermFileWrite
	PermDockerRead     = client.PermDockerRead
	PermDockerWrite    = client.PermDockerWrite
	PermDockerExecute  = client.PermDockerExecute
	PermCronRead       = client.PermCronRead
	PermCronWrite      = client.PermCronWrite
	PermSysctlRead     = client.PermSysctlRead
	PermSysctlWrite    = client.PermSysctlWrite
	PermNtpRead        = client.PermNtpRead
	PermNtpWrite       = client.PermNtpWrite
	PermTimezoneRead   = client.PermTimezoneRead
	PermTimezoneWrite  = client.PermTimezoneWrite
)

// AllPermissions is the full set of known permissions.
var AllPermissions = []Permission{
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
}

// DefaultRolePermissions maps built-in role names to their granted permissions.
var DefaultRolePermissions = map[string][]Permission{
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
	},
}

// ResolvePermissions computes the effective permission set for a token.
// If directPermissions is non-empty, it is returned directly (IdP override).
// Otherwise roles are expanded through customRoles first, then DefaultRolePermissions.
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
		if perms, ok := DefaultRolePermissions[role]; ok {
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

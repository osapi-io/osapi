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

// Permission represents a fine-grained resource:verb permission.
type Permission = string

// Permission constants using resource:verb format.
const (
	PermSystemRead   Permission = "system:read"
	PermNetworkRead  Permission = "network:read"
	PermNetworkWrite Permission = "network:write"
	PermJobRead      Permission = "job:read"
	PermJobWrite     Permission = "job:write"
	PermHealthRead   Permission = "health:read"
	PermAuditRead    Permission = "audit:read"
)

// AllPermissions is the full set of known permissions.
var AllPermissions = []Permission{
	PermSystemRead,
	PermNetworkRead,
	PermNetworkWrite,
	PermJobRead,
	PermJobWrite,
	PermHealthRead,
	PermAuditRead,
}

// DefaultRolePermissions maps built-in role names to their granted permissions.
var DefaultRolePermissions = map[string][]Permission{
	"admin": {
		PermSystemRead,
		PermNetworkRead,
		PermNetworkWrite,
		PermJobRead,
		PermJobWrite,
		PermHealthRead,
		PermAuditRead,
	},
	"write": {
		PermSystemRead,
		PermNetworkRead,
		PermNetworkWrite,
		PermJobRead,
		PermJobWrite,
		PermHealthRead,
	},
	"read": {
		PermSystemRead,
		PermNetworkRead,
		PermJobRead,
		PermHealthRead,
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

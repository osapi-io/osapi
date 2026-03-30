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

package client

// Permission represents a fine-grained resource:verb permission.
// Permissions use the format "resource:verb" (e.g., "node:read",
// "docker:execute").
type Permission = string

// Permission constants.
const (
	PermAgentRead      Permission = "agent:read"
	PermAgentWrite     Permission = "agent:write"
	PermNodeRead       Permission = "node:read"
	PermNodeWrite      Permission = "node:write"
	PermNetworkRead    Permission = "network:read"
	PermNetworkWrite   Permission = "network:write"
	PermJobRead        Permission = "job:read"
	PermJobWrite       Permission = "job:write"
	PermHealthRead     Permission = "health:read"
	PermAuditRead      Permission = "audit:read"
	PermCommandExecute Permission = "command:execute"
	PermFileRead       Permission = "file:read"
	PermFileWrite      Permission = "file:write"
	PermDockerRead     Permission = "docker:read"
	PermDockerWrite    Permission = "docker:write"
	PermDockerExecute  Permission = "docker:execute"
	PermCronRead       Permission = "cron:read"
	PermCronWrite      Permission = "cron:write"
	PermSysctlRead     Permission = "sysctl:read"
	PermSysctlWrite    Permission = "sysctl:write"
)

// Role represents a built-in RBAC role name.
type Role = string

// Built-in role constants.
const (
	RoleAdmin Role = "admin"
	RoleWrite Role = "write"
	RoleRead  Role = "read"
)

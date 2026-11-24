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

package role

// The sample outputs of the examples, which the docs page of each command
// shows as the CLI prints it.
var (
	sampleRoles = ListOutput{Roles: []Role{
		{Name: "admin", Source: "built-in"},
		{Name: "operator", Source: "built-in"},
		{Name: "automation", Source: "built-in"},
		{Name: "auditor", Source: "built-in"},
		{Name: "read", Source: "built-in"},
		{Name: "observer", Source: "built-in"},
		{Name: "dns-ops", Source: "config"},
	}}

	sampleRole = Role{
		Name:        "admin",
		Source:      "built-in",
		Permissions: []string{"status:read", "file:read", "file:write", "health:read"},
	}
)

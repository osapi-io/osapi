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

package types

// The built-in roles, one for each kind of person or program that uses
// osapi. internal/authtoken gives each its permissions.
const (
	// RoleAdmin is the owner of the fleet: each permission.
	RoleAdmin Role = "admin"
	// RoleOperator is the person on call: each read, and each action that
	// makes something happen (execute), such as a restart, a signal or a
	// reboot, but no change to the configuration.
	RoleOperator Role = "operator"
	// RoleAutomation is CI or a workflow that changes the configuration:
	// each read and each write, and the use of a secret in a write, but no
	// action.
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

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

import "slices"

// Verb is the verb of a permission, the part after the colon of
// resource:verb. It also says whether a method of a behavior reads the host
// or changes it.
type Verb string

// The verbs of a permission.
const (
	// VerbRead reads and changes nothing.
	VerbRead Verb = "read"
	// VerbWrite changes something. The output of a write of a behavior
	// reports whether it did.
	VerbWrite Verb = "write"
	// VerbExecute is the permission verb of an action, such as a restart or
	// a reboot: it makes something happen on the host, but it does not change
	// what the host is configured to be. The operator role gets it and the
	// automation role does not.
	VerbExecute Verb = "execute"
	// VerbShell is the permission verb of an action that runs a shell
	// command line, which reaches anything the host can do. Only the admin
	// role gets a shell permission.
	VerbShell Verb = "shell"
	// VerbExec is the permission verb of an action that runs a command that
	// the caller gives inside something on the host, such as docker exec in
	// a container. Only the admin role gets an exec permission.
	VerbExec Verb = "exec"
)

// AdminOnly reports whether only the admin role gets a permission of verb
// v: shell and exec reach anything that the host can do.
func (v Verb) AdminOnly() bool {
	return v == VerbShell || v == VerbExec
}

// NewPermission is the permission of verb v on resource, as resource:verb,
// such as sysctl:read.
func NewPermission(
	resource string,
	v Verb,
) Permission {
	return Permission(resource + ":" + string(v))
}

// MethodKind is what a method of a behavior does to a host, by the verb of
// its permission. It is a separate set from Verb: execute, shell and exec
// are each an action.
type MethodKind string

// The kinds of a method.
const (
	// KindRead changes nothing.
	KindRead MethodKind = "read"
	// KindWrite changes what the host is configured to be.
	KindWrite MethodKind = "write"
	// KindAction makes something happen on the host, such as a restart or a
	// reboot.
	KindAction MethodKind = "action"
)

// Values are each kind of a method.
func (MethodKind) Values() []MethodKind {
	return []MethodKind{KindRead, KindWrite, KindAction}
}

// Valid is whether k is one of Values.
func (k MethodKind) Valid() bool {
	return slices.Contains(k.Values(), k)
}

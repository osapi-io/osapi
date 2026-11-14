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

// Package managedfile is the managedfile behavior: files osapi writes on a
// host and keeps a record of, so it can tell its own files from the rest,
// leave a file alone when it already holds what was asked, and stop managing
// one. It is internal: other behaviors require it, such as sysctl, and call
// its methods on the same host.
package managedfile

import "github.com/osapi-io/osapi/internal/behavior"

// path and token are where the examples write.
const (
	path  = "/etc/osapi/example.conf"
	token = "/etc/myapp/token"
)

// Behavior is the managedfile definition.
var Behavior = behavior.Register(behavior.Define(
	"managedfile", "Files osapi writes on a host and keeps a record of",
	behavior.Node,
	behavior.Internal,
	behavior.Platform("debian", newDebian),
	behavior.Sudo("chown"),
	behavior.Runs("stat"),
	behavior.Read("get", Get).
		Describe("Read the record of a file osapi wrote").
		Example("A file osapi wrote", GetInput{Path: path},
			Record{
				Path:     path,
				SHA256:   "b6ac5bc39377e44ee783dff946f6703a994f5eb31a438e69dc0735b40060d2e8",
				Metadata: map[string]string{"setting": "on"},
			},
			behavior.After(UpdateInput{Path: path, Content: "on\n", Metadata: map[string]string{"setting": "on"}}),
		).
		Example("A file osapi stopped managing", GetInput{Path: path},
			Record{
				Path:       path,
				SHA256:     "b6ac5bc39377e44ee783dff946f6703a994f5eb31a438e69dc0735b40060d2e8",
				Metadata:   map[string]string{"setting": "on"},
				Undeployed: true,
			},
			behavior.After(UpdateInput{Path: path, Content: "on\n", Metadata: map[string]string{"setting": "on"}}),
			behavior.After(DeleteInput{Path: path}),
		).
		Fails("A file osapi never wrote", GetInput{Path: path}, "not managed"),
	behavior.Write("update", Update).
		Describe("Write a file and keep a record of it").
		Explain("Leaves a file already holding the content alone and reports changed: false, "+
			"whatever its record says, so a file edited by hand is put back.").
		Example("A new file", UpdateInput{Path: path, Content: "on\n"},
			Result{Path: path, Changed: true},
			behavior.Wrote(path, "on\n"),
		).
		Example("A secret only root reads",
			UpdateInput{
				Path: token, Content: "s3cret\n", Mode: "0600",
				Owner: "root", Group: "root", Secret: true,
			},
			Result{Path: token, Changed: true},
			behavior.Prints("stat -c %U:%G -- "+token, "osapi:osapi"),
			behavior.PrintsAfter("chown -- root:root "+token, "stat -c %U:%G -- "+token, "root:root"),
			behavior.Wrote(token, "s3cret\n"),
			behavior.Ran("chown", "--", "root:root", token),
		).
		Fails("A mode that is not octal",
			UpdateInput{Path: path, Content: "on\n", Mode: "rw-------"}, "invalid mode").
		Fails("An owner that is not a user name",
			UpdateInput{Path: path, Content: "on\n", Owner: "root;reboot"}, "invalid owner").
		Example("A file whose mode was changed by hand",
			UpdateInput{Path: path, Content: "on\n", Mode: "0600"},
			Result{Path: path, Changed: true},
			behavior.Exists(path, "on\n"),
			behavior.Wrote(path, "on\n"),
		).
		Example("A file edited by hand", UpdateInput{Path: path, Content: "on\n"},
			Result{Path: path, Changed: true},
			behavior.After(UpdateInput{Path: path, Content: "on\n"}),
			behavior.Exists(path, "off\n"),
			behavior.Wrote(path, "on\n"),
		),
	behavior.Write("delete", Delete).
		Describe("Remove a file osapi wrote and stop managing it").
		Explain("Leaves a file osapi has no record of alone, so it never removes what "+
			"it did not write.").
		Example("A file osapi wrote", DeleteInput{Path: path},
			Result{Path: path, Changed: true},
			behavior.After(UpdateInput{Path: path, Content: "on\n"}),
			behavior.Removed(path),
		).
		Example("A file osapi never wrote", DeleteInput{Path: path},
			Result{Path: path},
			behavior.Exists(path, "theirs\n"),
		),
))

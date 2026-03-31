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

// ExecRequest contains parameters for direct command execution.
type ExecRequest struct {
	// Command is the binary to execute (required).
	Command string

	// Args is the argument list passed to the command.
	Args []string

	// Cwd is the working directory. Empty uses the agent default.
	Cwd string

	// Timeout in seconds. Zero uses the server default (30s).
	Timeout int

	// Target specifies the host: "_any", "_all", hostname, or
	// label ("group:web").
	Target string
}

// FileDeployOpts contains parameters for file deployment.
type FileDeployOpts struct {
	// ObjectName is the name of the file in the Object Store (required).
	ObjectName string

	// Path is the destination path on the target filesystem (required).
	Path string

	// ContentType is "raw" or "template" (required).
	ContentType string

	// Mode is the file permission mode (e.g., "0644"). Optional.
	Mode string

	// Owner is the file owner user. Optional.
	Owner string

	// Group is the file owner group. Optional.
	Group string

	// Vars are template variables when ContentType is "template". Optional.
	Vars map[string]any

	// Target specifies the host: "_any", "_all", hostname, or
	// label ("group:web").
	Target string
}

// FileUndeployOpts contains parameters for file undeployment.
type FileUndeployOpts struct {
	// Path is the filesystem path to remove from the target host (required).
	Path string

	// Target specifies the host: "_any", "_all", hostname, or
	// label ("group:web").
	Target string
}

// ShellRequest contains parameters for shell command execution.
type ShellRequest struct {
	// Command is the shell command string passed to /bin/sh -c (required).
	Command string

	// Cwd is the working directory. Empty uses the agent default.
	Cwd string

	// Timeout in seconds. Zero uses the server default (30s).
	Timeout int

	// Target specifies the host: "_any", "_all", hostname, or
	// label ("group:web").
	Target string
}

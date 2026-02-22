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

package command

// Provider implements the methods to execute commands on the system.
type Provider interface {
	// Exec executes a command directly without a shell.
	Exec(params ExecParams) (*Result, error)
	// Shell executes a command through /bin/sh -c.
	Shell(params ShellParams) (*Result, error)
}

// ExecParams contains parameters for direct command execution.
type ExecParams struct {
	// Command is the executable name or path.
	Command string
	// Args are the command arguments.
	Args []string
	// Cwd is the optional working directory.
	Cwd string
	// Timeout is the timeout in seconds (0 = default 30s).
	Timeout int
}

// ShellParams contains parameters for shell command execution.
type ShellParams struct {
	// Command is the full shell command string.
	Command string
	// Cwd is the optional working directory.
	Cwd string
	// Timeout is the timeout in seconds (0 = default 30s).
	Timeout int
}

// Result contains the output of a command execution.
type Result struct {
	// Stdout is the standard output.
	Stdout string `json:"stdout"`
	// Stderr is the standard error output.
	Stderr string `json:"stderr"`
	// ExitCode is the process exit code.
	ExitCode int `json:"exit_code"`
	// DurationMs is the execution time in milliseconds.
	DurationMs int64 `json:"duration_ms"`
}

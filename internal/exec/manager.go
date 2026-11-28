// Copyright (c) 2024 John Dewey

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

package exec

import (
	"context"
	"io"
)

// Manager manager responsible for exec operations.
//
// Every method takes a context. A command with no deadline of its own is bounded
// by DefaultCommandTimeout, and one still running when the agent shuts down is
// killed with it rather than outliving it.
type Manager interface {
	// RunCmd executes the provided command with arguments, using the current
	// working directory.
	RunCmd(
		ctx context.Context,
		name string,
		args []string,
	) (string, error)

	// RunPrivilegedCmd executes the provided command with arguments.
	// When sudo is enabled, the command is prepended with "sudo".
	RunPrivilegedCmd(
		ctx context.Context,
		name string,
		args []string,
	) (string, error)

	// RunPrivilegedCmdWithStdin executes the provided command with arguments
	// and writes stdin to its standard input. When sudo is enabled, the
	// command is prepended with "sudo". Use it to pass secrets, which must
	// never be placed in arguments.
	RunPrivilegedCmdWithStdin(
		ctx context.Context,
		name string,
		args []string,
		stdin string,
	) (string, error)

	// RunPrivilegedCmdInteractive runs name with args, through sudo when
	// privilege escalation is on, and calls interact with its standard input
	// while it runs, for a command that waits for an answer that the caller
	// gives later, such as the confirmation of netplan try. The standard
	// input closes when interact returns. The result is the combined output.
	RunPrivilegedCmdInteractive(
		ctx context.Context,
		name string,
		args []string,
		interact func(ctx context.Context, stdin io.Writer) error,
	) (string, error)
	// RunPrivilegedCmdFull runs a command through sudo when privilege
	// escalation is on, with the standard output, the standard error and
	// the exit code apart.
	// stdin, when it is not empty, goes to the command's standard input.
	RunPrivilegedCmdFull(
		ctx context.Context,
		name string,
		args []string,
		stdin string,
	) (string, string, int, error)

	// RunReadCmd runs a command to read the host, as gohai runs one: its
	// standard output is the data and its standard error is in the error,
	// with LC_ALL=C, a timeout and a limit on its output.
	RunReadCmd(
		ctx context.Context,
		name string,
		args []string,
	) ([]byte, error)

	// RunCmdFull executes a command with separate stdout/stderr capture,
	// an optional working directory, and a timeout in seconds.
	RunCmdFull(
		ctx context.Context,
		name string,
		args []string,
		cwd string,
		timeout int,
	) (*CmdResult, error)
}

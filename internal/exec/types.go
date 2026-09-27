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
	"log/slog"
	"time"
)

// DefaultCommandTimeout bounds a command whose caller supplied no deadline of
// its own. It is a backstop rather than a budget: long enough for a package
// installation over a slow mirror, short enough that a command waiting on a
// prompt nobody will answer does not outlive the agent.
//
// A caller's own deadline still wins whenever it falls earlier, because the
// deadline is derived from the context it passed.
const DefaultCommandTimeout = 10 * time.Minute

// CommandExecutor executes OS commands. The default implementation
// runs real commands via os/exec. Tests inject a mock to assert
// commands without executing them.
//
// Every method takes a context: it is what cancels a command that is still
// running when the agent shuts down, and what bounds one that would otherwise
// run forever.
type CommandExecutor interface {
	Execute(
		ctx context.Context,
		name string,
		args []string,
		cwd string,
	) (string, error)

	// ExecuteWithStdin runs the command with stdin written to its standard
	// input. Implementations must not log stdin.
	ExecuteWithStdin(
		ctx context.Context,
		name string,
		args []string,
		cwd string,
		stdin string,
	) (string, error)
}

// Exec disk implementation.
type Exec struct {
	logger   *slog.Logger
	sudo     bool
	executor CommandExecutor
}

// CmdResult contains the full result of a command execution
// with separate stdout and stderr streams.
type CmdResult struct {
	// Stdout contains the standard output of the command.
	Stdout string
	// Stderr contains the standard error output of the command.
	Stderr string
	// ExitCode is the exit code of the command.
	ExitCode int
	// DurationMs is the execution time in milliseconds.
	DurationMs int64
}

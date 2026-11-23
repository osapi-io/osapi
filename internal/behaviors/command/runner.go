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

import (
	"context"
	"fmt"

	"github.com/osapi-io/osapi/internal/behavior/host"
)

// runner runs commands through the agent's executor, as its own user, on
// each platform.
type runner struct {
	exec host.Exec
}

// newRunner is the provider on each platform.
func newRunner(
	h host.Host,
) (Provider, error) {
	return &runner{exec: h.Exec}, nil
}

// Exec runs in.Command with in.Args, with no shell.
func (r *runner) Exec(
	ctx context.Context,
	in ExecInput,
) (Result, error) {
	return r.run(ctx, in.Command, in.Args, in.Cwd, in.Timeout)
}

// Shell runs in.Command through /bin/sh -c.
func (r *runner) Shell(
	ctx context.Context,
	in ShellInput,
) (Result, error) {
	return r.run(ctx, shell, []string{"-c", in.Command}, in.Cwd, in.Timeout)
}

// shell is the program that runs a command line.
const shell = "/bin/sh"

// run runs name with args and is what it did. A command that exits non-zero
// is a result; only a command that cannot start is an error.
func (r *runner) run(
	ctx context.Context,
	name string,
	args []string,
	cwd string,
	timeout int,
) (Result, error) {
	out, err := r.exec.RunCmdFull(ctx, name, args, cwd, timeout)
	if err != nil {
		return Result{}, fmt.Errorf("run %s: %w", name, err)
	}

	return Result{
		Stdout:     out.Stdout,
		Stderr:     out.Stderr,
		ExitCode:   out.ExitCode,
		DurationMs: out.DurationMs,
		Changed:    true,
	}, nil
}

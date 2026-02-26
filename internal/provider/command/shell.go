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
	"fmt"
	"log/slog"
)

// Shell executes a command through /bin/sh -c.
func (c *Executor) Shell(
	params ShellParams,
) (*Result, error) {
	c.logger.Debug("executing shell command",
		slog.String("command", params.Command),
		slog.String("cwd", params.Cwd),
		slog.Int("timeout", params.Timeout),
	)

	cmdResult, err := c.execManager.RunCmdFull(
		"/bin/sh",
		[]string{"-c", params.Command},
		params.Cwd,
		params.Timeout,
	)
	if err != nil {
		return nil, fmt.Errorf("shell execution failed: %w", err)
	}

	return &Result{
		Stdout:     cmdResult.Stdout,
		Stderr:     cmdResult.Stderr,
		ExitCode:   cmdResult.ExitCode,
		DurationMs: cmdResult.DurationMs,
		Changed:    true,
	}, nil
}

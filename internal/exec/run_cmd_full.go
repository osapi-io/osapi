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

package exec

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"time"
)

// RunCmdFull executes the provided command with separate stdout and stderr
// capture, an optional working directory, and a timeout in seconds.
// A timeout of 0 defaults to 30 seconds.
func (e *Exec) RunCmdFull(
	ctx context.Context,
	name string,
	args []string,
	cwd string,
	timeout int,
) (*CmdResult, error) {
	return e.runFull(ctx, name, args, cwd, timeout, "")
}

// runFull is RunCmdFull with stdin written to the command's standard input
// when it is not empty. It never logs stdin, which can hold a secret.
func (e *Exec) runFull(
	ctx context.Context,
	name string,
	args []string,
	cwd string,
	timeout int,
	stdin string,
) (*CmdResult, error) {
	if timeout <= 0 {
		timeout = 30
	}

	ctx, span := startSpan(ctx, name, args)
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := newCommand(ctx, name, args)
	if cwd != "" {
		cmd.Dir = cwd
	}

	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	result := &CmdResult{
		Stdout:     stdout.String(),
		Stderr:     stderr.String(),
		DurationMs: duration.Milliseconds(),
	}
	err = fullResult(ctx, result, err, timeout)
	endSpan(span, cmd, err)

	e.logger.Debug(
		"exec full",
		slog.String("command", strings.Join(cmd.Args, " ")),
		slog.String("cwd", cwd),
		slog.Int("exit_code", result.ExitCode),
		slog.Int64("duration_ms", result.DurationMs),
		slog.Any("error", err),
	)

	return result, err
}

// fullResult sets the exit code of result from the error of Run, and gives
// the error of the run: a timeout, a program that did not start, or nil for
// a command that ran, with any exit code.
func fullResult(
	ctx context.Context,
	result *CmdResult,
	err error,
	timeout int,
) error {
	if err == nil {
		return nil
	}
	if ctx.Err() == context.DeadlineExceeded {
		result.ExitCode = -1

		return fmt.Errorf("command timed out after %ds", timeout)
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		result.ExitCode = exitErr.ExitCode()

		return nil
	}

	return fmt.Errorf("failed to execute command: %w", err)
}

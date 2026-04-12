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

// Package exec provides command execution utilities.
package exec

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
)

const maxLogOutputLen = 200

// defaultExecutor implements CommandExecutor by running real OS commands
// via os/exec.
type defaultExecutor struct {
	logger *slog.Logger
}

// Execute runs the command and returns its combined output.
func (d *defaultExecutor) Execute(
	name string,
	args []string,
	cwd string,
) (string, error) {
	cmd := exec.Command(name, args...)
	if cwd != "" {
		cmd.Dir = cwd
	}
	out, err := cmd.CombinedOutput()

	logOutput := string(out)
	if len(logOutput) > maxLogOutputLen {
		logOutput = logOutput[:maxLogOutputLen] + fmt.Sprintf("... (%d bytes total)", len(out))
	}

	d.logger.Debug(
		"exec",
		slog.String("command", strings.Join(cmd.Args, " ")),
		slog.String("cwd", cwd),
		slog.String("output", logOutput),
		slog.Any("error", err),
	)
	if err != nil {
		return string(out), err
	}

	return string(out), nil
}

// New factory to create a new Exec instance.
func New(
	logger *slog.Logger,
	sudo bool,
) *Exec {
	l := logger.With(slog.String("subsystem", "agent.exec"))

	return &Exec{
		logger:   l,
		sudo:     sudo,
		executor: &defaultExecutor{logger: l},
	}
}

// RunCmdImpl executes the provided command with the specified arguments and
// an optional working directory. It delegates to the CommandExecutor.
func (e *Exec) RunCmdImpl(
	name string,
	args []string,
	cwd string,
) (string, error) {
	return e.executor.Execute(name, args, cwd)
}

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

package exec_test

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/exec"
)

type RunCmdFullPublicTestSuite struct {
	suite.Suite

	logger *slog.Logger
}

func (suite *RunCmdFullPublicTestSuite) SetupTest() {
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (suite *RunCmdFullPublicTestSuite) TestRunCmdFull() {
	tests := []struct {
		name          string
		command       string
		args          []string
		cwd           string
		timeout       int
		ctxFunc       func() (context.Context, context.CancelFunc)
		expectError   bool
		errorContains string
		validateFunc  func(*exec.CmdResult)
	}{
		{
			name:    "successful command with stdout",
			command: "echo",
			args:    []string{"hello"},
			timeout: 5,
			validateFunc: func(r *exec.CmdResult) {
				suite.Equal("hello\n", r.Stdout)
				suite.Empty(r.Stderr)
				suite.Equal(0, r.ExitCode)
				suite.Greater(r.DurationMs, int64(-1))
			},
		},
		{
			name:    "command with stderr",
			command: "/bin/sh",
			args:    []string{"-c", "echo error >&2"},
			timeout: 5,
			validateFunc: func(r *exec.CmdResult) {
				suite.Equal("error\n", r.Stderr)
				suite.Equal(0, r.ExitCode)
			},
		},
		{
			name:    "command with non-zero exit code",
			command: "/bin/sh",
			args:    []string{"-c", "exit 42"},
			timeout: 5,
			validateFunc: func(r *exec.CmdResult) {
				suite.Equal(42, r.ExitCode)
			},
		},
		{
			name:    "command with working directory",
			command: "pwd",
			args:    []string{},
			cwd:     "/tmp",
			timeout: 5,
			validateFunc: func(r *exec.CmdResult) {
				suite.Contains(r.Stdout, "tmp")
				suite.Equal(0, r.ExitCode)
			},
		},
		{
			name:    "a zero timeout runs the command",
			command: "echo",
			args:    []string{"ok"},
			timeout: 0,
			validateFunc: func(r *exec.CmdResult) {
				suite.Equal("ok\n", r.Stdout)
				suite.Equal(0, r.ExitCode)
			},
		},
		{
			name:          "command timeout",
			command:       "sleep",
			args:          []string{"10"},
			timeout:       1,
			expectError:   true,
			errorContains: "command timed out after 1s",
		},
		{
			// The background sleep holds the output open after its shell is
			// killed. The timeout kills the whole process group, so Run
			// returns at the timeout, not when the sleep ends.
			name:          "command timeout with a child that holds the output",
			command:       "/bin/sh",
			args:          []string{"-c", "sleep 30 & sleep 30"},
			timeout:       1,
			expectError:   true,
			errorContains: "command timed out after 1s",
			validateFunc: func(r *exec.CmdResult) {
				suite.Less(r.DurationMs, int64(5000))
				suite.Equal(-1, r.ExitCode)
			},
		},
		{
			name:    "output over the limit is kept to the limit, with the rest counted",
			command: "/bin/sh",
			args: []string{
				"-c",
				"head -c " + strconv.Itoa(exec.MaxCommandOutput+3) + " /dev/zero; echo err >&2",
			},
			timeout: 5,
			validateFunc: func(r *exec.CmdResult) {
				suite.Equal(
					strings.Repeat("\x00", exec.MaxCommandOutput)+
						"\n[3 more bytes of output not kept]\n",
					r.Stdout,
				)
				suite.Equal("err\n", r.Stderr)
			},
		},
		{
			// The caller stops while the command runs, so Run gives the exit
			// error of the killed process, not the error of the context.
			name:    "a command that its caller stops is an error",
			command: "sleep",
			args:    []string{"5"},
			timeout: 5,
			ctxFunc: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				time.AfterFunc(100*time.Millisecond, cancel)

				return ctx, cancel
			},
			expectError:   true,
			errorContains: "command stopped: context canceled",
			validateFunc: func(r *exec.CmdResult) {
				suite.Equal(-1, r.ExitCode)
			},
		},
		{
			name:          "command not found",
			command:       "nonexistent-command-xyz",
			args:          []string{},
			timeout:       5,
			expectError:   true,
			errorContains: "failed to execute command",
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			ctx, cancel := context.Background(), context.CancelFunc(func() {})
			if tc.ctxFunc != nil {
				ctx, cancel = tc.ctxFunc()
			}
			defer cancel()
			em := exec.New(suite.logger, false)

			result, err := em.RunCmdFull(
				ctx,
				tc.command,
				tc.args,
				tc.cwd,
				tc.timeout,
			)

			if tc.expectError {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errorContains)
			} else {
				suite.Require().NoError(err)
			}
			suite.Require().NotNil(result)
			if tc.validateFunc != nil {
				tc.validateFunc(result)
			}
		})
	}
}

func TestRunCmdFullPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(RunCmdFullPublicTestSuite))
}

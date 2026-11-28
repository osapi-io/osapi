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
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/exec"
)

type RunPrivilegedCmdInteractivePublicTestSuite struct {
	suite.Suite
}

func (s *RunPrivilegedCmdInteractivePublicTestSuite) TearDownSubTest() {
	exec.ResetSudoProgram()
}

// answer is an interact that writes line to the command's standard input.
func answer(
	line string,
) func(context.Context, io.Writer) error {
	return func(
		_ context.Context,
		stdin io.Writer,
	) error {
		_, err := io.WriteString(stdin, line)

		return err
	}
}

func (s *RunPrivilegedCmdInteractivePublicTestSuite) TestRunPrivilegedCmdInteractive() {
	boom := errors.New("boom")
	tests := []struct {
		name         string
		sudo         bool
		command      string
		args         []string
		timeout      time.Duration
		interact     func(context.Context, io.Writer) error
		validateFunc func(out string, err error, took time.Duration)
	}{
		{
			name:     "the command reads what interact writes, while it runs",
			command:  "/bin/sh",
			args:     []string{"-c", "echo ready; read l; echo got:$l; echo err >&2"},
			interact: answer("yes\n"),
			validateFunc: func(out string, err error, _ time.Duration) {
				s.Require().NoError(err)
				s.Equal("ready\ngot:yes\nerr\n", out)
			},
		},
		{
			name:     "with sudo, the command runs through sudo -n",
			sudo:     true,
			command:  "/bin/sh",
			args:     []string{"-c", "read l; echo $l"},
			interact: answer("hi\n"),
			validateFunc: func(out string, err error, _ time.Duration) {
				s.Require().NoError(err)
				s.Equal("hi\n", out)
			},
		},
		{
			name:    "interact's context ends when the command exits",
			command: "true",
			interact: func(
				ctx context.Context,
				_ io.Writer,
			) error {
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(10 * time.Second):
					return errors.New("interact waited for a command that had exited")
				}
			},
			validateFunc: func(_ string, err error, took time.Duration) {
				s.Require().NoError(err)
				s.Less(took, 5*time.Second)
			},
		},
		{
			name:     "interact's error and the command's are both the error",
			command:  "/bin/sh",
			args:     []string{"-c", "exit 3"},
			interact: func(context.Context, io.Writer) error { return boom },
			validateFunc: func(_ string, err error, _ time.Duration) {
				s.ErrorIs(err, boom)
				s.ErrorContains(err, "exit status 3")
			},
		},
		{
			name:     "a command that ctx ends is the context's error",
			command:  "sleep",
			args:     []string{"10"},
			timeout:  200 * time.Millisecond,
			interact: func(context.Context, io.Writer) error { return nil },
			validateFunc: func(_ string, err error, _ time.Duration) {
				s.ErrorIs(err, context.DeadlineExceeded)
			},
		},
		{
			name:     "a program that does not start is the error",
			command:  "/nonexistent/osapi-probe",
			interact: func(context.Context, io.Writer) error { return nil },
			validateFunc: func(_ string, err error, _ time.Duration) {
				s.ErrorContains(err, "command /nonexistent/osapi-probe")
			},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.sudo {
				// The stand-in for sudo runs its command only when its first
				// argument is -n, so a sudo that could ask for a password fails.
				sudo := filepath.Join(s.T().TempDir(), "sudo")
				s.Require().NoError(os.WriteFile(sudo,
					[]byte("#!/bin/sh\n[ \"$1\" = -n ] || exit 9\nshift\nexec \"$@\"\n"), 0o700))
				exec.SetSudoProgram(sudo)
			}
			ctx := context.Background()
			if tc.timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, tc.timeout)
				defer cancel()
			}
			e := exec.New(slog.Default(), tc.sudo)
			start := time.Now()

			out, err := e.RunPrivilegedCmdInteractive(ctx, tc.command, tc.args, tc.interact)

			tc.validateFunc(out, err, time.Since(start))
		})
	}
}

func TestRunPrivilegedCmdInteractivePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(RunPrivilegedCmdInteractivePublicTestSuite))
}

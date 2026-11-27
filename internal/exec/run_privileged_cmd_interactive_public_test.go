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

func (s *RunPrivilegedCmdInteractivePublicTestSuite) TestRunPrivilegedCmdInteractive() {
	ctx := context.Background()
	answer := func(line string) func(context.Context, io.Writer) error {
		return func(_ context.Context, stdin io.Writer) error {
			_, err := io.WriteString(stdin, line)

			return err
		}
	}

	s.Run("the command reads what interact writes, while it runs", func() {
		e := exec.New(slog.Default(), false)
		out, err := e.RunPrivilegedCmdInteractive(ctx, "/bin/sh",
			[]string{"-c", "echo ready; read l; echo got:$l; echo err >&2"}, answer("yes\n"))
		s.Require().NoError(err)
		s.Equal("ready\ngot:yes\nerr\n", out)
	})
	s.Run("with sudo, the command runs through sudo -n", func() {
		sudo := filepath.Join(s.T().TempDir(), "sudo")
		s.Require().NoError(os.WriteFile(sudo,
			[]byte("#!/bin/sh\n[ \"$1\" = -n ] || exit 9\nshift\nexec \"$@\"\n"), 0o700))
		defer exec.SetSudoProgram(sudo)()
		e := exec.New(slog.Default(), true)
		out, err := e.RunPrivilegedCmdInteractive(ctx, "/bin/sh",
			[]string{"-c", "read l; echo $l"}, answer("hi\n"))
		s.Require().NoError(err)
		s.Equal("hi\n", out)
	})
	s.Run("interact's context ends when the command exits", func() {
		e := exec.New(slog.Default(), false)
		start := time.Now()
		_, err := e.RunPrivilegedCmdInteractive(ctx, "true", nil,
			func(ctx context.Context, _ io.Writer) error {
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(10 * time.Second):
					return errors.New("interact waited for a command that had exited")
				}
			})
		s.Require().NoError(err)
		s.Less(time.Since(start), 5*time.Second)
	})
	s.Run("interact's error and the command's are both the error", func() {
		e := exec.New(slog.Default(), false)
		boom := errors.New("boom")
		_, err := e.RunPrivilegedCmdInteractive(ctx, "/bin/sh", []string{"-c", "exit 3"},
			func(context.Context, io.Writer) error { return boom })
		s.ErrorIs(err, boom)
		s.ErrorContains(err, "exit status 3")
	})
	s.Run("a command that ctx ends is the context's error", func() {
		e := exec.New(slog.Default(), false)
		cctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
		defer cancel()
		_, err := e.RunPrivilegedCmdInteractive(cctx, "sleep", []string{"10"},
			func(context.Context, io.Writer) error { return nil })
		s.ErrorIs(err, context.DeadlineExceeded)
	})
	s.Run("a program that does not start is the error", func() {
		e := exec.New(slog.Default(), false)
		_, err := e.RunPrivilegedCmdInteractive(ctx, "/nonexistent/osapi-probe", nil,
			func(context.Context, io.Writer) error { return nil })
		s.ErrorContains(err, "command /nonexistent/osapi-probe")
	})
}

func TestRunPrivilegedCmdInteractivePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(RunPrivilegedCmdInteractivePublicTestSuite))
}

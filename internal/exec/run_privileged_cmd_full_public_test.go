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
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/exec"
)

type RunPrivilegedCmdFullPublicTestSuite struct {
	suite.Suite
}

func (s *RunPrivilegedCmdFullPublicTestSuite) TestRunPrivilegedCmdFull() {
	ctx := context.Background()

	s.Run("without sudo, the output, the errors and a code other than 0 are apart", func() {
		e := exec.New(slog.Default(), false)
		out, errOut, code, err := e.RunPrivilegedCmdFull(ctx, "/bin/sh",
			[]string{"-c", "echo out; echo err >&2; exit 3"})
		s.Require().NoError(err)
		s.Equal("out\n", out)
		s.Equal("err\n", errOut)
		s.Equal(3, code)
	})
	s.Run("with sudo, the command runs through it", func() {
		defer exec.SetSudoProgram("env")()
		e := exec.New(slog.Default(), true)
		out, _, code, err := e.RunPrivilegedCmdFull(ctx, "echo", []string{"hi"})
		s.Require().NoError(err)
		s.Equal("hi\n", out)
		s.Equal(0, code)
	})
	s.Run("a program that does not start is the error", func() {
		e := exec.New(slog.Default(), false)
		_, _, _, err := e.RunPrivilegedCmdFull(ctx, "/nonexistent/osapi-probe", nil)
		s.ErrorContains(err, "failed to execute command")
	})
}

func TestRunPrivilegedCmdFullPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(RunPrivilegedCmdFullPublicTestSuite))
}

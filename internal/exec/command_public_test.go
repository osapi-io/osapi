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

package exec_test

import (
	"context"
	"log/slog"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/exec"
)

type CommandPublicTestSuite struct {
	suite.Suite
}

func (s *CommandPublicTestSuite) TestTimeoutKill() {
	tests := []struct {
		name string
		kill func(int, syscall.Signal) error
	}{
		{
			name: "when the group is already gone the wait still ends",
			kill: func(int, syscall.Signal) error { return syscall.ESRCH },
		},
		{
			name: "when the agent may not signal the group it kills the process",
			kill: func(int, syscall.Signal) error { return syscall.EPERM },
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			defer exec.SetKillGroup(tc.kill, 100*time.Millisecond)()
			e := exec.New(slog.Default(), false)

			start := time.Now()
			result, err := e.RunCmdFull(context.Background(), "sleep", []string{"30"}, "", 1)

			s.Require().ErrorContains(err, "command timed out after 1s")
			s.Equal(-1, result.ExitCode)
			s.Less(time.Since(start), 5*time.Second)
		})
	}
}

func TestCommandPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(CommandPublicTestSuite))
}

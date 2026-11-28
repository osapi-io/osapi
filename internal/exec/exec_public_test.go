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
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/exec"
)

type ExecPublicTestSuite struct {
	suite.Suite

	logger *slog.Logger
}

func (s *ExecPublicTestSuite) SetupTest() {
	s.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (s *ExecPublicTestSuite) TestRunCmdImpl() {
	dir := s.T().TempDir()
	tests := []struct {
		name         string
		command      string
		args         []string
		cwd          string
		validateFunc func(string, error)
	}{
		{
			name:    "a command runs in the working directory that it is given",
			command: "pwd",
			cwd:     dir,
			validateFunc: func(output string, err error) {
				s.Require().NoError(err)
				resolved, _ := os.Readlink(dir)
				s.Contains([]string{dir + "\n", resolved + "\n", "/private" + dir + "\n"}, output)
			},
		},
		{
			name:    "a command with no working directory runs in the agent's",
			command: "echo",
			args:    []string{"-n", "ok"},
			validateFunc: func(output string, err error) {
				s.Require().NoError(err)
				s.Equal("ok", output)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			em := exec.New(s.logger, false)

			tt.validateFunc(em.RunCmdImpl(context.Background(), tt.command, tt.args, tt.cwd))
		})
	}
}

func TestExecPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ExecPublicTestSuite))
}

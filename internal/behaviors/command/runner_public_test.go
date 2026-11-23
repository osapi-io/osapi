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

package command_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/behavior/host"
	hostmocks "github.com/osapi-io/osapi/internal/behavior/host/mocks"
	"github.com/osapi-io/osapi/internal/behaviors/command"
	"github.com/osapi-io/osapi/internal/exec"
)

type RunnerPublicTestSuite struct {
	suite.Suite

	ctrl *gomock.Controller
	exec *hostmocks.MockExec
	p    command.Provider
}

func (s *RunnerPublicTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.exec = hostmocks.NewMockExec(s.ctrl)
	p, err := command.Behavior.Provider("debian", host.Host{Exec: s.exec, CommandEnabled: true})
	s.Require().NoError(err)
	s.p = p.(command.Provider)
}

func (s *RunnerPublicTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *RunnerPublicTestSuite) TestExec() {
	tests := []struct {
		name    string
		in      command.ExecInput
		answer  *exec.CmdResult
		err     error
		want    command.Result
		wantErr string
	}{
		{
			name: "when the program exits zero",
			in: command.ExecInput{
				Command: "echo",
				Args:    []string{"hi"},
				Cwd:     "/tmp",
				Timeout: 5,
			},
			answer: &exec.CmdResult{Stdout: "hi\n", DurationMs: 3},
			want:   command.Result{Stdout: "hi\n", DurationMs: 3, Changed: true},
		},
		{
			name:   "when the program exits non-zero it is a result",
			in:     command.ExecInput{Command: "false"},
			answer: &exec.CmdResult{Stderr: "no\n", ExitCode: 1},
			want:   command.Result{Stderr: "no\n", ExitCode: 1, Changed: true},
		},
		{
			name:    "when the program cannot start",
			in:      command.ExecInput{Command: "nope"},
			err:     errors.New("not found"),
			wantErr: "run nope: not found",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.exec.EXPECT().
				RunCmdFull(gomock.Any(), tc.in.Command, tc.in.Args, tc.in.Cwd, tc.in.Timeout).
				Return(tc.answer, tc.err)

			got, err := s.p.Exec(context.Background(), tc.in)
			if tc.wantErr != "" {
				s.EqualError(err, tc.wantErr)

				return
			}
			s.NoError(err)
			s.Equal(tc.want, got)
		})
	}
}

func (s *RunnerPublicTestSuite) TestShell() {
	s.exec.EXPECT().
		RunCmdFull(gomock.Any(), "/bin/sh", []string{"-c", "echo a | wc -l"}, "/", 9).
		Return(&exec.CmdResult{Stdout: "1\n"}, nil)

	got, err := s.p.Shell(context.Background(),
		command.ShellInput{Command: "echo a | wc -l", Cwd: "/", Timeout: 9})
	s.NoError(err)
	s.Equal(command.Result{Stdout: "1\n", Changed: true}, got)
}

// TestDisabled: an agent with agent.command.enabled false runs no command.
func (s *RunnerPublicTestSuite) TestDisabled() {
	p, err := command.Behavior.Provider("debian", host.Host{Exec: s.exec})
	s.Require().NoError(err)

	_, err = p.(command.Provider).Exec(context.Background(), command.ExecInput{Command: "true"})
	s.ErrorIs(err, host.ErrCommandDisabled)
	_, err = p.(command.Provider).Shell(context.Background(), command.ShellInput{Command: "true"})
	s.ErrorContains(err, "agent.command.enabled")
}

func TestRunnerPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(RunnerPublicTestSuite))
}

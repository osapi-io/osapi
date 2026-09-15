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
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/exec"
	"github.com/osapi-io/osapi/internal/exec/mocks"
)

type RunPrivilegedCmdWithStdinPublicTestSuite struct {
	suite.Suite

	ctrl         *gomock.Controller
	mockExecutor *mocks.MockCommandExecutor
	logger       *slog.Logger
}

func (s *RunPrivilegedCmdWithStdinPublicTestSuite) SetupTest() {
	s.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (s *RunPrivilegedCmdWithStdinPublicTestSuite) SetupSubTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockExecutor = mocks.NewMockCommandExecutor(s.ctrl)
}

func (s *RunPrivilegedCmdWithStdinPublicTestSuite) TearDownSubTest() {
	s.ctrl.Finish()
}

func (s *RunPrivilegedCmdWithStdinPublicTestSuite) TestRunPrivilegedCmdWithStdin() {
	tests := []struct {
		name    string
		sudo    bool
		command string
		args    []string
		stdin   string
		// setupMock is nil for rows that run a real command through the
		// default executor.
		setupMock    func()
		validateFunc func(string, error)
	}{
		{
			name:    "without sudo passes stdin to the command",
			sudo:    false,
			command: "chpasswd",
			args:    nil,
			stdin:   "john:secret\n",
			setupMock: func() {
				s.mockExecutor.EXPECT().
					ExecuteWithStdin("chpasswd", []string(nil), "", "john:secret\n").
					Return("", nil)
			},
			validateFunc: func(output string, err error) {
				s.NoError(err)
				s.Equal("", output)
			},
		},
		{
			name:    "with sudo prepends sudo and keeps stdin out of args",
			sudo:    true,
			command: "chpasswd",
			args:    nil,
			stdin:   "john:secret\n",
			setupMock: func() {
				s.mockExecutor.EXPECT().
					ExecuteWithStdin("sudo", []string{"chpasswd"}, "", "john:secret\n").
					Return("", nil)
			},
			validateFunc: func(output string, err error) {
				s.NoError(err)
				s.Equal("", output)
			},
		},
		{
			name:    "executor error propagates",
			sudo:    false,
			command: "chpasswd",
			args:    nil,
			stdin:   "john:secret\n",
			setupMock: func() {
				s.mockExecutor.EXPECT().
					ExecuteWithStdin("chpasswd", []string(nil), "", "john:secret\n").
					Return("chpasswd: line 1: user 'john' does not exist", fmt.Errorf("exit status 1"))
			},
			validateFunc: func(output string, err error) {
				s.Error(err)
				s.Contains(output, "does not exist")
			},
		},
		{
			name:    "default executor writes stdin to the real command",
			sudo:    false,
			command: "cat",
			args:    nil,
			stdin:   "it's; $(not) `a` shell\n",
			validateFunc: func(output string, err error) {
				s.NoError(err)
				s.Equal("it's; $(not) `a` shell\n", output)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			em := exec.New(s.logger, tc.sudo)
			if tc.setupMock != nil {
				tc.setupMock()
				exec.SetExecutor(em, s.mockExecutor)
			}

			output, err := em.RunPrivilegedCmdWithStdin(tc.command, tc.args, tc.stdin)

			tc.validateFunc(output, err)
		})
	}
}

// In order for `go test` to run this suite, we need to create
// a normal test function and pass our suite to suite.Run.
func TestRunPrivilegedCmdWithStdinPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(RunPrivilegedCmdWithStdinPublicTestSuite))
}

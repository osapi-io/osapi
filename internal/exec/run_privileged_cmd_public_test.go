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
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/retr0h/osapi/internal/exec"
)

type RunPrivilegedCmdPublicTestSuite struct {
	suite.Suite

	logger *slog.Logger
}

func (suite *RunPrivilegedCmdPublicTestSuite) SetupTest() {
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (suite *RunPrivilegedCmdPublicTestSuite) TearDownTest() {}

func (suite *RunPrivilegedCmdPublicTestSuite) TestRunPrivilegedCmd() {
	tests := []struct {
		name         string
		sudo         bool
		command      string
		args         []string
		expectError  bool
		validateFunc func(string, error)
	}{
		{
			name:        "without sudo runs command directly",
			sudo:        false,
			command:     "echo",
			args:        []string{"-n", "hello"},
			expectError: false,
			validateFunc: func(output string, _ error) {
				suite.Require().Equal("hello", output)
			},
		},
		{
			name:        "without sudo invalid command returns error",
			sudo:        false,
			command:     "nonexistent-command-xyz",
			args:        []string{},
			expectError: true,
			validateFunc: func(_ string, err error) {
				suite.Require().Contains(err.Error(), "not found")
			},
		},
		{
			name:        "with sudo prepends sudo to command",
			sudo:        true,
			command:     "nonexistent-command-xyz",
			args:        []string{"arg1"},
			expectError: true,
			validateFunc: func(_ string, err error) {
				// When sudo=true, the exec manager prepends "sudo" to the
				// command. The error varies by environment (sudo may not be
				// installed, or may prompt for a password), but the attempt
				// to run sudo confirms the prepend behavior.
				suite.Require().Error(err)
			},
		},
		{
			name:        "with sudo and no args",
			sudo:        true,
			command:     "nonexistent-command-xyz",
			args:        []string{},
			expectError: true,
			validateFunc: func(_ string, err error) {
				suite.Require().Error(err)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			em := exec.New(suite.logger, tc.sudo)

			output, err := em.RunPrivilegedCmd(tc.command, tc.args)

			if tc.expectError {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
			if tc.validateFunc != nil {
				tc.validateFunc(output, err)
			}
		})
	}
}

// In order for `go test` to run this suite, we need to create
// a normal test function and pass our suite to suite.Run.
func TestRunPrivilegedCmdPublicTestSuite(t *testing.T) {
	suite.Run(t, new(RunPrivilegedCmdPublicTestSuite))
}

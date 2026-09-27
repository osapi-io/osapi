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
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/exec"
)

type RunCmdPublicTestSuite struct {
	suite.Suite

	logger *slog.Logger
}

func (suite *RunCmdPublicTestSuite) SetupTest() {
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (suite *RunCmdPublicTestSuite) TearDownTest() {}

func (suite *RunCmdPublicTestSuite) TestRunCmd() {
	tests := []struct {
		name         string
		command      string
		args         []string
		validateFunc func(string, error)
	}{
		{
			name:    "Valid command with no arguments",
			command: "ls",
			args:    []string{},
			validateFunc: func(output string, err error) {
				suite.Require().NoError(err)
				suite.Require().NotEmpty(output)
			},
		},
		{
			name:    "Valid command with no arguments and working dir",
			command: "ls",
			args:    []string{},
			validateFunc: func(output string, err error) {
				suite.Require().NoError(err)
				suite.Require().NotEmpty(output)
			},
		},
		{
			name:    "Valid command with output",
			command: "echo",
			args:    []string{"-n", "foo"},
			validateFunc: func(output string, err error) {
				suite.Require().NoError(err)
				suite.Require().NotEmpty(output)
			},
		},
		{
			name:    "Invalid command",
			command: "invalid",
			args:    []string{"foo"},
			validateFunc: func(_ string, err error) {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), "not found")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			em := exec.New(suite.logger, false)

			tc.validateFunc(em.RunCmd(context.Background(), tc.command, tc.args))
		})
	}
}

// TestRunCmdCancellation covers what the context is for: a command that would
// otherwise run to completion is killed when the caller's context ends, and the
// reason it stopped comes from the context rather than from the command's own
// output, which a killed command does not produce.
func (suite *RunCmdPublicTestSuite) TestRunCmdCancellation() {
	tests := []struct {
		name         string
		ctxFunc      func() (context.Context, context.CancelFunc)
		validateFunc func(string, error)
	}{
		{
			name: "a context cancelled before the command starts",
			ctxFunc: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				return ctx, func() {}
			},
			validateFunc: func(_ string, err error) {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, context.Canceled)
				suite.Require().Contains(err.Error(), "command sleep")
			},
		},
		{
			name: "a deadline that passes while the command runs",
			ctxFunc: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 10*time.Millisecond)
			},
			validateFunc: func(_ string, err error) {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, context.DeadlineExceeded)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			ctx, cancel := tc.ctxFunc()
			defer cancel()

			em := exec.New(suite.logger, false)

			tc.validateFunc(em.RunCmd(ctx, "sleep", []string{"5"}))
		})
	}
}

// In order for `go test` to run this suite, we need to create
// a normal test function and pass our suite to suite.Run.
func TestRunCmdPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(RunCmdPublicTestSuite))
}

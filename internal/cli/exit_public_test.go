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

package cli_test

import (
	"bytes"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli"
)

type ExitPublicTestSuite struct {
	suite.Suite
}

func (suite *ExitPublicTestSuite) TearDownTest() {
	cli.ResetOsExit()
	cli.ResetErrOut()
	cli.RunExitHooks()
}

// A command that fails through HandleError or LogFatal exits there, and
// cobra's PersistentPostRun never runs. The hooks are how its span ends and
// its tracer flushes anyway.
func (suite *ExitPublicTestSuite) TestExitRunsHooks() {
	tests := []struct {
		name         string
		exit         func()
		validateFunc func(order []string)
	}{
		{
			name: "when HandleError exits the hooks run first, last added first",
			exit: func() { cli.HandleError(errors.New("boom"), slog.Default()) },
			validateFunc: func(order []string) {
				suite.Equal([]string{"second", "first", "exit"}, order)
			},
		},
		{
			name: "when LogFatal exits the hooks run first",
			exit: func() { cli.LogFatal(slog.Default(), "failed", errors.New("boom")) },
			validateFunc: func(order []string) {
				suite.Equal([]string{"second", "first", "exit"}, order)
			},
		},
		{
			name: "when RunExitHooks already ran the hooks do not run again",
			exit: func() {
				cli.RunExitHooks()
				cli.HandleError(errors.New("boom"), slog.Default())
			},
			validateFunc: func(order []string) {
				suite.Equal([]string{"second", "first", "exit"}, order)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var order []string
			cli.SetErrOut(&bytes.Buffer{})
			cli.SetOsExit(func(int) { order = append(order, "exit") })
			cli.AtExit(func() { order = append(order, "first") })
			cli.AtExit(func() { order = append(order, "second") })

			tc.exit()

			tc.validateFunc(order)
		})
	}
}

func TestExitPublicTestSuite(t *testing.T) {
	suite.Run(t, new(ExitPublicTestSuite))
}

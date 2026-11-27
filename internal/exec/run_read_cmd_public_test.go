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

type RunReadCmdPublicTestSuite struct {
	suite.Suite
}

// TestRunReadCmd is a read as gohai runs it: stdout is the data, stderr is
// in the error and never the data, and the locale is C.
func (s *RunReadCmdPublicTestSuite) TestRunReadCmd() {
	tests := []struct {
		name     string
		cmd      string
		args     []string
		wantOut  string
		errHas   string
		wantFail bool
	}{
		{
			name:    "when the command succeeds, its stdout is the data",
			cmd:     "sh",
			args:    []string{"-c", "echo data; echo warning >&2"},
			wantOut: "data\n",
		},
		{
			name:    "when it runs, LC_ALL is C",
			cmd:     "sh",
			args:    []string{"-c", "echo $LC_ALL"},
			wantOut: "C\n",
		},
		{
			name:     "when it fails, its stderr is in the error, and its stdout is returned",
			cmd:      "sh",
			args:     []string{"-c", "echo partial; echo broken >&2; exit 3"},
			wantOut:  "partial\n",
			errHas:   "broken",
			wantFail: true,
		},
		{
			name:     "when the program does not exist",
			cmd:      "osapi-no-such-program",
			errHas:   "osapi-no-such-program",
			wantFail: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			e := exec.New(slog.Default(), false)

			out, err := e.RunReadCmd(context.Background(), tc.cmd, tc.args)

			s.Equal(tc.wantOut, string(out))
			if tc.wantFail {
				s.Require().Error(err)
				s.Contains(err.Error(), tc.errHas)

				return
			}
			s.NoError(err)
		})
	}
}

func TestRunReadCmdPublicTestSuite(t *testing.T) {
	suite.Run(t, new(RunReadCmdPublicTestSuite))
}

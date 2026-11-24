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

package types_test

import (
	"testing"

	"github.com/osapi-io/osapi/pkg/sdk/types"
	"github.com/stretchr/testify/suite"
)

type JobStepsPublicTestSuite struct {
	suite.Suite
}

func (s *JobStepsPublicTestSuite) TestFormatSteps() {
	changed, same := true, false
	tests := []struct {
		name  string
		steps []types.Step
		want  string
	}{
		{
			name: "when one step calls another, they are a line",
			steps: []types.Step{
				{
					Behavior: "netplan",
					Method:   "update",
					Changed:  &changed,
					Steps: []types.Step{
						{Behavior: "managedfile", Method: "update", Changed: &changed},
					},
				},
			},
			want: "netplan update (changed) → managedfile update (changed)",
		},
		{
			name: "when a method calls more than one, they are in parentheses",
			steps: []types.Step{
				{Behavior: "hostname", Method: "get"},
				{Behavior: "uptime", Method: "get", Changed: &same},
			},
			want: "(hostname get, uptime get)",
		},
		{
			name:  "when a step failed, it is marked failed",
			steps: []types.Step{{Behavior: "managedfile", Method: "update", Error: "disk full"}},
			want:  "managedfile update (failed)",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, types.FormatSteps(tc.steps))
		})
	}
}

func TestJobStepsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(JobStepsPublicTestSuite))
}

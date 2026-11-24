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

type StatusPublicTestSuite struct {
	suite.Suite
}

func (s *StatusPublicTestSuite) TestValues() {
	tests := []struct {
		name    string
		values  []string
		has     []string
		hasNot  []string
		isFirst string
	}{
		{
			name:   "when a job has a status, it is a status of the job and not of a host",
			values: types.JobStatus("").Values(),
			has: []string{
				string(types.JobStatusPartialFailure),
				string(types.JobStatusSkipped),
			},
			hasNot: []string{
				string(types.AgentJobStatusTimeout),
				string(types.AgentJobStatusStarted),
			},
			isFirst: string(types.JobStatusSubmitted),
		},
		{
			name:   "when a host has a status, it is a status of the host and not of the job",
			values: types.AgentJobStatus("").Values(),
			has: []string{
				string(types.AgentJobStatusRetried),
				string(types.AgentJobStatusTimeout),
			},
			hasNot: []string{
				string(types.JobStatusPartialFailure),
				string(types.JobStatusProcessing),
			},
			isFirst: string(types.AgentJobStatusSubmitted),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.isFirst, tc.values[0])
			s.Subset(tc.values, tc.has)
			for _, v := range tc.hasNot {
				s.NotContains(tc.values, v)
			}
			seen := map[string]bool{}
			for _, v := range tc.values {
				s.False(seen[v], "%s is a value twice", v)
				seen[v] = true
			}
		})
	}
}

func TestStatusPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(StatusPublicTestSuite))
}

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

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/pkg/sdk/types"
)

type StatusPublicTestSuite struct {
	suite.Suite
}

func (s *StatusPublicTestSuite) TestJobStatus() {
	tests := []struct {
		name   string
		status types.JobStatus
		want   bool
	}{
		{
			name:   "when a job completed on some agents and failed on others, it is a status",
			status: types.JobStatusPartialFailure,
			want:   true,
		},
		{
			name:   "when each agent skipped a job, it is a status",
			status: types.JobStatusSkipped,
			want:   true,
		},
		{
			name:   "when a word is the status of a host, it is not a status of a job",
			status: types.JobStatus(types.AgentJobStatusTimeout),
		},
		{
			name:   "when a word is pending, it is not a status",
			status: "pending",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, tc.status.Valid())
		})
	}
	s.Equal(types.JobStatusSubmitted, types.JobStatus("").Values()[0])
}

func (s *StatusPublicTestSuite) TestAgentJobStatus() {
	tests := []struct {
		name   string
		status types.AgentJobStatus
		want   bool
	}{
		{
			name:   "when the controller retried a job, it is a status",
			status: types.AgentJobStatusRetried,
			want:   true,
		},
		{
			name:   "when an agent did not answer, it is a status",
			status: types.AgentJobStatusTimeout,
			want:   true,
		},
		{
			name:   "when a word is the status of a job, it is not a status of a host",
			status: types.AgentJobStatus(types.JobStatusPartialFailure),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, tc.status.Valid())
		})
	}
	s.Equal(types.AgentJobStatusSubmitted, types.AgentJobStatus("").Values()[0])
}

func TestStatusPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(StatusPublicTestSuite))
}

// TestValid: a value of a closed set is valid when its type lists it.
func (s *StatusPublicTestSuite) TestValid() {
	tests := []struct {
		name  string
		valid func() bool
		want  bool
	}{
		{name: "when a job status is listed", valid: types.JobStatusSkipped.Valid, want: true},
		{name: "when a job status is not listed", valid: types.JobStatus("done").Valid},
		{name: "when an agent state is listed", valid: types.AgentDraining.Valid, want: true},
		{name: "when an agent state is not listed", valid: types.AgentState("Busy").Valid},
		{name: "when a condition is listed", valid: types.ConditionDrainStuck.Valid, want: true},
		{name: "when a condition is not listed", valid: types.ConditionType("Hot").Valid},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, tc.valid())
		})
	}
}

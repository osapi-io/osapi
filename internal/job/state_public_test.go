// Copyright (c) 2025 John Dewey

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

package job_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/job"
)

type StatePublicTestSuite struct {
	suite.Suite
}

func TestStatePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(StatePublicTestSuite))
}

// TestEffectiveState is the state the controller reports for each state an
// agent can report, with and without its drain flag.
func (s *StatePublicTestSuite) TestEffectiveState() {
	tests := []struct {
		name     string
		reported string
		flag     bool
		want     string
	}{
		{
			name:     "when the flag is set and the agent has not applied it is Draining",
			reported: job.AgentStateReady,
			flag:     true,
			want:     job.AgentStateDraining,
		},
		{
			name:     "when the flag is set and the agent reported no state is Draining",
			reported: "",
			flag:     true,
			want:     job.AgentStateDraining,
		},
		{
			name:     "when the agent is draining it is Draining",
			reported: job.AgentStateDraining,
			flag:     true,
			want:     job.AgentStateDraining,
		},
		{
			name:     "when the agent has cordoned it is Cordoned",
			reported: job.AgentStateCordoned,
			flag:     true,
			want:     job.AgentStateCordoned,
		},
		{
			name:     "when the flag is removed and the agent has not undrained it stays Cordoned",
			reported: job.AgentStateCordoned,
			flag:     false,
			want:     job.AgentStateCordoned,
		},
		{
			name:     "when there is no flag and no state it is Ready",
			reported: "",
			flag:     false,
			want:     job.AgentStateReady,
		},
		{
			name:     "when a pending agent has a flag it stays Pending",
			reported: job.AgentStatePending,
			flag:     true,
			want:     job.AgentStatePending,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, job.EffectiveState(tc.reported, tc.flag))
		})
	}
}

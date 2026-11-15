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

package metrics_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/telemetry/metrics"
)

type SubsystemsPublicTestSuite struct {
	suite.Suite
}

func TestSubsystemsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(SubsystemsPublicTestSuite))
}

func (s *SubsystemsPublicTestSuite) TestSubsystems() {
	tests := []struct {
		name string
		scs  map[string]job.SubComponentInfo
		want map[string]string
	}{
		{
			name: "when there are sub-components reports the status of each",
			scs: map[string]job.SubComponentInfo{
				"agent.heartbeat": {Status: "ok"},
				"agent.metrics":   {Status: "disabled"},
			},
			want: map[string]string{
				"agent.heartbeat": "ok",
				"agent.metrics":   "disabled",
			},
		},
		{
			name: "when there are none returns none",
			want: map[string]string{},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			statuses := metrics.Subsystems(tc.scs)
			got := map[string]string{}
			for _, st := range statuses {
				got[st.Name] = st.StatusFn()
			}
			s.Len(statuses, len(tc.want))
			s.Equal(tc.want, got)
		})
	}
}

package orchestrator

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type RunnerBroadcastTestSuite struct {
	suite.Suite
}

func TestRunnerBroadcastTestSuite(t *testing.T) {
	suite.Run(t, new(RunnerBroadcastTestSuite))
}

func (s *RunnerBroadcastTestSuite) TestIsBroadcastTarget() {
	tests := []struct {
		name   string
		target string
		want   bool
	}{
		{
			name:   "all agents is broadcast",
			target: "_all",
			want:   true,
		},
		{
			name:   "label selector is broadcast",
			target: "role:web",
			want:   true,
		},
		{
			name:   "single agent is not broadcast",
			target: "agent-001",
			want:   false,
		},
		{
			name:   "empty string is not broadcast",
			target: "",
			want:   false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := IsBroadcastTarget(tt.target)
			s.Equal(tt.want, got)
		})
	}
}

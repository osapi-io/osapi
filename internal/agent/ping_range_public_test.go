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

package agent_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/osapi-io/gohai/pkg/gohai/source"
	sourcemocks "github.com/osapi-io/gohai/pkg/gohai/source/mocks"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"k8s.io/utils/ptr"

	"github.com/osapi-io/osapi/internal/agent"
)

type PingRangePublicTestSuite struct {
	suite.Suite
}

func (s *PingRangePublicTestSuite) TearDownSubTest() {
	agent.ResetProcessGroups()
}

func (s *PingRangePublicTestSuite) TestPingProblem() {
	tests := []struct {
		name      string
		setting   *string
		groups    []int
		groupsErr error
		want      string
	}{
		{name: "when the host has no setting, such as macOS, there is no problem", want: ""},
		{
			name:    "when sysctl cannot run there is no problem to report",
			setting: ptr.To(""),
			want:    "",
		},
		{
			name:    "when a group of the agent is in the range a ping works",
			setting: ptr.To("0\t2147483647\n"),
			groups:  []int{990},
			want:    "",
		},
		{
			name:    "when a supplementary group is in the range a ping works",
			setting: ptr.To("1000 1000"),
			groups:  []int{990, 1000},
			want:    "",
		},
		{
			name:    "when the kernel default permits no group each ping fails",
			setting: ptr.To("1\t0\n"),
			groups:  []int{0},
			want:    "no group of the agent [0] is in net.ipv4.ping_group_range (1 0); each ping fails",
		},
		{
			name:    "when the setting does not have two values",
			setting: ptr.To("1"),
			want:    `cannot read net.ipv4.ping_group_range "1"`,
		},
		{
			name:    "when a value is not a number",
			setting: ptr.To("a 0"),
			want:    `cannot read net.ipv4.ping_group_range "a 0"`,
		},
		{
			name:      "when the groups cannot be read",
			setting:   ptr.To("0 2147483647"),
			groupsErr: errors.New("boom"),
			want:      "cannot read the groups of the agent: boom",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// gohai's sysctl collector runs sysctl -a: the setting is one
			// line of its output, as Linux prints it.
			runner := sourcemocks.NewMockRunner(gomock.NewController(s.T()))
			out, err := []byte("kernel.ostype = Linux\n"), error(nil)
			switch {
			case tc.setting == nil:
			case *tc.setting == "":
				out, err = nil, errors.New("no sysctl")
			default:
				out = append(out, []byte("net.ipv4.ping_group_range = "+*tc.setting+"\n")...)
			}
			runner.EXPECT().Execute(gomock.Any(), "sysctl", "-a").Return(out, err)
			agent.SetProcessGroups(func() ([]int, error) { return tc.groups, tc.groupsErr })

			s.Equal(tc.want, agent.PingProblem(
				context.Background(),
				source.Sources{Exec: runner, Family: "debian"},
			))
		})
	}
}

func (s *PingRangePublicTestSuite) TestDefaultProcessGroups() {
	s.Run("when read from the process it has the agent's group", func() {
		groups, err := agent.ExportDefaultProcessGroups()
		s.Require().NoError(err)
		s.Contains(groups, os.Getgid())
	})

	s.Run("when the groups cannot be read it returns the error", func() {
		agent.SetGetgroups(func() ([]int, error) { return nil, errors.New("boom") })
		defer agent.ResetGetgroups()

		_, err := agent.ExportDefaultProcessGroups()
		s.EqualError(err, "boom")
	})
}

func TestPingRangePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(PingRangePublicTestSuite))
}

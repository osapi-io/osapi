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

package commandtest_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior/command/commandtest"
	_ "github.com/osapi-io/osapi/internal/behaviors/network/dns/gen"
	_ "github.com/osapi-io/osapi/internal/behaviors/schedule/gen"
	_ "github.com/osapi-io/osapi/internal/behaviors/sysctl/gen"
)

type CommandtestPublicTestSuite struct {
	suite.Suite
}

func (s *CommandtestPublicTestSuite) TestSent() {
	tests := []struct {
		name string
		args []string
		want commandtest.Request
	}{
		{
			name: "when a command runs, the request it sends is returned",
			args: []string{
				"node",
				"sysctl",
				"get",
				"--target",
				"server1",
				"--key",
				"vm.swappiness",
			},
			want: commandtest.Request{
				Method: "GET",
				Path:   "/api/node/server1/sysctl/vm.swappiness",
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, commandtest.Sent(s.T(), tc.args))
		})
	}
}

// TestFlagsDoNotCarryOver runs two command lines on the same commands: the
// list, the bool and the string that the first sets are not in the second.
func (s *CommandtestPublicTestSuite) TestFlagsDoNotCarryOver() {
	first := commandtest.Sent(s.T(), []string{
		"node", "dns", "update", "--target", "server1", "--interface-name", "eth0",
		"--servers", "9.9.9.9", "--override-dhcp",
	})
	s.Equal(`{"servers":["9.9.9.9"],"override_dhcp":true}`, first.Body)

	second := commandtest.Sent(s.T(), []string{
		"node", "dns", "update", "--target", "server1", "--interface-name", "eth1",
		"--search-domains", "corp.example",
	})
	s.Equal("/api/node/server1/dns/eth1", second.Path)
	s.Equal(`{"search_domains":["corp.example"]}`, second.Body)
}

// TestMapFlagsDoNotCarryOver: a map flag adds each value to what it holds,
// so the vars of the first command line are not in the second.
func (s *CommandtestPublicTestSuite) TestMapFlagsDoNotCarryOver() {
	first := commandtest.Sent(s.T(), []string{
		"node", "schedule", "update", "--target", "server1", "--name", "backup",
		"--vars", "env=prod",
	})
	s.Equal(`{"vars":{"env":"prod"}}`, first.Body)

	second := commandtest.Sent(s.T(), []string{
		"node", "schedule", "update", "--target", "server1", "--name", "backup",
		"--schedule", "0 3 * * *",
	})
	s.Equal(`{"schedule":"0 3 * * *"}`, second.Body)
}

func TestCommandtestPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(CommandtestPublicTestSuite))
}

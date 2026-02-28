//go:build integration

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

package integration_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type AgentSmokeSuite struct {
	suite.Suite
}

func (s *AgentSmokeSuite) TestAgentList() {
	tests := []struct {
		name         string
		args         []string
		validateFunc func(stdout string, exitCode int)
	}{
		{
			name: "returns at least one agent",
			args: []string{"client", "agent", "list", "--json"},
			validateFunc: func(
				stdout string,
				exitCode int,
			) {
				s.Require().Equal(0, exitCode)

				var resp struct {
					Agents []struct {
						Hostname string `json:"hostname"`
					} `json:"agents"`
				}
				s.Require().NoError(parseJSON(stdout, &resp))
				s.GreaterOrEqual(len(resp.Agents), 1)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			stdout, _, exitCode := runCLI(tt.args...)
			tt.validateFunc(stdout, exitCode)
		})
	}
}

func (s *AgentSmokeSuite) TestAgentGet() {
	listOut, _, listCode := runCLI("client", "agent", "list", "--json")
	s.Require().Equal(0, listCode)

	var listResp struct {
		Agents []struct {
			Hostname string `json:"hostname"`
		} `json:"agents"`
	}
	s.Require().NoError(parseJSON(listOut, &listResp))
	s.Require().NotEmpty(listResp.Agents, "agent list must contain at least one entry")

	hostname := listResp.Agents[0].Hostname

	tests := []struct {
		name         string
		args         []string
		validateFunc func(stdout string)
	}{
		{
			name: "returns agent details for known hostname",
			args: []string{"client", "agent", "get", "--hostname", hostname, "--json"},
			validateFunc: func(stdout string) {
				var resp struct {
					Hostname string `json:"hostname"`
					Status   string `json:"status"`
				}
				s.Require().NoError(parseJSON(stdout, &resp))
				s.Equal(hostname, resp.Hostname)
				s.Equal("Ready", resp.Status)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			stdout, _, exitCode := runCLI(tt.args...)
			s.Require().Equal(0, exitCode)
			tt.validateFunc(stdout)
		})
	}
}

func TestAgentSmokeSuite(
	t *testing.T,
) {
	suite.Run(t, new(AgentSmokeSuite))
}

//go:build integration

// Copyright (c) 2024 John Dewey

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

type NetworkSmokeSuite struct {
	suite.Suite
}

func (s *NetworkSmokeSuite) TestNetworkDnsGet() {
	tests := []struct {
		name         string
		args         []string
		validateFunc func(stdout string, exitCode int)
	}{
		{
			name: "returns dns results",
			args: []string{
				"client",
				"node",
				"network",
				"dns",
				"get",
				"--interface-name",
				"eth0",
				"--json",
			},
			validateFunc: func(
				stdout string,
				exitCode int,
			) {
				s.Require().Equal(0, exitCode)

				var result map[string]any
				err := parseJSON(stdout, &result)
				s.Require().NoError(err)

				results, ok := result["results"].([]any)
				s.Require().True(ok)
				s.GreaterOrEqual(len(results), 1)
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

func (s *NetworkSmokeSuite) TestNetworkDnsUpdate() {
	skipWrite(s.T())

	tests := []struct {
		name         string
		args         []string
		validateFunc func(stdout string, exitCode int)
	}{
		{
			name: "returns dns update results",
			args: []string{
				"client",
				"node",
				"network",
				"dns",
				"update",
				"--interface-name",
				"eth0",
				"--servers",
				"1.1.1.1,8.8.8.8",
				"--search-domains",
				"example.com",
				"--json",
			},
			validateFunc: func(
				stdout string,
				exitCode int,
			) {
				s.Require().Equal(0, exitCode)

				var result map[string]any
				err := parseJSON(stdout, &result)
				s.Require().NoError(err)

				results, ok := result["results"].([]any)
				s.Require().True(ok)
				s.GreaterOrEqual(len(results), 1)
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

func (s *NetworkSmokeSuite) TestNetworkPingPost() {
	tests := []struct {
		name         string
		args         []string
		validateFunc func(stdout string, exitCode int)
	}{
		{
			name: "returns ping results",
			args: []string{"client", "node", "network", "ping", "--address", "127.0.0.1", "--json"},
			validateFunc: func(
				stdout string,
				exitCode int,
			) {
				s.Require().Equal(0, exitCode)

				var result map[string]any
				err := parseJSON(stdout, &result)
				s.Require().NoError(err)

				results, ok := result["results"].([]any)
				s.Require().True(ok)
				s.GreaterOrEqual(len(results), 1)
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

func TestNetworkSmokeSuite(
	t *testing.T,
) {
	suite.Run(t, new(NetworkSmokeSuite))
}

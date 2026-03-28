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

type NodeSmokeSuite struct {
	suite.Suite
}

func (s *NodeSmokeSuite) TestNodeHostnameGet() {
	tests := []struct {
		name         string
		args         []string
		validateFunc func(stdout string, exitCode int)
	}{
		{
			name: "returns hostname results",
			args: []string{"client", "node", "hostname", "get", "--json"},
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

				first, ok := results[0].(map[string]any)
				s.Require().True(ok)
				s.NotEmpty(first["hostname"])
				s.Contains(first, "changed")
				s.Equal(false, first["changed"])
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

func (s *NodeSmokeSuite) TestNodeStatusGet() {
	tests := []struct {
		name         string
		args         []string
		validateFunc func(stdout string, exitCode int)
	}{
		{
			name: "returns status results",
			args: []string{"client", "node", "status", "--json"},
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

				first, ok := results[0].(map[string]any)
				s.Require().True(ok)
				s.Contains(first, "changed")
				s.Equal(false, first["changed"])
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

func (s *NodeSmokeSuite) TestNodeHostnameUpdate() {
	skipWriteOp(s.T(), "HOSTNAME_UPDATE")

	// Use the current hostname to ensure idempotency (Changed: false).
	stdout, _, exitCode := runCLI("client", "node", "hostname", "get", "--json")
	s.Require().Equal(0, exitCode)

	var getResult map[string]any
	err := parseJSON(stdout, &getResult)
	s.Require().NoError(err)

	results, ok := getResult["results"].([]any)
	s.Require().True(ok)
	s.Require().GreaterOrEqual(len(results), 1)

	first, ok := results[0].(map[string]any)
	s.Require().True(ok)

	currentHostname, ok := first["hostname"].(string)
	s.Require().True(ok)

	tests := []struct {
		name         string
		args         []string
		validateFunc func(stdout string, exitCode int)
	}{
		{
			name: "updates hostname idempotently",
			args: []string{
				"client", "node", "hostname", "update",
				"--name", currentHostname,
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

				first, ok := results[0].(map[string]any)
				s.Require().True(ok)
				s.Contains(first, "hostname")
				s.Contains(first, "changed")
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

func TestNodeSmokeSuite(
	t *testing.T,
) {
	suite.Run(t, new(NodeSmokeSuite))
}

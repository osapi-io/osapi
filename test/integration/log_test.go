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

type LogSmokeSuite struct {
	suite.Suite
}

// TestLogQuery verifies the log query endpoint is reachable and returns
// a valid JSON response with a results array.
func (s *LogSmokeSuite) TestLogQuery() {
	tests := []struct {
		name         string
		args         []string
		validateFunc func(stdout string, exitCode int)
	}{
		{
			name: "query endpoint responds",
			args: []string{
				"client", "node", "log", "query",
				"--target", "_any",
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
				s.NotNil(results)
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

// TestLogUnit verifies the log unit endpoint is reachable and returns
// a valid JSON response with a results array.
func (s *LogSmokeSuite) TestLogUnit() {
	tests := []struct {
		name         string
		args         []string
		validateFunc func(stdout string, exitCode int)
	}{
		{
			name: "unit endpoint responds",
			args: []string{
				"client", "node", "log", "unit",
				"--target", "_any",
				"--name", "sshd.service",
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
				s.NotNil(results)
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

func TestLogSmokeSuite(
	t *testing.T,
) {
	suite.Run(t, new(LogSmokeSuite))
}

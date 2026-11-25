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

package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type DescribePublicTestSuite struct {
	suite.Suite
}

// describers are the groups of a part of osapi-io that describes itself:
// each DSL, the Go SDK and gohai's collectors. CONTRIBUTING.md of the specs
// says each has the same four commands.
var describers = []string{
	"dsl behavior", "dsl endpoint", "dsl command", "dsl chain", "sdk", "agent collector",
}

// TestFourCommands: each group that describes a part of osapi-io has list,
// get, spec get and validate, each with --json, so a person or an agent
// reads each part the same way.
func (s *DescribePublicTestSuite) TestFourCommands() {
	for _, g := range describers {
		for _, sub := range []string{"list", "get", "spec get", "validate"} {
			path := strings.Fields(g + " " + sub)
			c, rest, err := rootCmd.Find(path)
			s.Require().NoError(err, "osapi %s", strings.Join(path, " "))
			s.Empty(rest)
			s.Equal(path[len(path)-1], c.Name(), "osapi %s is a command", strings.Join(path, " "))
			s.NotNil(c.Flag("json"), "osapi %s has --json", strings.Join(path, " "))
		}
	}
}

func TestDescribePublicTestSuite(t *testing.T) {
	suite.Run(t, new(DescribePublicTestSuite))
}

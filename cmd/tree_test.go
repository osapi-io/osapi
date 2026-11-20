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

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
)

// TreeTestSuite holds that no two commands under one parent have one name.
// Cobra runs the first of them and hides the other: when a definition's
// group goes Under a command written by hand that AttachLocal is not given,
// the definition makes a second command of that name, and each command
// under the hand-written one is gone.
type TreeTestSuite struct {
	suite.Suite
}

func (s *TreeTestSuite) TestNamesUnique() {
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		seen := map[string]bool{}
		for _, sub := range c.Commands() {
			s.False(seen[sub.Name()], "%s has two %s commands", c.CommandPath(), sub.Name())
			seen[sub.Name()] = true
			walk(sub)
		}
	}
	walk(rootCmd)
}

// TestEveryBehaviorCommandIsAttached finds each behavior command in the
// binary's own tree. The behaviors' generated commands register through
// the import of internal/behaviors/all/gen in client.go: without it they
// are in no tree, and a test in their own package still passes.
func (s *TreeTestSuite) TestEveryBehaviorCommandIsAttached() {
	for _, b := range behavior.All() {
		if !b.CLI {
			continue
		}
		for _, m := range b.Methods {
			words := strings.Fields(b.Command(m))[1:]
			found, _, err := rootCmd.Find(words)
			s.Require().NoError(err, b.Command(m))
			s.Equal(m.Name, found.Name(), "%s is not in the tree", b.Command(m))
		}
	}
}

func TestTreeTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(TreeTestSuite))
}

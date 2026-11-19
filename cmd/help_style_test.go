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
	"unicode"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/command"
)

// HelpStyleTestSuite holds the help of every command to one style, whether
// a person wrote it or a definition made it: the one line and each flag's
// help start with a capital letter and have no period at the end, as
// cobra's own lines have none. The long help is sentences.
type HelpStyleTestSuite struct {
	suite.Suite
}

func (s *HelpStyleTestSuite) TestEveryCommand() {
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		path := c.CommandPath()
		if c.Name() == "help" || c.Name() == "completion" || c.Hidden {
			return
		}
		s.styled(path, "short help", c.Short)
		c.LocalFlags().VisitAll(func(f *pflag.Flag) {
			if f.Name != "help" && !f.Hidden {
				s.styled(path+" --"+f.Name, "flag help", f.Usage)
			}
		})
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	walk(rootCmd)
}

// TestCommandShape holds each command under osapi client, and each command
// of a local group, to <resource> <verb>, with the verbs and actions of the
// command layer, and each flag's help to its own words: a required flag is
// marked by MarkFlagRequired, which --help shows, not by its text.
func (s *HelpStyleTestSuite) TestCommandShape() {
	client, _, err := rootCmd.Find([]string{"client"})
	s.Require().NoError(err)
	roots := make([]*cobra.Command, 0, 1+len(command.All()))
	roots = append(roots, client)
	s.Require().NotEmpty(command.All())
	for _, g := range command.All() {
		c, _, err := rootCmd.Find(g.CommandPath())
		s.Require().NoError(err)
		roots = append(roots, c)
	}

	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		if c.Name() == "help" || c.Hidden {
			return
		}
		c.LocalFlags().VisitAll(func(f *pflag.Flag) {
			s.NotContainsf(
				f.Usage,
				"(required",
				"%s --%s: say that a flag is required with MarkFlagRequired",
				c.CommandPath(),
				f.Name,
			)
		})
		if len(c.Commands()) == 0 && c.Runnable() {
			s.Truef(command.Verbs[c.Name()] || command.Actions[c.Name()],
				"%s: %q is not a verb; a command is <resource> <verb>, such as facts keys list",
				c.CommandPath(), c.Name())
		}
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	for _, r := range roots {
		walk(r)
	}
}

// styled fails on text that is empty, starts with a small letter, or ends
// with a period.
func (s *HelpStyleTestSuite) styled(
	where string,
	what string,
	text string,
) {
	if text == "" {
		s.Failf("empty", "%s: %s is empty", where, what)

		return
	}
	first := []rune(text)[0]
	s.Falsef(unicode.IsLower(first), "%s: %s starts with a small letter: %q", where, what, text)
	s.Falsef(strings.HasSuffix(text, "."), "%s: %s ends with a period: %q", where, what, text)
}

func TestHelpStyleTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(HelpStyleTestSuite))
}

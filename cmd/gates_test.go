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
	"io/fs"
	"regexp"
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/skills"
	sdkclient "github.com/osapi-io/osapi/pkg/sdk/client"
)

// GatesTestSuite holds the gates of the command tree: the rules that each
// command of osapi keeps, which no one production file holds.
type GatesTestSuite struct {
	suite.Suite
}

func TestGatesTestSuite(t *testing.T) {
	suite.Run(t, new(GatesTestSuite))
}

// describers are the groups of a part of osapi-io that describes itself:
// each DSL, the Go SDK and gohai's collectors. CONTRIBUTING.md of the specs
// says each has the same four commands.
var describers = []string{
	"dsl behavior", "dsl endpoint", "dsl command", "dsl chain", "sdk", "agent collector", "config",
}

// TestFourCommands: each group that describes a part of osapi-io has list,
// get, spec get and validate, each with --json, so a person or an agent
// reads each part the same way.
func (s *GatesTestSuite) TestFourCommands() {
	for _, g := range describers {
		for _, sub := range []string{"list", "get", "spec get", "validate"} {
			path := strings.Fields(g + " " + sub)
			c, rest, err := rootCmd.Find(path)
			s.Require().NoError(err)
			s.Empty(rest)
			s.Equal(path[len(path)-1], c.Name())
			s.NotNil(c.Flag("json"))
		}
	}
}

// TestNamesUnique holds that no two commands under one parent have one name.
// Cobra runs the first of them and hides the other: when a definition's
// group goes Under a command written by hand that AttachLocal is not given,
// the definition makes a second command of that name, and each command
// under the hand-written one is gone.
func (s *GatesTestSuite) TestNamesUnique() {
	var twice []string
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		seen := map[string]bool{}
		for _, sub := range c.Commands() {
			if seen[sub.Name()] {
				twice = append(twice, sub.CommandPath())
			}
			seen[sub.Name()] = true
			walk(sub)
		}
	}
	walk(rootCmd)
	s.Empty(twice)
}

// TestEveryBehaviorCommandIsAttached finds each behavior command in the
// binary's own tree. The behaviors' generated commands register through
// the import of internal/behaviors/all/gen in client.go: without it they
// are in no tree, and a test in their own package still passes.
func (s *GatesTestSuite) TestEveryBehaviorCommandIsAttached() {
	for _, b := range behavior.All() {
		if !b.CLI {
			continue
		}
		for _, m := range b.Methods {
			words := strings.Fields(b.Command(m))[1:]
			found, _, err := rootCmd.Find(words)
			s.Require().NoError(err)
			s.Equal(b.Command(m)+": "+m.Name, b.Command(m)+": "+found.Name())
		}
	}
}

// userSkillsText is the text of each skill that a user installs.
func userSkillsText(t *testing.T) string {
	t.Helper()
	var b strings.Builder
	err := fs.WalkDir(skills.Installed(), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := fs.ReadFile(skills.Installed(), p)
		b.Write(data)

		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	return b.String()
}

// groups is the path of each command that has commands under it, below root.
func groups(c *cobra.Command, out map[string]bool) {
	for _, sub := range c.Commands() {
		if sub.Hidden || sub.Name() == "help" || sub.Name() == "completion" {
			continue
		}
		if sub.HasSubCommands() {
			out[strings.TrimPrefix(sub.CommandPath(), "osapi ")] = true
			groups(sub, out)
		}
	}
}

// taughtAsASet are the groups whose commands a skill teaches through a
// catalog of the binary, not one by one, and why.
var taughtAsASet = map[string]string{
	"client node": "behaviors: osapi dsl behavior list|get reads each one's definition, " +
		"and the osapi skill routes to it",
	"dsl": "authoring osapi: the dev skills teach it, to a person with the source",
}

// TestSkillsTeachEachGroup: each command group is named by a skill that a
// user installs, or is under a group that a catalog teaches, so a person with
// only the binary can learn it (osapi-io/osapi#1473).
func (s *GatesTestSuite) TestSkillsTeachEachGroup() {
	t := s.T()
	text := userSkillsText(t)
	all := map[string]bool{}
	groups(rootCmd, all)
	var missing []string
	for g := range all {
		if strings.Contains(text, "osapi "+g) || inSet(g) {
			continue
		}
		missing = append(missing, g)
	}
	sort.Strings(missing)
	// A group that no skill names goes in a reference of osapi-operations or
	// osapi, with its --json use.
	s.Empty(missing)
	for set := range taughtAsASet {
		s.Contains(text, "osapi "+set)
	}
}

// inSet is whether g is under a group that a catalog teaches.
func inSet(
	g string,
) bool {
	for set := range taughtAsASet {
		if strings.HasPrefix(g, set+" ") {
			return true
		}
	}

	return false
}

// TestSkillsTeachTheSDK: the SDK skill teaches each service from the SDK's
// catalog, and each service in it has a doc to teach.
func (s *GatesTestSuite) TestSkillsTeachTheSDK() {
	t := s.T()
	text := userSkillsText(t)
	for _, c := range []string{"osapi sdk list", "osapi sdk get"} {
		s.Contains(text, c)
	}
	var undocumented []string
	for _, svc := range sdkclient.Services() {
		if strings.TrimSpace(svc.Doc) == "" {
			undocumented = append(undocumented, svc.Name)
		}
	}
	s.Empty(undocumented)
}

// repoPath is a path into osapi's repository, which a user does not have.
var repoPath = regexp.MustCompile(
	`(^|[\s(\x60])(internal|cmd|test|ui/src|docs/docs)/[\w./-]+|\.go\b`,
)

// TestSkillsPointAtNoSource: a skill that a user installs names no file of
// the repository: a Go import path is the only path it has.
func (s *GatesTestSuite) TestSkillsPointAtNoSource() {
	t := s.T()
	var source []string
	for _, line := range strings.Split(userSkillsText(t), "\n") {
		if repoPath.MatchString(line) {
			source = append(source, strings.TrimSpace(line))
		}
	}
	s.Empty(source)
}

// RootTestSuite holds the help of every command to one style, whether
// a person wrote it or a definition made it: the one line and each flag's
// help start with a capital letter and have no period at the end, as
// cobra's own lines have none. The long help is sentences.
func (s *GatesTestSuite) TestEveryCommand() {
	// The root is a closed list: each word under it is one of
	// command.RootGroups, with its purpose as its help. help and completion
	// are cobra's own.
	for _, c := range rootCmd.Commands() {
		if c.Name() == "help" || c.Name() == "completion" {
			continue
		}
		r, ok := command.Root(c.Name())
		s.True(ok)
		s.Equal(c.Name()+": "+r.Purpose, c.Name()+": "+c.Short)
	}

	var bad []string
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		path := c.CommandPath()
		if c.Name() == "help" || c.Name() == "completion" || c.Hidden {
			return
		}
		bad = append(bad, styled(path+" short help", c.Short)...)
		c.LocalFlags().VisitAll(func(f *pflag.Flag) {
			if f.Name != "help" && !f.Hidden {
				bad = append(bad, styled(path+" --"+f.Name+" flag help", f.Usage)...)
			}
		})
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	walk(rootCmd)
	s.Empty(bad)
}

// TestCommandShape holds each command under osapi client, and each command
// of a local group, to <resource> <verb>, with the verbs and actions of the
// command layer, each word of the root to command.RootGroups, and each flag's help to its own words: a required flag is
// marked by MarkFlagRequired, which --help shows, not by its text.
func (s *GatesTestSuite) TestCommandShape() {
	client, _, err := rootCmd.Find([]string{"client"})
	s.Require().NoError(err)
	roots := make([]*cobra.Command, 0, 1+len(command.All()))
	roots = append(roots, client)
	s.Require().NotEmpty(command.All())
	for _, g := range command.All() {
		// A flat group adds no word: each of its commands is under its
		// parent, so each is a root of the walk.
		if g.Flat {
			for _, m := range g.Methods {
				c, _, err := rootCmd.Find(append(g.CommandPath(), m.Name))
				s.Require().NoError(err)
				roots = append(roots, c)
			}

			continue
		}
		c, _, err := rootCmd.Find(g.CommandPath())
		s.Require().NoError(err)
		roots = append(roots, c)
	}

	// A required flag says so with MarkFlagRequired, not in its text. A leaf
	// of the root is a leaf of command.RootGroups, and each other leaf is a
	// verb: <resource> <verb>, such as agent facts list.
	var bad []string
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		if c.Name() == "help" || c.Hidden {
			return
		}
		c.LocalFlags().VisitAll(func(f *pflag.Flag) {
			if strings.Contains(f.Usage, "(required") {
				bad = append(bad, c.CommandPath()+" --"+f.Name+": required in its text")
			}
		})
		if len(c.Commands()) == 0 && c.Runnable() && c.Parent() == rootCmd {
			if r, _ := command.Root(c.Name()); !r.Leaf {
				bad = append(bad, c.CommandPath()+": not a leaf of command.RootGroups")
			}
		} else if len(c.Commands()) == 0 && c.Runnable() &&
			!command.Verbs[c.Name()] && !command.Actions[c.Name()] {
			bad = append(bad, c.CommandPath()+": not a verb")
		}
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	for _, r := range roots {
		walk(r)
	}
	s.Empty(bad)
}

// styled names each way that text is wrong: it is empty, starts with a
// small letter, or ends with a period.
func styled(
	where string,
	text string,
) []string {
	if text == "" {
		return []string{where + ": empty"}
	}
	var bad []string
	if unicode.IsLower([]rune(text)[0]) {
		bad = append(bad, where+": starts with a small letter: "+text)
	}
	if strings.HasSuffix(text, ".") {
		bad = append(bad, where+": ends with a period: "+text)
	}

	return bad
}

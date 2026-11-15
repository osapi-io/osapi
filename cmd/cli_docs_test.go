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
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
)

// cliDocs is the directory of the CLI reference pages, relative to cmd/.
const cliDocs = "../docs/docs/sidebar/usage/cli"

type CLIDocsTestSuite struct {
	suite.Suite
}

// docPage is the page of a command, without its extension. A command with
// subcommands has a group page, <path>/<name>/<name>. A command without them
// has a leaf page, <path>/<name>. The root is cli.
func docPage(
	path []string,
	group bool,
) string {
	switch {
	case len(path) == 0:
		return "cli"
	case group:
		return filepath.Join(append(path, path[len(path)-1])...)
	default:
		return filepath.Join(path...)
	}
}

// documented reports if osapi documents the command on a CLI page. help and
// completion come from cobra. The commands of behaviors, under client next,
// have generated pages under behaviors/.
func documented(
	c *cobra.Command,
	path []string,
) bool {
	if c.Hidden || c.Name() == "help" || c.Name() == "completion" {
		return false
	}

	return strings.Join(path, " ") != "client next"
}

func walk(
	c *cobra.Command,
	path []string,
	pages map[string]bool,
) {
	var subs []*cobra.Command
	for _, sub := range c.Commands() {
		if documented(sub, append(path, sub.Name())) {
			subs = append(subs, sub)
		}
	}
	pages[docPage(path, len(subs) > 0)] = true
	for _, sub := range subs {
		walk(sub, append(append([]string{}, path...), sub.Name()), pages)
	}
}

func (s *CLIDocsTestSuite) TestEachCommandHasOnePageAndEachPageOneCommand() {
	want := map[string]bool{}
	walk(rootCmd, nil, want)

	have := map[string]bool{}
	err := filepath.WalkDir(cliDocs, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		ext := filepath.Ext(p)
		if ext != ".md" && ext != ".mdx" {
			return nil
		}
		rel, _ := filepath.Rel(cliDocs, p)
		have[strings.TrimSuffix(rel, ext)] = true

		return nil
	})
	s.Require().NoError(err)

	var missing, stale []string
	for p := range want {
		if !have[p] {
			missing = append(missing, p)
		}
	}
	for p := range have {
		if !want[p] {
			stale = append(stale, p)
		}
	}
	sort.Strings(missing)
	sort.Strings(stale)

	s.Empty(missing, "these commands have no page under %s", cliDocs)
	s.Empty(stale, "these pages under %s have no command: move or delete them", cliDocs)
}

func TestCLIDocsTestSuite(
	t *testing.T,
) {
	if _, err := os.Stat(cliDocs); err != nil {
		t.Skip("the docs are not in this checkout")
	}
	suite.Run(t, new(CLIDocsTestSuite))
}

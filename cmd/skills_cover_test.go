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

	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/skills"
	sdkclient "github.com/osapi-io/osapi/pkg/sdk/client"
)

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
func TestSkillsTeachEachGroup(t *testing.T) {
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
	for _, g := range missing {
		t.Errorf("osapi %s: no skill that a user installs names it; add it to a reference of "+
			"osapi-operations or osapi, with its --json use", g)
	}
	for set := range taughtAsASet {
		if !strings.Contains(text, "osapi "+set) {
			t.Errorf("osapi %s: a skill must teach the set", set)
		}
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
func TestSkillsTeachTheSDK(t *testing.T) {
	text := userSkillsText(t)
	for _, c := range []string{"osapi sdk list", "osapi sdk get"} {
		if !strings.Contains(text, c) {
			t.Errorf("the osapi-sdk skill must teach %s, the catalog of each service", c)
		}
	}
	for _, svc := range sdkclient.Services() {
		if strings.TrimSpace(svc.Doc) == "" {
			t.Errorf("SDK service %s has no doc in the catalog", svc.Name)
		}
	}
}

// repoPath is a path into osapi's repository, which a user does not have.
var repoPath = regexp.MustCompile(
	`(^|[\s(\x60])(internal|cmd|test|ui/src|docs/docs)/[\w./-]+|\.go\b`,
)

// TestSkillsPointAtNoSource: a skill that a user installs names no file of
// the repository: a Go import path is the only path it has.
func TestSkillsPointAtNoSource(t *testing.T) {
	for _, line := range strings.Split(userSkillsText(t), "\n") {
		if repoPath.MatchString(line) {
			t.Errorf("a user skill points at the source: %q", strings.TrimSpace(line))
		}
	}
}

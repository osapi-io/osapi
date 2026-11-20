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

package reference_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/reference"
)

type ReferencePublicTestSuite struct {
	suite.Suite
}

func TestReferencePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ReferencePublicTestSuite))
}

// tree is a small command tree: a root with a global flag, a group with a
// flag of its own, a leaf with flags and an example, and the commands that
// get no page.
func tree() *cobra.Command {
	root := &cobra.Command{Use: "osapi", Short: "Manage hosts"}
	root.PersistentFlags().BoolP("json", "j", false, "Enable JSON output")

	group := &cobra.Command{
		Use:     "node",
		Short:   "Send jobs to hosts.",
		Long:    "A target selects the hosts.",
		Example: "  osapi node status\n\n  osapi node status --target web",
	}
	group.PersistentFlags().String("target", "_all", "Target: <host> or {label}")
	group.Annotations = map[string]string{
		reference.Permission: "node:read",
		reference.Notes:      "Each command sends a job.",
	}

	leaf := &cobra.Command{
		Use:     "ssh-key",
		Short:   "Manage SSH keys",
		Long:    "Each `<key>` is one line. A key in <angle> brackets.",
		Example: "    osapi node ssh-key --name a \\\n      --force",
		Run:     func(*cobra.Command, []string) {},
		Annotations: map[string]string{
			reference.Result:     "It returns `changed: false` when the key <exists>.",
			reference.Permission: "node:write",
			reference.Notes:      "The agent writes the key file.",
		},
	}
	leaf.Flags().String("name", "", "Name of the key")
	leaf.Flags().String("required-one", "", "Needed")
	leaf.Flags().String("already", "", "Name (required)")
	leaf.Flags().Bool("force", false, "Replace the key | always")
	leaf.Flags().Int("count", 3, "Count")
	leaf.Flags().String("secret", "", "hidden")
	_ = leaf.Flags().MarkHidden("secret")
	_ = leaf.MarkFlagRequired("required-one")
	_ = leaf.MarkFlagRequired("already")

	plain := &cobra.Command{Use: "dns", Short: "Get DNS", Run: func(*cobra.Command, []string) {}}
	hidden := &cobra.Command{Use: "secret", Hidden: true, Run: func(*cobra.Command, []string) {}}

	group.AddCommand(leaf, plain, hidden)
	root.AddCommand(
		group,
		&cobra.Command{
			Use:   "version",
			Short: "Show the version",
			Run:   func(*cobra.Command, []string) {},
		},
	)
	root.InitDefaultHelpCmd()
	root.InitDefaultCompletionCmd()

	return root
}

func (s *ReferencePublicTestSuite) read(
	dir string,
	file string,
) string {
	b, err := os.ReadFile(filepath.Join(dir, file))
	s.Require().NoError(err)

	return string(b)
}

func (s *ReferencePublicTestSuite) TestWrite() {
	s.Run("when the tree has commands writes one page for each", func() {
		dir := s.T().TempDir()
		stale := filepath.Join(dir, "old", "gone.md")
		s.Require().NoError(os.MkdirAll(filepath.Dir(stale), 0o755))
		s.Require().NoError(os.WriteFile(stale, []byte("# Gone\n"), 0o644))
		s.Require().
			NoError(os.WriteFile(filepath.Join(dir, "output.md"), []byte("# Output\n"), 0o644))
		s.Require().NoError(os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("x"), 0o644))

		n, err := reference.Write(tree(), dir)
		s.Require().NoError(err)
		s.Positive(n, "it counts the pages it wrote")

		root := s.read(dir, "cli.mdx")
		s.Contains(root, "# CLI\n\nManage hosts.\n")
		s.Contains(root, "<DocCardList />")
		s.Contains(root, "| `-j, --json` | Enable JSON output |  |")

		group := s.read(dir, "node/node.mdx")
		s.Contains(group, "# Node\n\nSend jobs to hosts.\n\nA target selects the hosts.\n")
		s.Contains(group, "```bash\nosapi node status\n\nosapi node status --target web\n```")
		s.Contains(group, "Target: &lt;host&gt; or &#123;label&#125;")
		s.Contains(group, "`_all`")
		s.Contains(group, "## Permission\n\nEach command needs the `node:read` permission.")
		s.Contains(group, "## Notes\n\nEach command sends a job.")

		leaf := s.read(dir, "node/ssh-key.md")
		s.Contains(leaf, "# SSH Key\n")
		s.Contains(leaf, "Each `<key>` is one line. A key in &lt;angle&gt; brackets.")
		s.Contains(leaf, "osapi node ssh-key [flags]")
		s.Contains(leaf, "osapi node ssh-key --name a \\\n  --force")
		s.Contains(leaf, "| `--required-one` | Needed (required) |  |")
		s.Contains(leaf, "| `--already` | Name (required) |  |")
		s.Contains(leaf, "Replace the key \\| always")
		s.Contains(leaf, "| `--count` | Count | `3` |")
		s.NotContains(leaf, "--secret")
		s.Contains(leaf, "## Result\n\nIt returns `changed: false` when the key &lt;exists&gt;.")
		s.Contains(leaf, "## Permission\n\nIt needs the `node:write` permission.")
		s.Contains(leaf, "## Notes\n\nThe agent writes the key file.")
		s.Less(strings.Index(leaf, "## Notes"), strings.Index(leaf, "## Flags"))
		s.Contains(leaf, "## Global flags")
		s.Contains(leaf, "[Output](../output.md)")

		plain := s.read(dir, "node/dns.md")
		s.Contains(plain, "# DNS\n\nGet DNS.\n")
		s.NotContains(plain, "## Examples")

		s.FileExists(filepath.Join(dir, "version.md"))
		s.Contains(s.read(dir, "version.md"), "[Output](output.md)")
		s.NoFileExists(filepath.Join(dir, "node/secret.md"))
		s.NoFileExists(filepath.Join(dir, "help.md"))
		s.NoFileExists(stale)
		s.FileExists(filepath.Join(dir, "output.md"))
		s.FileExists(filepath.Join(dir, "notes.txt"))
	})

	s.Run("when a directory cannot be made returns the error", func() {
		dir := s.T().TempDir()
		s.Require().NoError(os.WriteFile(filepath.Join(dir, "node"), nil, 0o644))

		_, err := reference.Write(tree(), dir)

		s.ErrorContains(err, "make directory for ")
	})

	s.Run("when a page cannot be written returns the error", func() {
		dir := s.T().TempDir()
		s.Require().NoError(os.MkdirAll(filepath.Join(dir, "node", "node.mdx"), 0o755))

		_, err := reference.Write(tree(), dir)

		s.ErrorContains(err, "write ")
	})

	s.Run("when a stale page cannot be removed returns the error", func() {
		dir := s.T().TempDir()
		locked := filepath.Join(dir, "locked")
		s.Require().NoError(os.MkdirAll(locked, 0o755))
		s.Require().NoError(os.WriteFile(filepath.Join(locked, "stale.md"), nil, 0o644))
		s.Require().NoError(os.Chmod(locked, 0o500))
		s.T().Cleanup(func() { _ = os.Chmod(locked, 0o755) })

		_, err := reference.Write(&cobra.Command{Use: "osapi"}, dir)

		s.ErrorContains(err, "remove stale page locked/stale.md")
	})

	s.Run("when the walk fails returns the error", func() {
		dir := s.T().TempDir()
		closed := filepath.Join(dir, "closed")
		s.Require().NoError(os.MkdirAll(closed, 0o755))
		s.Require().NoError(os.Chmod(closed, 0o000))
		s.T().Cleanup(func() { _ = os.Chmod(closed, 0o755) })

		_, err := reference.Write(&cobra.Command{Use: "osapi"}, dir)

		s.Error(err)
	})
}

func (s *ReferencePublicTestSuite) TestTitle() {
	tests := []struct {
		name string
		want string
	}{
		{name: "node", want: "Node"},
		{name: "ssh-key", want: "SSH Key"},
		{name: "nats", want: "NATS"},
		{name: "list-sources", want: "List Sources"},
	}

	for _, tc := range tests {
		s.Run("when the name is "+tc.name+" gives "+tc.want, func() {
			s.Equal(tc.want, reference.Title(tc.name))
		})
	}
}

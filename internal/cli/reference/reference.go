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

// Package reference writes the CLI reference pages of the docs site from the
// cobra command tree. Each command has one page: a group page that lists its
// subcommands, or a leaf page with its usage, examples and flags.
package reference

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// The keys of the annotations of a command that its page shows. --help does
// not show them, so a page has detail that a terminal does not need.
const (
	// Result tells what the command reports: when it returns changed: true
	// or false, and when it fails.
	Result = "osapi.docs.result"
	// Permission is the permission that the command needs, for example
	// node:write.
	Permission = "osapi.docs.permission"
	// Notes is more detail for the docs site. Use it rarely.
	Notes = "osapi.docs.notes"
)

// Keep names the pages in the directory that a person writes. Write does not
// delete them.
var Keep = []string{"output.md"}

// acronyms are the words that a title writes in capitals.
var acronyms = map[string]bool{
	"api": true, "cli": true, "dns": true, "id": true, "jwt": true,
	"nats": true, "ntp": true, "os": true, "pki": true, "ssh": true,
	"ui": true, "url": true,
}

// Skip tells whether a command has no page. Hidden commands, help and
// completion have none.
type Skip func(c *cobra.Command, path []string) bool

// Write writes the page of each command under root into dir, and removes each
// page in dir that no command has, except the pages in Keep.
func Write(
	root *cobra.Command,
	dir string,
	skip Skip,
) error {
	written := map[string]bool{}
	if err := writeTree(root, nil, dir, skip, written); err != nil {
		return err
	}

	return removeStale(dir, written)
}

func documented(
	c *cobra.Command,
	path []string,
	skip Skip,
) bool {
	if c.Hidden || c.Name() == "help" || c.Name() == "completion" {
		return false
	}

	return skip == nil || !skip(c, path)
}

func writeTree(
	c *cobra.Command,
	path []string,
	dir string,
	skip Skip,
	written map[string]bool,
) error {
	var subs []*cobra.Command
	for _, sub := range c.Commands() {
		if documented(sub, append(append([]string{}, path...), sub.Name()), skip) {
			subs = append(subs, sub)
		}
	}

	file := pageFile(path, len(subs) > 0)
	if err := writePage(filepath.Join(dir, file), Page(c, path, subs)); err != nil {
		return err
	}
	written[file] = true

	for _, sub := range subs {
		if err := writeTree(sub, append(append([]string{}, path...), sub.Name()), dir, skip, written); err != nil {
			return err
		}
	}

	return nil
}

// pageFile is the file of the page of the command at path, relative to the
// directory of the reference. A group page is named as its directory.
func pageFile(
	path []string,
	group bool,
) string {
	switch {
	case len(path) == 0:
		return "cli.mdx"
	case group:
		return filepath.Join(append(path, path[len(path)-1]+".mdx")...)
	default:
		return filepath.Join(path...) + ".md"
	}
}

func writePage(
	file string,
	content string,
) error {
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return fmt.Errorf("make directory for %s: %w", file, err)
	}
	if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", file, err)
	}

	return nil
}

func removeStale(
	dir string,
	written map[string]bool,
) error {
	keep := map[string]bool{}
	for _, k := range Keep {
		keep[k] = true
	}

	return filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		ext := filepath.Ext(p)
		if ext != ".md" && ext != ".mdx" {
			return nil
		}
		rel, _ := filepath.Rel(dir, p)
		if written[rel] || keep[rel] {
			return nil
		}
		if err := os.Remove(p); err != nil {
			return fmt.Errorf("remove stale page %s: %w", rel, err)
		}

		return nil
	})
}

// Title is the title of the page of a command: its name, with each word
// capitalized and each acronym in capitals.
func Title(
	name string,
) string {
	words := strings.Split(name, "-")
	for i, w := range words {
		if acronyms[w] {
			words[i] = strings.ToUpper(w)
			continue
		}
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}

	return strings.Join(words, " ")
}

// Page returns the markdown of the page of c. subs are its documented
// subcommands: a command that has some gets a group page.
func Page(
	c *cobra.Command,
	path []string,
	subs []*cobra.Command,
) string {
	var b strings.Builder

	title := "CLI"
	if len(path) > 0 {
		title = Title(c.Name())
	}
	fmt.Fprintf(&b, "# %s\n\n", title)
	if c.Short != "" {
		fmt.Fprintf(&b, "%s.\n\n", mdx(strings.TrimSuffix(c.Short, ".")))
	}
	if c.Long != "" {
		fmt.Fprintf(&b, "%s\n\n", mdx(strings.TrimSpace(c.Long)))
	}

	if len(subs) > 0 || len(path) == 0 {
		if c.Example != "" {
			fmt.Fprintf(&b, "```bash\n%s\n```\n\n", strings.TrimRight(dedent(c.Example), "\n"))
		}
		b.WriteString("import DocCardList from '@theme/DocCardList';\n\n<DocCardList />\n")
		if p := c.Annotations[Permission]; p != "" {
			fmt.Fprintf(
				&b,
				"\n## Permission\n\nEach command needs the `%s` permission. `osapi token roles` lists the roles that grant it.\n",
				p,
			)
		}
		if n := c.Annotations[Notes]; n != "" {
			fmt.Fprintf(&b, "\n## Notes\n\n%s\n", mdx(n))
		}
		if t := flagTable(c.LocalNonPersistentFlags(), c.PersistentFlags()); t != "" {
			fmt.Fprintf(&b, "\n## Flags\n\n%s", t)
		}

		return b.String()
	}

	fmt.Fprintf(&b, "```bash\n%s\n```\n", c.UseLine())

	if c.Example != "" {
		fmt.Fprintf(
			&b,
			"\n## Examples\n\n```bash\n%s\n```\n",
			strings.TrimRight(dedent(c.Example), "\n"),
		)
	}
	if r := c.Annotations[Result]; r != "" {
		fmt.Fprintf(&b, "\n## Result\n\n%s\n", mdx(r))
	}
	if p := c.Annotations[Permission]; p != "" {
		fmt.Fprintf(
			&b,
			"\n## Permission\n\nIt needs the `%s` permission. `osapi token roles` lists the roles that grant it.\n",
			p,
		)
	}
	if n := c.Annotations[Notes]; n != "" {
		fmt.Fprintf(&b, "\n## Notes\n\n%s\n", mdx(n))
	}
	if t := flagTable(c.LocalFlags()); t != "" {
		fmt.Fprintf(&b, "\n## Flags\n\n%s", t)
	}
	if t := flagTable(c.InheritedFlags()); t != "" {
		fmt.Fprintf(&b, "\n## Global flags\n\n%s", t)
	}

	fmt.Fprintf(
		&b,
		"\n[Output](%s) tells how the result is printed, and what `--json` gives.\n",
		strings.Repeat("../", len(path)-1)+"output.md",
	)

	return b.String()
}

// flagTable is a markdown table of the flags in sets, sorted by name. It is
// empty if there are none.
func flagTable(
	sets ...*pflag.FlagSet,
) string {
	seen := map[string]bool{}
	var flags []*pflag.Flag
	for _, set := range sets {
		set.VisitAll(func(f *pflag.Flag) {
			if f.Hidden || f.Name == "help" || seen[f.Name] {
				return
			}
			seen[f.Name] = true
			flags = append(flags, f)
		})
	}
	if len(flags) == 0 {
		return ""
	}
	sort.Slice(flags, func(i, j int) bool { return flags[i].Name < flags[j].Name })

	var b strings.Builder
	b.WriteString("| Flag | Description | Default |\n| --- | --- | --- |\n")
	for _, f := range flags {
		name := "`--" + f.Name + "`"
		if f.Shorthand != "" {
			name = "`-" + f.Shorthand + ", --" + f.Name + "`"
		}
		usage := mdx(strings.ReplaceAll(f.Usage, "|", `\|`))
		if _, ok := f.Annotations[cobra.BashCompOneRequiredFlag]; ok &&
			!strings.Contains(strings.ToLower(usage), "required") {
			usage += " (required)"
		}
		fmt.Fprintf(&b, "| %s | %s | %s |\n", name, usage, defaultValue(f))
	}

	return b.String()
}

// defaultValue is the default of f in code format, or empty for a zero value.
func defaultValue(
	f *pflag.Flag,
) string {
	switch f.DefValue {
	case "", "false", "0", "[]", "0s":
		return ""
	}

	return "`" + f.DefValue + "`"
}

// dedent removes the indent that the lines of s have in common.
func dedent(
	s string,
) string {
	lines := strings.Split(strings.Trim(s, "\n"), "\n")
	indent := -1
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		n := len(l) - len(strings.TrimLeft(l, " \t"))
		if indent < 0 || n < indent {
			indent = n
		}
	}
	for i, l := range lines {
		if len(l) >= indent {
			lines[i] = l[indent:]
		}
	}

	return strings.Join(lines, "\n")
}

// mdxEscapes are the characters that MDX reads as the start of a tag or of an
// expression.
var mdxEscapes = strings.NewReplacer("<", "&lt;", ">", "&gt;", "{", "&#123;", "}", "&#125;")

// mdx escapes the characters of s that MDX would parse, outside code spans.
func mdx(
	s string,
) string {
	parts := strings.Split(s, "`")
	for i := 0; i < len(parts); i += 2 {
		parts[i] = mdxEscapes.Replace(parts[i])
	}

	return strings.Join(parts, "`")
}

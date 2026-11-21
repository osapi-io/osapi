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

package clidocs

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/cli/command"

	"github.com/osapi-io/osapi/internal/guide"

	"github.com/osapi-io/osapi/internal/cli/reference"
)

// Generate writes the page of each command of the binary into the
// directory, and removes each page there that no command has.
func Generate(
	_ context.Context,
	d Deps,
	in GenerateInput,
) (GenerateOutput, error) {
	dir := in.Dir
	if dir == "" {
		dir = "docs/docs/sidebar/usage/cli"
	}
	n, err := reference.Write(d.Root, dir)
	if err != nil {
		return GenerateOutput{}, fmt.Errorf("write the CLI reference into %s: %w", dir, err)
	}

	return GenerateOutput{Dir: dir, Pages: n}, nil
}

// GenerateDSL writes a page for each DSL of the binary into in.Dir, and removes each
// generated page there that no DSL has. A page that a person wrote, with no
// guide.Generated line, stays.
func GenerateDSL(
	_ context.Context,
	d Deps,
	in DSLInput,
) (DSLOutput, error) {
	dir := in.Dir
	if dir == "" {
		dir = "docs/docs/sidebar/development/dsl"
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return DSLOutput{}, fmt.Errorf("write the DSL pages into %s: %w", dir, err)
	}
	written := map[string]bool{}
	for _, l := range d.DSLs() {
		name := l.Slug + ".md"
		page := l.Spec.Markdown(l.Title, l.Intro, l.Position)
		if err := os.WriteFile(filepath.Join(dir, name), []byte(page), 0o644); err != nil {
			return DSLOutput{}, fmt.Errorf("write the DSL pages into %s: %w", dir, err)
		}
		written[name] = true
	}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if written[e.Name()] || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		if data, err := os.ReadFile(
			path,
		); err == nil &&
			strings.Contains(string(data), guide.Generated) {
			_ = os.Remove(path)
		}
	}

	return DSLOutput{Dir: dir, Pages: len(written)}, nil
}

// referenceDirs are where the reference pages go. A generated page in one
// of them that no command writes now is removed.
var referenceDirs = []string{
	"docs/docs/sidebar/reference/behaviors",
	"docs/docs/sidebar/reference/chains",
	"docs/docs/sidebar/reference/endpoints",
	"docs/docs/sidebar/reference/cli",
}

// GenerateReference writes each page of the reference, from the outputs of
// the commands that show it, and removes each generated page there that no
// command writes now.
func GenerateReference(
	_ context.Context,
	d Deps,
	in ReferenceInput,
) (ReferenceOutput, error) {
	root := in.Root
	if root == "" {
		root = "."
	}
	pages, err := d.Pages()
	if err != nil {
		return ReferenceOutput{}, err
	}
	written := map[string]bool{}
	for _, p := range pages {
		text, err := page(p)
		if err != nil {
			return ReferenceOutput{}, fmt.Errorf("%s: %w", p.Path, err)
		}
		path := filepath.Join(root, p.Path)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return ReferenceOutput{}, err
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			return ReferenceOutput{}, err
		}
		written[filepath.Clean(path)] = true
	}
	for _, dir := range referenceDirs {
		_ = filepath.WalkDir(
			filepath.Join(root, dir),
			func(path string, e fs.DirEntry, err error) error {
				if err != nil || e.IsDir() || written[filepath.Clean(path)] {
					return nil
				}
				if data, err := os.ReadFile(
					path,
				); err == nil &&
					strings.Contains(string(data), guide.Generated) {
					_ = os.Remove(path)
				}

				return nil
			},
		)
	}

	return ReferenceOutput{Pages: len(pages)}, nil
}

// page is the text of p: its front matter and title, its intro, then each
// part as the CLI prints it.
func page(
	p Page,
) (string, error) {
	var sb strings.Builder
	if p.Partial {
		sb.WriteString(
			"{/* " + strings.TrimSuffix(
				strings.TrimPrefix(guide.Generated, "<!-- "),
				" -->",
			) + " */}\n",
		)
	} else {
		fmt.Fprintf(&sb, "---\ntitle: %s\n---\n\n%s\n\n# %s\n", p.Title, guide.Generated, p.Title)
	}
	if p.Intro != "" {
		sb.WriteString("\n" + p.Intro + "\n")
	}
	for _, part := range p.Parts {
		md := &cli.Markdown{Title: part.Heading, Link: part.Link}
		if err := command.Render(md, part.Group, part.Out); err != nil {
			return "", err
		}
		sb.WriteString(md.String())
	}

	return sb.String(), nil
}

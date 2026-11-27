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
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/osfs"
	"github.com/osapi-io/osapi/internal/atomicfile"

	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/cli/command"

	"github.com/osapi-io/osapi/internal/guide"

	"github.com/osapi-io/osapi/internal/cli/reference"
)

// hostFS is the filesystem that the generator reads and writes, which a
// test replaces.
var hostFS avfs.VFS = osfs.NewWithNoIdm()

// The directories of the generated pages, under the root of the repository.
const (
	cliDir = "docs/docs/sidebar/usage/cli"
	dslDir = "docs/docs/sidebar/development/dsl"
)

// Generate writes each generated page of the docs site under in.Root: the
// page of each command, of each DSL and of the reference.
func Generate(
	ctx context.Context,
	d Deps,
	in GenerateInput,
) (GenerateOutput, error) {
	root := in.Root
	if root == "" {
		root = "."
	}
	c, err := WriteCLI(ctx, d, CLIInput{Dir: filepath.Join(root, cliDir)})
	if err != nil {
		return GenerateOutput{}, err
	}
	ds, err := WriteDSL(ctx, d, DSLInput{Dir: filepath.Join(root, dslDir)})
	if err != nil {
		return GenerateOutput{}, err
	}
	ref, err := WriteReference(ctx, d, ReferenceInput{Root: root})
	if err != nil {
		return GenerateOutput{}, err
	}

	return GenerateOutput{CLI: c.Pages, DSL: ds.Pages, Reference: ref.Pages}, nil
}

// WriteCLI writes the page of each command of the binary into the
// directory, and removes each page there that no command has.
func WriteCLI(
	_ context.Context,
	d Deps,
	in CLIInput,
) (CLIOutput, error) {
	dir := in.Dir
	if dir == "" {
		dir = cliDir
	}
	n, err := reference.Write(d.Root, dir)
	if err != nil {
		return CLIOutput{}, fmt.Errorf("write the CLI reference into %s: %w", dir, err)
	}

	return CLIOutput{Dir: dir, Pages: n}, nil
}

// WriteDSL writes a page for each DSL of the binary into in.Dir, and removes each
// generated page there that no DSL has. A page that a person wrote, with no
// guide.Generated line, stays.
func WriteDSL(
	_ context.Context,
	d Deps,
	in DSLInput,
) (DSLOutput, error) {
	dir := in.Dir
	if dir == "" {
		dir = dslDir
	}
	if err := hostFS.MkdirAll(dir, 0o755); err != nil {
		return DSLOutput{}, fmt.Errorf("write the DSL pages into %s: %w", dir, err)
	}
	written := map[string]bool{}
	for _, l := range d.DSLs() {
		name := l.Slug + ".md"
		page := l.Spec.Markdown(l.Title, l.Intro, l.Position)
		if err := atomicfile.Write(
			context.Background(),
			hostFS,
			filepath.Join(dir, name),
			[]byte(page),
			0o644,
		); err != nil {
			return DSLOutput{}, fmt.Errorf("write the DSL pages into %s: %w", dir, err)
		}
		written[name] = true
	}
	entries, _ := hostFS.ReadDir(dir)
	for _, e := range entries {
		if written[e.Name()] || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		if data, err := avfs.ReadFile(hostFS,
			path,
		); err == nil &&
			strings.Contains(string(data), guide.Generated) {
			_ = hostFS.Remove(path)
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

// categoryFile is the file that gives a folder of the docs site its label
// and its place in the sidebar.
const categoryFile = "_category_.json"

// WriteReference writes each page of the reference, from the outputs of
// the commands that show it, and removes each generated page there that no
// command writes now. A _category_.json in a folder under a reference
// directory is generated too: one that no command writes now is removed,
// then each folder that is empty.
func WriteReference(
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
		if err := hostFS.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return ReferenceOutput{}, err
		}
		if err := atomicfile.Write(
			context.Background(),
			hostFS,
			path,
			[]byte(text),
			0o644,
		); err != nil {
			return ReferenceOutput{}, err
		}
		written[filepath.Clean(path)] = true
	}
	for _, dir := range referenceDirs {
		top := filepath.Join(root, dir)
		var folders []string
		_ = filepath.WalkDir(
			top,
			func(path string, e fs.DirEntry, err error) error {
				if err != nil || written[filepath.Clean(path)] {
					return nil
				}
				if e.IsDir() {
					folders = append(folders, path)

					return nil
				}
				if e.Name() == categoryFile && filepath.Dir(path) != top {
					_ = hostFS.Remove(path)

					return nil
				}
				if data, err := avfs.ReadFile(hostFS,
					path,
				); err == nil &&
					strings.Contains(string(data), guide.Generated) {
					_ = hostFS.Remove(path)
				}

				return nil
			},
		)
		// The walk lists a folder before the folders in it, so the last
		// is removed first; a folder that is not empty stays.
		for i := len(folders) - 1; i > 0; i-- {
			_ = hostFS.Remove(folders[i])
		}
	}

	if err := writeTables(root, d.Tables); err != nil {
		return ReferenceOutput{}, err
	}

	return ReferenceOutput{Pages: len(pages)}, nil
}

// tableDirs are where pages that a person wrote have generated tables.
var tableDirs = []string{
	"docs/docs/sidebar/features",
	"docs/docs/sidebar/operate",
	"docs/docs/sidebar/architecture",
}

// writeTables writes each generated table of the pages in tableDirs again,
// and keeps the rest of each page as it is.
func writeTables(
	root string,
	tables func(string, string) (string, error),
) error {
	if tables == nil {
		return nil
	}
	for _, dir := range tableDirs {
		err := filepath.WalkDir(
			filepath.Join(root, dir),
			func(path string, e fs.DirEntry, err error) error {
				// A tree with no such directory has no tables in it.
				if errors.Is(err, fs.ErrNotExist) {
					return nil
				}
				if err != nil || e.IsDir() || !strings.HasSuffix(path, ".md") {
					return err
				}
				data, err := avfs.ReadFile(hostFS, path)
				if err != nil {
					return err
				}
				text, err := tables(root, string(data))
				if err != nil {
					return fmt.Errorf("%s: %w", path, err)
				}
				if text == string(data) {
					return nil
				}

				return atomicfile.Write(context.Background(), hostFS, path, []byte(text), 0o644)
			},
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// page is the text of p: its front matter and title, its intro, then each
// part as the CLI prints it.
func page(
	p Page,
) (string, error) {
	if p.Text != "" {
		return p.Text, nil
	}
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

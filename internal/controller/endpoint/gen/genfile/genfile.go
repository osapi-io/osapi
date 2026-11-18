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

// Package genfile writes generated files for the endpoint generator and the
// behavior generator alike: it removes the files a run made before, writes
// each file under the repository's root, groups a Go file's imports as the
// repository does, and formats every Go file it wrote once, at the end.
package genfile

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/dave/jennifer/jen"
	"golang.org/x/tools/imports"
)

// module is the import path of the repository, the third group of imports.
const module = "github.com/osapi-io/osapi"

// Writer writes generated files under Root, a path relative to the working
// directory, such as ../../.. from a go:generate line.
type Writer struct {
	Root string
	gofs []string
}

// Clean removes each file of dir that a generator made, by the comment that
// heads it, so a method or a group that is gone leaves nothing behind. A
// file that a person wrote stays.
func (w *Writer) Clean(
	dir string,
) error {
	entries, err := os.ReadDir(filepath.Join(w.Root, dir))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, e := range entries {
		p := filepath.Join(w.Root, dir, e.Name())
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if bytes.HasPrefix(data, []byte("// Code generated")) ||
			bytes.HasPrefix(data, []byte("# Code generated")) {
			if err := os.Remove(p); err != nil {
				return err
			}
		}
	}

	return nil
}

// Write writes data at rel under the root, with the directories it needs.
func (w *Writer) Write(
	rel string,
	data []byte,
) error {
	p := filepath.Join(w.Root, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}

	return os.WriteFile(p, data, 0o644)
}

// Go renders f to rel, with its imports in the repository's three groups:
// the standard library, then other modules, then this one. Format formats
// it at the end.
func (w *Writer) Go(
	rel string,
	f *jen.File,
) error {
	var buf bytes.Buffer
	if err := f.Render(&buf); err != nil {
		return fmt.Errorf("%s: %w", rel, err)
	}
	imports.LocalPrefix = module
	src, err := imports.Process(rel, buf.Bytes(), &imports.Options{
		Comments: true, TabIndent: true, TabWidth: 8, FormatOnly: true,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", rel, err)
	}
	w.gofs = append(w.gofs, rel)

	return w.Write(rel, src)
}

// Remove removes rel under the root, when it is there.
func (w *Writer) Remove(
	rel string,
) error {
	if err := os.Remove(filepath.Join(w.Root, rel)); err != nil &&
		!errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}

// Format runs the repository's formatters, golines with gofumpt, over every
// Go file written, once.
func (w *Writer) Format() error {
	if len(w.gofs) == 0 {
		return nil
	}
	args := append([]string{
		"tool", "github.com/segmentio/golines",
		"--base-formatter=go tool mvdan.cc/gofumpt", "-w",
	}, w.gofs...)
	cmd := exec.Command("go", args...)
	cmd.Dir = w.Root
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("format: %w: %s", err, out)
	}

	return nil
}

// Registry is the file at dir/gen.go that imports each of pkgs for its init,
// so a binary registers each one with a single import, such as each
// endpoint group's routes. The file has header and the package comment doc.
func (w *Writer) Registry(
	dir string,
	header string,
	doc []string,
	pkgs []string,
) error {
	sorted := append([]string(nil), pkgs...)
	sort.Strings(sorted)
	f := jen.NewFilePathName(module+"/"+filepath.ToSlash(dir), filepath.Base(dir))
	f.HeaderComment(header)
	for _, line := range doc {
		f.PackageComment(line)
	}
	for _, p := range sorted {
		f.Anon(p)
	}
	if err := w.Clean(dir); err != nil {
		return err
	}

	return w.Go(filepath.Join(dir, "gen.go"), f)
}

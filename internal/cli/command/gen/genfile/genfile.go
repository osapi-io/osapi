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

// Package genfile writes generated files for the local, endpoint and
// behavior generators alike: it removes the files a run made before, writes
// each file under the repository's root, and groups a Go file's imports as
// the repository does.
//
// A Writer changes nothing on disk until Flush. Clean, Write, Go, Remove and
// Registry only record what to do, so a run that fails or panics part way,
// such as on a literal that jennifer cannot write, leaves the files of the
// last run as they were. Flush writes each file to a temporary file beside
// it and renames it, then removes the stale ones.
package genfile

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/osfs"
	"github.com/osapi-io/osapi/internal/atomicfile"

	"github.com/dave/jennifer/jen"
	"golang.org/x/tools/imports"
)

// hostFS is the filesystem that the generator reads and writes, which a
// test replaces.
var hostFS avfs.VFS = osfs.NewWithNoIdm()

// module is the import path of the repository, the third group of imports.
const module = "github.com/osapi-io/osapi"

// Writer writes generated files under Root, a path relative to the working
// directory, such as ../../.. from a go:generate line. Call Flush when the
// whole run succeeds.
type Writer struct {
	Root string

	// files are the files to write, by their path under Root.
	files map[string][]byte
	// stale are the files to remove, by their path under Root, unless the
	// run writes them again.
	stale map[string]bool
}

// Clean removes each file of dir that a generator made, by the comment that
// heads it, so a method or a group that is gone leaves nothing behind. A
// file that a person wrote stays.
func (w *Writer) Clean(
	dir string,
) error {
	entries, err := hostFS.ReadDir(filepath.Join(w.Root, dir))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, e := range entries {
		p := filepath.Join(w.Root, dir, e.Name())
		data, err := avfs.ReadFile(hostFS, p)
		if err != nil {
			return err
		}
		if bytes.HasPrefix(data, []byte("// Code generated")) ||
			bytes.HasPrefix(data, []byte("# Code generated")) {
			w.removeLater(filepath.Join(dir, e.Name()))
		}
	}

	return nil
}

// removeLater records that Flush removes rel, unless the run writes it.
func (w *Writer) removeLater(
	rel string,
) {
	if w.stale == nil {
		w.stale = map[string]bool{}
	}
	w.stale[filepath.Clean(rel)] = true
}

// Write records data for rel under the root. Flush writes it.
func (w *Writer) Write(
	rel string,
	data []byte,
) error {
	if w.files == nil {
		w.files = map[string][]byte{}
	}
	rel = filepath.Clean(rel)
	w.files[rel] = data
	delete(w.stale, rel)

	return nil
}

// Go renders f to rel, with its imports in the repository's three groups:
// the standard library, then other modules, then this one. That is all the
// format it gets: golangci-lint skips a generated file, so no check reads its
// format.
func (w *Writer) Go(
	rel string,
	f *jen.File,
) (err error) {
	// jennifer panics on a literal of a type it cannot write, such as a
	// named string type.
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("%s: %v", rel, p)
		}
	}()
	var buf bytes.Buffer
	var src []byte
	err = f.Render(&buf)
	if err == nil {
		imports.LocalPrefix = module
		src, err = imports.Process(rel, buf.Bytes(), &imports.Options{
			Comments: true, TabIndent: true, TabWidth: 8, FormatOnly: true,
		})
	}
	if err != nil {
		return fmt.Errorf("%s: %w", rel, err)
	}

	return w.Write(rel, src)
}

// Remove records that Flush removes rel under the root, when it is there.
func (w *Writer) Remove(
	rel string,
) error {
	rel = filepath.Clean(rel)
	delete(w.files, rel)
	w.removeLater(rel)

	return nil
}

// Flush makes the changes that the run recorded: it writes each file, with
// the directories it needs, through a temporary file and a rename, then
// removes each stale file. Call it only when the whole run succeeded.
func (w *Writer) Flush() error {
	paths := make([]string, 0, len(w.files))
	for rel := range w.files {
		paths = append(paths, rel)
	}
	sort.Strings(paths)
	for _, rel := range paths {
		if err := w.writeFile(rel, w.files[rel]); err != nil {
			return err
		}
	}
	stale := make([]string, 0, len(w.stale))
	for rel := range w.stale {
		stale = append(stale, rel)
	}
	sort.Strings(stale)
	for _, rel := range stale {
		if err := hostFS.Remove(filepath.Join(w.Root, rel)); err != nil &&
			!errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	w.files, w.stale = nil, nil

	return nil
}

// writeFile writes data at rel under the root through a temporary file in
// the same directory, which atomicfile renames over rel.
func (w *Writer) writeFile(
	rel string,
	data []byte,
) error {
	p := filepath.Join(w.Root, rel)
	if err := hostFS.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}

	return atomicfile.Write(context.Background(), hostFS, p, data, 0o644)
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

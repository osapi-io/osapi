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

package behaviorcmd

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	gofmt "go/format"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

//go:embed scaffold
var scaffoldFS embed.FS

// scaffoldFiles is where the templates are read from; a test swaps it.
var scaffoldFiles fs.FS = scaffoldFS

// module is the module a scaffold is written into.
const module = "github.com/osapi-io/osapi"

// allFile registers every behavior, one blank import each.
const allFile = "internal/behaviors/all/all.go"

// scaffoldName is a behavior name that is also a Go package name.
var scaffoldName = regexp.MustCompile(`^[a-z][a-z0-9]*$`)

// Scaffold writes a new behavior's standard layout under
// internal/behaviors/<name>/, or internal/behaviors/<group>/<name>/ for a
// behavior nested under a group, in the osapi checkout at root, and registers it in
// internal/behaviors/all. It returns the files written, relative to root, the
// registration last. Everything is rendered and checked before anything is
// written, so a taken name or a root that is not a checkout writes nothing.
func Scaffold(
	root string,
	group string,
	name string,
	doc string,
) ([]string, error) {
	if !scaffoldName.MatchString(name) {
		return nil, fmt.Errorf(
			"name %q must be lowercase letters and digits, starting with a letter", name)
	}
	if group != "" && !scaffoldName.MatchString(group) {
		return nil, fmt.Errorf(
			"group %q must be lowercase letters and digits, starting with a letter", group)
	}
	if doc == "" || strings.ContainsAny(doc, "\r\n") {
		return nil, errors.New("description must be one line")
	}

	mod, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil || !strings.HasPrefix(string(mod), "module "+module+"\n") {
		return nil, fmt.Errorf("%s is not an osapi checkout", root)
	}

	// A grouped behavior lives under its group, as its routes and commands
	// do: internal/behaviors/network/dns.
	path := name
	if group != "" {
		path = group + "/" + name
	}
	dir := filepath.Join("internal", "behaviors", filepath.FromSlash(path))
	if _, err := os.Stat(filepath.Join(root, dir)); err == nil {
		return nil, fmt.Errorf("%s already exists", dir)
	}

	all, err := os.ReadFile(filepath.Join(root, allFile))
	if err != nil {
		return nil, err
	}
	registered, err := register(all, path)
	if err != nil {
		return nil, err
	}

	var tmpls []string
	_ = fs.WalkDir(scaffoldFiles, "scaffold", func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".tmpl") {
			tmpls = append(tmpls, p)
		}

		return nil
	})
	type file struct {
		rel  string
		data []byte
	}
	files := make([]file, 0, len(tmpls)+1)
	for _, t := range tmpls {
		src, err := renderScaffold(t, group, name, path, doc)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", t, err)
		}
		rel := filepath.Join(dir, strings.TrimSuffix(strings.TrimPrefix(t, "scaffold/"), ".tmpl"))
		files = append(files, file{rel, src})
	}
	files = append(files, file{allFile, registered})

	written := make([]string, 0, len(files))
	for _, f := range files {
		p := filepath.Join(root, f.rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return written, err
		}
		if err := os.WriteFile(p, f.data, 0o644); err != nil {
			return written, err
		}
		written = append(written, f.rel)
	}

	return written, nil
}

// renderScaffold fills one template and formats the result as Go.
func renderScaffold(
	tmpl string,
	group string,
	name string,
	path string,
	doc string,
) ([]byte, error) {
	t, err := template.ParseFS(scaffoldFiles, tmpl)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	data := struct{ Group, Name, Path, Upper, Doc string }{
		group, name, path, strings.ToUpper(name), doc,
	}
	if err := t.Execute(&buf, data); err != nil {
		return nil, err
	}

	return gofmt.Source(buf.Bytes())
}

// register adds the blank import of the behavior at path, under
// internal/behaviors, to internal/behaviors/all, keeping the imports sorted.
func register(
	src []byte,
	path string,
) ([]byte, error) {
	i := bytes.Index(src, []byte("import ("))
	if i < 0 {
		return nil, fmt.Errorf("%s has no import block", allFile)
	}
	// The new line goes last in the block, after any comment that heads it,
	// and formatting sorts it among the others.
	end := i + bytes.Index(src[i:], []byte("\n)")) + 1
	line := "\t_ \"" + module + "/internal/behaviors/" + path + "\"\n"

	out := make([]byte, 0, len(src)+len(line))
	out = append(out, src[:end]...)
	out = append(out, line...)
	out = append(out, src[end:]...)

	return gofmt.Source(out)
}

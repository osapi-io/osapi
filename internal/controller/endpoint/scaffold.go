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

package endpoint

import (
	"bytes"
	"errors"
	"fmt"
	gofmt "go/format"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// module is the module a scaffold is written into.
const module = "github.com/osapi-io/osapi"

// allFile registers every endpoint group, one blank import each.
const allFile = "internal/controller/endpoints/all/all.go"

// Scaffold writes a new group's layout under
// internal/controller/endpoints/<name>/ in the osapi checkout at root, and
// registers it in internal/controller/endpoints/all. It returns the files
// written, relative to root, the registration last. Everything is rendered
// and checked before anything is written, so a taken name or a root that is
// not a checkout writes nothing.
func Scaffold(
	root string,
	name string,
	doc string,
) ([]string, error) {
	if !regexp.MustCompile(`^[a-z][a-z0-9]*$`).MatchString(name) {
		return nil, fmt.Errorf(
			"name %q must be lowercase letters and digits, starting with a letter", name)
	}
	if doc == "" || strings.ContainsAny(doc, "\r\n") {
		return nil, errors.New("description must be one line")
	}

	mod, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil || !strings.HasPrefix(string(mod), "module "+module+"\n") {
		return nil, fmt.Errorf("%s is not an osapi checkout", root)
	}

	dir := filepath.Join("internal", "controller", "endpoints", name)
	if _, err := os.Stat(filepath.Join(root, dir)); err == nil {
		return nil, fmt.Errorf("%s already exists", dir)
	}

	all, err := os.ReadFile(filepath.Join(root, allFile))
	if err != nil {
		return nil, err
	}
	registered, err := register(all, name)
	if err != nil {
		return nil, err
	}

	type file struct {
		rel  string
		data []byte
	}
	sources := scaffoldSources(groupScaffold{name: name, doc: doc})
	files := make([]file, 0, len(sources)+1)
	for _, src := range sources {
		data, err := renderGo(src.file)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", src.rel, err)
		}
		files = append(files, file{filepath.Join(dir, src.rel), data})
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

// register adds the blank import of the group name to
// internal/controller/endpoints/all, keeping the imports sorted.
func register(
	src []byte,
	name string,
) ([]byte, error) {
	i := bytes.Index(src, []byte("import ("))
	if i < 0 {
		return nil, fmt.Errorf("%s has no import block", allFile)
	}
	end := i + bytes.Index(src[i:], []byte("\n)")) + 1
	line := "\t_ \"" + module + "/internal/controller/endpoints/" + name + "\"\n"

	out := make([]byte, 0, len(src)+len(line))
	out = append(out, src[:end]...)
	out = append(out, line...)
	out = append(out, src[end:]...)

	return gofmt.Source(out)
}

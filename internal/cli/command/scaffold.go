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
package command

import (
	"bytes"
	"errors"
	"fmt"
	gofmt "go/format"
	"os"
	"path/filepath"
	"strings"
)

// module is the module a scaffold is written into.
const module = "github.com/osapi-io/osapi"

// allFile registers every local group, one blank import each.
const allFile = "internal/cli/commands/all/all.go"

// parentIdents are the Parent constants that a scaffold writes, by word.
var parentIdents = map[Parent]string{
	ParentAgent:      "ParentAgent",
	ParentController: "ParentController",
	ParentDSL:        "ParentDSL",
	ParentNATS:       "ParentNATS",
	ParentPKI:        "ParentPKI",
}

// parentWords are the words of KnownParents, for a message.
func parentWords() []string {
	out := make([]string, 0, len(KnownParents()))
	for _, p := range KnownParents() {
		out = append(out, string(p))
	}

	return out
}

// Scaffold writes a new local group's layout under
// internal/cli/commands/<name>/ in the osapi checkout at root, Under the
// word of the root that under names, such as controller, and
// registers it in internal/cli/commands/all. It returns the files
// written, relative to root, the registration last. Everything is rendered
// and checked before anything is written, so a taken name or a root that is
// not a checkout writes nothing.
func Scaffold(
	root string,
	group string,
	doc string,
	under Parent,
) ([]string, error) {
	if parentIdents[under] == "" {
		return nil, fmt.Errorf(
			"under %q is not a word of the root that a local group goes under: %s",
			under, strings.Join(parentWords(), ", "))
	}
	if !name.MatchString(group) {
		return nil, fmt.Errorf(
			"name %q must be lowercase letters and digits, starting with a letter", group)
	}
	if doc == "" || strings.ContainsAny(doc, "\r\n") {
		return nil, errors.New("description must be one line")
	}

	mod, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil || !strings.HasPrefix(string(mod), "module "+module+"\n") {
		return nil, fmt.Errorf("%s is not an osapi checkout", root)
	}

	dir := filepath.Join("internal", "cli", "commands", group)
	if _, err := os.Stat(filepath.Join(root, dir)); err == nil {
		return nil, fmt.Errorf("%s already exists", dir)
	}

	all, err := os.ReadFile(filepath.Join(root, allFile))
	if err != nil {
		return nil, err
	}
	registered, err := register(all, group)
	if err != nil {
		return nil, err
	}

	type file struct {
		rel  string
		data []byte
	}
	sources := scaffoldSources(groupScaffold{name: group, doc: doc, under: under})
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
// internal/cli/commands/all, keeping the imports sorted.
func register(
	src []byte,
	name string,
) ([]byte, error) {
	i := bytes.Index(src, []byte("import ("))
	if i < 0 {
		return nil, fmt.Errorf("%s has no import block", allFile)
	}
	end := i + bytes.Index(src[i:], []byte("\n)")) + 1
	line := "\t_ \"" + module + "/internal/cli/commands/" + name + "\"\n"

	out := make([]byte, 0, len(src)+len(line))
	out = append(out, src[:end]...)
	out = append(out, line...)
	out = append(out, src[end:]...)

	return gofmt.Source(out)
}

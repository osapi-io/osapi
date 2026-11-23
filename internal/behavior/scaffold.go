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

package behavior

import (
	"bytes"
	"errors"
	"fmt"
	gofmt "go/format"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// module is the module a scaffold is written into.
const module = "github.com/osapi-io/osapi"

// allFile registers every behavior, one blank import each.
const allFile = "internal/behaviors/all/all.go"

// scaffoldName is a behavior name or a Go package name of a new behavior:
// lowercase letters and digits.
var scaffoldName = regexp.MustCompile(`^[a-z][a-z0-9]*$`)

// Scaffold writes a new behavior's standard layout under
// internal/behaviors/<name>/, or internal/behaviors/<group>/<name>/ for a
// behavior nested under a group, in the osapi checkout at root, and registers it in
// internal/behaviors/all. It returns the files written, relative to root, the
// registration last. Everything is rendered and checked before anything is
// written, so a taken name or a root that is not a checkout writes nothing.
//
// requires names registered behaviors that the new one is built on. The
// scaffold writes the wiring of each one: behavior.Requires in the
// definition, host.Require in the builder and a field of the provider. The
// author then writes the behavior.Call of each, which the tests ask for.
func Scaffold(
	root string,
	group string,
	name string,
	goPackage string,
	doc string,
	requires []string,
) ([]string, error) {
	if !scaffoldName.MatchString(name) {
		return nil, fmt.Errorf(
			"name %q must be lowercase letters and digits, starting with a letter", name)
	}
	// The Go package is what the caller names; with none it is the name.
	// Nothing changes a name to make it a Go package name.
	if goPackage == "" {
		goPackage = name
	}
	if !scaffoldName.MatchString(goPackage) {
		return nil, fmt.Errorf(
			"package %q must be lowercase letters and digits, starting with a letter", goPackage)
	}
	if token.IsKeyword(goPackage) {
		return nil, fmt.Errorf(
			"%q is a Go keyword, so it cannot be the Go package: pass --package with another name",
			goPackage)
	}
	if group != "" && !scaffoldName.MatchString(group) {
		return nil, fmt.Errorf(
			"group %q must be lowercase letters and digits, starting with a letter", group)
	}
	if group != "" && !slices.Contains(KnownGroups(), GroupName(group)) {
		return nil, fmt.Errorf(
			"group %q is not one of KnownGroups; add its constant to KnownGroups first", group)
	}
	if doc == "" || strings.ContainsAny(doc, "\r\n") {
		return nil, errors.New("description must be one line")
	}
	reqs, err := resolveRequires(requires)
	if err != nil {
		return nil, err
	}

	mod, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil || !strings.HasPrefix(string(mod), "module "+module+"\n") {
		return nil, fmt.Errorf("%s is not an osapi checkout", root)
	}

	// A grouped behavior lives under its group, as its routes and commands
	// do: internal/behaviors/network/dns.
	path := goPackage
	if group != "" {
		path = group + "/" + goPackage
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

	type file struct {
		rel  string
		data []byte
	}
	sources := scaffoldSources(
		scaffold{
			group: group, name: name, goPackage: goPackage, path: path, doc: doc,
			requires: reqs,
		},
	)
	files := make([]file, 0, len(sources)+1)
	for _, src := range sources {
		data, err := renderGo(src.file)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", src.rel, err)
		}
		files = append(files, file{filepath.Join(dir, filepath.FromSlash(src.rel)), data})
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

// requirement is a behavior that a new behavior requires, as the scaffold
// writes its wiring.
type requirement struct {
	// name is the behavior's name in the registry.
	name string
	// path is the import path of its Go package.
	path string
	// pkg is its Go package name.
	pkg string
	// fn is the Go function of one of its methods, such as Update, for the
	// example call in the provider's comment.
	fn string
}

// resolveRequires finds each named behavior in the registry. A name that is
// not registered, or that is given twice, is refused.
func resolveRequires(
	names []string,
) ([]requirement, error) {
	reqs := make([]requirement, 0, len(names))
	for i, n := range names {
		if slices.Contains(names[:i], n) {
			return nil, fmt.Errorf("requires %q twice: name each behavior once", n)
		}
		b := Lookup(n)
		if b == nil {
			return nil, fmt.Errorf(
				"requires %q, which is not a registered behavior: "+
					"osapi behavior list names each one", n)
		}
		if len(b.Methods) == 0 {
			return nil, fmt.Errorf("requires %q, which has no method to call", n)
		}
		// The Go function of a method, such as
		// github.com/osapi-io/osapi/internal/behaviors/managedfile.Update,
		// is the package's import path, a dot and the function.
		fn := b.Methods[0].Func()
		slash := strings.LastIndex(fn, "/") + 1
		dot := slash + strings.Index(fn[slash:], ".")
		reqs = append(reqs, requirement{
			name: b.Name,
			path: fn[:dot],
			pkg:  fn[slash:dot],
			fn:   fn[dot+1:],
		})
	}

	return reqs, nil
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

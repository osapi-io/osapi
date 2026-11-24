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

// Command gen writes the typed services of the Go SDK, pkg/sdk/client, from
// the full spec of the API, internal/controller/api/gen/api.yaml. For each
// service, <service>.gen.go has its methods, and <service>_types.gen.go the
// types only it sends and receives; types.gen.go has the types that more
// than one service has, and services.gen.go puts each service on the
// Client. A method sends its request through the SDK's Send, and decodes
// the answer into the SDK's types: the SDK uses nothing that oapi-codegen
// writes. just sdk-generate runs it, after go generate writes the spec. It
// is not a go:generate directive: go generate lists the files of each
// package before it runs, so a run that removed the file of a group that is
// gone made go generate fail on that file (osapi-io/osapi#1261).
//
// The same run writes a program for each service whose methods have
// examples, examples/sdk/client/<service>.go: it calls each method with the
// input of each example of the method's definition, which the spec has in
// x-osapi-examples. With -docs, it writes the SDK's reference pages under
// docs/docs/sidebar/sdk instead, from the same model and the methods
// of the SDK's source, and the table of services of client.md.
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/getkin/kin-openapi/openapi3"

	"github.com/osapi-io/osapi/internal/cli/command/gen/genfile"
	"github.com/osapi-io/osapi/internal/sdk/gen/sdkmodel"
)

// spec is the full spec of the API, from the root of the repository.
const spec = "internal/controller/api/gen/api.yaml"

// dir is the SDK's package, from the root of the repository.
const dir = "pkg/sdk/client"

func main() {
	root := flag.String("root", ".", "the root of the repository")
	docs := flag.Bool("docs", false, "write the SDK's reference pages, not the SDK")
	flag.Parse()
	if err := generate(*root, *docs); err != nil {
		fmt.Fprintln(os.Stderr, "sdk gen:", err)
		os.Exit(1)
	}
}

// generate writes the SDK and its examples, or with docs the SDK's
// reference pages, from the spec under root.
func generate(
	root string,
	docs bool,
) error {
	doc, err := openapi3.NewLoader().LoadFromFile(filepath.Join(root, spec))
	if err != nil {
		return err
	}
	m, err := sdkmodel.New(doc)
	if err != nil {
		return err
	}
	if docs {
		return runPages(root, m)
	}

	return run(root, m)
}

// run writes the SDK and a program of the examples of each service.
func run(
	root string,
	m *sdkmodel.Model,
) error {
	w := &genfile.Writer{Root: root}
	written, err := clean(w, root)
	if err != nil {
		return err
	}
	out := map[string]*jen.File{}
	for path, f := range files(m, written) {
		out[filepath.Join(dir, path)] = f
	}
	if err := w.Clean(exampleDir); err != nil {
		return err
	}
	for _, s := range m.Services {
		f, err := exampleFile(s)
		if err != nil {
			return err
		}
		if f == nil {
			continue
		}
		rel := filepath.Join(exampleDir, snake(s.Name)+".go")
		if data, err := os.ReadFile(filepath.Join(root, rel)); err == nil &&
			!strings.HasPrefix(string(data), header) {
			return fmt.Errorf("%s is a person's: the example of %s goes there", rel, s.Name)
		}
		out[rel] = f
	}
	for rel, f := range out {
		if err := w.Go(rel, f); err != nil {
			return err
		}
	}

	return w.Flush()
}

// clean removes each file of the SDK that a run of this generator made, by
// the comment that heads it, so a service that is gone leaves nothing
// behind. permissions.gen.go, which a different generator makes, stays. It
// returns the types that the other files of the package declare: a type a
// person wrote, such as Step, is the SDK's, and the generator does not write
// it.
func clean(
	w *genfile.Writer,
	root string,
) (map[string]bool, error) {
	entries, err := os.ReadDir(filepath.Join(root, dir))
	if err != nil {
		return nil, err
	}
	written := map[string]bool{}
	fset := token.NewFileSet()
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		path := filepath.Join(root, dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if strings.HasPrefix(string(data), header) {
			if err := w.Remove(filepath.Join(dir, e.Name())); err != nil {
				return nil, err
			}

			continue
		}
		if strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, path, data, parser.SkipObjectResolution)
		if err != nil {
			return nil, err
		}
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, spec := range gd.Specs {
				written[spec.(*ast.TypeSpec).Name.Name] = true
			}
		}
	}

	return written, nil
}

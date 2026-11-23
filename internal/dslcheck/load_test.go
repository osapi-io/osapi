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

package dslcheck_test

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"

	"golang.org/x/tools/go/packages"
)

// checked is one package, parsed and type-checked.
type checked struct {
	Fset  *token.FileSet
	Info  *types.Info
	Files []*ast.File
}

// loadChecked parses and type-checks the packages that patterns match, one
// at a time. packages.Load type-checks them in parallel, and go/types races
// when two checkers read one imported type that it completes lazily
// (osapi-io/osapi#1141). Here each package gets its own importer, which reads
// its imports from their export data, and the next package starts only after
// the one before it is done.
func loadChecked(
	dir string,
	patterns ...string,
) ([]checked, error) {
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedCompiledGoFiles | packages.NeedImports |
			packages.NeedDeps | packages.NeedExportFile,
		Dir: dir,
	}, patterns...)
	if err != nil {
		return nil, err
	}

	exports := map[string]string{}
	packages.Visit(pkgs, nil, func(p *packages.Package) {
		exports[p.PkgPath] = p.ExportFile
	})
	lookup := func(path string) (io.ReadCloser, error) {
		f, ok := exports[path]
		if !ok || f == "" {
			return nil, fmt.Errorf("no export data for %s", path)
		}

		return os.Open(f)
	}

	out := make([]checked, 0, len(pkgs))
	for _, p := range pkgs {
		if len(p.Errors) > 0 {
			return nil, fmt.Errorf("%s: %v", p.PkgPath, p.Errors)
		}
		fset := token.NewFileSet()
		files := make([]*ast.File, 0, len(p.CompiledGoFiles))
		for _, name := range p.CompiledGoFiles {
			f, err := parser.ParseFile(fset, name, nil, parser.ParseComments)
			if err != nil {
				return nil, err
			}
			files = append(files, f)
		}
		info := &types.Info{
			Types: map[ast.Expr]types.TypeAndValue{},
			Defs:  map[*ast.Ident]types.Object{},
			Uses:  map[*ast.Ident]types.Object{},
		}
		conf := types.Config{Importer: importer.ForCompiler(fset, "gc", lookup)}
		if _, err := conf.Check(p.PkgPath, fset, files, info); err != nil {
			return nil, fmt.Errorf("%s: %w", p.PkgPath, err)
		}
		out = append(out, checked{Fset: fset, Info: info, Files: files})
	}

	return out, nil
}

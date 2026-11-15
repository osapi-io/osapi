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

package behaviortest

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/osapi-io/osapi/internal/behavior"
)

// sourceDir is where Run reads the Go files of the behavior under test: the
// directory of the test, which is the package of the behavior.
var sourceDir = "."

// directCalls returns a problem for each call, in the Go files of dir, of a
// method of a behavior that b requires, made without behavior.Call. Such a
// call works, but skips the span and the metrics of the method, so a trace
// does not show it as a step. A test of a behavior runs in the directory of
// its package, so Run reads ".".
func directCalls(
	b *behavior.Behavior,
	dir string,
) []string {
	methods := map[string]string{}
	for _, r := range b.Requires {
		for _, m := range r.Methods {
			methods[m.Func()] = r.Name + " " + m.Name
		}
	}
	if len(methods) == 0 {
		return nil
	}

	files, _ := filepath.Glob(filepath.Join(dir, "*.go"))
	fset := token.NewFileSet()
	var problems []string
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, name, nil, parser.SkipObjectResolution)
		if err != nil {
			problems = append(problems, fmt.Sprintf("read %s: %v", name, err))
			continue
		}
		imports := map[string]string{}
		for _, imp := range f.Imports {
			p, _ := strconv.Unquote(imp.Path.Value)
			local := path.Base(p)
			if imp.Name != nil {
				local = imp.Name.Name
			}
			imports[local] = p
		}
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			pkg, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}
			if m, ok := methods[imports[pkg.Name]+"."+sel.Sel.Name]; ok {
				problems = append(problems, fmt.Sprintf(
					"%s calls %s.%s directly, so the trace has no step for %s; call it with behavior.Call",
					fset.Position(call.Pos()),
					pkg.Name,
					sel.Sel.Name,
					m,
				))
			}

			return true
		})
	}

	return problems
}

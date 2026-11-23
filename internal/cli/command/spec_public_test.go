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

package command_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/command"
)

type SpecPublicTestSuite struct {
	suite.Suite
}

// notDSL are the exported names of the package that a definition does not
// write: what reads a definition, and what the registry and the commands
// use.
var notDSL = map[string]bool{
	"Attach":                       true,
	"AttachLocal":                  true,
	"AddCommands":                  true,
	"Origin":                       true,
	"Output":                       true,
	"Report":                       true,
	"ReadFile":                     true,
	"SetTraceID":                   true,
	"TraceID":                      true,
	"Lookup":                       true,
	"All":                          true,
	"Run":                          true,
	"Serve":                        true,
	"IsProcess":                    true,
	"ConfigKeys":                   true,
	"Check":                        true,
	"FlagName":                     true,
	"Validate":                     true,
	"Scaffold":                     true,
	"Refusals":                     true,
	"Handler":                      true,
	"MethodName":                   true,
	"Input":                        true,
	"CommandName":                  true,
	"SubPath":                      true,
	"CommandShort":                 true,
	"CommandLong":                  true,
	"Examples":                     true,
	"PageNotes":                    true,
	"PageResult":                   true,
	"Method:LocalGroup":            true,
	"GroupName:LocalGroup":         true,
	"GroupDoc:LocalGroup":          true,
	"GroupNotes:LocalGroup":        true,
	"GroupLong:LocalGroup":         true,
	"SubDoc:LocalGroup":            true,
	"Path:LocalGroup":              true,
	"CommandPath:LocalGroup":       true,
	"Commands:LocalGroup":          true,
	"CommandPermission:LocalGroup": true,
	"Error:UsageError":             true,
	"Error:FailedError":            true,
	"String:Problem":               true,
	"Rendered":                     true,
	"Reason":                       true,
	"Groups":                       true,
	"ValidateOutput":               true,
	"TagKeys":                      true,
	"KnownParents":                 true,
	"InputFlags":                   true,
	"ExampleFlagProblems":          true,
	"Now":                          true,
	"SetNow":                       true,
	"Derive":                       true,
	"Done":                         true,
	"Columns":                      true,
	"Format:Column":                true,
	"FormatKV:Column":              true,
	"PrintsGroups":                 true,
	"Key":                          true,
	"Saying":                       true,
}

// TestSpecCoversTheDSL fails on an exported function of the package, or an
// exported method of LocalMethod, that the spec has no term for. Thus osapi
// cli command spec get cannot fall behind the code.
func (s *SpecPublicTestSuite) TestSpecCoversTheDSL() {
	entries, err := os.ReadDir(".")
	s.Require().NoError(err)
	fset := token.NewFileSet()
	var files []*ast.File
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, e.Name(), nil, 0)
		s.Require().NoError(err)
		files = append(files, f)
	}

	var names []string
	for _, f := range files {
		for _, d := range f.Decls {
			fn, ok := d.(*ast.FuncDecl)
			if !ok || !fn.Name.IsExported() {
				continue
			}
			name := fn.Name.Name
			if fn.Recv != nil {
				recv := fn.Recv.List[0].Type
				if star, ok := recv.(*ast.StarExpr); ok {
					recv = star.X
				}
				id, _ := recv.(*ast.Ident)
				if id == nil || !id.IsExported() {
					continue
				}
				if id.Name != "LocalMethod" && id.Name != "Renderer" {
					name += ":" + id.Name
				}
			}
			if !notDSL[name] {
				names = append(names, name)
			}
		}
	}
	s.NotEmpty(names)

	for _, name := range names {
		found := false
		for _, t := range command.Spec.Terms {
			if strings.Contains(t.Name, name+"(") || strings.Contains(t.Name, name+"[") {
				found = true
			}
		}
		s.True(found, "the spec has no term for %s", name)
	}
}

// TestSpecHasEveryDeps: the type of each field of each group's Deps has a
// term of the kind deps, so an author learns what a Deps can take and how a
// test replaces it.
func (s *SpecPublicTestSuite) TestSpecHasEveryDeps() {
	files, err := filepath.Glob("../commands/*/deps.go")
	s.Require().NoError(err)
	s.Require().NotEmpty(files)
	deps := command.Spec.Of(command.KindDeps)
	names := make([]string, 0, len(deps))
	for _, t := range deps {
		names = append(names, t.Name)
	}
	terms := strings.Join(names, "\n")
	for _, file := range files {
		f, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
		s.Require().NoError(err)
		ast.Inspect(f, func(n ast.Node) bool {
			spec, ok := n.(*ast.TypeSpec)
			if !ok || spec.Name.Name != "Deps" {
				return true
			}
			for _, field := range spec.Type.(*ast.StructType).Fields.List {
				kind := depsKind(field.Type)
				s.Contains(
					terms,
					kind,
					"%s: a Deps field of type %s has no term of the kind deps; add one",
					file,
					kind,
				)
			}

			return false
		})
	}
}

// depsKind is the term that names the type of a Deps field.
func depsKind(
	expr ast.Expr,
) string {
	switch t := expr.(type) {
	case *ast.FuncType:
		return "func(...)"
	case *ast.Ident:
		if t.Name == "string" {
			return "string"
		}

		return "an interface of the group"
	case *ast.StarExpr:
		return "*" + depsKind(t.X)
	case *ast.SelectorExpr:
		if t.Sel.Name == "Deps" {
			return "another group's Deps"
		}

		return t.X.(*ast.Ident).Name + "." + t.Sel.Name
	}

	return fmt.Sprintf("%T", expr)
}

func (s *SpecPublicTestSuite) TestSpecKinds() {
	for _, k := range command.Spec.Kinds {
		s.NotEmpty(command.Spec.Of(k.Kind), "kind %s has no terms", k.Kind)
	}
	for _, t := range command.Spec.Terms {
		s.NotEmpty(t.Does, t.Name)
		s.NotEmpty(t.When, t.Name)
	}
}

func TestSpecPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(SpecPublicTestSuite))
}

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

package endpoint_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/controller/endpoint"
)

type SpecPublicTestSuite struct {
	suite.Suite
}

// notDSL are the exported names of the package that a definition does not
// write: what reads a definition, and what the registry and the commands
// use.
var notDSL = map[string]bool{
	"OpenAPI":              true,
	"Validate":             true,
	"Lookup":               true,
	"All":                  true,
	"Path:Group":           true,
	"Permission:Group":     true,
	"Key":                  true,
	"OperationID":          true,
	"Handler":              true,
	"Mapped":               true,
	"SuccessStatus":        true,
	"LocationField":        true,
	"Answering":            true,
	"Method:Group":         true,
	"CommandShort":         true,
	"CommandLong":          true,
	"CommandPath:Group":    true,
	"Examples":             true,
	"Refusals":             true,
	"Scaffold":             true,
	"PageNotes":            true,
	"PageResult":           true,
	"CommandFlags":         true,
	"SubPath":              true,
	"CommandName":          true,
	"Error:InternalError":  true,
	"Unwrap:InternalError": true,
	"Error:AnswerError":    true,
	"String:Problem":       true,
}

// TestSpecCoversTheDSL fails on an exported function of the package, or an
// exported method of Method, that the spec has no term for. Thus osapi
// controller endpoint spec cannot fall behind the code.
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
				if id.Name != "Method" {
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
		for _, t := range endpoint.Spec.Terms {
			if strings.Contains(t.Name, name+"(") || strings.Contains(t.Name, name+"[") {
				found = true
			}
		}
		s.True(found, "the spec has no term for %s", name)
	}
}

func (s *SpecPublicTestSuite) TestSpecKinds() {
	for _, k := range endpoint.Spec.Kinds {
		s.NotEmpty(endpoint.Spec.Of(k.Kind), "kind %s has no terms", k.Kind)
	}
	for _, t := range endpoint.Spec.Terms {
		s.NotEmpty(t.Does, t.Name)
		s.NotEmpty(t.When, t.Name)
	}
}

func TestSpecPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(SpecPublicTestSuite))
}

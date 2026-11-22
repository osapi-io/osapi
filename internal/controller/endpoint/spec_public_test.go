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
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

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
	"OpenAPI":                 true,
	"Validate":                true,
	"Lookup":                  true,
	"All":                     true,
	"Path:Group":              true,
	"Permission:Group":        true,
	"Declared:Group":          true,
	"Key":                     true,
	"OperationID":             true,
	"Handler":                 true,
	"Mapped":                  true,
	"SuccessStatus":           true,
	"LocationField":           true,
	"Answering":               true,
	"Saying":                  true,
	"Output":                  true,
	"Method:Group":            true,
	"CommandShort":            true,
	"CommandLong":             true,
	"CommandPath:Group":       true,
	"Examples":                true,
	"Refusals":                true,
	"Scaffold":                true,
	"PageNotes":               true,
	"PageResult":              true,
	"SubPath":                 true,
	"CommandName":             true,
	"MethodName":              true,
	"Input":                   true,
	"GroupName:Group":         true,
	"GroupDoc:Group":          true,
	"GroupNotes:Group":        true,
	"GroupLong:Group":         true,
	"SubDoc:Group":            true,
	"Commands:Group":          true,
	"CommandPermission:Group": true,
	"HTTPMethod":              true,
	"Error:InternalError":     true,
	"Unwrap:InternalError":    true,
	"Error:AnswerError":       true,
	"String:Problem":          true,
}

// TestSpecCoversTheDSL fails on an exported function of the package, or an
// exported method of Method, that the spec has no term for. Thus osapi
// controller endpoint spec get cannot fall behind the code.
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
			// A refinement of another type, such as Because on Error, is
			// written as a call on a value: .Because(reason).
			if base, _, ok := strings.Cut(name, ":"); ok && strings.Contains(t.Name, "."+base+"(") {
				found = true
			}
		}
		s.True(found, "the spec has no term for %s", name)
	}
}

// TestSpecHasEveryDeps: each field of each group's Deps has a term of the
// kind deps, so an author learns each service of the controller that a
// handler can take, and how its tests replace it.
func (s *SpecPublicTestSuite) TestSpecHasEveryDeps() {
	files, err := filepath.Glob("../endpoints/*/deps.go")
	s.Require().NoError(err)
	s.Require().NotEmpty(files)
	for _, file := range files {
		f, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
		s.Require().NoError(err)
		ast.Inspect(f, func(n ast.Node) bool {
			spec, ok := n.(*ast.TypeSpec)
			if !ok || spec.Name.Name != "Deps" {
				return true
			}
			for _, field := range spec.Type.(*ast.StructType).Fields.List {
				for _, name := range field.Names {
					word := regexp.MustCompile(`(^|[^A-Za-z])` + name.Name + `([^A-Za-z]|$)`)
					found := false
					for _, t := range endpoint.Spec.Of(endpoint.KindDeps) {
						found = found || word.MatchString(t.Name)
					}
					s.True(
						found,
						"%s: Deps.%s has no term of the kind deps; add one",
						file,
						name.Name,
					)
				}
			}

			return false
		})
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

// TestReasonsMatchTheSchema: the reasons that the server sends are the enum
// of the ErrorResponse schema, in its order, and the schema says what each
// one means, so a client's typed reason and the server cannot drift.
func (s *SpecPublicTestSuite) TestReasonsMatchTheSchema() {
	raw, err := os.ReadFile("../api/components/gen/api.yaml")
	s.Require().NoError(err)
	var doc struct {
		Components struct {
			Schemas map[string]struct {
				Properties map[string]struct {
					Description string   `yaml:"description"`
					Enum        []string `yaml:"enum"`
				} `yaml:"properties"`
			} `yaml:"schemas"`
		} `yaml:"components"`
	}
	s.Require().NoError(yaml.Unmarshal(raw, &doc))
	reason := doc.Components.Schemas["ErrorResponse"].Properties["reason"]
	want := make([]string, 0, len(endpoint.Reasons))
	for _, r := range endpoint.Reasons {
		want = append(want, string(r))
	}
	s.Equal(want, reason.Enum)
	for _, r := range endpoint.Reasons {
		s.Contains(reason.Description, r+":", "the schema says what %s means", r)
	}
}

// TestReasons: each helper gives the reason of its status, Because names
// another, and a status that no error names has the reason of its class.
func (s *SpecPublicTestSuite) TestReasons() {
	err := errors.New("x")
	s.Equal(endpoint.ReasonNotFound, endpoint.NotFound(err, "").Reason)
	s.Equal(endpoint.ReasonConflict, endpoint.Conflict(err, "").Reason)
	s.Equal(endpoint.ReasonNotManaged,
		endpoint.Conflict(err, "").Because(endpoint.ReasonNotManaged).Reason)
	s.Equal(endpoint.ReasonInvalidRequest, endpoint.Invalid(err, "").Reason)
	s.Equal(endpoint.ReasonForbidden, endpoint.Forbidden(err, "").Reason)
	s.Equal(endpoint.ReasonUnavailable, endpoint.Unavailable(err, "").Reason)
	for status, want := range map[int]endpoint.Reason{
		401: endpoint.ReasonUnauthorized, 403: endpoint.ReasonForbidden,
		404: endpoint.ReasonNotFound, 405: endpoint.ReasonInvalidRequest,
		409: endpoint.ReasonConflict, 503: endpoint.ReasonUnavailable,
		500: endpoint.ReasonServerError,
	} {
		s.Equal(want, endpoint.ReasonFor(status), "status %d", status)
	}
}

func TestSpecPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(SpecPublicTestSuite))
}

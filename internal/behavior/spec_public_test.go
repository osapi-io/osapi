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

package behavior_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/osapi-io/osapi/internal/behavior/host"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
)

type SpecPublicTestSuite struct {
	suite.Suite
}

// sources parses the Go files of dir, without the tests.
func (s *SpecPublicTestSuite) sources(
	dir string,
) []*ast.File {
	entries, err := testFS.ReadDir(dir)
	s.Require().NoError(err)

	fset := token.NewFileSet()
	files := make([]*ast.File, 0, len(entries))
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, 0)
		s.Require().NoError(err)
		files = append(files, f)
	}

	return files
}

// resultName is the name of a function's one result type, with * kept.
func resultName(
	ft *ast.FuncType,
) string {
	if ft.Results == nil || len(ft.Results.List) != 1 {
		return ""
	}
	switch t := ft.Results.List[0].Type.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		if id, ok := t.X.(*ast.Ident); ok {
			return "*" + id.Name
		}
	}

	return ""
}

// dsl are the exported names of this package that an author writes in a
// definition: the Option values, and the functions that return an Option, a
// *Method, a Given or an Effect, with the ones a provider calls.
func (s *SpecPublicTestSuite) dsl() []string {
	names := []string{"Define", "Register", "Call", "Settings"}
	for _, f := range s.sources(".") {
		for _, d := range f.Decls {
			switch d := d.(type) {
			case *ast.GenDecl:
				for _, sp := range d.Specs {
					vs, ok := sp.(*ast.ValueSpec)
					if !ok {
						continue
					}
					if id, ok := vs.Type.(*ast.Ident); ok && id.Name == "Option" {
						for _, n := range vs.Names {
							names = append(names, n.Name)
						}
					}
				}
			case *ast.FuncDecl:
				if !d.Name.IsExported() {
					continue
				}
				switch resultName(d.Type) {
				case "Option", "Given", "Effect":
					if d.Recv == nil {
						names = append(names, d.Name.Name)
					}
				case "*Method":
					names = append(names, d.Name.Name)
				}
			}
		}
	}

	return names
}

func (s *SpecPublicTestSuite) TestSpecHasEveryPartOfTheDSL() {
	dsl := s.dsl()
	s.Contains(dsl, "Key", "the walk finds the refinements of a method")
	s.Contains(dsl, "KeepsState", "the walk finds the Option values")

	for _, name := range dsl {
		s.Run(name, func() {
			word := regexp.MustCompile(`(^|[^A-Za-z])` + name + `([^A-Za-z]|$)`)
			found := false
			// The chain layer owns the terms of calls between behaviors.
			for _, t := range append(behavior.Spec.Terms, behavior.ChainSpec.Terms...) {
				if word.MatchString(t.Name) {
					found = true
				}
			}
			s.True(found, "behavior.Spec and behavior.ChainSpec have no term for %s; add one", name)
		})
	}
}

// TestSpecHasEveryHostService: each field of host.Host, and Sources and
// Time, has a term of the kind host, so an author learns each service that a
// provider gets and how the tests fake it.
func (s *SpecPublicTestSuite) TestSpecHasEveryHostService() {
	t := reflect.TypeFor[host.Host]()
	names := make([]string, 0, 2+t.NumField())
	names = append(names, "Sources()", "Time()")
	for i := range t.NumField() {
		names = append(names, t.Field(i).Name)
	}
	for _, name := range names {
		s.Run(name, func() {
			word := regexp.MustCompile(
				`(^|[^A-Za-z])h\.` + regexp.QuoteMeta(name) + `([^A-Za-z(]|$)`,
			)
			found := false
			for _, term := range behavior.Spec.Terms {
				if term.Kind == behavior.KindHost && word.MatchString(term.Name) {
					found = true
				}
			}
			s.True(found, "behavior.Spec has no host term for h.%s; add one", name)
		})
	}
}

// TestSpecHasEveryError: each error that errors.go exports has a term of the
// kind error, which says what the API answers and when to return it.
func (s *SpecPublicTestSuite) TestSpecHasEveryError() {
	f, err := parser.ParseFile(token.NewFileSet(), "errors.go", nil, 0)
	s.Require().NoError(err)
	var names []string
	for _, d := range f.Decls {
		g, ok := d.(*ast.GenDecl)
		if !ok || g.Tok != token.VAR {
			continue
		}
		for _, spec := range g.Specs {
			for _, n := range spec.(*ast.ValueSpec).Names {
				if n.IsExported() {
					names = append(names, n.Name)
				}
			}
		}
	}
	s.Require().NotEmpty(names)
	for _, name := range names {
		found := false
		for _, t := range behavior.Spec.Of(behavior.KindError) {
			found = found || t.Name == "behavior."+name
		}
		s.True(found, "behavior.Spec has no error term for %s; add one", name)
	}
}

func (s *SpecPublicTestSuite) TestSpecHasEveryFieldTag() {
	get := regexp.MustCompile(`Tag\.Get\("([a-z]+)"\)`)
	var tags []string
	for _, dir := range []string{".", "../cli"} {
		for _, f := range s.sources(dir) {
			ast.Inspect(f, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok || len(call.Args) != 1 {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				lit, isLit := call.Args[0].(*ast.BasicLit)
				if ok && isLit && (sel.Sel.Name == "Get" || sel.Sel.Name == "Lookup") &&
					get.MatchString("Tag.Get("+lit.Value+")") {
					tag, _ := strconv.Unquote(lit.Value)
					tags = append(tags, tag)
				}

				return true
			})
		}
	}
	s.Contains(tags, "telemetry")

	for _, tag := range tags {
		found := false
		for _, t := range behavior.Spec.Of(behavior.KindTag) {
			if strings.HasPrefix(t.Name, tag+":") {
				found = true
			}
		}
		s.True(found, "behavior.Spec has no term for the %s tag; add one", tag)
	}
}

func (s *SpecPublicTestSuite) TestSpecKinds() {
	for _, t := range behavior.Spec.Terms {
		s.NotEmpty(behavior.Spec.Of(t.Kind), t.Name)
		s.NotEmpty(t.Does, t.Name)
		s.NotEmpty(t.When, t.Name)
	}
	for _, k := range behavior.Spec.Kinds {
		s.NotEmpty(behavior.Spec.Of(k.Kind), "kind %s has no terms", k.Kind)
	}
}

func (s *SpecPublicTestSuite) TestChecksHaveEveryRule() {
	var rules []string
	for _, f := range s.sources(".") {
		for _, d := range f.Decls {
			gd, ok := d.(*ast.GenDecl)
			if !ok || gd.Tok != token.CONST {
				continue
			}
			for _, sp := range gd.Specs {
				vs := sp.(*ast.ValueSpec)
				for i, n := range vs.Names {
					if strings.HasPrefix(n.Name, "Rule") && i < len(vs.Values) {
						if lit, ok := vs.Values[i].(*ast.BasicLit); ok {
							v, _ := strconv.Unquote(lit.Value)
							rules = append(rules, v)
						}
					}
				}
			}
		}
	}
	s.Contains(rules, behavior.RuleSpan)

	checked := map[string]bool{}
	for _, c := range append(behavior.Checks, behavior.ChainChecks...) {
		checked[c.Rule] = true
		s.NotEmpty(c.Requires, c.Rule)
		s.NotEmpty(c.Fix, c.Rule)
	}
	for _, r := range rules {
		s.True(checked[r], "the rule %s has no check in behavior.Checks or behavior.ChainChecks", r)
	}
	s.Len(checked, len(rules), "each check is of a Rule constant")
}

func (s *SpecPublicTestSuite) TestCurrentChainSpec() {
	behavior.Reset()
	defer behavior.Reset()
	behavior.Register(behavior.Define("files", "Files it keeps", behavior.Node, behavior.Internal,
		behavior.Read("get", get), behavior.Write("update", set)))
	behavior.Register(behavior.Define("kv", "Key values", behavior.Node, behavior.Read("get", get)))

	spec := behavior.CurrentChainSpec()

	s.Equal(behavior.KindBlock, spec.Kinds[len(spec.Kinds)-1].Kind, "building blocks come last")
	blocks := spec.Of(behavior.KindBlock)
	s.Require().Len(blocks, 1, "only an internal behavior is a building block")
	s.Equal("files", blocks[0].Name)
	s.Equal("Files it keeps. Methods: get, update.", blocks[0].Does)
	s.Contains(blocks[0].When, "osapi dsl behavior get --name files")
	s.Empty(behavior.ChainSpec.Of(behavior.KindBlock), "ChainSpec itself is not changed")
	for _, t := range behavior.ChainSpec.Terms {
		s.NotEmpty(t.Does, t.Name)
		s.NotEmpty(t.When, t.Name)
	}
	for _, k := range behavior.ChainSpec.Kinds {
		s.NotEmpty(behavior.ChainSpec.Of(k.Kind), "kind %s has no terms", k.Kind)
	}
}

func (s *SpecPublicTestSuite) TestCurrentSpec() {
	behavior.Reset()
	defer behavior.Reset()
	behavior.Register(behavior.Define("files", "Files it keeps", behavior.Node, behavior.Internal,
		behavior.Read("get", get), behavior.Write("update", set)))
	behavior.Register(behavior.Define("kv", "Key values", behavior.Node, behavior.Read("get", get)))

	spec := behavior.CurrentSpec()

	s.Empty(spec.Of(behavior.KindBlock), "the building blocks are in the chain spec")

	for _, t := range behavior.CurrentSpec("architecture", "fqdn").Of(behavior.KindTemplate) {
		if t.Name == ".Facts.<key>" {
			s.Contains(t.Does, "The keys: architecture, fqdn.")
		}
	}
	for _, t := range behavior.Spec.Of(behavior.KindTemplate) {
		s.NotContains(t.Does, "The keys:", "Spec itself is not changed")
	}
}

func (s *SpecPublicTestSuite) TestTestChecks() {
	tests := []struct {
		name string
		rule string
		want string
	}{
		{
			name: "when the rule has a check the message ends with its fix",
			rule: behavior.TestIdempotent,
			want: "running it again reports changed: true. Fix: Compare the host with the input " +
				"before the provider acts, and change nothing when they agree.",
		},
		{
			name: "when the rule has no check the fix is empty",
			rule: "nope",
			want: "running it again reports changed: true. Fix: ",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(
				tc.want,
				behavior.TestFailure(tc.rule, "running it again reports changed: true"),
			)
		})
	}

	for _, c := range behavior.TestChecks {
		s.NotEmpty(c.Requires, c.Rule)
		s.NotEmpty(c.Fix, c.Rule)
	}
}

func TestSpecPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(SpecPublicTestSuite))
}

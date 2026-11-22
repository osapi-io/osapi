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
	"go/token"
	"go/types"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"golang.org/x/tools/go/packages"
)

// definitions are the packages that define behaviors, endpoint groups and
// commands with the DSLs.
var definitions = []string{
	"github.com/osapi-io/osapi/internal/behaviors/...",
	"github.com/osapi-io/osapi/internal/controller/endpoints/...",
	"github.com/osapi-io/osapi/internal/cli/commands/...",
}

// typed are the named string types of the DSLs, which have a constant for
// each value. Go converts an untyped string literal to each, so a typo
// compiles; this test is the guard.
var typed = map[string]bool{
	"github.com/osapi-io/osapi/internal/behavior.PlatformName":      true,
	"github.com/osapi-io/osapi/internal/behavior.GroupName":         true,
	"github.com/osapi-io/osapi/internal/behavior.AreaName":          true,
	"github.com/osapi-io/osapi/internal/behavior.Verb":              true,
	"github.com/osapi-io/osapi/internal/behavior.Placement":         true,
	"github.com/osapi-io/osapi/internal/cli/command.Parent":         true,
	"github.com/osapi-io/osapi/internal/controller/endpoint.Reason": true,
	"github.com/osapi-io/osapi/pkg/sdk/client.Permission":           true,
	"github.com/osapi-io/osapi/pkg/sdk/client.Role":                 true,
	"github.com/osapi-io/osapi/pkg/sdk/client.ContentType":          true,
}

type LiteralsPublicTestSuite struct {
	suite.Suite
}

// TestNoLiteralForATypedValue fails on a string literal in a definition, not
// in a test and not generated, where a typed value of a DSL goes: an argument, a conversion,
// or a field of a composite literal. Use the constant, such as
// behavior.PlatformDebian, not "debian".
func (s *LiteralsPublicTestSuite) TestNoLiteralForATypedValue() {
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo,
	}, definitions...)
	s.Require().NoError(err)

	var found []string
	for _, p := range pkgs {
		s.Require().Empty(p.Errors)
		for _, f := range p.Syntax {
			// A generator writes a permission as a literal: the constant
			// of a new behavior's permission is written after its routes,
			// from the registry that the routes register in.
			if ast.IsGenerated(f) {
				continue
			}
			found = append(found, literals(p.Fset, p.TypesInfo, f)...)
		}
	}
	sort.Strings(found)
	s.Empty(found, "a string literal where a typed value of a DSL goes: use its constant")
}

// TestFindsALiteral is the guard of the guard: it gives literals a file
// with each kind of literal it must find.
func (s *LiteralsPublicTestSuite) TestFindsALiteral() {
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo,
		Dir: "testdata/literals",
	}, ".")
	s.Require().NoError(err)
	s.Require().Len(pkgs, 1)
	s.Require().Empty(pkgs[0].Errors)
	found := make([]string, 0, 4)
	for _, f := range pkgs[0].Syntax {
		found = append(found, literals(pkgs[0].Fset, pkgs[0].TypesInfo, f)...)
	}
	s.Equal([]string{
		"literals.go:31: \"debain\" as behavior.PlatformName",
		"literals.go:32: \"netwrok\" as behavior.GroupName",
		"literals.go:33: \"helth:read\" as client.Permission",
		"literals.go:36: \"raw\" as client.ContentType",
	}, found)
}

// literals are the string literals of f where a typed value goes.
func literals(
	fset *token.FileSet,
	info *types.Info,
	f *ast.File,
) []string {
	var out []string
	report := func(e ast.Expr, t types.Type) {
		lit, ok := ast.Unparen(e).(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return
		}
		named, ok := types.Unalias(t).(*types.Named)
		if !ok || named.Obj().Pkg() == nil {
			return
		}
		full := named.Obj().Pkg().Path() + "." + named.Obj().Name()
		if !typed[full] {
			return
		}
		pos := fset.Position(lit.Pos())
		file := pos.Filename[strings.LastIndex(pos.Filename, "/")+1:]
		out = append(out, fmt.Sprintf("%s:%d: %s as %s.%s",
			file, pos.Line, lit.Value, named.Obj().Pkg().Name(), named.Obj().Name()))
	}
	ast.Inspect(f, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.CallExpr:
			if tv, ok := info.Types[n.Fun]; ok && tv.IsType() {
				if len(n.Args) == 1 {
					report(n.Args[0], tv.Type)
				}

				return true
			}
			sig, ok := info.TypeOf(n.Fun).(*types.Signature)
			if !ok {
				return true
			}
			for i, a := range n.Args {
				report(a, paramType(sig, i))
			}
		case *ast.CompositeLit:
			for _, e := range n.Elts {
				if kv, ok := e.(*ast.KeyValueExpr); ok {
					report(kv.Value, info.TypeOf(kv.Key))
				} else {
					report(e, elemType(info.TypeOf(n)))
				}
			}
		}

		return true
	})

	return out
}

// paramType is the type of argument i of a call of sig, with the element
// type of a variadic parameter for each argument it takes.
func paramType(
	sig *types.Signature,
	i int,
) types.Type {
	n := sig.Params().Len()
	if n == 0 {
		return nil
	}
	if sig.Variadic() && i >= n-1 {
		return sig.Params().At(n - 1).Type().(*types.Slice).Elem()
	}
	if i >= n {
		return nil
	}

	return sig.Params().At(i).Type()
}

// elemType is the type of an element of a slice or an array literal of
// type t, or nil.
func elemType(
	t types.Type,
) types.Type {
	if t == nil {
		return nil
	}
	switch u := t.Underlying().(type) {
	case *types.Slice:
		return u.Elem()
	case *types.Array:
		return u.Elem()
	}

	return nil
}

func TestLiteralsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(LiteralsPublicTestSuite))
}

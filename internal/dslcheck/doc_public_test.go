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
	"bufio"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"golang.org/x/tools/go/packages"

	"github.com/osapi-io/osapi/internal/audit"
	"github.com/osapi-io/osapi/internal/behavior"
	_ "github.com/osapi-io/osapi/internal/behaviors/all"
	"github.com/osapi-io/osapi/internal/cli/command"
	_ "github.com/osapi-io/osapi/internal/cli/commands/all"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	_ "github.com/osapi-io/osapi/internal/controller/endpoints/all"
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
	"github.com/osapi-io/osapi/internal/behavior.PlatformName": true,
	"github.com/osapi-io/osapi/internal/area.Name":             true,
	"github.com/osapi-io/osapi/pkg/sdk/types.Verb":             true,
	"github.com/osapi-io/osapi/pkg/sdk/types.MethodKind":       true,
	"github.com/osapi-io/osapi/internal/behavior.Placement":    true,
	"github.com/osapi-io/osapi/internal/cli/command.Parent":    true,
	"github.com/osapi-io/osapi/pkg/sdk/types.Reason":           true,
	"github.com/osapi-io/osapi/pkg/sdk/types.Permission":       true,
	"github.com/osapi-io/osapi/pkg/sdk/types.Role":             true,
	"github.com/osapi-io/osapi/pkg/sdk/types.ContentType":      true,
}

type DocPublicTestSuite struct {
	suite.Suite
}

// TestNoLiteralForATypedValue fails on a string literal in a definition, not
// in a test and not generated, where a typed value of a DSL goes: an argument, a conversion,
// or a field of a composite literal. Use the constant, such as
// behavior.PlatformDebian, not "debian".
func (s *DocPublicTestSuite) TestNoLiteralForATypedValue() {
	pkgs, err := loadChecked("", definitions...)
	s.Require().NoError(err)

	var found []string
	for _, p := range pkgs {
		for _, f := range p.Files {
			// A generator writes a permission as a literal: the constant
			// of a new behavior's permission is written after its routes,
			// from the registry that the routes register in.
			if ast.IsGenerated(f) {
				continue
			}
			found = append(found, literals(p.Fset, p.Info, f)...)
		}
	}
	sort.Strings(found)
	s.Empty(found)
}

// TestFindsALiteral is the guard of the guard: it gives literals a file
// with each kind of literal it must find.
func (s *DocPublicTestSuite) TestFindsALiteral() {
	pkgs, err := loadChecked("testdata/literals", ".")
	s.Require().NoError(err)
	s.Require().Len(pkgs, 1)
	found := make([]string, 0, 4)
	for _, f := range pkgs[0].Files {
		found = append(found, literals(pkgs[0].Fset, pkgs[0].Info, f)...)
	}
	s.Equal([]string{
		"literals.go:32: \"debain\" as behavior.PlatformName",
		"literals.go:33: \"hardwar\" as area.Name",
		"literals.go:34: \"helth:read\" as types.Permission",
		"literals.go:37: \"raw\" as types.ContentType",
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

func TestDocPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(DocPublicTestSuite))
}

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

// TestSecretsAreTyped reads the input of each method of the three DSLs, at
// each depth, and fails on a string field whose JSON name is on the audit
// denylist, unless testdata/secrets.txt gives a reason. Such a field holds
// a secret, and a secret is a secret.Value: the type keeps it out of the
// spans and the chain records, makes it write-only in the spec, and keeps
// it out of a flag. The audit's name match alone does none of that.
func (s *DocPublicTestSuite) TestSecretsAreTyped() {
	allowed := readAllowed(s.T())
	denied := audit.SensitiveFields()

	var got []string
	add := func(dsl string, group string, method string, in reflect.Type) {
		for _, path := range stringFields(in, "", denied, map[reflect.Type]bool{}) {
			got = append(got, dsl+" "+group+" "+method+" "+path)
		}
	}
	for _, b := range behavior.All() {
		for _, m := range b.Methods {
			add("behavior", b.Name, m.Name, m.In)
		}
	}
	for _, g := range endpoint.All() {
		// A behavior's group is the behavior's request, read above.
		if behavior.Lookup(g.Name) != nil {
			continue
		}
		for _, m := range g.Methods {
			add("endpoint", g.Name, m.Name, m.In)
		}
	}
	for _, g := range command.All() {
		for _, m := range g.Methods {
			add("command", strings.Join(g.CommandPath(), "-"), m.MethodName(), m.In)
		}
	}
	sort.Strings(got)

	var untyped []string
	for _, e := range got {
		if _, ok := allowed[e]; !ok {
			untyped = append(untyped, e)
		}
	}
	// A string input field whose name is on the audit denylist holds a
	// secret: give it the type secret.Value, or add it to
	// testdata/secrets.txt with the reason that it is not a secret.
	s.Empty(untyped)

	var stale []string
	for e := range allowed {
		if !slices.Contains(got, e) {
			stale = append(stale, e)
		}
	}
	sort.Strings(stale)
	s.Empty(stale)
}

// stringFields are the paths of the string fields of t, at each depth,
// whose JSON name, in lower case, is in names.
func stringFields(
	t reflect.Type,
	prefix string,
	names map[string]bool,
	seen map[reflect.Type]bool,
) []string {
	for t.Kind() == reflect.Pointer || t.Kind() == reflect.Slice ||
		t.Kind() == reflect.Array || t.Kind() == reflect.Map {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct || seen[t] {
		return nil
	}
	seen[t] = true
	var out []string
	for f := range t.Fields() {
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if name == "" {
			name = f.Name
		}
		if !f.IsExported() || name == "-" {
			continue
		}
		if f.Type == reflect.TypeFor[string]() && names[strings.ToLower(name)] {
			out = append(out, prefix+name)
		}
		out = append(out, stringFields(f.Type, prefix+name+".", names, seen)...)
	}

	return out
}

// readAllowed reads testdata/secrets.txt: each line is an input field, as
// <dsl> <group> <method> <field>, a tab, and why it is not a secret.
func readAllowed(
	t *testing.T,
) map[string]string {
	f, err := os.Open("testdata/secrets.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	out := map[string]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		entry, reason, ok := strings.Cut(line, "\t")
		if !ok || strings.TrimSpace(reason) == "" {
			t.Fatalf("testdata/secrets.txt: %q has no reason", line)
		}
		out[entry] = reason
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}

	return out
}

// dsls are the packages whose exported API the tests read.
var dsls = []string{
	"github.com/osapi-io/osapi/internal/behavior",
	"github.com/osapi-io/osapi/internal/controller/endpoint",
	"github.com/osapi-io/osapi/internal/cli/command",
}

// categories are why a string of a DSL stays a string.
var categories = []string{
	"identity", "free text", "program", "path pattern", "fixture", "boundary", "checked",
}

// TestStringsAreReviewed lists each exported function or method parameter,
// and each exported struct field, of the DSLs whose type is string, a slice
// of string or a map with string keys, and compares the list with
// testdata/strings.txt. A string that picks from a fixed set is a typed
// value with constants; each other string is in the file with its reason.
func (s *DocPublicTestSuite) TestStringsAreReviewed() {
	reviewed := readReviewed(s.T())
	got := exportedStrings(s.T())

	var unreviewed []string
	for _, e := range got {
		if _, ok := reviewed[e]; !ok {
			unreviewed = append(unreviewed, e)
		}
	}
	// A string in the API of a DSL that nobody reviewed: make it a type with
	// constants and a validate rule, or add it to testdata/strings.txt with
	// its category and reason.
	s.Empty(unreviewed)

	var stale []string
	for e := range reviewed {
		if !slices.Contains(got, e) {
			stale = append(stale, e)
		}
	}
	sort.Strings(stale)
	s.Empty(stale)
}

// readReviewed is testdata/strings.txt: each entry, with its category and
// reason.
func readReviewed(
	t *testing.T,
) map[string]string {
	t.Helper()
	f, err := os.Open("testdata/strings.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	out := map[string]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) != 3 || parts[2] == "" {
			t.Errorf("%q: want entry, category and reason, separated by tabs", line)

			continue
		}
		if !slices.Contains(categories, parts[1]) {
			t.Errorf("%q: the category %q is not one of %v", line, parts[1], categories)
		}
		out[parts[0]] = parts[1]
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}

	return out
}

// exportedStrings are the strings of the API of the DSLs, as
// testdata/strings.txt names them: pkg.Func(param type), pkg.Type.Method(param
// type) or pkg.Type.Field type.
func exportedStrings(
	t *testing.T,
) []string {
	t.Helper()
	pkgs, err := packages.Load(
		&packages.Config{Mode: packages.NeedTypes | packages.NeedName},
		dsls...)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	params := func(prefix string, sig *types.Signature) {
		for i := range sig.Params().Len() {
			if v := sig.Params().At(i); isString(v.Type()) {
				out = append(out, fmt.Sprintf("%s(%s %s)", prefix, v.Name(), v.Type()))
			}
		}
	}
	for _, p := range pkgs {
		if len(p.Errors) > 0 {
			t.Fatal(p.Errors)
		}
		scope := p.Types.Scope()
		for _, name := range scope.Names() {
			switch o := scope.Lookup(name).(type) {
			case *types.Func:
				if o.Exported() {
					params(p.Name+"."+name, o.Type().(*types.Signature))
				}
			case *types.TypeName:
				named, ok := o.Type().(*types.Named)
				if !ok || !o.Exported() {
					continue
				}
				if st, ok := named.Underlying().(*types.Struct); ok {
					for i := range st.NumFields() {
						if f := st.Field(i); f.Exported() && isString(f.Type()) {
							out = append(
								out,
								fmt.Sprintf("%s.%s.%s %s", p.Name, name, f.Name(), f.Type()),
							)
						}
					}
				}
				for i := range named.NumMethods() {
					if m := named.Method(i); m.Exported() {
						params(p.Name+"."+name+"."+m.Name(), m.Type().(*types.Signature))
					}
				}
			}
		}
	}
	sort.Strings(out)

	return out
}

// isString reports whether t is string, a slice of string, or a map with
// string keys. A named string type is a typed value, and is not one.
func isString(
	t types.Type,
) bool {
	plain := func(t types.Type) bool {
		b, ok := t.(*types.Basic)

		return ok && b.Kind() == types.String
	}
	switch u := t.(type) {
	case *types.Slice:
		return plain(u.Elem())
	case *types.Map:
		return plain(u.Key())
	default:
		return plain(t)
	}
}

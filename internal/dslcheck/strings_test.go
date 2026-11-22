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
	"go/types"
	"os"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"golang.org/x/tools/go/packages"
)

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

type StringsPublicTestSuite struct {
	suite.Suite
}

// TestStringsAreReviewed lists each exported function or method parameter,
// and each exported struct field, of the DSLs whose type is string, a slice
// of string or a map with string keys, and compares the list with
// testdata/strings.txt. A string that picks from a fixed set is a typed
// value with constants; each other string is in the file with its reason.
func (s *StringsPublicTestSuite) TestStringsAreReviewed() {
	reviewed := readReviewed(s.T())
	got := exportedStrings(s.T())

	var unreviewed []string
	for _, e := range got {
		if _, ok := reviewed[e]; !ok {
			unreviewed = append(unreviewed, e)
		}
	}
	s.Empty(unreviewed,
		"a string in the API of a DSL that nobody reviewed: make it a type with constants "+
			"and a validate rule, or add it to internal/dslcheck/testdata/strings.txt with its "+
			"category and reason")

	var stale []string
	for e := range reviewed {
		if !slices.Contains(got, e) {
			stale = append(stale, e)
		}
	}
	sort.Strings(stale)
	s.Empty(stale, "entries of testdata/strings.txt that the API no longer has: remove them")
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

func TestStringsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(StringsPublicTestSuite))
}

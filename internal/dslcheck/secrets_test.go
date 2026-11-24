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
	"os"
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/audit"
	"github.com/osapi-io/osapi/internal/behavior"
	_ "github.com/osapi-io/osapi/internal/behaviors/all"
	"github.com/osapi-io/osapi/internal/cli/command"
	_ "github.com/osapi-io/osapi/internal/cli/commands/all"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	_ "github.com/osapi-io/osapi/internal/controller/endpoints/all"
)

type SecretsPublicTestSuite struct {
	suite.Suite
}

// TestSecretsAreTyped reads the input of each method of the three DSLs, at
// each depth, and fails on a string field whose JSON name is on the audit
// denylist, unless testdata/secrets.txt gives a reason. Such a field holds
// a secret, and a secret is a secret.Value: the type keeps it out of the
// spans and the chain records, makes it write-only in the spec, and keeps
// it out of a flag. The audit's name match alone does none of that.
func (s *SecretsPublicTestSuite) TestSecretsAreTyped() {
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
	s.Empty(untyped,
		"a string input field whose name is on the audit denylist holds a secret: give it "+
			"the type secret.Value, or add it to internal/dslcheck/testdata/secrets.txt "+
			"with the reason that it is not a secret")

	var stale []string
	for e := range allowed {
		if !slices.Contains(got, e) {
			stale = append(stale, e)
		}
	}
	sort.Strings(stale)
	s.Empty(stale, "entries of testdata/secrets.txt that no input has: remove them")
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

func TestSecretsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(SecretsPublicTestSuite))
}

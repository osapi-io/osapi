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

package config_test

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/config"
)

type ChecksPublicTestSuite struct {
	suite.Suite
}

func (s *ChecksPublicTestSuite) TestChecksHaveEveryRule() {
	entries, err := os.ReadDir(".")
	s.Require().NoError(err)

	var rules []string
	fset := token.NewFileSet()
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(".", e.Name()), nil, 0)
		s.Require().NoError(err)
		for _, d := range f.Decls {
			gd, ok := d.(*ast.GenDecl)
			if !ok || gd.Tok != token.CONST {
				continue
			}
			for _, sp := range gd.Specs {
				vs := sp.(*ast.ValueSpec)
				for i, n := range vs.Names {
					if !strings.HasPrefix(n.Name, "Rule") || i >= len(vs.Values) {
						continue
					}
					if lit, ok := vs.Values[i].(*ast.BasicLit); ok {
						v, _ := strconv.Unquote(lit.Value)
						rules = append(rules, v)
					}
				}
			}
		}
	}
	s.Contains(rules, config.RuleBucket)

	checked := make([]string, 0, len(config.Checks))
	for _, c := range config.Checks {
		checked = append(checked, c.Rule)
		s.NotEmpty(c.Requires, c.Rule)
		s.NotEmpty(c.Fix, c.Rule)
	}
	s.ElementsMatch(rules, checked, "each Rule constant has one check in config.Checks")
}

func (s *ChecksPublicTestSuite) TestRuleOf() {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "when the error has a rule returns it",
			err:  &config.Error{Rule: config.RuleFile, Err: errors.New("gone")},
			want: config.RuleFile,
		},
		{name: "when the error has no rule returns nothing", err: errors.New("boom"), want: ""},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, config.RuleOf(tc.err))
		})
	}

	inner := errors.New("gone")
	e := &config.Error{Rule: config.RuleFile, Err: inner}
	s.Equal("gone", e.Error())
	s.ErrorIs(e, inner)
}

type keyed struct {
	Outer struct {
		Port int `mapstructure:"port_number"`
	} `mapstructure:"outer"`
	List  []struct{ Name string } `mapstructure:"list"`
	Plain string
}

func (s *ChecksPublicTestSuite) TestKeyOf() {
	t := reflect.TypeFor[keyed]()

	tests := []struct {
		name      string
		namespace string
		prefix    string
		want      string
	}{
		{
			name:      "when the fields have mapstructure names uses them",
			namespace: "keyed.Outer.Port",
			want:      "outer.port_number",
		},
		{
			name:      "when a field is in a list keeps its index",
			namespace: "keyed.List[0].Name",
			want:      "list[0].Name",
		},
		{
			name:      "when a field has no mapstructure name keeps its Go name",
			namespace: "keyed.Plain",
			want:      "Plain",
		},
		{
			name:      "when the type has no such field keeps the path",
			namespace: "keyed.Gone.X",
			want:      "Gone.X",
		},
		{
			name:      "when a prefix is given puts it first",
			namespace: "keyed.Plain",
			prefix:    "behaviors.kv",
			want:      "behaviors.kv.Plain",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, config.KeyOf(t, tc.namespace, tc.prefix))
		})
	}
}

func (s *ChecksPublicTestSuite) TestValueProblems() {
	err := errors.New("not a validation error")
	s.Equal(
		[]config.Problem{{Rule: config.RuleValue, Message: "not a validation error"}},
		config.ValueProblems(err, config.RuleValue, reflect.TypeFor[keyed](), ""),
	)
}

func (s *ChecksPublicTestSuite) TestProblemString() {
	s.Equal("a.b breaks it", config.Problem{Key: "a.b", Message: "breaks it"}.String())
	s.Equal("no key", config.Problem{Message: "no key"}.String())
}

func TestChecksPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ChecksPublicTestSuite))
}

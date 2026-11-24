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
package rule_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/validation/rule"
)

// got records what a validate tag set.
type got struct {
	enum    []string
	pattern string
	format  string
	bounds  map[string]float64
	items   *got
}

func (g *got) Enum(values []string)      { g.enum = values }
func (g *got) Pattern(re string)         { g.pattern = re }
func (g *got) Format(format string)      { g.format = format }
func (g *got) Bound(r string, n float64) { g.bounds[r] = n }
func newGot() *got                       { return &got{bounds: map[string]float64{}} }

func (g *got) Items() rule.Target {
	if g.items == nil {
		g.items = newGot()
	}

	return g.items
}

// shade is a closed set, as an enum of pkg/sdk/types is.
type shade string

func (shade) Values() []shade { return []shade{"light", "dark"} }

func (c shade) Valid() bool { return c == "light" || c == "dark" }

// listed has Values but no Valid, so it is not a closed set.
type listed string

func (listed) Values() []listed { return []listed{"a"} }

// strange has Values and Valid, but its Values are not of its type.
type strange string

func (strange) Values() []string { return []string{"a"} }

func (strange) Valid() bool { return true }

type RulePublicTestSuite struct {
	suite.Suite
}

func (s *RulePublicTestSuite) SetupSuite() {
	rule.Register("test_pattern", rule.Pattern("^x+$"))
	rule.Register("test_format", rule.Format("hostname"))
	rule.Register("test_none", rule.None)
}

func (s *RulePublicTestSuite) TestApply() {
	tests := []struct {
		name         string
		tag          string
		typ          reflect.Type
		validateFunc func(*got, bool, error)
	}{
		{
			name: "when the tag is empty, nothing is set and the field is optional",
			tag:  "",
			validateFunc: func(g *got, required bool, err error) {
				s.NoError(err)
				s.False(required)
				s.Empty(g.bounds)
			},
		},
		{
			name: "when the tag has the standard rules, each is set",
			tag:  "required,min=1,max=255,len=3,oneof=a b,startswith=/etc,omitempty,dive",
			validateFunc: func(g *got, required bool, err error) {
				s.Require().NoError(err)
				s.True(required)
				s.Equal(map[string]float64{"min": 1, "max": 255, "len": 3}, g.bounds)
				s.Equal([]string{"a", "b"}, g.enum)
				s.Equal(`^/etc`, g.pattern)
			},
		},
		{
			name: "when the tag has dive, the rules after it are each item's",
			tag:  "required,max=10,dive,max=255,oneof=a b",
			validateFunc: func(g *got, required bool, err error) {
				s.Require().NoError(err)
				s.True(required)
				s.Equal(map[string]float64{"max": 10}, g.bounds, "the array has at most 10 items")
				s.Empty(g.enum)
				s.Require().NotNil(g.items)
				s.Equal(map[string]float64{"max": 255}, g.items.bounds, "each item is at most 255")
				s.Equal([]string{"a", "b"}, g.items.enum)
			},
		},
		{
			name: "when the tag has a standard format, it is the format",
			tag:  "url",
			validateFunc: func(g *got, _ bool, err error) {
				s.NoError(err)
				s.Equal("uri", g.format)
			},
		},
		{
			name: "when the tag has registered rules, their schemas are set",
			tag:  "test_pattern,test_format,test_none",
			validateFunc: func(g *got, _ bool, err error) {
				s.NoError(err)
				s.Equal("^x+$", g.pattern)
				s.Equal("hostname", g.format)
			},
		},
		{
			name: "when a rule has no schema, it is an error",
			tag:  "required,nosuchrule",
			validateFunc: func(_ *got, _ bool, err error) {
				s.EqualError(err,
					`validate tag "nosuchrule" has no schema; register one with rule.Register`)
			},
		},
		{
			name: "when the tag has startsnotwith, the first character is any other",
			tag:  "startsnotwith=-",
			validateFunc: func(g *got, _ bool, err error) {
				s.NoError(err)
				s.Equal(`^[^-]`, g.pattern)
			},
		},
		{
			name: "when startsnotwith has more than one character, it is an error",
			tag:  "startsnotwith=--",
			validateFunc: func(_ *got, _ bool, err error) {
				s.EqualError(err, "validate startsnotwith=--: the spec shows one character only")
			},
		},
		{
			name: "when a bound is not a number, it is an error",
			tag:  "max=lots",
			validateFunc: func(_ *got, _ bool, err error) {
				s.EqualError(err, "validate max=lots: not a number")
			},
		},
		{
			name: "when the type is a closed set, enum gives its values",
			tag:  "required,enum",
			typ:  reflect.TypeFor[*shade](),
			validateFunc: func(g *got, required bool, err error) {
				s.Require().NoError(err)
				s.True(required)
				s.Equal([]string{"light", "dark"}, g.enum)
			},
		},
		{
			name: "when a list holds a closed set, enum after dive gives each item's values",
			tag:  "omitempty,dive,enum",
			typ:  reflect.TypeFor[*[]shade](),
			validateFunc: func(g *got, _ bool, err error) {
				s.Require().NoError(err)
				s.Empty(g.enum)
				s.Require().NotNil(g.items)
				s.Equal([]string{"light", "dark"}, g.items.enum)
			},
		},
		{
			name: "when the type is not a closed set, enum is an error",
			tag:  "enum",
			validateFunc: func(_ *got, _ bool, err error) {
				s.EqualError(err, "validate enum on string: the type has no Values and Valid")
			},
		},
		{
			name: "when dive is on a type with no items, the rules after it keep the type",
			tag:  "dive,enum",
			typ:  reflect.TypeFor[shade](),
			validateFunc: func(g *got, _ bool, err error) {
				s.Require().NoError(err)
				s.Equal([]string{"light", "dark"}, g.items.enum)
			},
		},
		{
			name: "when the type is a closed set, oneof is an error",
			tag:  "oneof=light dark",
			typ:  reflect.TypeFor[shade](),
			validateFunc: func(_ *got, _ bool, err error) {
				s.EqualError(err,
					"validate oneof on rule_test.shade: the type has Values; use validate enum")
			},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			g := newGot()
			typ := tc.typ
			if typ == nil {
				typ = reflect.TypeFor[string]()
			}
			required, err := rule.Apply(tc.tag, typ, g)
			tc.validateFunc(g, required, err)
		})
	}
}

func (s *RulePublicTestSuite) TestHas() {
	s.True(rule.Has("omitempty,min=1", "omitempty"))
	s.False(rule.Has("omitempty,min=1", "min"))
}

type retry struct {
	Target string `json:"target" validate:"omitempty"`
	Reason string `json:"reason"`
}

type input struct {
	ID      string `json:"id"             in:"path" validate:"uuid"`
	Note    string `json:"note,omitempty"`
	Plain   string
	Skipped string `json:"-"`
	hidden  string //nolint:unused
	Body    retry  `                      in:"body"`
}

func (s *RulePublicTestSuite) TestUntagged() {
	tests := []struct {
		name string
		t    reflect.Type
		want []string
	}{
		{
			name: "when fields have no validate tag, each is named, with the body's",
			t:    reflect.TypeFor[*input](),
			want: []string{"note", "Plain", "Body", "Body.reason"},
		},
		{
			name: "when the input is not a struct, there is nothing to tag",
			t:    reflect.TypeFor[string](),
			want: nil,
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, rule.Untagged(tc.t))
		})
	}
}

type needsBody struct {
	Target string `json:"target" validate:"omitempty,min=1,max=3"`
}

type needsInput struct {
	Key    string    `json:"key"   in:"path" validate:"required,min=1,max=255"`
	Name   string    `json:"name"            validate:"required,min=1,uuid"`
	Note   string    `json:"note"            validate:"min=1"`
	Count  int       `json:"count"           validate:"omitempty,min=1"`
	Tags   []string  `json:"tags"            validate:"required,min=1,dive"`
	Bad    string    `json:"bad"             validate:"min=x"`
	Skip   string    `json:"-"               validate:"max=1"`
	Free   string    `json:"free"`
	hidden string    //nolint:unused
	Body   needsBody `             in:"body" validate:"required"`
}

// TestNeeds names each rule that needs a bad input, and leaves out what a
// left-out field tests and a min whose only bad value is empty.
func (s *RulePublicTestSuite) TestNeeds() {
	got := rule.Needs(reflect.TypeFor[*needsInput](), false)
	keys := make([]string, 0, len(got))
	for _, n := range got {
		keys = append(keys, n.Key())
	}

	s.Equal([]string{
		"Key:max", "Name:uuid", "Note:min", "Count:min", "Tags:min", "Bad:min", "Body.Target:max",
	}, keys)
	s.Equal(rule.Need{Field: "Key", Rule: "max=255"}, got[0])
	s.Nil(rule.Needs(reflect.TypeFor[string](), false))

	required := []string{}
	for _, n := range rule.Needs(reflect.TypeFor[needsInput](), true) {
		if strings.HasSuffix(n.Key(), ":required") {
			required = append(required, n.Key())
		}
	}
	s.Equal([]string{"Name:required", "Tags:required"}, required,
		"with required, each required field but a path field or a struct value needs one")
}

func (s *RulePublicTestSuite) TestDeclare() {
	var seen []string
	rule.Declare(rule.Custom{Tag: "rule_test_before", Schema: rule.Pattern("^b$")})
	rule.OnDeclare(func(c rule.Custom) { seen = append(seen, c.Tag) })
	s.Contains(seen, "rule_test_before", "a rule declared before OnDeclare is replayed")

	rule.Declare(rule.Custom{Tag: "rule_test_after", Schema: rule.None})
	s.Contains(seen, "rule_test_after", "a rule declared after OnDeclare is passed on")

	g := newGot()
	_, err := rule.Apply("rule_test_before", reflect.TypeFor[string](), g)
	s.NoError(err)
	s.Equal("^b$", g.pattern, "Declare registers the rule's schema")
}

func (s *RulePublicTestSuite) TestValues() {
	tests := []struct {
		name string
		typ  reflect.Type
		want []string
		ok   bool
	}{
		{
			name: "when the type is a closed set, they are its values",
			typ:  reflect.TypeFor[*shade](),
			want: []string{"light", "dark"},
			ok:   true,
		},
		{name: "when the type is a string, there are none", typ: reflect.TypeFor[string]()},
		{name: "when the type has no Valid, there are none", typ: reflect.TypeFor[listed]()},
		{
			name: "when Values is not of the type, there are none",
			typ:  reflect.TypeFor[strange](),
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			got, ok := rule.Values(tc.typ)
			s.Equal(tc.ok, ok)
			s.Equal(tc.want, got)
		})
	}
}

func (s *RulePublicTestSuite) TestOptions() {
	tests := []struct {
		name string
		tag  string
		typ  reflect.Type
		want []string
	}{
		{
			name: "when the field has enum, they are the values of its type",
			tag:  "omitempty,enum",
			typ:  reflect.TypeFor[shade](),
			want: []string{"light", "dark"},
		},
		{
			name: "when a list has oneof after dive, they are each item's values",
			tag:  "omitempty,max=3,dive,oneof=a b",
			typ:  reflect.TypeFor[[]string](),
			want: []string{"a", "b"},
		},
		{
			name: "when the field has neither, there are none",
			tag:  "required,min=1,startswith=/,email",
			typ:  reflect.TypeFor[string](),
		},
		{
			name: "when Apply refuses the tag, there are none",
			tag:  "enum",
			typ:  reflect.TypeFor[string](),
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Equal(tc.want, rule.Options(tc.tag, tc.typ))
		})
	}
}

func TestRulePublicTestSuite(t *testing.T) {
	suite.Run(t, new(RulePublicTestSuite))
}

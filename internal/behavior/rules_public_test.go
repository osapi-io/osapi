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
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/behavior"
)

// RulesPublicTestSuite drives the rules that the DSL audit added (osapi
// #973-#983) through ValidateBehavior, one input or output each.
type RulesPublicTestSuite struct {
	suite.Suite
}

type optionalKeyIn struct {
	Name string `json:"name" validate:"omitempty" doc:"test"`
}

type badInput struct {
	NoJSON   string `validate:"omitempty" doc:"test"`
	Hostname string `validate:"omitempty" doc:"test" json:"host"`
	Target   string `validate:"omitempty" doc:"test" json:"target"`
	NoDoc    string `validate:"omitempty"            json:"nodoc"`
	Def      string `validate:"omitempty" doc:"test" json:"def"    default:"x"`
	Arr      [2]int `validate:"omitempty" doc:"test" json:"arr"`
	Hidden   string `                                json:"-"`
	_        string
}

type badOutput struct {
	NoJSON string
	Arr    *[2]string        `json:"arr"`
	Seen   map[string][]bool `json:"seen"`
	Size   int               `json:"size" render:"furlongs"`
	_      string
}

type okOutput struct {
	Key string `json:"key" doc:"The key"`
}

func readOptionalKey(context.Context, fakeProvider, optionalKeyIn) (okOutput, error) {
	return okOutput{}, nil
}

func readBadInput(context.Context, fakeProvider, badInput) (okOutput, error) {
	return okOutput{}, nil
}

func readBadOutput(context.Context, fakeProvider, optionalKeyIn) (badOutput, error) {
	return badOutput{}, nil
}

func (s *RulesPublicTestSuite) problems(
	b *behavior.Behavior,
) map[string][]string {
	out := map[string][]string{}
	for _, p := range behavior.ValidateBehavior(b) {
		out[p.Rule] = append(out[p.Rule], p.Message)
	}

	return out
}

func (s *RulesPublicTestSuite) TestKeyIsRequired() {
	got := s.problems(behavior.Define("note", "x", behavior.Node, onDebian,
		behavior.Read("get", readOptionalKey).Key("name").Describe("One")))
	s.Equal([]string{`key "name" is optional, but a key is a path segment, which is never empty`},
		got[behavior.RuleKey])
}

func (s *RulesPublicTestSuite) TestInput() {
	got := s.problems(behavior.Define("note", "x", behavior.Node, onDebian,
		behavior.API, behavior.Area("config"), behavior.CLI,
		behavior.Read("list", readBadInput).Describe("List")))
	s.Equal([]string{
		"input field NoJSON has no json name",
		"input field host is named like the target, which every request has",
		"input field target is named like the target, which every request has",
		"input field nodoc has no doc",
		"input field def has a default, which the server does not apply",
		"input field arr is an array; use a slice",
	}, got[behavior.RuleInput])
}

func (s *RulesPublicTestSuite) TestOutput() {
	got := s.problems(behavior.Define("note", "x", behavior.Node, onDebian,
		behavior.Read("get", readBadOutput).Describe("Get")))
	s.Contains(got[behavior.RuleOutputShape], "output field NoJSON has no json name")
	s.Contains(got[behavior.RuleOutputShape], "output field Arr is an array; use a slice")
	s.Len(got[behavior.RuleOutputShape], 3, "and the render tag the output does not know")
}

func (s *RulesPublicTestSuite) TestNames() {
	got := s.problems(behavior.Define("ip-forward", "x", behavior.Node, onDebian,
		behavior.Read("list", readList).Describe("List")))
	s.Equal([]string{
		`name "ip-forward" must be lowercase letters and digits, starting with a letter`,
	}, got[behavior.RuleName])
}

// usageA and usageB are two types named Usage, as two behaviors could
// each declare: in one spec they would be one schema.
func usageA() any {
	type Usage struct {
		Total int `json:"total" doc:"test"`
	}
	type Out struct {
		Usage Usage                  `json:"usage" doc:"test"`
		Same  *[]map[string]okOutput `json:"same"  doc:"test"`
		Again okOutput               `json:"again" doc:"test"`
	}

	return Out{}
}

func usageB() any {
	type Usage struct {
		Free int `json:"free" doc:"test"`
	}
	type Out struct {
		Usage Usage `json:"usage" doc:"test"`
	}

	return Out{}
}

func (s *RulesPublicTestSuite) TestSchemaNames() {
	define := func(name string, _ any) *behavior.Behavior {
		return behavior.Define(
			name,
			"x",
			behavior.Node,
			onDebian,
			behavior.Read("get", func(context.Context, fakeProvider, optionalKeyIn) (okOutput, error) {
				return okOutput{}, nil
			}).
				Describe("Get"),
		)
	}
	a, b := define("memory", nil), define("swap", nil)
	a.Methods[0].Out = reflect.TypeOf(usageA())
	b.Methods[0].Out = reflect.TypeOf(usageB())
	same := define("again", nil)
	same.Methods[0].Out = reflect.TypeOf(usageA())

	got := behavior.ExportSchemaNameProblems([]*behavior.Behavior{a, same, b})
	s.Require().Len(got, 1, "the same type in two behaviors is one schema, which is right")
	s.Equal("swap", got[0].Behavior)
	s.Equal(behavior.RuleSchema, got[0].Rule)
	s.Equal("Usage is also a type of memory; in the spec the two are one schema", got[0].Message)
}

func (s *RulesPublicTestSuite) TestDiveOnAFieldWithNoItems() {
	type in struct {
		Name string `json:"name" validate:"dive,max=3" doc:"test"`
	}
	schema, err := behavior.SchemaFor(reflect.TypeFor[in]())
	s.Require().NoError(err)
	s.Nil(schema.Properties["name"].MaxLength, "the rules after dive constrain no item")
}

func TestRulesPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(RulesPublicTestSuite))
}

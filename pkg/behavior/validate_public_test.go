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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/pkg/behavior"
)

type ValidatePublicTestSuite struct {
	suite.Suite
}

func (s *ValidatePublicTestSuite) TearDownTest() {
	behavior.Reset()
}

type noChanged struct {
	Key string `json:"key"`
}

type stringChanged struct {
	Changed string `json:"changed"`
}

type reserved struct {
	Hostname string `json:"hostname"`
}

func readList(
	_ context.Context,
	_ fakeProvider,
	_ getIn,
) ([]entry, error) {
	return nil, nil
}

func readPointer(
	_ context.Context,
	_ fakeProvider,
	_ getIn,
) (*entry, error) {
	return &entry{}, nil
}

func readString(
	_ context.Context,
	_ fakeProvider,
	_ getIn,
) (string, error) {
	return "", nil
}

func readBadTag(
	_ context.Context,
	_ fakeProvider,
	_ badTag,
) (entry, error) {
	return entry{}, nil
}

func readBadOut(
	_ context.Context,
	_ fakeProvider,
	_ getIn,
) (badKind, error) {
	return badKind{}, nil
}

func readReserved(
	_ context.Context,
	_ fakeProvider,
	_ getIn,
) (reserved, error) {
	return reserved{}, nil
}

func writeNoChanged(
	_ context.Context,
	_ fakeProvider,
	_ getIn,
) (noChanged, error) {
	return noChanged{}, nil
}

func writeStringChanged(
	_ context.Context,
	_ fakeProvider,
	_ getIn,
) (stringChanged, error) {
	return stringChanged{}, nil
}

func (s *ValidatePublicTestSuite) SetupSuite() {
	behavior.ValidatorSchema("test_custom", behavior.Pattern(`^x+$`))
}

func (s *ValidatePublicTestSuite) TestValidateBehavior() {
	tests := []struct {
		name      string
		b         *behavior.Behavior
		wantRules []string
	}{
		{
			name: "when the definition is complete",
			b: behavior.Define("sysctl", "Kernel parameters", behavior.Node,
				behavior.Read("list", readList).Describe("List"),
				behavior.Read("one", readPointer).Describe("One"),
				behavior.Read("get", get).Key("key").Describe("Get"),
				behavior.Write("create", set).Describe("Create"),
			),
		},
		{
			name: "when names are not lowercase",
			b: behavior.Define("Sysctl", "x", behavior.Node, behavior.Under("Net"),
				behavior.Read("Get", get).Describe("Get"),
			),
			wantRules: []string{behavior.RuleName, behavior.RuleName, behavior.RuleName},
		},
		{
			name:      "when there is no placement, description or method",
			b:         behavior.Define("bare", ""),
			wantRules: []string{behavior.RulePlacement, behavior.RuleDoc, behavior.RuleMethods},
		},
		{
			name: "when a method is declared twice without a description",
			b: behavior.Define("sysctl", "x", behavior.Node,
				behavior.Read("get", get),
				behavior.Read("get", get).Describe("Get"),
			),
			wantRules: []string{behavior.RuleDoc, behavior.RuleDuplicate},
		},
		{
			name: "when the key is not an input field",
			b: behavior.Define("sysctl", "x", behavior.Node,
				behavior.Read("get", get).Key("nope").Describe("Get"),
			),
			wantRules: []string{behavior.RuleKey},
		},
		{
			name: "when the input has no schema",
			b: behavior.Define("sysctl", "x", behavior.Node,
				behavior.Read("get", readBadTag).Describe("Get"),
			),
			wantRules: []string{behavior.RuleSchema},
		},
		{
			name: "when the output has no schema",
			b: behavior.Define("sysctl", "x", behavior.Node,
				behavior.Read("get", readBadOut).Describe("Get"),
			),
			wantRules: []string{behavior.RuleSchema},
		},
		{
			name: "when the output is not a struct",
			b: behavior.Define("sysctl", "x", behavior.Node,
				behavior.Read("get", readString).Describe("Get"),
			),
			wantRules: []string{behavior.RuleOutputShape},
		},
		{
			name: "when the output declares a per-host field",
			b: behavior.Define("sysctl", "x", behavior.Node,
				behavior.Read("get", readReserved).Describe("Get"),
			),
			wantRules: []string{behavior.RuleOutputShape},
		},
		{
			name: "when a write has no changed field",
			b: behavior.Define("sysctl", "x", behavior.Node,
				behavior.Write("create", writeNoChanged).Describe("Create"),
			),
			wantRules: []string{behavior.RuleChanged},
		},
		{
			name: "when a write's changed is not a boolean",
			b: behavior.Define("sysctl", "x", behavior.Node,
				behavior.Write("create", writeStringChanged).Describe("Create"),
			),
			wantRules: []string{behavior.RuleChanged},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			rules := []string{}
			for _, p := range behavior.ValidateBehavior(tc.b) {
				assert.NotEmpty(s.T(), p.Message)
				rules = append(rules, p.Rule)
			}

			if tc.wantRules == nil {
				tc.wantRules = []string{}
			}
			assert.Equal(s.T(), tc.wantRules, rules)
		})
	}
}

func (s *ValidatePublicTestSuite) TestValidate() {
	tests := []struct {
		name      string
		setup     func()
		wantRules []string
	}{
		{
			name:      "when nothing is registered",
			setup:     func() {},
			wantRules: []string{},
		},
		{
			name: "when a registered behavior breaks a rule",
			setup: func() {
				behavior.Register(behavior.Define("bare", "x", behavior.Node))
			},
			wantRules: []string{behavior.RuleMethods},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			behavior.Reset()
			tc.setup()

			problems := behavior.Validate()
			rules := make([]string, 0, len(problems))
			for _, p := range problems {
				rules = append(rules, p.Rule)
			}
			assert.Equal(s.T(), tc.wantRules, rules)
		})
	}
}

func TestValidatePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(ValidatePublicTestSuite))
}

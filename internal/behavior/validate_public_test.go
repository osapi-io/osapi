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

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
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
		// wantMessages, when set, are the messages of the settings
		// problems, in order.
		wantMessages []string
	}{
		{
			name: "when the definition is complete",
			b: behavior.Define("sysctl", "Kernel parameters", behavior.Node, onDebian,
				behavior.Read("list", readList).Describe("List"),
				behavior.Read("status", readPointer).Key("key").Describe("One"),
				behavior.Write("create", set).Describe("Create"),
			),
		},
		{
			name: "when names are not lowercase",
			b: behavior.Define("Sysctl", "x", behavior.Node, onDebian, behavior.Under("Net"),
				behavior.Read("Get", get).Describe("Get"),
			),
			wantRules: []string{behavior.RuleName, behavior.RuleName, behavior.RuleName},
		},
		{
			name: "when there is no placement, description, method or platform",
			b:    behavior.Define("bare", ""),
			wantRules: []string{
				behavior.RulePlacement,
				behavior.RuleDoc,
				behavior.RuleMethods,
				behavior.RulePlatform,
			},
		},
		{
			name: "when a method is declared twice without a description",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Read("get", get),
				behavior.Read("get", get).Describe("Get"),
			),
			wantRules: []string{behavior.RuleDoc, behavior.RuleDuplicate},
		},
		{
			name: "when the key is not an input field",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Read("get", get).Key("nope").Describe("Get"),
			),
			wantRules: []string{behavior.RuleKey},
		},
		{
			name: "when the input has no schema",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Read("get", readBadTag).Describe("Get"),
			),
			wantRules: []string{behavior.RuleSchema},
		},
		{
			name: "when the output has no schema",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Read("get", readBadOut).Describe("Get"),
			),
			wantRules: []string{behavior.RuleSchema},
		},
		{
			name: "when the output is not a struct",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Read("get", readString).Describe("Get"),
			),
			wantRules: []string{behavior.RuleOutputShape},
		},
		{
			name: "when the output declares a per-host field",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Read("get", readReserved).Describe("Get"),
			),
			wantRules: []string{behavior.RuleOutputShape},
		},
		{
			name: "when a write has no changed field",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Write("create", writeNoChanged).Describe("Create"),
			),
			wantRules: []string{behavior.RuleChanged},
		},
		{
			name: "when a behavior has a CLI without an API",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian, behavior.CLI,
				behavior.Read("get", get).Describe("Get").
					Example("One", getIn{Key: "k"}, entry{Key: "k"}),
			),
			wantRules: []string{behavior.RuleSurface},
		},
		{
			name: "when a CLI method has no example",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian, behavior.API, behavior.CLI,
				behavior.Read("get", get).Describe("Get"),
			),
			wantRules: []string{behavior.RuleExample},
		},
		{
			name: "when an example has no description and the wrong types",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Read("get", get).Describe("Get").
					Example("", result{}, getIn{}),
			),
			wantRules: []string{behavior.RuleExample, behavior.RuleExample, behavior.RuleExample},
		},
		{
			name: "when two methods share a route",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Read("list", readList).Describe("List"),
				behavior.Read("status", readList).Describe("Status"),
			),
			wantRules: []string{behavior.RuleRoute},
		},
		{
			name: "when a method is named with a synonym",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Write("remove", set).Describe("Remove"),
			),
			wantRules: []string{behavior.RuleVerb},
		},
		{
			name: "when a read is not list, get or status",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Read("fetch", get).Describe("Fetch"),
			),
			wantRules: []string{behavior.RuleVerb},
		},
		{
			name: "when a write is named for an action",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Write("restart", set).Describe("Restart"),
			),
			wantRules: []string{behavior.RuleVerb},
		},
		{
			name: "when a write's changed is not a boolean",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Write("create", writeStringChanged).Describe("Create"),
			),
			wantRules: []string{behavior.RuleChanged},
		},
		{
			name: "when a CLI input field cannot be a flag",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian, behavior.API, behavior.CLI,
				behavior.Read("list", listTagged).Describe("List").
					Example("All", taggedIn{}, []entry{}),
			),
			wantRules: []string{behavior.RuleSurface},
		},
		{
			name: "when an input is not a struct",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian, behavior.API, behavior.CLI,
				behavior.Read("status", readBare).Describe("Status").
					Example("All", "x", entry{}),
			),
			wantRules: []string{behavior.RuleSchema},
		},
		{
			name: "when every setting is named and documented",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Config(goodSettings{}),
				behavior.Read("get", get).Describe("Get"),
			),
		},
		{
			name: "when settings are not a struct",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Config("dir"),
				behavior.Read("get", get).Describe("Get"),
			),
			wantRules:    []string{behavior.RuleSettings},
			wantMessages: []string{"settings must be a struct, not string"},
		},
		{
			name: "when settings fields lack a name or a doc",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Config(badSettings{}),
				behavior.Read("get", get).Describe("Get"),
			),
			wantRules: []string{
				behavior.RuleSettings,
				behavior.RuleSettings,
				behavior.RuleSettings,
			},
			wantMessages: []string{
				"settings field NoName has no mapstructure name",
				"setting no_doc has no doc",
				"setting nested.inner has no doc",
			},
		},
		{
			name: "when a refusal declares effects",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Read("get", get).Describe("Get").
					Fails("Unmanaged", getIn{Key: "k"}, "not managed", behavior.Wrote("/etc/k", "x")),
			),
			wantRules: []string{behavior.RuleExample},
		},
		{
			name: "when a refusal has no text and the wrong input",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Read("get", get).Describe("Get").
					Fails("", entry{}, ""),
			),
			wantRules: []string{behavior.RuleExample, behavior.RuleExample},
		},
		{
			name: "when a secret is also recorded on the span",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Read("get", readLeaky).Describe("Get"),
			),
			wantRules: []string{behavior.RuleSecret},
		},
		{
			name: "when an internal behavior is also served",
			b: behavior.Define(
				"sysctl",
				"x",
				behavior.Node,
				onDebian,
				behavior.Internal,
				behavior.API,
				behavior.Read("get", get).Describe("Get"),
			),
			wantRules: []string{behavior.RuleSurface},
		},
		{
			name: "when a platform is declared twice",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian, onDebian,
				behavior.Read("get", get).Describe("Get"),
			),
			wantRules: []string{behavior.RulePlatform},
		},
		{
			name: "when a platform's provider does not fit a method",
			b: behavior.Define("sysctl", "x", behavior.Node,
				behavior.Platform("darwin", func(host.Host) (string, error) { return "", nil }),
				behavior.Read("get", get).Describe("Get"),
			),
			wantRules: []string{behavior.RulePlatform},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			rules := []string{}
			var messages []string
			for _, p := range behavior.ValidateBehavior(tc.b) {
				assert.NotEmpty(s.T(), p.Message)
				rules = append(rules, p.Rule)
				if p.Rule == behavior.RuleSettings {
					messages = append(messages, p.Message)
				}
			}

			if tc.wantRules == nil {
				tc.wantRules = []string{}
			}
			assert.Equal(s.T(), tc.wantRules, rules)
			if tc.wantMessages != nil {
				assert.Equal(s.T(), tc.wantMessages, messages)
			}
		})
	}
}

// complete is a behavior that breaks no rule of its own, with opts added.
func complete(
	name string,
	opts ...behavior.Option,
) *behavior.Behavior {
	return behavior.Define(name, "x", append([]behavior.Option{
		behavior.Node, onDebian, behavior.Read("get", get).Describe("Get"),
	}, opts...)...)
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
				behavior.Register(behavior.Define("bare", "x", behavior.Node, onDebian))
			},
			wantRules: []string{behavior.RuleMethods},
		},
		{
			name: "when a behavior requires a registered one",
			setup: func() {
				files := behavior.Register(complete("files", behavior.Internal))
				behavior.Register(complete("sysctl", behavior.Requires(files)))
			},
			wantRules: []string{},
		},
		{
			name: "when a required behavior is not registered, or is nil",
			setup: func() {
				behavior.Register(complete("sysctl", behavior.Requires(complete("files"), nil)))
			},
			wantRules: []string{behavior.RuleRequires, behavior.RuleRequires},
		},
		{
			name: "when behaviors require each other",
			setup: func() {
				a := behavior.Register(complete("a"))
				b := behavior.Register(complete("b", behavior.Requires(a)))
				c := behavior.Register(complete("c", behavior.Requires(b)))
				a.Requires = []*behavior.Behavior{c}
			},
			wantRules: []string{
				behavior.RuleRequires,
				behavior.RuleRequires,
				behavior.RuleRequires,
			},
		},
		{
			name: "when a behavior requires a cycle it is not part of, only the cycle fails",
			setup: func() {
				b := behavior.Register(complete("b"))
				c := behavior.Register(complete("c", behavior.Requires(b)))
				b.Requires = []*behavior.Behavior{c}
				behavior.Register(complete("a", behavior.Requires(b)))
			},
			wantRules: []string{behavior.RuleRequires, behavior.RuleRequires},
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

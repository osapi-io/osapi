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

	"github.com/osapi-io/osapi/internal/validation/rule"

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
	Key string `json:"key" doc:"test"`
}

type stringChanged struct {
	Changed string `json:"changed" doc:"test"`
}

type reserved struct {
	Hostname string `json:"hostname" doc:"test"`
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

// noFields is an input with no field that a span can record.
type noFields struct{}

// onlySecret is an input whose only field is a secret.
type onlySecret struct {
	Password string `json:"password" secret:"true" validate:"omitempty" doc:"test"`
}

// taggedName is an input whose name is tagged for the span.
// sentEmptyIn has an optional body field whose json tag lacks omitempty,
// and a key and an ignored field that the rule leaves alone.
type sentEmptyIn struct {
	Name    string `json:"name"           validate:"required"        doc:"test"`
	Version string `json:"version"        validate:"omitempty,max=9" doc:"test"`
	Note    string `json:"note,omitempty" validate:"omitempty,max=9" doc:"test"`
	Skip    string `json:"-"              validate:"omitempty"`
}

func setSentEmpty(context.Context, fakeProvider, sentEmptyIn) (result, error) {
	return result{}, nil
}

// wideIn has an int64, which a CLI flag does not hold: the command
// generator builds an int flag only.
type wideIn struct {
	Limit int64 `json:"limit" validate:"omitempty" doc:"test"`
}

func listWide(context.Context, fakeProvider, wideIn) ([]entry, error) {
	return nil, nil
}

type untaggedIn struct {
	Name string `json:"name" doc:"test"`
}

type limitedIn struct {
	Name string `json:"name" validate:"omitempty,max=3" doc:"test"`
}

func readLimited(context.Context, fakeProvider, limitedIn) (entry, error) {
	return entry{}, nil
}

func readUntagged(context.Context, fakeProvider, untaggedIn) (entry, error) {
	return entry{}, nil
}

type taggedName struct {
	Name string `json:"name" telemetry:"attr" validate:"required" doc:"test"`
}

func setNoFields(context.Context, fakeProvider, noFields) (result, error) {
	return result{}, nil
}

func setOnlySecret(context.Context, fakeProvider, onlySecret) (result, error) {
	return result{}, nil
}

func setTagged(context.Context, fakeProvider, taggedName) (result, error) {
	return result{}, nil
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
	rule.Register("test_custom", rule.Pattern(`^x+$`))
}

// resultIn is an input whose key, result, is named as the row type Result.
type resultIn struct {
	Result string `json:"result" validate:"required" doc:"The result"`
}

func getResult(
	_ context.Context,
	_ fakeProvider,
	_ resultIn,
) (Result, error) {
	return Result{}, nil
}

// Result is a row type that a key named result collides with.
type Result struct {
	Value string `json:"value" doc:"The value"`
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
				behavior.Write("create", set).Key("key").Describe("Create"),
			),
		},
		{
			name: "when a write's optional body field is sent empty",
			b: behavior.Define(
				"pkg",
				"Packages",
				behavior.Node,
				behavior.API,
				behavior.Area("config"),
				onDebian,
				behavior.Write("create", setSentEmpty).Key("name").Describe("Install").
					Fails("too long", sentEmptyIn{Name: "a", Version: "0123456789"}, "x").
					Fails("too long", sentEmptyIn{Name: "a", Note: "0123456789"}, "x"),
			),
			wantRules: []string{behavior.RuleOptional},
		},
		{
			name: "when a served behavior has no area of the UI",
			b: behavior.Define("sysctl", "Kernel parameters", behavior.Node, onDebian, behavior.API,
				behavior.Read("list", readList).Describe("List"),
			),
			wantRules: []string{behavior.RuleSurface},
		},
		{
			name: "when a behavior is Under a group that is not one of KnownGroups",
			b: behavior.Define("dns", "Resolvers", behavior.Node, onDebian,
				behavior.Under("netwrok"),
				behavior.Read("list", readList).Describe("List"),
			),
			wantRules: []string{behavior.RuleName},
		},
		{
			name: "when a served behavior's area is not one of Areas",
			b: behavior.Define("sysctl", "Kernel parameters", behavior.Node, onDebian, behavior.API,
				behavior.Area("kitchen"),
				behavior.Read("list", readList).Describe("List"),
			),
			wantRules: []string{behavior.RuleSurface},
		},
		{
			name: "when the optional field is on a behavior with no API",
			b: behavior.Define("pkg", "Packages", behavior.Node, behavior.Internal, onDebian,
				behavior.Write("create", setSentEmpty).Key("name").Describe("Install").
					Fails("too long", sentEmptyIn{Name: "a", Version: "0123456789"}, "x").
					Fails("too long", sentEmptyIn{Name: "a", Note: "0123456789"}, "x"),
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
			name: "when a behavior or its group is named all",
			b: behavior.Define("all", "x", behavior.Node, onDebian, behavior.Under("all"),
				behavior.Read("get", get).Describe("Get"),
			),
			wantRules: []string{behavior.RuleName},
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
			name: "when a description or explanation has markup",
			b: behavior.Define("sysctl", "a <key>", behavior.Node, onDebian,
				behavior.Read("get", get).Describe("Get {key}").Explain("Writes osapi-<key>.conf"),
			),
			wantRules: []string{behavior.RuleDoc, behavior.RuleDoc, behavior.RuleDoc},
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
				behavior.Write("create", writeNoChanged).Key("key").Describe("Create"),
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
			b: behavior.Define(
				"sysctl",
				"x",
				behavior.Node,
				onDebian,
				behavior.API,
				behavior.Area("config"),
				behavior.CLI,
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
				behavior.Write("remove", set).Key("key").Describe("Remove"),
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
				behavior.Write("restart", set).Key("key").Describe("Restart"),
			),
			wantRules: []string{behavior.RuleVerb},
		},
		{
			name: "when an action is not one of the CLI's actions",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Action("bounce", set).Key("key").Describe("Bounce"),
			),
			wantRules: []string{behavior.RuleVerb},
		},
		{
			name: "when an action is named with a standard verb",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Action("update", set).Key("key").Describe("Update"),
			),
			wantRules: []string{behavior.RuleVerb},
		},
		{
			name: "when an action is named with a read verb that is also a CLI action",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Action("status", set).Key("key").Describe("Status"),
			),
			wantRules: []string{behavior.RuleVerb},
		},
		{
			name: "when a key's name in the spec is a row type's",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Read("get", getResult).Key("result").Describe("Get"),
			),
			wantRules: []string{behavior.RuleKey},
		},
		{
			name: "when a child is also under a group",
			b: behavior.Define(
				"part",
				"x",
				behavior.Node,
				onDebian,
				behavior.Of(
					behavior.Define(
						"item",
						"x",
						behavior.Node,
						behavior.Read("get", get).Key("key"),
					),
				),
				behavior.Under("network"),
				behavior.Read("list", get).Describe("List"),
			),
			wantRules: []string{behavior.RuleParent},
		},
		{
			name: "when a child's parent has no key",
			b: behavior.Define(
				"part",
				"x",
				behavior.Node,
				onDebian,
				behavior.Of(
					behavior.Define("item", "x", behavior.Node, behavior.Read("list", get)),
				),
				behavior.Read("list", get).Describe("List"),
			),
			wantRules: []string{behavior.RuleParent},
		},
		{
			name: "when a child's parent is not registered",
			b: behavior.Define(
				"part",
				"x",
				behavior.Node,
				onDebian,
				behavior.Of(
					behavior.Define(
						"item",
						"x",
						behavior.Node,
						behavior.Read("get", get).Key("key"),
					),
				),
				behavior.Read("list", get).Describe("List"),
			),
			wantRules: []string{behavior.RuleParent},
		},
		{
			name: "when a write that is not an action always changes",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Write("update", set).Key("key").Describe("Update").Always(),
			),
			wantRules: []string{behavior.RuleVerb},
		},
		{
			name: "when a write records nothing on its span",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Write("create", set).Describe("Create"),
			),
			wantRules: []string{behavior.RuleSpan},
		},
		{
			name: "when a write without a key tags the field that names the item",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Write("create", setTagged).Describe("Create"),
			),
		},
		{
			name: "when a write has no field that a span can record",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Write("create", setNoFields).Describe("Create"),
				behavior.Write("update", setOnlySecret).Describe("Update"),
			),
		},
		{
			name: "when a write's changed is not a boolean",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Write("create", writeStringChanged).Key("key").Describe("Create"),
			),
			wantRules: []string{behavior.RuleChanged},
		},
		{
			name: "when a CLI input field cannot be a flag",
			b: behavior.Define(
				"sysctl",
				"x",
				behavior.Node,
				onDebian,
				behavior.API,
				behavior.Area("config"),
				behavior.CLI,
				behavior.Read("list", listWide).Describe("List").
					Example("All", wideIn{}, []entry{}),
			),
			wantRules: []string{behavior.RuleSurface},
		},
		{
			name: "when an input is not a struct",
			b: behavior.Define(
				"sysctl",
				"x",
				behavior.Node,
				onDebian,
				behavior.API,
				behavior.Area("config"),
				behavior.CLI,
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
			name: "when no .Fails input breaks a rule",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Read("get", readLimited).Describe("Get"),
			),
			wantRules: []string{behavior.RuleFails},
		},
		{
			name: "when a .Fails input breaks each rule",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Read("get", readLimited).Describe("Get").
					Fails("a name over 3 characters", limitedIn{Name: "abcd"}, "too long"),
			),
		},
		{
			name: "when an input field has no validate tag",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Read("get", readUntagged).Describe("Get"),
			),
			wantRules: []string{behavior.RuleValidate},
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
				behavior.API, behavior.Area("config"),
				behavior.Read("get", get).Describe("Get"),
			),
			wantRules: []string{behavior.RuleSurface},
		},
		{
			name: "when On names a platform the behavior has no provider for",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Read("get", get).Describe("Get").On(behavior.PlatformDarwin),
			),
			wantRules: []string{behavior.RulePlatform},
		},
		{
			name: "when an input field's secret tag is not true",
			b: behavior.Define(
				"sysctl",
				"x",
				behavior.Node,
				onDebian,
				behavior.Read("get", func(context.Context, fakeProvider, badSecretIn) (entry, error) {
					return entry{}, nil
				}).
					Describe("Get"),
			),
			wantRules: []string{behavior.RuleTag},
		},
		{
			name: "when an input field has a tag key that nothing reads",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Read("get", func(context.Context, fakeProvider, badKeyIn) (entry, error) {
					return entry{}, nil
				}).Describe("Get"),
			),
			wantRules: []string{behavior.RuleTag},
		},
		{
			name: "when an output field's secret tag is not true",
			b: behavior.Define(
				"sysctl",
				"x",
				behavior.Node,
				onDebian,
				behavior.Read("get", func(context.Context, fakeProvider, getIn) (badSecretOut, error) {
					return badSecretOut{}, nil
				}).
					Describe("Get"),
			),
			wantRules: []string{behavior.RuleTag},
		},
		{
			name: "when an input field's telemetry tag is not attr",
			b: behavior.Define(
				"sysctl",
				"x",
				behavior.Node,
				onDebian,
				behavior.Read("get", func(context.Context, fakeProvider, badTelemetryIn) (entry, error) {
					return entry{}, nil
				}).
					Describe("Get"),
			),
			wantRules: []string{behavior.RuleTag},
		},
		{
			name: "when a field of the output, in a slice, has a tag key that nothing reads",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Read("get", func(context.Context, fakeProvider, getIn) (badKeyOut, error) {
					return badKeyOut{}, nil
				}).Describe("Get"),
			),
			wantRules: []string{behavior.RuleOutputShape},
		},
		{
			name: "when Platform names a platform that is not known",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Platform("debain", func(host.Host) (fakeProvider, error) {
					return fakeProvider{}, nil
				}),
				behavior.Read("get", get).Describe("Get"),
			),
			wantRules: []string{behavior.RulePlatform},
		},
		{
			name: "when On names a platform that is not known, with a linux provider",
			b: behavior.Define("sysctl", "x", behavior.Node,
				behavior.Platform(behavior.PlatformLinux, func(host.Host) (fakeProvider, error) {
					return fakeProvider{}, nil
				}),
				behavior.Read("get", get).Describe("Get").On("debain"),
			),
			wantRules: []string{behavior.RulePlatform},
		},
		{
			name: "when RunsOnPlatform names a platform that is not known",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.RunsOnPlatform("darwn", "vm_stat"),
				behavior.Read("get", get).Describe("Get"),
			),
			wantRules: []string{behavior.RulePlatform},
		},
		{
			name: "when an example or a refusal names a platform that is not known",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Read("get", get).Describe("Get").
					Example("A read", getIn{Key: "a"}, entry{}, behavior.OnPlatform("debain")).
					Fails("A refusal", getIn{Key: "b"}, "no", behavior.OnPlatform("debain")),
			),
			wantRules: []string{behavior.RulePlatform, behavior.RulePlatform},
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
				assert.NotEmpty(s.T(), p.Fix, "every rule has a fix")
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
		{
			name: "when an internal behavior is required by none",
			setup: func() {
				behavior.Register(complete("files", behavior.Internal))
			},
			wantRules: []string{behavior.RuleInternal},
		},
		{
			name: "when the paths a behavior writes are absolute patterns, or any",
			setup: func() {
				behavior.Register(complete("conf", behavior.Writes("/etc/conf.d/*.conf", "*")))
			},
			wantRules: []string{},
		},
		{
			name: "when a path it writes is relative, or not a pattern",
			setup: func() {
				behavior.Register(complete("conf", behavior.Writes("etc/conf", "/etc/[")))
			},
			wantRules: []string{behavior.RuleWrites, behavior.RuleWrites},
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

type badSecretIn struct {
	Key string `json:"key" validate:"required" doc:"The key" secret:"True"`
}

type badTelemetryIn struct {
	Key string `json:"key" validate:"required" doc:"The key" telemetry:"atr"`
}

type badKeyOut struct {
	Entries []struct {
		Key string `json:"key" doc:"The \"key\"" secrte:"true" `
	} `json:"entries" doc:"The entries"`
}

type badKeyIn struct {
	Key string `json:"key" validate:"required" doc:"The key" secrte:"true"`
}

type badSecretOut struct {
	Key string `json:"key" doc:"The key" secret:"yes"`
}

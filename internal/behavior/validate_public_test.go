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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/host"
	"github.com/osapi-io/osapi/internal/secret"
	"github.com/osapi-io/osapi/internal/validation/rule"
	"github.com/osapi-io/osapi/pkg/sdk/types"
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

// readNames answers a list of strings, a shape that no read has: its rows
// are not structs.
func readNames(
	_ context.Context,
	_ fakeProvider,
	_ getIn,
) ([]string, error) {
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
	Password secret.Value `json:"password" validate:"omitempty" doc:"test"`
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
		// validateFunc, when set, checks the messages of each rule in
		// place of wantRules.
		validateFunc func(byRule map[string][]string)
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
			b: behavior.Define("Sysctl", "x", behavior.Node, onDebian,
				behavior.Read("Get", get).Describe("Get"),
			),
			wantRules: []string{behavior.RuleName, behavior.RuleName},
		},
		{
			name: "when a behavior is named all",
			b: behavior.Define("all", "x", behavior.Node, onDebian,
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
			name: "when a child under the collection has the name of an action of its parent",
			b: behavior.Define(
				"reboot",
				"x",
				behavior.Node,
				onDebian,
				behavior.Under(
					behavior.Define(
						"item",
						"x",
						behavior.Node,
						behavior.Action("reboot", set).Describe("Reboot"),
					),
				),
				behavior.Read("list", get).Describe("List"),
			),
			wantRules: []string{behavior.RuleParent},
		},
		{
			name: "when a child under the collection's parent is not registered",
			b: behavior.Define(
				"part",
				"x",
				behavior.Node,
				onDebian,
				behavior.Under(
					behavior.Define("item", "x", behavior.Node, behavior.Read("list", get)),
				),
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
			name: "when a delete always changes",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Write("delete", set).Key("key").Describe("Delete").Always(),
			),
			wantRules: []string{behavior.RuleVerb},
		},
		{
			name: "when an update always changes, as a password does",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Write("update", set).Key("key").Describe("Update").Always(),
			),
		},
		{
			name: "when a method disrupts the host with no reason",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Write("update", set).Key("key").Describe("Update").Disrupts(" "),
			),
			wantRules: []string{behavior.RuleVerb},
		},
		{
			name: "when a write that is not an action runs a shell",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Write("update", set).Key("key").Describe("Update").Shell(),
			),
			wantRules: []string{behavior.RuleVerb},
		},
		{
			name: "when an action takes a permission verb that only the admin role gets",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Action("exec", set).Key("key").Describe("Exec").
					PermitAs(types.VerbExec),
			),
		},
		{
			name: "when an action takes a permission verb that other roles get",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.Action("exec", set).Key("key").Describe("Exec").
					PermitAs(types.VerbWrite),
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
			name: "when a .Before step has the method's input in and out, it breaks no rule",
			b: behavior.Define("account", "Accounts", behavior.Node, onDebian,
				behavior.API, behavior.Area(area.Config),
				behavior.Write("update", setPassword).Describe("Update").Before(hashStep).
					Fails("no name", pwIn{}, "name"),
			),
		},
		{
			name: "when each write example reads back its result",
			b: behavior.Define(
				"sysctl",
				"x",
				behavior.Node,
				behavior.API,
				behavior.CLI,
				behavior.Area(area.Config),
				onDebian,
				behavior.Read("get", get).Key("key").Describe("Get").
					Example("A key", getIn{Key: "a"}, entry{Key: "a"}).
					Fails("no key", getIn{}, "key"),
				behavior.Write("create", set).Key("key").Describe("Create").
					Example("Set it", getIn{Key: "a"}, result{Key: "a", Changed: true},
						behavior.Reads(getIn{Key: "a"}, entry{Key: "a", Value: "1"})).
					Example("Clear it", getIn{Key: "a"}, result{Key: "a"},
						behavior.ReadsNone(getIn{Key: "a"}, entry{Key: "a"}, "key", "value")).
					Example("A key it has no read of", getIn{Key: "a"}, result{Key: "a"},
						behavior.NoReadBack("the read cannot name this key on a live host")).
					Fails("no key", getIn{}, "key"),
				behavior.Action("restart", set).Key("key").Describe("Restart").
					ReadsNothing("its result is its own output").
					Example("Restart it", getIn{Key: "a"}, result{Key: "a"}).
					Fails("no key", getIn{}, "key"),
			),
		},
		{
			name: "when a write example of a list reads back a row of the list",
			b: behavior.Define(
				"sysctl",
				"x",
				behavior.Node,
				behavior.API,
				behavior.CLI,
				behavior.Area(area.Config),
				onDebian,
				behavior.Read("list", readList).Describe("List").
					Example("Each key", getIn{Key: "a"}, []entry{{Key: "a"}}).
					Fails("no key", getIn{}, "key"),
				behavior.Write("create", set).Key("key").Describe("Create").
					Example("Set it", getIn{Key: "a"}, result{Key: "a", Changed: true},
						behavior.Reads(getIn{Key: "a"}, entry{Key: "a"})).
					Fails("no key", getIn{}, "key"),
			),
		},
		{
			name: "when a write example does not read back its result, or reads it wrong",
			b: behavior.Define(
				"sysctl",
				"x",
				behavior.Node,
				behavior.API,
				behavior.CLI,
				behavior.Area(area.Config),
				onDebian,
				behavior.Read("get", get).Key("key").Describe("Get").
					Example("A key", getIn{Key: "a"}, entry{Key: "a"},
						behavior.Reads(getIn{Key: "a"}, entry{Key: "a"})).
					Fails("no key", getIn{}, "key"),
				behavior.Write("create", set).Key("key").Describe("Create").
					Example("No read", getIn{Key: "a"}, result{Key: "a", Changed: true}).
					Example("No read takes it", getIn{Key: "a"}, result{Key: "a"},
						behavior.Reads(result{Key: "a"}, entry{Key: "a"})).
					Example("Another type", getIn{Key: "a"}, result{Key: "a"},
						behavior.Reads(getIn{Key: "a"}, result{Key: "a"})).
					Example("No such field", getIn{Key: "a"}, result{Key: "a"},
						behavior.Reads(getIn{Key: "a"}, entry{Key: "a"}, "nope")).
					Example("Nothing to compare", getIn{Key: "a"}, result{Key: "a"},
						behavior.Reads(getIn{Key: "a"}, entry{})).
					Fails("no key", getIn{}, "key"),
			),
			wantRules: []string{
				behavior.RuleReadBack, behavior.RuleReadBack, behavior.RuleReadBack,
				behavior.RuleReadBack, behavior.RuleReadBack, behavior.RuleReadBack,
			},
		},
		{
			name: "when a read back reads rows that are not structs it is a problem, not a panic",
			b: behavior.Define(
				"sysctl",
				"x",
				behavior.Node,
				behavior.API,
				behavior.CLI,
				behavior.Area(area.Config),
				onDebian,
				behavior.Read("list", readNames).Describe("List").
					Example("Each name", getIn{Key: "a"}, []string{"a"}).
					Fails("no key", getIn{}, "key"),
				behavior.Write("create", set).Key("key").Describe("Create").
					Example("Set it", getIn{Key: "a"}, result{Key: "a", Changed: true},
						behavior.Reads(getIn{Key: "a"}, "a")).
					Fails("no key", getIn{}, "key"),
			),
			wantRules: []string{behavior.RuleOutputShape, behavior.RuleReadBack},
		},
		{
			name: "when a .Before step is not func(context.Context, In) (In, error)",
			b: behavior.Define("account", "Accounts", behavior.Node, onDebian,
				behavior.API, behavior.Area(area.Config),
				behavior.Write("update", setPassword).Describe("Update").
					Before(func(pwIn) pwIn { return pwIn{} }).
					Fails("no name", pwIn{}, "name"),
			),
			wantRules: []string{behavior.RuleBefore},
		},
		{
			name: "when a .Before step is on a behavior that is not served",
			b: behavior.Define("account", "Accounts", behavior.Node, onDebian,
				behavior.Write("update", setPassword).Describe("Update").Before(hashStep).
					Fails("no name", pwIn{}, "name"),
			),
			wantRules: []string{behavior.RuleBefore},
		},
		{
			name: "when a .Before step has no secret field to keep out of the job",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.API, behavior.Area(area.Config),
				behavior.Write("update", set).Describe("Update").Key("key").
					Before(func(_ context.Context, in getIn) (getIn, error) { return in, nil }).
					Fails("no key", getIn{}, "key"),
			),
			wantRules: []string{behavior.RuleBefore},
		},
		{
			name: "when a field is tagged before with no step, or a value other than set",
			b: behavior.Define(
				"account",
				"Accounts",
				behavior.Node,
				onDebian,
				behavior.API,
				behavior.Area(area.Config),
				behavior.Write("update", func(context.Context, fakeProvider, badBeforeIn) (result, error) {
					return result{}, nil
				}).
					Describe("Update").
					Fails("no name", badBeforeIn{}, "name"),
			),
			wantRules: []string{behavior.RuleBefore, behavior.RuleBefore},
		},
		{
			name: "when an input field has the old secret tag, which is not a key",
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
			name: "when an output field has the old secret tag, which is not a key",
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
			wantRules: []string{behavior.RuleOutputShape},
		},
		{
			name: "when an output field's telemetry tag is not attr",
			b: behavior.Define(
				"sysctl",
				"x",
				behavior.Node,
				onDebian,
				behavior.Read("get", func(context.Context, fakeProvider, getIn) (badTelemetryOut, error) {
					return badTelemetryOut{}, nil
				}).
					Describe("Get"),
			),
			wantRules: []string{behavior.RuleTag},
		},
		{
			name: "when a served read takes a secret, which goes in its query",
			b: behavior.Define(
				"sysctl",
				"x",
				behavior.Node,
				onDebian,
				behavior.API,
				behavior.Area(area.Config),
				behavior.Read("list", func(context.Context, fakeProvider, onlySecret) ([]entry, error) {
					return nil, nil
				}).
					Describe("List"),
			),
			wantRules: []string{behavior.RuleSecret},
		},
		{
			name: "when the key of a served write is a secret, which goes in its path",
			b: behavior.Define(
				"sysctl",
				"x",
				behavior.Node,
				onDebian,
				behavior.API,
				behavior.Area(area.Config),
				behavior.Write("update", func(context.Context, fakeProvider, secretKeyIn) (result, error) {
					return result{}, nil
				}).
					Describe("Update").
					Key("password").
					Fails("no password", secretKeyIn{}, "required"),
			),
			wantRules: []string{behavior.RuleSecret},
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
			name: "when SudoOnPlatform names a platform that is not known",
			b: behavior.Define("sysctl", "x", behavior.Node, onDebian,
				behavior.SudoOnPlatform("debain", "update-ca-certificates"),
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
		{
			name: "when a key is optional, which a path segment never is",
			b: behavior.Define("note", "x", behavior.Node, onDebian,
				behavior.Read("get", readOptionalKey).Key("name").Describe("One")),
			validateFunc: func(byRule map[string][]string) {
				s.Equal([]string{
					`key "name" is optional, but a key is a path segment, which is never empty`,
				}, byRule[behavior.RuleKey])
			},
		},
		{
			name: "when the input fields break the input rules",
			b: behavior.Define("note", "x", behavior.Node, onDebian,
				behavior.API, behavior.Area("config"), behavior.CLI,
				behavior.Read("list", readBadInput).Describe("List")),
			validateFunc: func(byRule map[string][]string) {
				s.Equal([]string{
					"input field NoJSON has no json name",
					"input field host is named like the target, which every request has",
					"input field target is named like the target, which every request has",
					"input field nodoc has no doc",
					"input field def has a default, which the server does not apply",
					"input field arr is an array; use a slice",
				}, byRule[behavior.RuleInput])
			},
		},
		{
			name: "when the output fields break the output rules",
			b: behavior.Define("note", "x", behavior.Node, onDebian,
				behavior.Read("get", readBadOutput).Describe("Get")),
			validateFunc: func(byRule map[string][]string) {
				s.Contains(byRule[behavior.RuleOutputShape], "output field NoJSON has no json name")
				s.Contains(
					byRule[behavior.RuleOutputShape],
					"output field Arr is an array; use a slice",
				)
				s.Len(byRule[behavior.RuleOutputShape], 3)
			},
		},
		{
			name: "when the name is not lowercase letters and digits",
			b: behavior.Define("ip-forward", "x", behavior.Node, onDebian,
				behavior.Read("list", readList).Describe("List")),
			validateFunc: func(byRule map[string][]string) {
				s.Equal([]string{
					`name "ip-forward" must be lowercase letters and digits, starting with a letter`,
				}, byRule[behavior.RuleName])
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			rules := []string{}
			byRule := map[string][]string{}
			var messages []string
			for _, p := range behavior.ValidateBehavior(tc.b) {
				assert.NotEmpty(s.T(), p.Message)
				assert.NotEmpty(s.T(), p.Fix)
				rules = append(rules, p.Rule)
				byRule[p.Rule] = append(byRule[p.Rule], p.Message)
				if p.Rule == behavior.RuleSettings {
					messages = append(messages, p.Message)
				}
			}

			if tc.validateFunc != nil {
				tc.validateFunc(byRule)

				return
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

type secretKeyIn struct {
	Password secret.Value `json:"password" validate:"required" doc:"test"`
}

type badSecretIn struct {
	Key string `json:"key" validate:"required" doc:"The key" secret:"true"`
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

type badTelemetryOut struct {
	Key string `json:"key" doc:"The key" telemetry:"atr"`
}

type badSecretOut struct {
	Key string `json:"key" doc:"The key" secret:"true"`
}

type badBeforeIn struct {
	Name string `json:"name"           validate:"required"  doc:"The user" telemetry:"attr"`
	Hash string `json:"hash,omitempty" validate:"omitempty" doc:"The hash"                  before:"later"`
}

// ValidatePublicTestSuite drives the rules that the DSL audit added (osapi
// #973-#983) through ValidateBehavior, one input or output each.
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

func (s *ValidatePublicTestSuite) TestSchemaNameProblems() {
	define := func(name string) *behavior.Behavior {
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
	withOut := func(name string, out any) *behavior.Behavior {
		b := define(name)
		b.Methods[0].Out = reflect.TypeOf(out)

		return b
	}

	tests := []struct {
		name         string
		behaviors    []*behavior.Behavior
		validateFunc func(got []behavior.Problem)
	}{
		{
			name: "when two behaviors have different types of one name",
			behaviors: []*behavior.Behavior{
				withOut("memory", usageA()),
				withOut("again", usageA()),
				withOut("swap", usageB()),
			},
			validateFunc: func(got []behavior.Problem) {
				s.Require().Len(got, 1)
				s.Equal("swap", got[0].Behavior)
				s.Equal(behavior.RuleSchema, got[0].Rule)
				s.Equal(
					"Usage is also a type of memory; in the spec the two are one schema",
					got[0].Message,
				)
			},
		},
		{
			name: "when the same type is in two behaviors, it is one schema",
			behaviors: []*behavior.Behavior{
				withOut("memory", usageA()),
				withOut("again", usageA()),
			},
			validateFunc: func(got []behavior.Problem) {
				s.Empty(got)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.validateFunc(behavior.ExportSchemaNameProblems(tt.behaviors))
		})
	}
}

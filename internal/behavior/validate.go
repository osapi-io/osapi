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

package behavior

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// Problem is one rule a behavior breaks.
type Problem struct {
	Behavior string `json:"behavior"`
	Method   string `json:"method,omitempty"`
	Rule     string `json:"rule"`
	Message  string `json:"message"`
}

// The rules Validate checks, by the name a Problem reports.
const (
	RuleName        = "name"
	RulePlacement   = "placement"
	RuleDoc         = "doc"
	RuleMethods     = "methods"
	RuleDuplicate   = "duplicate"
	RuleKey         = "key"
	RuleSchema      = "schema"
	RuleChanged     = "changed"
	RuleOutputShape = "output"
	RuleSurface     = "surface"
	RuleExample     = "example"
	RuleVerb        = "verb"
	RuleRoute       = "route"
	// RuleSettings fails settings that are not a struct, or a field
	// without the mapstructure name or doc the config needs.
	RuleSettings = "settings"
	// RuleRequires: every behavior required is registered, and none
	// requires itself through others.
	RuleRequires = "requires"
	// RuleSecret: an input field tagged secret is not also recorded on the
	// method's span.
	RuleSecret = "secret"
	// RulePlatform: a behavior declares a provider for at least one
	// platform, once each, and each implements every method's provider type.
	RulePlatform = "platform"
)

// synonyms are method names that mean a standard verb, and the verb to use.
// A behavior's method names become its commands, so they follow the CLI's
// one-verb-per-meaning rule.
var synonyms = map[string]string{
	"remove":   "delete",
	"rm":       "delete",
	"destroy":  "delete",
	"add":      "create",
	"new":      "create",
	"describe": "get",
	"show":     "get",
	"ls":       "list",
}

// readVerbs are the names a read method may have.
var readVerbs = map[string]bool{"list": true, "get": true, "status": true}

// writeVerbs are the names a write method may have. A write named for an
// action, such as restart, would be served as a create at the collection's
// route, so it is refused until the platform has action methods.
var writeVerbs = map[string]bool{"create": true, "update": true, "delete": true}

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// reservedFields are what the platform puts in every per-host result, so an
// output type that declares one would collide with it.
var reservedFields = map[string]bool{
	"hostname": true,
	"status":   true,
	"error":    true,
}

// Validate checks every registered behavior against the rules the platform
// relies on, and returns what each breaks. No problems means every route,
// command and schema can be built.
func Validate() []Problem {
	all := All()
	problems := make([]Problem, 0, len(all))

	mu.RLock()
	for _, name := range duplicates {
		problems = append(problems, Problem{
			Behavior: name,
			Rule:     RuleDuplicate,
			Message:  "registered more than once",
		})
	}
	mu.RUnlock()

	for _, b := range all {
		problems = append(problems, ValidateBehavior(b)...)
	}

	return append(problems, chainProblems(all)...)
}

// ValidateBehavior checks one behavior.
func ValidateBehavior(
	b *Behavior,
) []Problem {
	var problems []Problem
	add := func(method, rule, format string, args ...any) {
		problems = append(problems, Problem{
			Behavior: b.Name,
			Method:   method,
			Rule:     rule,
			Message:  fmt.Sprintf(format, args...),
		})
	}

	if !namePattern.MatchString(b.Name) {
		add("", RuleName, "name %q must be lowercase letters, digits and dashes", b.Name)
	}
	if b.Group != "" && !namePattern.MatchString(b.Group) {
		add("", RuleName, "group %q must be lowercase letters, digits and dashes", b.Group)
	}
	if b.Placement == "" {
		add("", RulePlacement, "no placement; add behavior.Node")
	}
	if b.Doc == "" {
		add("", RuleDoc, "no description")
	}
	if b.CLI && !b.API {
		add("", RuleSurface, "CLI without API; the commands call the API, so add behavior.API")
	}
	if b.Internal && (b.API || b.CLI) {
		add("", RuleSurface, "internal, but served; an internal behavior has no API or CLI")
	}
	if b.CLI {
		for _, m := range b.Methods {
			for _, name := range unflaggable(m.In) {
				add(m.Name, RuleSurface,
					"input field %s cannot be a CLI flag; flags hold a string, bool, number or "+
						"float, so give it one of those types or leave CLI off", name)
			}
		}
	}
	for _, msg := range settingsProblems(b.SettingsType(), "") {
		add("", RuleSettings, "%s", msg)
	}
	if len(b.Methods) == 0 {
		add("", RuleMethods, "no methods")
	}
	if len(b.Platforms) == 0 {
		add("", RulePlatform, "no provider for any platform; add behavior.Platform")
	}
	platforms := map[string]bool{}
	for _, pp := range b.Platforms {
		if platforms[pp.Name] {
			add("", RulePlatform, "platform %q is declared twice", pp.Name)
		}
		platforms[pp.Name] = true
		for _, m := range b.Methods {
			if want := m.ProviderType(); !implements(pp.Type, want) {
				add(m.Name, RulePlatform, "the %s provider %s does not implement %s",
					pp.Name, pp.Type, want)
			}
		}
	}

	seen := map[string]bool{}
	routes := map[string]string{}
	for _, m := range b.Methods {
		// Two methods on one verb and path would be one route, and only one
		// of them could be served.
		route := m.HTTPMethod() + " " + b.Path(m)
		if other, ok := routes[route]; ok && other != m.Name {
			add(m.Name, RuleRoute, "%s is also %s's route; give one of them a key", route, other)
		}
		routes[route] = m.Name

		if seen[m.Name] {
			add(m.Name, RuleDuplicate, "method declared twice")
		}
		// A badly cased name is the name rule's problem, not this one's.
		verb := strings.ToLower(m.Name)
		if use, ok := synonyms[verb]; ok {
			add(m.Name, RuleVerb, "%q is a synonym; name the method %q", m.Name, use)
		} else if m.Verb == VerbRead && !readVerbs[verb] {
			add(m.Name, RuleVerb, "a read is named list, get or status, not %q", m.Name)
		} else if m.Verb == VerbWrite && !writeVerbs[verb] {
			add(m.Name, RuleVerb, "a write is named create, update or delete, not %q; "+
				"action methods such as restart are not supported yet", m.Name)
		}
		// The docs pages are the examples run through the real renderer, so
		// an example of the wrong type would publish output the method can
		// never produce.
		if b.CLI && len(m.Examples) == 0 {
			add(
				m.Name,
				RuleExample,
				"no example; a CLI behavior's docs page is built from its examples",
			)
		}
		for _, ex := range m.Examples {
			if ex.Doc == "" {
				add(m.Name, RuleExample, "an example has no description")
			}
			if got := reflect.TypeOf(ex.Input); got != m.In {
				add(m.Name, RuleExample, "example %q: input is %v, want %v", ex.Doc, got, m.In)
			}
			if got := reflect.TypeOf(ex.Output); got != m.Out {
				add(m.Name, RuleExample, "example %q: output is %v, want %v", ex.Doc, got, m.Out)
			}
		}
		if m.In.Kind() != reflect.Struct {
			add(
				m.Name,
				RuleSchema,
				"input is %v; an input is a struct, whose fields are the request's and the flags",
				m.In,
			)
		} else {
			for i := range m.In.NumField() {
				if f := m.In.Field(i); Secret(f) && f.Tag.Get("telemetry") == "attr" {
					add(m.Name, RuleSecret,
						"input field %s is secret and tagged telemetry:\"attr\"; a secret is never recorded on a span",
						JSONName(f))
				}
			}
		}

		for _, r := range m.Refusals {
			if r.Doc == "" || r.Error == "" {
				add(m.Name, RuleExample, "a refusal has no description or no error text")
			}
			if len(r.Effects) > 0 {
				add(
					m.Name,
					RuleExample,
					"refusal %q declares effects; a call that fails changes nothing to check",
					r.Doc,
				)
			}
			if got := reflect.TypeOf(r.Input); got != m.In {
				add(m.Name, RuleExample, "refusal %q: input is %v, want %v", r.Doc, got, m.In)
			}
		}
		seen[m.Name] = true

		if !namePattern.MatchString(m.Name) {
			add(m.Name, RuleName, "method name must be lowercase letters, digits and dashes")
		}
		if m.Doc == "" {
			add(m.Name, RuleDoc, "no description; add .Describe(...)")
		}

		in, err := SchemaFor(m.In)
		if err != nil {
			add(m.Name, RuleSchema, "input: %v", err)
		} else if m.KeyField != "" {
			if _, ok := in.Properties[m.KeyField]; !ok {
				add(m.Name, RuleKey, "key %q is not a field of the input", m.KeyField)
			}
		}

		out, err := SchemaFor(m.Out)
		if err != nil {
			add(m.Name, RuleSchema, "output: %v", err)

			continue
		}

		elem := m.Out
		for elem.Kind() == reflect.Pointer {
			elem = elem.Elem()
		}
		if elem.Kind() == reflect.Slice {
			elem = elem.Elem()
			out = out.Items
		}
		if elem.Kind() != reflect.Struct {
			add(m.Name, RuleOutputShape,
				"output must be a struct or a slice of structs, got %s", elem.Kind())

			continue
		}
		for name := range out.Properties {
			if reservedFields[name] {
				add(
					m.Name,
					RuleOutputShape,
					"output field %q collides with the per-host result",
					name,
				)
			}
		}

		changed, ok := out.Properties["changed"]
		if m.Verb == VerbWrite && (!ok || changed.Type != "boolean") {
			add(m.Name, RuleChanged, `a write's output needs Changed bool `+"`json:\"changed\"`")
		}
	}

	return problems
}

// settingsProblems checks a settings struct: every exported field has a
// mapstructure name and a doc, walking into nested structs. A nil type is a
// behavior with no settings, which is fine.
func settingsProblems(
	t reflect.Type,
	prefix string,
) []string {
	if t == nil {
		return nil
	}
	if t.Kind() != reflect.Struct {
		return []string{fmt.Sprintf("settings must be a struct, not %s", t.Kind())}
	}

	var out []string
	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		name, _, _ := strings.Cut(f.Tag.Get("mapstructure"), ",")
		if name == "" {
			out = append(
				out,
				fmt.Sprintf("settings field %s%s has no mapstructure name", prefix, f.Name),
			)

			continue
		}
		if f.Type.Kind() == reflect.Struct {
			out = append(out, settingsProblems(f.Type, prefix+name+".")...)

			continue
		}
		if f.Tag.Get("doc") == "" {
			out = append(out, fmt.Sprintf("setting %s%s has no doc", prefix, name))
		}
	}

	return out
}

// implements reports whether a provider of type have can be handed to a
// method whose provider type is want.
func implements(
	have reflect.Type,
	want reflect.Type,
) bool {
	return have == want || (want.Kind() == reflect.Interface && have.Implements(want))
}

// unflaggable are the JSON names of in's fields that no CLI flag can hold.
// It reads fields the way the CLI does: one without a json name is not a
// flag at all.
func unflaggable(
	in reflect.Type,
) []string {
	if in.Kind() != reflect.Struct {
		return nil
	}

	var names []string
	for i := range in.NumField() {
		f := in.Field(i)
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if !f.IsExported() || name == "" || name == "-" {
			continue
		}
		switch f.Type.Kind() {
		case reflect.String, reflect.Bool,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
		default:
			names = append(names, name)
		}
	}

	return names
}

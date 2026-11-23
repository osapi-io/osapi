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
	"slices"
	"strings"

	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/internal/validation"

	"github.com/osapi-io/osapi/internal/validation/rule"
)

// Problem is one rule a behavior breaks.
type Problem struct {
	Behavior string `json:"behavior"`
	Method   string `json:"method,omitempty"`
	Rule     string `json:"rule"`
	Message  string `json:"message"`
	// Fix is how to fix it, from the check of Rule.
	Fix string `json:"fix" table:"-"`
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
	// RuleParent: a behavior declared with Of is under a registered
	// behavior that has a key, and does not also declare Under.
	RuleParent = "parent"
	// RuleSettings fails settings that are not a struct, or a field
	// without the mapstructure name or doc the config needs.
	RuleSettings = "settings"
	// RuleRequires: every behavior required is registered, and none
	// requires itself through others.
	RuleRequires = "requires"
	// RuleSecret: an input field tagged secret is not also recorded on the
	// method's span.
	RuleSecret = "secret"
	// RuleValidate: each input field has a validate tag. The endpoint DSL
	// holds its inputs to the same check, from package rule.
	RuleValidate = "validate"
	// RuleFails: each rule of an input field has a .Fails input that breaks
	// it. The endpoint DSL holds its inputs to the same check.
	RuleFails = "fails"
	// RuleOptional: an optional field of a write's request body is left out
	// of the JSON when it is empty, so the client does not send "" for it.
	RuleOptional = "optional"
	// RuleInput: each input field has a json name, a doc and no default,
	// and is not named hostname or target, which every request has.
	RuleInput = "input"
	// RulePlatform: a behavior declares a provider for at least one
	// platform, once each, and each implements every method's provider type.
	RulePlatform = "platform"
	// RuleInternal: an internal behavior is required by another. Nothing
	// else can reach it, because it has no API and no CLI.
	RuleInternal = "internal"
	// RuleWrites: each path a behavior declares with Writes is "*", or an
	// absolute path.Match pattern.
	RuleWrites = "writes"
	// RuleSpan: a write records on its span what it acts on, through Key or
	// a field tagged telemetry:"attr". Thus a trace says what each write
	// changed. A write whose input has no field that can be recorded is
	// exempt.
	//
	// Checks has the requirement and the fix of each rule.
	RuleSpan = "span"
	// RuleTemplate: a field tagged template:"<vars>" is a string, and <vars>
	// names an input field of type map[string]string.
	RuleTemplate = "template"
	// RuleTag: each struct tag of an input field has a known key
	// (command.KnownTags), and secret is "true" and telemetry is "attr" on an
	// input or an output field. A typo of either drops what
	// the tag does with no error: a secret reaches the audit log.
	RuleTag = "tag"
	// RuleBefore: a .Before step is func(context.Context, In) (In, error),
	// on a served method with a secret field, and each field tagged before
	// is set by one.
	RuleBefore = "before"
)

// tagProblems are the RuleTag problems of t's fields, and of the fields of
// each struct that a field holds, named from prefix. keys checks each tag
// key against command.KnownTags too; an output's keys are checked by
// command.TagProblems, under the output's rule.
func tagProblems(
	t reflect.Type,
	prefix string,
	keys bool,
	seen map[reflect.Type]bool,
) []string {
	for t != nil && (t.Kind() == reflect.Pointer || t.Kind() == reflect.Slice ||
		t.Kind() == reflect.Array || t.Kind() == reflect.Map) {
		t = t.Elem()
	}
	if t == nil || t.Kind() != reflect.Struct || seen[t] {
		return nil
	}
	seen[t] = true
	var out []string
	for i := range t.NumField() {
		f := t.Field(i)
		name := prefix + f.Name
		for _, key := range command.TagKeys(f.Tag) {
			if keys && !slices.Contains(command.KnownTags, key) {
				out = append(
					out,
					fmt.Sprintf("field %s has the tag key %q, which nothing reads", name, key),
				)
			}
		}
		if v, ok := f.Tag.Lookup("secret"); ok && v != "true" {
			out = append(
				out,
				fmt.Sprintf("field %s is tagged secret:%q; the only value is \"true\"", name, v),
			)
		}
		if v, ok := f.Tag.Lookup("telemetry"); ok && v != "attr" {
			out = append(
				out,
				fmt.Sprintf("field %s is tagged telemetry:%q; the only value is \"attr\"", name, v),
			)
		}
		out = append(out, tagProblems(f.Type, name+".", keys, seen)...)
	}

	return out
}

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

// writeVerbs are the names a write method may have. A write named for what
// it does, such as restart, is an action: Action declares it, and its name is
// one of the CLI's actions.
var writeVerbs = map[string]bool{"create": true, "update": true, "delete": true}

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9]*$`)

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

	problems = append(problems, schemaNameProblems(all)...)

	return withFixes(append(problems, chainProblems(all)...))
}

// schemaNameProblems is each schema name that two Go types would have. A
// method's input and output are named with the behavior's TypeName, but a
// struct they hold is named by its own name, in one spec for every
// behavior: a second Usage would replace the first.
func schemaNameProblems(
	all []*Behavior,
) []Problem {
	owner := map[string]reflect.Type{}
	where := map[string]string{}
	var problems []Problem
	for _, b := range all {
		for _, m := range b.Methods {
			for _, t := range append(nested(m.In), nested(m.Out)...) {
				prev, ok := owner[t.Name()]
				switch {
				case !ok:
					owner[t.Name()], where[t.Name()] = t, b.Name
				case prev != t:
					problems = append(problems, Problem{
						Behavior: b.Name,
						Method:   m.Name,
						Rule:     RuleSchema,
						Message: fmt.Sprintf(
							"%s is also a type of %s; in the spec the two are one schema",
							t.Name(),
							where[t.Name()],
						),
					})
				}
			}
		}
	}

	return problems
}

// nested is each named struct that t's fields hold, at any depth, but not
// t itself: those keep their own names in the spec.
func nested(
	t reflect.Type,
) []reflect.Type {
	var out []reflect.Type
	seen := map[reflect.Type]bool{}
	var walk func(t reflect.Type, top bool)
	walk = func(t reflect.Type, top bool) {
		for t.Kind() == reflect.Pointer || t.Kind() == reflect.Slice || t.Kind() == reflect.Map {
			t = t.Elem()
		}
		if t.Kind() != reflect.Struct || seen[t] {
			return
		}
		seen[t] = true
		if !top && t.Name() != "" {
			out = append(out, t)
		}
		for i := range t.NumField() {
			if t.Field(i).IsExported() {
				walk(t.Field(i).Type, false)
			}
		}
	}
	walk(t, true)

	return out
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
		add(
			"",
			RuleName,
			"name %q must be lowercase letters and digits, starting with a letter",
			b.Name,
		)
	}
	if b.Name == "all" || b.Group == "all" {
		add(
			"",
			RuleName,
			"all is internal/behaviors/all, which registers every behavior; choose another name",
		)
	}
	problems = append(problems, keyNameProblems(b)...)
	if b.Parent != nil {
		switch {
		case b.Group != "":
			add(
				"",
				RuleParent,
				"it declares Of(%s) and Under(%q); a child is under its parent only",
				b.Parent.Name,
				b.Group,
			)
		case b.ParentKey() == "":
			add(
				"",
				RuleParent,
				"its parent %s has no keyed method, so its items have no key to be under",
				b.Parent.Name,
			)
		case Lookup(b.Parent.Name) != b.Parent:
			add("", RuleParent, "its parent %s is not registered", b.Parent.Name)
		}
	}
	if b.Group != "" && !namePattern.MatchString(b.Group) {
		add(
			"",
			RuleName,
			"group %q must be lowercase letters and digits, starting with a letter",
			b.Group,
		)
	}
	if b.Group != "" && b.Group != "all" && namePattern.MatchString(b.Group) &&
		!slices.Contains(KnownGroups(), GroupName(b.Group)) {
		add(
			"",
			RuleName,
			"group %q is not one of KnownGroups; use a behavior.Group constant, such as GroupNetwork",
			b.Group,
		)
	}
	if b.Placement == "" {
		add("", RulePlacement, "no placement; add behavior.Node")
	}
	if b.Doc == "" {
		add("", RuleDoc, "no description")
	}
	if markup(b.Doc) {
		add(
			"",
			RuleDoc,
			"description %q has < > { or }, which the docs site reads as markup",
			b.Doc,
		)
	}
	if b.API && !area.Valid(b.Area) {
		add(
			"",
			RuleSurface,
			"served, but area %q is not one of %s; add behavior.Area(...), so the UI can place its blocks",
			b.Area,
			area.List(),
		)
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
					"input field %s cannot be a CLI flag; a flag holds a string, bool, int, "+
						"[]string, map[string]string or []byte (a file), so give it one of those "+
						"types or leave CLI off", name)
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
	// unknown reports a platform name that no host reports: a declaration
	// with it would never apply, and a provider or a program would be
	// skipped with no error.
	unknown := func(method, what string, p PlatformName) bool {
		if slices.Contains(KnownPlatforms(), p) {
			return false
		}
		add(
			method,
			RulePlatform,
			"%s names %q, which is not a platform; use a behavior.Platform constant (%s)",
			what,
			p,
			knownPlatformList(),
		)

		return true
	}
	platforms := map[PlatformName]bool{}
	for _, pp := range b.Platforms {
		unknown("", "Platform", pp.Name)
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

	for _, pr := range b.PlatformRuns {
		unknown("", "RunsOnPlatform", pr.Platform)
	}
	for _, ps := range b.ProviderSudo {
		unknown("", "SudoOnPlatform", ps.Platform)
	}
	for _, m := range b.Methods {
		for _, e := range m.Examples {
			for _, p := range e.Platforms {
				unknown(m.Name, "OnPlatform", p)
			}
		}
		for _, r := range m.Refusals {
			for _, p := range r.Platforms {
				unknown(m.Name, "OnPlatform", p)
			}
		}
		for _, p := range m.Platforms {
			if unknown(m.Name, "On", p) {
				continue
			}
			if b.resolve(p) == nil {
				add(
					m.Name,
					RulePlatform,
					"On names %q, but the behavior declares no provider for it; add behavior.Platform",
					p,
				)
			}
		}
	}

	seen := map[string]bool{}
	routes := map[string]string{}
	for _, m := range b.Methods {
		for _, msg := range beforeProblems(b, m) {
			add(m.Name, RuleBefore, "%s", msg)
		}
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
		} else if m.Action && (!command.Actions[verb] || command.Verbs[verb] || readVerbs[verb]) {
			add(m.Name, RuleVerb, "an action is named for what it does, with one of the CLI's "+
				"actions (osapi cli command spec get), not %q", m.Name)
		} else if m.Verb == VerbRead && !readVerbs[verb] {
			add(m.Name, RuleVerb, "a read is named list, get or status, not %q", m.Name)
		} else if m.Verb == VerbWrite && !m.Action && !writeVerbs[verb] {
			add(m.Name, RuleVerb, "a write is named create, update or delete, not %q; "+
				"declare a write named for what it does, such as restart, with behavior.Action", m.Name)
		}
		if m.Disruptive && strings.TrimSpace(m.Disruption) == "" {
			add(m.Name, RuleVerb, ".Disrupts has no reason; say why a run takes the host down")
		}
		if m.AlwaysChanges && !m.Action {
			add(m.Name, RuleVerb, ".Always is for an action, such as restart; a create, update or "+
				"delete compares the host with its input and changes nothing when they agree")
		}
		if m.PermitVerb != "" && !m.Action {
			add(
				m.Name,
				RuleVerb,
				".PermitAs and .Shell are for an action, such as docker exec or command shell; "+
					"a read or a write keeps the permission of its verb",
			)
		}
		if m.PermitVerb != "" && !slices.Contains(adminVerbs, m.PermitVerb) {
			add(m.Name, RuleVerb, ".PermitAs takes a verb that only the admin role gets, "+
				"behavior.VerbShell or behavior.VerbExec, not %q", m.PermitVerb)
		}
		// The examples are the command lines on the CLI page of the method,
		// and the generated tests send each one. The tests also check each
		// output, so an example of the wrong type fails.
		if b.CLI && len(m.Examples) == 0 {
			add(
				m.Name,
				RuleExample,
				"no example; the CLI page of a CLI behavior shows its examples",
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
			for _, f := range rule.Untagged(m.In) {
				add(m.Name, RuleValidate, "input field %s has no validate tag", f)
			}
			for _, n := range unrefused(m) {
				add(m.Name, RuleFails, "no .Fails input breaks %s on %s", n.Rule, n.Field)
			}
			if b.API && m.Verb == VerbWrite {
				for _, f := range sentEmpty(m) {
					add(m.Name, RuleOptional,
						"optional field %s is sent as its zero value when it is empty", f)
				}
			}
			for _, msg := range tagProblems(m.In, "", true, map[reflect.Type]bool{}) {
				add(m.Name, RuleTag, "input %s", msg)
			}
			for _, msg := range tagProblems(m.Out, "", false, map[reflect.Type]bool{}) {
				add(m.Name, RuleTag, "output %s", msg)
			}
			recordable := false
			for i := range m.In.NumField() {
				f := m.In.Field(i)
				if Secret(f) && f.Tag.Get("telemetry") == "attr" {
					add(
						m.Name,
						RuleSecret,
						"input field %s is secret and tagged telemetry:\"attr\"; a secret is never recorded on a span",
						JSONName(f),
					)
				}
				if f.IsExported() && !Secret(f) {
					recordable = true
				}
			}
			for _, tf := range TemplateFields(m.In) {
				f := m.In.FieldByIndex(tf.index)
				switch {
				case f.Type.Kind() != reflect.String:
					add(
						m.Name,
						RuleTemplate,
						"input field %s is tagged template but is %s, not a string",
						tf.Field,
						f.Type,
					)
				case tf.vars == nil:
					add(
						m.Name,
						RuleTemplate,
						"input field %s names %q for its vars, which the input does not have",
						tf.Field,
						tf.Vars,
					)
				case m.In.FieldByIndex(tf.vars).Type != reflect.TypeFor[map[string]string]():
					add(
						m.Name,
						RuleTemplate,
						"input field %s is the vars of %s but is %s, not map[string]string",
						tf.Vars,
						tf.Field,
						m.In.FieldByIndex(tf.vars).Type,
					)
				}
			}
			if m.Verb == VerbWrite && recordable && len(m.SpanInputs()) == 0 {
				add(
					m.Name,
					RuleSpan,
					"its span records no input, so a trace cannot say what it changed",
				)
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
			add(
				m.Name,
				RuleName,
				"method name must be lowercase letters and digits, starting with a letter",
			)
		}
		if m.Doc == "" {
			add(m.Name, RuleDoc, "no description; add .Describe(...)")
		}
		for _, text := range []string{m.Doc, m.Long} {
			if markup(text) {
				add(
					m.Name,
					RuleDoc,
					"description %q has < > { or }, which the docs site reads as markup",
					text,
				)
			}
		}

		in, err := SchemaFor(m.In)
		if err != nil {
			add(m.Name, RuleSchema, "input: %v", err)
		} else {
			for _, key := range m.PathKeys() {
				if _, ok := in.Properties[key]; !ok {
					add(m.Name, RuleKey, "key %q is not a field of the input", key)
				} else if !slices.Contains(in.Required, key) {
					add(
						m.Name,
						RuleKey,
						"key %q is optional, but a key is a path segment, which is never empty",
						key,
					)
				}
			}
		}
		for _, msg := range inputProblems(m.In) {
			add(m.Name, RuleInput, "%s", msg)
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

		for _, msg := range outputProblems(elem) {
			add(m.Name, RuleOutputShape, "%s", msg)
		}

		changed, ok := out.Properties["changed"]
		if m.Verb == VerbWrite && (!ok || changed.Type != "boolean") {
			add(m.Name, RuleChanged, `a write's output needs Changed bool `+"`json:\"changed\"`")
		}
	}

	return withFixes(problems)
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
		// The CLI owns which types a flag holds; its generator builds each.
		if !command.Flaggable(f.Type) {
			names = append(names, name)
		}
	}

	return names
}

// unrefused is each rule of m's input that no .Fails input breaks.
func unrefused(
	m *Method,
) []rule.Need {
	// An input whose spec does not build has a tag the validator may not
	// know; RuleSchema reports it.
	if _, err := SchemaFor(m.In); err != nil {
		return nil
	}
	broken := map[string]bool{}
	for _, r := range m.Refusals {
		for _, b := range validation.Breaks(r.Input) {
			broken[b] = true
		}
	}

	var out []rule.Need
	for _, n := range rule.Needs(m.In, false) {
		if !broken[n.Key()] {
			out = append(out, n)
		}
	}

	return out
}

// sentEmpty is each field of m's request body that a client sends empty:
// the endpoint layer owns the rule, and the body is every input field but
// the key, which is in the path.
func sentEmpty(
	m *Method,
) []string {
	var body []reflect.StructField
	for i := range m.In.NumField() {
		f := m.In.Field(i)
		if name, _, _ := strings.Cut(f.Tag.Get("json"), ","); name != m.KeyField {
			body = append(body, f)
		}
	}

	return endpoint.SentEmpty(body)
}

// reservedInputs are what the generated request of every method has: the
// target host is the hostname path segment and the --target flag. An input
// field of the same name would be a second of each.
var reservedInputs = map[string]bool{"hostname": true, "target": true}

// inputProblems is each input field that the generators cannot build or
// that would mislead: one with no json name, which the CLI and the tables
// drop; one named like the target; one with no doc, which the spec, the
// flag and the page show empty; one with a default, which the server does
// not apply; and an array, which has no JSON shape of its own.
func inputProblems(
	in reflect.Type,
) []string {
	if in.Kind() != reflect.Struct {
		return nil
	}

	var out []string
	for i := range in.NumField() {
		f := in.Field(i)
		if !f.IsExported() {
			continue
		}
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		switch name {
		case "-":
			continue
		case "":
			out = append(out, fmt.Sprintf("input field %s has no json name", f.Name))

			continue
		}
		if reservedInputs[name] || f.Name == "Hostname" {
			out = append(out, fmt.Sprintf(
				"input field %s is named like the target, which every request has", name))
		}
		if f.Tag.Get("doc") == "" {
			out = append(out, fmt.Sprintf("input field %s has no doc", name))
		}
		if _, ok := f.Tag.Lookup("default"); ok {
			out = append(out, fmt.Sprintf(
				"input field %s has a default, which the server does not apply", name))
		}
		if hasArray(f.Type) {
			out = append(out, fmt.Sprintf("input field %s is an array; use a slice", name))
		}
	}

	return out
}

// outputProblems is each output field with no json name, which the tables
// drop, or an array, and each render tag that the output does not know.
func outputProblems(
	out reflect.Type,
) []string {
	var problems []string
	for i := range out.NumField() {
		f := out.Field(i)
		if !f.IsExported() {
			continue
		}
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if name == "" {
			problems = append(problems, fmt.Sprintf("output field %s has no json name", f.Name))
		}
		if hasArray(f.Type) {
			problems = append(
				problems,
				fmt.Sprintf("output field %s is an array; use a slice", f.Name),
			)
		}
	}

	return append(problems, command.TagProblems(out)...)
}

// hasArray reports whether t is a fixed-size array, or a slice, map or
// pointer of one: the generators write each as a slice or a pointer.
func hasArray(
	t reflect.Type,
) bool {
	for {
		switch t.Kind() {
		case reflect.Array:
			return true
		case reflect.Slice, reflect.Pointer, reflect.Map:
			t = t.Elem()
		default:
			return false
		}
	}
}

// markup reports whether text has a character that the docs site's MDX reads
// as a tag or an expression, so the page fails to build.
func markup(
	text string,
) bool {
	return strings.ContainsAny(text, "<>{}")
}

// keyNameProblems is each path key whose name in the spec is the name of a
// row type of the behavior: both are the TypeName and the title, so the key
// part of sshkey and a row type Part are both SshkeyPart, and the spec
// cannot have the two.
func keyNameProblems(
	b *Behavior,
) []Problem {
	rows := map[string]bool{}
	for _, m := range b.Methods {
		t := m.Out
		for t != nil && (t.Kind() == reflect.Pointer || t.Kind() == reflect.Slice) {
			t = t.Elem()
		}
		if t != nil {
			rows[t.Name()] = true
		}
	}
	var problems []Problem
	seen := map[string]bool{}
	for _, m := range b.Methods {
		for _, key := range m.PathKeys() {
			if rows[title(key)] && !seen[key] {
				seen[key] = true
				problems = append(problems, Problem{
					Behavior: b.Name, Method: m.Name, Rule: RuleKey,
					Message: fmt.Sprintf("key %q and the type %s are both %s%s in the spec; "+
						"rename the type, such as %sEntry", key, title(key), b.TypeName(), title(key), title(key)),
				})
			}
		}
	}

	return problems
}

// knownPlatformList is KnownPlatforms, for a message.
func knownPlatformList() string {
	names := make([]string, 0, len(KnownPlatforms()))
	for _, p := range KnownPlatforms() {
		names = append(names, string(p))
	}

	return strings.Join(names, ", ")
}

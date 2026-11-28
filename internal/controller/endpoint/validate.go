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

package endpoint

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"unicode"

	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/cli/command"

	"github.com/osapi-io/osapi/internal/validation"

	"github.com/osapi-io/osapi/internal/validation/rule"

	"github.com/osapi-io/osapi/internal/guide"
	"github.com/osapi-io/osapi/internal/secret"
)

// name is a group, a method, a sub-path or an action: lowercase letters and
// digits, starting with a letter.
var name = regexp.MustCompile(`^[a-z][a-z0-9]*$`)

// The rules Validate checks, by the name a Problem reports.
const (
	RuleName    = "name"
	RuleArea    = "area"
	RuleMethods = "methods"
	RuleUnique  = "unique"
	RuleSegment = "segment"
	RuleAction  = "action"
	RuleKey     = "key"
	RuleID      = "id"
	RuleRoute   = "route"
	RuleSchema  = "schema"
	RuleHandler = "handler"
	RuleSubDoc  = "subdoc"
	// RuleValidate is the same check in the behavior DSL, from package rule.
	RuleValidate = "validate"
	// RuleFails is the same check in the behavior DSL, from package rule.
	RuleFails = "fails"
	// RuleOptional: an optional field of a request body is left out of the
	// JSON when it is empty, so a client does not send "" for it.
	RuleOptional = "optional"
	// RuleIn: each in tag of an input field is path, query, form or body.
	// Any other value falls into the request body with no error.
	RuleIn = "in"
	// RuleField: a field that Location or Example names is a json field of
	// its type. A misspelled name gave a Location of <nil>, or dropped the
	// example from the spec, with no error.
	RuleField = "field"
	// RuleReason: each error that a method maps has a valid types.Reason. Go
	// converts an untyped literal to types.Reason, so Because("not_manged")
	// compiles, and the caller would get a word the schema does not list.
	RuleReason = "reason"
	// RuleExample: each flag that an Example names is a flag of the
	// method's command. A misspelled flag showed a command on its docs page
	// that fails.
	RuleExample = "example"
	// RulePrints: the output that Prints gives an example is a value of the
	// method's output type, and Prints follows an Example. Prints takes any
	// value, so a sample of another type compiles.
	RulePrints = "prints"
	// RuleSecret: a secret.Value input field is not in the path or the
	// query, which logs keep with the URL, and is not recorded on a span.
	RuleSecret = "secret"
	// RuleStutter: a type of the group's package does not repeat the name
	// of its package, as agent.AgentInfo does. Go reads it as agent.Info.
	RuleStutter = "stutter"
)

// inValues are the values of the in tag: where a field of a request comes
// from.
var inValues = []string{"path", "query", "form", "body"}

// hasField reports whether t, or the type t points to, has a field with the
// json name name.
func hasField(
	t reflect.Type,
	name string,
) bool {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	for _, f := range fields(t) {
		if jsonName(f) == name {
			return true
		}
	}

	return false
}

// Checks are the rules Validate applies. osapi dsl endpoint
// validate --help prints each one, and each problem carries the fix of its
// rule, so an author or an agent can read the rules before breaking one.
// The tests trigger each rule, and fail on a rule without a check.
var Checks = guide.Checks{
	{
		Rule:     RuleStutter,
		Requires: "No type of the group's package that a method uses repeats the name of its package, such as agent.AgentInfo: Go code reads it as agent.Info. revive's exported rule refuses it too.",
		Fix:      "Drop the package's name from the type, and keep its API name with a schema tag on its blank field: type Info struct { _ struct{} `schema:\"AgentInfo\"` }. Never write //nolint:revive.",
	},
	{
		Rule:     RuleSecret,
		Requires: "No secret.Value input field is tagged in:\"path\" or in:\"query\", or telemetry:\"attr\": a URL goes in logs, and a span is no place for a secret.",
		Fix:      "Send the secret in the body: drop its in tag, or tag it in:\"body\" or in:\"form\". Remove telemetry:\"attr\" from it.",
	},
	{
		Rule:     RuleExample,
		Requires: "Each flag that an Example names is a flag of the method's command: a field of its input, a field of the struct that in:\"body\" names, or --json, --debug or --osapi-file.",
		Fix:      "Spell each flag of the example as the command has it: osapi dsl endpoint get --name <group> lists the methods, and the command's --help lists its flags.",
	},
	{
		Rule:     RulePrints,
		Requires: "The output that Prints gives an example is a value of the method's output type, and Prints follows an Example.",
		Fix:      "Give Prints a value of the method's output type, after the Example that it shows.",
	},
	{
		Rule:     RuleIn,
		Requires: "Each in tag of an input field is path, query, form or body.",
		Fix:      "Correct the value. A field with another value is read from the request body, not from where the tag says.",
	},
	{
		Rule:     RuleReason,
		Requires: "Each error that a method maps has a reason that types.Reason lists.",
		Fix:      "Use a types.Reason constant, such as endpoint.ReasonNotManaged, in Because.",
	},
	{
		Rule:     RuleField,
		Requires: "The field that Location names is a json field of the method's output, and the field that Example[T] names is a json field of T.",
		Fix:      "Correct the name to the json name of the field, such as job_id.",
	},
	{
		Rule:     RuleArea,
		Requires: "The group has an area, one of area.All, so the UI, the SDK's pages and the docs site can place its operations.",
		Fix:      "Add endpoint.Area(...) to the group's definition with a constant of package area, such as area.Fleet. Add a new area to area.All with its label first.",
	},
	{
		Rule:     RuleName,
		Requires: "The group and each method have a name of lowercase letters and digits, starting with a letter.",
		Fix:      "Name the group as its path segment, such as agent, and each method for what it does, such as list or drain.",
	},
	{
		Rule:     RuleMethods,
		Requires: "The group has at least one method.",
		Fix:      "Add a method with endpoint.Get, Post, Put or Delete.",
	},
	{
		Rule:     RuleUnique,
		Requires: "No two methods of a group have one name.",
		Fix:      "Give each method its own name.",
	},
	{
		Rule:     RuleSegment,
		Requires: "Each Sub and Action segment is lowercase letters and digits.",
		Fix:      "Use one word, such as pending or drain.",
	},
	{
		Rule:     RuleAction,
		Requires: "A method with an Action has a key: an input field tagged in:\"path\".",
		Fix:      "Tag the field the action is on in:\"path\", or use Sub for a fixed path.",
	},
	{
		Rule:     RuleKey,
		Requires: "An input has at most one field tagged in:\"path\" besides the fields of the Prefix, or .Key names each of them in the order of the path. Each field that .Key or the Prefix names is an in:\"path\" field of the input, is not a secret, and is named once.",
		Fix:      "Tag one field in:\"path\". For a key with more than one field, such as a key in a vault, name each in order with .Key(\"vault\", \"key\").",
	},
	{
		Rule:     RuleID,
		Requires: "Each operationId is used once, and starts with a capital letter, as each operationId of the API does.",
		Fix:      "Give the method its own ID with .ID, in Pascal case, such as GetAgents.",
	},
	{
		Rule:     RuleRoute,
		Requires: "No two methods have one verb and path.",
		Fix:      "Change the verb, the Sub or the Action.",
	},
	{
		Rule:     RuleHandler,
		Requires: "A group that Uses a type gives each method a handler, func(context.Context, D, In) (Out, error), with D that type and In and Out the method's types. A group with a handler Uses a type.",
		Fix:      "Add .Handle(fn) to each method, with the signature that the message names, and endpoint.Uses[Deps]() to Define.",
	},
	{
		Rule:     RuleSubDoc,
		Requires: "Each Sub segment that makes a command of its own, such as stale in osapi client file stale list, has a SubDoc: the one line of help of that command.",
		Fix:      "Add endpoint.SubDoc(\"segment\", \"...\") to Define, with what the segment's commands do, such as Read the deployments whose file changed since.",
	},
	{
		Rule:     RuleValidate,
		Requires: rule.Requires,
		Fix:      rule.Fix,
	},
	{
		Rule:     RuleFails,
		Requires: rule.FailsRequires,
		Fix:      rule.FailsFix,
	},
	{
		Rule:     RuleOptional,
		Requires: "An optional field of a write's request body, with omitempty in its validate tag, has omitempty in its json tag too, so a client leaves it out when it is empty instead of sending \"\" or 0, which its other rules can refuse.",
		Fix:      "Add ,omitempty to the field's json tag, such as json:\"reason,omitempty\".",
	},
	{
		Rule:     RuleSchema,
		Requires: "The spec builds: each input and output type is one that JSON can encode, and each validate rule has a schema.",
		Fix:      "Use structs, slices, maps, strings, numbers, bools and time.Time; not channels or functions. Register the schema of a custom validate rule with rule.Register, beside its validator in internal/validation.",
	},
}

// Problem is a rule a definition breaks, and how to fix it.
type Problem struct {
	// Group is the group it is about.
	Group string `json:"group"`
	// Method is the method it is about, or "" for the group.
	Method string `json:"method,omitempty"`
	// Rule is the name of the rule, from Checks.
	Rule string `json:"rule"`
	// Message says what is wrong.
	Message string `json:"message"`
	// Fix is how to fix it, from the check of Rule.
	Fix string `json:"fix" table:"-"`
}

// String is the problem as one line: its message, then its fix.
func (p Problem) String() string {
	return p.Message + " Fix: " + p.Fix
}

// Validate is each rule g breaks. A group with no problems makes a spec
// that loads and routes that do not collide.
func Validate(
	g *Group,
) []Problem {
	var out []Problem
	add := func(method, rule, message string) {
		out = append(out, Problem{
			Group: g.Name, Method: method, Rule: rule, Message: message, Fix: Checks.Fix(rule),
		})
	}

	if !name.MatchString(g.Name) {
		add("", RuleName, fmt.Sprintf("group name %q is not lowercase letters and digits.", g.Name))
	}
	if len(g.Methods) == 0 {
		add("", RuleMethods, "the group has no methods.")
	}
	if !area.Valid(g.Area) {
		add("", RuleArea,
			fmt.Sprintf("area %q is not one of %s.", g.Area, area.List()))
	}
	var named []string
	for t, examples := range g.Examples {
		for field := range examples {
			if !hasField(t, field) {
				named = append(named, fmt.Sprintf(
					"Example names %s, which is not a json field of %s.", field, t))
			}
		}
	}
	slices.Sort(named)
	for _, msg := range named {
		add("", RuleField, msg)
	}

	for _, msg := range stutters(g) {
		add("", RuleStutter, msg)
	}

	names := map[string]bool{}
	ids := map[string]string{}
	routes := map[string]string{}
	for _, m := range g.Methods {
		if !name.MatchString(m.Name) {
			add(m.Name, RuleName,
				fmt.Sprintf("method name %q is not lowercase letters and digits.", m.Name))
		}
		if names[m.Name] {
			add(m.Name, RuleUnique, "two methods have the name "+m.Name+".")
		}
		names[m.Name] = true

		for _, seg := range []string{m.sub, m.action, m.cmdSub} {
			if seg != "" && !name.MatchString(seg) {
				add(m.Name, RuleSegment,
					fmt.Sprintf("path segment %q is not lowercase letters and digits.", seg))
			}
		}
		if m.action != "" && len(m.Keys()) == 0 {
			add(m.Name, RuleAction, "the method has an action but no key.")
		}
		if seg := m.SubPath(); seg != "" && seg != m.CommandName() && g.SubDocs[seg] == "" {
			add(m.Name, RuleSubDoc, fmt.Sprintf("the Sub segment %q has no SubDoc.", seg))
		}

		for _, f := range fields(m.In) {
			if v, ok := f.Tag.Lookup("in"); ok && !slices.Contains(inValues, v) {
				add(m.Name, RuleIn, fmt.Sprintf(
					"input field %s is tagged in:%q; the values are path, query, form and body.",
					jsonName(f), v))
			}
			if !secret.Is(f.Type) {
				continue
			}
			if in := f.Tag.Get("in"); in == "path" || in == "query" {
				add(m.Name, RuleSecret, fmt.Sprintf(
					"input field %s is a secret in the %s; a URL goes in logs.", jsonName(f), in))
			}
			if f.Tag.Get("telemetry") == "attr" {
				add(m.Name, RuleSecret, fmt.Sprintf(
					"input field %s is a secret and tagged telemetry:\"attr\"; a secret is never recorded on a span.",
					jsonName(f),
				))
			}
		}
		if m.unprinted {
			add(
				m.Name,
				RulePrints,
				"Prints comes before any Example; it gives the last Example its output.",
			)
		}
		if m.unchecked {
			add(
				m.Name,
				RulePrints,
				"Check comes before any Example; it gives the last Example its read.",
			)
		}
		for _, ex := range m.examples {
			if ex.Output == nil {
				continue
			}
			if t := reflect.TypeOf(ex.Output); t != m.Out {
				add(m.Name, RulePrints, fmt.Sprintf(
					"Prints gives the example %q a %s, not a %s.", ex.Flags, t, m.Out))
			}
		}
		for _, msg := range command.ExampleFlagProblems(m.Examples(), command.InputFlags(m.In)) {
			add(m.Name, RuleExample, msg)
		}
		for _, e := range m.Mapped() {
			if !e.Reason.Valid() {
				add(m.Name, RuleReason, fmt.Sprintf(
					"an error of %d has the reason %q, which is not one of the values of types.Reason.",
					e.Code,
					e.Reason,
				))
			}
		}
		for _, l := range m.location {
			if !hasField(m.Out, l) {
				add(m.Name, RuleField, fmt.Sprintf(
					"Location names %s, which is not a json field of %s.", l, m.Out))
			}
		}

		paths := map[string]reflect.StructField{}
		for _, f := range fields(m.In) {
			if f.Tag.Get("in") == "path" {
				paths[jsonName(f)] = f
			}
		}
		for _, msg := range keyProblems(m, paths) {
			add(m.Name, RuleKey, msg)
		}
		for _, p := range m.prefix {
			if _, ok := paths[p]; !ok {
				add(
					m.Name,
					RuleKey,
					"the prefix names {"+p+"}, and the input has no path field "+p+".",
				)
			}
		}

		id := m.OperationID(g)
		if id != "" && unicode.IsLower([]rune(id)[0]) {
			add(m.Name, RuleID, "operationId "+id+" starts with a small letter.")
		}
		if other, ok := ids[id]; ok {
			add(m.Name, RuleID, "operationId "+id+" is also the ID of "+other+".")
		}
		ids[id] = m.Name

		for _, f := range rule.Untagged(m.In) {
			add(m.Name, RuleValidate, "input field "+f+" has no validate tag.")
		}
		for _, msg := range refusalProblems(m) {
			add(m.Name, RuleFails, msg)
		}
		if m.HTTPMethod() != http.MethodGet {
			for _, f := range SentEmpty(bodyFields(m.In)) {
				add(
					m.Name,
					RuleOptional,
					"optional body field "+f+" is sent as its zero value when it is empty.",
				)
			}
		}

		if msg := handlerProblem(g, m); msg != "" {
			add(m.Name, RuleHandler, msg)
		}

		route := m.HTTPMethod() + " " + g.Path(m)
		if other, ok := routes[route]; ok {
			add(m.Name, RuleRoute, route+" is also the route of "+other+".")
		}
		routes[route] = m.Name
	}

	if _, err := OpenAPI(g); err != nil {
		add("", RuleSchema, "the spec does not build: "+err.Error()+".")
	}

	return out
}

// handlerProblem says what is wrong with the method's handler, or "".
func handlerProblem(
	g *Group,
	m *Method,
) string {
	switch {
	case m.handler == nil && g.Deps == nil:
		return ""
	case m.handler == nil:
		return "the group Uses " + g.Deps.String() + ", and the method has no handler."
	case g.Deps == nil:
		return "the method has a handler, and the group Uses no type."
	}
	want := reflect.FuncOf(
		[]reflect.Type{reflect.TypeFor[context.Context](), g.Deps, m.In},
		[]reflect.Type{m.Out, reflect.TypeFor[error]()},
		false,
	)
	if got := reflect.TypeOf(m.handler); got != want {
		return fmt.Sprintf("the handler is %s; it must be %s.", got, want)
	}

	return ""
}

// refusalProblems is each Fails input that is not the method's input, and
// each rule of the input that no Fails input breaks.
func refusalProblems(
	m *Method,
) []string {
	var out []string
	broken := map[string]bool{}
	for _, r := range m.refusals {
		if got := reflect.TypeOf(r.Input); got != m.In {
			out = append(out, fmt.Sprintf("the .Fails input %q is %v, not %v.", r.Doc, got, m.In))

			continue
		}
		for _, b := range validation.Breaks(r.Input) {
			broken[b] = true
		}
	}
	for _, n := range rule.Needs(m.In, true) {
		if !broken[n.Key()] {
			out = append(out, "no .Fails input breaks "+n.Rule+" on "+n.Field+".")
		}
	}

	return out
}

// bodyFields is the fields of a request body: the fields of the type that
// in:"body" names, and each input field that is not a path, query or form
// value.
func bodyFields(
	in reflect.Type,
) []reflect.StructField {
	var body []reflect.StructField
	for _, f := range fields(in) {
		switch f.Tag.Get("in") {
		case "path", "query", "form":
		case "body":
			t := f.Type
			for t.Kind() == reflect.Pointer {
				t = t.Elem()
			}
			body = append(body, fields(t)...)
		default:
			body = append(body, f)
		}
	}

	return body
}

// SentEmpty is each field of a request body, fs, that is optional, with
// omitempty in its validate tag, but has no omitempty in its json tag: a
// client sends "" or 0 for it, which its other rules can refuse. The
// behavior DSL asks it of a write's body, so the two cannot disagree.
func SentEmpty(
	fs []reflect.StructField,
) []string {
	var out []string
	for _, f := range fs {
		name, opts, _ := strings.Cut(f.Tag.Get("json"), ",")
		if name == "" || name == "-" {
			continue
		}
		if !slices.Contains(strings.Split(f.Tag.Get("validate"), ","), "omitempty") {
			continue
		}
		if !slices.Contains(strings.Split(opts, ","), "omitempty") {
			out = append(out, name)
		}
	}

	return out
}

// keyProblems are what is wrong with the key of m: a field that .Key names
// and the input does not have in the path, a field named twice, a secret,
// and a path field that is neither the prefix's nor a field of the key.
func keyProblems(
	m *Method,
	paths map[string]reflect.StructField,
) []string {
	var out []string
	if len(m.keys) == 0 {
		extra := 0
		for name := range paths {
			if !slices.Contains(m.prefix, name) {
				extra++
			}
		}
		if extra > 1 {
			out = append(out, "the input has more than one path field besides the prefix; "+
				"name each in order with .Key.")
		}

		return out
	}
	seen := map[string]bool{}
	for _, k := range m.keys {
		f, ok := paths[k]
		switch {
		case seen[k]:
			out = append(out, "Key names {"+k+"} twice.")
		case !ok:
			out = append(out, "Key names {"+k+"}, and the input has no path field "+k+".")
		case f.Type == reflect.TypeFor[secret.Value]():
			out = append(out, "Key names {"+k+"}, which is a secret; a URL goes in logs.")
		}
		seen[k] = true
	}
	names := make([]string, 0, len(paths))
	for n := range paths {
		names = append(names, n)
	}
	slices.Sort(names)
	for _, f := range names {
		if !seen[f] && !slices.Contains(m.prefix, f) {
			out = append(out, "path field "+f+" is not in Key or the prefix.")
		}
	}

	return out
}

// stutters are the types that the group's methods use, of a package under
// internal/, whose name repeats the name of their package, in order. A type
// of pkg/sdk is the SDK's to name.
func stutters(
	g *Group,
) []string {
	var out []string
	seen := map[reflect.Type]bool{}
	var walk func(t reflect.Type)
	walk = func(t reflect.Type) {
		for t.Kind() == reflect.Pointer || t.Kind() == reflect.Slice ||
			t.Kind() == reflect.Array || t.Kind() == reflect.Map {
			t = t.Elem()
		}
		if t.Kind() != reflect.Struct || t.Name() == "" || seen[t] {
			return
		}
		seen[t] = true
		if pkg := exportedPkg(t); strings.Contains(t.PkgPath(), "/internal/") &&
			strings.HasPrefix(t.Name(), pkg) && len(t.Name()) > len(pkg) {
			out = append(out, fmt.Sprintf(
				"type %s repeats its package: name it %s, and give it _ struct{} `schema:%q` "+
					"to keep its schema name.", t, strings.TrimPrefix(t.Name(), pkg), SchemaName(t)))
		}
		for f := range t.Fields() {
			walk(f.Type)
		}
	}
	for _, m := range g.Methods {
		if m.In != nil {
			walk(m.In)
		}
		if m.Out != nil {
			walk(m.Out)
		}
	}
	slices.Sort(out)

	return out
}

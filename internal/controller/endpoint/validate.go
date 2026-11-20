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
	"reflect"
	"regexp"
	"slices"
	"unicode"

	"github.com/osapi-io/osapi/internal/validation"

	"github.com/osapi-io/osapi/internal/validation/rule"

	"github.com/osapi-io/osapi/internal/guide"
)

// name is a group, a method, a sub-path or an action: lowercase letters and
// digits, starting with a letter.
var name = regexp.MustCompile(`^[a-z][a-z0-9]*$`)

// The rules Validate checks, by the name a Problem reports.
const (
	RuleName    = "name"
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
)

// Checks are the rules Validate applies. osapi controller endpoint
// validate --help prints each one, and each problem carries the fix of its
// rule, so an author or an agent can read the rules before breaking one.
// The tests trigger each rule, and fail on a rule without a check.
var Checks = guide.Checks{
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
		Requires: "An input has at most one field tagged in:\"path\".",
		Fix:      "Tag one field in:\"path\"; a group has one key.",
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
		Requires: "Each Sub segment that makes a command of its own, such as trace in osapi client telemetry trace get, has a SubDoc: the one line of help of that command.",
		Fix:      "Add endpoint.SubDoc(\"segment\", \"...\") to Define, with what the segment's commands do, such as Read a trace from the tracing backend.",
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

		for _, seg := range []string{m.sub, m.action} {
			if seg != "" && !name.MatchString(seg) {
				add(m.Name, RuleSegment,
					fmt.Sprintf("path segment %q is not lowercase letters and digits.", seg))
			}
		}
		if m.action != "" && m.Key() == "" {
			add(m.Name, RuleAction, "the method has an action but no key.")
		}
		if seg := m.sub; seg != "" && seg != m.CommandName() && g.SubDocs[seg] == "" {
			add(m.Name, RuleSubDoc, fmt.Sprintf("the Sub segment %q has no SubDoc.", seg))
		}

		keys := 0
		paths := map[string]bool{}
		for _, f := range fields(m.In) {
			if f.Tag.Get("in") == "path" {
				paths[jsonName(f)] = true
				if !slices.Contains(m.prefix, jsonName(f)) {
					keys++
				}
			}
		}
		if keys > 1 {
			add(m.Name, RuleKey, "the input has more than one path field besides the prefix.")
		}
		for _, p := range m.prefix {
			if !paths[p] {
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

		if msg := handlerProblem(g, m); msg != "" {
			add(m.Name, RuleHandler, msg)
		}

		route := m.Verb + " " + g.Path(m)
		if other, ok := routes[route]; ok {
			add(m.Name, RuleRoute, route+" is also the route of "+other+".")
		}
		routes[route] = m.Name
	}

	if _, err := OpenAPI(g, ErrorRef); err != nil {
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

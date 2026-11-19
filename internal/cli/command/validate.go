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

package command

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/osapi-io/osapi/internal/guide"
	"github.com/osapi-io/osapi/internal/validation"
	"github.com/osapi-io/osapi/internal/validation/rule"
)

// name is a group, a method, a segment or a word of Under: lowercase
// letters and digits, starting with a letter.
var name = regexp.MustCompile(`^[a-z][a-z0-9]*$`)

// Verbs are the verbs of a resource: each command is <resource> <verb>, and
// a last word that is not one of these is an action in Actions.
var Verbs = map[string]bool{
	"list":   true,
	"get":    true,
	"create": true,
	"update": true,
	"delete": true,
}

// Actions are the commands that do something a CRUD verb does not say.
// Add one here only for a real action, never for a noun: list the noun's
// items with list under it instead, such as facts keys list.
var Actions = map[string]bool{
	"admit": true, "reject": true, "revoke": true, "drain": true, "undrain": true, "retry": true,
	"export": true, "exec": true, "shell": true, "start": true, "stop": true,
	"restart": true, "enable": true, "disable": true, "install": true,
	"uninstall": true, "upload": true, "deploy": true, "undeploy": true,
	"pull": true, "signal": true, "reboot": true, "shutdown": true, "ping": true,
	"status": true, "fingerprint": true, "liveness": true, "ready": true,
	"generate": true, "validate": true,
}

// The rules Validate checks, by the name a Problem reports.
const (
	RuleName     = "name"
	RuleMethods  = "methods"
	RuleUnique   = "unique"
	RuleVerb     = "verb"
	RuleHelp     = "help"
	RuleFlag     = "flag"
	RuleDoc      = "doc"
	RuleValidate = "validate"
	RuleFails    = "fails"
	RuleHandler  = "handler"
	RuleReason   = "reason"
	RuleTag      = "tag"
	RuleSays     = "says"
)

// Checks are the rules Validate applies. osapi cli command validate --help
// prints each one, and each problem carries the fix of its rule, so an
// author or an agent can read the rules before breaking one. The tests
// trigger each rule, and fail on a rule without a check.
var Checks = guide.Checks{
	{
		Rule:     RuleName,
		Requires: "The group, each word of Under, each method and each Sub have a name of lowercase letters and digits, starting with a letter.",
		Fix:      "Name the group for what its commands are about, such as token, and each method for what it does, such as generate.",
	},
	{
		Rule:     RuleMethods,
		Requires: "The group has at least one method.",
		Fix:      "Add a method with command.Method, with its input and output types and its name.",
	},
	{
		Rule:     RuleUnique,
		Requires: "No two methods of a group have one name: the name is the key of the method's renderer.",
		Fix:      "Give each method its own name.",
	},
	{
		Rule:     RuleVerb,
		Requires: "The last word of each command is a verb (list, get, create, update, delete) or an action in command.Actions, as TestCommandShape holds each command to.",
		Fix:      "Name the method for a verb, and put a noun before it with .Sub, such as roles list. Add a word to command.Actions only for a real action, never for a noun.",
	},
	{
		Rule:     RuleHelp,
		Requires: "Each method has a Short, and a method with input has an Example.",
		Fix:      "Add .Short(\"<what it does>\") and .Example(\"--flag value\") to the method.",
	},
	{
		Rule:     RuleFlag,
		Requires: "Each value is a flag: the input is a struct, each field has a json name that is its flag with _ as -, no flag or in tag, and a type a flag holds (string, bool, int, []string, map[string]string or []byte). A command takes no positional argument.",
		Fix:      "Make the input a struct, tag each field json:\"name\", drop the flag and in tags, and use one of the flag types.",
	},
	{
		Rule:     RuleDoc,
		Requires: "Each input field has a doc tag: the help of its flag.",
		Fix:      "Add doc:\"...\" to the field: a capital letter first and no period at the end.",
	},
	{
		Rule:     RuleValidate,
		Requires: "Each input field has a validate tag: its rules, or omitempty for an optional field with none. Run checks the value against it before the handler runs, and a value that breaks it is a usage error.",
		Fix:      rule.Fix,
	},
	{
		Rule:     RuleFails,
		Requires: "Each rule of an input field has a .Fails input that breaks it, which the command refuses with a usage error, required included. A min of 1 on a string that is required or optional needs none: its only bad value is empty.",
		Fix:      rule.FailsFix,
	},
	{
		Rule:     RuleHandler,
		Requires: "The group Uses a type, and each method has a handler, func(context.Context, D, In) (Out, error), with D that type and In and Out the method's types.",
		Fix:      "Add command.Uses[Deps]() to Define, and .Handle(fn) to each method, with the signature that the message names.",
	},
	{
		Rule:     RuleReason,
		Requires: "A method of any group, endpoint, behavior or local, prints from its output type. A renderer that Render sets in its place says why the output type cannot print the answer, with .Because.",
		Fix:      "Delete the renderer, and give the output fields the tags that osapi cli command spec lists. When the output type cannot print the answer, such as a tree, add .Because(\"<why>\") to the renderer.",
	},
	{
		Rule:     RuleTag,
		Requires: "Each render tag of an output field is one the output knows: -, bytes, age, time, duration, duration=ms, status or short=N with N above 0. An empty tag is on a list, and a title tag is on the blank field _ of the output type.",
		Fix:      "Use a render value that osapi cli command spec lists, put empty on a list field, and put title on a field _ struct{} of the output type.",
	},
	{
		Rule:     RuleSays,
		Requires: "A method that answers with no body says what occurred: its name has a past tense, such as admitted for admit, or .Says gives the text.",
		Fix:      "Add .Says(\"<state>\") to the method, such as .Says(\"draining\") for drain.",
	},
}

// Problem is a rule a definition breaks, and how to fix it.
type Problem struct {
	// Group is the command path of the group it is about.
	Group string `json:"group"`
	// Method is the method it is about, or "" for the group.
	Method string `json:"method,omitempty"`
	// Rule is the name of the rule, from Checks.
	Rule string `json:"rule"`
	// Message says what is wrong.
	Message string `json:"message"`
	// Fix is how to fix it, from the check of Rule.
	Fix string `json:"fix"`
}

func (p Problem) String() string {
	return p.Message + " Fix: " + p.Fix
}

// flagTypes are the types a flag holds.
var flagTypes = map[reflect.Type]bool{
	reflect.TypeFor[string]():            true,
	reflect.TypeFor[bool]():              true,
	reflect.TypeFor[int]():               true,
	reflect.TypeFor[[]string]():          true,
	reflect.TypeFor[map[string]string](): true,
	reflect.TypeFor[[]byte]():            true,
}

// Validate is each rule g breaks. A group with no problems makes commands
// that hold to the CLI's rules.
func Validate(
	g *LocalGroup,
) []Problem {
	var out []Problem
	group := strings.Join(g.CommandPath(), " ")
	add := func(method, rule, message string) {
		out = append(out, Problem{
			Group: group, Method: method, Rule: rule, Message: message, Fix: Checks.Fix(rule),
		})
	}

	for _, w := range g.CommandPath() {
		if !name.MatchString(w) {
			add("", RuleName, fmt.Sprintf("the word %q is not lowercase letters and digits.", w))
		}
	}
	if len(g.Methods) == 0 {
		add("", RuleMethods, "the group has no methods.")
	}
	if g.Deps == nil {
		add("", RuleHandler, "the group Uses no type.")
	}

	seen := map[string]bool{}
	for _, m := range g.Methods {
		for _, w := range []string{m.Name, m.sub} {
			if w != "" && !name.MatchString(w) {
				add(m.Name, RuleName, fmt.Sprintf("%q is not lowercase letters and digits.", w))
			}
		}
		if seen[m.Name] {
			add(m.Name, RuleUnique, "two methods have the name "+m.Name+".")
		}
		seen[m.Name] = true
		if !Verbs[m.Name] && !Actions[m.Name] {
			add(m.Name, RuleVerb, fmt.Sprintf("%q is not a verb or an action.", m.Name))
		}

		fields := structFields(m.In)
		if m.short == "" {
			add(m.Name, RuleHelp, "the method has no Short.")
		}
		if len(fields) > 0 && len(m.examples) == 0 {
			add(m.Name, RuleHelp, "the method has input and no Example.")
		}

		if m.In.Kind() != reflect.Struct {
			add(m.Name, RuleFlag, fmt.Sprintf("the input is %v, not a struct.", m.In))
		}
		for _, f := range fields {
			for _, msg := range flagProblems(f) {
				add(m.Name, RuleFlag, msg)
			}
			if f.Tag.Get("doc") == "" {
				add(m.Name, RuleDoc, "input field "+FlagName(f)+" has no doc.")
			}
		}
		for _, f := range rule.Untagged(m.In) {
			add(m.Name, RuleValidate, "input field "+f+" has no validate tag.")
		}
		for _, msg := range refusalProblems(m) {
			add(m.Name, RuleFails, msg)
		}
		if msg := handlerProblem(g, m); msg != "" {
			add(m.Name, RuleHandler, msg)
		}
	}

	return out
}

// structFields are the exported fields of t that JSON reads, when t is a
// struct.
func structFields(
	t reflect.Type,
) []reflect.StructField {
	if t.Kind() != reflect.Struct {
		return nil
	}
	var out []reflect.StructField
	for f := range t.Fields() {
		if f.IsExported() && !strings.HasPrefix(f.Tag.Get("json"), "-") {
			out = append(out, f)
		}
	}

	return out
}

// flagProblems says what keeps f from being a flag of its own json name.
func flagProblems(
	f reflect.StructField,
) []string {
	var out []string
	if n, _, _ := strings.Cut(f.Tag.Get("json"), ","); n == "" {
		out = append(out, "input field "+f.Name+" has no json name.")
	}
	if fl, ok := f.Tag.Lookup("flag"); ok && fl != FlagName(f) {
		out = append(
			out,
			"input field "+f.Name+" has the flag tag "+fl+": its flag is its json name.",
		)
	}
	if in, ok := f.Tag.Lookup("in"); ok {
		out = append(
			out,
			"input field "+f.Name+" has the in tag "+in+": a local command has no route.",
		)
	}
	if !flagTypes[f.Type] {
		out = append(out, fmt.Sprintf("input field %s is %v, which no flag holds.", f.Name, f.Type))
	}

	return out
}

// handlerProblem says what is wrong with the method's handler, or "".
func handlerProblem(
	g *LocalGroup,
	m *LocalMethod,
) string {
	if m.handler == nil {
		return "the method has no handler."
	}
	if g.Deps == nil {
		return ""
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
	m *LocalMethod,
) []string {
	var out []string
	broken := map[string]bool{}
	for _, r := range m.refusals {
		if got := reflect.TypeOf(r.Input); got != m.In {
			out = append(out, fmt.Sprintf("the .Fails input %q is %v, not %v.", r.Doc, got, m.In))

			continue
		}
		for _, b := range validation.Breaks(r.Input) {
			// A rule after dive breaks on an item, such as Roles[0]: it is
			// the field's rule.
			field, tag, _ := strings.Cut(b, ":")
			field, _, _ = strings.Cut(field, "[")
			broken[field+":"+tag] = true
		}
	}
	for _, n := range rule.Needs(m.In, true) {
		if !broken[n.Key()] {
			out = append(out, "no .Fails input breaks "+n.Rule+" on "+n.Field+".")
		}
	}

	return out
}

// renderValues are the render tags that the output knows, besides short=N.
var renderValues = map[string]bool{
	RenderHidden: true, RenderBytes: true, RenderAge: true, RenderTime: true,
	RenderDuration: true, RenderDurationMs: true, RenderStatus: true,
}

// ValidateOutput is each rule that the output of a registered group breaks:
// each group whose commands AddCommands added, endpoint, behavior and local.
// A renderer has a reason, each tag of an output type has a value that the
// output knows, and a method with no body says what occurred.
func ValidateOutput() []Problem {
	var out []Problem
	for _, g := range Groups() {
		group := strings.Join(g.CommandPath(), " ")
		add := func(method, rule, message string) {
			out = append(out, Problem{
				Group: group, Method: method, Rule: rule, Message: message, Fix: Checks.Fix(rule),
			})
		}
		for _, m := range g.Commands() {
			r, rendered := Rendered(g.GroupName(), m.MethodName())
			if rendered && r.Reason() == "" {
				add(m.MethodName(), RuleReason, "the renderer has no Because.")
			}
			for _, msg := range tagProblems(m.Output(), map[reflect.Type]bool{}) {
				add(m.MethodName(), RuleTag, msg)
			}
			t := m.Output()
			if !rendered && t.Kind() == reflect.Struct && answersNothing(t) && said(m) == "" {
				add(m.MethodName(), RuleSays, fmt.Sprintf("%q has no past tense.", m.MethodName()))
			}
		}
	}

	return out
}

// tagProblems says which output tags of t, and of each type in it, the
// output does not know.
func tagProblems(
	t reflect.Type,
	seen map[reflect.Type]bool,
) []string {
	t = element(t)
	if t.Kind() != reflect.Struct || seen[t] {
		return nil
	}
	seen[t] = true

	var out []string
	for f := range t.Fields() {
		if f.Name == "_" {
			continue
		}
		where := t.Name() + "." + f.Name
		if render, ok := f.Tag.Lookup("render"); ok && !knownRender(render) {
			out = append(
				out,
				fmt.Sprintf("%s has render:%q, which the output does not know.", where, render),
			)
		}
		if _, ok := f.Tag.Lookup("empty"); ok && f.Type.Kind() != reflect.Slice {
			out = append(out, where+" has an empty tag and is not a list.")
		}
		if _, ok := f.Tag.Lookup("title"); ok {
			out = append(out, where+" has a title tag: it goes on the field _ struct{}.")
		}
		out = append(out, tagProblems(f.Type, seen)...)
	}

	return out
}

// knownRender reports whether a render tag is one the output knows.
func knownRender(
	render string,
) bool {
	if n, ok := strings.CutPrefix(render, RenderShort); ok {
		v, err := strconv.Atoi(n)

		return err == nil && v > 0
	}

	return renderValues[render]
}

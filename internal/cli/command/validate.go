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
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/guide"
	"github.com/osapi-io/osapi/internal/secret"
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
// items with list under it instead, such as agent facts list.
var Actions = map[string]bool{
	"admit": true, "reject": true, "revoke": true, "drain": true, "undrain": true, "retry": true,
	"export": true, "exec": true, "shell": true, "start": true, "stop": true,
	"restart": true, "enable": true, "disable": true, "install": true,
	"uninstall": true, "upload": true, "deploy": true, "undeploy": true,
	"pull": true, "signal": true, "reboot": true, "shutdown": true, "ping": true,
	"status": true, "fingerprint": true, "liveness": true, "ready": true,
	"generate": true, "validate": true, "setup": true,
}

// RootGroup is a word directly under osapi, and what goes under it.
type RootGroup struct {
	// Name is the word, such as controller.
	Name string
	// Purpose says what goes under the word, in one line: the help of its
	// command, and its row on the CLI page of the docs site.
	Purpose string
	// Leaf is true for a command of the root itself, with no words under
	// it, such as version.
	Leaf bool
}

// RootGroups are the words directly under osapi, a closed list. A command
// goes where it runs and what it touches: client for a call to the API,
// controller, agent or nats for the process or the files of that component
// on this host, config for the config that each component reads, dsl for
// the definitions of osapi, and skill for the skills in the binary. Only
// start and version act on the whole binary or host. TestCommandShape fails
// on any other word. A new word is a decision of the owner, recorded in the
// specs cli.md.
var RootGroups = []RootGroup{
	{Name: "start", Purpose: "Start each component of osapi on this host", Leaf: true},
	{Name: "version", Purpose: "Show the version of this binary", Leaf: true},
	{
		Name:    "controller",
		Purpose: "Run the controller on this host, and work on its key, its tokens and its roles",
	},
	{Name: "agent", Purpose: "Run the agent on this host, and work on its key"},
	{
		Name:    "nats",
		Purpose: "Run the embedded NATS server on this host, or prepare a NATS server",
	},
	{Name: "config", Purpose: "List, get and validate the config that each component reads"},
	{
		Name:    "skill",
		Purpose: "Show the agent skills in this binary, and install them into a repository",
	},
	{Name: "dsl", Purpose: "Describe, check and document the definitions of osapi, for authors"},
	{Name: "client", Purpose: "Call the API of the controller"},
}

// Root is the RootGroup of word, and false for a word that is not one.
func Root(
	word string,
) (RootGroup, bool) {
	for _, r := range RootGroups {
		if r.Name == word {
			return r, true
		}
	}

	return RootGroup{}, false
}

// DescribeRoot gives each command directly under root the purpose of its
// word in RootGroups as its help, so that osapi --help and the CLI page of
// the docs site list each word with its purpose from one list.
func DescribeRoot(
	root *cobra.Command,
) {
	for _, c := range root.Commands() {
		if r, ok := Root(c.Name()); ok {
			c.Short = r.Purpose
		}
	}
}

// The rules Validate checks, by the name a Problem reports.
const (
	RuleName     = "name"
	RuleMethods  = "methods"
	RuleUnique   = "unique"
	RuleVerb     = "verb"
	RuleRoot     = "root"
	RuleHelp     = "help"
	RuleFlag     = "flag"
	RuleDoc      = "doc"
	RuleValidate = "validate"
	RuleFails    = "fails"
	RuleHandler  = "handler"
	RuleTag      = "tag"
	RuleSays     = "says"
	RuleSubDoc   = "subdoc"
	RulePrints   = "prints"
	// RuleSecret: a secret.Value input field is not tagged telemetry:"attr".
	RuleSecret = "secret"
)

// Checks are the rules Validate applies. osapi dsl command validate --help
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
		Requires: "No two methods of a group have one name.",
		Fix:      "Give each method its own name.",
	},
	{
		Rule:     RuleVerb,
		Requires: "The last word of each command is a verb (list, get, create, update, delete) or an action in command.Actions, as TestCommandShape holds each command to.",
		Fix:      "Name the method for a verb, and put a noun before it with .Sub, such as key fingerprint. Add a word to command.Actions only for a real action, never for a noun.",
	},
	{
		Rule:     RuleRoot,
		Requires: "A command goes where it runs and what it touches. Its first word is one of command.RootGroups: controller, agent or nats for the process or the files of that component on this host, config for the config that each component reads, dsl for the definitions of osapi, skill for the skills in the binary, and client for a call to the API. A local group calls no API, so it is never under client. Only start and version are commands of the root itself.",
		Fix:      "Put the group Under the command.Parent of the component whose process or files it works on, or of dsl for a command for authors. Add a word to command.RootGroups only as a decision of the owner, recorded in the specs cli.md.",
	},
	{
		Rule:     RuleHelp,
		Requires: "Each method has a Short, and a method with input has an Example. Each flag that an Example names is a flag of the command: a field of its input, a key that its Flag names, or --json, --debug or --osapi-file.",
		Fix:      "Add .Short(\"<what it does>\") and .Example(\"--flag value\") to the method, and spell each flag of an example as the command has it.",
	},
	{
		Rule:     RuleFlag,
		Requires: "Each value is a flag: the input is a struct, each field has a json name that is its flag with _ as -, no flag or in tag, and a type a flag holds (string, bool, int, []string, map[string]string, []byte, secret.Value or map[string]secret.Value). A command takes no positional argument.",
		Fix:      "Make the input a struct, tag each field json:\"name\", drop the flag and in tags, and use one of the flag types.",
	},
	{
		Rule:     RuleSecret,
		Requires: "No secret.Value input field is tagged telemetry:\"attr\": a span is no place for a secret. A secret flag takes a reference, secret://<vault>/<key>, or - to read the value from standard input, and the command resolves a reference itself.",
		Fix:      "Remove telemetry:\"attr\" from the secret field.",
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
		Rule:     RuleSubDoc,
		Requires: "Each Sub segment that makes a command of its own, such as method in osapi dsl behavior method get, has a SubDoc: the one line of help of that command.",
		Fix:      "Add command.SubDoc(\"segment\", \"...\") to Define, with what the segment's commands do, such as Show one method of a behavior in full.",
	},
	{
		Rule:     RuleHandler,
		Requires: "The group Uses a type, and each method has a handler, func(context.Context, D, In) (Out, error), with D that type and In and Out the method's types.",
		Fix:      "Add command.Uses[Deps]() to Define, and .Handle(fn) to each method, with the signature that the message names.",
	},
	{
		Rule:     RuleTag,
		Requires: "Each tag key of an output field is one of command.KnownTags. Each render tag is one the output knows: -, bytes, age, time, duration, duration=ms, status or short=N with N above 0. An empty tag is on a list, and a title tag is on the blank field _ of the output type.",
		Fix:      "Use a render value that osapi dsl command spec get lists, put empty on a list field, and put title on a field _ struct{} of the output type.",
	},
	{
		Rule:     RuleSays,
		Requires: "A method that answers with no body says what occurred: its name has a past tense, such as admitted for admit, or .Says gives the text.",
		Fix:      "Add .Says(\"<state>\") to the method, such as .Says(\"draining\") for drain.",
	},
	{
		Rule:     RulePrints,
		Requires: "Each method that prints a body, and is not a process, has an Example that Prints gives its output: a value of the method's output type, which its docs page shows as the CLI prints it. Each process has an Example that Logs gives the lines it writes when it starts, which its docs page shows as its log handler writes them. Prints and Logs follow the Example that they show.",
		Fix:      "Add .Prints(sample) after the Example that it shows, with sample a value of the method's output type, in samples.go of the group. For a process, add .Logs(lines...) after its Example, with the lines in samples.go of the group.",
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
	Fix string `json:"fix" table:"-"`
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
	// A secret is a string flag, and a map of secrets a key=value flag:
	// the command reads each as Secret and Secrets read it.
	reflect.TypeFor[secret.Value]():            true,
	reflect.TypeFor[map[string]secret.Value](): true,
}

// Flaggable reports whether a flag can hold a value of type t: the types
// that the command generator builds a flag for. The behavior DSL asks it,
// so the two cannot disagree.
func Flaggable(
	t reflect.Type,
) bool {
	return flagTypes[t]
}

// Validate is each rule g breaks. A group with no problems makes commands
// that hold to the CLI's rules.
func Validate(
	g *LocalGroup,
) []Problem {
	var out []Problem
	group := g.Path()
	add := func(method, rule, message string) {
		out = append(out, Problem{
			Group: group, Method: method, Rule: rule, Message: message, Fix: Checks.Fix(rule),
		})
	}

	for _, w := range append(append([]string(nil), g.Parents...), g.Name) {
		if !name.MatchString(w) {
			add("", RuleName, fmt.Sprintf("the word %q is not lowercase letters and digits.", w))
		}
	}
	for _, w := range g.Parents {
		if name.MatchString(w) && !slices.Contains(KnownParents(), Parent(w)) {
			add("", RuleName, fmt.Sprintf(
				"the group is Under %q, which is not one of KnownParents; use a command.Parent constant.",
				w,
			))
		}
	}
	if words := g.CommandPath(); len(words) > 0 {
		switch r, ok := Root(words[0]); {
		case !ok || r.Leaf:
			add("", RuleRoot, fmt.Sprintf("%q is not a group of RootGroups.", words[0]))
		case r.Name == "client":
			add("", RuleRoot, "a local group calls no API, so it is not under client.")
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
		if seen[m.MethodName()] {
			add(m.Name, RuleUnique, "two methods are "+m.MethodName()+".")
		}
		seen[m.MethodName()] = true
		switch {
		case g.Flat && len(g.Parents) == 0 && m.sub == "":
			if r, ok := Root(m.Name); !ok || !r.Leaf {
				add(
					m.Name,
					RuleRoot,
					fmt.Sprintf(
						"%q is not a command of the root; only the leaves of RootGroups are.",
						m.Name,
					),
				)
			}
		case !Verbs[m.Name] && !Actions[m.Name]:
			add(m.Name, RuleVerb, fmt.Sprintf("%q is not a verb or an action.", m.Name))
		}

		fields := structFields(m.In)
		if m.short == "" {
			add(m.Name, RuleHelp, "the method has no Short.")
		}
		if seg := m.sub; seg != "" && seg != m.Name && g.SubDocs[seg] == "" {
			add(m.Name, RuleSubDoc, fmt.Sprintf("the Sub segment %q has no SubDoc.", seg))
		}
		if len(fields) > 0 && len(m.examples) == 0 {
			add(m.Name, RuleHelp, "the method has input and no Example.")
		}
		for _, msg := range exampleFlagProblems(m) {
			add(m.Name, RuleHelp, msg)
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
			if secret.Is(f.Type) && f.Tag.Get("telemetry") == "attr" {
				add(m.Name, RuleSecret, "input field "+FlagName(f)+
					" is a secret and tagged telemetry:\"attr\"; a secret is never recorded on a span.")
			}
		}
		for _, f := range rule.Untagged(m.In) {
			add(m.Name, RuleValidate, "input field "+f+" has no validate tag.")
		}
		if !m.process && len(m.keys) > 0 {
			add(m.Name, RuleFlag, "a method's flags are its input, so it has no Flag.")
		}
		if m.process && len(m.refusals) > 0 {
			add(m.Name, RuleFails, "a process has no input, so nothing Fails.")
		}
		for _, msg := range refusalProblems(m) {
			add(m.Name, RuleFails, msg)
		}
		for _, msg := range printsProblems(m) {
			add(m.Name, RulePrints, msg)
		}
		if msg := handlerProblem(g, m); msg != "" {
			add(m.Name, RuleHandler, msg)
		}
	}

	return out
}

// printsProblems says what is wrong with the outputs that Prints and Logs
// gave the examples of m: one before any Example, one of another type than
// m's output, none for a method that prints a body, a log on a method that
// is not a process, or none for a process. A method with no body prints one
// line, which its page derives from the example's flags.
func printsProblems(
	m *LocalMethod,
) []string {
	var out []string
	if m.unprinted {
		out = append(out, "Prints comes before any Example; it gives the last Example its output.")
	}
	if m.unlogged {
		out = append(out, "Logs comes before any Example; it gives the last Example its log.")
	}
	logged := false
	for _, ex := range m.examples {
		if len(ex.Logs) > 0 {
			logged = true
		}
	}
	if logged && !m.process {
		out = append(out, "Logs is for a process: give this method's output with Prints.")
	}
	if !logged && m.process {
		out = append(out, "no Example has a log: add .Logs(lines...) after one.")
	}
	printed := false
	for _, ex := range m.examples {
		if ex.Output == nil {
			continue
		}
		printed = true
		if t := reflect.TypeOf(ex.Output); t != m.Out {
			out = append(out, fmt.Sprintf(
				"Prints gives the example %q a %s, not a %s.", ex.Flags, t, m.Out))
		}
	}
	if !printed && !m.process && m.Out != nil && !answersNothing(m.Out) {
		out = append(out, "no Example has an output: add .Prints(sample) after one.")
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
	if m.process {
		want := reflect.FuncOf(
			[]reflect.Type{reflect.TypeFor[context.Context](), g.Deps},
			[]reflect.Type{reflect.TypeFor[Running](), reflect.TypeFor[error]()},
			false,
		)
		if got := reflect.TypeOf(m.handler); got != want {
			return fmt.Sprintf("the Setup is %s; it must be %s.", got, want)
		}

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
	RenderDuration: true, RenderDurationMs: true, RenderStatus: true, RenderBody: true,
}

// ValidateOutput is each rule that the output of a registered group breaks:
// each group whose commands AddCommands added, endpoint, behavior and local.
// Each tag of an output type has a value that the output knows, and a method
// with no body says what occurred.
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
			// A process has no output type: its page shows its log.
			if lm, ok := m.(*LocalMethod); ok && lm.process {
				continue
			}
			for _, msg := range tagProblems(m.Output(), map[reflect.Type]bool{}) {
				add(m.MethodName(), RuleTag, msg)
			}
			t := m.Output()
			if t.Kind() == reflect.Struct && answersNothing(t) && said(m) == "" {
				add(m.MethodName(), RuleSays, fmt.Sprintf("%q has no past tense.", m.MethodName()))
			}
		}
	}

	return out
}

// tagProblems says which output tags of t, and of each type in it, the
// output does not know.
// TagProblems is each tag of output type t, and of the types it holds,
// that the output does not know: the check that ValidateOutput makes on a
// registered command, for a validator that has the type before a command.
func TagProblems(
	t reflect.Type,
) []string {
	return tagProblems(t, map[reflect.Type]bool{})
}

// KnownTags are the struct tag keys that a field of an input or an output
// of a command, a behavior or an endpoint can carry: the DSLs, the
// validator, the output's renderer and the settings loader read them. A key
// outside the list is a typo: Go reads a tag only by its exact key, so the
// field gets nothing from it.
var KnownTags = []string{
	"before", "default", "doc", "empty", "example", "flag", "in", "json", "label",
	"mapstructure", "none", "param", "pattern", "prefix", "render",
	"short", "table", "telemetry", "template", "title", "validate", "yaml",
}

// TagKeys are the keys of tag, in order, by the convention that
// reflect.StructTag.Get reads: key:"value" pairs separated by spaces.
func TagKeys(
	tag reflect.StructTag,
) []string {
	var keys []string
	s := string(tag)
	for s != "" {
		s = strings.TrimLeft(s, " ")
		i := strings.Index(s, ":\"")
		if i <= 0 {
			break
		}
		keys = append(keys, s[:i])
		s = s[i+2:]
		for j := 0; j < len(s); j++ {
			if s[j] == '\\' {
				j++

				continue
			}
			if s[j] == '"' {
				s = s[j+1:]

				break
			}
		}
	}

	return keys
}

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
		for _, key := range TagKeys(f.Tag) {
			if !slices.Contains(KnownTags, key) {
				out = append(
					out,
					fmt.Sprintf("%s has the tag key %q, which nothing reads.", where, key),
				)
			}
		}
		if render, ok := f.Tag.Lookup("render"); ok && !knownRender(render) {
			out = append(
				out,
				fmt.Sprintf("%s has render:%q, which the output does not know.", where, render),
			)
		}
		if _, ok := f.Tag.Lookup("empty"); ok && f.Type.Kind() != reflect.Slice {
			out = append(out, where+" has an empty tag and is not a list.")
		}
		if tb, ok := f.Tag.Lookup(
			"table",
		); ok && tb != "-" && tb != TableBlocks &&
			tb != TableSections {
			out = append(
				out,
				fmt.Sprintf("%s has table:%q, which the output does not know.", where, tb),
			)
		}
		if tb := f.Tag.Get("table"); (tb == TableBlocks || tb == TableSections) && !isList(f.Type) {
			out = append(
				out,
				fmt.Sprintf("%s has table:%q and is not a list of objects.", where, tb),
			)
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

// globalFlags are the flags of each command, which the root command adds:
// the long names and their short letters.
var globalFlags = map[string]bool{
	"json": true, "debug": true, "osapi-file": true, "help": true,
	"j": true, "d": true, "f": true, "h": true,
}

// GlobalFlag reports whether name, without its dashes, is a flag that the
// root command adds to each command, such as json.
func GlobalFlag(
	name string,
) bool {
	return globalFlags[name]
}

// exampleFlagProblems are the flags that an example of m names and the
// command does not have: its input's flags, and the flag of each key that
// its Flag names.
func exampleFlagProblems(
	m *LocalMethod,
) []string {
	flags := InputFlags(m.In)
	for _, k := range m.keys {
		// As config.FlagName names the flag of a key.
		flags[strings.NewReplacer(".", "-", "_", "-").Replace(k)] = true
	}

	return ExampleFlagProblems(m.examples, flags)
}

// InputFlags are the flags of the command of a method whose input is in,
// as the command generator declares them: each field's flag tag, else its
// FlagName, and its short tag. A field tagged in:"body" that is a struct
// gives a flag for each of its fields. The keys are the names without the
// dashes.
func InputFlags(
	in reflect.Type,
) map[string]bool {
	flags := map[string]bool{}
	add := func(f reflect.StructField) {
		name := FlagName(f)
		if tag := f.Tag.Get("flag"); tag != "" {
			name = tag
		}
		flags[name] = true
		if short := f.Tag.Get("short"); short != "" {
			flags[short] = true
		}
	}
	for _, f := range structFields(in) {
		if f.Tag.Get("in") == "body" && f.Type.Kind() == reflect.Struct {
			for _, bf := range structFields(f.Type) {
				add(bf)
			}

			continue
		}
		add(f)
	}

	return flags
}

// ExampleFlagProblems are the flags that examples name and that are not in
// flags or the flags of each command. A misspelled flag made a docs page
// show a command that fails. The command DSL and the endpoint DSL both
// check their examples with it.
func ExampleFlagProblems(
	examples []Example,
	flags map[string]bool,
) []string {
	var out []string
	for _, e := range examples {
		for _, word := range strings.Fields(e.Flags) {
			name, ok := strings.CutPrefix(word, "-")
			if !ok || name == "" {
				continue
			}
			name = strings.TrimPrefix(name, "-")
			name, _, _ = strings.Cut(name, "=")
			if !flags[name] && !globalFlags[name] {
				out = append(
					out,
					fmt.Sprintf(
						"the example %q names %s, which the command does not have.",
						e.Flags,
						word,
					),
				)
			}
		}
	}

	return out
}

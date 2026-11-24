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
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/osapi-io/osapi/pkg/sdk/client"
)

// LocalGroup is a group of local commands: commands that run in the CLI
// process, with no API, such as osapi controller token generate. Define makes one; its
// gen/cli.go makes a cobra command for each method, and Run runs it.
type LocalGroup struct {
	// Name is the group's command, the first word under its path.
	Name string
	// Doc says what the group is, in one line: its command's help.
	Doc string
	// Parents are the words of the command the group goes under, which
	// Under sets, such as controller. The root has none.
	Parents []string
	// Deps is the type Uses names: what the methods get from the CLI.
	Deps reflect.Type
	// Methods are the group's commands, in the order they were declared.
	Methods []*LocalMethod
	// Flat is true for a group that adds no word of its own, which Flat
	// sets: its commands are under its parent, such as osapi version.
	Flat bool
	// Long is the rest of the help of the group's command, which Long
	// sets.
	Long string
	// SubDocs are the one line of help of each command that a Sub makes,
	// such as roles in osapi controller role list, which SubDoc sets.
	SubDocs map[string]string

	provide func() any
}

// Option is a setting of a local group, or one of its methods.
type Option interface {
	apply(g *LocalGroup)
}

type optionFunc func(g *LocalGroup)

func (f optionFunc) apply(
	g *LocalGroup,
) {
	f(g)
}

// Define builds a local group from its options. Validate checks the result.
func Define(
	name string,
	doc string,
	opts ...Option,
) *LocalGroup {
	g := &LocalGroup{Name: name, Doc: doc}
	for _, o := range opts {
		o.apply(g)
	}

	return g
}

// Uses names D, the struct of what the group's methods use, such as the
// config or the signing of tokens. The CLI gives one D with Provide, and a
// test gives one with fakes. Each handler takes it as its second argument.
func Uses[D any]() Option {
	return optionFunc(func(g *LocalGroup) { g.Deps = reflect.TypeFor[D]() })
}

// SubDoc is the one line of help of the command that the Sub segment makes,
// such as roles in osapi controller role list.
func SubDoc(
	segment string,
	short string,
) Option {
	return optionFunc(func(g *LocalGroup) {
		if g.SubDocs == nil {
			g.SubDocs = map[string]string{}
		}
		g.SubDocs[segment] = short
	})
}

// Long is the rest of the help of the group's command, after its one line,
// such as what each of its commands reads.
func Long(
	text string,
) Option {
	return optionFunc(func(g *LocalGroup) { g.Long = text })
}

// Flat makes the group add no word of its own: each method is a command
// under the group's parent, the root or the words of Under. osapi version is
// the method version of a flat group at the root; osapi controller start
// would be the method start of a flat group Under controller. A command
// directly under the root is a leaf of RootGroups, and any other has a verb or
// an action.
func Flat() Option {
	return optionFunc(func(g *LocalGroup) { g.Flat = true })
}

// Parent is a command that local groups go under, such as controller for
// osapi dsl endpoint. A misspelled parent made a new command at the
// root with no error, so validate refuses a word outside KnownParents.
type Parent string

// The commands that local groups go under.
const (
	ParentAgent      Parent = "agent"
	ParentController Parent = "controller"
	ParentDSL        Parent = "dsl"
	ParentNATS       Parent = "nats"
)

// KnownParents are the commands that a group can be Under. Add a constant
// here to make a new one.
func KnownParents() []Parent {
	return []Parent{ParentAgent, ParentController, ParentDSL, ParentNATS}
}

// Under puts the group's command under the command of words, such as
// ParentController for osapi dsl endpoint. Without it, the group is
// under the root.
func Under(
	words ...Parent,
) Option {
	return optionFunc(func(g *LocalGroup) {
		g.Parents = make([]string, 0, len(words))
		for _, w := range words {
			g.Parents = append(g.Parents, string(w))
		}
	})
}

var (
	localMu sync.RWMutex
	locals  = map[string]*LocalGroup{}
)

// Register adds g to the local groups the CLI has, and returns it. Two
// groups with one command path are a mistake in the code, so it panics.
func Register(
	g *LocalGroup,
) *LocalGroup {
	localMu.Lock()
	defer localMu.Unlock()
	key := g.Path()
	if _, ok := locals[key]; ok {
		panic(fmt.Sprintf("command: group %q registered twice", key))
	}
	locals[key] = g

	return g
}

// Lookup is the registered local group with the command path, such as
// token, or nil.
func Lookup(
	path string,
) *LocalGroup {
	localMu.RLock()
	defer localMu.RUnlock()

	return locals[path]
}

// All is each registered local group, by command path.
func All() []*LocalGroup {
	localMu.RLock()
	defer localMu.RUnlock()
	out := make([]*LocalGroup, 0, len(locals))
	for _, g := range locals {
		out = append(out, g)
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := strings.Join(out[i].CommandPath(), " "), strings.Join(out[j].CommandPath(), " ")
		if a != b {
			return a < b
		}

		// Two flat groups under one parent have the same command path, as
		// nats start and nats setup do. Their path keeps the order fixed.
		return out[i].Path() < out[j].Path()
	})

	return out
}

// Provide gives the group's methods their deps: fn makes them when a
// command runs, after the config is read. D must be the type the group
// Uses; any other is a mistake in the code, so it panics. A group whose Deps
// has no field needs none.
func Provide[D any](
	g *LocalGroup,
	fn func() D,
) {
	if t := reflect.TypeFor[D](); t != g.Deps {
		panic(fmt.Sprintf("command: %s uses %v, and Provide gives %v", g.Name, g.Deps, t))
	}
	g.provide = func() any { return fn() }
}

// Method is the method of g with the name, its MethodName, such as get or
// spec get, or nil.
func (g *LocalGroup) Method(
	name string,
) *LocalMethod {
	for _, m := range g.Methods {
		if m.MethodName() == name {
			return m
		}
	}

	return nil
}

// A local group is a group of the command layer, and a method one of its
// commands, as an endpoint group and its methods are.
var (
	_ Group   = (*LocalGroup)(nil)
	_ Command = (*LocalMethod)(nil)
)

// GroupName is the group's Name.
func (g *LocalGroup) GroupName() string {
	return g.Name
}

// GroupDoc is the group's Doc.
func (g *LocalGroup) GroupDoc() string {
	return g.Doc
}

// GroupNotes is "": a local group's page has no Notes section.
func (g *LocalGroup) GroupNotes() string {
	return ""
}

// GroupLong is the help Long set, or "".
func (g *LocalGroup) GroupLong() string {
	return g.Long
}

// SubDoc is the help SubDoc set for the segment, or "".
func (g *LocalGroup) SubDoc(
	segment string,
) string {
	return g.SubDocs[segment]
}

// CommandPath is the words of the group's command under the root: the words
// Under set, then its name.
func (g *LocalGroup) CommandPath() []string {
	if g.Flat {
		return append([]string(nil), g.Parents...)
	}

	return append(append([]string(nil), g.Parents...), g.Name)
}

// Path is the group's words and its name, such as dsl endpoint, or
// version for a flat group at the root: the key that Register and Lookup
// use, and the name that a problem and dsl command get give it.
func (g *LocalGroup) Path() string {
	return strings.Join(append(append([]string(nil), g.Parents...), g.Name), " ")
}

// Commands are the group's methods.
func (g *LocalGroup) Commands() []Command {
	out := make([]Command, 0, len(g.Methods))
	for _, m := range g.Methods {
		out = append(out, m)
	}

	return out
}

// CommandPermission is "": a local command calls no API, so it needs no
// permission.
func (g *LocalGroup) CommandPermission(
	_ Command,
) client.Permission {
	return ""
}

// LocalMethod is one local command. Method makes one; its methods set the
// rest and return it, so a definition chains them.
type LocalMethod struct {
	// Name is the method's name in its group, and the command's last word.
	Name string
	// In is the input type: each field is a flag.
	In reflect.Type
	// Out is the output type, which prints as JSON with --json.
	Out reflect.Type

	sub, short, long string
	says             string
	process          bool
	unprinted        bool
	unlogged         bool
	keys             []string
	examples         []Example
	refusals         []Refusal
	handler          any
}

func (m *LocalMethod) apply(
	g *LocalGroup,
) {
	g.Methods = append(g.Methods, m)
}

// Method is a local command named name that takes In, whose fields are its
// flags, and prints Out.
func Method[In, Out any](
	name string,
) *LocalMethod {
	return &LocalMethod{Name: name, In: reflect.TypeFor[In](), Out: reflect.TypeFor[Out]()}
}

// Process is a local command named name that runs a process until the
// context ends, such as osapi nats start: it has no input and no
// output. Setup sets what it runs.
func Process(
	name string,
) *LocalMethod {
	return &LocalMethod{Name: name, In: reflect.TypeFor[struct{}](), process: true}
}

// Setup sets the function that makes a process's components:
// func(ctx context.Context, d D) (command.Running, error), with D the type
// the group Uses. Serve calls it with the context of the command, then runs
// what it returns until that context ends.
func (m *LocalMethod) Setup(
	fn any,
) *LocalMethod {
	m.handler = fn

	return m
}

// Flag adds a flag for each config key, such as agent.max_jobs, to a
// process: --agent-max-jobs, with the key's type, default and doc, which
// sets the key over the environment and the file.
func (m *LocalMethod) Flag(
	keys ...string,
) *LocalMethod {
	m.keys = append(m.keys, keys...)

	return m
}

// ConfigKeys are the config keys that Flag named.
func (m *LocalMethod) ConfigKeys() []string {
	return m.keys
}

// IsProcess is true for a method that Process made.
func (m *LocalMethod) IsProcess() bool {
	return m.process
}

// Short sets the command's one line of help.
func (m *LocalMethod) Short(
	text string,
) *LocalMethod {
	m.short = text

	return m
}

// Long sets the rest of the command's help.
func (m *LocalMethod) Long(
	text string,
) *LocalMethod {
	m.long = text

	return m
}

// Example adds an example of the command: its flags, such as --name web,
// and, when doc is given, what it shows, as a # line above it.
func (m *LocalMethod) Example(
	flags string,
	doc ...string,
) *LocalMethod {
	m.examples = append(m.examples, Example{Flags: flags, Doc: strings.Join(doc, " ")})

	return m
}

// Prints gives the last Example its output: what the command prints, a
// value of the method's output type. The docs page of the command shows it
// as the CLI prints it, and as JSON for --json. validate refuses a value of
// another type, Prints before an Example, and a method with output that no
// example Prints.
func (m *LocalMethod) Prints(
	out any,
) *LocalMethod {
	if len(m.examples) == 0 {
		m.unprinted = true

		return m
	}
	m.examples[len(m.examples)-1].Output = out

	return m
}

// Logs gives the last Example the lines that a process writes when it
// starts, at the info level or above. The docs page of the command shows
// them as its log handler writes them in a terminal, and as JSON for
// --json. validate refuses Logs before an Example, Logs on a method that is
// not a process, and a process that no example Logs.
func (m *LocalMethod) Logs(
	lines ...LogLine,
) *LocalMethod {
	if len(m.examples) == 0 {
		m.unlogged = true

		return m
	}
	m.examples[len(m.examples)-1].Logs = lines

	return m
}

// Sub puts the command under a command of its own in the group, such as
// roles in osapi controller role list.
func (m *LocalMethod) Sub(
	segment string,
) *LocalMethod {
	m.sub = segment

	return m
}

// Says is what the one line of a method with no output says, when it is a
// state and not the past tense of the method, such as draining for drain.
func (m *LocalMethod) Says(
	text string,
) *LocalMethod {
	m.says = text

	return m
}

// Refusal is an input that the command refuses with a usage error, as Fails
// declares it.
type Refusal struct {
	// Doc says what is wrong with the input, such as "an unknown role".
	Doc string
	// Input is the method's input, with one rule broken.
	Input any
}

// Fails declares an input that the command refuses with a usage error: the
// method's input with the other fields valid and one rule broken. Validate
// requires one for each rule of each input field, and a test runs each one.
func (m *LocalMethod) Fails(
	doc string,
	input any,
) *LocalMethod {
	m.refusals = append(m.refusals, Refusal{Doc: doc, Input: input})

	return m
}

// Refusals are the inputs Fails declared.
func (m *LocalMethod) Refusals() []Refusal {
	return m.refusals
}

// Handle sets the function that runs the command:
// func(ctx context.Context, d D, in In) (Out, error), with D the type the
// group Uses. Validate checks the signature; Run calls it.
func (m *LocalMethod) Handle(
	fn any,
) *LocalMethod {
	m.handler = fn

	return m
}

// Handler is the function Handle set, or nil.
func (m *LocalMethod) Handler() any {
	return m.handler
}

// MethodName is the method's identity in its group: its Sub and its Name,
// such as roles list, or its Name for a method with no Sub. Two methods of
// a group can have one Name under two Subs, such as get and spec get.
func (m *LocalMethod) MethodName() string {
	return strings.TrimSpace(m.sub + " " + m.Name)
}

// Input is the method's In.
func (m *LocalMethod) Input() reflect.Type {
	return m.In
}

// CommandName is the command's name, the method's.
func (m *LocalMethod) CommandName() string {
	return m.Name
}

// SubPath is the segment Sub set, or "".
func (m *LocalMethod) SubPath() string {
	return m.sub
}

// CommandShort is the one line Short set.
func (m *LocalMethod) CommandShort() string {
	return m.short
}

// CommandLong is the help Long set, or "".
func (m *LocalMethod) CommandLong() string {
	return m.long
}

// Examples are the examples Example added.
func (m *LocalMethod) Examples() []Example {
	return m.examples
}

// PageNotes is "": a local command's page has no Notes section.
func (m *LocalMethod) PageNotes() string {
	return ""
}

// PageResult is "": a local command's page has no Result section.
func (m *LocalMethod) PageResult() string {
	return ""
}

// Output is the method's Out, which its output prints from.
func (m *LocalMethod) Output() reflect.Type {
	return m.Out
}

// KeyField is "": a local command acts on no item of an API.
func (m *LocalMethod) KeyField() string {
	return ""
}

// Saying is the text Says set, or "".
func (m *LocalMethod) Saying() string {
	return m.says
}

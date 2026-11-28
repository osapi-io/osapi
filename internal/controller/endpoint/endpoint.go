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

// Package endpoint is the DSL for the controller's own API: the operations
// that run on the controller, such as the agents, the jobs and the audit
// log, and not on a host. A group's definition is Go; its OpenAPI part, and
// later its routes, CLI commands and tests, come from it. A behavior is work
// on a host and has its own DSL, internal/behavior.
package endpoint

import (
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/secret"
	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/osapi-io/osapi/internal/wiring"
)

// Group is a set of endpoints under one path, /api/<name>, such as agent.
type Group struct {
	// Name is the segment of each path after the prefix, and the CLI group.
	Name string
	// CommandNotes is the Notes section of the page of the group's command,
	// which CommandNotes sets.
	CommandNotes string
	// CommandLong is the rest of the help of the group's command, which
	// CommandLong sets.
	CommandLong string
	// SubDocs are the one line of help of each command that a Sub makes,
	// such as stale in osapi client file stale, which SubDoc sets.
	SubDocs map[string]string
	// Prefix is the part of each path between /api/ and the name, such as
	// next/node/{hostname}. Each {field} in it is a path field of each
	// method's input, besides the method's key.
	Prefix string
	// Doc says what the group is, in one line.
	Doc string
	// Area is the area of the group's operations, one of area.All, which
	// Area sets.
	Area area.Name
	// Title is the title of the group's spec. The full spec puts it in
	// front of each tag, and the UI's client files take their names from
	// the result, so a group that moves keeps the title it had.
	Title string
	// Tag is the tag of each operation, with its name in the docs and its
	// description.
	Tag, TagDisplay, TagDoc string
	// Deps is the type Uses names: what the group's handlers get from the
	// controller, such as the job client. Nil for a group that only has a
	// spec.
	Deps reflect.Type
	// Docs are the descriptions of the group's named types, which Doc sets:
	// Go has no tag on a type.
	Docs map[reflect.Type]string
	// Examples are the examples of fields of the group's named types, by
	// type and json name, which Example sets: for a value a tag cannot hold.
	Examples map[reflect.Type]map[string]string
	// Methods are the group's endpoints, in the order they were declared.
	Methods []*Method
	// Defaults are the group's own settings and their defaults, which
	// Config sets, or nil for a group with none.
	Defaults any
	// settings are the decoded settings, which SetSettings sets once the
	// config is read.
	settings any
}

// Option is a setting of a group, or one of its endpoints.
type Option interface {
	apply(g *Group)
}

type optionFunc func(g *Group)

func (f optionFunc) apply(
	g *Group,
) {
	f(g)
}

// Area puts the group's operations in an area, one of area.All: the folder
// of the SDK's page and its section of the docs site's SDK menu. endpoint validate refuses a group with none.
func Area(
	name area.Name,
) Option {
	return optionFunc(func(g *Group) { g.Area = name })
}

// Title sets the title of the group's spec.
func Title(
	title string,
) Option {
	return optionFunc(func(g *Group) { g.Title = title })
}

// Tag sets the tag of each operation of the group: its name, its name in
// the docs, and its description.
func Tag(
	name string,
	display string,
	doc string,
) Option {
	return optionFunc(func(g *Group) {
		g.Tag, g.TagDisplay, g.TagDoc = name, display, doc
	})
}

// Uses names D, the struct of what the group's handlers use, such as the
// job client. The controller gives one D at startup, and a test gives one
// with fakes. Each handler takes it as its second argument.
func Uses[D any]() Option {
	return optionFunc(func(g *Group) { g.Deps = reflect.TypeFor[D]() })
}

// Example gives the field of T with the json name field its example. A
// struct tag holds an example with no double quote, and Example holds the
// rest, such as a JSON document: gofumpt and golines drop the backslashes of
// a quote in a tag.
func Example[T any](
	field string,
	example string,
) Option {
	return optionFunc(func(g *Group) {
		if g.Examples == nil {
			g.Examples = map[reflect.Type]map[string]string{}
		}
		t := reflect.TypeFor[T]()
		if g.Examples[t] == nil {
			g.Examples[t] = map[string]string{}
		}
		g.Examples[t][field] = example
	})
}

// Doc describes T, a named type of the group, in its schema.
func Doc[T any](
	doc string,
) Option {
	return optionFunc(func(g *Group) {
		if g.Docs == nil {
			g.Docs = map[reflect.Type]string{}
		}
		g.Docs[reflect.TypeFor[T]()] = doc
	})
}

// Define builds a group from its options. Validate checks the result.
func Define(
	name string,
	doc string,
	opts ...Option,
) *Group {
	g := &Group{Name: name, Doc: doc, Title: name, Tag: name}
	for _, o := range opts {
		o.apply(g)
	}
	for _, m := range g.Methods {
		m.prefix = g.prefixFields()
	}

	return g
}

// Prefix puts each path of the group under a prefix, between /api/ and the
// group's name, such as next/node/{hostname} for a behavior. Each {field} of
// the prefix is a path field of each method's input, which is not its key.
func Prefix(
	path string,
) Option {
	return optionFunc(func(g *Group) { g.Prefix = strings.Trim(path, "/") })
}

// CommandLong is the rest of the help of the group's command, after its one
// line, such as what most of its commands read.
func CommandLong(
	text string,
) Option {
	return optionFunc(func(g *Group) { g.CommandLong = text })
}

// SubDoc is the one line of help of the command that the Sub segment makes,
// such as stale in osapi client file stale.
func SubDoc(
	segment string,
	short string,
) Option {
	return optionFunc(func(g *Group) {
		if g.SubDocs == nil {
			g.SubDocs = map[string]string{}
		}
		g.SubDocs[segment] = short
	})
}

// CommandNotes adds a Notes section to the page of the group's own command,
// such as a link to the page that tells more about the group.
func CommandNotes(
	text string,
) Option {
	return optionFunc(func(g *Group) { g.CommandNotes = text })
}

// CommandPath is the words of the group's command under osapi client: the
// fixed segments of its prefix, then its name. sysctl under the prefix
// next/node/{hostname} is next node sysctl.
func (g *Group) CommandPath() []string {
	var out []string
	for _, seg := range strings.Split(g.Prefix, "/") {
		if seg != "" && !strings.HasPrefix(seg, "{") {
			out = append(out, seg)
		}
	}

	return append(out, g.Name)
}

// prefixFields are the json names of the path fields that the prefix names.
func (g *Group) prefixFields() []string {
	var out []string
	for _, seg := range strings.Split(g.Prefix, "/") {
		if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
			out = append(out, seg[1:len(seg)-1])
		}
	}

	return out
}

// Path is the method's path: /api/, the group's prefix, its name, then its
// fixed sub-path, then its key, then its action, each only when it has one.
func (g *Group) Path(
	m *Method,
) string {
	p := "/api/" + g.Name
	if g.Prefix != "" {
		p = "/api/" + g.Prefix + "/" + g.Name
	}
	if m.sub != "" {
		p += "/" + m.sub
	}
	for _, key := range m.Keys() {
		p += "/{" + key + "}"
	}
	if m.action != "" {
		p += "/" + m.action
	}

	return p
}

// Permission is the permission a method needs: the one it names, or
// <group>:read for a GET and <group>:write for the other verbs. A public
// method needs none, and has "".
func (g *Group) Permission(
	m *Method,
) types.Permission {
	switch {
	case m.public:
		return ""
	case m.permission != "":
		return m.permission
	case m.verb == http.MethodGet:
		return types.NewPermission(g.Name, types.VerbRead)
	default:
		return types.NewPermission(g.Name, types.VerbWrite)
	}
}

// Declared is a permission that a method of a group needs: the permission,
// whether the method marks it admin only, and whether the method names it,
// with Permission, rather than takes its group's.
type Declared struct {
	Permission types.Permission
	AdminOnly  bool
	Named      bool
	// Where is the group and method, such as agent drain, for an error.
	Where string
}

// Declared is the permission of each method of g that needs one.
func (g *Group) Declared() []Declared {
	var out []Declared
	for _, m := range g.Methods {
		if m.public {
			continue
		}
		out = append(out, Declared{
			Permission: g.Permission(m),
			AdminOnly:  m.adminOnly,
			Named:      m.permission != "",
			Where:      g.Name + " " + m.Name,
		})
	}

	return out
}

// A group is a group of the command layer, and a method one of its
// commands: the CLI's commands are made from them.
var (
	_ command.Group   = (*Group)(nil)
	_ command.Command = (*Method)(nil)
)

// GroupName is the group's Name, for the command layer.
func (g *Group) GroupName() string {
	return g.Name
}

// GroupArea is the group's Area, for the command layer: the heading that
// its command is under in the help of its parent.
func (g *Group) GroupArea() area.Name {
	return g.Area
}

// GroupDoc is the group's Doc, for the command layer.
func (g *Group) GroupDoc() string {
	return g.Doc
}

// GroupLong is the help CommandLong set, or "", for the command layer.
func (g *Group) GroupLong() string {
	return g.CommandLong
}

// SubDoc is the help SubDoc set for the segment, or "", for the command
// layer.
func (g *Group) SubDoc(
	segment string,
) string {
	return g.SubDocs[segment]
}

// GroupNotes is the Notes section CommandNotes set, or "", for the command
// layer.
func (g *Group) GroupNotes() string {
	return g.CommandNotes
}

// Commands are the group's methods, for the command layer, but a method
// that NoCommand marks.
func (g *Group) Commands() []command.Command {
	out := make([]command.Command, 0, len(g.Methods))
	for _, m := range g.Methods {
		if m.noCommand {
			continue
		}
		out = append(out, m)
	}

	return out
}

// CommandPermission is the permission of method m, which is one of the
// group's methods, for the command layer.
func (g *Group) CommandPermission(
	m command.Command,
) types.Permission {
	return g.Permission(m.(*Method))
}

var (
	registryMu sync.RWMutex
	registry   = map[string]*Group{}
)

// Register adds g to the groups the controller has, and returns it. Two
// groups with one name are a mistake in the code: Register reports it to
// wiring and keeps the first.
func Register(
	g *Group,
) *Group {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, ok := registry[g.Name]; ok {
		wiring.Report(fmt.Errorf("endpoint: group %q registered twice", g.Name))

		return g
	}
	registry[g.Name] = g

	return g
}

// Lookup is the registered group with the name, or nil.
func Lookup(
	name string,
) *Group {
	registryMu.RLock()
	defer registryMu.RUnlock()

	return registry[name]
}

// All is each registered group, by name.
func All() []*Group {
	registryMu.RLock()
	defer registryMu.RUnlock()
	out := make([]*Group, 0, len(registry))
	for _, g := range registry {
		out = append(out, g)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })

	return out
}

// SecretFields are the JSON names of every secret.Value input field, at each
// depth, across the methods of every registered group: a behavior's group
// too. The audit log redacts each by name.
func SecretFields() map[string]bool {
	names := map[string]bool{}
	for _, g := range All() {
		for _, m := range g.Methods {
			for _, name := range secret.Fields(m.In) {
				names[name] = true
			}
		}
	}

	return names
}

// Method is one endpoint of a group. Get, Post, Put and Delete make one;
// its methods set the rest and return it, so a definition chains them.
type Method struct {
	// Name is the method's name in its group, such as list or drain.
	Name string
	// verb is the HTTP method, which Get, Post, Put and Delete set.
	verb string
	// In is the input type: a field tagged in:"path" is the key, a field
	// tagged in:"query" is a query parameter, and the others are the JSON
	// body.
	In reflect.Type
	// Out is the response type.
	Out reflect.Type

	sub, action, id  string
	keys             []string
	cmdSub           string
	summary, explain string
	returns          string
	permission       types.Permission
	public           bool
	// noCommand is a method with no CLI command, which NoCommand sets.
	noCommand     bool
	adminOnly     bool
	handler       any
	cmdShort      string
	location      []string
	cmdUse        string
	cmdLong       string
	examples      []CommandExample
	unprinted     bool
	unchecked     bool
	noCheck       string
	prefix        []string
	refusals      []Refusal
	notes, result string
	status        int
	says          string
	errors        []Error
	answers       []Answered
}

func (m *Method) apply(
	g *Group,
) {
	g.Methods = append(g.Methods, m)
}

func newMethod[In, Out any](
	verb string,
	name string,
) *Method {
	return &Method{
		Name:   name,
		verb:   verb,
		In:     reflect.TypeFor[In](),
		Out:    reflect.TypeFor[Out](),
		status: http.StatusOK,
	}
}

// Get is a GET endpoint that takes In and answers with Out.
func Get[In, Out any](
	name string,
) *Method {
	return newMethod[In, Out](http.MethodGet, name)
}

// Post is a POST endpoint that takes In and answers with Out.
func Post[In, Out any](
	name string,
) *Method {
	return newMethod[In, Out](http.MethodPost, name)
}

// Put is a PUT endpoint that takes In and answers with Out.
func Put[In, Out any](
	name string,
) *Method {
	return newMethod[In, Out](http.MethodPut, name)
}

// Delete is a DELETE endpoint that takes In and answers with Out.
func Delete[In, Out any](
	name string,
) *Method {
	return newMethod[In, Out](http.MethodDelete, name)
}

// Sub puts a fixed segment after the group's name, such as pending in
// /api/agent/pending, or facts in /api/agent/facts.
func (m *Method) Sub(
	segment string,
) *Method {
	m.sub = segment

	return m
}

// Key names the path fields of the key, in the order of the path, such as
// Key("vault", "key") for /api/secret/{vault}/{key}: a key that is unique
// only in the thing that holds it. Without it, the key is the one input
// field tagged in:"path" that the prefix does not name.
func (m *Method) Key(
	fields ...string,
) *Method {
	m.keys = fields

	return m
}

// Action puts a segment after the key, such as drain in
// /api/agent/{hostname}/drain.
func (m *Method) Action(
	segment string,
) *Method {
	m.action = segment

	return m
}

// ID sets the operationId. The Go SDK and the UI client take their names
// from it, so a group that moves keeps the IDs it had.
func (m *Method) ID(
	id string,
) *Method {
	m.id = id

	return m
}

// Describe sets the summary: what the method does, in one line.
func (m *Method) Describe(
	summary string,
) *Method {
	m.summary = summary

	return m
}

// Explain sets the description: the rest of what a caller needs to know.
func (m *Method) Explain(
	text string,
) *Method {
	m.explain = text

	return m
}

// Returns describes the response of a success, such as "List of active
// agents."
func (m *Method) Returns(
	doc string,
) *Method {
	m.returns = doc

	return m
}

// HTTPMethod is the method's HTTP method, such as GET. Get, Post, Put and
// Delete set it, so an author cannot name one that is not.
func (m *Method) HTTPMethod() string {
	return m.verb
}

// Permission names the permission the method needs, in place of the
// group's read or write permission.
func (m *Method) Permission(
	permission types.Permission,
) *Method {
	m.permission = permission

	return m
}

// AdminOnly makes the method's permission one that only the admin role
// grants, such as admission:write: write and read do not get it. Each method
// with one permission must agree. An execute or shell permission is admin
// only with no mark.
func (m *Method) AdminOnly() *Method {
	m.adminOnly = true

	return m
}

// NoCommand makes the method have no CLI command: a program calls it, not a
// person, such as the agent's join. Its route, its spec and its SDK method
// are as each other method's.
func (m *Method) NoCommand() *Method {
	m.noCommand = true

	return m
}

// Public makes the method need no auth: the health probes, and the agent's
// join, which the join token authorizes.
func (m *Method) Public() *Method {
	m.public = true

	return m
}

// Status sets the status code of a success, such as 201 or 204.
func (m *Method) Status(
	code int,
) *Method {
	m.status = code

	return m
}

// Says is what the CLI prints for a method that answers with no body, when
// it is a state and not the past tense of the method: the agent's key and
// draining, for drain.
func (m *Method) Says(
	text string,
) *Method {
	m.says = text

	return m
}

// Saying is the text Says set, or "", for the command layer.
func (m *Method) Saying() string {
	return m.says
}

// Output is the method's Out, which the command layer prints from.
func (m *Method) Output() reflect.Type {
	return m.Out
}

// Errors are the errors the method can return, each with its status code.
// An error that is not in the list is a 500.
func (m *Method) Errors(
	errs ...Error,
) *Method {
	m.errors = append(m.errors, errs...)

	return m
}

// Handle sets the function that runs the method:
// func(ctx context.Context, d D, in In) (Out, error), with D the type the
// group Uses. Validate checks the signature; the generated route calls it.
func (m *Method) Handle(
	fn any,
) *Method {
	m.handler = fn

	return m
}

// Command sets the CLI command's own help: its one line, and the rest. The
// API's summary and description are for a caller of the API; the CLI's are
// for a person at a shell.
func (m *Method) Command(
	short string,
	long string,
) *Method {
	m.cmdShort, m.cmdLong = short, long

	return m
}

// CommandExample is an example of a CLI command: its flags, and what it
// shows, which the help writes as a # line above it.
type CommandExample = command.Example

// Example adds an example of the CLI command: its flags, such as
// --hostname web-01, and, when doc is given, what it shows. The help shows
// each one after the command, under its doc as a # line.
func (m *Method) Example(
	flags string,
	doc ...string,
) *Method {
	m.examples = append(m.examples, CommandExample{Flags: flags, Doc: strings.Join(doc, " ")})

	return m
}

// Check gives the last Example a read that shows its result: the command
// words after the binary, such as client agent config get, its flags, and
// the paths and wants of the fields of its --json output, as command.Fields
// pairs them. The integration suite runs it after the example. endpoint
// validate refuses a Check before an Example.
func (m *Method) Check(
	words string,
	flags string,
	pathsAndWants ...any,
) *Method {
	c := command.ReadBack{Command: words, Flags: flags, Fields: command.Fields(pathsAndWants...)}
	if !command.WithCheck(m.examples, c) {
		m.unchecked = true
	}

	return m
}

// CheckFails gives the last Example a read that fails with errorText,
// such as not found: the example removed the thing that the read names.
func (m *Method) CheckFails(
	words string,
	flags string,
	errorText string,
) *Method {
	if !command.WithCheck(
		m.examples,
		command.ReadBack{Command: words, Flags: flags, Fails: errorText},
	) {
		m.unchecked = true
	}

	return m
}

// NoCheck says why no read shows the result of the method's examples, for
// a write whose result no command reads back. The integration gate requires
// a Check of each example of a write, or this.
func (m *Method) NoCheck(
	why string,
) *Method {
	m.noCheck = why

	return m
}

// Unchecked is why no read shows the result of the method's examples, as
// NoCheck says it, or "".
func (m *Method) Unchecked() string {
	return m.noCheck
}

// Prints gives the last Example its output: what the API answers, a value
// of the method's output type. The docs page of the command shows it as the
// CLI prints it, and as JSON for --json. endpoint validate refuses a value
// of another type, and Prints before an Example.
func (m *Method) Prints(
	out any,
) *Method {
	if len(m.examples) == 0 {
		m.unprinted = true

		return m
	}
	m.examples[len(m.examples)-1].Output = out

	return m
}

// Refusal is an input that the API refuses with 400, as Fails declares it.
type Refusal struct {
	// Doc says what is wrong with the input, such as "a name over 255
	// characters".
	Doc string
	// Input is the method's input, with one rule broken.
	Input any
}

// Fails declares an input that the API refuses with 400: the method's input
// with the other fields valid and one rule broken. Validate requires one for
// each rule of each input field, and a test sends each one through the
// routes.
func (m *Method) Fails(
	doc string,
	input any,
) *Method {
	m.refusals = append(m.refusals, Refusal{Doc: doc, Input: input})

	return m
}

// Refusals are the inputs Fails declared.
func (m *Method) Refusals() []Refusal {
	return m.refusals
}

// Notes adds a Notes section to the CLI command's page: what a person needs
// to know that the help does not say, such as what --json has that the
// table does not.
func (m *Method) Notes(
	text string,
) *Method {
	m.notes = text

	return m
}

// Result adds a Result section to the CLI command's page: what the command
// does in a case that needs words, such as when two agents share a name.
func (m *Method) Result(
	text string,
) *Method {
	m.result = text

	return m
}

// Use names the CLI command apart from the method, whose name must be
// unique in the group: such as list under pending, when the group has a
// list already.
func (m *Method) Use(
	name string,
) *Method {
	m.cmdUse = name

	return m
}

// MethodName is the method's Name, for the command layer.
func (m *Method) MethodName() string {
	return m.Name
}

// Input is the method's In, for the command layer.
func (m *Method) Input() reflect.Type {
	return m.In
}

// CommandName is the CLI command's name: the one Use set, or the method's.
func (m *Method) CommandName() string {
	if m.cmdUse != "" {
		return m.cmdUse
	}

	return m.Name
}

// Location names the fields of the output that hold the ID of the
// resource a 201 created, such as job_id: the answer's Location header is
// the group's path and the value of each field, joined by /, such as
// /api/job/7c9e6679-... or /api/secret/default/registry-token.
func (m *Method) Location(
	fields ...string,
) *Method {
	m.location = fields

	return m
}

// LocationFields are the fields Location named, or none.
func (m *Method) LocationFields() []string {
	return m.location
}

// SubPath is the word that the CLI command is under: the one CommandSub
// set, or the fixed segment Sub set, or "".
func (m *Method) SubPath() string {
	if m.cmdSub != "" {
		return m.cmdSub
	}

	return m.sub
}

// CommandSub puts the CLI command under a word that is not in the path,
// such as vault in osapi client secret vault list for GET /api/secret.
func (m *Method) CommandSub(
	word string,
) *Method {
	m.cmdSub = word

	return m
}

// PageNotes is the Notes section Notes set, or "".
func (m *Method) PageNotes() string {
	return m.notes
}

// PageResult is the Result section Result set, or "".
func (m *Method) PageResult() string {
	return m.result
}

// CommandShort is the CLI command's one line: the one Command set, or the
// summary.
func (m *Method) CommandShort() string {
	if m.cmdShort != "" {
		return m.cmdShort
	}

	return m.summary
}

// CommandLong is the rest of the CLI command's help: the one Command set,
// or the description.
func (m *Method) CommandLong() string {
	if m.cmdLong != "" {
		return m.cmdLong
	}

	return m.explain
}

// Examples are the CLI command's examples, the flags and the doc of each.
func (m *Method) Examples() []CommandExample {
	return m.examples
}

// Mapped are the errors Errors set, each with its status code.
func (m *Method) Mapped() []Error {
	return m.errors
}

// Answered is a status a method answers with its own output, which Answers
// sets: such as a probe's 503, which says what is not ready.
type Answered struct {
	// Code is the status.
	Code int
	// Doc is its description in the spec.
	Doc string
}

// Answers adds a status that the method answers with its output, not with
// an error: such as the 503 of a probe, whose body says what failed. The
// method returns its output with Answer(code).
func (m *Method) Answers(
	code int,
	doc string,
) *Method {
	m.answers = append(m.answers, Answered{Code: code, Doc: doc})

	return m
}

// Answering is the statuses Answers added.
func (m *Method) Answering() []Answered {
	return m.answers
}

// AnswerError is the error Answer makes: the status to send the output
// with.
type AnswerError struct {
	// Code is the status.
	Code int
}

// Error is the message of the error.
func (e *AnswerError) Error() string {
	return http.StatusText(e.Code)
}

// Answer is the error a method returns, with its output, to answer code
// with that output: a status that Answers added.
func Answer(
	code int,
) error {
	return &AnswerError{Code: code}
}

// SuccessStatus is the status of a success: the one Status set, or 200.
func (m *Method) SuccessStatus() int {
	return m.status
}

// Method is the method of g with the name, or nil.
func (g *Group) Method(
	name string,
) *Method {
	for _, m := range g.Methods {
		if m.Name == name {
			return m
		}
	}

	return nil
}

// Handler is the function Handle set, or nil.
func (m *Method) Handler() any {
	return m.handler
}

// Keys are the json names of the path fields of the key, in the order of
// the path: the ones Key named, or the one input field tagged in:"path"
// that the prefix does not name. It is empty when the method has no key.
func (m *Method) Keys() []string {
	if len(m.keys) > 0 {
		return m.keys
	}
	for _, f := range fields(m.In) {
		if f.Tag.Get("in") == "path" && !slices.Contains(m.prefix, jsonName(f)) {
			return []string{jsonName(f)}
		}
	}

	return nil
}

// KeyField is the json name of the last field of the key, which names what
// the method acts on, such as key in /api/secret/{vault}/{key}, or "".
func (m *Method) KeyField() string {
	keys := m.Keys()
	if len(keys) == 0 {
		return ""
	}

	return keys[len(keys)-1]
}

// OperationID is the operationId: the one ID sets, or the verb and the
// path in Pascal case, as each operationId of the API is.
func (m *Method) OperationID(
	g *Group,
) string {
	if m.id != "" {
		return m.id
	}

	return upper(strings.ToLower(m.verb)) + upper(g.Name) + upper(m.sub) + upper(m.Name)
}

// Error is an error a method can return, and the status code it maps to.
type Error struct {
	// Err is the error, matched with errors.Is.
	Err error
	// Code is the status code.
	Code int
	// Reason is the word of the answer's reason field, one of the values of types.Reason.
	Reason types.Reason
	// Doc says when it occurs, for the spec.
	Doc string
}

// ReasonDocs say when each reason is the answer, with its status. The
// ErrorResponse schema of the spec has them.
var ReasonDocs = map[types.Reason]string{
	types.ReasonInvalidRequest: "the request breaks a rule, such as a missing field (400).",
	types.ReasonUnauthorized:   "no token, or a token that is not valid (401).",
	types.ReasonForbidden:      "the token does not have the permission (403).",
	types.ReasonNotFound:       "the host or osapi does not have what the request names (404).",
	types.ReasonNotManaged:     "the item is there, but osapi did not deploy it, so osapi does not change it (409).",
	types.ReasonNotInstalled:   "the host does not have a program that the request needs (409).",
	types.ReasonConflict:       "the state of the item blocks the request, such as a create of one that exists (409).",
	types.ReasonUnavailable:    "a part that osapi depends on is down, such as NATS (503).",
	types.ReasonServerError:    "the controller failed; its log has the cause (500).",
}

// ErrorResponse is the body of each answer that is not a success: the
// ErrorResponse schema of the spec.
type ErrorResponse struct {
	Error string `json:"error"`
	// Reason is why, as one of the values of types.Reason.
	Reason types.Reason `json:"reason"`
}

// ReasonFor is the reason of a status that no error names: the reason of
// a 404 that no route matches, or a 405.
func ReasonFor(
	status int,
) types.Reason {
	switch {
	case status == http.StatusUnauthorized:
		return types.ReasonUnauthorized
	case status == http.StatusForbidden:
		return types.ReasonForbidden
	case status == http.StatusNotFound:
		return types.ReasonNotFound
	case status == http.StatusConflict:
		return types.ReasonConflict
	case status == http.StatusServiceUnavailable:
		return types.ReasonUnavailable
	case status >= http.StatusInternalServerError:
		return types.ReasonServerError
	default:
		return types.ReasonInvalidRequest
	}
}

// Because is e with reason instead of the reason of its status, such as
// not_managed for a 409 that osapi did not deploy.
func (e Error) Because(
	reason types.Reason,
) Error {
	e.Reason = reason

	return e
}

// Errorf is an error whose text is the message, and that errors.Is
// matches to sentinel. A handler returns it to answer with the status that
// Errors maps sentinel to, and with a message that names what it was
// called with, such as no pending agent with hostname "web-01".
func Errorf(
	sentinel error,
	format string,
	args ...any,
) error {
	return &wrapped{msg: fmt.Sprintf(format, args...), sentinel: sentinel}
}

type wrapped struct {
	msg      string
	sentinel error
}

// Error is the message of the error.
func (w *wrapped) Error() string {
	return w.msg
}

// Unwrap is the error that this one wraps.
func (w *wrapped) Unwrap() error {
	return w.sentinel
}

// Internal is a failure of the controller whose cause stays out of the
// answer: the caller gets msg, a 500, and the log gets the cause, such as
// the address of a backend that could not be read.
func Internal(
	msg string,
	cause error,
) error {
	return &InternalError{Msg: msg, Cause: cause}
}

// InternalError is the error Internal makes.
type InternalError struct {
	// Msg is what the caller gets.
	Msg string
	// Cause is what the log gets.
	Cause error
}

// Error is the message of the error.
func (e *InternalError) Error() string {
	return e.Msg
}

// Unwrap is the error that this one wraps.
func (e *InternalError) Unwrap() error {
	return e.Cause
}

// Invalid maps err to 400: a request that is well formed, but that the
// method cannot do, such as a retry of a job with nothing to run again.
func Invalid(
	err error,
	doc string,
) Error {
	return Error{
		Err:    err,
		Code:   http.StatusBadRequest,
		Reason: types.ReasonInvalidRequest,
		Doc:    doc,
	}
}

// Forbidden maps err to 403: a request that the token may make, but not
// on this thing, such as a write to a file that osapi owns.
func Forbidden(
	err error,
	doc string,
) Error {
	return Error{Err: err, Code: http.StatusForbidden, Reason: types.ReasonForbidden, Doc: doc}
}

// NotFound maps err to 404.
func NotFound(
	err error,
	doc string,
) Error {
	return Error{Err: err, Code: http.StatusNotFound, Reason: types.ReasonNotFound, Doc: doc}
}

// Conflict maps err to 409.
func Conflict(
	err error,
	doc string,
) Error {
	return Error{Err: err, Code: http.StatusConflict, Reason: types.ReasonConflict, Doc: doc}
}

// Unavailable maps err to 503.
func Unavailable(
	err error,
	doc string,
) Error {
	return Error{
		Err:    err,
		Code:   http.StatusServiceUnavailable,
		Reason: types.ReasonUnavailable,
		Doc:    doc,
	}
}

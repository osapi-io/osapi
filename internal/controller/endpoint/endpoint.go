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
	"sort"
	"sync"
)

// Group is a set of endpoints under one path, /api/<name>, such as agent.
type Group struct {
	// Name is the first segment of each path and the CLI group.
	Name string
	// Doc says what the group is, in one line.
	Doc string
	// Title is the title of the group's spec. The joined spec puts it in
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

	return g
}

// Path is the method's path: /api/<group>, then its fixed sub-path, then
// its key, then its action, each only when it has one.
func (g *Group) Path(
	m *Method,
) string {
	p := "/api/" + g.Name
	if m.sub != "" {
		p += "/" + m.sub
	}
	if key := m.Key(); key != "" {
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
) string {
	switch {
	case m.public:
		return ""
	case m.permission != "":
		return m.permission
	case m.Verb == http.MethodGet:
		return g.Name + ":read"
	default:
		return g.Name + ":write"
	}
}

var (
	registryMu sync.RWMutex
	registry   = map[string]*Group{}
)

// Register adds g to the groups the controller has, and returns it. Two
// groups with one name are a mistake in the code, so it panics.
func Register(
	g *Group,
) *Group {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, ok := registry[g.Name]; ok {
		panic(fmt.Sprintf("endpoint: group %q registered twice", g.Name))
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

// Method is one endpoint of a group. Get, Post, Put and Delete make one;
// its methods set the rest and return it, so a definition chains them.
type Method struct {
	// Name is the method's name in its group, such as list or drain.
	Name string
	// Verb is the HTTP method.
	Verb string
	// In is the input type: a field tagged in:"path" is the key, a field
	// tagged in:"query" is a query parameter, and the others are the JSON
	// body.
	In reflect.Type
	// Out is the response type.
	Out reflect.Type

	sub, action, id  string
	summary, explain string
	returns          string
	permission       string
	public           bool
	handler          any
	cmdShort         string
	cmdLong          string
	examples         []string
	notes, result    string
	flags            []Flag
	status           int
	errors           []Error
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
		Verb:   verb,
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
// /api/agent/pending, or trace in /api/telemetry/trace/{id}.
func (m *Method) Sub(
	segment string,
) *Method {
	m.sub = segment

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

// Permission names the permission the method needs, in place of the
// group's read or write permission.
func (m *Method) Permission(
	permission string,
) *Method {
	m.permission = permission

	return m
}

// Public makes the method need no auth. Only the health probes are public.
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

// Example adds an example of the CLI command: its flags, such as
// --hostname web-01. The help shows each one after the command.
func (m *Method) Example(
	flags string,
) *Method {
	m.examples = append(m.examples, flags)

	return m
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

// Flag is a flag of the CLI command that is not part of the request: it
// changes only how the answer prints, such as --all. The renderer reads it.
type Flag struct {
	// Name is the flag's name, without the dashes.
	Name string
	// Usage is its help.
	Usage string
	// String is true for a string flag, which StringFlag adds; false for a
	// bool flag.
	String bool
	// Value is a string flag's value when the caller gives none.
	Value string
	// Required is true for a string flag the caller must give.
	Required bool
}

// Flag adds a bool flag to the CLI command that the request does not send,
// for the renderer: such as --all, to show what it leaves out.
func (m *Method) Flag(
	name string,
	usage string,
) *Method {
	m.flags = append(m.flags, Flag{Name: name, Usage: usage})

	return m
}

// StringFlag adds a string flag to the CLI command that the request does
// not send, for the renderer: such as --output, the file an export writes.
// value is its value when the caller gives none.
func (m *Method) StringFlag(
	name string,
	value string,
	usage string,
	required bool,
) *Method {
	m.flags = append(m.flags, Flag{
		Name:     name,
		Usage:    usage,
		String:   true,
		Value:    value,
		Required: required,
	})

	return m
}

// CommandFlags are the flags Flag and StringFlag added.
func (m *Method) CommandFlags() []Flag {
	return m.flags
}

// SubPath is the fixed segment Sub set, or "".
func (m *Method) SubPath() string {
	return m.sub
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

// Examples are the CLI command's examples, the flags of each.
func (m *Method) Examples() []string {
	return m.examples
}

// Mapped are the errors Errors set, each with its status code.
func (m *Method) Mapped() []Error {
	return m.errors
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

// Key is the name of the input's path parameter, or "" when it has none.
func (m *Method) Key() string {
	for _, f := range fields(m.In) {
		if f.Tag.Get("in") == "path" {
			return jsonName(f)
		}
	}

	return ""
}

// OperationID is the operationId: the one ID sets, or the verb and the
// path in camel case.
func (m *Method) OperationID(
	g *Group,
) string {
	if m.id != "" {
		return m.id
	}

	return lowerVerb(m.Verb) + upper(g.Name) + upper(m.sub) + upper(m.Name)
}

// Error is an error a method can return, and the status code it maps to.
type Error struct {
	// Err is the error, matched with errors.Is.
	Err error
	// Code is the status code.
	Code int
	// Doc says when it occurs, for the spec.
	Doc string
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

func (w *wrapped) Error() string {
	return w.msg
}

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

func (e *InternalError) Error() string {
	return e.Msg
}

func (e *InternalError) Unwrap() error {
	return e.Cause
}

// Invalid maps err to 400: a request that is well formed, but that the
// method cannot do, such as a retry of a job with nothing to run again.
func Invalid(
	err error,
	doc string,
) Error {
	return Error{Err: err, Code: http.StatusBadRequest, Doc: doc}
}

// NotFound maps err to 404.
func NotFound(
	err error,
	doc string,
) Error {
	return Error{Err: err, Code: http.StatusNotFound, Doc: doc}
}

// Conflict maps err to 409.
func Conflict(
	err error,
	doc string,
) Error {
	return Error{Err: err, Code: http.StatusConflict, Doc: doc}
}

// Unavailable maps err to 503.
func Unavailable(
	err error,
	doc string,
) Error {
	return Error{Err: err, Code: http.StatusServiceUnavailable, Doc: doc}
}

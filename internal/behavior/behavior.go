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
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// Placement says where a behavior's work happens.
type Placement string

// Node is the only placement: the work runs on a host, through the job
// system. What the controller answers itself is not a behavior.
const NodePlacement Placement = "node"

// Verb says whether a method reads the host or changes it. It decides the HTTP
// method, the permission, and whether the result carries changed.
type Verb string

const (
	// VerbRead reads the host and changes nothing.
	VerbRead Verb = "read"
	// VerbWrite changes the host. Its output reports whether it did.
	VerbWrite Verb = "write"
)

// Behavior is one area of a host that osapi manages, such as kernel
// parameters or DNS resolver settings.
type Behavior struct {
	// Name is the behavior's name: the path segment, the CLI command and the
	// permission prefix when it has no group.
	Name string
	// Doc says what the behavior manages, in one line.
	Doc string
	// Placement is where the work runs.
	Placement Placement
	// Group nests the behavior under another path and command, such as
	// network for dns. Empty sits it directly under node.
	Group string
	// API serves the behavior over REST: routes, the OpenAPI spec and its
	// API docs. Without it the behavior is reachable only through another
	// behavior or the job system.
	API bool
	// CLI gives the behavior client commands and CLI docs pages. The
	// commands call the API, so CLI needs API.
	CLI bool
	// Requires are the behaviors this one is built on. The agent builds each
	// one's provider first and hands it to this one's builder, and this
	// one's methods call theirs as Go calls on the same host.
	Requires []*Behavior
	// Internal marks a behavior only other behaviors use: it has no route,
	// no command, no spec and no permissions.
	Internal bool
	// Methods are the operations, in the order they were declared.
	Methods []*Method
	// Defaults is the behavior's settings struct as declared with Config,
	// its field values the defaults. Nil when the behavior has none.
	Defaults any

	// settings is the decoded settings struct, set by the config loader.
	settings any
	// Platforms are how the behavior's provider is built on each platform it
	// supports, in the order declared. On any other platform its methods are
	// skipped.
	Platforms []*PlatformProvider
	// Sudo lists the programs the behavior's providers run with sudo, which
	// the agent's preflight checks it may run.
	Sudo []string
	// Runs lists the programs its providers run without sudo, declared with
	// Runs.
	Runs []string
}

// Option configures a behavior in Define.
type Option interface {
	apply(b *Behavior)
}

type optionFunc func(b *Behavior)

func (f optionFunc) apply(
	b *Behavior,
) {
	f(b)
}

// Node places a behavior on the hosts, through the job system.
var Node Option = optionFunc(func(b *Behavior) { b.Placement = NodePlacement })

// API serves a behavior over REST: its routes, its part of the OpenAPI
// spec, and its API reference pages.
var API Option = optionFunc(func(b *Behavior) { b.API = true })

// CLI gives a behavior client commands, generated into cmd/next, and CLI
// docs pages. It needs API, because the commands call it.
var CLI Option = optionFunc(func(b *Behavior) { b.CLI = true })

// Under nests a behavior inside a group: Under("network") puts dns at
// /api/node/{hostname}/network/dns and `osapi client node network dns`.
func Under(
	group string,
) Option {
	return optionFunc(func(b *Behavior) { b.Group = group })
}

// Define declares a behavior. Methods are options too, so a definition reads
// top to bottom as name, description, placement and methods.
func Define(
	name string,
	doc string,
	opts ...Option,
) *Behavior {
	b := &Behavior{Name: name, Doc: doc}
	for _, o := range opts {
		o.apply(b)
	}
	for _, m := range b.Methods {
		m.owner = b
	}

	return b
}

// Permission is the permission a method needs: the top-level group, or the
// behavior's name without one, and the verb. sysctl get needs sysctl:read and
// dns under network needs network:read.
func (b *Behavior) Permission(
	v Verb,
) string {
	prefix := b.Name
	if b.Group != "" {
		prefix = b.Group
	}

	return prefix + ":" + string(v)
}

// Method returns the method with the given name, or nil.
func (b *Behavior) Method(
	name string,
) *Method {
	for _, m := range b.Methods {
		if m.Name == name {
			return m
		}
	}

	return nil
}

// Operation is the job operation name the agent dispatches on, such as
// sysctl.get.
func (b *Behavior) Operation(
	m *Method,
) string {
	return b.Name + "." + m.Name
}

// Method is one operation of a behavior.
type Method struct {
	// Name is the method's name: the CLI subcommand and the operation suffix.
	Name string
	// Doc says what the method does, in one line. It is the command's short
	// help, the operation's summary, and the docs page's first line.
	Doc string
	// Long is what a reader needs beyond Doc: what the method touches on the
	// host and how it behaves when run twice. It is the command's long help
	// and the body of its docs page. Optional.
	Long string
	// Verb is read or write.
	Verb Verb
	// KeyField is the input field that goes in the path, by its JSON name.
	// Empty means the method addresses the whole collection.
	KeyField string
	// In and Out are the method's input and output types. An output that is a
	// slice becomes one result row per element, each with the host's name.
	In  reflect.Type
	Out reflect.Type
	// Examples are sample calls for the docs pages.
	Examples []Example
	// Refusals are calls the method fails, for the docs pages and the
	// provider's tests.
	Refusals []Refusal

	provider reflect.Type
	owner    *Behavior
	// invoke decodes the input, calls the method, and returns the decoded
	// input beside the output so telemetry can read its tagged fields.
	invoke func(ctx context.Context, provider any, input []byte) (out any, in any, err error)
}

func (m *Method) apply(
	b *Behavior,
) {
	b.Methods = append(b.Methods, m)
}

// Key names the input field, by its JSON name, that is addressed in the path.
func (m *Method) Key(
	field string,
) *Method {
	m.KeyField = field

	return m
}

// Explain sets the method's longer description, for its command's help and
// its docs page.
func (m *Method) Explain(
	long string,
) *Method {
	m.Long = long

	return m
}

// Describe sets the method's one-line description.
func (m *Method) Describe(
	doc string,
) *Method {
	m.Doc = doc

	return m
}

// ProviderType is the type the method's function takes as its provider.
func (m *Method) ProviderType() reflect.Type {
	return m.provider
}

// Invoke decodes input as the method's input type and calls it with
// provider, inside the span and metrics every behavior gets (see
// telemetry.go). The agent calls this; nothing else should.
func (m *Method) Invoke(
	ctx context.Context,
	provider any,
	input []byte,
) (any, error) {
	return m.observe(ctx, provider, input)
}

// Read declares a method that reads the host.
func Read[P any, In any, Out any](
	name string,
	fn func(ctx context.Context, p P, in In) (Out, error),
) *Method {
	return newMethod(name, VerbRead, fn)
}

// Write declares a method that changes the host. Its output type reports
// whether anything changed, through a Changed bool field tagged
// json:"changed"; behavior validate fails a write without one.
func Write[P any, In any, Out any](
	name string,
	fn func(ctx context.Context, p P, in In) (Out, error),
) *Method {
	return newMethod(name, VerbWrite, fn)
}

func newMethod[P any, In any, Out any](
	name string,
	verb Verb,
	fn func(ctx context.Context, p P, in In) (Out, error),
) *Method {
	return &Method{
		Name:     name,
		Verb:     verb,
		In:       reflect.TypeFor[In](),
		Out:      reflect.TypeFor[Out](),
		provider: reflect.TypeFor[P](),
		invoke: func(ctx context.Context, provider any, input []byte) (any, any, error) {
			p, ok := provider.(P)
			if !ok {
				return nil, nil, fmt.Errorf(
					"method %s: provider is %T, want %s",
					name, provider, reflect.TypeFor[P](),
				)
			}

			var in In
			if len(input) > 0 {
				if err := json.Unmarshal(input, &in); err != nil {
					return nil, nil, fmt.Errorf("method %s: decode input: %w", name, err)
				}
			}

			out, err := fn(ctx, p, in)

			return out, in, err
		},
	}
}

// HTTPMethod is the method's HTTP verb: GET for a read, and for a write POST,
// PUT or DELETE by the method's name, POST for any other write.
func (m *Method) HTTPMethod() string {
	if m.Verb == VerbRead {
		return "GET"
	}

	switch m.Name {
	case "update":
		return "PUT"
	case "delete":
		return "DELETE"
	default:
		return "POST"
	}
}

// RoutePrefix is where behaviors are served while they run beside the
// hand-written routes: Path's /api/node/... is served at /api/next/node/...
const RoutePrefix = "/api/next"

// CommandPrefix is where behaviors' commands sit while they run beside the
// hand-written ones: osapi client next node sysctl get.
const CommandPrefix = "osapi client next"

// CLICommand is the command a method is run with today, under CommandPrefix.
// Command is where it moves at the cutover.
func (b *Behavior) CLICommand(
	m *Method,
) string {
	return strings.Replace(b.Command(m), "osapi client", CommandPrefix, 1)
}

// Route is the path a method is served at, under RoutePrefix.
func (b *Behavior) Route(
	m *Method,
) string {
	return RoutePrefix + strings.TrimPrefix(b.Path(m), "/api")
}

// Path is the method's route after the cutover: /api/node/{hostname}, the
// group and name, and the key when the method addresses one item.
func (b *Behavior) Path(
	m *Method,
) string {
	p := "/api/node/{hostname}/"
	if b.Group != "" {
		p += b.Group + "/"
	}
	p += b.Name
	if m.KeyField != "" {
		p += "/{" + m.KeyField + "}"
	}

	return p
}

// Command is the method's CLI command, such as
// osapi client node network dns get.
func (b *Behavior) Command(
	m *Method,
) string {
	c := "osapi client node "
	if b.Group != "" {
		c += b.Group + " "
	}

	return c + b.Name + " " + m.Name
}

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

// Package behavior defines what osapi can do to a host. A behavior is one
// directory: a definition, typed inputs and outputs, and the logic for each
// platform. The API routes, the agent's dispatch, the CLI commands, the
// OpenAPI document and the docs pages are all built from the definition.
package behavior

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
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
	// Methods are the operations, in the order they were declared.
	Methods []*Method
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

// Example is one sample call of a method, used for the docs pages.
type Example struct {
	// Doc says what the example shows.
	Doc string
	// Input is a value of the method's input type.
	Input any
	// Output is a value of the method's output type, as one host would
	// return it.
	Output any
}

// Method is one operation of a behavior.
type Method struct {
	// Name is the method's name: the CLI subcommand and the operation suffix.
	Name string
	// Doc says what the method does, in one line.
	Doc string
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

	provider reflect.Type
	invoke   func(ctx context.Context, provider any, input []byte) (any, error)
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

// Describe sets the method's one-line description.
func (m *Method) Describe(
	doc string,
) *Method {
	m.Doc = doc

	return m
}

// Example adds a sample call for the docs pages.
func (m *Method) Example(
	doc string,
	input any,
	output any,
) *Method {
	m.Examples = append(m.Examples, Example{Doc: doc, Input: input, Output: output})

	return m
}

// ProviderType is the type the method's function takes as its provider.
func (m *Method) ProviderType() reflect.Type {
	return m.provider
}

// Invoke decodes input as the method's input type and calls it with
// provider. The agent calls this; nothing else should.
func (m *Method) Invoke(
	ctx context.Context,
	provider any,
	input []byte,
) (any, error) {
	return m.invoke(ctx, provider, input)
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
		invoke: func(ctx context.Context, provider any, input []byte) (any, error) {
			p, ok := provider.(P)
			if !ok {
				return nil, fmt.Errorf(
					"method %s: provider is %T, want %s",
					name, provider, reflect.TypeFor[P](),
				)
			}

			var in In
			if len(input) > 0 {
				if err := json.Unmarshal(input, &in); err != nil {
					return nil, fmt.Errorf("method %s: decode input: %w", name, err)
				}
			}

			return fn(ctx, p, in)
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

// Path is the method's route: /api/node/{hostname}, the group and name, and
// the key when the method addresses one item.
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

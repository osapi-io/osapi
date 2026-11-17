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

// Package command is the runtime of the generated CLI commands of the
// endpoint groups. Each group's gen/cli.go makes one cobra command for each
// method and registers it here; Attach puts them under osapi client
// <group>. Run builds the method's request from its typed input, sends it
// through the SDK, and prints the answer: the body for --json, or the
// group's renderer for the method.
package command

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/pkg/sdk/client"
)

// Caller sends one request to the API. The SDK client satisfies it.
type Caller interface {
	Request(
		ctx context.Context,
		method string,
		path string,
		body []byte,
	) (*client.Response[json.RawMessage], error)
}

// Deps is what the commands need from the CLI.
type Deps struct {
	// Caller returns the SDK client.
	Caller func() Caller
	// JSON reports whether --json was given.
	JSON func() bool
	// Error reports a failed call.
	Error func(err error)
}

type entry struct {
	group *endpoint.Group
	cmds  []*cobra.Command
}

var (
	registered []entry
	deps       Deps
	renderers  = map[string]renderer{}
)

// renderer prints a method's answer: fn gets the input and a value of out,
// the type the answer decodes into.
type renderer struct {
	fn  func(in, out any)
	out reflect.Type
}

// Register adds a group's generated commands. Each gen/cli.go calls it from
// init.
func Register(
	g *endpoint.Group,
	cmds ...*cobra.Command,
) {
	registered = append(registered, entry{group: g, cmds: cmds})
}

// Attach puts each registered group's commands under parent's command of
// the group's name, which it makes when parent has none, and keeps d for
// Run. A generated command with the name of one that is there already is a
// command written twice, so it is an error.
func Attach(
	parent *cobra.Command,
	d Deps,
) error {
	deps = d
	for _, e := range registered {
		group := child(parent, e.group.Name)
		if group == nil {
			group = &cobra.Command{Use: e.group.Name, Short: e.group.Doc}
			parent.AddCommand(group)
		}
		for _, c := range e.cmds {
			if child(group, c.Name()) != nil {
				return fmt.Errorf(
					"osapi client %s %s is both a hand-written command and the %s group's",
					e.group.Name, c.Name(), e.group.Name)
			}
			group.AddCommand(c)
		}
	}

	return nil
}

func child(
	parent *cobra.Command,
	name string,
) *cobra.Command {
	for _, c := range parent.Commands() {
		if c.Name() == name {
			return c
		}
	}

	return nil
}

// Render sets how a method's answer prints without --json. fn gets the
// input the command sent and the output the API answered.
func Render[In, Out any](
	g *endpoint.Group,
	method string,
	fn func(in In, out Out),
) {
	renderers[g.Name+"."+method] = renderer{
		fn:  func(in, out any) { fn(in.(In), out.(Out)) },
		out: reflect.TypeFor[Out](),
	}
}

// Run sends method m of g with in, and prints the answer. The answer
// decodes into the type of the method's renderer, so the generated command
// does not name its output type.
func Run[In any](
	cmd *cobra.Command,
	g *endpoint.Group,
	method string,
	in In,
) {
	m := g.Method(method)
	path, body, err := Build(g, m, in)
	if err != nil {
		deps.Error(err)

		return
	}

	resp, err := deps.Caller().Request(cmd.Context(), m.Verb, path, body)
	if err != nil {
		deps.Error(err)

		return
	}
	if deps.JSON() {
		cli.PrintJSON(resp.RawJSON())

		return
	}

	render, ok := renderers[g.Name+"."+method]
	if !ok {
		cli.PrintJSON(resp.RawJSON())

		return
	}
	out := reflect.New(render.out)
	if len(resp.Data) > 0 {
		if err := json.Unmarshal(resp.Data, out.Interface()); err != nil {
			deps.Error(fmt.Errorf("decode %s %s: %w", g.Name, method, err))

			return
		}
	}
	render.fn(in, out.Elem().Interface())
}

// Build is the request that sends in to method m of g: the key in the path,
// the query fields in the query, and the other fields as the JSON body.
func Build(
	g *endpoint.Group,
	m *endpoint.Method,
	in any,
) (string, []byte, error) {
	v := reflect.ValueOf(in)
	path := g.Path(m)
	query := url.Values{}
	body := map[string]any{}
	var whole any

	if v.Kind() == reflect.Struct {
		for i := range v.NumField() {
			f := v.Type().Field(i)
			if !f.IsExported() {
				continue
			}
			name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
			if name == "" {
				name = f.Name
			}
			if name == "-" {
				continue
			}
			value := v.Field(i)
			switch f.Tag.Get("in") {
			case "path":
				path = strings.Replace(path, "{"+name+"}",
					url.PathEscape(fmt.Sprint(value.Interface())), 1)
			case "query":
				if !value.IsZero() {
					query.Set(name, fmt.Sprint(value.Interface()))
				}
			case "body":
				whole = value.Interface()
			default:
				body[name] = value.Interface()
			}
		}
	}
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	switch {
	case whole != nil:
		data, err := json.Marshal(whole)

		return path, data, err
	case len(body) > 0:
		data, err := json.Marshal(body)

		return path, data, err
	}

	return path, nil, nil
}

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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/url"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/cli/reference"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/pkg/sdk/client"
)

// Caller sends one request to the API. The SDK client satisfies it.
type Caller interface {
	Send(
		ctx context.Context,
		method string,
		path string,
		contentType string,
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
	fn  func(cmd *cobra.Command, in, out any)
	out reflect.Type
	// acts is true for a renderer that Acts set, which runs with --json too.
	acts bool
}

// SubKey is the annotation of a generated command that goes under a
// command of its own: the Sub segment of its method, such as trace in
// osapi client telemetry trace get.
const SubKey = "osapi.endpoint.sub"

// Register adds a group's generated commands. Each gen/cli.go calls it from
// init.
func Register(
	g *endpoint.Group,
	cmds ...*cobra.Command,
) {
	registered = append(registered, entry{group: g, cmds: cmds})
}

// Attach adds every registered group's commands under parent, osapi client,
// and keeps d for the commands to call through. A group goes under the
// commands of its command path: next node sysctl is under next, then node,
// which Attach adds when parent does not have them. A group or a command
// that is among groups, by name, is that one, so the generated commands
// join a hand-written command of the same name, such as agent with key and
// status.
func Attach(
	parent *cobra.Command,
	d Deps,
	groups ...*cobra.Command,
) error {
	deps = d
	for _, e := range registered {
		words := e.group.CommandPath()
		into := parent
		for _, w := range words[:len(words)-1] {
			next := child(into, w)
			if next == nil {
				next = &cobra.Command{Use: w, Short: strings.ToUpper(w[:1]) + w[1:] + " commands"}
				into.AddCommand(next)
			}
			into = next
		}
		group := child(into, e.group.Name)
		for _, g := range groups {
			if g.Name() == e.group.Name {
				group = g
			}
		}
		if group == nil {
			group = &cobra.Command{Use: e.group.Name, Short: e.group.Doc}
			if e.group.CommandNotes != "" {
				group.Annotations = map[string]string{reference.Notes: e.group.CommandNotes}
			}
			into.AddCommand(group)
		}
		for _, c := range e.cmds {
			parent := group
			if sub := c.Annotations[SubKey]; sub != "" {
				parent = child(group, sub)
				for _, g := range groups {
					if g.Name() == sub && g != group {
						parent = g
					}
				}
				if parent == nil {
					parent = &cobra.Command{
						Use:   sub,
						Short: strings.ToUpper(sub[:1]) + sub[1:] + " commands",
					}
					group.AddCommand(parent)
				}
			}
			if child(parent, c.Name()) != nil {
				return fmt.Errorf(
					"osapi client %s %s is both a hand-written command and the %s group's: "+
						"delete the hand-written one from cmd/",
					strings.Join(words, " "), c.Name(), e.group.Name)
			}
			parent.AddCommand(c)
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
// command, for a flag that Flag added, the input the command sent and the
// output the API answered.
func Render[In, Out any](
	g *endpoint.Group,
	method string,
	fn func(cmd *cobra.Command, in In, out Out),
) {
	renderers[g.Name+"."+method] = renderer{
		fn:  func(cmd *cobra.Command, in, out any) { fn(cmd, in.(In), out.(Out)) },
		out: reflect.TypeFor[Out](),
	}
}

// Acts sets what a command does with a method's answer, with or without
// --json: such as export, which writes the answer to a file. fn prints its
// own output, and reads --json from the command when it prints differently.
func Acts[In, Out any](
	g *endpoint.Group,
	method string,
	fn func(cmd *cobra.Command, in In, out Out),
) {
	Render(g, method, fn)
	r := renderers[g.Name+"."+method]
	r.acts = true
	renderers[g.Name+"."+method] = r
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
	path, body, contentType, err := Build(g, m, in)
	if err != nil {
		deps.Error(err)

		return
	}

	resp, err := deps.Caller().Send(cmd.Context(), m.Verb, path, contentType, body)
	data, ok := answered(m, err)
	if err != nil && !ok {
		deps.Error(err)

		return
	}
	if ok {
		resp = client.NewResponse(data, []byte(data))
	}
	traceID = resp.TraceID()
	render, ok := renderers[g.Name+"."+method]
	if deps.JSON() && !render.acts {
		cli.PrintJSON(resp.RawJSON())

		return
	}
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
	render.fn(cmd, in, out.Elem().Interface())
}

// Build is the request that sends in to method m of g: the key in the path,
// the query fields in the query, and the other fields as the JSON body.
func Build(
	g *endpoint.Group,
	m *endpoint.Method,
	in any,
) (string, []byte, string, error) {
	v := reflect.ValueOf(in)
	path := g.Path(m)
	query := url.Values{}
	body := map[string]any{}
	form := &bytes.Buffer{}
	parts := multipart.NewWriter(form)
	hasForm := false
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
			case "form":
				hasForm = true
				formPart(parts, name, value)
			default:
				body[name] = value.Interface()
			}
		}
	}
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	switch {
	case hasForm:
		_ = parts.Close()

		return path, form.Bytes(), parts.FormDataContentType(), nil
	case whole != nil:
		data, err := json.Marshal(whole)

		return path, data, jsonType, err
	case len(body) > 0:
		data, err := json.Marshal(body)

		return path, data, jsonType, err
	}

	return path, nil, jsonType, nil
}

// jsonType is the content type of a JSON body.
const jsonType = "application/json"

// formPart writes a field of a form: a []byte as the file of that name,
// and each other value as text. A zero value is not sent, so the server
// applies its default.
func formPart(
	parts *multipart.Writer,
	name string,
	value reflect.Value,
) {
	if value.IsZero() {
		return
	}
	if data, ok := value.Interface().([]byte); ok {
		file, _ := parts.CreateFormFile(name, name)
		_, _ = file.Write(data)

		return
	}
	_ = parts.WriteField(name, fmt.Sprint(value.Interface()))
}

// answered is the body of an error answer whose status the method
// .Answers with its output, such as a probe's 503: it prints as an answer.
func answered(
	m *endpoint.Method,
	err error,
) (json.RawMessage, bool) {
	var a interface{ Answer() (int, []byte) }
	if !errors.As(err, &a) {
		return nil, false
	}
	code, body := a.Answer()
	for _, ans := range m.Answering() {
		if ans.Code == code {
			return body, true
		}
	}

	return nil, false
}

// ReadFile is the content of the local file at path, for a flag of a []byte
// field. A file that it cannot read is reported, and ok is false.
func ReadFile(
	path string,
) ([]byte, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		deps.Error(fmt.Errorf("read %s: %w", path, err))

		return nil, false
	}

	return data, true
}

// traceID is the trace of the call Run made last. A command runs one call,
// so a renderer reads it with TraceID.
var traceID string

// TraceID is the trace of the call that the running command made, from the
// controller's traceresponse header, for a renderer that prints it, as each
// behavior's result table does. It is "" when the answer had none.
func TraceID() string {
	return traceID
}

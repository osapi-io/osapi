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

package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/dave/jennifer/jen"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/command"
	"github.com/osapi-io/osapi/internal/cli/reference"
)

const (
	cliCommandPath = module + "/internal/cli/command"
	areaPath       = module + "/internal/area"
	endpointPath   = module + "/internal/controller/endpoint"
	behaviorPkg    = module + "/internal/behavior"
	dispatchPath   = module + "/internal/behavior/dispatch"
)

// hostnameDoc is the description of the target of each behavior method.
const hostnameDoc = "Target: _all, _any, hostname, or label (group:web.dev)."

// jobID is the example of a response's job_id.
const jobID = "550e8400-e29b-41d4-a716-446655440000"

// groupFile is the endpoint group of b: for each method, a request type with
// the target and the method's input, the response with a row for each host,
// and the method of the group. The endpoint layer makes b's spec from it.
func groupFile(
	b *behavior.Behavior,
) (*jen.File, error) {
	f := genFile(b)
	f.ImportName(endpointPath, "endpoint")
	f.ImportName(behaviorPkg, "behavior")
	f.ImportName(dispatchPath, "dispatch")
	f.ImportAlias(cliCommandPath, "clicommand")

	prefix, err := groupPrefix(b)
	if err != nil {
		return nil, err
	}

	written := map[string]bool{}
	var methods []jen.Code
	for _, m := range b.Methods {
		req := title(m.Name) + "Request"
		resp := responseName(b, m)

		f.Comment(
			fmt.Sprintf(
				"%s is the request of %s %s: its target and its input.",
				req,
				b.Name,
				m.Name,
			),
		)
		f.Type().Id(req).Struct(requestFields(b, m, f)...)
		f.Line()

		row, rowName := rowType(b, m)
		if !written[rowName] {
			written[rowName] = true
			f.Comment(fmt.Sprintf("%s is the result of %s on one host.", rowName, b.Name))
			f.Type().Id(rowName).Struct(row...)
			f.Line()
		}
		if !written[resp] {
			written[resp] = true
			f.Comment(fmt.Sprintf("%s is each target host's result.", resp))
			f.Type().Id(resp).Struct(
				jen.Id("JobID").String().Tag(map[string]string{
					"json": "job_id,omitempty", "validate": "uuid", "example": jobID,
					"doc": "The job ID used to process this request.",
				}),
				jen.Id("Results").Index().Id(rowName).Tag(map[string]string{"json": "results"}),
			)
			f.Line()
		}

		ok := "Each target host's result."
		if m.Verb != behavior.VerbRead {
			ok = "Each target host's result, with whether it changed."
		}
		verb := map[string]string{"GET": "Get", "POST": "Post", "PUT": "Put", "DELETE": "Delete"}[m.HTTPMethod()]
		handle := "Handle" + title(m.Name)
		f.Comment(
			fmt.Sprintf(
				"%s runs %s %s on the target, for the endpoint server.",
				handle,
				b.Name,
				m.Name,
			),
		)
		f.Func().Id(handle).Add(params(
			jen.Id("ctx").Qual("context", "Context"),
			jen.Id("d").Qual(dispatchPath, "Deps"),
			jen.Id("in").Id(req),
		)).Params(jen.Id(resp), jen.Error()).Block(
			jen.Return(jen.Qual(dispatchPath, "Run").Types(jen.Id(resp)).Call(
				jen.Id("ctx"), jen.Id("d"), jen.Qual(pkgPath(b), "Behavior"), jen.Lit(m.Name),
				jen.Id("in").Dot("Hostname"), input(m),
			)),
		)
		f.Line()

		def := jen.Qual(endpointPath, verb).Types(jen.Id(req), jen.Id(resp)).Call(jen.Lit(m.Name)).
			Op(".").Line().Id("Handle").Call(jen.Id(handle)).
			Op(".").Line().Id("ID").Call(jen.Lit(b.OperationID(m))).
			Op(".").Line().Id("Describe").Call(jen.Lit(m.Doc)).
			Op(".").Line().Id("Permission").Call(jen.Lit(string(b.Permission(m.PermissionVerb())))).
			Op(".").Line().Id("Returns").Call(jen.Lit(ok))
		if m.Action {
			def = def.Op(".").Line().Id("Action").Call(jen.Lit(m.Name))
		}
		if m.Long != "" {
			def = def.Op(".").Line().Id("Explain").Call(jen.Lit(m.Long))
		}
		// The CLI command's help: each example, with what it shows, and the
		// tables of what it does on the host and what it refuses.
		for _, ex := range m.Examples {
			def = def.Op(".").Line().Id("Example").Call(
				jen.Lit(command.ExampleFlags(ex.Input)), jen.Lit(ex.Doc))
		}
		if r := command.Annotations(b, m)[reference.Result]; r != "" {
			def = def.Op(".").Line().Id("Result").Call(jen.Lit(r))
		}
		// What a provider returns, as the API answers it on one host.
		missing := "The host does not have what the request names, such as the object of a deploy."
		if m.KeyField != "" {
			missing = fmt.Sprintf("The host has no %s with that %s.", b.Name, m.KeyField)
		}
		conflict := "osapi did not deploy the item, so it does not change it, " +
			"or the host does not have a program that the method needs."
		invalid := "The request is not valid, or the host refuses the input, " +
			"such as a path that is a directory."
		if len(behavior.TemplateFields(m.In)) > 0 {
			invalid = "The request is not valid, a template reads a var that the " +
				"behavior does not have, or the host refuses the input."
		}
		def = def.Op(".").Line().Id("Errors").Call(
			jen.Qual(endpointPath, "NotFound").
				Call(jen.Qual(behaviorPkg, "ErrNotFound"), jen.Lit(missing)),
			jen.Qual(endpointPath, "Conflict").
				Call(jen.Qual(behaviorPkg, "ErrNotManaged"), jen.Lit(conflict)).
				Dot("Because").
				Call(jen.Qual("github.com/osapi-io/osapi/pkg/sdk/types", "ReasonNotManaged")),
			jen.Qual(endpointPath, "Conflict").
				Call(jen.Qual(behaviorPkg, "ErrNotInstalled"), jen.Lit(conflict)).
				Dot("Because").
				Call(jen.Qual("github.com/osapi-io/osapi/pkg/sdk/types", "ReasonNotInstalled")),
			jen.Qual(endpointPath, "Invalid").
				Call(jen.Qual(dispatchPath, "ErrInvalid"), jen.Lit(invalid)),
		)
		methods = append(methods, def)
	}

	opts := append([]jen.Code{
		jen.Lit(b.Name), jen.Lit(b.Doc),
		// The behavior's area is the constant of package area that its
		// definition names, such as area.Hardware.
		jen.Qual(endpointPath, "Area").Call(jen.Qual(areaPath, title(string(b.Area)))),
		jen.Qual(endpointPath, "Prefix").Call(jen.Lit(prefix)),
		jen.Qual(endpointPath, "Uses").Index(jen.Qual(dispatchPath, "Deps")).Call(),
		jen.Qual(endpointPath, "Tag").Call(
			jen.Lit(b.Tag()),
			jen.Lit(b.Title), jen.Lit(b.Doc)),
		jen.Qual(endpointPath, "CommandNotes").Call(
			jen.Lit(command.GroupAnnotations(b)[reference.Notes])),
	}, methods...)

	f.Comment("Group is " + b.Name + " as an endpoint group, which its spec comes from.")
	f.Var().Id("Group").Op("=").Qual(endpointPath, "Define").Add(multi("(", ")", opts...))
	f.Line()
	// The behavior runtime prints each method's answer as the result
	// table, so the group registers no printer of its own.
	inits := []jen.Code{
		jen.Qual(behaviorPkg, "RegisterGroup").Call(jen.Lit(b.Name), jen.Id("Group")),
	}
	f.Func().Id("init").Params().Block(inits...)

	return f, nil
}

// groupPrefix is the part of b's paths between /api/ and its name, such as
// next/node/{hostname}. Each method's route must be the endpoint path.
func groupPrefix(
	b *behavior.Behavior,
) (string, error) {
	route := b.Path(b.Methods[0])
	i := strings.LastIndex(route, "/"+b.Name)
	if i < 0 {
		return "", fmt.Errorf("route %s does not have /%s", route, b.Name)
	}

	return strings.TrimPrefix(route[:i], "/api/"), nil
}

// requestFields are the fields of a method's request: the target, the key,
// and the input as query fields for a read or a body for a write.
func requestFields(
	b *behavior.Behavior,
	m *behavior.Method,
	f *jen.File,
) []jen.Code {
	out := []jen.Code{
		jen.Id("Hostname").String().Tag(map[string]string{
			"json": "hostname", "in": "path", "param": "Hostname",
			"validate": "required,min=1,valid_target", "doc": hostnameDoc,
			"flag": "target", "short": "T", "default": "_all",
		}),
	}

	var body []jen.Code
	for i := range m.In.NumField() {
		fld := m.In.Field(i)
		name := behavior.JSONName(fld)
		if !fld.IsExported() || name == "-" || behavior.SetByBefore(fld) {
			continue
		}
		tags := map[string]string{"json": fld.Tag.Get("json")}
		for _, k := range []string{"validate", "doc", "example", "default", "pattern"} {
			if v, ok := fld.Tag.Lookup(k); ok {
				tags[k] = v
			}
		}
		field := func() *jen.Statement { return jen.Id(fld.Name).Add(typeOf(fld.Type)) }
		switch {
		case m.InPath(name):
			tags["in"] = "path"
			tags["param"] = b.TypeName() + title(name)
			out = append(out, field().Tag(tags))
		case m.Verb == behavior.VerbRead:
			tags["in"] = "query"
			out = append(out, field().Tag(tags))
		default:
			body = append(body, field().Tag(tags))
		}
	}
	if len(body) > 0 {
		name := b.TypeName() + title(m.Name) + "Request"
		f.Comment(fmt.Sprintf("%s is the body of %s %s.", name, b.Name, m.Name))
		f.Type().Id(name).Struct(body...)
		f.Line()
		out = append(out, jen.Id("Body").Id(name).Tag(map[string]string{
			"in": "body", "validate": "required",
		}))
	}

	return out
}

// rowType is the fields and the name of a method's row: the host, its
// status, the method's output fields and the error.
func rowType(
	b *behavior.Behavior,
	m *behavior.Method,
) ([]jen.Code, string) {
	elem := m.Out
	for elem.Kind() == reflect.Slice || elem.Kind() == reflect.Pointer {
		elem = elem.Elem()
	}

	fields := []jen.Code{
		jen.Id("Hostname").String().Tag(map[string]string{
			"json": "hostname", "doc": "Hostname of the agent that produced this result.",
		}),
		jen.Id("Status").String().Tag(map[string]string{
			"json": "status", "validate": "oneof=ok failed skipped timeout",
			"doc": "The status of the operation for this host.",
		}),
	}
	for i := range elem.NumField() {
		fld := elem.Field(i)
		name := behavior.JSONName(fld)
		if !fld.IsExported() || name == "-" || name == "hostname" || name == "status" ||
			name == "error" {
			continue
		}
		tags := map[string]string{"json": name + ",omitempty"}
		if v, ok := fld.Tag.Lookup("doc"); ok {
			tags["doc"] = v
		}
		fields = append(fields, jen.Id(fld.Name).Add(optional(fld.Type)).Tag(tags))
	}
	// A pointer keeps a value that the host sent as it was, a false or an
	// empty string included, where omitempty would drop it; the schema is
	// the same.
	fields = append(fields, jen.Id("Error").Op("*").String().Tag(map[string]string{
		"json": "error,omitempty", "doc": "Error message if the agent failed.",
	}))

	return fields, b.TypeName() + elem.Name()
}

// responseName is the name of a method's response: the collection of a
// list, and <Behavior><Method>Response for each other method.
func responseName(
	b *behavior.Behavior,
	m *behavior.Method,
) string {
	if m.Name == "list" {
		return b.TypeName() + "CollectionResponse"
	}

	return b.TypeName() + title(m.Name) + "Response"
}

// optional is t as the type of an optional row field: a pointer, so a zero
// value that the host sent is kept.
func optional(
	t reflect.Type,
) *jen.Statement {
	if t.Kind() == reflect.Pointer {
		return typeOf(t)
	}

	return jen.Op("*").Add(typeOf(t))
}

// input is the method's input, made from its request: the key from the
// path, and the rest from the query of a read or the body of a write.
func input(
	m *behavior.Method,
) *jen.Statement {
	var values []jen.Code
	for i := range m.In.NumField() {
		fld := m.In.Field(i)
		name := behavior.JSONName(fld)
		if !fld.IsExported() || name == "-" || behavior.SetByBefore(fld) {
			continue
		}
		from := jen.Id("in").Dot(fld.Name)
		if !m.InPath(name) && m.Verb != behavior.VerbRead {
			from = jen.Id("in").Dot("Body").Dot(fld.Name)
		}
		values = append(values, jen.Id(fld.Name).Op(":").Add(from))
	}

	return jen.Qual(m.In.PkgPath(), m.In.Name()).Values(values...)
}

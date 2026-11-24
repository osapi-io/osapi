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

package endpointdsl

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/osapi-io/osapi/internal/behavior"

	"gopkg.in/yaml.v3"

	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/cli/commands/behaviordsl"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
)

// List is each endpoint group, with the names of its methods.
func List(
	_ context.Context,
	d Deps,
	_ ListInput,
) (ListOutput, error) {
	out := ListOutput{Items: []Summary{}}
	for _, g := range d.Groups() {
		names := make([]string, 0, len(g.Methods))
		for _, m := range g.Methods {
			names = append(names, m.Name)
		}
		out.Items = append(
			out.Items,
			Summary{Name: g.Name, Area: g.Area, Methods: names, Doc: g.Doc},
		)
	}

	return out, nil
}

// find is the group with the name, or an error that says to list them.
func find(
	d Deps,
	name string,
) (*endpoint.Group, error) {
	for _, g := range d.Groups() {
		if g.Name == name {
			return g, nil
		}
	}

	return nil, fmt.Errorf(
		"no endpoint group named %q: osapi dsl endpoint list shows them",
		name,
	)
}

// Get is the group with the name, and the route, permission and
// operation of each method.
func Get(
	_ context.Context,
	d Deps,
	in GetInput,
) (GetOutput, error) {
	g, err := find(d, in.Name)
	if err != nil {
		return GetOutput{}, err
	}
	out := GetOutput{
		Name: g.Name, Doc: g.Doc, Area: g.Area, Methods: []Method{}, Types: []Type{},
	}
	for _, m := range g.Methods {
		out.Methods = append(out.Methods, Method{
			Name: m.Name, Verb: m.HTTPMethod(), Path: g.Path(m),
			Permission: string(g.Permission(m)), OperationID: m.OperationID(g),
		})
		out.Types = append(out.Types, Type{
			Name:   m.Name,
			Input:  behavior.FieldShape(m.In, true),
			Output: behavior.FieldShape(m.Out, false),
		})
	}

	return out, nil
}

// OpenAPIGet is the group's part of the OpenAPI spec, as YAML.
func OpenAPIGet(
	_ context.Context,
	d Deps,
	in GetInput,
) (OpenAPIOutput, error) {
	g, err := find(d, in.Name)
	if err != nil {
		return OpenAPIOutput{}, err
	}
	doc, err := d.OpenAPI(g)
	if err != nil {
		return OpenAPIOutput{}, fmt.Errorf("the spec of %s does not build: %w", g.Name, err)
	}
	// The spec goes through JSON, so its keys are its json names.
	raw, _ := json.Marshal(doc)
	var tree any
	_ = json.Unmarshal(raw, &tree)
	var b strings.Builder
	enc := yaml.NewEncoder(&b)
	enc.SetIndent(2)
	_ = enc.Encode(tree)

	return OpenAPIOutput{Name: g.Name, YAML: b.String()}, nil
}

// MethodGet is one method of the group in full: its route, permission,
// command, input fields, output, examples, refusals and errors.
func MethodGet(
	_ context.Context,
	d Deps,
	in MethodInput,
) (MethodOutput, error) {
	g, err := find(d, in.Name)
	if err != nil {
		return MethodOutput{}, err
	}
	var m *endpoint.Method
	for _, candidate := range g.Methods {
		if candidate.Name == in.Method {
			m = candidate
		}
	}
	if m == nil {
		return MethodOutput{}, fmt.Errorf(
			"endpoint group %s has no method %q: osapi dsl endpoint get --name %s shows them",
			g.Name,
			in.Method,
			g.Name,
		)
	}
	input, err := behavior.SchemaFor(m.In)
	if err != nil {
		return MethodOutput{}, fmt.Errorf("%s %s input: %w", g.Name, m.Name, err)
	}
	output, err := behavior.SchemaFor(m.Out)
	if err != nil {
		return MethodOutput{}, fmt.Errorf("%s %s output: %w", g.Name, m.Name, err)
	}
	words := append([]string{"osapi", "client"}, g.CommandPath()...)
	if sub := m.SubPath(); sub != "" && sub != m.CommandName() {
		words = append(words, sub)
	}
	out := MethodOutput{
		Name: m.Name, Doc: m.CommandShort(), Long: m.CommandLong(),
		HTTPMethod: m.HTTPMethod(), Route: g.Path(m), Permission: string(g.Permission(m)),
		OperationID: m.OperationID(g), Command: strings.Join(append(words, m.CommandName()), " "),
		InputType: behavior.TypeName(m.In), OutputType: behavior.TypeName(m.Out),
		Fields: []Field{}, Results: behaviordsl.Results(output), Examples: []Example{},
		Input: input, Output: output,
	}
	for _, dc := range g.Declared() {
		if dc.Where == g.Name+" "+m.Name {
			out.AdminOnly = dc.AdminOnly
		}
	}
	where := map[string]string{}
	for _, f := range fields(m.In) {
		in := f.Tag.Get("in")
		if in == "" {
			in = "body"
		}
		where[jsonName(f)] = in
	}
	for _, f := range behaviordsl.Fields(input) {
		// A field tagged in:"body" that is a struct is the whole body, so
		// its fields are the fields of the request.
		if body := input.Properties[f.Field]; where[f.Field] == "body" && len(body.Order) > 0 {
			for _, bf := range behaviordsl.Fields(body) {
				out.Fields = append(out.Fields, Field{
					Field: bf.Field, In: "body", Type: bf.Type, Required: bf.Required,
					Constraints: bf.Constraints, Description: bf.Description,
				})
			}

			continue
		}
		out.Fields = append(out.Fields, Field{
			Field: f.Field, In: where[f.Field], Type: f.Type, Required: f.Required,
			Constraints: f.Constraints, Description: f.Description,
		})
	}
	for _, e := range m.Examples() {
		out.Examples = append(out.Examples, Example{Flags: e.Flags, Doc: e.Doc})
	}
	for _, r := range m.Refusals() {
		out.Refuses = append(out.Refuses, r.Doc)
	}
	for _, e := range m.Mapped() {
		out.Errors = append(out.Errors, Error{Code: e.Code, Reason: string(e.Reason), Doc: e.Doc})
	}

	return out, nil
}

// fields are the exported fields of t, a struct, that JSON reads. An
// endpoint method's input is a struct: its schema is built first.
func fields(
	t reflect.Type,
) []reflect.StructField {
	var out []reflect.StructField
	for f := range t.Fields() {
		if f.IsExported() && jsonName(f) != "-" {
			out = append(out, f)
		}
	}

	return out
}

// jsonName is the json name of f, or its Go name when it has none.
func jsonName(
	f reflect.StructField,
) string {
	name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
	if name == "" {
		return f.Name
	}

	return name
}

// Create writes a new endpoint group in the checkout, and is its files and
// the steps after it.
func Create(
	_ context.Context,
	d Deps,
	in CreateInput,
) (CreateOutput, error) {
	files, err := endpoint.Scaffold(d.Dir, in.Name, in.Description)
	if err != nil {
		return CreateOutput{}, err
	}
	out := CreateOutput{Name: in.Name}
	for _, f := range files {
		out.Files = append(out.Files, File{Path: f})
	}
	for _, step := range []string{
		"go generate ./internal/controller/endpoints/all",
		"mise exec -- just generate",
		"go test ./internal/controller/endpoints/" + in.Name + "/...",
		"go run . dsl endpoint validate",
	} {
		out.Next = append(out.Next, Step{Command: step})
	}

	return out, nil
}

// Spec is each part of the endpoint DSL, by kind.
func Spec(
	_ context.Context,
	_ Deps,
	_ SpecInput,
) (SpecOutput, error) {
	return SpecOutput{Kinds: endpoint.Spec.Sections()}, nil
}

// Validate checks each endpoint group. When a rule fails, it is the report
// with command.Fail: the report prints, then the command exits 1.
func Validate(
	_ context.Context,
	d Deps,
	_ ValidateInput,
) (ValidateOutput, error) {
	groups := d.Groups()
	out := ValidateOutput{Groups: len(groups)}
	for _, g := range groups {
		out.Problems = append(out.Problems, endpoint.Validate(g)...)
	}
	out.Valid = len(out.Problems) == 0
	if out.Valid {
		return out, nil
	}
	out.Fixes = map[string]any{}
	for _, p := range out.Problems {
		if p.Fix != "" {
			out.Fixes[p.Rule] = p.Fix
		}
	}

	return out, command.Fail()
}

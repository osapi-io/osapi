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
	"strings"

	"github.com/osapi-io/osapi/internal/behavior"

	"gopkg.in/yaml.v3"

	"github.com/osapi-io/osapi/internal/cli/command"
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
		out.Items = append(out.Items, Summary{Name: g.Name, Methods: names, Doc: g.Doc})
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
		"no endpoint group named %q: osapi controller endpoint list shows them",
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
	out := GetOutput{Name: g.Name, Doc: g.Doc, Methods: []Method{}, Types: []Type{}}
	for _, m := range g.Methods {
		out.Methods = append(out.Methods, Method{
			Name: m.Name, Verb: m.Verb, Path: g.Path(m),
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
		"go run . controller endpoint validate",
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

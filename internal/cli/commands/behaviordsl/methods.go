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

package behaviordsl

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/telemetry/instruments"
	"gopkg.in/yaml.v3"
)

// List is each behavior, with the names of its methods.
func List(
	_ context.Context,
	d Deps,
	_ ListInput,
) (ListOutput, error) {
	out := ListOutput{Items: []Summary{}}
	for _, b := range d.Behaviors() {
		names := make([]string, 0, len(b.Methods))
		for _, m := range b.Methods {
			names = append(names, m.Name)
		}
		out.Items = append(out.Items, Summary{
			Name: b.Name, Group: b.Group, Area: b.Area, Methods: names,
			Doc: b.Doc, Internal: b.Internal,
		})
	}

	return out, nil
}

// find is the behavior with the name, or an error that says to list them.
func find(
	d Deps,
	name string,
) (*behavior.Behavior, error) {
	for _, b := range d.Behaviors() {
		if b.Name == name {
			return b, nil
		}
	}

	return nil, fmt.Errorf("no behavior named %q: osapi behavior list shows them", name)
}

// Get is the behavior with the name.
func Get(
	_ context.Context,
	d Deps,
	in GetInput,
) (GetOutput, error) {
	b, err := find(d, in.Name)
	if err != nil {
		return GetOutput{}, err
	}
	out := GetOutput{
		Name: b.Name, Doc: b.Doc, Placement: b.Placement, Group: b.Group, Area: b.Area,
		Internal: b.Internal, API: b.API, CLI: b.CLI,
		Platforms:  platformNames(b),
		Sudo:       sudoOf(b),
		Runs:       append([]string{}, b.Runs...),
		Writes:     append([]string{}, b.Writes...),
		KeepsState: b.KeepsState,
		Uses:       names(b.Requires),
		UsedBy:     names(behavior.RequiredBy(b)),
		Methods:    []Method{},
		Types:      []Type{},
		Settings:   []Setting{},
		Telemetry: Telemetry{
			Span:       "behavior." + b.Name + ".<method>",
			ChildSpans: append([]string{}, behavior.ChildSpans(b)...),
			Metrics: []string{
				instruments.BehaviorCalls.Exported(),
				instruments.BehaviorDuration.Exported(),
				instruments.BehaviorChanges.Exported(),
			},
			Attributes: []string{
				string(
					behavior.AttrBehavior,
				),
				string(behavior.AttrMethod),
				string(behavior.AttrVerb),
				string(behavior.AttrOutcome),
				string(behavior.AttrChanged),
			},
			Inputs: map[string][]string{},
		},
	}
	for _, m := range b.Methods {
		in, err := behavior.SchemaFor(m.In)
		if err != nil {
			return GetOutput{}, fmt.Errorf("%s %s input: %w", b.Name, m.Name, err)
		}
		res, err := behavior.SchemaFor(m.Out)
		if err != nil {
			return GetOutput{}, fmt.Errorf("%s %s output: %w", b.Name, m.Name, err)
		}
		if recorded := m.SpanInputs(); len(recorded) > 0 {
			out.Telemetry.Inputs[m.Name] = recorded
		}
		row := Method{Name: m.Name, Verb: m.Verb, Doc: m.Doc}
		if b.API {
			row.HTTPMethod, row.Route, row.Permission = m.HTTPMethod(), b.Path(
				m,
			), string(b.Permission(
				m.Verb,
			))
		}
		if b.CLI {
			row.Command = b.Command(m)
		}
		out.Methods = append(out.Methods, row)
		out.Types = append(out.Types, Type{
			Name:   m.Name,
			Input:  behavior.Shape(m.In, in, true),
			Output: behavior.Shape(m.Out, res, false),
		})
	}
	out.Telemetry.Recorded = recordedInputs(out.Telemetry.Inputs, b.Methods)
	for _, k := range d.Keys() {
		if k.Behavior != b.Name {
			continue
		}
		def := ""
		if k.Default != nil {
			def = fmt.Sprint(k.Default)
		}
		out.Settings = append(
			out.Settings,
			Setting{Key: k.Name, Flag: k.Flag, Type: k.Type, Default: def},
		)
	}

	return out, nil
}

// MethodGet is one method of the behavior in full.
func MethodGet(
	_ context.Context,
	d Deps,
	in MethodInput,
) (MethodOutput, error) {
	b, err := find(d, in.Name)
	if err != nil {
		return MethodOutput{}, err
	}
	m := b.Method(in.Method)
	if m == nil {
		return MethodOutput{}, fmt.Errorf(
			"behavior %s has no method %q: osapi behavior get --name %s shows them",
			b.Name,
			in.Method,
			b.Name,
		)
	}
	input, err := behavior.SchemaFor(m.In)
	if err != nil {
		return MethodOutput{}, fmt.Errorf("%s %s input: %w", b.Name, m.Name, err)
	}
	output, err := behavior.SchemaFor(m.Out)
	if err != nil {
		return MethodOutput{}, fmt.Errorf("%s %s output: %w", b.Name, m.Name, err)
	}
	out := MethodOutput{
		Name: m.Name, Verb: m.Verb, Doc: m.Doc, Long: m.Long,
		InputType: behavior.TypeName(m.In), OutputType: behavior.TypeName(m.Out),
		Fields: Fields(input), Results: Results(output), Examples: []ExampleRow{},
		Input: input, Output: output, Calls: m.Examples, Refusals: m.Refusals,
	}
	if b.API {
		out.HTTPMethod, out.Route, out.Permission = m.HTTPMethod(), b.Path(
			m,
		), string(
			b.Permission(m.PermissionVerb()),
		)
	}
	if b.CLI {
		out.Command = b.Command(m)
	}
	for _, ex := range m.Examples {
		out.Examples = append(out.Examples, ExampleRow{
			Example: ex.Doc, Input: pairs(ex.Input), Output: pairs(ex.Output),
		})
	}

	return out, nil
}

// Create writes a new behavior's layout in the checkout, and is its files
// and the steps after it.
func Create(
	_ context.Context,
	_ Deps,
	in CreateInput,
) (CreateOutput, error) {
	root := in.Root
	if root == "" {
		root = "."
	}
	files, err := behavior.Scaffold(
		root, in.Group, in.Name, in.Package, in.Description, in.Requires)
	if err != nil {
		return CreateOutput{}, err
	}
	dir := in.Package
	if dir == "" {
		dir = in.Name
	}
	path := dir
	if in.Group != "" {
		path = in.Group + "/" + dir
	}
	out := CreateOutput{Name: in.Name}
	for _, f := range files {
		out.Files = append(out.Files, File{Path: f})
	}
	steps := []string{
		"go generate ./internal/behaviors/" + path + "/mocks ./internal/behaviors/all",
		"change the types, the provider and the examples in internal/behaviors/" + path,
	}
	if len(in.Requires) > 0 {
		// The scaffold wires each required behavior; the calls are the
		// author's, and the tests fail until an example makes one.
		steps = append(steps, "call "+strings.Join(in.Requires, " and ")+
			" with behavior.Call in internal/behaviors/"+path+"/debian.go, from an example:"+
			" until then, go test fails the rule used")
	}
	steps = append(steps,
		"go generate ./internal/behaviors/all",
		"osapi behavior validate",
		"just test",
	)
	for _, step := range steps {
		out.Next = append(out.Next, Step{Step: step})
	}

	return out, nil
}

// Spec is each part of the behavior DSL, by kind.
func Spec(
	_ context.Context,
	d Deps,
	_ SpecInput,
) (SpecOutput, error) {
	return SpecOutput{Kinds: d.Spec().Sections()}, nil
}

// Validate checks each behavior. When a rule fails, it is the report with
// command.Fail: the report prints, then the command exits 1.
func Validate(
	_ context.Context,
	d Deps,
	_ ValidateInput,
) (ValidateOutput, error) {
	out := ValidateOutput{Behaviors: len(d.Behaviors()), Problems: d.Validate()}
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

// Fields are the fields of an input schema in declared order, with their
// constraints. The method get of the endpoint and the command DSLs uses it
// too, so the three show a method the same way.
func Fields(
	s *behavior.Schema,
) []Field {
	out := make([]Field, 0, len(s.Order))
	for _, name := range s.Order {
		f := s.Properties[name]
		out = append(out, Field{
			Field: name, Type: f.Type, Required: slices.Contains(s.Required, name),
			Constraints: constraints(f), Description: f.Description,
		})
	}

	return out
}

// Results are the fields of an output schema in declared order. A list is
// described by its element, since each element is one result row. The method
// get of the endpoint and the command DSLs uses it too.
func Results(
	s *behavior.Schema,
) []Result {
	if s.Items != nil {
		s = s.Items
	}
	out := make([]Result, 0, len(s.Order))
	for _, name := range s.Order {
		f := s.Properties[name]
		out = append(out, Result{Field: name, Type: f.Type, Description: f.Description})
	}

	return out
}

// constraints is a field's validation in one cell.
func constraints(
	f *behavior.Schema,
) string {
	var parts []string
	if f.Pattern != "" {
		parts = append(parts, "pattern "+f.Pattern)
	}
	if f.Format != "" {
		parts = append(parts, "format "+f.Format)
	}
	if len(f.Enum) > 0 {
		parts = append(parts, "one of "+strings.Join(f.Enum, " "))
	}
	if f.MinLength != nil {
		parts = append(parts, fmt.Sprintf("min length %d", *f.MinLength))
	}
	if f.MaxLength != nil {
		parts = append(parts, fmt.Sprintf("max length %d", *f.MaxLength))
	}
	if f.Minimum != nil {
		parts = append(parts, fmt.Sprintf("min %g", *f.Minimum))
	}
	if f.Maximum != nil {
		parts = append(parts, fmt.Sprintf("max %g", *f.Maximum))
	}

	return strings.Join(parts, ", ")
}

// pairs is a value as its fields, key=value, in the order JSON gives them:
// key=vm.swappiness, value=10. A list is each element in brackets.
func pairs(
	v any,
) string {
	raw, _ := json.Marshal(v)
	var list []json.RawMessage
	if json.Unmarshal(raw, &list) == nil {
		parts := make([]string, 0, len(list))
		for _, item := range list {
			var x any
			_ = json.Unmarshal(item, &x)
			parts = append(parts, "["+pairs(x)+"]")
		}

		return strings.Join(parts, " ")
	}
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	if tok, err := dec.Token(); err != nil || tok != json.Delim('{') {
		return string(raw)
	}
	var parts []string
	for dec.More() {
		key, _ := dec.Token()
		var val any
		_ = dec.Decode(&val)
		parts = append(parts, fmt.Sprintf("%v=%v", key, val))
	}
	if len(parts) == 0 {
		return "(none)"
	}

	return strings.Join(parts, ", ")
}

// recordedInputs is each input field that the methods record on their
// spans, and the methods that record it: key (get, create, update, delete).
func recordedInputs(
	inputs map[string][]string,
	methods []*behavior.Method,
) string {
	byField := map[string][]string{}
	var order []string
	for _, m := range methods {
		for _, f := range inputs[m.Name] {
			if _, ok := byField[f]; !ok {
				order = append(order, f)
			}
			byField[f] = append(byField[f], m.Name)
		}
	}
	parts := make([]string, 0, len(order))
	for _, f := range order {
		parts = append(parts, f+" ("+strings.Join(byField[f], ", ")+")")
	}

	return strings.Join(parts, ", ")
}

// platformNames are the platforms b has a provider for, in declared order.
func platformNames(
	b *behavior.Behavior,
) []string {
	out := make([]string, 0, len(b.Platforms))
	for _, p := range b.Platforms {
		out = append(out, string(p.Name))
	}

	return out
}

// names are the behaviors' names, in order, never nil.
func names(
	bs []*behavior.Behavior,
) []string {
	out := make([]string, 0, len(bs))
	for _, b := range bs {
		out = append(out, b.Name)
	}

	return out
}

// OpenAPIGet is the served behavior's part of the OpenAPI spec, as YAML: the
// part that go generate writes to its gen/api.yaml and puts in the full spec
// of the API.
func OpenAPIGet(
	_ context.Context,
	d Deps,
	in GetInput,
) (OpenAPIOutput, error) {
	b, err := find(d, in.Name)
	if err != nil {
		return OpenAPIOutput{}, err
	}
	if !b.API {
		return OpenAPIOutput{}, fmt.Errorf(
			"behavior %s serves no route: it has no behavior.API, so it has no part of the spec",
			b.Name,
		)
	}
	doc, err := d.OpenAPI(b)
	if err != nil {
		return OpenAPIOutput{}, fmt.Errorf("the spec of %s does not build: %w", b.Name, err)
	}
	// The spec goes through JSON, so its keys are its json names.
	raw, _ := json.Marshal(doc)
	var tree any
	_ = json.Unmarshal(raw, &tree)
	var out strings.Builder
	enc := yaml.NewEncoder(&out)
	enc.SetIndent(2)
	_ = enc.Encode(tree)

	return OpenAPIOutput{Name: b.Name, YAML: out.String()}, nil
}

// sudoOf is each program b runs with sudo on every platform, then each that
// one platform's provider runs with it, once each: the providers of several
// platforms can run the same program, such as update-ca-certificates on
// debian and suse.
func sudoOf(
	b *behavior.Behavior,
) []string {
	out := append([]string{}, b.Sudo...)
	for _, ps := range b.ProviderSudo {
		for _, p := range ps.Programs {
			if !slices.Contains(out, p) {
				out = append(out, p)
			}
		}
	}

	return out
}

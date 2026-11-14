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

package cmd

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/config"
)

// methodDescription is one method in `behavior get --json`. Route and
// command are where the method is served and run today; command is empty
// when the behavior has no CLI, and route when it has no API.
type methodDescription struct {
	Name       string        `json:"name"`
	Doc        string        `json:"doc"`
	Long       string        `json:"long,omitempty"`
	Verb       behavior.Verb `json:"verb"`
	HTTPMethod string        `json:"http_method,omitempty"`
	Route      string        `json:"route,omitempty"`
	// Permission is what a token needs to call the method; empty for a
	// behavior that is not served, which no token can call.
	Permission string             `json:"permission,omitempty"`
	Command    string             `json:"command,omitempty"`
	Input      *behavior.Schema   `json:"input"`
	Output     *behavior.Schema   `json:"output"`
	InputType  string             `json:"input_type"`
	OutputType string             `json:"output_type"`
	Examples   []behavior.Example `json:"examples,omitempty"`
	// Refusals are calls the method fails, with text its error contains.
	Refusals []behavior.Refusal `json:"refusals,omitempty"`
}

// telemetryDescription is what a behavior's calls record: the span and
// metrics every method gets, and the input fields this behavior tags onto
// the span, by method.
type telemetryDescription struct {
	Span       string              `json:"span"`
	Metrics    []string            `json:"metrics"`
	Attributes []string            `json:"attributes"`
	Inputs     map[string][]string `json:"inputs"`
}

// behaviorDescription is what `behavior get --json` prints.
type behaviorDescription struct {
	Name      string             `json:"name"`
	Doc       string             `json:"doc"`
	Placement behavior.Placement `json:"placement"`
	Group     string             `json:"group,omitempty"`
	// Internal is set for a behavior only other behaviors use: it has no API
	// or CLI of its own.
	Internal bool `json:"internal,omitempty"`
	API      bool `json:"api"`
	CLI      bool `json:"cli"`
	// Uses lists the behaviors this one is built on, and UsedBy those built
	// on it. Both are empty until a behavior can require another.
	Uses      []string             `json:"uses"`
	UsedBy    []string             `json:"used_by"`
	Telemetry telemetryDescription `json:"telemetry"`
	// Platforms are the platforms the behavior has a provider for; on any
	// other its methods are skipped. Sudo lists the programs it runs with
	// sudo, and Runs those it runs without.
	Platforms []string `json:"platforms"`
	Sudo      []string `json:"sudo"`
	Runs      []string `json:"runs"`
	// Settings are the behavior's own config keys, under behaviors.<name>.
	Settings []config.Key        `json:"settings"`
	Methods  []methodDescription `json:"methods"`
}

var behaviorGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Show a behavior: its surfaces, chain, methods, types and telemetry",
	Long: `Show a behavior: where it is served, its platforms, the programs it runs
with and without sudo, its methods, types, settings and telemetry.

Uses lists the behaviors it is built on: their providers are built first on
the agent and its methods call theirs on the same host, in the same job. If a
behavior it uses cannot run on a platform, this one reports skipped there too.
Used by lists the behaviors built on this one. Internal: yes means it has no
route, command or permission, and only the behaviors in Used by call it.`,
	Example: `  osapi behavior get --name sysctl
  osapi behavior get --name sysctl --method create
  osapi behavior get --name sysctl --method create --json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")
		method, _ := cmd.Flags().GetString("method")

		b := behavior.Lookup(name)
		if b == nil {
			return fmt.Errorf("no behavior named %q; run 'osapi behavior list'", name)
		}

		methods := b.Methods
		if method != "" {
			m := b.Method(method)
			if m == nil {
				return fmt.Errorf("behavior %s has no method %q", b.Name, method)
			}
			methods = []*behavior.Method{m}
		}

		desc, err := describeBehavior(b, methods)
		if err != nil {
			return err
		}

		if jsonOutput {
			out, _ := json.Marshal(desc)
			cli.PrintJSON(out)

			return nil
		}

		if method != "" {
			printMethod(desc.Methods[0])

			return nil
		}
		printBehavior(desc)

		return nil
	},
}

func describeBehavior(
	b *behavior.Behavior,
	methods []*behavior.Method,
) (behaviorDescription, error) {
	desc := behaviorDescription{
		Name: b.Name, Doc: b.Doc, Placement: b.Placement, Group: b.Group,
		API: b.API, CLI: b.CLI, Internal: b.Internal,
		Uses: names(b.Requires), UsedBy: names(behavior.RequiredBy(b)),
		Settings:  behaviorSettings(b.Name),
		Platforms: platformNames(b),
		Sudo:      append([]string{}, b.Sudo...),
		Runs:      append([]string{}, b.Runs...),
		Telemetry: telemetryDescription{
			Span: "behavior." + b.Name + ".<method>",
			Metrics: []string{
				behavior.MetricCalls,
				behavior.MetricDuration,
				behavior.MetricChanges,
			},
			Attributes: []string{
				string(behavior.AttrBehavior), string(behavior.AttrMethod),
				string(
					behavior.AttrVerb,
				), string(behavior.AttrOutcome), string(behavior.AttrChanged),
			},
			Inputs: map[string][]string{},
		},
	}

	for _, m := range methods {
		in, err := behavior.SchemaFor(m.In)
		if err != nil {
			return desc, fmt.Errorf("%s %s input: %w", b.Name, m.Name, err)
		}
		out, err := behavior.SchemaFor(m.Out)
		if err != nil {
			return desc, fmt.Errorf("%s %s output: %w", b.Name, m.Name, err)
		}

		if tagged := behavior.TaggedInputs(m.In); len(tagged) > 0 {
			desc.Telemetry.Inputs[m.Name] = tagged
		}

		md := methodDescription{
			Name:       m.Name,
			Doc:        m.Doc,
			Long:       m.Long,
			Verb:       m.Verb,
			Input:      in,
			Output:     out,
			InputType:  typeName(m.In),
			OutputType: typeName(m.Out),
			Examples:   m.Examples,
			Refusals:   m.Refusals,
		}

		if b.API {
			md.Permission = b.Permission(m.Verb)
			md.HTTPMethod = m.HTTPMethod()
			md.Route = b.Route(m)
		}
		if b.CLI {
			md.Command = b.CLICommand(m)
		}
		desc.Methods = append(desc.Methods, md)
	}

	return desc, nil
}

// typeName is a type's name, with [] for a list: []Entry, MutationResult.
func typeName(
	t reflect.Type,
) string {
	if t.Kind() == reflect.Slice {
		return "[]" + t.Elem().Name()
	}

	return t.Name()
}

func printBehavior(
	desc behaviorDescription,
) {
	methods := make([][]string, 0, len(desc.Methods))
	types := make([][]string, 0, len(desc.Methods))
	for _, m := range desc.Methods {
		route := "-"
		if m.Route != "" {
			route = m.HTTPMethod + " " + m.Route
		}
		permission := m.Permission
		if permission == "" {
			permission = "-"
		}
		methods = append(methods, []string{m.Name, string(m.Verb), route, permission, m.Doc})
		types = append(types, []string{
			m.Name,
			shape(m.InputType, m.Input, true),
			shape(m.OutputType, m.Output, false),
		})
	}

	cli.PrintHeading("Behavior")
	cli.PrintKV(
		"Name", desc.Name,
		"Description", desc.Doc,
		"Placement", string(desc.Placement),
		"Internal", yesNo(desc.Internal),
		"API", surface(desc.API, behavior.RoutePrefix),
		"CLI", surface(desc.CLI, behavior.CommandPrefix),
		"Platforms", listOrNone(desc.Platforms),
		"Sudo", listOrNone(desc.Sudo),
		"Runs", listOrNone(desc.Runs),
	)

	cli.PrintCompactTable([]cli.Section{
		chain("Uses", desc.Uses),
		chain("Used by", desc.UsedBy),
		{
			Title:     fmt.Sprintf("Methods (%d)", len(desc.Methods)),
			Headers:   []string{"NAME", "VERB", "ROUTE", "PERMISSION", "DESCRIPTION"},
			Rows:      methods,
			Reference: true,
		},
		{
			Title:     fmt.Sprintf("Types (%d)", len(desc.Methods)),
			Headers:   []string{"NAME", "INPUT", "OUTPUT"},
			Rows:      types,
			Reference: true,
		},
		settingsSection(desc.Settings),
	})

	cli.PrintHeading("Telemetry")
	cli.PrintKV(
		"Span", desc.Telemetry.Span,
		"Metrics", strings.Join(desc.Telemetry.Metrics, ", "),
		"Attributes", strings.Join(desc.Telemetry.Attributes, ", "),
		"Recorded inputs", recordedInputs(desc.Telemetry.Inputs, desc.Methods),
	)
}

// behaviorSettings are the config keys a behavior declared with
// behavior.Config, empty when it declared none.
func behaviorSettings(
	name string,
) []config.Key {
	out := []config.Key{}
	for _, k := range config.Keys() {
		if k.Behavior == name {
			out = append(out, k)
		}
	}

	return out
}

// settingsSection lists a behavior's settings: the key, the agent start flag
// that sets it, its type and default.
func settingsSection(
	keys []config.Key,
) cli.Section {
	rows := make([][]string, 0, len(keys))
	for _, k := range keys {
		def := ""
		if k.Default != nil {
			def = fmt.Sprint(k.Default)
		}
		rows = append(rows, []string{k.Name, "--" + k.Flag, k.Type, def})
	}

	return cli.Section{
		Title:     fmt.Sprintf("Settings (%d)", len(keys)),
		Headers:   []string{"KEY", "AGENT FLAG", "TYPE", "DEFAULT"},
		Rows:      rows,
		Reference: true,
		Empty:     "Declare them with behavior.Config.",
	}
}

// chain is a Uses or Used by section: each behavior in the chain with its
// description, or None.
func chain(
	title string,
	names []string,
) cli.Section {
	rows := make([][]string, 0, len(names))
	for _, n := range names {
		desc := ""
		if b := behavior.Lookup(n); b != nil {
			desc = b.Doc
		}
		rows = append(rows, []string{n, desc})
	}

	return cli.Section{
		Title:     title,
		Headers:   []string{"NAME", "DESCRIPTION"},
		Rows:      rows,
		Reference: true,
	}
}

// surface says where a behavior's surface is served, or none.
func surface(
	on bool,
	where string,
) string {
	if !on {
		return "none"
	}

	return where
}

// shape renders a type as its name and fields: CreateInput { key*, value* },
// []Entry { key, value }. A required input field is marked *.
func shape(
	name string,
	s *behavior.Schema,
	input bool,
) string {
	fields := fieldList(s, input)
	if fields == "(none)" {
		return name
	}

	return name + " { " + fields + " }"
}

// fieldList is a schema's fields in one line, required inputs marked with *:
// key*, value*. A list output is described by its element.
func fieldList(
	s *behavior.Schema,
	input bool,
) string {
	if s.Items != nil {
		s = s.Items
	}
	if len(s.Order) == 0 {
		return "(none)"
	}

	parts := make([]string, 0, len(s.Order))
	for _, name := range s.Order {
		if input && slices.Contains(s.Required, name) {
			name += "*"
		}
		parts = append(parts, name)
	}

	return strings.Join(parts, ", ")
}

func printMethod(
	m methodDescription,
) {
	cli.PrintHeading("Method")
	cli.PrintKV("Name", m.Name, "Verb", string(m.Verb))
	cli.PrintKV("Description", m.Doc)
	if m.Long != "" {
		cli.PrintKV("Detail", m.Long)
	}
	if m.Route != "" {
		cli.PrintKV("Route", m.HTTPMethod+" "+m.Route, "Permission", m.Permission)
	} else {
		cli.PrintKV("Called by", "other behaviors only; it is not served")
	}
	if m.Command != "" {
		cli.PrintKV("Command", m.Command)
	}

	examples := make([][]string, 0, len(m.Examples))
	for _, ex := range m.Examples {
		examples = append(examples, []string{ex.Doc, pairs(ex.Input), pairs(ex.Output)})
	}

	cli.PrintCompactTable([]cli.Section{
		{
			Title:     "Input (" + m.InputType + ")",
			Headers:   []string{"FIELD", "TYPE", "REQUIRED", "CONSTRAINTS", "DESCRIPTION"},
			Rows:      fieldRows(m.Input, true),
			Reference: true,
		},
		{
			Title:     "Output (" + m.OutputType + ")",
			Headers:   []string{"FIELD", "TYPE", "DESCRIPTION"},
			Rows:      fieldRows(m.Output, false),
			Reference: true,
		},
		{
			Title:     fmt.Sprintf("Examples (%d)", len(m.Examples)),
			Headers:   []string{"EXAMPLE", "INPUT", "OUTPUT"},
			Rows:      examples,
			Reference: true,
		},
	})
}

// fieldRows lists a schema's fields in declaration order. A slice output is
// described by its element, since each element is one result row.
func fieldRows(
	s *behavior.Schema,
	withConstraints bool,
) [][]string {
	if s.Type == "array" && s.Items != nil {
		s = s.Items
	}

	rows := make([][]string, 0, len(s.Order))
	for _, name := range s.Order {
		f := s.Properties[name]
		if withConstraints {
			rows = append(rows, []string{
				name, f.Type, strconv.FormatBool(slices.Contains(s.Required, name)),
				constraints(f), f.Description,
			})

			continue
		}
		rows = append(rows, []string{name, f.Type, f.Description})
	}

	return rows
}

// constraints summarizes a field's validation in one cell.
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

func init() {
	behaviorCmd.AddCommand(behaviorGetCmd)

	behaviorGetCmd.Flags().
		String("name", "", "The behavior, from 'osapi behavior list' (required)")
	behaviorGetCmd.Flags().
		String("method", "", "One method to show in full: its fields, constraints and examples")
	_ = behaviorGetCmd.MarkFlagRequired("name")
}

// pairs renders a value as its fields, key=value, in the order JSON gives
// them: key=vm.swappiness, value=10. A list is each element in brackets. No
// raw JSON reaches a table cell.
func pairs(
	v any,
) string {
	raw, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprint(v)
	}

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

// recordedInputs lists the input fields a behavior tags onto its spans, and
// the methods each is recorded by: key (get, create, update, delete).
func recordedInputs(
	inputs map[string][]string,
	methods []methodDescription,
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
	if len(order) == 0 {
		return "none"
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
	names := make([]string, 0, len(b.Platforms))
	for _, pp := range b.Platforms {
		names = append(names, pp.Name)
	}

	return names
}

// listOrNone joins a list for a key-value line, or says none.
func listOrNone(
	items []string,
) string {
	if len(items) == 0 {
		return "none"
	}

	return strings.Join(items, ", ")
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

// yesNo says a flag for a key-value line.
func yesNo(
	on bool,
) string {
	if on {
		return "yes"
	}

	return "no"
}

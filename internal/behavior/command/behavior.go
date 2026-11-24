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

// Package behavior holds what the generated behavior commands share:
// flags derived from an input type's tags, and the one table layout every
// behavior's output renders in, driven by its output type's tags. The
// commands themselves are generated into cmd/next_<behavior>.go.
package command

import (
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/cli"
	clicommand "github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/internal/secret"
)

// Flag is one CLI flag derived from an input field.
type Flag struct {
	// Name is the flag name: the field's JSON name with _ written as -.
	Name string
	// JSON is the field's JSON name, the key the input is sent under.
	JSON string
	// Kind is the field's kind: string, bool, int, float, or map for a
	// map[string]string, which takes key=value.
	Kind reflect.Kind
	// Usage is the field's doc tag.
	Usage string
	// Required is set when the field's validate tag has required.
	Required bool
	// Index is the field's index in the input type.
	Index int
}

// Flags derives one flag per exported field of the input type, in order.
// A field without a JSON name, or of a kind a flag cannot hold, is skipped.
func Flags(
	in reflect.Type,
) []Flag {
	if in.Kind() != reflect.Struct {
		return nil
	}

	flags := make([]Flag, 0, in.NumField())
	for i := range in.NumField() {
		f := in.Field(i)
		name := jsonName(f)
		// A field that the controller step sets is not the caller's to send.
		if !f.IsExported() || name == "" || behavior.SetByBefore(f) {
			continue
		}

		kind := flagKind(f.Type)
		if kind == reflect.Invalid {
			continue
		}

		usage := f.Tag.Get("doc")
		switch f.Type {
		case reflect.TypeFor[secret.Value]():
			usage += clicommand.SecretHelp
		case reflect.TypeFor[map[string]secret.Value]():
			usage += clicommand.SecretsHelp
		}
		// A required flag says so through MarkFlagRequired, which --help
		// shows, as each command's flag does.
		required := hasRule(f.Tag.Get("validate"), "required")

		flags = append(flags, Flag{
			Name:     strings.ReplaceAll(name, "_", "-"),
			JSON:     name,
			Kind:     kind,
			Usage:    usage,
			Required: required,
			Index:    i,
		})
	}

	return flags
}

// Args are the flags that set input's non-zero fields, as a command line
// takes them. A secret is its reference, or - for a literal value, which
// the command reads from standard input: Stdin is what it reads.
func Args(
	input any,
) []string {
	if input == nil {
		return nil
	}

	var args []string
	v := reflect.ValueOf(input)
	for _, f := range Flags(v.Type()) {
		fv := v.Field(f.Index)
		if fv.IsZero() {
			continue
		}
		if f.Kind == reflect.Map {
			for _, k := range slices.Sorted(maps.Keys(pairs(fv))) {
				args = append(args, "--"+f.Name, k+"="+pairs(fv)[k])
			}

			continue
		}
		if s, ok := fv.Interface().(secret.Value); ok {
			args = append(args, "--"+f.Name, flagOf(s))

			continue
		}
		// A bool flag is set by its name: cobra reads a word after it as
		// an argument, so --force true is a command with an extra word.
		if f.Kind == reflect.Bool {
			args = append(args, "--"+f.Name)

			continue
		}
		// A list is the flag given once for each value.
		if f.Kind == reflect.Slice {
			values, _ := fv.Interface().([]string)
			for _, value := range values {
				args = append(args, "--"+f.Name, value)
			}

			continue
		}
		args = append(args, "--"+f.Name, fmt.Sprint(fv.Interface()))
	}

	return args
}

// Stdin is what the command line of Args reads from standard input: the
// literal value of input's secret field, or "" when it has none. A command
// reads one value from standard input, so a second literal is not read.
func Stdin(
	input any,
) string {
	if input == nil {
		return ""
	}
	v := reflect.ValueOf(input)
	for _, f := range Flags(v.Type()) {
		if s, ok := v.Field(f.Index).Interface().(secret.Value); ok && s != "" && !s.IsReference() {
			return s.Reveal()
		}
	}

	return ""
}

// flagOf is a secret as its flag takes it: its reference, or - for a
// literal, which the command reads from standard input.
func flagOf(
	s secret.Value,
) string {
	if s.IsReference() {
		return s.Reveal()
	}

	return clicommand.Stdin
}

// pairs is a map flag's value as key=value pairs: a map of secrets gives
// each reference, or - for a literal, which the command refuses.
func pairs(
	fv reflect.Value,
) map[string]string {
	if m, ok := fv.Interface().(map[string]secret.Value); ok {
		out := make(map[string]string, len(m))
		for k, s := range m {
			out[k] = flagOf(s)
		}

		return out
	}
	m, _ := fv.Interface().(map[string]string)

	return m
}

// Columns are the output fields that a behavior's result table shows, in
// field order: the command layer's columns of the output type, except
// changed, which the table renders as the row's status.
func Columns(
	out reflect.Type,
) []clicommand.Column {
	var cols []clicommand.Column
	for _, c := range clicommand.Columns(out) {
		if c.JSON != "changed" {
			cols = append(cols, c)
		}
	}

	return cols
}

// Table renders a result the one way every behavior command prints: one row
// per result, the output fields as columns, and changed folded into the
// row's status for a write. A host that failed or was skipped is listed under
// the table with its reason. Each value prints as the command layer prints
// it, by its field's tags.
func Table(
	m *behavior.Method,
	r Result,
) cli.Section {
	cols := Columns(m.Out)
	headers := make([]string, 0, len(cols))
	for _, c := range cols {
		headers = append(headers, c.Header)
	}

	rows := make([]cli.ResultRow, 0, len(r.Results))
	for _, res := range r.Results {
		row := cli.ResultRow{
			Hostname: fmt.Sprint(res["hostname"]),
			Status:   types.RowStatus(fmt.Sprint(res["status"])),
		}
		if e, ok := res["error"].(string); ok && e != "" {
			row.Error = &e
		}
		if m.Verb == types.VerbWrite {
			if c, ok := res["changed"].(bool); ok {
				row.Changed = &c
			}
		}
		for _, c := range cols {
			row.Fields = append(row.Fields, c.Format(cli.Printer{}, res[c.JSON]))
		}
		rows = append(rows, row)
	}

	tr := cli.BuildBroadcastTable(rows, headers)
	if m.Verb == types.VerbWrite {
		tr = cli.BuildMutationTable(rows, headers)
	}

	return cli.Section{Headers: tr.Headers, Rows: tr.Rows, Errors: tr.Errors}
}

func jsonName(
	f reflect.StructField,
) string {
	name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
	if name == "-" {
		return ""
	}

	return name
}

func flagKind(
	t reflect.Type,
) reflect.Kind {
	// The CLI owns which types a flag holds: a field it does not flag is
	// not a flag here either, so the docs, the examples and the tests
	// show the command that the generator builds.
	if !clicommand.Flaggable(t) || t == reflect.TypeFor[[]byte]() {
		return reflect.Invalid
	}
	switch t.Kind() {
	case reflect.Map:
		return reflect.Map
	case reflect.Slice:
		return reflect.Slice
	default:
		return t.Kind()
	}
}

func hasRule(
	tag string,
	rule string,
) bool {
	for _, r := range strings.Split(tag, ",") {
		if r == rule {
			return true
		}
	}

	return false
}

func init() {
	clicommand.PrintsGroups(printResults)
}

// printResults prints the answer of a behavior's method as its result
// table, and is false for a group that is not a behavior.
func printResults(
	g clicommand.Group,
	method string,
	answer []byte,
) bool {
	b := behavior.Lookup(g.GroupName())
	if b == nil || b.Method(method) == nil {
		return false
	}
	// A group with the behavior's name that is not its group, such as the
	// controller's file endpoints beside the file behavior, is not it.
	if eg, ok := g.(*endpoint.Group); ok && eg != behavior.GroupOf(b) {
		return false
	}
	Print(b, method, json.RawMessage(answer), clicommand.TraceID())

	return true
}

// Print writes the answer of method of b, out, the method's response type:
// the job ID, the trace ID of the call, and the table every behavior
// renders. The command runtime calls it through printResults.
func Print(
	b *behavior.Behavior,
	method string,
	out any,
	traceID string,
) {
	var r Result
	data, _ := json.Marshal(out)
	_ = json.Unmarshal(data, &r)

	cli.PrintJobID(r.JobID)
	cli.PrintTraceID(traceID)
	cli.PrintCompactTable([]cli.Section{Table(b.Method(method), r)})
}

// Result is what every behavior method answers: the job and one row per
// host, or per element when the method returns a list.
type Result struct {
	JobID   string           `json:"job_id"`
	Results []map[string]any `json:"results"`
}

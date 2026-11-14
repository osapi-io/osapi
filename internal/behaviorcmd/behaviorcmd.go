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

// Package behaviorcmd holds what the generated behavior commands share:
// flags derived from an input type's tags, and the one table layout every
// behavior's output renders in, driven by its output type's tags. The
// commands themselves are generated into cmd/next_<behavior>.go.
package behaviorcmd

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/cli"
)

// Flag is one CLI flag derived from an input field.
type Flag struct {
	// Name is the flag name: the field's JSON name with _ written as -.
	Name string
	// JSON is the field's JSON name, the key the input is sent under.
	JSON string
	// Kind is the field's kind: string, bool, int or float.
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
		if !f.IsExported() || name == "" {
			continue
		}

		kind := flagKind(f.Type.Kind())
		if kind == reflect.Invalid {
			continue
		}

		usage := f.Tag.Get("doc")
		if behavior.Secret(f) {
			usage += " (secret: kept out of the audit log and traces)"
		}
		required := hasRule(f.Tag.Get("validate"), "required")
		if required {
			usage += " (required)"
		}

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
// takes them.
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
		args = append(args, "--"+f.Name, fmt.Sprint(fv.Interface()))
	}

	return args
}

// Render values for the render tag on an output field.
const (
	// RenderHidden leaves the field out of the table. It is still in --json.
	RenderHidden = "-"
	// RenderBytes prints a byte count as 4.0 KB.
	RenderBytes = "bytes"
	// RenderAge prints a number of seconds as 3d 4h.
	RenderAge = "age"
	// RenderStatus colors the value as a status, like the STATUS column.
	RenderStatus = "status"
)

// Column is one table column, from one field of the output type.
type Column struct {
	// JSON is the field's JSON name, the key its value has in a result row.
	JSON string
	// Header is the column header: the JSON name in capitals.
	Header string
	// Render is the field's render tag, empty for a plain value.
	Render string
}

// Columns returns the output fields shown as table columns, in field order:
// every field of the output type, or of its element type for a list, except
// changed, which the table renders as the row's status, and any field tagged
// render:"-".
func Columns(
	out reflect.Type,
) []Column {
	for out.Kind() == reflect.Slice || out.Kind() == reflect.Pointer {
		out = out.Elem()
	}
	if out.Kind() != reflect.Struct {
		return nil
	}

	cols := make([]Column, 0, out.NumField())
	for i := range out.NumField() {
		f := out.Field(i)
		name := jsonName(f)
		render := f.Tag.Get("render")
		if !f.IsExported() || name == "" || name == "changed" || render == RenderHidden {
			continue
		}
		cols = append(cols, Column{
			JSON:   name,
			Header: strings.ToUpper(strings.ReplaceAll(name, "_", " ")),
			Render: render,
		})
	}

	return cols
}

// Table renders a result the one way every behavior command prints: one row
// per result, the output fields as columns, and changed folded into the
// row's status for a write. A host that failed or was skipped is listed under
// the table with its reason.
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
			Status:   fmt.Sprint(res["status"]),
		}
		if e, ok := res["error"].(string); ok && e != "" {
			row.Error = &e
		}
		if m.Verb == behavior.VerbWrite {
			if c, ok := res["changed"].(bool); ok {
				row.Changed = &c
			}
		}
		for _, c := range cols {
			row.Fields = append(row.Fields, format(c, res[c.JSON]))
		}
		rows = append(rows, row)
	}

	tr := cli.BuildBroadcastTable(rows, headers)
	if m.Verb == behavior.VerbWrite {
		tr = cli.BuildMutationTable(rows, headers)
	}

	return cli.Section{Headers: tr.Headers, Rows: tr.Rows, Errors: tr.Errors}
}

// format renders one value by its column's render tag.
func format(
	c Column,
	v any,
) string {
	if v == nil {
		return ""
	}

	n, isNumber := v.(float64)
	switch {
	case c.Render == RenderBytes && isNumber:
		return cli.FormatBytes(int(n))
	case c.Render == RenderAge && isNumber:
		return cli.FormatAge(time.Duration(n) * time.Second)
	case isNumber && n == float64(int64(n)):
		return strconv.FormatInt(int64(n), 10)
	default:
		return fmt.Sprint(v)
	}
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
	k reflect.Kind,
) reflect.Kind {
	switch k {
	case reflect.String:
		return reflect.String
	case reflect.Bool:
		return reflect.Bool
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return reflect.Int
	case reflect.Float32, reflect.Float64:
		return reflect.Float64
	default:
		return reflect.Invalid
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

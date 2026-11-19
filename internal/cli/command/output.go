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

package command

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Printer is how the output of a method prints: the headings, blocks and
// tables of the CLI, and the formats of its values. The CLI gives the
// helpers of internal/cli, which has the one theme, through Deps.
type Printer interface {
	// Heading prints the heading of a block.
	Heading(title string)
	// KV adds labelled values to the key-value block under the heading.
	KV(pairs ...string)
	// Table prints a table under its heading.
	Table(t Table)
	// Verdict prints one line at the left margin, such as m-1 admitted.
	Verdict(text string)
	// Status is a status word in its color.
	Status(word string) string
	// Bytes is a size, such as 3.4 GB.
	Bytes(n int64) string
	// Age is the time since something, such as 3s.
	Age(d time.Duration) string
	// Time is a time in the one time format of the CLI.
	Time(t time.Time) string
	// Duration is a duration rounded to a useful precision.
	Duration(d time.Duration) string
}

// Table is a table of the output: its heading with the count, the column
// headers, the rows, and the hint that an empty list prints after None.
type Table struct {
	Title   string
	Headers []string
	Rows    [][]string
	Empty   string
}

// The values of the render tag of an output field.
const (
	// RenderHidden leaves the field out of the output. --json has it.
	RenderHidden = "-"
	// RenderBytes prints a number of bytes as a size, such as 3.4 GB.
	RenderBytes = "bytes"
	// RenderAge prints the time since a time, or a number of seconds, such
	// as 3s.
	RenderAge = "age"
	// RenderTime prints a time in the one time format of the CLI.
	RenderTime = "time"
	// RenderDuration prints a duration: a Go duration, as a string or as a
	// number of nanoseconds.
	RenderDuration = "duration"
	// RenderDurationMs prints an integer of milliseconds as a duration.
	RenderDurationMs = "duration=ms"
	// RenderStatus prints a status word in its color.
	RenderStatus = "status"
	// RenderShort is the prefix of short=N, which prints the first N
	// characters, with ... after them.
	RenderShort = "short="
)

// Column is one field of an output type: a column of a table, or a pair of
// a key-value block.
type Column struct {
	// JSON is the field's json name, its key in the answer.
	JSON string
	// Label is the json name in sentence case, or the label tag.
	Label string
	// Header is the label in capitals.
	Header string
	// Render is the field's render tag, or "".
	Render string
	// None is the text for a field with no value, from the none tag.
	None string
	// Empty is the hint after None. for an empty list, from the empty tag.
	Empty string
	// Prefix is the text in front of each value, from the prefix tag.
	Prefix string
	// OmitEmpty is true for a field whose json tag has omitempty or
	// omitzero: with no value, it is not shown.
	OmitEmpty bool
	// NoTable is true for a field tagged table:"-": a table leaves it out,
	// and a key-value block shows it.
	NoTable bool
	// Type is the field's type.
	Type reflect.Type
}

// Columns are the fields of out that the output shows, in the order of the
// fields: each exported field with a json name, of out or of its element
// for a list, except a field tagged render:"-".
func Columns(
	out reflect.Type,
) []Column {
	out = element(out)
	if out.Kind() != reflect.Struct {
		return nil
	}

	cols := make([]Column, 0, out.NumField())
	for f := range out.Fields() {
		name, opts, _ := strings.Cut(f.Tag.Get("json"), ",")
		render := f.Tag.Get("render")
		if !f.IsExported() || name == "" || name == "-" || render == RenderHidden {
			continue
		}
		label := f.Tag.Get("label")
		if label == "" {
			label = sentence(name)
		}
		cols = append(cols, Column{
			JSON:      name,
			Label:     label,
			Header:    strings.ToUpper(label),
			Render:    render,
			None:      f.Tag.Get("none"),
			Empty:     f.Tag.Get("empty"),
			Prefix:    f.Tag.Get("prefix"),
			OmitEmpty: strings.Contains(opts, "omitempty") || strings.Contains(opts, "omitzero"),
			NoTable:   f.Tag.Get("table") == "-",
			Type:      f.Type,
		})
	}

	return cols
}

// element is t without its pointers and lists: the type of one item.
func element(
	t reflect.Type,
) reflect.Type {
	for t.Kind() == reflect.Slice || t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	return t
}

// initialisms are the words a label keeps in capitals.
var initialisms = map[string]string{
	"id": "ID", "ip": "IP", "ipv4": "IPv4", "ipv6": "IPv6", "os": "OS", "cpu": "CPU",
	"sha": "SHA", "sha256": "SHA256", "url": "URL", "fqdn": "FQDN", "mac": "MAC",
	"api": "API", "dlq": "DLQ", "kv": "KV", "nats": "NATS", "http": "HTTP", "uuid": "UUID",
}

// sentence is a json name in sentence case, the label of its field:
// last_seen is Last seen, and an initialism such as id stays in capitals.
func sentence(
	name string,
) string {
	words := strings.Split(name, "_")
	for i, w := range words {
		switch {
		case initialisms[w] != "":
			words[i] = initialisms[w]
		case i == 0 && w != "":
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}

	return strings.Join(words, " ")
}

// plural is the heading of a list: the name with s, or as it is when it
// ends in s.
func plural(
	name string,
) string {
	if strings.HasSuffix(name, "s") {
		return name
	}

	return name + "s"
}

// Format is the text of v, the value of the column's field in the answer as
// JSON decodes it, by its render tag and its type, for a table: short=N cuts
// it.
func (c Column) Format(
	p Printer,
	v any,
) string {
	if noValue(c, v) {
		return c.None
	}

	return c.Prefix + c.format(p, v)
}

// FormatKV is the text of v for a key-value block: as Format, except that
// short=N does not cut it, because a block has the room for all of it.
func (c Column) FormatKV(
	p Printer,
	v any,
) string {
	if strings.HasPrefix(c.Render, RenderShort) {
		c.Render = ""
	}

	return c.Format(p, v)
}

// noValue reports whether v is no value: absent, null, an empty string, or
// a zero time.
func noValue(
	c Column,
	v any,
) bool {
	switch t := v.(type) {
	case nil:
		return true
	case string:
		if t == "" {
			return true
		}
		if isTime(c.Type) {
			parsed, err := time.Parse(time.RFC3339Nano, t)

			return err == nil && parsed.IsZero()
		}
	}

	return false
}

func (c Column) format(
	p Printer,
	v any,
) string {
	n, isNumber := v.(float64)
	s, isString := v.(string)
	switch {
	case c.Render == RenderBytes && isNumber:
		return p.Bytes(int64(n))
	case c.Render == RenderAge && isNumber:
		return p.Age(time.Duration(n) * time.Second)
	case c.Render == RenderAge && isString:
		if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
			return p.Age(time.Since(t))
		}
	case (c.Render == RenderTime || isTime(c.Type)) && isString:
		if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
			return p.Time(t)
		}
	case c.Render == RenderDuration && isNumber:
		return p.Duration(time.Duration(n))
	case c.Render == RenderDuration && isString:
		if d, err := time.ParseDuration(s); err == nil {
			return p.Duration(d)
		}
	case c.Render == RenderDurationMs && isNumber:
		return p.Duration(time.Duration(n) * time.Millisecond)
	case c.Render == RenderStatus && isString:
		return p.Status(s)
	case strings.HasPrefix(c.Render, RenderShort) && isString:
		if n, err := strconv.Atoi(strings.TrimPrefix(c.Render, RenderShort)); err == nil &&
			len(s) > n {
			return s[:n] + "..."
		}
	}

	return scalar(v)
}

// scalar is the text of a value with no render tag: a bool is yes or no, a
// whole number has no decimals, a list of values is joined with commas, an
// object of values is its key:value pairs, and a list or object with more
// in it is a count, which --json prints all of.
func scalar(
	v any,
) string {
	switch t := v.(type) {
	case bool:
		if t {
			return "yes"
		}

		return "no"
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}

		return strconv.FormatFloat(t, 'f', -1, 64)
	case []any:
		parts := make([]string, 0, len(t))
		for _, item := range t {
			if !flat(item) {
				return count(len(t), "item")
			}
			parts = append(parts, scalar(item))
		}

		return strings.Join(parts, ", ")
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k, item := range t {
			if !flat(item) {
				return count(len(t), "field")
			}
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			parts = append(parts, k+":"+scalar(t[k]))
		}

		return strings.Join(parts, ", ")
	default:
		return fmt.Sprint(t)
	}
}

// flat reports whether v is one value, not a list or an object.
func flat(
	v any,
) bool {
	switch v.(type) {
	case []any, map[string]any:
		return false
	default:
		return true
	}
}

// count is n of noun, with the noun that agrees with it.
func count(
	n int,
	noun string,
) string {
	if n == 1 {
		return "1 " + noun
	}

	return strconv.Itoa(n) + " " + noun + "s"
}

var timeType = reflect.TypeFor[time.Time]()

// isTime reports whether t is a time.Time, or a pointer to one.
func isTime(
	t reflect.Type,
) bool {
	return t != nil && (t == timeType || t.Kind() == reflect.Pointer && t.Elem() == timeType)
}

// isObject reports whether t is an object of the output: a struct, or a
// pointer to one, that is not a time.
func isObject(
	t reflect.Type,
) bool {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	return t.Kind() == reflect.Struct && t != timeType
}

// isList reports whether t is a list of objects.
func isList(
	t reflect.Type,
) bool {
	return t.Kind() == reflect.Slice && isObject(t.Elem())
}

// title is the heading the title tag of t's blank field sets, or "".
func title(
	t reflect.Type,
) string {
	if t.Kind() != reflect.Struct {
		return ""
	}
	for f := range t.Fields() {
		if f.Name == "_" {
			if s := f.Tag.Get("title"); s != "" {
				return s
			}
		}
	}

	return ""
}

// pastTense is what a method did, for the one line of a method with no
// body: the past tense of each verb and action that can answer with none.
var pastTense = map[string]string{
	"create": "created", "update": "updated", "delete": "deleted",
	"admit": "admitted", "reject": "rejected", "revoke": "revoked",
	"drain": "drained", "undrain": "undrained", "retry": "retried",
	"export": "exported", "start": "started", "stop": "stopped",
	"restart": "restarted", "enable": "enabled", "disable": "disabled",
	"install": "installed", "uninstall": "uninstalled", "upload": "uploaded",
	"deploy": "deployed", "undeploy": "undeployed", "pull": "pulled",
	"signal": "signaled", "reboot": "rebooted", "shutdown": "shut down",
	"generate": "generated", "validate": "validated",
}

// said is what the one line of a method with no body says: what Says set,
// or the past tense of the method.
func said(
	m Command,
) string {
	if s := m.Saying(); s != "" {
		return s
	}

	return pastTense[m.MethodName()]
}

// answersNothing reports whether the output type has no field: the method
// answers with no body, a 204.
func answersNothing(
	t reflect.Type,
) bool {
	return t.Kind() == reflect.Struct && len(Columns(t)) == 0
}

// derive prints the answer of m of g from its output type, by the output
// style of the CLI. The answer is the JSON that the method answered.
func derive(
	p Printer,
	g Group,
	m Command,
	in any,
	answer []byte,
) error {
	t := m.Output()
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if len(answer) == 0 || answersNothing(t) {
		p.Verdict(keyOf(g, m, in) + " " + said(m))

		return nil
	}

	var v any
	if err := json.Unmarshal(answer, &v); err != nil {
		return fmt.Errorf("decode %s %s: %w", g.GroupName(), m.MethodName(), err)
	}

	heading := title(t)
	// An object with one field that is an object or a list is an envelope:
	// its value prints, not the envelope.
	if cols := Columns(t); t.Kind() == reflect.Struct && len(cols) == 1 &&
		(isObject(cols[0].Type) || isList(cols[0].Type)) {
		obj, _ := v.(map[string]any)
		v = obj[cols[0].JSON]
		if heading == "" {
			heading = title(element(cols[0].Type))
		}
		if isList(cols[0].Type) {
			if heading == "" {
				heading = plural(sentence(g.GroupName()))
			}
			items, _ := v.([]any)
			list(p, heading, cols[0], items, -1)

			return nil
		}
		t = element(cols[0].Type)
	}

	name := sentence(g.GroupName())
	switch {
	case t.Kind() == reflect.Slice:
		if heading == "" {
			heading = plural(name)
		}
		items, _ := v.([]any)
		list(p, heading, Column{Type: t}, items, -1)
	case listOf(t) != nil:
		if heading == "" {
			heading = plural(name)
		}
		obj, _ := v.(map[string]any)
		items, _ := obj["items"].([]any)
		total := -1
		if n, ok := obj["total_items"].(float64); ok {
			total = int(n)
		}
		list(p, heading, *listOf(t), items, total)
	default:
		if heading == "" {
			heading = name
		}
		obj, _ := v.(map[string]any)
		object(p, heading, t, obj)
	}

	return nil
}

// keyOf is the key of the input, which the one line of a method with no
// body names: the value of the input's key field, or the group's name for
// an input with none.
func keyOf(
	g Group,
	m Command,
	in any,
) string {
	key := m.Key()
	v := reflect.ValueOf(in)
	if key != "" && v.Kind() == reflect.Struct {
		for i := range v.NumField() {
			name, _, _ := strings.Cut(v.Type().Field(i).Tag.Get("json"), ",")
			if name == key {
				return fmt.Sprint(v.Field(i).Interface())
			}
		}
	}

	return g.GroupName()
}

// keyName is the json name of the input field that the one line of a
// method with no body names, or "name" for an input with none.
func keyName(
	m Command,
) string {
	if k := m.Key(); k != "" {
		return k
	}

	return "name"
}

// listOf is the Items field of t when t is a list with its count: Items
// []T, with TotalItems or without, and no other field. It is nil for any
// other type.
func listOf(
	t reflect.Type,
) *Column {
	if t.Kind() != reflect.Struct {
		return nil
	}
	var items *Column
	for _, c := range Columns(t) {
		switch {
		case c.JSON == "items" && c.Type.Kind() == reflect.Slice:
			items = &c
		case c.JSON == "total_items":
		default:
			return nil
		}
	}

	return items
}

// list prints items, a list of the element of c's type, under its heading
// with the count: (3), or (3 of 120) when total is more.
func list(
	p Printer,
	heading string,
	c Column,
	items []any,
	total int,
) {
	counted := fmt.Sprintf("%s (%d)", heading, len(items))
	if total > len(items) {
		counted = fmt.Sprintf("%s (%d of %d)", heading, len(items), total)
	}

	elem := element(c.Type)
	cols := tableColumns(Columns(elem))
	if !isObject(elem) {
		cols = []Column{{JSON: "", Header: "VALUE", Type: elem}}
	}
	// A column with no value in any row says nothing, so it is left out.
	cols = filled(p, cols, items)
	headers := make([]string, 0, len(cols))
	for _, col := range cols {
		headers = append(headers, col.Header)
	}
	rows := make([][]string, 0, len(items))
	for _, item := range items {
		row := make([]string, 0, len(cols))
		for _, col := range cols {
			row = append(row, col.Format(p, cell(col, item)))
		}
		rows = append(rows, row)
	}

	p.Table(Table{Title: counted, Headers: headers, Rows: rows, Empty: c.Empty})
}

// cell is the value of col in item: the field of an object item, or the
// item itself for a list of values.
func cell(
	col Column,
	item any,
) any {
	if obj, ok := item.(map[string]any); ok {
		return obj[col.JSON]
	}

	return item
}

// filled are the columns of cols that have a value in at least one of
// items: an object item's field, or the item itself for a list of values.
func filled(
	p Printer,
	cols []Column,
	items []any,
) []Column {
	out := make([]Column, 0, len(cols))
	for _, col := range cols {
		for _, item := range items {
			if col.Format(p, cell(col, item)) != "" {
				out = append(out, col)

				break
			}
		}
	}

	return out
}

// tableColumns are the columns of a table: each column but one tagged
// table:"-".
func tableColumns(
	cols []Column,
) []Column {
	out := make([]Column, 0, len(cols))
	for _, c := range cols {
		if !c.NoTable {
			out = append(out, c)
		}
	}

	return out
}

// object prints v, an object of type t, under its heading: its values as a
// key-value block, then each object in it as a section and each list in it
// as a table, in the order of the fields.
func object(
	p Printer,
	heading string,
	t reflect.Type,
	v map[string]any,
) {
	p.Heading(heading)
	cols := Columns(t)
	var pairs []string
	for _, c := range cols {
		value, ok := v[c.JSON]
		switch {
		case isObject(c.Type) || isList(c.Type):
		case c.OmitEmpty && c.None == "" && (!ok || noValue(c, value) || zero(value)):
		default:
			pairs = append(pairs, c.Label, c.FormatKV(p, value))
		}
	}
	if len(pairs) > 0 {
		p.KV(pairs...)
	}

	for _, c := range cols {
		value := v[c.JSON]
		switch {
		case isObject(c.Type):
			if obj, ok := value.(map[string]any); ok {
				object(p, c.Label, element(c.Type), obj)
			}
		case isList(c.Type):
			items, _ := value.([]any)
			if c.OmitEmpty && len(items) == 0 {
				continue
			}
			list(p, c.Label, c, items, -1)
		}
	}
}

// zero reports whether v is the zero value of its JSON kind, which a field
// with omitempty leaves out: false, 0, or an empty list or object.
func zero(
	v any,
) bool {
	switch t := v.(type) {
	case bool:
		return !t
	case float64:
		return t == 0
	case []any:
		return len(t) == 0
	case map[string]any:
		return len(t) == 0
	}

	return false
}

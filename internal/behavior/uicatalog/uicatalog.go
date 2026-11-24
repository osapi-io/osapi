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

// Package uicatalog writes the UI's catalog of the behaviors: for each
// behavior, its title, area, platforms and the behaviors that it calls, and
// for each method its verb, permission, route, signature, form, result
// columns, chain and examples. The Behaviors and Plans pages read it, so a
// behavior that changes changes the UI, and nobody edits the UI by hand.
package uicatalog

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/behaviortest"
	clicommand "github.com/osapi-io/osapi/internal/cli/command"
)

// File is where the UI keeps the catalog, from the root of the repository.
const File = "ui/src/lib/catalog.gen.ts"

// Field is one input field of a step's form.
type Field struct {
	Name     string   `json:"name"`
	In       string   `json:"in"`
	Kind     string   `json:"kind"`
	Required bool     `json:"required"`
	Doc      string   `json:"doc"`
	Options  []string `json:"options,omitempty"`
	Secret   bool     `json:"secret,omitempty"`
}

// Column is one field of the rows that a method answers, as the CLI's
// output shows it: its label, and its render tag, such as bytes. A field
// that holds an object or a list of objects has the columns of the object,
// so the UI shows it as the CLI does, with no card of its own.
type Column struct {
	Name    string   `json:"name"`
	Label   string   `json:"label"`
	Doc     string   `json:"doc"`
	Render  string   `json:"render,omitempty"`
	Kind    string   `json:"kind"`
	Columns []Column `json:"columns,omitempty"`
}

// Call is one call that a method makes to another behavior, with the calls
// that it makes in turn, as the examples of the method record them.
type Call struct {
	Behavior string `json:"behavior"`
	Method   string `json:"method"`
	Calls    []Call `json:"calls,omitempty"`
}

// Example is one example of a method: what it does, its input and its
// output.
type Example struct {
	Doc    string `json:"doc"`
	Input  any    `json:"input"`
	Output any    `json:"output,omitempty"`
}

// Method is one method of a behavior, as the UI shows and runs it.
type Method struct {
	Name string `json:"name"`
	// Kind is what the method does to a host: read, write, or action, by
	// the verb of its permission.
	Kind       string   `json:"kind"`
	Doc        string   `json:"doc"`
	Long       string   `json:"long,omitempty"`
	HTTPMethod string   `json:"httpMethod,omitempty"`
	Path       string   `json:"path,omitempty"`
	Permission string   `json:"permission,omitempty"`
	Platforms  []string `json:"platforms"`
	// RunsOn are the host platforms that the method runs on, as the agent
	// decides, of behavior.HostFamilies, and "*" when it runs on a Linux host
	// whose platform is none of them.
	RunsOn     []string  `json:"runsOn"`
	Input      string    `json:"input"`
	Output     string    `json:"output"`
	Fields     []Field   `json:"fields"`
	Columns    []Column  `json:"columns"`
	Calls      []Call    `json:"calls"`
	Examples   []Example `json:"examples"`
	Disruptive bool      `json:"disruptive,omitempty"`
	Disruption string    `json:"disruption,omitempty"`
}

// Behavior is one behavior, as the UI shows it. An internal behavior is in
// the catalog so a chain can name it, but it has no route: only another
// behavior calls it.
type Behavior struct {
	Name      string   `json:"name"`
	Title     string   `json:"title"`
	Area      string   `json:"area"`
	Doc       string   `json:"doc"`
	Internal  bool     `json:"internal"`
	Parent    string   `json:"parent,omitempty"`
	Platforms []string `json:"platforms"`
	Uses      []string `json:"uses"`
	UsedBy    []string `json:"usedBy"`
	Chain     string   `json:"chain"`
	Methods   []Method `json:"methods"`
}

// Runs are the runs of the examples of a behavior, with the calls that each
// made: behaviortest.Chains.
type Runs func(b *behavior.Behavior) []behaviortest.ChainRun

// Catalog is each behavior that the API serves, and each internal behavior,
// in the order of all.
func Catalog(
	all []*behavior.Behavior,
	runs Runs,
) []Behavior {
	out := []Behavior{}
	for _, b := range all {
		if !b.API && !b.Internal {
			continue
		}
		cb := Behavior{
			Name:      b.Name,
			Title:     b.Title,
			Area:      string(b.Area),
			Doc:       b.Doc,
			Internal:  b.Internal,
			Platforms: []string{},
			Uses:      names(b.Requires),
			UsedBy:    names(behavior.RequiredBy(b)),
			Chain:     behavior.Tree(b),
			Methods:   []Method{},
		}
		if b.Parent != nil {
			cb.Parent = b.Parent.Name
		}
		for _, pp := range b.Platforms {
			cb.Platforms = append(cb.Platforms, string(pp.Name))
		}
		calls := map[string][]Call{}
		for _, r := range runs(b) {
			calls[r.Method] = merge(calls[r.Method], r.Root.Steps)
		}
		for _, m := range b.Methods {
			cb.Methods = append(cb.Methods, method(b, m, calls[m.Name]))
		}
		out = append(out, cb)
	}

	return out
}

// otherLinux is a Linux platform that has no family, such as NixOS.
const otherLinux behavior.PlatformName = "*"

// runsOn are the host platforms that m of b runs on.
func runsOn(
	b *behavior.Behavior,
	m *behavior.Method,
) []string {
	out := []string{}
	for _, f := range behavior.HostFamilies {
		if b.RunsOnHost(m, f) {
			out = append(out, string(f))
		}
	}
	if b.RunsOnHost(m, otherLinux) {
		out = append(out, string(otherLinux))
	}

	return out
}

// method is m of b, as the UI shows and runs it.
func method(
	b *behavior.Behavior,
	m *behavior.Method,
	calls []Call,
) Method {
	cm := Method{
		Name:       m.Name,
		Kind:       kindOf(m),
		Doc:        m.Doc,
		Long:       m.Long,
		Platforms:  []string{},
		Input:      behavior.FieldShape(m.In, true),
		Output:     behavior.FieldShape(m.Out, false),
		Fields:     fields(m),
		Columns:    columns(m.Out),
		Calls:      calls,
		Examples:   []Example{},
		Disruptive: m.Disruptive,
		Disruption: m.Disruption,
	}
	if cm.Calls == nil {
		cm.Calls = []Call{}
	}
	if b.API {
		cm.HTTPMethod = m.HTTPMethod()
		cm.Path = b.Path(m)
		cm.Permission = string(b.Permission(m.PermissionVerb()))
	}
	cm.RunsOn = runsOn(b, m)
	for _, pp := range b.Platforms {
		if b.Serves(m, pp.Name) {
			cm.Platforms = append(cm.Platforms, string(pp.Name))
		}
	}
	for _, ex := range m.Examples {
		cm.Examples = append(cm.Examples, Example{
			Doc:    ex.Doc,
			Input:  behavior.Redact(ex.Input),
			Output: ex.Output,
		})
	}

	return cm
}

// kindOf is what m does to a host, by the verb of its permission: a read
// changes nothing, a write changes what the host is configured to be, and
// an action makes something happen, such as a restart or a reboot.
func kindOf(
	m *behavior.Method,
) string {
	switch m.PermissionVerb() {
	case behavior.VerbRead:
		return "read"
	case behavior.VerbWrite:
		return "write"
	default:
		return "action"
	}
}

// merge adds the calls of steps to calls: a call that is there already
// gains the calls under it, and a new one goes at the end, so each call
// that any example made shows once, in the order that they first ran.
func merge(
	calls []Call,
	steps []*behavior.Step,
) []Call {
	for _, s := range steps {
		i := -1
		for j, c := range calls {
			if c.Behavior == s.Behavior && c.Method == s.Method {
				i = j

				break
			}
		}
		if i < 0 {
			calls = append(calls, Call{Behavior: s.Behavior, Method: s.Method})
			i = len(calls) - 1
		}
		calls[i].Calls = merge(calls[i].Calls, s.Steps)
	}

	return calls
}

// names are the names of bs, in order, never nil.
func names(
	bs []*behavior.Behavior,
) []string {
	out := make([]string, 0, len(bs))
	for _, b := range bs {
		out = append(out, b.Name)
	}

	return out
}

// fields is the form of m's input: a path field, a query field of a read,
// or a body field of a write.
func fields(
	m *behavior.Method,
) []Field {
	out := []Field{}
	t := m.In
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == nil || t.Kind() != reflect.Struct {
		return out
	}
	for i := range t.NumField() {
		f := t.Field(i)
		name := behavior.JSONName(f)
		if !f.IsExported() || name == "-" || behavior.SetByBefore(f) {
			continue
		}
		in := "body"
		switch {
		case m.InPath(name):
			in = "path"
		case m.Verb == behavior.VerbRead:
			in = "query"
		}
		rules := f.Tag.Get("validate")
		field := Field{
			Name:     name,
			In:       in,
			Kind:     kind(f.Type),
			Required: strings.Contains(","+rules+",", ",required,"),
			Doc:      f.Tag.Get("doc"),
			Secret:   behavior.Secret(f),
		}
		for _, rule := range strings.Split(rules, ",") {
			if v, ok := strings.CutPrefix(rule, "oneof="); ok {
				field.Options = strings.Fields(v)
			}
		}
		out = append(out, field)
	}

	return out
}

// kind is how the form takes a value of t: a string, a number, a flag, or
// a list of strings.
func kind(
	t reflect.Type,
) string {
	switch t.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Slice:
		return "list"
	case reflect.Map, reflect.Struct:
		return "object"
	default:
		return "string"
	}
}

// columns is the fields of the rows of out, the element of a list or out,
// as the CLI's output reads them: clicommand.Columns, with its labels and
// render tags. A struct field, or a list of structs, has its own columns.
func columns(
	t reflect.Type,
) []Column {
	e := t
	for e.Kind() == reflect.Pointer || e.Kind() == reflect.Slice {
		e = e.Elem()
	}
	docs := map[string]string{}
	if e.Kind() == reflect.Struct {
		for f := range e.Fields() {
			docs[behavior.JSONName(f)] = f.Tag.Get("doc")
		}
	}

	cols := clicommand.Columns(t)
	out := make([]Column, 0, len(cols))
	for _, c := range cols {
		col := Column{
			Name:   c.JSON,
			Label:  c.Label,
			Doc:    docs[c.JSON],
			Render: c.Render,
			Kind:   "value",
		}
		inner := c.Type
		for inner.Kind() == reflect.Pointer {
			inner = inner.Elem()
		}
		switch {
		case inner.Kind() == reflect.Struct:
			col.Kind, col.Columns = "object", columns(inner)
		case inner.Kind() == reflect.Slice && elem(inner).Kind() == reflect.Struct:
			col.Kind, col.Columns = "list", columns(inner)
		case inner.Kind() == reflect.Slice:
			col.Kind = "values"
		}
		out = append(out, col)
	}

	return out
}

// elem is the element of a list, without its pointers.
func elem(
	t reflect.Type,
) reflect.Type {
	t = t.Elem()
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	return t
}

// TypeScript is the catalog as the UI's CATALOG, the union of the names of
// the behaviors, and area.All as the UI's AREAS: the sections of the
// Behaviors page, with their labels, in their order.
func TypeScript(
	catalog []Behavior,
) string {
	var sb strings.Builder
	sb.WriteString("// Code generated by go generate ./internal/behaviors/all. DO NOT EDIT.\n")
	sb.WriteString("// From the behavior registry, by internal/behavior/uicatalog.\n\n")
	sb.WriteString("import type { CatalogBehavior } from \"./catalog\";\n\n")
	names := make([]string, 0, len(catalog))
	sb.WriteString("export const CATALOG: readonly CatalogBehavior[] = [\n")
	for _, b := range catalog {
		// A Behavior is strings, bools, lists of them and the inputs and
		// outputs of examples, which encode, as the tests of each example
		// encode them.
		data, _ := json.Marshal(b)
		fmt.Fprintf(&sb, "  %s,\n", data)
		names = append(names, fmt.Sprintf("%q", b.Name))
	}
	sb.WriteString("];\n\n")
	fmt.Fprintf(&sb, "export type BehaviorName = %s;\n\n", strings.Join(names, " | "))
	sb.WriteString("export const AREAS = [\n")
	for _, a := range area.All {
		// An Area is two strings, which always encode.
		data, _ := json.Marshal(a)
		fmt.Fprintf(&sb, "  %s,\n", data)
	}
	sb.WriteString("] as const;\n\n")
	sb.WriteString("export type AreaName = (typeof AREAS)[number][\"name\"];\n\n")
	sb.WriteString(
		"/** The host platforms that gohai names; runsOn has \"*\" for any other Linux. */\n",
	)
	sb.WriteString("export const HOST_FAMILIES: readonly string[] = [\n")
	for _, f := range behavior.HostFamilies {
		fmt.Fprintf(&sb, "  %q,\n", f)
	}
	sb.WriteString("];\n")

	return sb.String()
}

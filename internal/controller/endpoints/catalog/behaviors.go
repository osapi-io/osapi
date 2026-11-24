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

package catalog

import (
	"reflect"
	"strings"

	"github.com/osapi-io/osapi/internal/validation/rule"
	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/behaviortest"
	clicommand "github.com/osapi-io/osapi/internal/cli/command"
)

// Runs are the runs of the examples of a behavior, with the calls that each
// made: behaviortest.Chains.
type Runs func(b *behavior.Behavior) []behaviortest.ChainRun

// Behaviors is each behavior that the API serves, and each internal behavior,
// in the order of all.
func Behaviors(
	all []*behavior.Behavior,
	runs Runs,
) []CatalogBehavior {
	out := []CatalogBehavior{}
	for _, b := range all {
		if !b.API && !b.Internal {
			continue
		}
		cb := CatalogBehavior{
			Name:      b.Name,
			Title:     b.Title,
			Area:      string(b.Area),
			Doc:       b.Doc,
			Internal:  b.Internal,
			Platforms: []string{},
			Uses:      names(b.Requires),
			UsedBy:    names(behavior.RequiredBy(b)),
			Chain:     behavior.Tree(b),
			Methods:   []CatalogMethod{},
		}
		if b.Parent != nil {
			cb.Parent = b.Parent.Name
		}
		for _, pp := range b.Platforms {
			cb.Platforms = append(cb.Platforms, string(pp.Name))
		}
		calls := map[string][]CatalogCall{}
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
	calls []CatalogCall,
) CatalogMethod {
	cm := CatalogMethod{
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
		Examples:   []CatalogExample{},
		Disruptive: m.Disruptive,
		Disruption: m.Disruption,
	}
	if cm.Calls == nil {
		cm.Calls = []CatalogCall{}
	}
	if b.API {
		cm.HTTPMethod = m.HTTPMethod()
		cm.Path = b.Path(m)
		cm.Permission = b.Permission(m.PermissionVerb())
	}
	cm.RunsOn = runsOn(b, m)
	for _, pp := range b.Platforms {
		if b.Serves(m, pp.Name) {
			cm.Platforms = append(cm.Platforms, string(pp.Name))
		}
	}
	for _, ex := range m.Examples {
		cm.Examples = append(cm.Examples, CatalogExample{
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
) types.MethodKind {
	switch m.PermissionVerb() {
	case types.VerbRead:
		return types.KindRead
	case types.VerbWrite:
		return types.KindWrite
	default:
		return types.KindAction
	}
}

// merge adds the calls of steps to calls: a call that is there already
// gains the calls under it, and a new one goes at the end, so each call
// that any example made shows once, in the order that they first ran.
func merge(
	calls []CatalogCall,
	steps []*behavior.Step,
) []CatalogCall {
	for _, s := range steps {
		i := -1
		for j, c := range calls {
			if c.Behavior == s.Behavior && c.Method == s.Method {
				i = j

				break
			}
		}
		if i < 0 {
			calls = append(calls, CatalogCall{Behavior: s.Behavior, Method: s.Method})
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
) []CatalogField {
	out := []CatalogField{}
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
		case m.Verb == types.VerbRead:
			in = "query"
		}
		rules := f.Tag.Get("validate")
		field := CatalogField{
			Name:     name,
			In:       in,
			Kind:     kind(f.Type),
			Required: strings.Contains(","+rules+",", ",required,"),
			Doc:      f.Tag.Get("doc"),
			Secret:   behavior.Secret(f),
			Options:  rule.Options(rules, f.Type),
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
) []CatalogColumn {
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
	out := make([]CatalogColumn, 0, len(cols))
	for _, c := range cols {
		col := CatalogColumn{
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

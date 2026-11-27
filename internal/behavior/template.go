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

package behavior

import (
	"bytes"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"text/template"
	"text/template/parse"
)

// TemplateData is what a template that a behavior renders can read. A field
// tagged template:"<vars>" holds such a template, and the input field named
// <vars> holds its .Vars.
type TemplateData struct {
	// Facts are the facts of the agent that renders the template, by their
	// JSON names, such as architecture.
	Facts map[string]any
	// Vars are the values that the request gives, from the field that the
	// template tag names.
	Vars map[string]string
	// Hostname is the hostname of the agent.
	Hostname string
}

// templateFields are the fields of TemplateData, the names that a template
// can start a field with.
var templateFields = []string{"Facts", "Vars", "Hostname"}

// Render renders text, a template, with data. A key that data does not have
// is an error, not an empty value.
func Render(
	text string,
	data TemplateData,
) (string, error) {
	t, err := template.New("template").Option("missingkey=error").Parse(text)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render template: %w", err)
	}

	return buf.String(), nil
}

// TemplateField is an input field that holds a template, and the field that
// holds its vars.
type TemplateField struct {
	// Field is the template's field, by its JSON name.
	Field string
	// Vars is the field of the vars, by its JSON name.
	Vars  string
	index []int
	vars  []int
}

// TemplateFields are the fields of t tagged template:"<vars>".
func TemplateFields(
	t reflect.Type,
) []TemplateField {
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == nil || t.Kind() != reflect.Struct {
		return nil
	}

	var out []TemplateField
	for i := range t.NumField() {
		f := t.Field(i)
		vars, ok := f.Tag.Lookup("template")
		if !ok || !f.IsExported() {
			continue
		}
		tf := TemplateField{Field: JSONName(f), Vars: vars, index: f.Index}
		for j := range t.NumField() {
			if v := t.Field(j); JSONName(v) == vars {
				tf.vars = v.Index
			}
		}
		out = append(out, tf)
	}

	return out
}

// CheckTemplates checks each template field of in before a job is queued: it
// parses, it starts each field with .Facts, .Vars or .Hostname, and each
// .Vars.<name> it reads is in its vars. Facts are known only on each host, so
// they are not checked here. The message names the field and what is wrong.
func CheckTemplates(
	in any,
) (string, bool) {
	v := reflect.ValueOf(in)
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return "", true
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return "", true
	}

	for _, tf := range TemplateFields(v.Type()) {
		text, _ := v.FieldByIndex(tf.index).Interface().(string)
		t, err := template.New(tf.Field).Parse(text)
		if err != nil {
			return fmt.Sprintf("%s is not a template: %v", tf.Field, err), false
		}

		var given map[string]string
		if tf.vars != nil {
			given, _ = v.FieldByIndex(tf.vars).Interface().(map[string]string)
		}
		for _, field := range readsOf(t.Root, true) {
			switch {
			case !slices.Contains(templateFields, field[0]):
				return fmt.Sprintf(
					"%s reads .%s; a template reads only .Facts, .Vars and .Hostname",
					tf.Field,
					strings.Join(field, "."),
				), false
			case field[0] == "Vars" && len(field) > 1:
				if _, ok := given[field[1]]; !ok {
					return fmt.Sprintf("%s reads .Vars.%s, which %s does not have",
						tf.Field, field[1], tf.Vars), false
				}
			}
		}
	}

	return "", true
}

// readsOf are the fields that node reads from the template's data, each as
// its path, such as [Vars region]. top is false inside range and with, where
// the dot is no longer the data, so a read of the dot there is not checked;
// a read of $ is, since $ is always the data.
func readsOf(
	node parse.Node,
	top bool,
) [][]string {
	var out [][]string
	switch n := node.(type) {
	case *parse.ListNode:
		if n == nil {
			return nil
		}
		for _, c := range n.Nodes {
			out = append(out, readsOf(c, top)...)
		}
	case *parse.ActionNode:
		out = readsOf(n.Pipe, top)
	case *parse.PipeNode:
		for _, c := range n.Cmds {
			out = append(out, readsOf(c, top)...)
		}
	case *parse.CommandNode:
		// index .Vars "x" reads the var x, as .Vars.x does.
		if len(n.Args) == 3 {
			fn, isFn := n.Args[0].(*parse.IdentifierNode)
			field, isField := dataPath(n.Args[1], top)
			key, isKey := n.Args[2].(*parse.StringNode)
			if isFn && fn.Ident == "index" && isField && isKey && len(field) == 1 {
				return [][]string{{field[0], key.Text}}
			}
		}
		for _, a := range n.Args {
			out = append(out, readsOf(a, top)...)
		}
	case *parse.FieldNode, *parse.VariableNode:
		if path, ok := dataPath(n, top); ok {
			out = append(out, path)
		}
	case *parse.IfNode:
		out = branch(&n.BranchNode, top, top)
	case *parse.RangeNode:
		out = branch(&n.BranchNode, top, false)
	case *parse.WithNode:
		out = branch(&n.BranchNode, top, false)
	}

	return out
}

// dataPath is the path that node reads from the template's data: .Vars.x
// where the dot is the data, and $.Vars.x anywhere, since $ is always the
// data. A variable that a template declares, such as $k, reads nothing from
// it.
func dataPath(
	node parse.Node,
	top bool,
) ([]string, bool) {
	switch n := node.(type) {
	case *parse.FieldNode:
		return n.Ident, top
	case *parse.VariableNode:
		if len(n.Ident) > 1 && n.Ident[0] == "$" {
			return n.Ident[1:], true
		}
	}

	return nil, false
}

// branch is readsOf for an if, range or with: its pipe, with the dot as it
// was, then its bodies, read only when body is true.
func branch(
	n *parse.BranchNode,
	top bool,
	body bool,
) [][]string {
	out := readsOf(n.Pipe, top)
	out = append(out, readsOf(n.List, body)...)

	return append(out, readsOf(n.ElseList, top)...)
}

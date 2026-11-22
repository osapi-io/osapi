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

// Package uiblocks writes the UI's blocks of the behaviors: for each method
// that the API serves, the form that its input makes, the request that runs
// it, the permission that it needs, and the columns of the rows that it
// answers. The Configure page builds its behavior blocks from it, so a
// behavior that changes changes its block, and nobody edits the UI by hand.
package uiblocks

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/behavior/command"
)

// File is where the UI keeps the blocks, from the root of the repository.
const File = "ui/src/lib/blocks.gen.ts"

// Field is one input field of a block's form.
type Field struct {
	Name     string   `json:"name"`
	In       string   `json:"in"`
	Kind     string   `json:"kind"`
	Required bool     `json:"required"`
	Doc      string   `json:"doc"`
	Options  []string `json:"options,omitempty"`
	Secret   bool     `json:"secret,omitempty"`
}

// Column is one field of the rows that a block answers.
type Column struct {
	Name string `json:"name"`
	Doc  string `json:"doc"`
}

// Block is one method of a behavior, as the UI runs it.
type Block struct {
	Type        string   `json:"type"`
	Behavior    string   `json:"behavior"`
	Title       string   `json:"title"`
	Method      string   `json:"method"`
	Label       string   `json:"label"`
	Description string   `json:"description"`
	HTTPMethod  string   `json:"httpMethod"`
	Path        string   `json:"path"`
	Permission  string   `json:"permission"`
	Write       bool     `json:"write"`
	Fields      []Field  `json:"fields"`
	Columns     []Column `json:"columns"`
}

// Blocks is the block of each method of each behavior that the API serves,
// in the order of all.
func Blocks(
	all []*behavior.Behavior,
) []Block {
	var out []Block
	for _, b := range all {
		if !b.API {
			continue
		}
		for _, m := range b.Methods {
			out = append(out, Block{
				Type:        b.Name + "-" + m.Name,
				Behavior:    b.Name,
				Title:       command.Title(b),
				Method:      m.Name,
				Label:       strings.ToUpper(m.Name[:1]) + m.Name[1:],
				Description: m.Doc,
				HTTPMethod:  m.HTTPMethod(),
				Path:        b.Path(m),
				Permission:  b.Permission(m.Verb),
				Write:       m.Verb == behavior.VerbWrite,
				Fields:      fields(m),
				Columns:     columns(m.Out),
			})
		}
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
		if !f.IsExported() || name == "-" {
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

// columns is the fields of the rows of out: the element of a list, or out.
func columns(
	t reflect.Type,
) []Column {
	out := []Column{}
	for t != nil && (t.Kind() == reflect.Pointer || t.Kind() == reflect.Slice) {
		t = t.Elem()
	}
	if t == nil || t.Kind() != reflect.Struct {
		return out
	}
	for i := range t.NumField() {
		f := t.Field(i)
		name := behavior.JSONName(f)
		if !f.IsExported() || name == "-" {
			continue
		}
		out = append(out, Column{Name: name, Doc: f.Tag.Get("doc")})
	}

	return out
}

// TypeScript is blocks as the UI's BEHAVIOR_BLOCKS, keyed by block type,
// and the union of the behavior names.
func TypeScript(
	blocks []Block,
) string {
	var sb strings.Builder
	sb.WriteString("// Code generated by go generate ./internal/behaviors/all. DO NOT EDIT.\n")
	sb.WriteString("// From the behavior registry, by internal/behavior/uiblocks.\n\n")
	sb.WriteString("import type { BehaviorBlockDef } from \"./behavior-block\";\n\n")
	names := []string{}
	seen := map[string]bool{}
	sb.WriteString("export const BEHAVIOR_BLOCKS = {\n")
	for _, b := range blocks {
		// A Block is strings, bools and lists of them, which always encode.
		data, _ := json.Marshal(b)
		fmt.Fprintf(&sb, "  %q: %s,\n", b.Type, data)
		if !seen[b.Behavior] {
			seen[b.Behavior] = true
			names = append(names, fmt.Sprintf("%q", b.Behavior))
		}
	}
	sb.WriteString("} as const satisfies Record<string, BehaviorBlockDef>;\n\n")
	sb.WriteString("export type BehaviorBlockType = keyof typeof BEHAVIOR_BLOCKS;\n\n")
	fmt.Fprintf(&sb, "export type BehaviorName = %s;\n", strings.Join(names, " | "))

	return sb.String()
}

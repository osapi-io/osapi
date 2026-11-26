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
	"fmt"
	"reflect"

	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// ReadMethod is the read of b, or of b's parent, whose input type is the
// type of input, and the behavior that has it, or nil. A ReadBack names
// its read this way, as After names its method.
func ReadMethod(
	b *Behavior,
	input any,
) (*Behavior, *Method) {
	for _, on := range []*Behavior{b, b.Parent} {
		if on == nil {
			continue
		}
		for _, m := range on.Methods {
			if m.Verb == types.VerbRead && m.In == reflect.TypeOf(input) {
				return on, m
			}
		}
	}

	return nil, nil
}

// ReadRow is the type of one row of m's output: the element of a list, or
// the output itself.
func ReadRow(
	m *Method,
) reflect.Type {
	if m.Out.Kind() == reflect.Slice {
		return m.Out.Elem()
	}

	return m.Out
}

// readBackProblems are the RuleReadBack problems of m: a write example
// that runs on the live host with no Reads, ReadsNone or NoReadBack, and a
// read back that names no read, a want of another type, or a field that
// the want does not have.
func readBackProblems(
	b *Behavior,
	m *Method,
) []string {
	var out []string
	for _, ex := range m.Examples {
		live := b.CLI && m.Verb == types.VerbWrite && !m.Disruptive && ex.NotLive == ""
		if live && len(ex.ReadBacks) == 0 && ex.NoReadBack == "" && m.NoReadBack == "" {
			out = append(out, fmt.Sprintf(
				"example %q is a write with no read of its result: add behavior.Reads or "+
					"behavior.ReadsNone, or behavior.NoReadBack with the reason",
				ex.Doc,
			))
		}
		if m.Verb != types.VerbWrite && len(ex.ReadBacks) > 0 {
			out = append(out, fmt.Sprintf("example %q: a read back is for a write example", ex.Doc))
		}
		for _, rb := range ex.ReadBacks {
			out = append(out, readBackProblem(b, ex, rb)...)
		}
	}

	return out
}

// readBackProblem are the problems of one read back of ex.
func readBackProblem(
	b *Behavior,
	ex Example,
	rb ReadBack,
) []string {
	_, read := ReadMethod(b, rb.Input)
	if read == nil {
		return []string{fmt.Sprintf(
			"example %q: no read of %s or its parent takes %T", ex.Doc, b.Name, rb.Input,
		)}
	}
	row := ReadRow(read)
	if got := reflect.TypeOf(rb.Want); got != row {
		return []string{fmt.Sprintf(
			"example %q: the read back of %s wants a %v, and %s answers a %v",
			ex.Doc, read.Name, got, read.Name, row,
		)}
	}
	var out []string
	names := map[string]bool{}
	for i := range row.NumField() {
		names[JSONName(row.Field(i))] = true
	}
	for _, f := range rb.Fields {
		if !names[f] {
			out = append(out, fmt.Sprintf(
				"example %q: the read back of %s names %s, which is not a field of %v",
				ex.Doc, read.Name, f, row,
			))
		}
	}
	if len(rb.Fields) == 0 && reflect.ValueOf(rb.Want).IsZero() {
		out = append(out, fmt.Sprintf(
			"example %q: the read back of %s wants a zero %v, which compares nothing: "+
				"name the fields to compare", ex.Doc, read.Name, row,
		))
	}

	return out
}

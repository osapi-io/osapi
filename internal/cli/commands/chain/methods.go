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

package chain

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/cli/command"
)

// maxValue is the length of an input or an output in the table. --json has
// each in full.
const maxValue = 70

// List is each behavior that calls others or that others call.
func List(
	_ context.Context,
	d Deps,
	_ ListInput,
) (ListOutput, error) {
	out := ListOutput{Items: []Summary{}}
	for _, b := range d.Behaviors() {
		usedBy := names(behavior.RequiredBy(b))
		if len(b.Requires) == 0 && len(usedBy) == 0 {
			continue
		}
		out.Items = append(out.Items, Summary{
			Behavior: b.Name,
			Internal: b.Internal,
			Chain:    behavior.Tree(b),
			Calls:    names(b.Requires),
			UsedBy:   usedBy,
		})
	}

	return out, nil
}

// Get is the chain of one behavior, with the calls of each example.
func Get(
	_ context.Context,
	d Deps,
	in GetInput,
) (GetOutput, error) {
	var b *behavior.Behavior
	for _, candidate := range d.Behaviors() {
		if candidate.Name == in.Name {
			b = candidate
		}
	}
	if b == nil {
		return GetOutput{}, fmt.Errorf(
			"no behavior %q: osapi behavior chain list shows each chain", in.Name)
	}

	out := GetOutput{
		Behavior: b.Name,
		Chain:    behavior.Tree(b),
		Calls:    names(b.Requires),
		UsedBy:   names(behavior.RequiredBy(b)),
		Examples: []ExampleCalls{},
		Runs:     d.Runs(b),
	}
	for _, r := range out.Runs {
		if len(r.Root.Steps) == 0 {
			continue
		}
		ex := ExampleCalls{Example: r.Example + " (" + r.Method + ")"}
		var walk func(s *behavior.Step, depth int)
		walk = func(s *behavior.Step, depth int) {
			call := s.Behavior + " " + s.Method
			// The table trims leading spaces, so a deeper call is marked
			// with a dot for each level above it.
			if depth > 0 {
				call = strings.Repeat("· ", depth-1) + "└─ " + call
			}
			gets := fields(s.Output)
			if s.Error != "" {
				gets = "error: " + s.Error
			}
			ex.Calls = append(ex.Calls, Call{Call: call, Sends: fields(s.Input), Gets: gets})
			for _, c := range s.Steps {
				walk(c, depth+1)
			}
		}
		walk(r.Root, 0)
		out.Examples = append(out.Examples, ex)
	}

	return out, nil
}

// Validate checks every chain against the chain rules, and fails when one
// breaks.
func Validate(
	_ context.Context,
	d Deps,
	_ ValidateInput,
) (ValidateOutput, error) {
	all := d.Behaviors()
	problems := d.Problems()
	for _, b := range all {
		problems = append(problems, d.RunProblems(b)...)
	}
	out := ValidateOutput{Behaviors: len(all), Problems: problems, Valid: len(problems) == 0}
	if out.Valid {
		return out, nil
	}
	out.Fixes = map[string]any{}
	for _, p := range problems {
		out.Fixes[p.Rule] = p.Fix
	}

	return out, command.Fail()
}

// Spec is the spec of a chain, with the building blocks of the binary.
func Spec(
	_ context.Context,
	d Deps,
	_ SpecInput,
) (SpecOutput, error) {
	return SpecOutput{Kinds: d.Spec().Sections()}, nil
}

// names are the names of bs.
func names(
	bs []*behavior.Behavior,
) []string {
	out := make([]string, 0, len(bs))
	for _, b := range bs {
		out = append(out, b.Name)
	}

	return out
}

// fields is v as the fields that have a value, key=value in key order: a
// text of several lines is its count of lines, and the whole is cut to
// maxValue for the table.
func fields(
	v any,
) string {
	data, _ := json.Marshal(v)
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return cut(string(data))
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		if text := value(m[k]); text != "" {
			parts = append(parts, k+"="+text)
		}
	}

	return cut(strings.Join(parts, " "))
}

// value is one field's value as fields shows it, or "" for one with none.
func value(
	v any,
) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		if n := strings.Count(strings.TrimRight(t, "\n"), "\n") + 1; n > 1 {
			return fmt.Sprintf("<%d lines>", n)
		}
		if strings.ContainsAny(t, " ") {
			return strconv.Quote(t)
		}

		return t
	case []any:
		if len(t) == 0 {
			return ""
		}
		items := make([]string, 0, len(t))
		for _, item := range t {
			items = append(items, value(item))
		}

		return "[" + strings.Join(items, ", ") + "]"
	case map[string]any:
		if len(t) == 0 {
			return ""
		}
		data, _ := json.Marshal(t)

		return string(data)
	default:
		return fmt.Sprint(t)
	}
}

// cut is s cut to maxValue, with … where it was cut.
func cut(
	s string,
) string {
	if r := []rune(s); len(r) > maxValue {
		return string(r[:maxValue-1]) + "…"
	}

	return s
}

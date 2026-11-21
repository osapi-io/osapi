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
	"fmt"
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
			Chain:    tree(b),
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
		Chain:    tree(b),
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
		for _, c := range r.Calls(maxValue) {
			ex.Calls = append(ex.Calls, Call(c))
		}
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

// tree is b and the behaviors that it calls, down the chain:
// dns → netplan → managedfile, or status → (hostname, uptime) when a
// behavior calls more than one.
func tree(
	b *behavior.Behavior,
) string {
	if len(b.Requires) == 0 {
		return b.Name
	}
	parts := make([]string, 0, len(b.Requires))
	for _, r := range b.Requires {
		parts = append(parts, tree(r))
	}
	if len(parts) == 1 {
		return b.Name + " → " + parts[0]
	}

	return b.Name + " → (" + strings.Join(parts, ", ") + ")"
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

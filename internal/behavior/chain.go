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
	"slices"
	"strings"
)

// Requires declares the behaviors this one is built on. The agent builds
// each required behavior's provider before this one's and hands it to this
// one's builder, which reads it with host.Require; this behavior's provider
// then calls the required behavior's methods as Go calls, on the same host in
// the same job. Chaining behaviors across hosts or steps is a workflow, not a
// requirement.
func Requires(
	behaviors ...*Behavior,
) Option {
	return optionFunc(func(b *Behavior) { b.Requires = append(b.Requires, behaviors...) })
}

// Internal marks a behavior that only other behaviors use. It gets no route,
// no command, no spec and no permissions; behavior list and behavior get still
// show it, with what uses it. A primitive can start internal and be exposed
// later by removing the option.
var Internal Option = optionFunc(func(b *Behavior) { b.Internal = true })

// RequiredBy are the registered behaviors that require b, in registry order.
func RequiredBy(
	b *Behavior,
) []*Behavior {
	var out []*Behavior
	for _, other := range All() {
		for _, r := range other.Requires {
			if r == b {
				out = append(out, other)

				break
			}
		}
	}

	return out
}

// chainProblems are the requirements of the registered behaviors that cannot
// be built: one that is not registered, and a cycle.
func chainProblems(
	all []*Behavior,
) []Problem {
	var problems []Problem
	for _, b := range all {
		for _, r := range b.Requires {
			if r == nil || Lookup(r.Name) != r {
				name := "<nil>"
				if r != nil {
					name = r.Name
				}
				problems = append(problems, Problem{
					Behavior: b.Name,
					Rule:     RuleRequires,
					Message:  "requires " + name + ", which is not registered",
				})
			}
		}
		if cycle := cycleFrom(b, b, nil); cycle != nil {
			problems = append(problems, Problem{
				Behavior: b.Name,
				Rule:     RuleRequires,
				Message:  "requires itself: " + strings.Join(cycle, " -> "),
			})
		}
	}

	return problems
}

// cycleFrom is the path from at back to start through requirements, or nil
// when there is none. path is the way taken so far.
func cycleFrom(
	start *Behavior,
	at *Behavior,
	path []string,
) []string {
	path = append(path, at.Name)
	for _, r := range at.Requires {
		if r == nil {
			continue
		}
		if r == start {
			return append(path, r.Name)
		}
		if slices.Contains(path, r.Name) {
			continue
		}
		if cycle := cycleFrom(start, r, path); cycle != nil {
			return cycle
		}
	}

	return nil
}

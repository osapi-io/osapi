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

package guide

import (
	"fmt"
	"strings"
)

// Term is one part of a DSL or format: how it is written, what it makes
// osapi do, and when to use it.
type Term struct {
	// Kind groups the term, as one of the Spec's Kinds names it.
	Kind string `json:"kind"`
	// Name is how the term is written, such as Key("field").
	Name string `json:"name"`
	// Does is what the term makes osapi do.
	Does string `json:"does"`
	// When is when to use it.
	When string `json:"when"`
}

// Kind is a group of terms, with the title that spec shows above it.
type Kind struct {
	Kind  string `json:"kind"`
	Title string `json:"title"`
}

// Spec is every term of a DSL or format, in groups.
type Spec struct {
	Kinds []Kind `json:"kinds"`
	Terms []Term `json:"terms"`
}

// Of are the terms of kind, in order.
func (s Spec) Of(
	kind string,
) []Term {
	var out []Term
	for _, t := range s.Terms {
		if t.Kind == kind {
			out = append(out, t)
		}
	}

	return out
}

// Check is one rule that a validate command applies: what is required, and
// how to fix what breaks it.
type Check struct {
	// Rule is the name that a problem reports.
	Rule string `json:"rule"`
	// Requires is what must be true.
	Requires string `json:"requires"`
	// Fix is how to fix what breaks the rule.
	Fix string `json:"fix"`
}

// Checks are the rules of a validate command, in order.
type Checks []Check

// Fix is the fix of rule, or "" for a rule without a check.
func (c Checks) Fix(
	rule string,
) string {
	for _, k := range c {
		if k.Rule == rule {
			return k.Fix
		}
	}

	return ""
}

// Help is the list of checks for the help of a validate command, one item
// for each rule.
func (c Checks) Help() string {
	var sb strings.Builder
	for _, k := range c {
		fmt.Fprintf(&sb, "- `%s`: %s Fix: %s\n", k.Rule, k.Requires, k.Fix)
	}

	return sb.String()
}

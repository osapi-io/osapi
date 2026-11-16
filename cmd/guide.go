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

package cmd

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/guide"
)

// newSpecCommand is the spec command of a command group: it prints the spec
// that spec returns, the terms of the group's DSL or format, as text under a
// heading for each kind, or as the whole Spec with --json. A group that
// teaches an author, or an agent, how to build for it declares a guide.Spec
// and adds this command. spec is a function, so a part of the spec can come
// from what the binary has, such as its building blocks.
func newSpecCommand(
	spec func() guide.Spec,
	short string,
	long string,
	example string,
) *cobra.Command {
	return &cobra.Command{
		Use:     "spec",
		Short:   short,
		Long:    long,
		Example: example,
		Args:    cobra.NoArgs,
		Run: func(_ *cobra.Command, _ []string) {
			spec := spec()
			if jsonOutput {
				out, _ := json.Marshal(spec)
				cli.PrintJSON(out)

				return
			}
			for _, k := range spec.Kinds {
				cli.PrintHeading(k.Title)
				cli.PrintTerms(spec.Of(k.Kind))
			}
		},
	}
}

// printFixes prints, under a Fixes heading, the fix of each rule in rules,
// one time each, from checks.
func printFixes(
	checks guide.Checks,
	rules []string,
) {
	seen := map[string]bool{}
	var terms []guide.Term
	for _, r := range rules {
		if seen[r] || checks.Fix(r) == "" {
			continue
		}
		seen[r] = true
		terms = append(terms, guide.Term{Name: r, Does: checks.Fix(r)})
	}
	if len(terms) == 0 {
		return
	}
	cli.PrintHeading("Fixes")
	cli.PrintTerms(terms)
}

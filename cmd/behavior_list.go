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
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/cli"
)

// behaviorSummary is one behavior in `behavior list --json`.
type behaviorSummary struct {
	Name    string   `json:"name"`
	Group   string   `json:"group,omitempty"`
	Doc     string   `json:"doc"`
	Methods []string `json:"methods"`
	// Internal is set for a behavior only other behaviors use.
	Internal bool `json:"internal,omitempty"`
}

var behaviorListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the behaviors this binary has",
	Example: `  osapi behavior list
  osapi behavior list --json`,
	Args: cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		all := behavior.All()

		summaries := make([]behaviorSummary, 0, len(all))
		for _, b := range all {
			names := make([]string, 0, len(b.Methods))
			for _, m := range b.Methods {
				names = append(names, m.Name)
			}
			summaries = append(summaries, behaviorSummary{
				Name: b.Name, Group: b.Group, Doc: b.Doc, Methods: names, Internal: b.Internal,
			})
		}

		if jsonOutput {
			out, _ := json.Marshal(summaries)
			cli.PrintJSON(out)

			return
		}

		if len(summaries) == 0 {
			cli.PrintHeading("Behaviors (0)")
			cli.PrintNote("None.")

			return
		}

		rows := make([][]string, 0, len(summaries))
		for _, s := range summaries {
			doc := s.Doc
			if s.Internal {
				doc += " (internal)"
			}
			rows = append(rows, []string{s.Name, s.Group, strings.Join(s.Methods, ", "), doc})
		}
		cli.PrintCompactTable([]cli.Section{{
			Title:     fmt.Sprintf("Behaviors (%d)", len(summaries)),
			Headers:   []string{"NAME", "GROUP", "METHODS", "DESCRIPTION"},
			Rows:      rows,
			Reference: true,
		}})
	},
}

func init() {
	behaviorCmd.AddCommand(behaviorListCmd)
}

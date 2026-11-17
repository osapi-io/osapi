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

	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/endpoint"
)

// endpointSummary is one group in `endpoint list --json`.
type endpointSummary struct {
	Name    string   `json:"name"`
	Doc     string   `json:"doc"`
	Methods []string `json:"methods"`
}

var endpointListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the endpoint groups this binary has",
	Example: `  osapi endpoint list
  osapi endpoint list --json`,
	Args: cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		all := endpoint.All()
		summaries := make([]endpointSummary, 0, len(all))
		for _, g := range all {
			names := make([]string, 0, len(g.Methods))
			for _, m := range g.Methods {
				names = append(names, m.Name)
			}
			summaries = append(summaries, endpointSummary{Name: g.Name, Doc: g.Doc, Methods: names})
		}

		if jsonOutput {
			out, _ := json.Marshal(summaries)
			cli.PrintJSON(out)

			return
		}

		rows := make([][]string, 0, len(summaries))
		for _, s := range summaries {
			rows = append(rows, []string{s.Name, strings.Join(s.Methods, ", "), s.Doc})
		}
		cli.PrintCompactTable([]cli.Section{{
			Title:     fmt.Sprintf("Endpoint groups (%d)", len(summaries)),
			Headers:   []string{"NAME", "METHODS", "DESCRIPTION"},
			Rows:      rows,
			Reference: true,
		}})
	},
}

func init() {
	endpointCmd.AddCommand(endpointListCmd)
}

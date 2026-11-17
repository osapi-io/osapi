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
	"os"

	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
)

// endpointValidateResult is what `controller endpoint validate --json` prints.
type endpointValidateResult struct {
	Valid    bool               `json:"valid"`
	Problems []endpoint.Problem `json:"problems"`
}

var endpointValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Check every endpoint group against the rules of a definition",
	Long: "It exits 1 when a rule fails, and names the group, the method, the rule,\n" +
		"what is wrong and the fix. Each problem in `--json` has `group`, `method`\n" +
		"(when the problem is in one), `rule`, `message` and `fix`.\n" +
		"`osapi controller endpoint spec` describes each part of a group.\n\n" +
		"The rules:\n\n" + endpoint.Checks.Help(),
	Example: `  osapi controller endpoint validate
  osapi controller endpoint validate --json`,
	Args: cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		result := endpointValidateResult{Problems: []endpoint.Problem{}}
		for _, g := range endpoint.All() {
			result.Problems = append(result.Problems, endpoint.Validate(g)...)
		}
		result.Valid = len(result.Problems) == 0

		switch {
		case jsonOutput:
			out, _ := json.Marshal(result)
			cli.PrintJSON(out)
		case result.Valid:
			n := len(endpoint.All())
			noun := "groups"
			if n == 1 {
				noun = "group"
			}
			fmt.Println(cli.Success(os.Stdout, fmt.Sprintf("%d endpoint %s valid", n, noun)))
		default:
			rows := make([][]string, 0, len(result.Problems))
			rules := make([]string, 0, len(result.Problems))
			for _, p := range result.Problems {
				rows = append(rows, []string{p.Group, p.Method, p.Rule, p.Message})
				rules = append(rules, p.Rule)
			}
			cli.PrintCompactTable([]cli.Section{{
				Title:     "Problems",
				Headers:   []string{"GROUP", "METHOD", "RULE", "MESSAGE"},
				Rows:      rows,
				Reference: true,
			}})
			printFixes(endpoint.Checks, rules)
		}

		if !result.Valid {
			cliExit(cli.ExitError)
		}
	},
}

func init() {
	endpointCmd.AddCommand(endpointValidateCmd)
}

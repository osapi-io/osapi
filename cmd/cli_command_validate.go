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
	"github.com/osapi-io/osapi/internal/cli/command"
)

// commandValidateResult is what `cli command validate --json` prints.
type commandValidateResult struct {
	Valid    bool              `json:"valid"`
	Problems []command.Problem `json:"problems"`
}

var cliCommandValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Check every local group against the rules of a definition, and the output of every group",
	Long: "It exits 1 when a rule fails, and names the group, the method, the rule,\n" +
		"what is wrong and the fix. Each problem in `--json` has `group`, `method`\n" +
		"(when the problem is in one), `rule`, `message` and `fix`.\n" +
		"`osapi cli command spec` describes each part of a group.\n\n" +
		"The rules:\n\n" + command.Checks.Help(),
	Example: `  osapi cli command validate
  osapi cli command validate --json`,
	Args: cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		result := commandValidateResult{Problems: []command.Problem{}}
		for _, g := range command.All() {
			result.Problems = append(result.Problems, command.Validate(g)...)
		}
		result.Problems = append(result.Problems, command.ValidateOutput()...)
		result.Valid = len(result.Problems) == 0

		switch {
		case jsonOutput:
			out, _ := json.Marshal(result)
			cli.PrintJSON(out)
		case result.Valid:
			fmt.Println(cli.Success(os.Stdout, fmt.Sprintf(
				"%s valid, and the output of %s",
				cli.Plural(len(command.All()), "local group", "local groups"),
				cli.Plural(len(command.Groups()), "group", "groups"),
			)))
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
			printFixes(command.Checks, rules)
		}

		if !result.Valid {
			cliExit(cli.ExitError)
		}
	},
}

func init() {
	cliCommandCmd.AddCommand(cliCommandValidateCmd)
}

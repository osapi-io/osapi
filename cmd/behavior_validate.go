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
	"github.com/osapi-io/osapi/pkg/behavior"
)

// validateResult is what `behavior validate --json` prints.
type validateResult struct {
	Valid    bool               `json:"valid"`
	Problems []behavior.Problem `json:"problems"`
}

var behaviorValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Check every behavior against the rules the platform relies on",
	Long: `Checks each definition: a placement, a description on the behavior and
each method, a key that is a field of the input, input and output types that
have a JSON Schema, no output field that collides with the per-host result,
and a changed field on every write. Exits 1 when anything fails.`,
	Example: `  osapi behavior validate
  osapi behavior validate --json`,
	Args: cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		problems := behavior.Validate()
		result := validateResult{Valid: len(problems) == 0, Problems: problems}
		if result.Problems == nil {
			result.Problems = []behavior.Problem{}
		}

		if jsonOutput {
			out, _ := json.Marshal(result)
			cli.PrintJSON(out)
		} else if result.Valid {
			n := len(behavior.All())
			noun := "behaviors"
			if n == 1 {
				noun = "behavior"
			}
			fmt.Println(cli.Success(os.Stdout, fmt.Sprintf("%d %s valid", n, noun)))
		} else {
			rows := make([][]string, 0, len(problems))
			for _, p := range problems {
				rows = append(rows, []string{p.Behavior, p.Method, p.Rule, p.Message})
			}
			cli.PrintCompactTable([]cli.Section{{
				Title:     "Problems",
				Headers:   []string{"BEHAVIOR", "METHOD", "RULE", "MESSAGE"},
				Rows:      rows,
				Reference: true,
			}})
		}

		if !result.Valid {
			cliExit(cli.ExitError)
		}
	},
}

// cliExit is os.Exit, replaced in tests.
var cliExit = os.Exit

func init() {
	behaviorCmd.AddCommand(behaviorValidateCmd)
}

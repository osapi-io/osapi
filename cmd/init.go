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
	"strings"

	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/skills"
)

// initResult is what `osapi init --json` prints.
type initResult struct {
	Version string          `json:"version"`
	Tools   []string        `json:"tools"`
	Files   []skills.Result `json:"files"`
}

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Install osapi's agent skills into a repository",
	Long: "The command writes the skills in this binary for each `--tool`. It also\n" +
		"writes a short section, with links to the skills, in the instructions file\n" +
		"of that tool: `CLAUDE.md` or `AGENTS.md`. It does not change text outside\n" +
		"the osapi section. Run it again after you upgrade osapi. It writes the\n" +
		"skills again to agree with the binary. It removes the files that an older\n" +
		"version installed.",
	Example: `  osapi init
  osapi init --tool claude --tool codex
  osapi init ./ops-repo --json`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := "."
		if len(args) == 1 {
			root = args[0]
		}

		names, _ := cmd.Flags().GetStringSlice("tool")
		tools := make([]skills.Tool, 0, len(names))
		for _, n := range names {
			t, ok := skills.Tools[n]
			if !ok {
				return fmt.Errorf(
					"invalid argument %q for \"--tool\": want one of %s",
					n, strings.Join(skills.ToolNames(), ", "),
				)
			}
			tools = append(tools, t)
		}

		src := skills.Installed()
		results, err := skills.Install(appFs, root, tools, src, displayVersion())
		if err != nil {
			return err
		}

		if jsonOutput {
			out, _ := json.Marshal(initResult{
				Version: displayVersion(), Tools: names, Files: results,
			})
			cli.PrintJSON(out)

			return nil
		}

		rows := make([][]string, 0, len(results))
		changed := 0
		for _, r := range results {
			rows = append(rows, []string{r.Path, string(r.Status)})
			if r.Status != skills.Unchanged {
				changed++
			}
		}
		cli.PrintCompactTable([]cli.Section{{
			Title:     "Skills",
			Headers:   []string{"FILE", "STATUS"},
			Rows:      rows,
			Reference: true,
		}})
		fmt.Println()
		fmt.Println("  " + cli.Success(os.Stdout, fmt.Sprintf(
			"osapi %s skills installed for %s, %d files changed",
			displayVersion(), strings.Join(names, " and "), changed,
		)))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	skipConfigCmds["init"] = true
	initCmd.Flags().StringSlice(
		"tool", []string{"claude"},
		"Agent tool to install for: "+strings.Join(skills.ToolNames(), ", ")+" (repeatable)",
	)
}

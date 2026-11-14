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

	"github.com/osapi-io/osapi/internal/behaviorcmd"
	"github.com/osapi-io/osapi/internal/cli"
)

// newResult is what `behavior new --json` prints.
type newResult struct {
	Name  string   `json:"name"`
	Files []string `json:"files"`
	Next  []string `json:"next"`
}

var behaviorNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Write a new behavior's standard layout in an osapi checkout",
	Long: `Writes internal/behaviors/<name>/ with the standard files, each with its test
where it has code: the definition, its settings, the input and output types,
the methods, the provider interface and its mocks, and a Debian provider that
keeps a value in a file. It registers the behavior in internal/behaviors/all.
Writes nothing when the name is taken.

After go generate writes the provider's mock and the behavior's gen/
directory (its API handlers, CLI commands, spec and their tests), the result
builds, validates and passes every test before you change anything.
Change the types, the provider and the examples to what the behavior manages;
each example's declared effects are the provider's tests.
go doc ./internal/behavior lists every option; internal/behaviors/sysctl is the reference.`,
	Example: `  osapi behavior new --name motd --description "The message shown at login"
  osapi behavior new --name motd --description "The message shown at login" --root ~/src/osapi
  osapi behavior new --group network --name neighbors --description "The host's ARP neighbors"`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")
		doc, _ := cmd.Flags().GetString("description")
		root, _ := cmd.Flags().GetString("root")
		group, _ := cmd.Flags().GetString("group")

		files, err := behaviorcmd.Scaffold(root, group, name, doc)
		if err != nil {
			return err
		}

		path := name
		if group != "" {
			path = group + "/" + name
		}
		next := []string{
			"go generate ./internal/behaviors/" + path + "/mocks ./internal/behaviors/all",
			"change the types, the provider and the examples in internal/behaviors/" + path,
			"go generate ./internal/behaviors/all",
			"osapi behavior validate",
			"just test",
		}

		if jsonOutput {
			out, _ := json.Marshal(newResult{Name: name, Files: files, Next: next})
			cli.PrintJSON(out)

			return nil
		}

		rows := make([][]string, 0, len(files))
		for _, f := range files {
			rows = append(rows, []string{f})
		}
		steps := make([][]string, 0, len(next))
		for _, n := range next {
			steps = append(steps, []string{n})
		}
		cli.PrintCompactTable([]cli.Section{
			{Title: "Written", Headers: []string{"FILE"}, Rows: rows, Reference: true},
			{Title: "Next", Headers: []string{"STEP"}, Rows: steps, Reference: true},
		})

		return nil
	},
}

func init() {
	behaviorCmd.AddCommand(behaviorNewCmd)

	behaviorNewCmd.Flags().
		String("name", "", "The behavior's name: lowercase letters and digits (required)")
	behaviorNewCmd.Flags().
		String("description", "", "What it manages, in one line (required)")
	behaviorNewCmd.Flags().
		String("group", "", "Nest it under a group, such as network: its routes, commands and directory")
	behaviorNewCmd.Flags().
		String("root", ".", "The osapi checkout to write into")
	_ = behaviorNewCmd.MarkFlagRequired("name")
	_ = behaviorNewCmd.MarkFlagRequired("description")
}

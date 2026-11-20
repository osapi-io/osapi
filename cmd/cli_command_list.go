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
	"github.com/osapi-io/osapi/internal/cli/command"
)

// commandSummary is one command in `cli command list --json`.
type commandSummary struct {
	Command string `json:"command"`
	Layer   string `json:"layer"`
	Group   string `json:"group,omitempty"`
	Doc     string `json:"doc"`
}

// The layers that make a command.
const (
	layerCommand     = "command"
	layerEndpoint    = "endpoint"
	layerBehavior    = "behavior"
	layerHandWritten = "hand-written"
)

var cliCommandListCmd = &cobra.Command{
	Use:   "list",
	Short: "List each command of this binary, and the layer that made it",
	Long: "The layer is `command` for a command of a local group, `endpoint` for a\n" +
		"command of an endpoint group, `behavior` for a command of a behavior, and\n" +
		"`hand-written` for a command in `cmd/` that no definition makes yet.",
	Example: `  osapi cli command list
  osapi cli command list --json`,
	Args: cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		var summaries []commandSummary
		var walk func(c *cobra.Command)
		walk = func(c *cobra.Command) {
			if c.Hidden || c.Name() == "help" || c.Name() == "completion" {
				return
			}
			if c.Runnable() && len(c.Commands()) == 0 {
				layer, group := layerOf(c)
				summaries = append(summaries, commandSummary{
					Command: c.CommandPath(), Layer: layer, Group: group, Doc: c.Short,
				})
			}
			for _, sub := range c.Commands() {
				walk(sub)
			}
		}
		walk(rootCmd)

		if jsonOutput {
			out, _ := json.Marshal(summaries)
			cli.PrintJSON(out)

			return
		}

		count := map[string]int{}
		rows := make([][]string, 0, len(summaries))
		for _, s := range summaries {
			count[s.Layer]++
			rows = append(rows, []string{s.Command, s.Layer, s.Doc})
		}
		cli.PrintCompactTable([]cli.Section{
			{
				Title: fmt.Sprintf(
					"Commands (%d: %d command, %d endpoint, %d behavior, %d hand-written)",
					len(summaries),
					count[layerCommand],
					count[layerEndpoint],
					count[layerBehavior],
					count[layerHandWritten],
				),
				Headers:   []string{"COMMAND", "LAYER", "DESCRIPTION"},
				Rows:      rows,
				Reference: true,
			},
		})
	},
}

// layerOf is the layer that made c, and the name of its group.
func layerOf(
	c *cobra.Command,
) (string, string) {
	g, ok := command.Origin(c)
	if !ok {
		return layerHandWritten, ""
	}
	group := strings.Join(g.CommandPath(), " ")
	if lg, ok := g.(*command.LocalGroup); ok {
		return layerCommand, lg.Path()
	}
	// A behavior's group is an endpoint group that the behavior DSL makes,
	// with the behavior's name.
	if behavior.Lookup(g.GroupName()) != nil {
		return layerBehavior, group
	}

	return layerEndpoint, group
}

func init() {
	cliCommandCmd.AddCommand(cliCommandListCmd)
}

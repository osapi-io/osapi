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
	"github.com/osapi-io/osapi/internal/cli/command"
)

// localFlag is one flag in `cli command get --json`.
type localFlag struct {
	Flag     string `json:"flag"`
	Short    string `json:"short,omitempty"`
	Type     string `json:"type"`
	Validate string `json:"validate"`
	Doc      string `json:"doc"`
}

// localCommand is one command in `cli command get --json`. A local group's
// command has its flags and the inputs it refuses.
type localCommand struct {
	Method  string      `json:"method"`
	Command string      `json:"command"`
	Short   string      `json:"short"`
	Flags   []localFlag `json:"flags,omitempty"`
	Fails   []string    `json:"fails,omitempty"`
}

// findGroup is the registered group with the command path, a local group
// or the group of an endpoint or a behavior, or nil.
func findGroup(
	path string,
) command.Group {
	if g := command.Lookup(path); g != nil {
		return g
	}
	for _, g := range command.Groups() {
		if strings.Join(g.CommandPath(), " ") == path {
			return g
		}
	}

	return nil
}

// describeCommand is c of g, for `cli command get`: its command, its
// output, and for a local method its flags and the inputs it refuses.
func describeCommand(
	g command.Group,
	c command.Command,
) localCommand {
	root := "osapi client"
	if _, ok := g.(*command.LocalGroup); ok {
		root = "osapi"
	}
	full := strings.Join(
		[]string{root, strings.Join(g.CommandPath(), " "), c.SubPath(), c.CommandName()},
		" ",
	)
	lc := localCommand{
		Method:  c.MethodName(),
		Command: strings.Join(strings.Fields(full), " "),
		Short:   c.CommandShort(),
	}
	m, ok := c.(*command.LocalMethod)
	if !ok {
		return lc
	}
	for f := range m.In.Fields() {
		if !f.IsExported() {
			continue
		}
		lc.Flags = append(lc.Flags, localFlag{
			Flag: "--" + command.FlagName(f), Short: f.Tag.Get("short"),
			Type: f.Type.String(), Validate: f.Tag.Get("validate"), Doc: f.Tag.Get("doc"),
		})
	}
	for _, r := range m.Refusals() {
		lc.Fails = append(lc.Fails, r.Doc)
	}

	return lc
}

var cliCommandGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Describe one group: its commands, and the flags and refusals of a local group",
	Long: "`--name` is the command path of the group, such as `token` or `agent`: a\n" +
		"local group, or the group of an endpoint or a behavior. Each command prints\n" +
		"from the output type of its method. A command of a local group also has\n" +
		"its flags, each with the validate\n" +
		"tag of its field, and the inputs that its `.Fails` declares, which\n" +
		"`TestRefusals` runs.",
	Example: `  osapi cli command get --name token
  osapi cli command get --name agent --json`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		name, _ := cmd.Flags().GetString("name")

		g := findGroup(name)
		if g == nil {
			cli.LogFatal(logger, "no group", fmt.Errorf("no group named %q", name))
		}

		path := strings.Join(g.CommandPath(), " ")
		if lg, ok := g.(*command.LocalGroup); ok {
			path = lg.Path()
		}
		commands := make([]localCommand, 0, len(g.Commands()))
		for _, c := range g.Commands() {
			commands = append(commands, describeCommand(g, c))
		}

		if jsonOutput {
			out, _ := json.Marshal(map[string]any{
				"name": g.GroupName(), "path": path, "doc": g.GroupDoc(), "commands": commands,
			})
			cli.PrintJSON(out)

			return
		}

		cli.PrintHeading("Group")
		cli.PrintKV("Name", path, "Description", g.GroupDoc())
		rows := make([][]string, 0, len(commands))
		for _, c := range commands {
			rows = append(rows, []string{c.Command, c.Short})
		}
		cli.PrintCompactTable([]cli.Section{{
			Title:     fmt.Sprintf("Commands (%d)", len(commands)),
			Headers:   []string{"COMMAND", "DESCRIPTION"},
			Rows:      rows,
			Reference: true,
		}})
		for _, c := range commands {
			if len(c.Flags) == 0 && len(c.Fails) == 0 {
				continue
			}
			rows := make([][]string, 0, len(c.Flags))
			for _, f := range c.Flags {
				rows = append(rows, []string{f.Flag, f.Type, f.Validate, f.Doc})
			}
			cli.PrintCompactTable([]cli.Section{{
				Title:     fmt.Sprintf("%s: %s", c.Command, c.Short),
				Headers:   []string{"FLAG", "TYPE", "VALIDATE", "HELP"},
				Rows:      rows,
				Reference: true,
			}})
			if len(c.Fails) > 0 {
				cli.PrintKV("Refuses", strings.Join(c.Fails, "; "))
			}
		}
	},
}

func init() {
	cliCommandCmd.AddCommand(cliCommandGetCmd)
	cliCommandGetCmd.Flags().
		String("name", "", "The command path of the group, such as token or agent")
	_ = cliCommandGetCmd.MarkFlagRequired("name")
}

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

// localCommand is one command in `cli command get --json`.
type localCommand struct {
	Method  string      `json:"method"`
	Command string      `json:"command"`
	Short   string      `json:"short"`
	Flags   []localFlag `json:"flags"`
	Fails   []string    `json:"fails"`
}

var cliCommandGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Describe one local group: its commands, their flags and the inputs they refuse",
	Long: "`--name` is the command path of the group, such as `token`. Each flag has\n" +
		"the validate tag of its field, and each command has the inputs that its\n" +
		"`.Fails` declares, which `TestRefusals` runs.",
	Example: `  osapi cli command get --name token
  osapi cli command get --name token --json`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		name, _ := cmd.Flags().GetString("name")

		g := command.Lookup(name)
		if g == nil {
			cli.LogFatal(logger, "no local group", fmt.Errorf("no group named %q", name))
		}

		path := strings.Join(g.CommandPath(), " ")
		commands := make([]localCommand, 0, len(g.Methods))
		for _, m := range g.Methods {
			full := strings.Join([]string{"osapi", path, m.SubPath(), m.Name}, " ")
			lc := localCommand{
				Method:  m.Name,
				Command: strings.Join(strings.Fields(full), " "),
				Short:   m.CommandShort(),
				Flags:   []localFlag{},
				Fails:   []string{},
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
			commands = append(commands, lc)
		}

		if jsonOutput {
			out, _ := json.Marshal(map[string]any{
				"name": g.Name, "path": path, "doc": g.Doc, "commands": commands,
			})
			cli.PrintJSON(out)

			return
		}

		cli.PrintHeading("Local group")
		cli.PrintKV("Name", path, "Description", g.Doc)
		for _, c := range commands {
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
		String("name", "", "The command path of the local group, such as token")
	_ = cliCommandGetCmd.MarkFlagRequired("name")
}

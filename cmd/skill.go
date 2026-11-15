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
	"io/fs"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/cmd/resources"
	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/skills"
)

// skillSources returns the skills that osapi init installs, and the skills
// for a person who changes osapi itself.
func skillSources() (fs.FS, fs.FS) {
	installed, _ := fs.Sub(resources.Skills, "skills")
	dev, _ := fs.Sub(resources.DevSkills, "dev-skills")

	return installed, dev
}

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Show the agent skills that this binary has",
	Long: `Shows the skills compiled into this binary. osapi init installs the skills
that operate hosts. The skills that change osapi itself, such as
build-a-behavior, are only served here: a checkout of osapi points to them.`,
}

var skillListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the skills that this binary has",
	Example: `  osapi skill list
  osapi skill list --json`,
	Args: cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		installed, dev := skillSources()
		all, err := skills.Catalog(installed, dev)
		if err != nil {
			return err
		}

		if jsonOutput {
			out, _ := json.Marshal(all)
			cli.PrintJSON(out)

			return nil
		}

		rows := make([][]string, 0, len(all))
		for _, s := range all {
			init := "no"
			if s.Init {
				init = "yes"
			}
			rows = append(rows, []string{s.Name, init, fmt.Sprint(len(s.References))})
		}
		cli.PrintCompactTable([]cli.Section{{
			Title:     fmt.Sprintf("Skills (%d)", len(all)),
			Headers:   []string{"NAME", "INSTALLED BY INIT", "REFERENCES"},
			Rows:      rows,
			Reference: true,
		}})

		return nil
	},
}

// skillFile is what `osapi skill get --json` prints.
type skillFile struct {
	Name      string `json:"name"`
	Reference string `json:"reference,omitempty"`
	Content   string `json:"content"`
}

var skillGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Print a skill, or one of its reference files",
	Long: `Prints the SKILL.md of a skill as markdown. With --reference, prints one of
its reference files, for example references/services.md. An agent reads the
output and obeys it.`,
	Example: `  osapi skill get --name build-a-behavior
  osapi skill get --name build-a-behavior --reference references/services.md`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")
		reference, _ := cmd.Flags().GetString("reference")

		installed, dev := skillSources()
		data, err := skills.Read(installed, dev, name, reference)
		if err != nil {
			return fmt.Errorf("%w: run osapi skill list", err)
		}

		if jsonOutput {
			out, _ := json.Marshal(skillFile{Name: name, Reference: reference, Content: string(data)})
			cli.PrintJSON(out)

			return nil
		}

		fmt.Fprint(os.Stdout, strings.TrimRight(string(data), "\n")+"\n")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(skillCmd)
	skipConfigCmds["skill"] = true
	skillCmd.AddCommand(skillListCmd, skillGetCmd)
	skillGetCmd.Flags().String("name", "", "Name of the skill")
	_ = skillGetCmd.MarkFlagRequired("name")
	skillGetCmd.Flags().String("reference", "", "A reference file of the skill, relative to its directory")
}

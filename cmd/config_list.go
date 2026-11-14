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
)

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List every setting, its value in effect and where it came from",
	Long: `Lists every key osapi reads with the value in effect and its source:
env, file, default or unset. The file and the environment are read the way
osapi start reads them; without a config file the values are the defaults and
the environment. Passwords, signing keys and tokens print as ********.

--json carries each key's type, default, rules and description too. osapi
config get --key <key> shows one key in full.`,
	Example: `  osapi config list
  osapi config list -f ./osapi.yaml --json`,
	Args: cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		file, settings, err := readSettings()
		if err != nil {
			cli.HandleError(err, logger)

			return
		}

		if jsonOutput {
			out, _ := json.Marshal(settings)
			cli.PrintJSON(out)

			return
		}

		cli.PrintHeading("Config")
		cli.PrintKV(
			"File", file,
			"Environment", "OSAPI_, then the key in capitals with dots as underscores",
			"One key", "osapi config get --key <key>",
		)

		// One table per top-level section, so the list reads like the file
		// it describes.
		var sections []cli.Section
		index := map[string]int{}
		for _, st := range settings {
			group, _, found := strings.Cut(st.Name, ".")
			if !found {
				group = "general"
			}
			i, ok := index[group]
			if !ok {
				i = len(sections)
				index[group] = i
				sections = append(sections, cli.Section{
					Title:     group,
					Headers:   []string{"KEY", "VALUE", "SOURCE", "ENV"},
					Reference: true,
				})
			}
			sections[i].Rows = append(
				sections[i].Rows,
				[]string{st.Name, display(st.Value), st.Source, st.Env},
			)
		}
		for i := range sections {
			sections[i].Title = fmt.Sprintf(
				"%s (%d)", titleCase(sections[i].Title), len(sections[i].Rows),
			)
		}
		cli.PrintCompactTable(sections)
	},
}

// titleCase capitalizes a section name for its heading: nats becomes NATS,
// the rest take a capital first letter.
func titleCase(
	s string,
) string {
	if s == "nats" {
		return "NATS"
	}

	return strings.ToUpper(s[:1]) + s[1:]
}

func init() {
	configCmd.AddCommand(configListCmd)
}

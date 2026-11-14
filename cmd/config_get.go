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

	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/cli"
)

var configGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Show one setting in full",
	Long: `Shows one setting: its value in effect and where it came from, its
environment variable, type, default, rules and description. The file and the
environment are read the way osapi start reads them. A password, signing key
or token prints as ********. osapi config list shows every setting.`,
	Example: `  osapi config get --key controller.api.port
  osapi config get --key agent.labels -f ./osapi.yaml --json`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		key, _ := cmd.Flags().GetString("key")
		file, settings, err := readSettings()
		if err != nil {
			cli.HandleError(err, logger)

			return
		}

		var st *setting
		for i := range settings {
			if settings[i].Name == key {
				st = &settings[i]
			}
		}
		if st == nil {
			cli.HandleError(
				fmt.Errorf("no setting %q: osapi config list shows them all", key),
				logger,
			)

			return
		}

		if jsonOutput {
			out, _ := json.Marshal(st)
			cli.PrintJSON(out)

			return
		}

		cli.PrintHeading("Config")
		cli.PrintKV("File", file)
		cli.PrintHeading("Setting")
		cli.PrintKV(
			"Key", st.Name,
			"Value", display(st.Value),
			"Source", st.Source,
			"Env", st.Env,
		)
		if st.Flag != "" {
			cli.PrintKV("Flag", "osapi agent start --"+st.Flag)
		}
		cli.PrintKV(
			"Type", st.Type,
			"Default", display(st.Default),
			"Validate", st.Validate,
			"Description", st.Doc,
		)
	},
}

func init() {
	configCmd.AddCommand(configGetCmd)
	configGetCmd.Flags().String("key", "", "The setting to show, such as controller.api.port")
	_ = configGetCmd.MarkFlagRequired("key")
}

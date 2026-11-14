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
	"github.com/spf13/viper"

	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/config"
)

// configCheck is what `config validate --json` prints.
type configCheck struct {
	File  string `json:"file"`
	Valid bool   `json:"valid"`
	Error string `json:"error,omitempty"`
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Check a config file before osapi starts with it",
	Long: `Loads the config file and the environment the way osapi start does
and runs the same checks, without starting anything. Exits 1 when the config
is not valid.`,
	Example: `  osapi config validate
  osapi config validate -f ./osapi.yaml --json`,
	Args: cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		file := viper.GetString("osapiFile")
		var c config.Config
		err := config.Load(viper.New(), file, &c)
		result := configCheck{File: file, Valid: err == nil}
		if err != nil {
			result.Error = err.Error()
		}

		if jsonOutput {
			out, _ := json.Marshal(result)
			cli.PrintJSON(out)
		} else {
			cli.PrintHeading("Config")
			cli.PrintKV("File", file)
			if result.Valid {
				cli.PrintKV("Status", "valid")
			} else {
				cli.PrintKV("Status", "invalid", "Error", result.Error)
			}
		}

		if !result.Valid {
			cli.FlushKV()
			cliExit(cli.ExitError)
		}
	},
}

func init() {
	configCmd.AddCommand(configValidateCmd)
}

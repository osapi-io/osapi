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

	"github.com/osapi-io/osapi/internal/cli"
)

// clientFileDeleteCmd represents the clientFileDelete command.
var clientFileDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a file from the Object Store",
	Example: `  osapi client file delete --name app.conf
  osapi client file delete --name app.conf --json`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		name, _ := cmd.Flags().GetString("name")

		if err := sdkClient.File.Delete(ctx, name); err != nil {
			cli.HandleError(err, logger)
			return
		}

		// A delete answers with no body, so --json prints what it did, as
		// job delete does.
		if jsonOutput {
			result, _ := json.Marshal(map[string]any{"name": name, "status": "deleted"})
			cli.PrintJSON(result)
			return
		}

		cli.PrintHeading("File")
		cli.PrintKV("Name", name, "Status", "Deleted")
	},
}

func init() {
	clientFileCmd.AddCommand(clientFileDeleteCmd)

	clientFileDeleteCmd.PersistentFlags().
		String("name", "", "Name of the file in the Object Store")

	_ = clientFileDeleteCmd.MarkPersistentFlagRequired("name")
}

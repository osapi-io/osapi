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
	"fmt"

	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
)

// clientFileDeleteCmd represents the clientFileDelete command.
var clientFileDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a file from the Object Store",
	Long:  `Delete a specific file from the OSAPI Object Store.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		name, _ := cmd.Flags().GetString("name")

		resp, err := sdkClient.File.Delete(ctx, name)
		if err != nil {
			cli.HandleError(err, logger)
			return
		}

		if jsonOutput {
			fmt.Println(string(resp.RawJSON()))
			return
		}

		fmt.Println()
		cli.PrintKV("Name", resp.Data.Name)
		cli.PrintKV("Deleted", fmt.Sprintf("%v", resp.Data.Deleted))
	},
}

func init() {
	clientFileCmd.AddCommand(clientFileDeleteCmd)

	clientFileDeleteCmd.PersistentFlags().
		String("name", "", "Name of the file in the Object Store (required)")

	_ = clientFileDeleteCmd.MarkPersistentFlagRequired("name")
}

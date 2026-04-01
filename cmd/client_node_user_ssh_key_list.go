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

// clientNodeUserSSHKeyListCmd represents the user ssh-key list command.
var clientNodeUserSSHKeyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List SSH authorized keys",
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		name, _ := cmd.Flags().GetString("name")

		resp, err := sdkClient.User.ListKeys(ctx, host, name)
		if err != nil {
			cli.HandleError(err, logger)
			return
		}

		if jsonOutput {
			fmt.Println(string(resp.RawJSON()))
			return
		}

		if resp.Data.JobID != "" {
			fmt.Println()
			cli.PrintKV("Job ID", resp.Data.JobID)
		}

		results := make([]cli.ResultRow, 0)
		for _, r := range resp.Data.Results {
			var errPtr *string
			if r.Error != "" {
				errPtr = &r.Error
			}
			if errPtr != nil || len(r.Keys) == 0 {
				results = append(results, cli.ResultRow{
					Hostname: r.Hostname,
					Status:   r.Status,
					Error:    errPtr,
					Fields:   []string{"", "", ""},
				})

				continue
			}
			for _, k := range r.Keys {
				results = append(results, cli.ResultRow{
					Hostname: r.Hostname,
					Status:   r.Status,
					Fields: []string{
						k.Type,
						k.Fingerprint,
						k.Comment,
					},
				})
			}
		}
		headers, rows := cli.BuildBroadcastTable(results, []string{
			"TYPE", "FINGERPRINT", "COMMENT",
		})
		cli.PrintCompactTable([]cli.Section{{Headers: headers, Rows: rows}})
	},
}

func init() {
	clientNodeUserSSHKeyCmd.AddCommand(clientNodeUserSSHKeyListCmd)

	clientNodeUserSSHKeyListCmd.PersistentFlags().
		String("name", "", "Username to list SSH keys for (required)")

	_ = clientNodeUserSSHKeyListCmd.MarkPersistentFlagRequired("name")
}

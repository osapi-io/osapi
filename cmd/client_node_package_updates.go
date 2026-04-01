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

// clientNodePackageUpdatesCmd represents the package updates command.
var clientNodePackageUpdatesCmd = &cobra.Command{
	Use:   "updates",
	Short: "List available package updates",
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")

		resp, err := sdkClient.Package.ListUpdates(ctx, host)
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

		results := make([]cli.ResultRow, 0, len(resp.Data.Results))
		for _, r := range resp.Data.Results {
			var errPtr *string
			if r.Error != "" {
				errPtr = &r.Error
			}
			for _, u := range r.Updates {
				results = append(results, cli.ResultRow{
					Hostname: r.Hostname,
					Status:   r.Status,
					Error:    errPtr,
					Fields:   []string{u.Name, u.CurrentVersion, u.NewVersion},
				})
			}
			if len(r.Updates) == 0 && r.Error != "" {
				results = append(results, cli.ResultRow{
					Hostname: r.Hostname,
					Status:   r.Status,
					Error:    errPtr,
				})
			}
		}
		headers, rows := cli.BuildBroadcastTable(results, []string{
			"NAME", "CURRENT", "NEW",
		})
		cli.PrintCompactTable([]cli.Section{{Headers: headers, Rows: rows}})
	},
}

func init() {
	clientNodePackageCmd.AddCommand(clientNodePackageUpdatesCmd)
}

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
	"github.com/retr0h/osapi/pkg/sdk/client"
)

// clientNodeLogQueryCmd represents the log query command.
var clientNodeLogQueryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query journal entries",
	Long:  `Query journal log entries on the target node.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")

		opts := client.LogQueryOpts{}

		if cmd.Flags().Changed("lines") {
			lines, _ := cmd.Flags().GetInt("lines")
			opts.Lines = &lines
		}

		since, _ := cmd.Flags().GetString("since")
		if since != "" {
			opts.Since = &since
		}

		priority, _ := cmd.Flags().GetString("priority")
		if priority != "" {
			opts.Priority = &priority
		}

		resp, err := sdkClient.Log.Query(ctx, host, opts)
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
			fmt.Println()
		}

		results := make([]cli.ResultRow, 0)
		for _, r := range resp.Data.Results {
			if r.Error != "" {
				var errPtr *string
				e := r.Error
				errPtr = &e
				results = append(results, cli.ResultRow{
					Hostname: r.Hostname,
					Status:   r.Status,
					Error:    errPtr,
				})

				continue
			}

			for _, e := range r.Entries {
				message := e.Message
				if len(message) > 80 {
					message = message[:77] + "..."
				}

				results = append(results, cli.ResultRow{
					Hostname: r.Hostname,
					Status:   r.Status,
					Fields: []string{
						e.Timestamp,
						e.Unit,
						message,
					},
				})
			}
		}
		tr := cli.BuildBroadcastTable(
			results,
			[]string{"TIMESTAMP", "UNIT", "MESSAGE"},
		)
		cli.PrintCompactTable(
			[]cli.Section{{Headers: tr.Headers, Rows: tr.Rows, Errors: tr.Errors}},
		)
	},
}

func init() {
	clientNodeLogCmd.AddCommand(clientNodeLogQueryCmd)

	clientNodeLogQueryCmd.PersistentFlags().
		Int("lines", 100, "Maximum number of log lines to return")
	clientNodeLogQueryCmd.PersistentFlags().
		String("since", "", "Return entries since this time (e.g., '1h', '2026-01-01 00:00:00')")
	clientNodeLogQueryCmd.PersistentFlags().
		String("priority", "", "Filter by priority level (e.g., 'err', 'warning', 'info')")
}

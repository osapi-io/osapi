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

// clientNodeProcessGetCmd represents the process get command.
var clientNodeProcessGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get process information by PID",
	Long:  `Get detailed information about a specific process on the target node.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		pid, _ := cmd.Flags().GetInt("pid")

		resp, err := sdkClient.Process.Get(ctx, host, pid)
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

			for _, p := range r.Processes {
				command := p.Command
				if len(command) > 60 {
					command = command[:57] + "..."
				}

				results = append(results, cli.ResultRow{
					Hostname: r.Hostname,
					Status:   r.Status,
					Fields: []string{
						fmt.Sprintf("%d", p.PID),
						p.Name,
						p.User,
						p.State,
						fmt.Sprintf("%.1f%%", p.CPUPercent),
						fmt.Sprintf("%.1f%%", p.MemPercent),
						command,
					},
				})
			}
		}
		headers, rows := cli.BuildBroadcastTable(
			results,
			[]string{"PID", "NAME", "USER", "STATE", "CPU%", "MEM%", "COMMAND"},
		)
		cli.PrintCompactTable([]cli.Section{{Headers: headers, Rows: rows}})
	},
}

func init() {
	clientNodeProcessCmd.AddCommand(clientNodeProcessGetCmd)

	clientNodeProcessGetCmd.PersistentFlags().
		Int("pid", 0, "Process ID to inspect (required)")

	_ = clientNodeProcessGetCmd.MarkPersistentFlagRequired("pid")
}

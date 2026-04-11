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

// clientNodeDiskGetCmd represents the clientNodeDiskGet command.
var clientNodeDiskGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get disk usage",
	Long:  `Get disk usage information from the target node.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		resp, err := sdkClient.Disk.Get(ctx, host)
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
				e := r.Error
				results = append(results, cli.ResultRow{
					Hostname: r.Hostname,
					Status:   r.Status,
					Error:    &e,
				})

				continue
			}

			for _, disk := range r.Disks {
				usage := ""
				if disk.Total > 0 {
					pct := float64(disk.Used) / float64(disk.Total) * 100
					usage = fmt.Sprintf("%.0f%%", pct)
				}

				results = append(results, cli.ResultRow{
					Hostname: r.Hostname,
					Status:   r.Status,
					Fields: []string{
						disk.Name,
						cli.FormatBytes(disk.Total),
						cli.FormatBytes(disk.Used),
						usage,
					},
				})
			}
		}
		tr := cli.BuildBroadcastTable(
			results,
			[]string{"MOUNT", "TOTAL", "USED", "USAGE"},
		)
		cli.PrintCompactTable(
			[]cli.Section{{Headers: tr.Headers, Rows: tr.Rows, Errors: tr.Errors}},
		)
	},
}

func init() {
	clientNodeDiskCmd.AddCommand(clientNodeDiskGetCmd)
}

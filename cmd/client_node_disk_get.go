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

		for _, r := range resp.Data.Results {
			if r.Error != "" {
				fmt.Println()
				cli.PrintKV("Hostname", r.Hostname, "Error", r.Error)
				continue
			}

			fmt.Println()
			cli.PrintKV("Hostname", r.Hostname)

			diskRows := make([][]string, 0, len(r.Disks))
			for _, disk := range r.Disks {
				usage := ""
				if disk.Total > 0 {
					pct := float64(disk.Used) / float64(disk.Total) * 100
					usage = fmt.Sprintf("%.0f%%", pct)
				}

				diskRows = append(diskRows, []string{
					disk.Name,
					cli.FormatBytes(disk.Total),
					cli.FormatBytes(disk.Used),
					cli.FormatBytes(disk.Free),
					usage,
				})
			}

			sections := []cli.Section{
				{
					Headers: []string{"MOUNT", "TOTAL", "USED", "FREE", "USAGE"},
					Rows:    diskRows,
				},
			}
			cli.PrintCompactTable(sections)
		}
	},
}

func init() {
	clientNodeDiskCmd.AddCommand(clientNodeDiskGetCmd)
}

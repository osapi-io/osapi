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
	"strconv"

	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
)

var (
	auditListLimit  int
	auditListOffset int
)

// clientAuditListCmd represents the clientAuditList command.
var clientAuditListCmd = &cobra.Command{
	Use:   "list",
	Short: "List audit log entries",
	Long: `List audit log entries with pagination.

Displays a table of recent API activity including user, method, path,
response status, and duration. Requires audit:read permission.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		resp, err := sdkClient.Audit.List(ctx, auditListLimit, auditListOffset)
		if err != nil {
			cli.HandleError(err, logger)
			return
		}

		if jsonOutput {
			fmt.Println(string(resp.RawJSON()))
			return
		}

		fmt.Println()
		cli.PrintKV("Total", strconv.Itoa(resp.Data.TotalItems))

		if len(resp.Data.Items) == 0 {
			fmt.Println("  No audit entries found.")
			return
		}

		rows := make([][]string, 0, len(resp.Data.Items))
		for _, entry := range resp.Data.Items {
			rows = append(rows, []string{
				entry.ID,
				entry.Timestamp.Format("2006-01-02 15:04:05"),
				entry.User,
				entry.Method,
				entry.Path,
				strconv.Itoa(entry.ResponseCode),
				strconv.FormatInt(entry.DurationMs, 10) + "ms",
			})
		}

		cli.PrintCompactTable([]cli.Section{
			{
				Title: "Audit Entries",
				Headers: []string{
					"ID",
					"TIMESTAMP",
					"USER",
					"METHOD",
					"PATH",
					"STATUS",
					"DURATION",
				},
				Rows: rows,
			},
		})
	},
}

func init() {
	clientAuditCmd.AddCommand(clientAuditListCmd)
	clientAuditListCmd.Flags().
		IntVar(&auditListLimit, "limit", 20, "Maximum number of entries to return")
	clientAuditListCmd.Flags().IntVar(&auditListOffset, "offset", 0, "Number of entries to skip")
}

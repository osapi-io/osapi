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
	"net/http"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
	"github.com/retr0h/osapi/internal/client"
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
		auditHandler := handler.(client.AuditHandler)
		resp, err := auditHandler.GetAuditLogs(ctx, auditListLimit, auditListOffset)
		if err != nil {
			cli.LogFatal(logger, "failed to get audit logs", err)
		}

		switch resp.StatusCode() {
		case http.StatusOK:
			if jsonOutput {
				fmt.Println(string(resp.Body))
				return
			}

			if resp.JSON200 == nil {
				cli.LogFatal(logger, "failed response", fmt.Errorf("audit list response was nil"))
			}

			fmt.Println()
			cli.PrintKV("Total", strconv.Itoa(resp.JSON200.TotalItems))

			if len(resp.JSON200.Items) == 0 {
				fmt.Println("  No audit entries found.")
				return
			}

			rows := make([][]string, 0, len(resp.JSON200.Items))
			for _, entry := range resp.JSON200.Items {
				rows = append(rows, []string{
					entry.Id.String(),
					entry.Timestamp.Format("2006-01-02 15:04:05"),
					entry.User,
					entry.Method,
					entry.Path,
					strconv.Itoa(entry.ResponseCode),
					strconv.FormatInt(entry.DurationMs, 10) + "ms",
				})
			}

			cli.PrintStyledTable([]cli.Section{
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

		case http.StatusBadRequest:
			cli.HandleUnknownError(resp.JSON400, resp.StatusCode(), logger)
		case http.StatusUnauthorized:
			cli.HandleAuthError(resp.JSON401, resp.StatusCode(), logger)
		case http.StatusForbidden:
			cli.HandleAuthError(resp.JSON403, resp.StatusCode(), logger)
		default:
			cli.HandleUnknownError(resp.JSON500, resp.StatusCode(), logger)
		}
	},
}

func init() {
	clientAuditCmd.AddCommand(clientAuditListCmd)
	clientAuditListCmd.Flags().
		IntVar(&auditListLimit, "limit", 20, "Maximum number of entries to return")
	clientAuditListCmd.Flags().IntVar(&auditListOffset, "offset", 0, "Number of entries to skip")
}

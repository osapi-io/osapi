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
	"strings"

	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
)

// clientAuditGetCmd represents the clientAuditGet command.
var clientAuditGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a single audit log entry",
	Long: `Get a single audit log entry by its UUID.

Requires audit:read permission.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		auditID, _ := cmd.Flags().GetString("audit-id")

		resp, err := sdkClient.Audit.Get(ctx, auditID)
		if err != nil {
			cli.LogFatal(logger, "failed to get audit log entry", err)
		}

		switch resp.StatusCode() {
		case http.StatusOK:
			if jsonOutput {
				fmt.Println(string(resp.Body))
				return
			}

			if resp.JSON200 == nil {
				cli.LogFatal(logger, "failed response", fmt.Errorf("audit entry response was nil"))
			}

			entry := resp.JSON200.Entry

			fmt.Println()
			cli.PrintKV("ID", entry.Id.String())
			cli.PrintKV("Timestamp", entry.Timestamp.Format("2006-01-02 15:04:05"))
			cli.PrintKV("User", entry.User)
			cli.PrintKV("Roles", strings.Join(entry.Roles, ", "))
			cli.PrintKV("Method", entry.Method, "Path", entry.Path)
			cli.PrintKV(
				"Status",
				strconv.Itoa(entry.ResponseCode),
				"Duration",
				strconv.FormatInt(entry.DurationMs, 10)+"ms",
			)
			cli.PrintKV("Source IP", entry.SourceIp)
			if entry.OperationId != nil {
				cli.PrintKV("Operation", *entry.OperationId)
			}

		case http.StatusUnauthorized:
			cli.HandleAuthError(resp.JSON401, resp.StatusCode(), logger)
		case http.StatusForbidden:
			cli.HandleAuthError(resp.JSON403, resp.StatusCode(), logger)
		case http.StatusNotFound:
			cli.HandleUnknownError(resp.JSON404, resp.StatusCode(), logger)
		default:
			cli.HandleUnknownError(resp.JSON500, resp.StatusCode(), logger)
		}
	},
}

func init() {
	clientAuditCmd.AddCommand(clientAuditGetCmd)

	clientAuditGetCmd.PersistentFlags().
		StringP("audit-id", "", "", "Audit entry ID to retrieve")

	_ = clientAuditGetCmd.MarkPersistentFlagRequired("audit-id")
}

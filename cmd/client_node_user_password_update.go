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
	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/cli"
)

// clientNodeUserPasswordCmd holds the commands of a user's password, such
// as osapi client node user password update.
var clientNodeUserPasswordCmd = &cobra.Command{
	Use:   "password",
	Short: "Manage a user's password",
}

// clientNodeUserPasswordUpdateCmd represents the user password update command.
var clientNodeUserPasswordUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a user's password",
	Long: "The client sends the password to the controller in plaintext over TLS. The\n" +
		"controller hashes it before it stores the job. The agent and the provider\n" +
		"see only the hash. osapi never writes the plaintext to disk or to logs.",
	Example: `  osapi client node user password update --target web-01 \
    --name deploy --password 'newpass123'
  osapi client node user password update --target _all \
    --name deploy --password 'newpass123'`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		name, _ := cmd.Flags().GetString("name")
		password, _ := cmd.Flags().GetString("password")

		resp, err := sdkClient.User.ChangePassword(ctx, host, name, password)
		if err != nil {
			cli.HandleError(err, logger)
			return
		}

		if jsonOutput {
			cli.PrintJSON(resp.RawJSON())
			return
		}

		cli.PrintJobID(resp.Data.JobID)

		cli.PrintTraceID(resp.TraceID())

		results := make([]cli.ResultRow, 0, len(resp.Data.Results))
		for _, r := range resp.Data.Results {
			var errPtr *string
			if r.Error != "" {
				errPtr = &r.Error
			}
			changed := r.Changed
			results = append(results, cli.ResultRow{
				Hostname: r.Hostname,
				Status:   r.Status,
				Changed:  &changed,
				Error:    errPtr,
				Fields:   []string{r.Name},
			})
		}
		tr := cli.BuildMutationTable(results, []string{"NAME"})
		cli.PrintCompactTable(
			[]cli.Section{{Headers: tr.Headers, Rows: tr.Rows, Errors: tr.Errors}},
		)
	},
}

func init() {
	clientNodeUserCmd.AddCommand(clientNodeUserPasswordCmd)
	clientNodeUserPasswordCmd.AddCommand(clientNodeUserPasswordUpdateCmd)

	clientNodeUserPasswordUpdateCmd.PersistentFlags().
		String("name", "", "Username to change password for")
	clientNodeUserPasswordUpdateCmd.PersistentFlags().
		String("password", "", "New password (hashed by the controller before storage)")

	_ = clientNodeUserPasswordUpdateCmd.MarkPersistentFlagRequired("name")
	_ = clientNodeUserPasswordUpdateCmd.MarkPersistentFlagRequired("password")
}

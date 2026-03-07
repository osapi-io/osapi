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

// clientNodeFileStatusCmd represents the clientNodeFileStatus command.
var clientNodeFileStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check deployment status of a file on a host",
	Long: `Check the deployment status of a file on the target host.
Reports whether the file is in-sync, drifted, or missing.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		path, _ := cmd.Flags().GetString("path")

		resp, err := sdkClient.Node.FileStatus(ctx, host, path)
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

		sha := ""
		if resp.Data.SHA256 != "" {
			sha = resp.Data.SHA256
		}
		results := []cli.ResultRow{
			{
				Hostname: resp.Data.Hostname,
				Fields:   []string{resp.Data.Path, resp.Data.Status, sha},
			},
		}
		headers, rows := cli.BuildBroadcastTable(results, []string{"PATH", "STATUS", "SHA256"})
		cli.PrintCompactTable([]cli.Section{{Headers: headers, Rows: rows}})
	},
}

func init() {
	clientNodeFileCmd.AddCommand(clientNodeFileStatusCmd)

	clientNodeFileStatusCmd.PersistentFlags().
		String("path", "", "Filesystem path to check (required)")

	_ = clientNodeFileStatusCmd.MarkPersistentFlagRequired("path")
}

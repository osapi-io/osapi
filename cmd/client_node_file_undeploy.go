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

	"github.com/retr0h/osapi/pkg/sdk/client"
	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
)

// clientNodeFileUndeployCmd represents the clientNodeFileUndeploy command.
var clientNodeFileUndeployCmd = &cobra.Command{
	Use:   "undeploy",
	Short: "Remove a deployed file from a host",
	Long: `Remove a previously deployed file from the target host's filesystem.
The object store entry is preserved; only the file on disk is removed.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		path, _ := cmd.Flags().GetString("path")

		if host == "_all" {
			fmt.Print("This will undeploy the file from ALL hosts. Continue? [y/N] ")
			var confirm string
			if _, err := fmt.Scanln(&confirm); err != nil || (confirm != "y" && confirm != "Y") {
				fmt.Println("Aborted.")
				return
			}
		}

		resp, err := sdkClient.Node.FileUndeploy(ctx, client.FileUndeployOpts{
			Target: host,
			Path:   path,
		})
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

		changed := resp.Data.Changed
		changedPtr := &changed
		results := []cli.MutationResultRow{
			{
				Hostname: resp.Data.Hostname,
				Changed:  changedPtr,
			},
		}
		headers, rows := cli.BuildMutationTable(results, nil)
		cli.PrintCompactTable([]cli.Section{{Headers: headers, Rows: rows}})
	},
}

func init() {
	clientNodeFileCmd.AddCommand(clientNodeFileUndeployCmd)

	clientNodeFileUndeployCmd.PersistentFlags().
		String("path", "", "Path of the file to undeploy on the target filesystem (required)")

	_ = clientNodeFileUndeployCmd.MarkPersistentFlagRequired("path")
}

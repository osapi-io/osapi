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

// clientNodeGroupUpdateCmd represents the group update command.
var clientNodeGroupUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a group",
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		name, _ := cmd.Flags().GetString("name")
		members, _ := cmd.Flags().GetStringSlice("members")

		opts := client.GroupUpdateOpts{}
		if cmd.Flags().Changed("members") {
			opts.Members = members
		}

		resp, err := sdkClient.Group.Update(ctx, host, name, opts)
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
		cli.PrintCompactTable([]cli.Section{{Headers: tr.Headers, Rows: tr.Rows, Errors: tr.Errors}})
	},
}

func init() {
	clientNodeGroupCmd.AddCommand(clientNodeGroupUpdateCmd)

	clientNodeGroupUpdateCmd.PersistentFlags().
		String("name", "", "Group name to update (required)")
	clientNodeGroupUpdateCmd.PersistentFlags().
		StringSlice("members", nil, "Group members (replaces existing, required)")

	_ = clientNodeGroupUpdateCmd.MarkPersistentFlagRequired("name")
	_ = clientNodeGroupUpdateCmd.MarkPersistentFlagRequired("members")
}

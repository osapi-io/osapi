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

// clientNodeScheduleCronUpdateCmd represents the cron update command.
var clientNodeScheduleCronUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a cron entry",
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		name, _ := cmd.Flags().GetString("name")
		object, _ := cmd.Flags().GetString("object")
		schedule, _ := cmd.Flags().GetString("schedule")
		user, _ := cmd.Flags().GetString("user")
		contentType, _ := cmd.Flags().GetString("content-type")

		resp, err := sdkClient.Schedule.CronUpdate(ctx, host, name, client.CronUpdateOpts{
			Object:      object,
			Schedule:    schedule,
			User:        user,
			ContentType: contentType,
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

		if len(resp.Data.Results) == 0 {
			fmt.Println("\n  No results.")
			return
		}

		rows := make([][]string, 0, len(resp.Data.Results))
		for _, r := range resp.Data.Results {
			errVal := r.Error
			if errVal == "" {
				errVal = "-"
			}
			rows = append(rows, []string{
				r.Hostname,
				r.Name,
				fmt.Sprintf("%v", r.Changed),
				errVal,
			})
		}
		cli.PrintCompactTable([]cli.Section{{
			Headers: []string{"HOSTNAME", "NAME", "CHANGED", "ERROR"},
			Rows:    rows,
		}})
	},
}

func init() {
	clientNodeScheduleCronCmd.AddCommand(clientNodeScheduleCronUpdateCmd)

	clientNodeScheduleCronUpdateCmd.PersistentFlags().
		String("name", "", "Name of the cron entry to update (required)")
	clientNodeScheduleCronUpdateCmd.PersistentFlags().
		String("object", "", "New object to deploy")
	clientNodeScheduleCronUpdateCmd.PersistentFlags().
		String("schedule", "", "New cron schedule expression")
	clientNodeScheduleCronUpdateCmd.PersistentFlags().
		String("user", "", "New user to run the command as")
	clientNodeScheduleCronUpdateCmd.PersistentFlags().
		String("content-type", "", "Content type: raw or template")

	_ = clientNodeScheduleCronUpdateCmd.MarkPersistentFlagRequired("name")
	clientNodeScheduleCronUpdateCmd.MarkFlagsOneRequired("object", "schedule", "user")
}

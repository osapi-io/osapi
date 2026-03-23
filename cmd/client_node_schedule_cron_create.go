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

// clientNodeScheduleCronCreateCmd represents the cron create command.
var clientNodeScheduleCronCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a cron entry",
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		name, _ := cmd.Flags().GetString("name")
		schedule, _ := cmd.Flags().GetString("schedule")
		command, _ := cmd.Flags().GetString("command")
		user, _ := cmd.Flags().GetString("user")

		resp, err := sdkClient.Schedule.CronCreate(ctx, host, client.CronCreateOpts{
			Name:     name,
			Schedule: schedule,
			Command:  command,
			User:     user,
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
				Hostname: resp.Data.Name,
				Changed:  changedPtr,
			},
		}
		headers, rows := cli.BuildMutationTable(results, nil)
		cli.PrintCompactTable([]cli.Section{{Headers: headers, Rows: rows}})
	},
}

func init() {
	clientNodeScheduleCronCmd.AddCommand(clientNodeScheduleCronCreateCmd)

	clientNodeScheduleCronCreateCmd.PersistentFlags().
		String("name", "", "Name for the cron drop-in entry (required)")
	clientNodeScheduleCronCreateCmd.PersistentFlags().
		String("schedule", "", "Cron schedule expression, e.g. \"*/5 * * * *\" (required)")
	clientNodeScheduleCronCreateCmd.PersistentFlags().
		String("command", "", "Command to execute on schedule (required)")
	clientNodeScheduleCronCreateCmd.PersistentFlags().
		String("user", "", "User to run the command as (default root)")

	_ = clientNodeScheduleCronCreateCmd.MarkPersistentFlagRequired("name")
	_ = clientNodeScheduleCronCreateCmd.MarkPersistentFlagRequired("schedule")
	_ = clientNodeScheduleCronCreateCmd.MarkPersistentFlagRequired("command")
}

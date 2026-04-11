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

// clientNodeUserUpdateCmd represents the user update command.
var clientNodeUserUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a user account",
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		name, _ := cmd.Flags().GetString("name")
		shell, _ := cmd.Flags().GetString("shell")
		home, _ := cmd.Flags().GetString("home")
		groups, _ := cmd.Flags().GetStringSlice("groups")
		lockFlag, _ := cmd.Flags().GetBool("lock")
		unlockFlag, _ := cmd.Flags().GetBool("unlock")

		opts := client.UserUpdateOpts{
			Shell: shell,
			Home:  home,
		}
		if cmd.Flags().Changed("groups") {
			opts.Groups = groups
		}
		if lockFlag {
			b := true
			opts.Lock = &b
		}
		if unlockFlag {
			b := false
			opts.Lock = &b
		}

		resp, err := sdkClient.User.Update(ctx, host, name, opts)
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
	clientNodeUserCmd.AddCommand(clientNodeUserUpdateCmd)

	clientNodeUserUpdateCmd.PersistentFlags().
		String("name", "", "Username to update (required)")
	clientNodeUserUpdateCmd.PersistentFlags().
		String("shell", "", "New login shell path")
	clientNodeUserUpdateCmd.PersistentFlags().
		String("home", "", "New home directory path")
	clientNodeUserUpdateCmd.PersistentFlags().
		StringSlice("groups", nil, "Supplementary groups (replaces existing)")
	clientNodeUserUpdateCmd.PersistentFlags().
		Bool("lock", false, "Lock the account")
	clientNodeUserUpdateCmd.PersistentFlags().
		Bool("unlock", false, "Unlock the account")

	_ = clientNodeUserUpdateCmd.MarkPersistentFlagRequired("name")
	clientNodeUserUpdateCmd.MarkFlagsMutuallyExclusive("lock", "unlock")
	clientNodeUserUpdateCmd.MarkFlagsOneRequired("shell", "home", "groups", "lock", "unlock")
}

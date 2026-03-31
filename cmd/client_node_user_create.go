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

// clientNodeUserCreateCmd represents the user create command.
var clientNodeUserCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a user account",
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		name, _ := cmd.Flags().GetString("name")
		uid, _ := cmd.Flags().GetInt("uid")
		gid, _ := cmd.Flags().GetInt("gid")
		home, _ := cmd.Flags().GetString("home")
		shell, _ := cmd.Flags().GetString("shell")
		groups, _ := cmd.Flags().GetStringSlice("groups")
		password, _ := cmd.Flags().GetString("password")
		system, _ := cmd.Flags().GetBool("system")

		resp, err := sdkClient.User.Create(ctx, host, client.UserCreateOpts{
			Name:     name,
			UID:      uid,
			GID:      gid,
			Home:     home,
			Shell:    shell,
			Groups:   groups,
			Password: password,
			System:   system,
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

		results := make([]cli.MutationResultRow, 0, len(resp.Data.Results))
		for _, r := range resp.Data.Results {
			var errPtr *string
			if r.Error != "" {
				errPtr = &r.Error
			}
			changed := r.Changed
			results = append(results, cli.MutationResultRow{
				Hostname: r.Hostname,
				Status:   r.Status,
				Changed:  &changed,
				Error:    errPtr,
				Fields:   []string{r.Name},
			})
		}
		headers, rows := cli.BuildMutationTable(results, []string{"NAME"})
		cli.PrintCompactTable([]cli.Section{{Headers: headers, Rows: rows}})
	},
}

func init() {
	clientNodeUserCmd.AddCommand(clientNodeUserCreateCmd)

	clientNodeUserCreateCmd.PersistentFlags().
		String("name", "", "Username for the new account (required)")
	clientNodeUserCreateCmd.PersistentFlags().
		Int("uid", 0, "Numeric user ID (system assigns if omitted)")
	clientNodeUserCreateCmd.PersistentFlags().
		Int("gid", 0, "Primary group ID (system assigns if omitted)")
	clientNodeUserCreateCmd.PersistentFlags().
		String("home", "", "Home directory path")
	clientNodeUserCreateCmd.PersistentFlags().
		String("shell", "", "Login shell path")
	clientNodeUserCreateCmd.PersistentFlags().
		StringSlice("groups", nil, "Supplementary groups (comma-separated)")
	clientNodeUserCreateCmd.PersistentFlags().
		String("password", "", "Initial password (plaintext, hashed by the agent)")
	clientNodeUserCreateCmd.PersistentFlags().
		Bool("system", false, "Create a system account")

	_ = clientNodeUserCreateCmd.MarkPersistentFlagRequired("name")
}

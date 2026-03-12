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
	"github.com/retr0h/osapi/pkg/sdk/client/gen"
)

// clientContainerListCmd represents the clientContainerList command.
var clientContainerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List containers on target node",
	Long:  `List containers on the target node, optionally filtered by state.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		stateFlag, _ := cmd.Flags().GetString("state")
		limit, _ := cmd.Flags().GetInt("limit")

		params := &gen.GetNodeContainerParams{}

		if stateFlag != "" {
			state := gen.GetNodeContainerParamsState(stateFlag)
			params.State = &state
		}
		if limit > 0 {
			params.Limit = &limit
		}

		resp, err := sdkClient.Container.List(ctx, host, params)
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
			fmt.Println()
		}

		for _, r := range resp.Data.Results {
			if r.Error != "" {
				cli.PrintKV("Hostname", r.Hostname, "Error", r.Error)
				continue
			}

			rows := make([][]string, 0, len(r.Containers))
			for _, c := range r.Containers {
				rows = append(rows, []string{
					c.ID,
					c.Name,
					c.Image,
					c.State,
					c.Created,
				})
			}

			cli.PrintCompactTable([]cli.Section{{
				Title:   r.Hostname,
				Headers: []string{"ID", "NAME", "IMAGE", "STATE", "CREATED"},
				Rows:    rows,
			}})
		}
	},
}

func init() {
	clientContainerCmd.AddCommand(clientContainerListCmd)

	clientContainerListCmd.PersistentFlags().
		String("state", "running", "Filter by state: running, stopped, all")
	clientContainerListCmd.PersistentFlags().
		Int("limit", 0, "Maximum number of containers to return")
}

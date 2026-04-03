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
	"strconv"

	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
)

// clientFileStaleCmd represents the clientFileStale command.
var clientFileStaleCmd = &cobra.Command{
	Use:   "stale",
	Short: "List stale file deployments",
	Long: `List deployments where the source object has been updated since
the file was last deployed. Shows which files need redeployment.

Requires file:read permission.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()

		resp, err := sdkClient.File.Stale(ctx)
		if err != nil {
			cli.HandleError(err, logger)
			return
		}

		if jsonOutput {
			fmt.Println(string(resp.RawJSON()))
			return
		}

		fmt.Println()
		cli.PrintKV("Total", strconv.Itoa(resp.Data.Total))

		if len(resp.Data.Stale) == 0 {
			fmt.Println("  All deployments are in sync.")
			return
		}

		rows := make([][]string, 0, len(resp.Data.Stale))
		for _, s := range resp.Data.Stale {
			deployedSHA := s.DeployedSHA
			if len(deployedSHA) > 12 {
				deployedSHA = deployedSHA[:12] + "…"
			}

			currentSHA := s.CurrentSHA
			if currentSHA == "" {
				currentSHA = "(deleted)"
			} else if len(currentSHA) > 12 {
				currentSHA = currentSHA[:12] + "…"
			}

			rows = append(rows, []string{
				s.ObjectName,
				s.Hostname,
				s.Provider,
				s.DeployedAt,
				deployedSHA,
				currentSHA,
			})
		}

		cli.PrintCompactTable([]cli.Section{
			{
				Title: "Stale Deployments",
				Headers: []string{
					"OBJECT",
					"HOSTNAME",
					"PROVIDER",
					"DEPLOYED",
					"DEPLOYED SHA",
					"CURRENT SHA",
				},
				Rows: rows,
			},
		})
	},
}

func init() {
	clientFileCmd.AddCommand(clientFileStaleCmd)
}

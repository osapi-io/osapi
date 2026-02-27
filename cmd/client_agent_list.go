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
	"net/http"
	"time"

	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
)

// clientAgentListCmd represents the clientAgentList command.
var clientAgentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List active agents",
	Long: `Discover all active agents by querying the agent registry.
Shows each agent's hostname, status, labels, age, load, and OS.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()

		resp, err := sdkClient.Agent.List(ctx)
		if err != nil {
			cli.LogFatal(logger, "failed to list agents", err)
		}

		switch resp.StatusCode() {
		case http.StatusOK:
			if jsonOutput {
				fmt.Println(string(resp.Body))
				return
			}

			if resp.JSON200 == nil {
				cli.LogFatal(logger, "failed response", fmt.Errorf("agents response was nil"))
			}

			agents := resp.JSON200.Agents
			if len(agents) == 0 {
				fmt.Println("No active agents found.")
				return
			}

			rows := make([][]string, 0, len(agents))
			for _, a := range agents {
				labels := cli.FormatLabels(a.Labels)
				age := ""
				if a.StartedAt != nil {
					age = cli.FormatAge(time.Since(*a.StartedAt))
				}
				loadStr := ""
				if a.LoadAverage != nil {
					loadStr = fmt.Sprintf("%.2f", a.LoadAverage.N1min)
				}
				osStr := ""
				if a.OsInfo != nil {
					osStr = a.OsInfo.Distribution + " " + a.OsInfo.Version
				}
				rows = append(rows, []string{
					a.Hostname,
					string(a.Status),
					labels,
					age,
					loadStr,
					osStr,
				})
			}

			sections := []cli.Section{
				{
					Title:   fmt.Sprintf("Active Agents (%d)", resp.JSON200.Total),
					Headers: []string{"HOSTNAME", "STATUS", "LABELS", "AGE", "LOAD (1m)", "OS"},
					Rows:    rows,
				},
			}
			cli.PrintCompactTable(sections)
		case http.StatusUnauthorized:
			cli.HandleAuthError(resp.JSON401, resp.StatusCode(), logger)
		case http.StatusForbidden:
			cli.HandleAuthError(resp.JSON403, resp.StatusCode(), logger)
		default:
			cli.HandleUnknownError(resp.JSON500, resp.StatusCode(), logger)
		}
	},
}

func init() {
	clientAgentCmd.AddCommand(clientAgentListCmd)
}

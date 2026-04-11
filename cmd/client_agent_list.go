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
	"strings"
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
		pending, _ := cmd.Flags().GetBool("pending")

		if pending {
			listPendingAgents(cmd)
			return
		}

		resp, err := sdkClient.Agent.List(ctx)
		if err != nil {
			cli.HandleError(err, logger)
			return
		}

		if jsonOutput {
			fmt.Println(string(resp.RawJSON()))
			return
		}

		agents := resp.Data.Agents
		if len(agents) == 0 {
			fmt.Println("No active agents found.")
			return
		}

		rows := make([][]string, 0, len(agents))
		for _, a := range agents {
			status := a.State
			if status == "" {
				status = "Ready"
			}
			conditions := "-"
			if len(a.Conditions) > 0 {
				active := make([]string, 0)
				for _, c := range a.Conditions {
					if c.Status {
						active = append(active, c.Type)
					}
				}
				if len(active) > 0 {
					conditions = strings.Join(active, ",")
				}
			}
			labels := cli.FormatLabels(a.Labels)
			age := ""
			if !a.StartedAt.IsZero() {
				age = cli.FormatAge(time.Since(a.StartedAt))
			}
			loadStr := ""
			if a.LoadAverage != nil {
				loadStr = fmt.Sprintf("%.2f", a.LoadAverage.OneMin)
			}
			osStr := ""
			if a.OSInfo != nil {
				osStr = a.OSInfo.Distribution + " " + a.OSInfo.Version
			}
			machineID := ""
			if a.MachineId != "" {
				machineID = a.MachineId
				if len(machineID) > 12 {
					machineID = machineID[:12]
				}
			}
			rows = append(rows, []string{
				machineID,
				a.Hostname,
				status,
				conditions,
				labels,
				age,
				loadStr,
				osStr,
			})
		}

		sections := []cli.Section{
			{
				Title: fmt.Sprintf("Active Agents (%d)", resp.Data.Total),
				Headers: []string{
					"MACHINE ID",
					"HOSTNAME",
					"STATUS",
					"CONDITIONS",
					"LABELS",
					"AGE",
					"LOAD (1m)",
					"OS",
				},
				Rows: rows,
			},
		}
		cli.PrintCompactTable(sections)
	},
}

func listPendingAgents(
	cmd *cobra.Command,
) {
	ctx := cmd.Context()

	resp, err := sdkClient.Agent.ListPending(ctx)
	if err != nil {
		cli.HandleError(err, logger)
		return
	}

	if jsonOutput {
		fmt.Println(string(resp.RawJSON()))
		return
	}

	agents := resp.Data.Agents
	if len(agents) == 0 {
		fmt.Println("No pending agents found.")
		return
	}

	rows := make([][]string, 0, len(agents))
	for _, a := range agents {
		machineID := a.MachineID
		if len(machineID) > 12 {
			machineID = machineID[:12]
		}
		fingerprint := a.Fingerprint
		if len(fingerprint) > 20 {
			fingerprint = fingerprint[:20] + "..."
		}
		rows = append(rows, []string{
			machineID,
			a.Hostname,
			fingerprint,
			cli.FormatAge(time.Since(a.RequestedAt)),
		})
	}

	sections := []cli.Section{
		{
			Title: fmt.Sprintf("Pending Agents (%d)", resp.Data.Total),
			Headers: []string{
				"MACHINE ID",
				"HOSTNAME",
				"FINGERPRINT",
				"REQUESTED",
			},
			Rows: rows,
		},
	}
	cli.PrintCompactTable(sections)
}

func init() {
	clientAgentCmd.AddCommand(clientAgentListCmd)
	clientAgentListCmd.Flags().Bool("pending", false, "List agents awaiting enrollment")
}

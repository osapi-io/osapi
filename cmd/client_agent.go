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

	"github.com/osapi-io/osapi/internal/cli"
	clicommand "github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/controller/endpoints/agent"
)

// clientAgentCmd represents the clientAgent command.
var clientAgentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage the agents in the fleet: discovery, drain and PKI enrollment",
	Long:  "Most commands read the NATS KV heartbeat registry directly, and make no job.",
}

func init() {
	clientCmd.AddCommand(clientAgentCmd)
}

// The agent group's commands are generated from its definition, and print
// from their output types. agent list has a renderer.
func init() {
	clicommand.Render(agent.Group, "list", renderAgentList).
		Because("five of its nine columns are made from other fields")
}

func renderAgentList(
	_ *cobra.Command,
	_ agent.ListInput,
	out agent.ListAgentsResponse,
) {
	if len(out.Items) == 0 {
		cli.PrintHeading("Agents (0)")
		cli.PrintNote("None. An agent appears here once it starts and connects.")

		return
	}

	rows := make([][]string, 0, len(out.Items))
	for _, a := range out.Items {
		status := a.State
		if status == "" {
			status = "Ready"
		}
		conditions := "-"
		var active []string
		for _, c := range a.Conditions {
			if c.Status {
				active = append(active, c.Type)
			}
		}
		if len(active) > 0 {
			conditions = strings.Join(active, ",")
		}
		age := ""
		if !a.StartedAt.IsZero() {
			age = cli.FormatAge(time.Since(a.StartedAt))
		}
		load := ""
		if a.LoadAverage != nil {
			load = fmt.Sprintf("%.2f", a.LoadAverage.OneMin)
		}
		os := ""
		if a.OSInfo != nil {
			os = a.OSInfo.Distribution + " " + a.OSInfo.Version
		}
		rows = append(rows, []string{
			a.MachineID, a.Hostname, status, conditions, cli.FormatLabels(a.Labels),
			age, load, os, a.Version,
		})
	}

	cli.PrintCompactTable([]cli.Section{{
		Title: fmt.Sprintf("Agents (%d)", out.TotalItems),
		Headers: []string{
			"MACHINE ID", "HOSTNAME", "STATUS", "CONDITIONS", "LABELS",
			"AGE", "LOAD (1m)", "OS", "VERSION",
		},
		Rows: rows,
	}})
}

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
	"github.com/osapi-io/osapi/internal/controller/endpoint/command"
	"github.com/osapi-io/osapi/internal/controller/endpoints/agent"
)

// clientAgentCmd represents the clientAgent command.
var clientAgentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage the agents in the fleet: discovery, drain and PKI enrollment",
	Long:  "Most commands read the NATS KV heartbeat registry directly, and make no job.",
}

// clientAgentPendingCmd holds the commands of the agents that wait for
// enrollment, such as osapi client agent pending list.
var clientAgentPendingCmd = &cobra.Command{
	Use:   "pending",
	Short: "Read the agents that wait for enrollment",
}

func init() {
	clientCmd.AddCommand(clientAgentCmd)
	clientAgentCmd.AddCommand(clientAgentPendingCmd)
}

// The agent group's commands are generated from its definition. These are
// how each answer prints without --json.
func init() {
	command.Render(agent.Group, "list", renderAgentList)
	command.Render(agent.Group, "pending", renderAgentPending)
	command.Render(
		agent.Group,
		"get",
		func(_ *cobra.Command, _ agent.GetInput, out agent.AgentInfo) {
			renderAgentDetail(out)
		},
	)
	command.Render(
		agent.Group,
		"accept",
		func(_ *cobra.Command, _ agent.AcceptInput, out agent.Message) {
			fmt.Println()
			cli.PrintKV("Status", "Accepted", "Message", out.Message)
		},
	)
	command.Render(
		agent.Group,
		"reject",
		func(_ *cobra.Command, _ agent.RejectInput, out agent.Message) {
			fmt.Println()
			cli.PrintKV("Status", "Rejected", "Message", out.Message)
		},
	)
	command.Render(
		agent.Group,
		"delete",
		func(_ *cobra.Command, _ agent.DeleteInput, out agent.Message) {
			fmt.Println()
			cli.PrintKV("Status", "Removed", "Message", out.Message)
		},
	)
	command.Render(
		agent.Group,
		"drain",
		func(_ *cobra.Command, in agent.DrainInput, out agent.Message) {
			fmt.Println()
			cli.PrintKV("Hostname", in.Hostname, "Status", "Draining")
			cli.PrintKV("Message", out.Message)
		},
	)
	command.Render(
		agent.Group,
		"undrain",
		func(_ *cobra.Command, in agent.UndrainInput, out agent.Message) {
			fmt.Println()
			cli.PrintKV("Hostname", in.Hostname, "Status", "Ready")
			cli.PrintKV("Message", out.Message)
		},
	)
}

func renderAgentList(
	_ *cobra.Command,
	_ agent.ListInput,
	out agent.ListAgentsResponse,
) {
	if len(out.Agents) == 0 {
		cli.PrintHeading("Agents (0)")
		cli.PrintNote("None. An agent appears here once it starts and connects.")

		return
	}

	rows := make([][]string, 0, len(out.Agents))
	for _, a := range out.Agents {
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
		Title: fmt.Sprintf("Agents (%d)", out.Total),
		Headers: []string{
			"MACHINE ID", "HOSTNAME", "STATUS", "CONDITIONS", "LABELS",
			"AGE", "LOAD (1m)", "OS", "VERSION",
		},
		Rows: rows,
	}})
}

func renderAgentPending(
	_ *cobra.Command,
	_ agent.ListInput,
	out agent.ListPendingAgentsResponse,
) {
	if len(out.Agents) == 0 {
		cli.PrintHeading("Pending agents (0)")
		cli.PrintNote("None.")

		return
	}

	rows := make([][]string, 0, len(out.Agents))
	for _, a := range out.Agents {
		fingerprint := a.Fingerprint
		if len(fingerprint) > 20 {
			fingerprint = fingerprint[:20] + "..."
		}
		rows = append(rows, []string{
			a.MachineID, a.Hostname, fingerprint, cli.FormatAge(time.Since(a.RequestedAt)),
		})
	}

	cli.PrintCompactTable([]cli.Section{{
		Title:   fmt.Sprintf("Pending agents (%d)", out.Total),
		Headers: []string{"MACHINE ID", "HOSTNAME", "FINGERPRINT", "REQUESTED"},
		Rows:    rows,
	}})
}

// renderAgentDetail is one agent, as agent get prints it.
func renderAgentDetail(
	data agent.AgentInfo,
) {
	cli.PrintHeading("Agent")

	kv := []string{"Hostname", data.Hostname}
	if data.MachineID != "" {
		kv = append(kv, "Machine ID", data.MachineID)
	}
	if data.Fingerprint != "" {
		kv = append(kv, "Fingerprint", data.Fingerprint)
	}
	if data.Version != "" {
		kv = append(kv, "Version", data.Version)
	}
	kv = append(kv, "Status", data.Status)
	cli.PrintKV(kv...)

	if data.State != "" && data.State != "Ready" {
		cli.PrintKV("State", data.State)
	}
	if len(data.Labels) > 0 {
		cli.PrintKV("Labels", cli.FormatLabels(data.Labels))
	}
	if data.OSInfo != nil {
		cli.PrintKV("OS", data.OSInfo.Distribution+" "+cli.DimStyle.Render(data.OSInfo.Version))
	}
	if data.Uptime != "" {
		cli.PrintKV("Uptime", data.Uptime)
	}
	if !data.StartedAt.IsZero() {
		cli.PrintKV("Age", cli.FormatAge(time.Since(data.StartedAt)))
	}
	if !data.RegisteredAt.IsZero() {
		cli.PrintKV("Last seen", cli.FormatAge(time.Since(data.RegisteredAt))+" ago")
	}
	if data.Architecture != "" {
		cli.PrintKV("Arch", data.Architecture)
	}
	if data.KernelVersion != "" {
		cli.PrintKV("Kernel", data.KernelVersion)
	}
	if data.CPUCount > 0 {
		cli.PrintKV("CPUs", fmt.Sprintf("%d", data.CPUCount))
	}
	if data.FQDN != "" {
		cli.PrintKV("FQDN", data.FQDN)
	}
	if data.LoadAverage != nil {
		cli.PrintKV("Load", fmt.Sprintf("%.2f, %.2f, %.2f",
			data.LoadAverage.OneMin, data.LoadAverage.FiveMin, data.LoadAverage.FifteenMin,
		)+" "+cli.DimStyle.Render("(1m, 5m, 15m)"))
	}
	if data.Memory != nil {
		mem := []string{cli.FormatBytes(data.Memory.Total) + " total"}
		if data.Memory.Used > 0 {
			mem = append(mem, cli.FormatBytes(data.Memory.Used)+" used")
		}
		mem = append(mem, cli.FormatBytes(data.Memory.Free)+" free")
		cli.PrintKV("Memory", strings.Join(mem, ", "))
	}
	if data.ServiceMgr != "" {
		cli.PrintKV("Service manager", data.ServiceMgr)
	}
	if data.PackageMgr != "" {
		cli.PrintKV("Package manager", data.PackageMgr)
	}
	if data.PrimaryInterface != "" {
		cli.PrintKV("Primary interface", data.PrimaryInterface)
	}
	for _, iface := range data.Interfaces {
		var parts []string
		if iface.IPv4 != "" {
			parts = append(parts, iface.IPv4)
		}
		if iface.IPv6 != "" {
			parts = append(parts, iface.IPv6)
		}
		if iface.MAC != "" {
			parts = append(parts, cli.DimStyle.Render(iface.MAC))
		}
		cli.PrintKV("Interface "+iface.Name, strings.Join(parts, "  "))
	}

	var sections []cli.Section
	if len(data.Routes) > 0 {
		rows := make([][]string, 0, len(data.Routes))
		for _, r := range data.Routes {
			rows = append(rows, []string{
				r.Destination, r.Gateway, r.Interface, r.Mask, fmt.Sprintf("%d", r.Metric),
			})
		}
		sections = append(sections, cli.Section{
			Title:   "Routes",
			Headers: []string{"DESTINATION", "GATEWAY", "INTERFACE", "MASK", "METRIC"},
			Rows:    rows,
		})
	}
	if len(data.Conditions) > 0 {
		rows := make([][]string, 0, len(data.Conditions))
		for _, c := range data.Conditions {
			since := ""
			if !c.LastTransitionTime.IsZero() {
				since = cli.FormatAge(time.Since(c.LastTransitionTime)) + " ago"
			}
			rows = append(rows, []string{c.Type, fmt.Sprintf("%t", c.Status), c.Reason, since})
		}
		sections = append(sections, cli.Section{
			Title:   "Conditions",
			Headers: []string{"TYPE", "STATUS", "REASON", "SINCE"},
			Rows:    rows,
		})
	}
	timeline := make([][]string, 0, len(data.Timeline))
	for _, te := range data.Timeline {
		timeline = append(timeline, []string{
			te.Timestamp.Format(time.RFC3339), te.Event, te.Hostname, te.Message, te.Error,
		})
	}
	sections = append(sections, cli.Section{
		Title:   "Timeline",
		Headers: []string{"TIMESTAMP", "EVENT", "HOSTNAME", "MESSAGE", "ERROR"},
		Rows:    timeline,
	})

	for _, sec := range sections {
		cli.PrintCompactTable([]cli.Section{sec})
	}
}

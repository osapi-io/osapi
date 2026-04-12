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

	"github.com/retr0h/osapi/pkg/sdk/client"
	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
)

// clientAgentGetCmd represents the clientAgentGet command.
var clientAgentGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get agent details",
	Long:  `Get detailed information about a specific agent by hostname.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		hostname, _ := cmd.Flags().GetString("hostname")

		resp, err := sdkClient.Agent.Get(ctx, hostname)
		if err != nil {
			cli.HandleError(err, logger)
			return
		}

		if jsonOutput {
			fmt.Println(string(resp.RawJSON()))
			return
		}

		displayAgentGetDetail(&resp.Data)
	},
}

// displayAgentGetDetail renders detailed agent information in PrintKV style.
func displayAgentGetDetail(
	data *client.Agent,
) {
	fmt.Println()

	kvArgs := []string{"Hostname", data.Hostname}
	if data.MachineID != "" {
		kvArgs = append(kvArgs, "Machine ID", data.MachineID)
	}
	if data.Fingerprint != "" {
		kvArgs = append(kvArgs, "Fingerprint", data.Fingerprint)
	}
	kvArgs = append(kvArgs, "Status", data.Status)
	cli.PrintKV(kvArgs...)

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
		cli.PrintKV("Last Seen", cli.FormatAge(time.Since(data.RegisteredAt))+" ago")
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

	if data.Fqdn != "" {
		cli.PrintKV("FQDN", data.Fqdn)
	}

	if data.LoadAverage != nil {
		cli.PrintKV("Load", fmt.Sprintf("%.2f, %.2f, %.2f",
			data.LoadAverage.OneMin, data.LoadAverage.FiveMin, data.LoadAverage.FifteenMin,
		)+" "+cli.DimStyle.Render("(1m, 5m, 15m)"))
	}

	if data.Memory != nil {
		memParts := []string{cli.FormatBytes(data.Memory.Total) + " total"}
		if data.Memory.Used > 0 {
			memParts = append(memParts, cli.FormatBytes(data.Memory.Used)+" used")
		}
		memParts = append(memParts, cli.FormatBytes(data.Memory.Free)+" free")
		cli.PrintKV("Memory", strings.Join(memParts, ", "))
	}

	if data.ServiceMgr != "" {
		cli.PrintKV("Service Mgr", data.ServiceMgr)
	}

	if data.PackageMgr != "" {
		cli.PrintKV("Package Mgr", data.PackageMgr)
	}

	if data.PrimaryInterface != "" {
		cli.PrintKV("Primary Iface", data.PrimaryInterface)
	}

	if len(data.Interfaces) > 0 {
		for _, iface := range data.Interfaces {
			parts := []string{}
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
	}

	var sections []cli.Section

	if len(data.Routes) > 0 {
		routeRows := make([][]string, 0, len(data.Routes))
		for _, r := range data.Routes {
			routeRows = append(routeRows, []string{
				r.Destination, r.Gateway, r.Interface, r.Mask, fmt.Sprintf("%d", r.Metric),
			})
		}
		sections = append(sections, cli.Section{
			Title:   "Routes",
			Headers: []string{"DESTINATION", "GATEWAY", "INTERFACE", "MASK", "METRIC"},
			Rows:    routeRows,
		})
	}

	if len(data.Conditions) > 0 {
		condRows := make([][]string, 0, len(data.Conditions))
		for _, c := range data.Conditions {
			status := "false"
			if c.Status {
				status = "true"
			}
			reason := c.Reason
			since := ""
			if !c.LastTransitionTime.IsZero() {
				since = cli.FormatAge(time.Since(c.LastTransitionTime)) + " ago"
			}
			condRows = append(condRows, []string{c.Type, status, reason, since})
		}
		sections = append(sections, cli.Section{
			Title:   "Conditions",
			Headers: []string{"TYPE", "STATUS", "REASON", "SINCE"},
			Rows:    condRows,
		})
	}

	timelineRows := make([][]string, 0, len(data.Timeline))
	for _, te := range data.Timeline {
		timelineRows = append(
			timelineRows,
			[]string{te.Timestamp, te.Event, te.Hostname, te.Message, te.Error},
		)
	}
	if len(timelineRows) == 0 {
		timelineRows = [][]string{{"No events"}}
	}
	sections = append(sections, cli.Section{
		Title:   "Timeline",
		Headers: []string{"TIMESTAMP", "EVENT", "HOSTNAME", "MESSAGE", "ERROR"},
		Rows:    timelineRows,
	})

	for _, sec := range sections {
		cli.PrintCompactTable([]cli.Section{sec})
	}
}

func init() {
	clientAgentCmd.AddCommand(clientAgentGetCmd)
	clientAgentGetCmd.Flags().String("hostname", "", "Hostname of the agent to retrieve")
	_ = clientAgentGetCmd.MarkFlagRequired("hostname")
}

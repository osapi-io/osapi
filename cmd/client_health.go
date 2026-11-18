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
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/controller/endpoint/command"
	"github.com/osapi-io/osapi/internal/controller/endpoints/health"
	"github.com/osapi-io/osapi/pkg/sdk/client"
)

// clientHealthCmd represents the health parent command.
var clientHealthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check the health of the controller",
}

func init() {
	clientCmd.AddCommand(clientHealthCmd)
}

func init() {
	// osapi client health liveness, ready and status are generated from the
	// health group's definition; this is how each answer prints. ready and
	// status print their 503 too, which says what is not healthy.
	command.Render(health.Group, "liveness", renderHealthLiveness)
	command.Render(health.Group, "ready", renderHealthReady)
	command.Render(health.Group, "status", renderHealthStatus)
}

// renderHealthLiveness prints that the process is alive.
func renderHealthLiveness(
	_ *cobra.Command,
	_ health.LivenessInput,
	out health.HealthResponse,
) {
	cli.PrintHeading("Health")
	cli.PrintKV("Status", out.Status)
}

// renderHealthReady prints whether the controller can take traffic, and
// why not.
func renderHealthReady(
	_ *cobra.Command,
	_ health.ReadyInput,
	out health.ReadyResponse,
) {
	cli.PrintHeading("Health")
	cli.PrintKV("Status", out.Status)
	if out.Error != "" {
		cli.PrintKV("Error", out.Error)
	}
}

// renderHealthStatus prints the controller, each component and the
// storage, as the status commands of agent, controller and nats do.
func renderHealthStatus(
	_ *cobra.Command,
	_ health.StatusInput,
	out health.StatusResponse,
) {
	// The SDK's SystemStatus has the same JSON as the answer.
	raw, _ := json.Marshal(out)
	var status client.SystemStatus
	_ = json.Unmarshal(raw, &status)
	status.ServiceUnavailable = out.Status != "ok"

	displayStatusHealth(&status)
}

// subComponent holds a name, status, optional address and optional URL for
// display under a parent.
type subComponent struct {
	name    string
	status  string
	address string
	url     string
}

// subComponentsFor discovers sub-components for the given component type
// by matching keys with the "{type}." prefix in the components map.
func subComponentsFor(
	componentType string,
	components map[string]client.ComponentHealth,
) []subComponent {
	if components == nil {
		return nil
	}

	prefix := componentType + "."
	var result []subComponent

	for key, c := range components {
		if strings.HasPrefix(key, prefix) {
			name := strings.TrimPrefix(key, prefix)
			result = append(result, subComponent{
				name:    name,
				status:  c.Status,
				address: c.Address,
				url:     c.URL,
			})
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].name < result[j].name
	})

	return result
}

// displayComponentTable renders the component registry table,
// optionally filtered by component type. Pass "" for all types.
func displayComponentTable(
	registry []client.RegistryEntry,
	filterType string,
	subComponents map[string]client.ComponentHealth,
) {
	var filtered []client.RegistryEntry
	for _, e := range registry {
		if filterType == "" || e.Type == filterType {
			filtered = append(filtered, e)
		}
	}

	if len(filtered) == 0 {
		cli.PrintNote("No components found.")

		return
	}

	rows := make([][]string, 0, len(filtered))
	for _, e := range filtered {
		conditions := "-"
		if len(e.Conditions) > 0 {
			conditions = strings.Join(e.Conditions, ", ")
		}

		cpu := fmt.Sprintf("%.1f%%", e.CPUPercent)
		mem := cli.FormatBytes(int(e.MemBytes))
		rows = append(rows, []string{
			e.Type,
			e.Hostname,
			e.Status,
			conditions,
			e.Age,
			cpu,
			mem,
		})

		// Show sub-components indented under their parent.
		scs := subComponentsFor(e.Type, subComponents)
		for i, sc := range scs {
			prefix := "  ├─ "
			if i == len(scs)-1 {
				prefix = "  └─ "
			}

			// The status word in its status color, then the address dim,
			// because nothing connects through it.
			status := cli.RenderStatus(sc.status)
			if sc.address != "" {
				status += " " + cli.DimStyle.Render(sc.address)
			}

			rows = append(rows, []string{
				"",
				prefix + sc.name,
				status,
				"",
				"",
				"",
				"",
			})

			// The url goes on a line of its own, so a long one is never cut
			// and every row keeps the same shape. The tree line continues
			// down to the next sibling; under the last one it stops.
			if sc.url != "" {
				branch := "  │"
				if i == len(scs)-1 {
					branch = ""
				}

				rows = append(rows, []string{
					"",
					branch,
					cli.DimStyle.Render("↳") + " " + cli.URLStyle.Render(sc.url),
					"",
					"",
					"",
					"",
				})
			}
		}
	}

	cli.PrintCompactTable([]cli.Section{{
		Headers: []string{"TYPE", "HOSTNAME", "STATUS", "CONDITIONS", "AGE", "CPU", "MEM"},
		Rows:    rows,
	}})
}

// displayStatusHealth renders health status output with system metrics.
func displayStatusHealth(
	data *client.SystemStatus,
) {
	cli.PrintHeading("Controller")
	cli.PrintKV("Status", data.Status)
	cli.PrintKV("Version", data.Version)
	cli.PrintKV("Uptime", data.Uptime)

	cli.PrintHeading("Components")
	displayComponentTable(data.Registry, "", data.Components)

	// Agent details are shown in the component table above.
	// Use "osapi client agent list" for labels and detailed info.

	cli.PrintHeading("Storage")
	if data.Jobs != nil {
		cli.PrintKV("Jobs", strings.Join([]string{
			cli.Count(data.Jobs.Total, "total"),
			cli.Count(data.Jobs.Completed, "completed"),
			cli.Count(data.Jobs.Unprocessed, "unprocessed"),
			cli.Count(data.Jobs.Failed, "failed"),
			cli.Count(data.Jobs.Dlq, "dlq"),
		}, ", "))
	}

	// Streams
	for _, s := range data.Streams {
		cli.PrintKV("Stream", fmt.Sprintf(
			"%s "+cli.DimStyle.Render("(%d msgs, %s, %d consumers)"),
			s.Name, s.Messages, cli.FormatBytes(s.Bytes), s.Consumers,
		))
	}

	// KV Buckets
	for _, b := range data.KVBuckets {
		cli.PrintKV("Bucket", fmt.Sprintf(
			"%s "+cli.DimStyle.Render("(%d keys, %s)"),
			b.Name, b.Keys, cli.FormatBytes(b.Bytes),
		))
	}

	// Object Stores
	for _, o := range data.ObjectStores {
		cli.PrintKV("Object Store", fmt.Sprintf(
			"%s "+cli.DimStyle.Render("(%s)"),
			o.Name, cli.FormatBytes(o.Size),
		))
	}

	// Consumers last — the table can be long with many agents
	if data.Consumers != nil {
		cli.PrintHeading("Consumers")
		cli.PrintKV("Total", fmt.Sprintf("%d", data.Consumers.Total))
		if len(data.Consumers.Consumers) > 0 {
			rows := make([][]string, 0, len(data.Consumers.Consumers))
			for _, c := range data.Consumers.Consumers {
				rows = append(rows, []string{
					c.Name,
					fmt.Sprintf("%d", c.Pending),
					fmt.Sprintf("%d", c.AckPending),
					fmt.Sprintf("%d", c.Redelivered),
				})
			}
			cli.PrintCompactTable([]cli.Section{{
				Headers: []string{"NAME", "PENDING", "ACK PENDING", "REDELIVERED"},
				Rows:    rows,
			}})
		}
	}
}

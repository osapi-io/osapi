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

	"github.com/osapi-io/osapi-sdk/pkg/osapi/gen"
	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
)

// clientHealthStatusCmd represents the clientHealthStatus command.
var clientHealthStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "System status and component health",
	Long: `Show per-component health status with system metrics.
Requires authentication.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		resp, err := sdkClient.Health.Status(ctx)
		if err != nil {
			cli.LogFatal(logger, "failed to get health status endpoint", err)
		}

		switch resp.StatusCode() {
		case http.StatusOK:
			if jsonOutput {
				fmt.Println(string(resp.Body))
				return
			}

			if resp.JSON200 == nil {
				cli.LogFatal(
					logger,
					"failed response",
					fmt.Errorf("health status response was nil"),
				)
			}

			displayStatusHealth(resp.JSON200)

		case http.StatusServiceUnavailable:
			if jsonOutput {
				fmt.Println(string(resp.Body))
				return
			}

			if resp.JSON503 == nil {
				cli.LogFatal(
					logger,
					"failed response",
					fmt.Errorf("health status response was nil"),
				)
			}

			displayStatusHealth(resp.JSON503)

		case http.StatusUnauthorized:
			cli.HandleAuthError(resp.JSON401, resp.StatusCode(), logger)
		case http.StatusForbidden:
			cli.HandleAuthError(resp.JSON403, resp.StatusCode(), logger)
		default:
			cli.HandleUnknownError(nil, resp.StatusCode(), logger)
		}
	},
}

// displayStatusHealth renders health status output with system metrics.
func displayStatusHealth(
	data *gen.StatusResponse,
) {
	fmt.Println()
	cli.PrintKV("Status", data.Status, "Version", data.Version, "Uptime", data.Uptime)

	// NATS connection info (merged with component health)
	if data.Nats != nil {
		natsStatus := "ok"
		if c, ok := data.Components["nats"]; ok && c.Status != "ok" {
			natsStatus = c.Status
			if c.Error != nil {
				natsStatus += " " + cli.DimStyle.Render(*c.Error)
			}
		}
		natsVal := natsStatus + " " + cli.DimStyle.Render(data.Nats.Url)
		if data.Nats.Version != "" {
			natsVal += " " + cli.DimStyle.Render("(v"+data.Nats.Version+")")
		}
		cli.PrintKV("NATS", natsVal)
	}

	// KV component (without duplicating the NATS line)
	if c, ok := data.Components["kv"]; ok {
		kvVal := c.Status
		if c.Error != nil {
			kvVal += " " + cli.DimStyle.Render(*c.Error)
		}
		cli.PrintKV("KV", kvVal)
	}

	// Non-standard components (skip nats/kv already shown above)
	for name, component := range data.Components {
		if name == "nats" || name == "kv" {
			continue
		}
		val := component.Status
		if component.Error != nil {
			val += " " + cli.DimStyle.Render(*component.Error)
		}
		cli.PrintKV(name, val)
	}

	if data.Agents != nil {
		cli.PrintKV("Agents", fmt.Sprintf(
			"%d total, %d ready",
			data.Agents.Total, data.Agents.Ready,
		))
		if data.Agents.Agents != nil {
			rows := make([][]string, 0, len(*data.Agents.Agents))
			for _, a := range *data.Agents.Agents {
				labels := ""
				if a.Labels != nil {
					labels = *a.Labels
				}
				rows = append(rows, []string{a.Hostname, labels, a.Registered})
			}
			cli.PrintCompactTable([]cli.Section{{
				Headers: []string{"HOSTNAME", "LABELS", "REGISTERED"},
				Rows:    rows,
			}})
			fmt.Println()
		}
	}

	if data.Jobs != nil {
		cli.PrintKV("Jobs", fmt.Sprintf(
			"%d total, %d completed, %d unprocessed, %d failed, %d dlq",
			data.Jobs.Total, data.Jobs.Completed,
			data.Jobs.Unprocessed, data.Jobs.Failed, data.Jobs.Dlq,
		))
	}

	// Streams
	if data.Streams != nil {
		for _, s := range *data.Streams {
			cli.PrintKV("Stream", fmt.Sprintf(
				"%s "+cli.DimStyle.Render("(%d msgs, %s, %d consumers)"),
				s.Name, s.Messages, cli.FormatBytes(s.Bytes), s.Consumers,
			))
		}
	}

	// KV Buckets
	if data.KvBuckets != nil {
		for _, b := range *data.KvBuckets {
			cli.PrintKV("Bucket", fmt.Sprintf(
				"%s "+cli.DimStyle.Render("(%d keys, %s)"),
				b.Name, b.Keys, cli.FormatBytes(b.Bytes),
			))
		}
	}

	// Consumers last — the table can be long with many agents
	if data.Consumers != nil {
		fmt.Println()
		cli.PrintKV("Consumers", fmt.Sprintf("%d total", data.Consumers.Total))
		if data.Consumers.Consumers != nil {
			rows := make([][]string, 0, len(*data.Consumers.Consumers))
			for _, c := range *data.Consumers.Consumers {
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

func init() {
	clientHealthCmd.AddCommand(clientHealthStatusCmd)
}

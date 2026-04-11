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

	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
)

// clientContainerDockerInspectCmd represents the clientContainerDockerInspect command.
var clientContainerDockerInspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect a container",
	Long:  `Retrieve detailed information about a specific container on the target node.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		id, _ := cmd.Flags().GetString("id")

		resp, err := sdkClient.Docker.Inspect(ctx, host, id)
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

		results := make([]cli.ResultRow, 0)
		for _, r := range resp.Data.Results {
			if r.Error != "" {
				var errPtr *string
				e := r.Error
				errPtr = &e
				results = append(results, cli.ResultRow{
					Hostname: r.Hostname,
					Status:   r.Status,
					Error:    errPtr,
				})

				continue
			}

			network := ""
			if len(r.NetworkSettings) > 0 {
				parts := make([]string, 0)
				if ip := r.NetworkSettings["ip_address"]; ip != "" {
					parts = append(parts, "ip="+ip)
				}
				if gw := r.NetworkSettings["gateway"]; gw != "" {
					parts = append(parts, "gw="+gw)
				}
				for k, v := range r.NetworkSettings {
					if k != "ip_address" && k != "gateway" && v != "" {
						parts = append(parts, k+"="+v)
					}
				}
				network = strings.Join(parts, ", ")
			}

			results = append(results, cli.ResultRow{
				Hostname: r.Hostname,
				Status:   r.Status,
				Fields: []string{
					r.ID,
					r.Name,
					r.Image,
					r.State,
					r.Created,
					r.Health,
					cli.FormatList(r.Ports),
					cli.FormatList(r.Mounts),
					network,
				},
			})
		}
		tr := cli.BuildBroadcastTable(
			results,
			[]string{
				"ID",
				"NAME",
				"IMAGE",
				"STATE",
				"CREATED",
				"HEALTH",
				"PORTS",
				"MOUNTS",
				"NETWORK",
			},
		)
		cli.PrintCompactTable([]cli.Section{{Headers: tr.Headers, Rows: tr.Rows, Errors: tr.Errors}})
	},
}

func init() {
	clientContainerDockerCmd.AddCommand(clientContainerDockerInspectCmd)

	clientContainerDockerInspectCmd.PersistentFlags().
		String("id", "", "Container ID to inspect (required)")

	_ = clientContainerDockerInspectCmd.MarkPersistentFlagRequired("id")
}

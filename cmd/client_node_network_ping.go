// Copyright (c) 2024 John Dewey

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

// clientNodeNetworkPingCmd represents the clientNodeNetworkPing command.
var clientNodeNetworkPingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Ping the specified server",
	Long: `Ping the specified server and return results.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		address, _ := cmd.Flags().GetString("address")

		resp, err := sdkClient.Node.Ping(ctx, host, address)
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

		results := make([]cli.ResultRow, 0, len(resp.Data.Results))
		for _, r := range resp.Data.Results {
			var errPtr *string
			if r.Error != "" {
				errPtr = &r.Error
			}
			results = append(results, cli.ResultRow{
				Hostname: r.Hostname,
				Status:   r.Status,
				Error:    errPtr,
				Fields: []string{
					r.AvgRtt,
					r.MaxRtt,
					r.MinRtt,
					fmt.Sprintf("%f", r.PacketLoss),
					strconv.Itoa(r.PacketsReceived),
					strconv.Itoa(r.PacketsSent),
				},
			})
		}
		headers, rows := cli.BuildBroadcastTable(results, []string{
			"AVG RTT",
			"MAX RTT",
			"MIN RTT",
			"PACKET LOSS",
			"PACKETS RECEIVED",
			"PACKETS SENT",
		})
		cli.PrintCompactTable([]cli.Section{{
			Title:   "Ping Response",
			Headers: headers,
			Rows:    rows,
		}})
	},
}

func init() {
	clientNodeNetworkCmd.AddCommand(clientNodeNetworkPingCmd)

	clientNodeNetworkPingCmd.PersistentFlags().
		StringP("address", "", "", "The address to ping")

	_ = clientNodeNetworkPingCmd.MarkPersistentFlagRequired("address")
}

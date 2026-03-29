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

	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
)

// clientNodeNetworkDNSGetCmd represents the clientNodeNetworkDNSGet command.
var clientNodeNetworkDNSGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the DNS configuration",
	Long: `Get the servers current DNS configuration.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		interfaceName, _ := cmd.Flags().GetString("interface-name")

		resp, err := sdkClient.Node.GetDNS(ctx, host, interfaceName)
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
		for _, cfg := range resp.Data.Results {
			var errPtr *string
			if cfg.Error != "" {
				errPtr = &cfg.Error
			}
			results = append(results, cli.ResultRow{
				Hostname: cfg.Hostname,
				Status:   cfg.Status,
				Error:    errPtr,
				Fields: []string{
					cli.FormatList(cfg.Servers),
					cli.FormatList(cfg.SearchDomains),
				},
			})
		}
		headers, rows := cli.BuildBroadcastTable(results, []string{
			"SERVERS",
			"SEARCH DOMAINS",
		})
		cli.PrintCompactTable([]cli.Section{{Headers: headers, Rows: rows}})
	},
}

func init() {
	clientNodeNetworkDNSCmd.AddCommand(clientNodeNetworkDNSGetCmd)

	clientNodeNetworkDNSGetCmd.PersistentFlags().
		String("interface-name", "", "Name of the network interface to retrieve DNS server configurations (required)")

	_ = clientNodeNetworkDNSGetCmd.MarkPersistentFlagRequired("interface-name")
}

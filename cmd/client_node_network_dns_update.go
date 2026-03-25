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
	"log/slog"
	"strings"

	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
)

// clientNodeNetworkDNSUpdateCmd represents the clientNodeNetworkDNSUpdate command.
var clientNodeNetworkDNSUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the DNS configuration",
	Long: `Update the current DNS configuration with the supplied options.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		servers, _ := cmd.Flags().GetStringSlice("servers")
		searchDomains, _ := cmd.Flags().GetStringSlice("search-domains")
		interfaceName, _ := cmd.Flags().GetString("interface-name")


		resp, err := sdkClient.Node.UpdateDNS(
			ctx,
			host,
			interfaceName,
			servers,
			searchDomains,
		)
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

		if len(resp.Data.Results) > 0 {
			results := make([]cli.MutationResultRow, 0, len(resp.Data.Results))
			for _, r := range resp.Data.Results {
				var errPtr *string
				if r.Error != "" {
					errPtr = &r.Error
				}
				var changedPtr *bool
				if r.Changed {
					changedPtr = &r.Changed
				}
				results = append(results, cli.MutationResultRow{
					Hostname: r.Hostname,
					Status:   r.Status,
					Changed:  changedPtr,
					Error:    errPtr,
				})
			}
			headers, rows := cli.BuildMutationTable(results, nil)
			cli.PrintCompactTable([]cli.Section{{Headers: headers, Rows: rows}})
		} else {
			logger.Info(
				"network dns put",
				slog.String("search_domains", strings.Join(searchDomains, ",")),
				slog.String("servers", strings.Join(servers, ",")),
				slog.String("status", "ok"),
			)
		}
	},
}

func init() {
	clientNodeNetworkDNSCmd.AddCommand(clientNodeNetworkDNSUpdateCmd)

	clientNodeNetworkDNSUpdateCmd.PersistentFlags().
		StringSlice("servers", []string{}, "List of DNS server IP addresses (comma-separated)")
	clientNodeNetworkDNSUpdateCmd.PersistentFlags().
		StringSlice("search-domains", []string{}, "List of DNS search domains (comma-separated)")
	clientNodeNetworkDNSUpdateCmd.PersistentFlags().
		String("interface-name", "", "Name of the network interface to retrieve DNS server configurations (required)")

	clientNodeNetworkDNSUpdateCmd.MarkFlagsOneRequired("servers", "search-domains")
	_ = clientNodeNetworkDNSUpdateCmd.MarkPersistentFlagRequired("interface-name")
}

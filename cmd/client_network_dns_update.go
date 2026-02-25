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
	"net/http"
	"strings"

	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
)

// clientNetworkDNSUpdateCmd represents the clientNetworkDNSUpdate command.
var clientNetworkDNSUpdateCmd = &cobra.Command{
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

		if host == "_all" {
			fmt.Print("This will modify DNS on ALL hosts. Continue? [y/N] ")
			var confirm string
			if _, err := fmt.Scanln(&confirm); err != nil || (confirm != "y" && confirm != "Y") {
				fmt.Println("Aborted.")
				return
			}
		}

		resp, err := sdkClient.Network.UpdateDNS(
			ctx,
			host,
			interfaceName,
			servers,
			searchDomains,
		)
		if err != nil {
			cli.LogFatal(logger, "failed to update network dns endpoint", err)
		}

		switch resp.StatusCode() {
		case http.StatusAccepted:
			if jsonOutput {
				fmt.Println(string(resp.Body))
				return
			}

			if resp.JSON202 != nil && resp.JSON202.JobId != nil {
				fmt.Println()
				cli.PrintKV("Job ID", resp.JSON202.JobId.String())
			}

			if resp.JSON202 != nil && len(resp.JSON202.Results) > 0 {
				results := make([]cli.MutationResultRow, 0, len(resp.JSON202.Results))
				for _, r := range resp.JSON202.Results {
					results = append(results, cli.MutationResultRow{
						Hostname: r.Hostname,
						Status:   string(r.Status),
						Error:    r.Error,
					})
				}
				headers, rows := cli.BuildMutationTable(results, nil)
				cli.PrintStyledTable([]cli.Section{{Headers: headers, Rows: rows}})
			} else {
				logger.Info(
					"network dns put",
					slog.String("search_domains", strings.Join(searchDomains, ",")),
					slog.String("servers", strings.Join(servers, ",")),
					slog.Int("code", resp.StatusCode()),
					slog.String("status", "ok"),
				)
			}

		case http.StatusBadRequest:
			cli.HandleUnknownError(resp.JSON400, resp.StatusCode(), logger)
		case http.StatusUnauthorized:
			cli.HandleAuthError(resp.JSON401, resp.StatusCode(), logger)
		case http.StatusForbidden:
			cli.HandleAuthError(resp.JSON403, resp.StatusCode(), logger)
		default:
			cli.HandleUnknownError(resp.JSON500, resp.StatusCode(), logger)
		}
	},
}

func init() {
	clientNetworkDNSCmd.AddCommand(clientNetworkDNSUpdateCmd)

	clientNetworkDNSUpdateCmd.PersistentFlags().
		StringSlice("servers", []string{}, "List of DNS server IP addresses (comma-separated)")
	clientNetworkDNSUpdateCmd.PersistentFlags().
		StringSlice("search-domains", []string{}, "List of DNS search domains (comma-separated)")
	clientNetworkDNSUpdateCmd.PersistentFlags().
		String("interface-name", "", "Name of the network interface to retrieve DNS server configurations (required)")

	clientNetworkDNSUpdateCmd.MarkFlagsOneRequired("servers", "search-domains")
	_ = clientNetworkDNSUpdateCmd.MarkPersistentFlagRequired("interface-name")
}

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
	"net/http"

	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
	"github.com/retr0h/osapi/internal/client"
)

// clientNetworkDNSGetCmd represents the clientNetworkDNSGet command.
var clientNetworkDNSGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the DNS configuration",
	Long: `Get the servers current DNS configuration.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		interfaceName, _ := cmd.Flags().GetString("interface-name")

		networkHandler := handler.(client.NetworkHandler)
		resp, err := networkHandler.GetNetworkDNSByInterface(ctx, host, interfaceName)
		if err != nil {
			cli.LogFatal(logger, "failed to get network dns endpoint", err)
		}

		switch resp.StatusCode() {
		case http.StatusOK:
			if jsonOutput {
				fmt.Println(string(resp.Body))
				return
			}

			if resp.JSON200 == nil {
				cli.LogFatal(logger, "failed response", fmt.Errorf("get dns response was nil"))
			}

			if resp.JSON200.JobId != nil {
				fmt.Println()
				cli.PrintKV("Job ID", resp.JSON200.JobId.String())
			}

			results := make([]cli.ResultRow, 0, len(resp.JSON200.Results))
			for _, cfg := range resp.JSON200.Results {
				var serversList, searchDomainsList []string
				if cfg.Servers != nil {
					serversList = *cfg.Servers
				}
				if cfg.SearchDomains != nil {
					searchDomainsList = *cfg.SearchDomains
				}
				results = append(results, cli.ResultRow{
					Hostname: cfg.Hostname,
					Error:    cfg.Error,
					Fields: []string{
						cli.FormatList(serversList),
						cli.FormatList(searchDomainsList),
					},
				})
			}
			headers, rows := cli.BuildBroadcastTable(results, []string{
				"SERVERS",
				"SEARCH DOMAINS",
			})
			cli.PrintStyledTable([]cli.Section{{Headers: headers, Rows: rows}})

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
	clientNetworkDNSCmd.AddCommand(clientNetworkDNSGetCmd)

	clientNetworkDNSGetCmd.PersistentFlags().
		String("interface-name", "", "Name of the network interface to retrieve DNS server configurations (required)")

	_ = clientNetworkDNSGetCmd.MarkPersistentFlagRequired("interface-name")
}

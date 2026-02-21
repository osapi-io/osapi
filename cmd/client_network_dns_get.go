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
			logFatal("failed to get network dns endpoint", err)
		}

		switch resp.StatusCode() {
		case http.StatusOK:
			if jsonOutput {
				fmt.Println(string(resp.Body))
				return
			}

			if resp.JSON200 == nil {
				logFatal("failed response", fmt.Errorf("get dns response was nil"))
			}

			if resp.JSON200.JobId != nil {
				fmt.Println()
				printKV("Job ID", resp.JSON200.JobId.String())
			}

			rows := make([][]string, 0, len(resp.JSON200.Results))
			for _, cfg := range resp.JSON200.Results {
				var serversList, searchDomainsList []string
				if cfg.Servers != nil {
					serversList = *cfg.Servers
				}
				if cfg.SearchDomains != nil {
					searchDomainsList = *cfg.SearchDomains
				}
				rows = append(rows, []string{
					cfg.Hostname,
					formatList(serversList),
					formatList(searchDomainsList),
				})
			}
			sections := []section{
				{
					Headers: []string{"HOSTNAME", "SERVERS", "SEARCH DOMAINS"},
					Rows:    rows,
				},
			}
			printStyledTable(sections)

		case http.StatusBadRequest:
			handleUnknownError(resp.JSON400, resp.StatusCode(), logger)
		case http.StatusUnauthorized:
			handleAuthError(resp.JSON401, resp.StatusCode(), logger)
		case http.StatusForbidden:
			handleAuthError(resp.JSON403, resp.StatusCode(), logger)
		default:
			handleUnknownError(resp.JSON500, resp.StatusCode(), logger)
		}
	},
}

func init() {
	clientNetworkDNSCmd.AddCommand(clientNetworkDNSGetCmd)

	clientNetworkDNSGetCmd.PersistentFlags().
		String("interface-name", "", "Name of the network interface to retrieve DNS server configurations (required)")

	_ = clientNetworkDNSGetCmd.MarkPersistentFlagRequired("interface-name")
}

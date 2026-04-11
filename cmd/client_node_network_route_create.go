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
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
	"github.com/retr0h/osapi/pkg/sdk/client"
)

// clientNodeNetworkRouteCreateCmd represents the route create command.
var clientNodeNetworkRouteCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create routes for an interface",
	Long: `Create network routes for an interface on the target node.

Route format: TO:VIA or TO:VIA:METRIC
  --route 10.1.0.0/16:10.0.0.1
  --route 10.2.0.0/16:10.0.0.1:100`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		interfaceName, _ := cmd.Flags().GetString("interface")
		routeStrs, _ := cmd.Flags().GetStringSlice("route")

		opts, err := parseRouteFlags(routeStrs)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		resp, err := sdkClient.Route.Create(ctx, host, interfaceName, opts)
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
			changed := r.Changed
			results = append(results, cli.ResultRow{
				Hostname: r.Hostname,
				Status:   r.Status,
				Changed:  &changed,
				Error:    errPtr,
				Fields:   []string{r.Interface},
			})
		}
		tr := cli.BuildMutationTable(results, []string{"INTERFACE"})
		cli.PrintCompactTable([]cli.Section{{Headers: tr.Headers, Rows: tr.Rows, Errors: tr.Errors}})
	},
}

// parseRouteFlags parses route flag values in TO:VIA or TO:VIA:METRIC format.
func parseRouteFlags(
	routeStrs []string,
) (client.RouteConfigOpts, error) {
	opts := client.RouteConfigOpts{
		Routes: make([]client.RouteItem, 0, len(routeStrs)),
	}

	for _, rs := range routeStrs {
		parts := strings.SplitN(rs, ":", 3)
		if len(parts) < 2 {
			return opts, fmt.Errorf("invalid route format %q: expected TO:VIA or TO:VIA:METRIC", rs)
		}

		item := client.RouteItem{
			To:  parts[0],
			Via: parts[1],
		}

		if len(parts) == 3 {
			metric, err := strconv.Atoi(parts[2])
			if err != nil {
				return opts, fmt.Errorf("invalid metric in route %q: %w", rs, err)
			}

			item.Metric = &metric
		}

		opts.Routes = append(opts.Routes, item)
	}

	return opts, nil
}

func init() {
	clientNodeNetworkRouteCmd.AddCommand(clientNodeNetworkRouteCreateCmd)

	clientNodeNetworkRouteCreateCmd.PersistentFlags().
		String("interface", "", "Interface name (required)")
	clientNodeNetworkRouteCreateCmd.PersistentFlags().
		StringSlice("route", []string{}, "Route in TO:VIA or TO:VIA:METRIC format (repeatable, required)")

	_ = clientNodeNetworkRouteCreateCmd.MarkPersistentFlagRequired("interface")
	_ = clientNodeNetworkRouteCreateCmd.MarkPersistentFlagRequired("route")
}

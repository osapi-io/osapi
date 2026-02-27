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
	"strings"
	"time"

	"github.com/osapi-io/osapi-sdk/pkg/osapi/gen"
	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
)

// clientAgentGetCmd represents the clientAgentGet command.
var clientAgentGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get agent details",
	Long:  `Get detailed information about a specific agent by hostname.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		hostname, _ := cmd.Flags().GetString("hostname")

		resp, err := sdkClient.Agent.Get(ctx, hostname)
		if err != nil {
			cli.LogFatal(logger, "failed to get agent details", err)
		}

		switch resp.StatusCode() {
		case http.StatusOK:
			if jsonOutput {
				fmt.Println(string(resp.Body))
				return
			}

			if resp.JSON200 == nil {
				cli.LogFatal(logger, "failed response", fmt.Errorf("agent response was nil"))
			}

			displayAgentGetDetail(resp.JSON200)
		case http.StatusUnauthorized:
			cli.HandleAuthError(resp.JSON401, resp.StatusCode(), logger)
		case http.StatusForbidden:
			cli.HandleAuthError(resp.JSON403, resp.StatusCode(), logger)
		case http.StatusNotFound:
			cli.HandleUnknownError(resp.JSON404, resp.StatusCode(), logger)
		default:
			cli.HandleUnknownError(resp.JSON500, resp.StatusCode(), logger)
		}
	},
}

// displayAgentGetDetail renders detailed agent information in PrintKV style.
func displayAgentGetDetail(
	data *gen.AgentInfo,
) {
	fmt.Println()

	kvArgs := []string{"Hostname", data.Hostname, "Status", string(data.Status)}
	cli.PrintKV(kvArgs...)

	if data.Labels != nil && len(*data.Labels) > 0 {
		cli.PrintKV("Labels", cli.FormatLabels(data.Labels))
	}

	if data.OsInfo != nil {
		cli.PrintKV("OS", data.OsInfo.Distribution+" "+cli.DimStyle.Render(data.OsInfo.Version))
	}

	if data.Uptime != nil {
		cli.PrintKV("Uptime", *data.Uptime)
	}

	if data.StartedAt != nil {
		cli.PrintKV("Age", cli.FormatAge(time.Since(*data.StartedAt)))
	}

	if data.RegisteredAt != nil {
		cli.PrintKV("Last Seen", cli.FormatAge(time.Since(*data.RegisteredAt))+" ago")
	}

	if data.LoadAverage != nil {
		cli.PrintKV("Load", fmt.Sprintf("%.2f, %.2f, %.2f",
			data.LoadAverage.N1min, data.LoadAverage.N5min, data.LoadAverage.N15min,
		)+" "+cli.DimStyle.Render("(1m, 5m, 15m)"))
	}

	if data.Memory != nil {
		memParts := []string{cli.FormatBytes(data.Memory.Total) + " total"}
		if data.Memory.Used > 0 {
			memParts = append(memParts, cli.FormatBytes(data.Memory.Used)+" used")
		}
		memParts = append(memParts, cli.FormatBytes(data.Memory.Free)+" free")
		cli.PrintKV("Memory", strings.Join(memParts, ", "))
	}
}

func init() {
	clientAgentCmd.AddCommand(clientAgentGetCmd)
	clientAgentGetCmd.Flags().String("hostname", "", "Hostname of the agent to retrieve")
	_ = clientAgentGetCmd.MarkFlagRequired("hostname")
}

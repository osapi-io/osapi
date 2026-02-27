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

// clientNodeStatusGetCmd represents the clientNodeStatusGet command.
var clientNodeStatusGetCmd = &cobra.Command{
	Use:   "status",
	Short: "Status of the server",
	Long: `Obtain the current node status.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		resp, err := sdkClient.Node.Status(ctx, host)
		if err != nil {
			cli.LogFatal(logger, "failed to get node status endpoint", err)
		}

		switch resp.StatusCode() {
		case http.StatusOK:
			if jsonOutput {
				fmt.Println(string(resp.Body))
				return
			}

			if resp.JSON200 == nil {
				cli.LogFatal(logger, "failed response", fmt.Errorf("node status response was nil"))
			}

			if resp.JSON200.JobId != nil {
				fmt.Println()
				cli.PrintKV("Job ID", resp.JSON200.JobId.String())
			}

			displayNodeStatusCollection(host, resp.JSON200)

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

// displayNodeStatusCollection renders node status results.
// For a single non-broadcast result, shows detailed output; otherwise shows a summary table.
func displayNodeStatusCollection(
	target string,
	data *gen.NodeStatusCollectionResponse,
) {
	if len(data.Results) == 1 && target != "_all" {
		displayNodeStatusDetail(&data.Results[0])
		return
	}

	fmt.Println()

	results := make([]cli.ResultRow, 0, len(data.Results))
	for _, s := range data.Results {
		uptime := ""
		if s.Uptime != nil {
			uptime = *s.Uptime
		}
		load := ""
		if s.LoadAverage != nil {
			load = fmt.Sprintf("%.2f", s.LoadAverage.N1min)
		}
		memory := ""
		if s.Memory != nil {
			memory = fmt.Sprintf(
				"%d GB / %d GB",
				s.Memory.Used/1024/1024/1024,
				s.Memory.Total/1024/1024/1024,
			)
		}
		results = append(results, cli.ResultRow{
			Hostname: s.Hostname,
			Error:    s.Error,
			Fields:   []string{uptime, load, memory},
		})
	}
	headers, rows := cli.BuildBroadcastTable(results, []string{
		"UPTIME",
		"LOAD (1m)",
		"MEMORY USED",
	})
	cli.PrintStyledTable([]cli.Section{{Headers: headers, Rows: rows}})
}

// displayNodeStatusDetail renders a single node status response with full details.
func displayNodeStatusDetail(
	data *gen.NodeStatusResponse,
) {
	fmt.Println()

	kvArgs := []string{"Hostname", data.Hostname}
	if data.OsInfo != nil {
		kvArgs = append(
			kvArgs,
			"OS",
			data.OsInfo.Distribution+" "+cli.DimStyle.Render(data.OsInfo.Version),
		)
	}
	cli.PrintKV(kvArgs...)

	if data.LoadAverage != nil {
		cli.PrintKV("Load", fmt.Sprintf("%.2f, %.2f, %.2f",
			data.LoadAverage.N1min, data.LoadAverage.N5min, data.LoadAverage.N15min,
		)+" "+cli.DimStyle.Render("(1m, 5m, 15m)"))
	}

	if data.Memory != nil {
		cli.PrintKV("Memory", fmt.Sprintf("%d GB used / %d GB total / %d GB free",
			data.Memory.Used/1024/1024/1024,
			data.Memory.Total/1024/1024/1024,
			data.Memory.Free/1024/1024/1024,
		))
	}

	diskRows := [][]string{}

	if data.Disks != nil {
		for _, disk := range *data.Disks {
			diskRows = append(diskRows, []string{
				disk.Name,
				fmt.Sprintf("%d GB", disk.Total/1024/1024/1024),
				fmt.Sprintf("%d GB", disk.Used/1024/1024/1024),
				fmt.Sprintf("%d GB", disk.Free/1024/1024/1024),
			})
		}
	}

	sections := []cli.Section{
		{
			Title:   "Disks",
			Headers: []string{"DISK NAME", "TOTAL", "USED", "FREE"},
			Rows:    diskRows,
		},
	}
	cli.PrintStyledTable(sections)
}

func init() {
	clientNodeCmd.AddCommand(clientNodeStatusGetCmd)
}

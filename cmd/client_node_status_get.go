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

	"github.com/retr0h/osapi/pkg/sdk/client"
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
		resp, err := sdkClient.Status.Get(ctx, host)
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

		displayNodeStatusCollection(host, &resp.Data)
	},
}

// displayNodeStatusCollection renders node status results.
// For a single non-broadcast result, shows detailed output; otherwise shows a summary table.
func displayNodeStatusCollection(
	target string,
	data *client.Collection[client.NodeStatus],
) {
	if len(data.Results) == 1 && target != "_all" {
		displayNodeStatusDetail(&data.Results[0])
		return
	}

	fmt.Println()

	results := make([]cli.ResultRow, 0, len(data.Results))
	for _, s := range data.Results {
		load := ""
		if s.LoadAverage != nil {
			load = fmt.Sprintf("%.2f", s.LoadAverage.OneMin)
		}
		memory := ""
		if s.Memory != nil {
			memory = fmt.Sprintf(
				"%d GB / %d GB",
				s.Memory.Used/1024/1024/1024,
				s.Memory.Total/1024/1024/1024,
			)
		}
		var errPtr *string
		if s.Error != "" {
			errPtr = &s.Error
		}
		results = append(results, cli.ResultRow{
			Hostname: s.Hostname,
			Status:   s.Status,
			Error:    errPtr,
			Fields:   []string{s.Uptime, load, memory},
		})
	}
	tr := cli.BuildBroadcastTable(results, []string{
		"UPTIME",
		"LOAD",
		"MEM",
	})
	cli.PrintCompactTable([]cli.Section{{Headers: tr.Headers, Rows: tr.Rows, Errors: tr.Errors}})
}

// displayNodeStatusDetail renders a single node status response with full details.
func displayNodeStatusDetail(
	data *client.NodeStatus,
) {
	fmt.Println()

	kvArgs := []string{"Hostname", data.Hostname}
	if data.OSInfo != nil {
		kvArgs = append(
			kvArgs,
			"OS",
			data.OSInfo.Distribution+" "+cli.DimStyle.Render(data.OSInfo.Version),
		)
	}
	cli.PrintKV(kvArgs...)

	if data.LoadAverage != nil {
		cli.PrintKV("Load", fmt.Sprintf(
			"%.2f, %.2f, %.2f",
			data.LoadAverage.OneMin, data.LoadAverage.FiveMin, data.LoadAverage.FifteenMin,
		)+" "+cli.DimStyle.Render("(1m, 5m, 15m)"))
	}

	if data.Memory != nil {
		cli.PrintKV("Memory", fmt.Sprintf(
			"%d GB used / %d GB total / %d GB free",
			data.Memory.Used/1024/1024/1024,
			data.Memory.Total/1024/1024/1024,
			data.Memory.Free/1024/1024/1024,
		))
	}

	diskRows := make([][]string, 0, len(data.Disks))
	for _, disk := range data.Disks {
		diskRows = append(diskRows, []string{
			disk.Name,
			fmt.Sprintf("%d GB", disk.Total/1024/1024/1024),
			fmt.Sprintf("%d GB", disk.Used/1024/1024/1024),
			fmt.Sprintf("%d GB", disk.Free/1024/1024/1024),
		})
	}

	sections := []cli.Section{
		{
			Title:   "Disks",
			Headers: []string{"DISK NAME", "TOTAL", "USED", "FREE"},
			Rows:    diskRows,
		},
	}
	cli.PrintCompactTable(sections)
}

func init() {
	clientNodeCmd.AddCommand(clientNodeStatusGetCmd)
}

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
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/client"
	"github.com/retr0h/osapi/internal/client/gen"
)

// clientSystemStatusGetCmd represents the clientSystemStatusGet command.
var clientSystemStatusGetCmd = &cobra.Command{
	Use:   "status",
	Short: "Status of the server",
	Long: `Obtain the current system status.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		systemHandler := handler.(client.SystemHandler)
		resp, err := systemHandler.GetSystemStatus(ctx, host)
		if err != nil {
			logFatal("failed to get system status endpoint", err)
		}

		switch resp.StatusCode() {
		case http.StatusOK:
			if jsonOutput {
				fmt.Println(string(resp.Body))
				return
			}

			if host == "_all" {
				displayMultiSystemStatus(resp.Body)
				return
			}

			if resp.JSON200 == nil {
				logFatal("failed response", fmt.Errorf("system data response was nil"))
			}

			displaySingleSystemStatus(resp.JSON200)

		case http.StatusUnauthorized:
			handleAuthError(resp.JSON401, resp.StatusCode(), logger)
		case http.StatusForbidden:
			handleAuthError(resp.JSON403, resp.StatusCode(), logger)
		default:
			handleUnknownError(resp.JSON500, resp.StatusCode(), logger)
		}
	},
}

// displaySingleSystemStatus renders a single system status response.
func displaySingleSystemStatus(
	data *gen.SystemStatusResponse,
) {
	systemData := map[string]interface{}{
		"Load Average (1m, 5m, 15m)": fmt.Sprintf(
			"%.2f, %.2f, %.2f",
			data.LoadAverage.N1min,
			data.LoadAverage.N5min,
			data.LoadAverage.N15min,
		),
		"Memory": fmt.Sprintf(
			"%d GB used / %d GB total / %d GB free",
			data.Memory.Used/1024/1024/1024,
			data.Memory.Total/1024/1024/1024,
			data.Memory.Free/1024/1024/1024,
		),
		"OS": fmt.Sprintf(
			"%s %s",
			data.OsInfo.Distribution,
			data.OsInfo.Version,
		),
	}
	printStyledMap(systemData)

	diskRows := [][]string{}

	if data.Disks != nil {
		for _, disk := range data.Disks {
			diskRows = append(diskRows, []string{
				disk.Name,
				fmt.Sprintf("%d GB", disk.Total/1024/1024/1024),
				fmt.Sprintf("%d GB", disk.Used/1024/1024/1024),
				fmt.Sprintf("%d GB", disk.Free/1024/1024/1024),
			})
		}
	}

	sections := []section{
		{
			Title:   "Disks",
			Headers: []string{"DISK NAME", "TOTAL", "USED", "FREE"},
			Rows:    diskRows,
		},
	}
	printStyledTable(sections)
}

// displayMultiSystemStatus renders multiple system status responses from _all broadcast.
func displayMultiSystemStatus(
	body []byte,
) {
	var multiResp struct {
		Results []gen.SystemStatusResponse `json:"results"`
	}
	if err := json.Unmarshal(body, &multiResp); err != nil {
		logFatal("failed to parse multi-host response", err)
	}

	rows := make([][]string, 0, len(multiResp.Results))
	for _, s := range multiResp.Results {
		rows = append(rows, []string{
			s.Hostname,
			s.Uptime,
			fmt.Sprintf("%.2f", s.LoadAverage.N1min),
			fmt.Sprintf(
				"%d GB / %d GB",
				s.Memory.Used/1024/1024/1024,
				s.Memory.Total/1024/1024/1024,
			),
		})
	}

	sections := []section{
		{
			Headers: []string{"HOSTNAME", "UPTIME", "LOAD (1m)", "MEMORY USED"},
			Rows:    rows,
		},
	}
	printStyledTable(sections)
}

func init() {
	clientSystemCmd.AddCommand(clientSystemStatusGetCmd)
}

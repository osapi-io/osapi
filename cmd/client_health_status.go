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
	"strconv"

	"github.com/osapi-io/osapi-sdk/pkg/osapi/gen"
	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
)

// clientHealthStatusCmd represents the clientHealthStatus command.
var clientHealthStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "System status and component health",
	Long: `Show per-component health status with system metrics.
Requires authentication.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		resp, err := sdkClient.Health.Status(ctx)
		if err != nil {
			cli.LogFatal(logger, "failed to get health status endpoint", err)
		}

		switch resp.StatusCode() {
		case http.StatusOK:
			if jsonOutput {
				fmt.Println(string(resp.Body))
				return
			}

			if resp.JSON200 == nil {
				cli.LogFatal(
					logger,
					"failed response",
					fmt.Errorf("health status response was nil"),
				)
			}

			displayStatusHealth(resp.JSON200)

		case http.StatusServiceUnavailable:
			if jsonOutput {
				fmt.Println(string(resp.Body))
				return
			}

			if resp.JSON503 == nil {
				cli.LogFatal(
					logger,
					"failed response",
					fmt.Errorf("health status response was nil"),
				)
			}

			displayStatusHealth(resp.JSON503)

		case http.StatusUnauthorized:
			cli.HandleAuthError(resp.JSON401, resp.StatusCode(), logger)
		case http.StatusForbidden:
			cli.HandleAuthError(resp.JSON403, resp.StatusCode(), logger)
		default:
			cli.HandleUnknownError(nil, resp.StatusCode(), logger)
		}
	},
}

// displayStatusHealth renders health status output with system metrics.
func displayStatusHealth(
	data *gen.StatusResponse,
) {
	fmt.Println()
	cli.PrintKV("Status", data.Status, "Version", data.Version, "Uptime", data.Uptime)

	if data.Nats != nil {
		natsVal := data.Nats.Url
		if data.Nats.Version != "" {
			natsVal += " " + cli.DimStyle.Render("(v"+data.Nats.Version+")")
		}
		cli.PrintKV("NATS", natsVal)
	}

	if data.Jobs != nil {
		cli.PrintKV("Jobs", fmt.Sprintf(
			"%d total, %d completed, %d unprocessed, %d failed, %d dlq",
			data.Jobs.Total, data.Jobs.Completed,
			data.Jobs.Unprocessed, data.Jobs.Failed, data.Jobs.Dlq,
		))
	}

	// Tables only for genuinely multi-row data
	var sections []cli.Section

	componentRows := make([][]string, 0, len(data.Components))
	for name, component := range data.Components {
		errMsg := ""
		if component.Error != nil {
			errMsg = *component.Error
		}
		componentRows = append(componentRows, []string{name, component.Status, errMsg})
	}
	sections = append(sections, cli.Section{
		Title:   "Components",
		Headers: []string{"COMPONENT", "STATUS", "ERROR"},
		Rows:    componentRows,
	})

	if data.Streams != nil && len(*data.Streams) > 0 {
		streamRows := make([][]string, 0, len(*data.Streams))
		for _, s := range *data.Streams {
			streamRows = append(streamRows, []string{
				s.Name,
				strconv.Itoa(s.Messages),
				strconv.Itoa(s.Bytes),
				strconv.Itoa(s.Consumers),
			})
		}
		sections = append(sections, cli.Section{
			Title:   "Streams",
			Headers: []string{"NAME", "MESSAGES", "BYTES", "CONSUMERS"},
			Rows:    streamRows,
		})
	}

	if data.KvBuckets != nil && len(*data.KvBuckets) > 0 {
		kvRows := make([][]string, 0, len(*data.KvBuckets))
		for _, b := range *data.KvBuckets {
			kvRows = append(kvRows, []string{
				b.Name,
				strconv.Itoa(b.Keys),
				strconv.Itoa(b.Bytes),
			})
		}
		sections = append(sections, cli.Section{
			Title:   "KV Buckets",
			Headers: []string{"NAME", "KEYS", "BYTES"},
			Rows:    kvRows,
		})
	}

	cli.PrintStyledTable(sections)
}

func init() {
	clientHealthCmd.AddCommand(clientHealthStatusCmd)
}

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

	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/client"
	"github.com/retr0h/osapi/internal/client/gen"
)

// clientHealthDetailedCmd represents the clientHealthDetailed command.
var clientHealthDetailedCmd = &cobra.Command{
	Use:   "detailed",
	Short: "Detailed component health",
	Long: `Show per-component health status with version and uptime.
Requires authentication.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		healthHandler := handler.(client.HealthHandler)
		resp, err := healthHandler.GetHealthDetailed(ctx)
		if err != nil {
			logFatal("failed to get health detailed endpoint", err)
		}

		switch resp.StatusCode() {
		case http.StatusOK:
			if jsonOutput {
				fmt.Println(string(resp.Body))
				return
			}

			if resp.JSON200 == nil {
				logFatal("failed response", fmt.Errorf("health detailed response was nil"))
			}

			displayDetailedHealth(resp.JSON200)

		case http.StatusServiceUnavailable:
			if jsonOutput {
				fmt.Println(string(resp.Body))
				return
			}

			if resp.JSON503 == nil {
				logFatal("failed response", fmt.Errorf("health detailed response was nil"))
			}

			displayDetailedHealth(resp.JSON503)

		case http.StatusUnauthorized:
			handleAuthError(resp.JSON401, resp.StatusCode(), logger)
		case http.StatusForbidden:
			handleAuthError(resp.JSON403, resp.StatusCode(), logger)
		default:
			handleUnknownError(nil, resp.StatusCode(), logger)
		}
	},
}

// displayDetailedHealth renders detailed health check output.
func displayDetailedHealth(
	data *gen.DetailedHealthResponse,
) {
	overview := map[string]interface{}{
		"Status":  data.Status,
		"Version": data.Version,
		"Uptime":  data.Uptime,
	}
	printStyledMap(overview)

	rows := make([][]string, 0, len(data.Components))
	for name, component := range data.Components {
		errMsg := ""
		if component.Error != nil {
			errMsg = *component.Error
		}
		rows = append(rows, []string{name, component.Status, errMsg})
	}

	sections := []section{
		{
			Title:   "Components",
			Headers: []string{"COMPONENT", "STATUS", "ERROR"},
			Rows:    rows,
		},
	}
	printStyledTable(sections)
}

func init() {
	clientHealthCmd.AddCommand(clientHealthDetailedCmd)
}

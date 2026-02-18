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
)

// clientJobWorkersListCmd represents the clientJobWorkersList command.
var clientJobWorkersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List active workers",
	Long: `Discover all active workers by broadcasting a hostname query
and collecting responses. Shows each worker's hostname.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()

		jobHandler := handler.(client.JobHandler)
		resp, err := jobHandler.GetJobWorkers(ctx)
		if err != nil {
			logFatal("failed to list workers", err)
		}

		switch resp.StatusCode() {
		case http.StatusOK:
			if jsonOutput {
				fmt.Println(string(resp.Body))
				return
			}

			if resp.JSON200 == nil {
				logFatal("failed response", fmt.Errorf("workers response was nil"))
			}

			workers := resp.JSON200.Workers
			if len(workers) == 0 {
				fmt.Println("No active workers found.")
				return
			}

			rows := make([][]string, 0, len(workers))
			for _, w := range workers {
				rows = append(rows, []string{w.Hostname})
			}

			sections := []section{
				{
					Title:   fmt.Sprintf("Active Workers (%d)", resp.JSON200.Total),
					Headers: []string{"HOSTNAME"},
					Rows:    rows,
				},
			}
			printStyledTable(sections)
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
	clientJobWorkersCmd.AddCommand(clientJobWorkersListCmd)
}

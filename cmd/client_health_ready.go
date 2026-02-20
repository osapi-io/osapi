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

// clientHealthReadyCmd represents the clientHealthReady command.
var clientHealthReadyCmd = &cobra.Command{
	Use:   "ready",
	Short: "Readiness probe",
	Long: `Check if the API server is ready to accept traffic.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		healthHandler := handler.(client.HealthHandler)
		resp, err := healthHandler.GetHealthReady(ctx)
		if err != nil {
			logFatal("failed to get health ready endpoint", err)
		}

		switch resp.StatusCode() {
		case http.StatusOK:
			if jsonOutput {
				fmt.Println(string(resp.Body))
				return
			}

			if resp.JSON200 == nil {
				logFatal("failed response", fmt.Errorf("health ready response was nil"))
			}

			fmt.Println()
			printKV("Status", resp.JSON200.Status)

		case http.StatusServiceUnavailable:
			if jsonOutput {
				fmt.Println(string(resp.Body))
				return
			}

			if resp.JSON503 == nil {
				logFatal("failed response", fmt.Errorf("health ready response was nil"))
			}

			fmt.Println()
			printKV("Status", resp.JSON503.Status)
			if resp.JSON503.Error != nil {
				printKV("Error", *resp.JSON503.Error)
			}

		default:
			handleUnknownError(nil, resp.StatusCode(), logger)
		}
	},
}

func init() {
	clientHealthCmd.AddCommand(clientHealthReadyCmd)
}

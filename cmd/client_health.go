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

	"github.com/retr0h/osapi/internal/cli"
)

// clientHealthCmd represents the clientHealth command.
var clientHealthCmd = &cobra.Command{
	Use:   "health",
	Short: "Health check endpoints",
	Long: `Check the health of the API server.

Running without a subcommand performs a liveness probe.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		resp, err := sdkClient.Health.Liveness(ctx)
		if err != nil {
			cli.LogFatal(logger, "failed to get health endpoint", err)
		}

		switch resp.StatusCode() {
		case http.StatusOK:
			if jsonOutput {
				fmt.Println(string(resp.Body))
				return
			}

			if resp.JSON200 == nil {
				cli.LogFatal(logger, "failed response", fmt.Errorf("health response was nil"))
			}

			fmt.Println()
			cli.PrintKV("Status", resp.JSON200.Status)

		default:
			cli.HandleUnknownError(nil, resp.StatusCode(), logger)
		}
	},
}

func init() {
	clientCmd.AddCommand(clientHealthCmd)
}

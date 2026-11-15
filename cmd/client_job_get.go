// Copyright (c) 2025 John Dewey

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
	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/cli"
)

// clientJobGetCmd represents the clientJobsGet command.
var clientJobGetCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get job details and status",
	Example: `  osapi client job get --id 550e8400-e29b-41d4-a716-446655440000`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		jobID, _ := cmd.Flags().GetString("id")

		resp, err := sdkClient.Job.Get(ctx, jobID)
		if err != nil {
			cli.HandleError(err, logger)
			return
		}

		if jsonOutput {
			cli.PrintJSON(resp.RawJSON())
			return
		}

		cli.DisplayJobDetail(&resp.Data)
	},
}

func init() {
	clientJobCmd.AddCommand(clientJobGetCmd)

	clientJobGetCmd.PersistentFlags().
		StringP("id", "", "", "Job ID to retrieve")

	_ = clientJobGetCmd.MarkPersistentFlagRequired("id")
}

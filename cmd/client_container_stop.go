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

	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
	"github.com/retr0h/osapi/pkg/sdk/client/gen"
)

// clientContainerStopCmd represents the clientContainerStop command.
var clientContainerStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a running container",
	Long:  `Stop a running container on the target node.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		id, _ := cmd.Flags().GetString("id")
		timeout, _ := cmd.Flags().GetInt("timeout")

		body := gen.ContainerStopRequest{}
		if cmd.Flags().Changed("timeout") {
			body.Timeout = &timeout
		}

		resp, err := sdkClient.Container.Stop(ctx, host, id, body)
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
			fmt.Println()
		}

		for _, r := range resp.Data.Results {
			cli.PrintKV("Hostname", r.Hostname)
			if r.Error != "" {
				cli.PrintKV("Error", r.Error)
				continue
			}
			cli.PrintKV("Message", r.Message)
		}
	},
}

func init() {
	clientContainerCmd.AddCommand(clientContainerStopCmd)

	clientContainerStopCmd.PersistentFlags().
		String("id", "", "Container ID to stop (required)")
	clientContainerStopCmd.PersistentFlags().
		Int("timeout", 10, "Seconds to wait before killing the container")

	_ = clientContainerStopCmd.MarkPersistentFlagRequired("id")
}

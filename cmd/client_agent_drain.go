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
)

// clientAgentDrainCmd represents the clientAgentDrain command.
var clientAgentDrainCmd = &cobra.Command{
	Use:   "drain",
	Short: "Drain an agent",
	Long:  `Stop an agent from accepting new jobs. In-flight jobs continue to completion.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		hostname, _ := cmd.Flags().GetString("hostname")

		resp, err := sdkClient.Agent.Drain(ctx, hostname)
		if err != nil {
			cli.HandleError(err, logger)
			return
		}

		if jsonOutput {
			fmt.Println(string(resp.RawJSON()))
			return
		}

		fmt.Println()
		cli.PrintKV("Hostname", hostname, "Status", "Draining")
		cli.PrintKV("Message", resp.Data.Message)
	},
}

func init() {
	clientAgentCmd.AddCommand(clientAgentDrainCmd)
	clientAgentDrainCmd.Flags().String("hostname", "", "Hostname of the agent to drain")
	_ = clientAgentDrainCmd.MarkFlagRequired("hostname")
}

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

	"github.com/osapi-io/osapi/internal/cli"
)

// clientAgentRemoveCmd represents the clientAgentRemove command.
var clientAgentRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove an enrolled agent",
	Long: `Remove an accepted agent's stored key. Its responses and registration stop
verifying at once, so targeting no longer resolves it.

The agent keeps the controller key it was given and does not enroll again by
itself. To enroll it again, delete controller.pub from its key directory on the
host and restart it.

Two machines may share a hostname. When they do, the command fails and names
their machine IDs; run it again with --machine-id.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		hostname, _ := cmd.Flags().GetString("hostname")
		machineID, _ := cmd.Flags().GetString("machine-id")

		resp, err := sdkClient.Agent.Remove(ctx, hostname, machineID)
		if err != nil {
			cli.HandleError(err, logger)
			return
		}

		if jsonOutput {
			cli.PrintJSON(resp.RawJSON())
			return
		}

		fmt.Println()
		cli.PrintKV("Status", "Removed", "Message", resp.Data.Message)
	},
}

func init() {
	clientAgentCmd.AddCommand(clientAgentRemoveCmd)
	clientAgentRemoveCmd.Flags().String("hostname", "", "Hostname of the accepted agent to remove")
	clientAgentRemoveCmd.Flags().
		String("machine-id", "", "Machine ID to remove, when more than one agent has the hostname")
	_ = clientAgentRemoveCmd.MarkFlagRequired("hostname")
}

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
	Long: "When PKI is enabled, the controller keeps the public key of each accepted\n" +
		"agent. This command deletes it. After that, the controller does not verify\n" +
		"the responses or the registration of the agent. Thus a target does not find\n" +
		"the agent, and a broadcast gives it no row. The agent does not get a\n" +
		"message. It keeps the controller key that it got at acceptance, and it does\n" +
		"not enroll again by itself. To enroll it again, delete `controller.pub` from\n" +
		"its key directory on the host (`/etc/osapi/pki` by default), and restart the\n" +
		"agent. Then it sends a new request, and `agent accept` accepts it. Two\n" +
		"machines can have the same hostname, because the identity of an agent is its\n" +
		"machine ID. When more than one accepted agent has the hostname, the command\n" +
		"fails and names their machine IDs.",
	Example: `  osapi client agent remove --hostname web-01
  osapi client agent remove --hostname web-01 --machine-id 3f2a9c...`,
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

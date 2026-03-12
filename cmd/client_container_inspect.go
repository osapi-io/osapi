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
	"strings"

	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
)

// clientContainerInspectCmd represents the clientContainerInspect command.
var clientContainerInspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect a container",
	Long:  `Retrieve detailed information about a specific container on the target node.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		id, _ := cmd.Flags().GetString("id")

		resp, err := sdkClient.Container.Inspect(ctx, host, id)
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
			cli.PrintKV("ID", r.ID, "Name", r.Name)
			cli.PrintKV("Image", r.Image, "State", r.State)
			cli.PrintKV("Created", r.Created)
			if r.Health != "" {
				cli.PrintKV("Health", r.Health)
			}
			cli.PrintKV("Ports", cli.FormatList(r.Ports))
			cli.PrintKV("Mounts", cli.FormatList(r.Mounts))

			if len(r.NetworkSettings) > 0 {
				ip := r.NetworkSettings["ip_address"]
				gateway := r.NetworkSettings["gateway"]
				if ip != "" {
					cli.PrintKV("Network IP", ip)
				}
				if gateway != "" {
					cli.PrintKV("Network Gateway", gateway)
				}

				// Display any other network settings
				other := make([]string, 0)
				for k, v := range r.NetworkSettings {
					if k != "ip_address" && k != "gateway" && v != "" {
						other = append(other, k+"="+v)
					}
				}
				if len(other) > 0 {
					cli.PrintKV("Network", strings.Join(other, ", "))
				}
			}
		}
	},
}

func init() {
	clientContainerCmd.AddCommand(clientContainerInspectCmd)

	clientContainerInspectCmd.PersistentFlags().
		String("id", "", "Container ID to inspect (required)")

	_ = clientContainerInspectCmd.MarkPersistentFlagRequired("id")
}

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

// clientNodeNetworkInterfaceGetCmd represents the interface get command.
var clientNodeNetworkInterfaceGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a network interface",
	Long:  `Get details of a specific network interface on the target node.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		name, _ := cmd.Flags().GetString("name")

		resp, err := sdkClient.Interface.Get(ctx, host, name)
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
			if r.Error != "" {
				cli.PrintKV("Hostname", r.Hostname)
				cli.PrintKV("Error", r.Error)

				continue
			}

			if r.Interface != nil {
				iface := r.Interface
				cli.PrintKV("Hostname", r.Hostname)
				cli.PrintKV("Name", iface.Name)
				cli.PrintKV("DHCP4", fmt.Sprintf("%t", iface.DHCP4))
				cli.PrintKV("DHCP6", fmt.Sprintf("%t", iface.DHCP6))
				cli.PrintKV("Addresses", cli.FormatList(iface.Addresses))
				cli.PrintKV("Gateway4", iface.Gateway4)
				cli.PrintKV("Gateway6", iface.Gateway6)
				cli.PrintKV("MTU", fmt.Sprintf("%d", iface.MTU))
				cli.PrintKV("MAC Address", iface.MACAddress)
				cli.PrintKV("Wake-on-LAN", fmt.Sprintf("%t", iface.WakeOnLAN))
				cli.PrintKV("Managed", fmt.Sprintf("%t", iface.Managed))
				cli.PrintKV("State", iface.State)
			}
		}
	},
}

func init() {
	clientNodeNetworkInterfaceCmd.AddCommand(clientNodeNetworkInterfaceGetCmd)

	clientNodeNetworkInterfaceGetCmd.PersistentFlags().
		String("name", "", "Interface name (required)")

	_ = clientNodeNetworkInterfaceGetCmd.MarkPersistentFlagRequired("name")
}

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
	"github.com/retr0h/osapi/pkg/sdk/client"
)

// clientNodeNetworkInterfaceCreateCmd represents the interface create command.
var clientNodeNetworkInterfaceCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a network interface configuration",
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		name, _ := cmd.Flags().GetString("name")

		opts := buildInterfaceConfigOpts(cmd)

		resp, err := sdkClient.Interface.Create(ctx, host, name, opts)
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
		}

		results := make([]cli.ResultRow, 0, len(resp.Data.Results))
		for _, r := range resp.Data.Results {
			var errPtr *string
			if r.Error != "" {
				errPtr = &r.Error
			}
			changed := r.Changed
			results = append(results, cli.ResultRow{
				Hostname: r.Hostname,
				Status:   r.Status,
				Changed:  &changed,
				Error:    errPtr,
				Fields:   []string{r.Name},
			})
		}
		tr := cli.BuildMutationTable(results, []string{"NAME"})
		cli.PrintCompactTable([]cli.Section{{Headers: tr.Headers, Rows: tr.Rows, Errors: tr.Errors}})
	},
}

// buildInterfaceConfigOpts reads interface configuration flags from the command.
func buildInterfaceConfigOpts(
	cmd *cobra.Command,
) client.InterfaceConfigOpts {
	opts := client.InterfaceConfigOpts{}

	if cmd.Flags().Changed("dhcp4") {
		v, _ := cmd.Flags().GetBool("dhcp4")
		opts.DHCP4 = &v
	}

	if cmd.Flags().Changed("dhcp6") {
		v, _ := cmd.Flags().GetBool("dhcp6")
		opts.DHCP6 = &v
	}

	if cmd.Flags().Changed("address") {
		opts.Addresses, _ = cmd.Flags().GetStringSlice("address")
	}

	if cmd.Flags().Changed("gateway4") {
		opts.Gateway4, _ = cmd.Flags().GetString("gateway4")
	}

	if cmd.Flags().Changed("gateway6") {
		opts.Gateway6, _ = cmd.Flags().GetString("gateway6")
	}

	if cmd.Flags().Changed("mtu") {
		v, _ := cmd.Flags().GetInt("mtu")
		opts.MTU = &v
	}

	if cmd.Flags().Changed("mac-address") {
		opts.MACAddress, _ = cmd.Flags().GetString("mac-address")
	}

	if cmd.Flags().Changed("wakeonlan") {
		v, _ := cmd.Flags().GetBool("wakeonlan")
		opts.WakeOnLAN = &v
	}

	return opts
}

func init() {
	clientNodeNetworkInterfaceCmd.AddCommand(clientNodeNetworkInterfaceCreateCmd)

	clientNodeNetworkInterfaceCreateCmd.PersistentFlags().
		String("name", "", "Interface name (required)")
	clientNodeNetworkInterfaceCreateCmd.PersistentFlags().
		Bool("dhcp4", false, "Enable DHCPv4")
	clientNodeNetworkInterfaceCreateCmd.PersistentFlags().
		Bool("dhcp6", false, "Enable DHCPv6")
	clientNodeNetworkInterfaceCreateCmd.PersistentFlags().
		StringSlice("address", []string{}, "IP address in CIDR notation (repeatable)")
	clientNodeNetworkInterfaceCreateCmd.PersistentFlags().
		String("gateway4", "", "IPv4 gateway address")
	clientNodeNetworkInterfaceCreateCmd.PersistentFlags().
		String("gateway6", "", "IPv6 gateway address")
	clientNodeNetworkInterfaceCreateCmd.PersistentFlags().
		Int("mtu", 0, "Maximum transmission unit")
	clientNodeNetworkInterfaceCreateCmd.PersistentFlags().
		String("mac-address", "", "Hardware MAC address")
	clientNodeNetworkInterfaceCreateCmd.PersistentFlags().
		Bool("wakeonlan", false, "Enable Wake-on-LAN")

	_ = clientNodeNetworkInterfaceCreateCmd.MarkPersistentFlagRequired("name")
}

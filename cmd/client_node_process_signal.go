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

// clientNodeProcessSignalCmd represents the process signal command.
var clientNodeProcessSignalCmd = &cobra.Command{
	Use:   "signal",
	Short: "Send a signal to a process",
	Long:  `Send a signal to a specific process on the target node.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		pid, _ := cmd.Flags().GetInt("pid")
		signal, _ := cmd.Flags().GetString("signal")

		opts := client.ProcessSignalOpts{
			Signal: signal,
		}

		resp, err := sdkClient.Process.Signal(ctx, host, pid, opts)
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

			results = append(results, cli.ResultRow{
				Hostname: r.Hostname,
				Status:   r.Status,
				Error:    errPtr,
				Fields: []string{
					fmt.Sprintf("%d", r.PID),
					r.Signal,
					fmt.Sprintf("%v", r.Changed),
				},
			})
		}
		headers, rows := cli.BuildBroadcastTable(
			results,
			[]string{"PID", "SIGNAL", "CHANGED"},
		)
		cli.PrintCompactTable([]cli.Section{{Headers: headers, Rows: rows}})
	},
}

func init() {
	clientNodeProcessCmd.AddCommand(clientNodeProcessSignalCmd)

	clientNodeProcessSignalCmd.PersistentFlags().
		Int("pid", 0, "Process ID to signal (required)")
	clientNodeProcessSignalCmd.PersistentFlags().
		String("signal", "", "Signal to send, e.g. TERM, KILL, HUP (required)")

	_ = clientNodeProcessSignalCmd.MarkPersistentFlagRequired("pid")
	_ = clientNodeProcessSignalCmd.MarkPersistentFlagRequired("signal")
}

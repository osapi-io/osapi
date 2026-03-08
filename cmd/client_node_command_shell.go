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
	"os"
	"strconv"

	"github.com/retr0h/osapi/pkg/sdk/client"
	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
)

var clientNodeCommandShellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Execute a shell command",
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		command, _ := cmd.Flags().GetString("command")
		cwd, _ := cmd.Flags().GetString("cwd")
		timeout, _ := cmd.Flags().GetInt("timeout")
		showStdout, _ := cmd.Flags().GetBool("stdout")
		showStderr, _ := cmd.Flags().GetBool("stderr")

		if host == "_all" {
			fmt.Print("This will execute shell command on ALL hosts. Continue? [y/N] ")
			var confirm string
			if _, err := fmt.Scanln(&confirm); err != nil || (confirm != "y" && confirm != "Y") {
				fmt.Println("Aborted.")
				return
			}
		}

		resp, err := sdkClient.Node.Shell(ctx, client.ShellRequest{
			Command: command,
			Cwd:     cwd,
			Timeout: timeout,
			Target:  host,
		})
		if err != nil {
			cli.HandleError(err, logger)
			return
		}

		if jsonOutput {
			fmt.Println(string(resp.RawJSON()))
			return
		}

		if showStdout || showStderr {
			fmt.Println()
			results := buildRawResults(resp.Data.Results)
			cli.PrintRawOutput(os.Stdout, os.Stderr, results, showStdout, showStderr)
			if code := cli.MaxExitCode(results); code != 0 {
				os.Exit(code)
			}
			return
		}

		if resp.Data.JobID != "" {
			fmt.Println()
			cli.PrintKV("Job ID", resp.Data.JobID)
		}

		if len(resp.Data.Results) > 0 {
			results := make([]cli.ResultRow, 0, len(resp.Data.Results))
			for _, r := range resp.Data.Results {
				var errPtr *string
				if r.Error != "" {
					errPtr = &r.Error
				}
				var changedPtr *bool
				if r.Changed {
					changedPtr = &r.Changed
				}
				durationStr := ""
				if r.DurationMs > 0 {
					durationStr = strconv.FormatInt(r.DurationMs, 10) + "ms"
				}
				results = append(results, cli.ResultRow{
					Hostname: r.Hostname,
					Changed:  changedPtr,
					Error:    errPtr,
					Fields: []string{
						r.Stdout,
						r.Stderr,
						strconv.Itoa(r.ExitCode),
						durationStr,
					},
				})
			}
			headers, rows := cli.BuildBroadcastTable(results, []string{
				"STDOUT",
				"STDERR",
				"EXIT CODE",
				"DURATION",
			})
			cli.PrintCompactTable([]cli.Section{{Headers: headers, Rows: rows}})
		}
	},
}

func init() {
	clientNodeCommandCmd.AddCommand(clientNodeCommandShellCmd)

	clientNodeCommandShellCmd.PersistentFlags().
		String("command", "", "The shell command to execute (required)")
	clientNodeCommandShellCmd.PersistentFlags().
		String("cwd", "", "Working directory for the command")
	clientNodeCommandShellCmd.PersistentFlags().
		Int("timeout", 30, "Timeout in seconds (default 30, max 300)")
	clientNodeCommandShellCmd.PersistentFlags().
		Bool("stdout", false, "Print only remote stdout")
	clientNodeCommandShellCmd.PersistentFlags().
		Bool("stderr", false, "Print only remote stderr")

	_ = clientNodeCommandShellCmd.MarkPersistentFlagRequired("command")
}

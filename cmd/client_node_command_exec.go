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

var clientNodeCommandExecCmd = &cobra.Command{
	Use:   "exec",
	Short: "Execute a command directly",
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		command, _ := cmd.Flags().GetString("command")
		args, _ := cmd.Flags().GetStringSlice("args")
		cwd, _ := cmd.Flags().GetString("cwd")
		timeout, _ := cmd.Flags().GetInt("timeout")
		showStdout, _ := cmd.Flags().GetBool("stdout")
		showStderr, _ := cmd.Flags().GetBool("stderr")

		resp, err := sdkClient.Node.Exec(ctx, client.ExecRequest{
			Command: command,
			Args:    args,
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

		results := make([]cli.ResultRow, 0, len(resp.Data.Results))
		for _, r := range resp.Data.Results {
			var errPtr *string
			if r.Error != "" {
				errPtr = &r.Error
			}
			changed := r.Changed
			durationStr := ""
			if r.DurationMs > 0 {
				durationStr = strconv.FormatInt(r.DurationMs, 10) + "ms"
			}
			results = append(results, cli.ResultRow{
				Hostname: r.Hostname,
				Changed:  &changed,
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
	},
}

func buildRawResults(
	items []client.CommandResult,
) []cli.RawResult {
	results := make([]cli.RawResult, 0, len(items))
	for _, r := range items {
		results = append(results, cli.RawResult{
			Hostname: r.Hostname,
			Stdout:   r.Stdout,
			Stderr:   r.Stderr,
			ExitCode: r.ExitCode,
		})
	}
	return results
}

func init() {
	clientNodeCommandCmd.AddCommand(clientNodeCommandExecCmd)

	clientNodeCommandExecCmd.PersistentFlags().
		String("command", "", "The command to execute (required)")
	clientNodeCommandExecCmd.PersistentFlags().
		StringSlice("args", []string{}, "Command arguments (comma-separated)")
	clientNodeCommandExecCmd.PersistentFlags().
		String("cwd", "", "Working directory for the command")
	clientNodeCommandExecCmd.PersistentFlags().
		Int("timeout", 30, "Timeout in seconds (default 30, max 300)")
	clientNodeCommandExecCmd.PersistentFlags().
		Bool("stdout", false, "Print only remote stdout")
	clientNodeCommandExecCmd.PersistentFlags().
		Bool("stderr", false, "Print only remote stderr")

	_ = clientNodeCommandExecCmd.MarkPersistentFlagRequired("command")
}

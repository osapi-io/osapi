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
	"strconv"

	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
	"github.com/retr0h/osapi/pkg/sdk/client"
)

// clientContainerDockerExecCmd represents the clientContainerDockerExec command.
var clientContainerDockerExecCmd = &cobra.Command{
	Use:   "exec",
	Short: "Execute a command in a container",
	Long:  `Execute a command inside a running container on the target node.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		id, _ := cmd.Flags().GetString("id")
		command, _ := cmd.Flags().GetStringSlice("command")
		envFlags, _ := cmd.Flags().GetStringSlice("env")
		workingDir, _ := cmd.Flags().GetString("working-dir")

		opts := client.DockerExecOpts{
			Command:    command,
			Env:        envFlags,
			WorkingDir: workingDir,
		}

		resp, err := sdkClient.Docker.Exec(ctx, host, id, opts)
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

		results := make([]cli.ResultRow, 0)
		for _, r := range resp.Data.Results {
			if r.Error != "" {
				var errPtr *string
				e := r.Error
				errPtr = &e
				results = append(results, cli.ResultRow{
					Hostname: r.Hostname,
					Status:   r.Status,
					Error:    errPtr,
				})

				continue
			}

			results = append(results, cli.ResultRow{
				Hostname: r.Hostname,
				Status:   r.Status,
				Fields: []string{
					strconv.Itoa(r.ExitCode),
					r.Stdout,
					r.Stderr,
				},
			})
		}
		tr := cli.BuildBroadcastTable(
			results,
			[]string{"EXIT CODE", "STDOUT", "STDERR"},
		)
		cli.PrintCompactTable([]cli.Section{{Headers: tr.Headers, Rows: tr.Rows, Errors: tr.Errors}})
	},
}

func init() {
	clientContainerDockerCmd.AddCommand(clientContainerDockerExecCmd)

	clientContainerDockerExecCmd.PersistentFlags().
		String("id", "", "Container ID to exec in (required)")
	clientContainerDockerExecCmd.PersistentFlags().
		StringSlice("command", []string{}, "Command to execute (required, comma-separated)")
	clientContainerDockerExecCmd.PersistentFlags().
		StringSlice("env", []string{}, "Environment variable in KEY=VALUE format (repeatable)")
	clientContainerDockerExecCmd.PersistentFlags().
		String("working-dir", "", "Working directory inside the container")

	_ = clientContainerDockerExecCmd.MarkPersistentFlagRequired("id")
	_ = clientContainerDockerExecCmd.MarkPersistentFlagRequired("command")
}

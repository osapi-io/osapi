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
	"github.com/retr0h/osapi/pkg/sdk/client/gen"
)

// clientContainerExecCmd represents the clientContainerExec command.
var clientContainerExecCmd = &cobra.Command{
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

		body := gen.ContainerExecRequest{
			Command: command,
		}

		if len(envFlags) > 0 {
			body.Env = &envFlags
		}
		if workingDir != "" {
			body.WorkingDir = &workingDir
		}

		resp, err := sdkClient.Container.Exec(ctx, host, id, body)
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

		for _, r := range resp.Data.Results {
			cli.PrintKV("Hostname", r.Hostname)
			if r.Error != "" {
				cli.PrintKV("Error", r.Error)
				continue
			}
			cli.PrintKV("Exit Code", strconv.Itoa(r.ExitCode))
			if r.Stdout != "" {
				cli.PrintKV("Stdout", r.Stdout)
			}
			if r.Stderr != "" {
				cli.PrintKV("Stderr", r.Stderr)
			}
		}
	},
}

func init() {
	clientContainerCmd.AddCommand(clientContainerExecCmd)

	clientContainerExecCmd.PersistentFlags().
		String("id", "", "Container ID to exec in (required)")
	clientContainerExecCmd.PersistentFlags().
		StringSlice("command", []string{}, "Command to execute (required, comma-separated)")
	clientContainerExecCmd.PersistentFlags().
		StringSlice("env", []string{}, "Environment variable in KEY=VALUE format (repeatable)")
	clientContainerExecCmd.PersistentFlags().
		String("working-dir", "", "Working directory inside the container")

	_ = clientContainerExecCmd.MarkPersistentFlagRequired("id")
	_ = clientContainerExecCmd.MarkPersistentFlagRequired("command")
}

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
	"net/http"

	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
	"github.com/retr0h/osapi/internal/client"
)

var clientCommandShellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Execute a shell command",
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		host, _ := cmd.Flags().GetString("target")
		command, _ := cmd.Flags().GetString("command")
		cwd, _ := cmd.Flags().GetString("cwd")
		timeout, _ := cmd.Flags().GetInt("timeout")

		if host == "_all" {
			fmt.Print("This will execute shell command on ALL hosts. Continue? [y/N] ")
			var confirm string
			if _, err := fmt.Scanln(&confirm); err != nil || (confirm != "y" && confirm != "Y") {
				fmt.Println("Aborted.")
				return
			}
		}

		commandHandler := handler.(client.CommandHandler)
		resp, err := commandHandler.PostCommandShell(
			ctx,
			host,
			command,
			cwd,
			timeout,
		)
		if err != nil {
			cli.LogFatal(logger, "failed to execute shell command", err)
		}

		switch resp.StatusCode() {
		case http.StatusAccepted:
			if jsonOutput {
				fmt.Println(string(resp.Body))
				return
			}

			if resp.JSON202 != nil && resp.JSON202.JobId != nil {
				fmt.Println()
				cli.PrintKV("Job ID", resp.JSON202.JobId.String())
			}

			if resp.JSON202 != nil && len(resp.JSON202.Results) > 0 {
				results := make([]cli.ResultRow, 0, len(resp.JSON202.Results))
				for _, r := range resp.JSON202.Results {
					results = append(results, cli.ResultRow{
						Hostname: r.Hostname,
						Error:    r.Error,
						Fields: []string{
							cli.SafeString(r.Stdout),
							cli.SafeString(r.Stderr),
							cli.IntToSafeString(r.ExitCode),
							formatDurationMs(r.DurationMs),
						},
					})
				}
				headers, rows := cli.BuildBroadcastTable(results, []string{
					"STDOUT",
					"STDERR",
					"EXIT CODE",
					"DURATION",
				})
				cli.PrintStyledTable([]cli.Section{{Headers: headers, Rows: rows}})
			}

		case http.StatusBadRequest:
			cli.HandleUnknownError(resp.JSON400, resp.StatusCode(), logger)
		case http.StatusUnauthorized:
			cli.HandleAuthError(resp.JSON401, resp.StatusCode(), logger)
		case http.StatusForbidden:
			cli.HandleAuthError(resp.JSON403, resp.StatusCode(), logger)
		default:
			cli.HandleUnknownError(resp.JSON500, resp.StatusCode(), logger)
		}
	},
}

func init() {
	clientCommandCmd.AddCommand(clientCommandShellCmd)

	clientCommandShellCmd.PersistentFlags().
		String("command", "", "The shell command to execute (required)")
	clientCommandShellCmd.PersistentFlags().
		String("cwd", "", "Working directory for the command")
	clientCommandShellCmd.PersistentFlags().
		Int("timeout", 30, "Timeout in seconds (default 30, max 300)")

	_ = clientCommandShellCmd.MarkPersistentFlagRequired("command")
}

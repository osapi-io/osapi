// Copyright (c) 2025 John Dewey

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
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
)

// clientJobAddCmd represents the clientJobAdd command.
var clientJobAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a job to the queue",
	Long:  `Adds a job to the queue via the REST API for processing.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()

		filePath, _ := cmd.Flags().GetString("json-file")
		targetHostname, _ := cmd.Flags().GetString("target-hostname")

		file, err := os.Open(filePath)
		if err != nil {
			cli.LogFatal(logger, "failed to open file", err)
		}
		defer func() { _ = file.Close() }()

		fileContents, err := io.ReadAll(file)
		if err != nil {
			cli.LogFatal(logger, "failed to read file", err)
		}

		var operationData map[string]interface{}
		if err := json.Unmarshal(fileContents, &operationData); err != nil {
			cli.LogFatal(logger, "failed to parse JSON operation file", err)
		}

		resp, err := sdkClient.Job.Create(ctx, operationData, targetHostname)
		if err != nil {
			cli.LogFatal(logger, "failed to create job", err)
		}

		switch resp.StatusCode() {
		case http.StatusCreated:
			if jsonOutput {
				fmt.Println(string(resp.Body))
				return
			}

			if resp.JSON201 == nil {
				cli.LogFatal(logger, "failed response", fmt.Errorf("create job response was nil"))
			}

			fmt.Println()
			cli.PrintKV("Job ID", resp.JSON201.JobId.String(), "Status", resp.JSON201.Status)
			if resp.JSON201.Revision != nil {
				cli.PrintKV("Revision", fmt.Sprintf("%d", *resp.JSON201.Revision))
			}

			logger.Info("job created successfully",
				slog.String("job_id", resp.JSON201.JobId.String()),
				slog.String("target_hostname", targetHostname),
			)
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
	clientJobCmd.AddCommand(clientJobAddCmd)

	clientJobAddCmd.PersistentFlags().
		StringP("json-file", "", "", "Path to the JSON file containing operation data to create a job")
	clientJobAddCmd.PersistentFlags().
		StringP("target-hostname", "", "", "Target hostname (_any, _all, or specific hostname)")

	_ = clientJobAddCmd.MarkPersistentFlagRequired("json-file")
}

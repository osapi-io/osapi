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
	"log/slog"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/client"
)

// clientJobRetryCmd represents the clientJobRetry command.
var clientJobRetryCmd = &cobra.Command{
	Use:   "retry",
	Short: "Retry a failed or stuck job",
	Long:  `Creates a new job using the same operation data as an existing job via the REST API.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		jobID, _ := cmd.Flags().GetString("job-id")
		targetHostname, _ := cmd.Flags().GetString("target-hostname")

		jobHandler := handler.(client.JobHandler)
		resp, err := jobHandler.RetryJobByID(ctx, jobID, targetHostname)
		if err != nil {
			logFatal("failed to retry job", err)
		}

		switch resp.StatusCode() {
		case http.StatusCreated:
			if jsonOutput {
				fmt.Println(string(resp.Body))
				return
			}

			if resp.JSON201 == nil {
				logFatal("failed response", fmt.Errorf("retry job response was nil"))
			}

			fmt.Println()
			printKV("Job ID", resp.JSON201.JobId.String(), "Status", resp.JSON201.Status)
			if resp.JSON201.Revision != nil {
				printKV("Revision", fmt.Sprintf("%d", *resp.JSON201.Revision))
			}

			logger.Info("job retried successfully",
				slog.String("original_job_id", jobID),
				slog.String("new_job_id", resp.JSON201.JobId.String()),
				slog.String("target_hostname", targetHostname),
			)
		case http.StatusBadRequest:
			handleUnknownError(resp.JSON400, resp.StatusCode(), logger)
		case http.StatusNotFound:
			handleUnknownError(resp.JSON404, resp.StatusCode(), logger)
		case http.StatusUnauthorized:
			handleAuthError(resp.JSON401, resp.StatusCode(), logger)
		case http.StatusForbidden:
			handleAuthError(resp.JSON403, resp.StatusCode(), logger)
		default:
			handleUnknownError(resp.JSON500, resp.StatusCode(), logger)
		}
	},
}

func init() {
	clientJobCmd.AddCommand(clientJobRetryCmd)

	clientJobRetryCmd.PersistentFlags().
		StringP("job-id", "", "", "Job ID to retry")
	clientJobRetryCmd.PersistentFlags().
		StringP("target-hostname", "", "_any", "Override target hostname (_any, _all, or specific hostname)")

	_ = clientJobRetryCmd.MarkPersistentFlagRequired("job-id")
}

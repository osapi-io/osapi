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
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// clientJobRunCmd represents the clientJobRun command.
var clientJobRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Submit a job and wait for completion",
	Long: `Submits a job request and polls for completion, then displays the results.
This combines job submission and retrieval into a single command for convenience.`,
	Run: func(cmd *cobra.Command, _ []string) {
		// Get flags
		timeoutSeconds, _ := cmd.Flags().GetInt("timeout")
		pollSeconds, _ := cmd.Flags().GetInt("poll-interval")
		jsonFilePath, _ := cmd.Flags().GetString("json-file")
		targetHostname, _ := cmd.Flags().GetString("target-hostname")
		if targetHostname == "" {
			targetHostname = "_any"
		}

		// Create context with timeout
		timeout := time.Duration(timeoutSeconds) * time.Second
		ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
		defer cancel()

		// Open and read the JSON file
		file, err := os.Open(jsonFilePath)
		if err != nil {
			logFatal("failed to open file", err)
		}
		defer func() { _ = file.Close() }()

		fileContents, err := io.ReadAll(file)
		if err != nil {
			logFatal("failed to read file", err)
		}

		// Parse the JSON operation
		var operationData map[string]interface{}
		if err := json.Unmarshal(fileContents, &operationData); err != nil {
			logFatal("failed to parse JSON operation file", err)
		}

		// Submit the job
		result, err := jobClient.CreateJob(ctx, operationData, targetHostname)
		if err != nil {
			logger.Error("failed to submit job", slog.String("error", err.Error()))
			return
		}

		jobID := result.JobID
		logger.Debug("job submitted", slog.String("job_id", jobID))

		// Poll for completion
		pollInterval := time.Duration(pollSeconds) * time.Second
		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()

		// Check immediately
		if checkJobComplete(ctx, jobID) {
			return
		}

		// Poll until timeout or completion
		for {
			select {
			case <-ctx.Done():
				logger.Error("job polling timeout",
					slog.String("job_id", jobID),
					slog.String("error", ctx.Err().Error()),
				)
				return
			case <-ticker.C:
				if checkJobComplete(ctx, jobID) {
					return
				}
			}
		}
	},
}

func checkJobComplete(ctx context.Context, jobID string) bool {
	jobInfo, err := jobClient.GetJobStatus(ctx, jobID)
	if err != nil {
		logger.Error("failed to get job status",
			slog.String("job_id", jobID),
			slog.String("error", err.Error()),
		)
		return false
	}

	logger.Debug("job status check",
		slog.String("job_id", jobID),
		slog.String("status", jobInfo.Status),
	)

	// Check if job is complete (including partial_failure)
	if jobInfo.Status == "completed" || jobInfo.Status == "failed" ||
		jobInfo.Status == "partial_failure" {
		logger.Debug("job finished",
			slog.String("job_id", jobID),
			slog.String("status", jobInfo.Status),
		)

		// Display results
		displayJobDetails(jobInfo)
		return true
	}

	return false
}

func init() {
	clientJobCmd.AddCommand(clientJobRunCmd)

	clientJobRunCmd.PersistentFlags().
		StringP("json-file", "", "", "Path to the JSON file containing operation data to create a job")
	clientJobRunCmd.PersistentFlags().
		StringP("target-hostname", "", "", "Target hostname (_any, _all, or specific hostname)")
	clientJobRunCmd.PersistentFlags().
		IntP("timeout", "t", 60, "Timeout in seconds")
	clientJobRunCmd.PersistentFlags().
		IntP("poll-interval", "p", 2, "Poll interval in seconds")

	_ = clientJobRunCmd.MarkPersistentFlagRequired("json-file")
}

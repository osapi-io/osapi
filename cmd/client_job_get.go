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
	"log/slog"

	"github.com/spf13/cobra"
)

// clientJobGetCmd represents the clientJobsGet command.
var clientJobGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get job details and status from KV store",
	Long: `Retrieves a job's details and current status from the NATS KV store.
Shows the job data, status (unprocessed/processing/completed/failed), 
and any results if completed.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		jobID, _ := cmd.Flags().GetString("job-id")

		// Use job client to get job status
		jobInfo, err := jobClient.GetJobStatus(ctx, jobID)
		if err != nil {
			if jsonOutput {
				result := map[string]interface{}{
					"status":  "not_found",
					"job_id":  jobID,
					"message": err.Error(),
				}
				resultJSON, _ := json.Marshal(result)
				logger.Info("job", slog.String("response", string(resultJSON)))
				return
			}

			jobData := map[string]interface{}{
				"Job ID":  jobID,
				"Status":  "Not Found",
				"Message": err.Error(),
			}
			printStyledMap(jobData)
			return
		}

		if jsonOutput {
			resultJSON, _ := json.Marshal(jobInfo)
			logger.Info("job", slog.String("response", string(resultJSON)))
			return
		}

		// Display job details
		jobData := map[string]interface{}{
			"Job ID":  jobInfo.ID,
			"Status":  jobInfo.Status,
			"Created": jobInfo.Created,
		}

		// Add error field if present
		if jobInfo.Error != "" {
			jobData["Error"] = jobInfo.Error
		}

		// Add subject if present
		if jobInfo.Subject != "" {
			jobData["Subject"] = jobInfo.Subject
		}

		printStyledMap(jobData)

		// Collect content for both sections to ensure consistent table widths
		var allContent []string
		var sections []section

		// Display the operation request
		if jobInfo.Operation != nil {
			jobOperationJSON, _ := json.MarshalIndent(jobInfo.Operation, "", "  ")
			operationRows := [][]string{{string(jobOperationJSON)}}
			sections = append(sections, section{
				Title:   "Job Request",
				Headers: []string{"DATA"},
				Rows:    operationRows,
			})
			allContent = append(allContent, string(jobOperationJSON))
		}

		// Display results if completed and available
		if jobInfo.Status == "completed" && len(jobInfo.Result) > 0 {
			var result interface{}
			if err := json.Unmarshal(jobInfo.Result, &result); err == nil {
				resultJSON, _ := json.MarshalIndent(result, "", "  ")
				resultRows := [][]string{{string(resultJSON)}}
				sections = append(sections, section{
					Title:   "Job Response",
					Headers: []string{"DATA"},
					Rows:    resultRows,
				})
				allContent = append(allContent, string(resultJSON))
			}
		}

		// Print each section individually to ensure consistent formatting
		for _, sec := range sections {
			printStyledTable([]section{sec})
		}

		logger.Info("job retrieved successfully",
			slog.String("job_id", jobID),
			slog.String("status", jobInfo.Status),
		)
	},
}

func init() {
	clientJobCmd.AddCommand(clientJobGetCmd)

	clientJobGetCmd.PersistentFlags().
		StringP("job-id", "", "", "Job ID to retrieve")

	_ = clientJobGetCmd.MarkPersistentFlagRequired("job-id")
}

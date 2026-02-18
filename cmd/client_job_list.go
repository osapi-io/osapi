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
	"log/slog"
	"net/http"
	"time"

	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/client"
	"github.com/retr0h/osapi/internal/client/gen"
)

// clientJobListCmd represents the clientJobsList command.
var clientJobListCmd = &cobra.Command{
	Use:   "list",
	Short: "List jobs",
	Long:  `Lists jobs with their current status via the REST API.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		statusFilter, _ := cmd.Flags().GetString("status")
		limitFlag, _ := cmd.Flags().GetInt("limit")
		offsetFlag, _ := cmd.Flags().GetInt("offset")

		jobHandler := handler.(client.JobHandler)

		// Get jobs list
		jobsResp, err := jobHandler.GetJobs(ctx, statusFilter)
		if err != nil {
			logFatal("failed to list jobs", err)
		}

		if jobsResp.StatusCode() != http.StatusOK {
			switch jobsResp.StatusCode() {
			case http.StatusUnauthorized:
				handleAuthError(jobsResp.JSON401, jobsResp.StatusCode(), logger)
			case http.StatusForbidden:
				handleAuthError(jobsResp.JSON403, jobsResp.StatusCode(), logger)
			default:
				handleUnknownError(jobsResp.JSON500, jobsResp.StatusCode(), logger)
			}
			return
		}

		// Get queue stats for summary
		statsResp, err := jobHandler.GetJobQueueStats(ctx)
		if err != nil {
			logFatal("failed to get queue stats", err)
		}

		if statsResp.StatusCode() != http.StatusOK {
			handleUnknownError(statsResp.JSON500, statsResp.StatusCode(), logger)
			return
		}

		// Extract jobs from response
		var jobs []gen.JobDetailResponse
		if jobsResp.JSON200 != nil && jobsResp.JSON200.Items != nil {
			jobs = *jobsResp.JSON200.Items
		}

		// Apply offset
		if offsetFlag > 0 && offsetFlag < len(jobs) {
			jobs = jobs[offsetFlag:]
		} else if offsetFlag >= len(jobs) {
			jobs = []gen.JobDetailResponse{}
		}

		// Apply limit
		if limitFlag > 0 && len(jobs) > limitFlag {
			jobs = jobs[:limitFlag]
		}

		stats := statsResp.JSON200

		if jsonOutput {
			totalJobs := 0
			if stats != nil && stats.TotalJobs != nil {
				totalJobs = *stats.TotalJobs
			}
			statusCounts := map[string]int{}
			if stats != nil && stats.StatusCounts != nil {
				statusCounts = *stats.StatusCounts
			}

			result := map[string]interface{}{
				"total_jobs":     totalJobs,
				"displayed_jobs": len(jobs),
				"status_counts":  statusCounts,
				"filter_applied": statusFilter != "",
				"limit_applied":  limitFlag > 0,
				"offset_applied": offsetFlag,
				"jobs":           jobs,
			}
			resultJSON, _ := json.Marshal(result)
			fmt.Println(string(resultJSON))
			return
		}

		// Display summary
		totalJobs := 0
		statusCounts := map[string]int{}
		if stats != nil {
			if stats.TotalJobs != nil {
				totalJobs = *stats.TotalJobs
			}
			if stats.StatusCounts != nil {
				statusCounts = *stats.StatusCounts
			}
		}

		summaryData := map[string]interface{}{
			"Total Jobs": totalJobs,
			"Submitted":  statusCounts["submitted"],
			"Processing": statusCounts["processing"],
			"Completed":  statusCounts["completed"],
			"Failed":     statusCounts["failed"],
			"Partial":    statusCounts["partial_failure"],
		}

		if statusFilter != "" {
			summaryData["Showing ("+statusFilter+")"] = len(jobs)
		} else {
			summaryData["Showing"] = "All jobs"
		}
		if offsetFlag > 0 {
			summaryData["Skipped"] = fmt.Sprintf("%d jobs", offsetFlag)
		}
		if limitFlag > 0 && len(jobs) >= limitFlag {
			summaryData["Limited to"] = fmt.Sprintf("First %d jobs", limitFlag)
		}

		printStyledMap(summaryData)

		if len(jobs) > 0 {
			jobRows := [][]string{}
			for _, j := range jobs {
				created := safeString(j.Created)
				if t, err := time.Parse(time.RFC3339, created); err == nil {
					created = t.Format("2006-01-02 15:04")
				}

				operationSummary := "Unknown"
				if j.Operation != nil {
					if operationType, ok := (*j.Operation)["type"].(string); ok {
						operationSummary = operationType
					}
				}

				target := safeString(j.Hostname)

				workers := ""
				if j.WorkerStates != nil && len(*j.WorkerStates) > 0 {
					var workerList []string
					for hostname := range *j.WorkerStates {
						workerList = append(workerList, hostname)
					}
					if len(workerList) == 1 {
						workers = workerList[0]
					} else {
						workers = fmt.Sprintf("%d workers", len(workerList))
					}
				}

				jobRows = append(jobRows, []string{
					safeString(j.Id),
					safeString(j.Status),
					created,
					target,
					operationSummary,
					workers,
				})
			}

			sections := []section{
				{
					Title: "Jobs",
					Headers: []string{
						"JOB ID",
						"STATUS",
						"CREATED",
						"TARGET",
						"OPERATION",
						"WORKERS",
					},
					Rows: jobRows,
				},
			}
			printStyledTable(sections)
		}

		logger.Info("jobs listed successfully",
			slog.Int("total", totalJobs),
			slog.Int("displayed", len(jobs)),
			slog.String("status_filter", statusFilter),
			slog.Int("limit", limitFlag),
			slog.Int("offset", offsetFlag),
		)
	},
}

func init() {
	clientJobCmd.AddCommand(clientJobListCmd)

	clientJobListCmd.Flags().
		String("status", "", "Filter jobs by status (submitted, processing, completed, failed, partial_failure)")
	clientJobListCmd.Flags().Int("limit", 10, "Limit number of jobs displayed (0 for no limit)")
	clientJobListCmd.Flags().Int("offset", 0, "Skip the first N jobs (for pagination)")
}

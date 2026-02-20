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
	"net/http"
	"strings"
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

		// Get jobs list (server-side pagination)
		jobsResp, err := jobHandler.GetJobs(ctx, statusFilter, limitFlag, offsetFlag)
		if err != nil {
			logFatal("failed to list jobs", err)
		}

		if jobsResp.StatusCode() != http.StatusOK {
			switch jobsResp.StatusCode() {
			case http.StatusBadRequest:
				handleUnknownError(jobsResp.JSON400, jobsResp.StatusCode(), logger)
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

		// Extract jobs from response (already paginated server-side)
		var jobs []gen.JobDetailResponse
		if jobsResp.JSON200 != nil && jobsResp.JSON200.Items != nil {
			jobs = *jobsResp.JSON200.Items
		}

		totalItems := 0
		if jobsResp.JSON200 != nil && jobsResp.JSON200.TotalItems != nil {
			totalItems = *jobsResp.JSON200.TotalItems
		}

		stats := statsResp.JSON200

		if jsonOutput {
			statusCounts := map[string]int{}
			if stats != nil && stats.StatusCounts != nil {
				statusCounts = *stats.StatusCounts
			}

			result := map[string]interface{}{
				"total_jobs":     totalItems,
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
		statusCounts := map[string]int{}
		if stats != nil {
			if stats.StatusCounts != nil {
				statusCounts = *stats.StatusCounts
			}
		}

		// Summary header
		showing := "All jobs"
		if statusFilter != "" {
			showing = fmt.Sprintf("%s (%d)", statusFilter, totalItems)
		}
		fmt.Println()
		printKV("Total", fmt.Sprintf("%d", totalItems), "Showing", showing)
		printKV(
			"Submitted", fmt.Sprintf("%d", statusCounts["submitted"]),
			"Completed", fmt.Sprintf("%d", statusCounts["completed"]),
			"Failed", fmt.Sprintf("%d", statusCounts["failed"]),
			"Partial", fmt.Sprintf("%d", statusCounts["partial_failure"]),
		)
		if offsetFlag > 0 || (limitFlag > 0 && len(jobs) >= limitFlag) {
			parts := []string{}
			if offsetFlag > 0 {
				parts = append(parts, fmt.Sprintf("offset %d", offsetFlag))
			}
			if limitFlag > 0 && len(jobs) >= limitFlag {
				parts = append(parts, fmt.Sprintf("limit %d", limitFlag))
			}
			printKV("Filter", dimStyle.Render(strings.Join(parts, ", ")))
		}

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

				jobRows = append(jobRows, []string{
					safeString(j.Id),
					safeString(j.Status),
					created,
					target,
					operationSummary,
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
					},
					Rows: jobRows,
				},
			}
			printStyledTable(sections)
		}
	},
}

func init() {
	clientJobCmd.AddCommand(clientJobListCmd)

	clientJobListCmd.Flags().
		String("status", "", "Filter jobs by status (submitted, processing, completed, failed, partial_failure)")
	clientJobListCmd.Flags().Int("limit", 10, "Limit number of jobs displayed (0 for no limit)")
	clientJobListCmd.Flags().Int("offset", 0, "Skip the first N jobs (for pagination)")
}

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
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/job"
)

// extractTargetFromSubject extracts the target hostname from a job subject
// Expected format: job.{type}.{hostname}.{category}.{operation}
func extractTargetFromSubject(subject string) string {
	if subject == "" {
		return "unknown"
	}

	parts := strings.Split(subject, ".")
	if len(parts) < 3 {
		return "unknown"
	}

	// Return the hostname part (3rd element)
	return parts[2]
}

// clientJobListCmd represents the clientJobsList command.
var clientJobListCmd = &cobra.Command{
	Use:   "list",
	Short: "List jobs from KV store",
	Long: `Lists jobs stored in the NATS KV bucket with their current status computed from events.
Shows job IDs, creation time, status, target, operation type, and worker information.
Job status is computed in real-time from append-only status events.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		// Get filter flags
		statusFilter, _ := cmd.Flags().GetString("status")
		limitFlag, _ := cmd.Flags().GetInt("limit")
		offsetFlag, _ := cmd.Flags().GetInt("offset")

		// Use job client to get jobs (this computes status from events)
		jobs, err := jobClient.ListJobs(ctx, statusFilter)
		if err != nil {
			logFatal("failed to list jobs", err)
		}

		// Apply offset if specified
		if offsetFlag > 0 && offsetFlag < len(jobs) {
			jobs = jobs[offsetFlag:]
		} else if offsetFlag >= len(jobs) {
			jobs = []*job.QueuedJob{} // No jobs to show if offset exceeds total
		}

		// Apply limit if specified
		if limitFlag > 0 && len(jobs) > limitFlag {
			jobs = jobs[:limitFlag]
		}

		// Get queue stats for summary
		stats, err := jobClient.GetQueueStats(ctx)
		if err != nil {
			logFatal("failed to get queue stats", err)
		}

		if jsonOutput {
			result := map[string]interface{}{
				"total_jobs":     stats.TotalJobs,
				"displayed_jobs": len(jobs),
				"status_counts":  stats.StatusCounts,
				"filter_applied": statusFilter != "",
				"limit_applied":  limitFlag > 0,
				"offset_applied": offsetFlag,
				"jobs":           jobs,
			}
			resultJSON, _ := json.Marshal(result)
			logger.Info("jobs list", slog.String("response", string(resultJSON)))
			return
		}

		// Display summary (always show total counts)
		summaryData := map[string]interface{}{
			"Total Jobs": stats.TotalJobs,
			"Submitted":  stats.StatusCounts["submitted"],
			"Processing": stats.StatusCounts["processing"],
			"Completed":  stats.StatusCounts["completed"],
			"Failed":     stats.StatusCounts["failed"],
			"Partial":    stats.StatusCounts["partial_failure"],
		}

		// Add filter info if applied
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

		// Display job details
		if len(jobs) > 0 {
			jobRows := [][]string{}
			for _, job := range jobs {
				// Format created time
				created := job.Created
				if t, err := time.Parse(time.RFC3339, job.Created); err == nil {
					created = t.Format("2006-01-02 15:04")
				}

				// Get operation summary
				operationSummary := "Unknown"
				if job.Operation != nil {
					if operationType, ok := job.Operation["type"].(string); ok {
						operationSummary = operationType
					}
				}

				// Get target from subject
				target := extractTargetFromSubject(job.Subject)

				// Get worker info if available
				workers := ""
				if len(job.WorkerStates) > 0 {
					var workerList []string
					for hostname := range job.WorkerStates {
						workerList = append(workerList, hostname)
					}
					if len(workerList) == 1 {
						workers = workerList[0]
					} else {
						workers = fmt.Sprintf("%d workers", len(workerList))
					}
				}

				jobRows = append(jobRows, []string{
					job.ID,
					job.Status,
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
			slog.Int("total", stats.TotalJobs),
			slog.Int("displayed", len(jobs)),
			slog.String("status_filter", statusFilter),
			slog.Int("limit", limitFlag),
			slog.Int("offset", offsetFlag),
		)
	},
}

func init() {
	clientJobCmd.AddCommand(clientJobListCmd)

	// Add filtering flags
	clientJobListCmd.Flags().
		String("status", "", "Filter jobs by status (submitted, processing, completed, failed, partial_failure)")
	clientJobListCmd.Flags().Int("limit", 10, "Limit number of jobs displayed (0 for no limit)")
	clientJobListCmd.Flags().Int("offset", 0, "Skip the first N jobs (for pagination)")
}

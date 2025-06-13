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
	Long: `Lists jobs stored in the NATS KV bucket with their current status.
By default, shows only active jobs (unprocessed and processing) to focus on pending work.
Shows job IDs, creation time, status, and operation type summaries.

Filtering options:
  --status: Show only jobs with specific status (unprocessed, processing, completed, failed)
  --limit: Limit number of jobs displayed (default: 10, use 0 for no limit)

Examples:
  osapi job list                     # Show active jobs (unprocessed/processing)
  osapi job list --status failed     # Show only failed jobs
  osapi job list --status completed  # Show only completed jobs
  osapi job list --limit 0           # Show all matching jobs (no limit)

Performance: Uses efficient prefix-based filtering to avoid loading all jobs into memory.`,
	Run: func(cmd *cobra.Command, _ []string) {
		// Get filter flags
		statusFilter, _ := cmd.Flags().GetString("status")
		limitFlag, _ := cmd.Flags().GetInt("limit")

		var jobs []map[string]interface{}
		statusCounts := map[string]int{
			"unprocessed": 0,
			"processing":  0,
			"completed":   0,
			"failed":      0,
		}

		// Determine which statuses to watch
		var watchPatterns []string
		if statusFilter != "" {
			// Only watch the specific status
			watchPatterns = []string{statusFilter + ".*"}
		} else {
			// Default: only show unprocessed and processing jobs
			watchPatterns = []string{"unprocessed.*", "processing.*"}
		}

		// Use Watch for efficient prefix-based filtering
		for _, pattern := range watchPatterns {
			watcher, err := jobsKV.Watch(pattern)
			if err != nil {
				if err.Error() != "nats: no keys found" {
					logFatal(
						"failed to watch job keys",
						fmt.Errorf("error watching pattern %s: %w", pattern, err),
					)
				}
				continue
			}

			// Process initial values
			for entry := range watcher.Updates() {
				if entry == nil {
					break // End of initial values
				}

				var jobWithStatus map[string]interface{}
				if err := json.Unmarshal(entry.Value(), &jobWithStatus); err != nil {
					continue // Skip jobs we can't parse
				}

				// Count status
				if status, ok := jobWithStatus["status"].(string); ok {
					statusCounts[status]++
				}

				jobs = append(jobs, jobWithStatus)

				// Apply limit if specified (0 means no limit)
				if limitFlag > 0 && len(jobs) >= limitFlag {
					break
				}
			}

			// Stop watcher after getting initial values
			watcher.Stop()

			// Check if we've reached our limit
			if limitFlag > 0 && len(jobs) >= limitFlag {
				break
			}
		}

		// For full statistics when no filter, we need total counts
		// This is a trade-off - we'll count all statuses only if needed
		if statusFilter == "" && !jsonOutput {
			// Count completed and failed for the summary
			for _, status := range []string{"completed", "failed"} {
				watcher, err := jobsKV.Watch(status + ".*")
				if err != nil {
					continue
				}

				count := 0
				for entry := range watcher.Updates() {
					if entry == nil {
						break
					}
					count++
				}
				statusCounts[status] = count
				watcher.Stop()
			}
		}

		if jsonOutput {
			totalJobs := 0
			for _, count := range statusCounts {
				totalJobs += count
			}

			result := map[string]interface{}{
				"total_jobs":     totalJobs,
				"displayed_jobs": len(jobs),
				"status_counts":  statusCounts,
				"filter_applied": statusFilter != "",
				"limit_applied":  limitFlag > 0 && len(jobs) >= limitFlag,
				"jobs":           jobs,
			}
			resultJSON, _ := json.Marshal(result)
			logger.Info("jobs list", slog.String("response", string(resultJSON)))
			return
		}

		// Calculate total for display
		totalJobs := 0
		for _, count := range statusCounts {
			totalJobs += count
		}

		// Display summary (always show total counts)
		summaryData := map[string]interface{}{
			"Total Jobs":  totalJobs,
			"Unprocessed": statusCounts["unprocessed"],
			"Processing":  statusCounts["processing"],
			"Completed":   statusCounts["completed"],
			"Failed":      statusCounts["failed"],
		}

		// Add filter info if applied
		if statusFilter != "" {
			summaryData["Showing ("+statusFilter+")"] = len(jobs)
		} else {
			summaryData["Showing"] = "Active jobs only (use --status to filter)"
		}
		if limitFlag > 0 && len(jobs) >= limitFlag {
			summaryData["Limited to"] = fmt.Sprintf("First %d jobs", limitFlag)
		}

		printStyledMap(summaryData)

		// Display job details
		if len(jobs) > 0 {
			jobRows := [][]string{}
			for _, job := range jobs {
				jobID := "unknown"
				if id, ok := job["id"].(string); ok {
					jobID = id
				}

				status := "unknown"
				if s, ok := job["status"].(string); ok {
					status = s
				}

				created := "unknown"
				if c, ok := job["created"].(string); ok {
					if t, err := time.Parse(time.RFC3339, c); err == nil {
						created = t.Format(time.RFC3339)
					}
				}

				// Try to get operation summary from job
				operationSummary := "Unknown"
				if jobOperationData, ok := job["operation"].(map[string]interface{}); ok {
					if operationType, ok := jobOperationData["type"].(string); ok {
						operationSummary = fmt.Sprintf("Type: %s", operationType)
					}
				}

				// Extract target hostname from subject
				target := "unknown"
				if subject, ok := job["subject"].(string); ok {
					target = extractTargetFromSubject(subject)
				}

				jobRows = append(jobRows, []string{
					jobID,
					status,
					created,
					target,
					operationSummary,
				})
			}

			sections := []section{
				{
					Title:   "Jobs",
					Headers: []string{"JOB ID", "STATUS", "CREATED", "TARGET", "OPERATION"},
					Rows:    jobRows,
				},
			}
			printStyledTable(sections)
		}

		logger.Info("jobs listed successfully",
			slog.Int("total", totalJobs),
			slog.Int("displayed", len(jobs)),
			slog.String("status_filter", statusFilter),
			slog.Int("limit", limitFlag),
			slog.Int("unprocessed", statusCounts["unprocessed"]),
			slog.Int("processing", statusCounts["processing"]),
			slog.Int("completed", statusCounts["completed"]),
			slog.Int("failed", statusCounts["failed"]),
		)
	},
}

func init() {
	clientJobCmd.AddCommand(clientJobListCmd)

	// Add filtering flags
	clientJobListCmd.Flags().
		String("status", "", "Filter jobs by status (unprocessed, processing, completed, failed)")
	clientJobListCmd.Flags().Int("limit", 10, "Limit number of jobs displayed (0 for no limit)")
}

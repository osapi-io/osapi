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
	"strings"
	"time"

	"github.com/osapi-io/osapi-sdk/pkg/osapi"
	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/cli"
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

		// Get jobs list (server-side pagination)
		jobsResp, err := sdkClient.Job.List(ctx, osapi.ListParams{
			Status: statusFilter,
			Limit:  limitFlag,
			Offset: offsetFlag,
		})
		if err != nil {
			cli.HandleError(err, logger)
			return
		}

		jobs := jobsResp.Data.Items
		totalItems := jobsResp.Data.TotalItems
		statusCounts := jobsResp.Data.StatusCounts

		if jsonOutput {
			displayJobListJSON(jobs, totalItems, statusCounts, statusFilter, limitFlag, offsetFlag)
			return
		}

		displayJobListSummary(
			totalItems,
			statusCounts,
			statusFilter,
			limitFlag,
			offsetFlag,
			len(jobs),
		)
		displayJobListTable(jobs)
	},
}

func displayJobListJSON(
	jobs []osapi.JobDetail,
	totalItems int,
	statusCounts map[string]int,
	statusFilter string,
	limitFlag int,
	offsetFlag int,
) {
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
}

func displayJobListSummary(
	totalItems int,
	statusCounts map[string]int,
	statusFilter string,
	limitFlag int,
	offsetFlag int,
	jobCount int,
) {
	showing := "All jobs"
	if statusFilter != "" {
		showing = fmt.Sprintf("%s (%d)", statusFilter, totalItems)
	}
	fmt.Println()
	cli.PrintKV("Total", fmt.Sprintf("%d", totalItems), "Showing", showing)
	cli.PrintKV(
		"Submitted", fmt.Sprintf("%d", statusCounts["submitted"]),
		"Completed", fmt.Sprintf("%d", statusCounts["completed"]),
		"Failed", fmt.Sprintf("%d", statusCounts["failed"]),
		"Partial", fmt.Sprintf("%d", statusCounts["partial_failure"]),
	)
	if offsetFlag > 0 || (limitFlag > 0 && jobCount >= limitFlag) {
		parts := []string{}
		if offsetFlag > 0 {
			parts = append(parts, fmt.Sprintf("offset %d", offsetFlag))
		}
		if limitFlag > 0 && jobCount >= limitFlag {
			parts = append(parts, fmt.Sprintf("limit %d", limitFlag))
		}
		cli.PrintKV("Filter", cli.DimStyle.Render(strings.Join(parts, ", ")))
	}
}

func displayJobListTable(
	jobs []osapi.JobDetail,
) {
	if len(jobs) == 0 {
		return
	}

	jobRows := [][]string{}
	for _, j := range jobs {
		created := j.Created
		if t, err := time.Parse(time.RFC3339, created); err == nil {
			created = t.Format("2006-01-02 15:04")
		}

		operationSummary := "Unknown"
		if j.Operation != nil {
			if operationType, ok := j.Operation["type"].(string); ok {
				operationSummary = operationType
			}
		}

		jobRows = append(jobRows, []string{
			j.ID,
			j.Status,
			created,
			j.Hostname,
			operationSummary,
		})
	}

	sections := []cli.Section{
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
	cli.PrintCompactTable(sections)
}

func init() {
	clientJobCmd.AddCommand(clientJobListCmd)

	clientJobListCmd.Flags().
		String("status", "", "Filter jobs by status (submitted, processing, completed, failed, partial_failure)")
	clientJobListCmd.Flags().Int("limit", 10, "Maximum number of jobs per page (1-100, default 10)")
	clientJobListCmd.Flags().Int("offset", 0, "Skip the first N jobs (for pagination)")
}

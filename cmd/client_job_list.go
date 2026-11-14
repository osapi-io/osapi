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

	"github.com/osapi-io/osapi/pkg/sdk/client"
	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/cli"
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
		jobsResp, err := sdkClient.Job.List(ctx, client.ListParams{
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

		displayJobListTable(jobs, totalItems, statusFilter)
		displayJobListSummary(
			totalItems,
			statusCounts,
			limitFlag,
			offsetFlag,
			len(jobs),
		)
	},
}

func displayJobListJSON(
	jobs []client.JobDetail,
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
	cli.PrintJSON(resultJSON)
}

// displayJobListSummary prints the summary line under the jobs table: the
// total by status, then, in gray, the page the table shows.
func displayJobListSummary(
	totalItems int,
	statusCounts map[string]int,
	limitFlag int,
	offsetFlag int,
	jobCount int,
) {
	noun := "jobs"
	if totalItems == 1 {
		noun = "job"
	}
	line := fmt.Sprintf("%d %s: %s", totalItems, noun, strings.Join([]string{
		cli.Count(statusCounts["completed"], "completed"),
		cli.Count(statusCounts["failed"], "failed"),
		cli.Count(statusCounts["partial_failure"], "partial"),
		cli.Count(statusCounts["submitted"], "submitted"),
	}, ", "))

	var page []string
	if offsetFlag > 0 {
		page = append(page, fmt.Sprintf("offset %d", offsetFlag))
	}
	if limitFlag > 0 && jobCount >= limitFlag {
		page = append(page, fmt.Sprintf("limit %d", limitFlag))
	}
	if len(page) > 0 {
		line += "   " + cli.DimStyle.Render(strings.Join(page, ", "))
	}

	fmt.Println()
	fmt.Println("  " + line)
}

func displayJobListTable(
	jobs []client.JobDetail,
	totalItems int,
	statusFilter string,
) {
	title := fmt.Sprintf("Jobs (%d of %d)", len(jobs), totalItems)
	if statusFilter != "" {
		title = fmt.Sprintf("Jobs %s (%d of %d)", statusFilter, len(jobs), totalItems)
	}

	jobRows := make([][]string, 0, len(jobs))
	for _, j := range jobs {
		created := j.Created
		if t, err := time.Parse(time.RFC3339, created); err == nil {
			created = t.Local().Format("2006-01-02 15:04:05")
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
			Title: title,
			Empty: "None.",
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

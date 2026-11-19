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
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/cli"
	clicommand "github.com/osapi-io/osapi/internal/cli/command"
	"github.com/osapi-io/osapi/internal/controller/endpoints/job"
	"github.com/osapi-io/osapi/pkg/sdk/client"
)

// clientJobCmd represents the clientJob command.
var clientJobCmd = &cobra.Command{
	Use:   "job",
	Short: "Manage the jobs in the queue",
}

func init() {
	clientCmd.AddCommand(clientJobCmd)

	// osapi client job list, get, delete and retry are generated from the
	// job group's definition; this is how each answer prints.
	clicommand.Render(job.Group, "list", renderJobList)
	clicommand.Render(job.Group, "get", renderJob)
	clicommand.Acts(job.Group, "delete", renderJobDeleted)
	clicommand.Render(job.Group, "retry", renderJobRetried)
}

// renderJobList prints a page of jobs as a table, then the count of each
// status and the page the table shows.
func renderJobList(
	_ *cobra.Command,
	in job.ListInput,
	out job.ListJobsResponse,
) {
	title := fmt.Sprintf("Jobs (%d of %d)", len(out.Items), out.TotalItems)
	if in.Status != "" {
		title = fmt.Sprintf("Jobs %s (%d of %d)", in.Status, len(out.Items), out.TotalItems)
	}

	rows := make([][]string, 0, len(out.Items))
	for _, j := range out.Items {
		created := cli.FormatTime(j.Created)

		operation := "Unknown"
		if kind, ok := j.Operation["type"].(string); ok {
			operation = kind
		}

		// The target the request gave; a job stored before targets were kept
		// has only the host that answered.
		target := j.Target
		if target == "" {
			target = j.Hostname
		}
		rows = append(rows, []string{j.ID, j.Status, created, target, operation})
	}

	cli.PrintCompactTable([]cli.Section{
		{
			Title:   title,
			Empty:   "None.",
			Headers: []string{"JOB ID", "STATUS", "CREATED", "TARGET", "OPERATION"},
			Rows:    rows,
		},
	})

	noun := "jobs"
	if out.TotalItems == 1 {
		noun = "job"
	}
	line := fmt.Sprintf("%d %s: %s", out.TotalItems, noun, strings.Join([]string{
		cli.Count(out.StatusCounts["completed"], "completed"),
		cli.Count(out.StatusCounts["failed"], "failed"),
		cli.Count(out.StatusCounts["partial_failure"], "partial"),
		cli.Count(out.StatusCounts["submitted"], "submitted"),
	}, ", "))

	var page []string
	if in.Offset > 0 {
		page = append(page, fmt.Sprintf("offset %d", in.Offset))
	}
	if in.Limit > 0 && len(out.Items) >= in.Limit {
		page = append(page, fmt.Sprintf("limit %d", in.Limit))
	}
	if len(page) > 0 {
		line += "   " + cli.DimStyle.Render(strings.Join(page, ", "))
	}

	fmt.Println()
	fmt.Println("  " + line)
}

// renderJob prints a job with what each agent did with it, as the commands
// that wait for a job do.
func renderJob(
	_ *cobra.Command,
	_ job.GetInput,
	out job.JobDetailResponse,
) {
	// The SDK's JobDetail has the same JSON as the answer.
	raw, _ := json.Marshal(out)
	var detail client.JobDetail
	_ = json.Unmarshal(raw, &detail)

	cli.DisplayJobDetail(&detail)
}

// renderJobDeleted prints the job that went. A delete answers with no
// body, so --json prints what it did.
func renderJobDeleted(
	_ *cobra.Command,
	in job.DeleteInput,
	_ struct{},
) {
	if jsonOutput {
		result, _ := json.Marshal(map[string]any{
			"status":    "deleted",
			"job_id":    in.ID,
			"timestamp": time.Now().Format(time.RFC3339),
		})
		cli.PrintJSON(result)

		return
	}

	cli.PrintHeading("Job")
	cli.PrintKV("Job ID", in.ID, "Status", "Deleted")
	logger.Info("job deleted successfully", slog.String("job_id", in.ID))
}

// renderJobRetried prints the new job that a retry sent.
func renderJobRetried(
	_ *cobra.Command,
	in job.RetryInput,
	out job.CreateJobResponse,
) {
	cli.PrintHeading("Job")
	cli.PrintKV("Job ID", out.JobID, "Status", out.Status)
	if out.Revision != 0 {
		cli.PrintKV("Revision", fmt.Sprintf("%d", out.Revision))
	}
	logger.Info(
		"job retried successfully",
		slog.String("original_job_id", in.ID),
		slog.String("new_job_id", out.JobID),
		slog.String("target_hostname", in.Body.TargetHostname),
	)
}

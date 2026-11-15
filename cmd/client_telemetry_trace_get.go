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
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/cli/reference"
)

// clientTelemetryTraceGetCmd represents the clientTelemetryTraceGet command.
var clientTelemetryTraceGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Show the span tree of a trace",
	Long: "Shows each span of a trace as a tree: the request on the controller,\n" +
		"the job, each agent that ran it with its host and status, and the\n" +
		"commands and file changes under each. The NATS calls, which record the\n" +
		"job's progress, are left out and counted; --all shows them. Each\n" +
		"command prints the ID of its trace after its job ID.\n\n" +
		"The controller reads the trace from the tracing backend at\n" +
		"`telemetry.tracing.url`, so this host needs no access to the backend.\n" +
		"The backend gets the spans a few seconds after the request: a trace\n" +
		"that is not found yet is there after a few seconds.",
	Annotations: map[string]string{
		reference.Permission: "trace:read",
	},
	Example: `  osapi client telemetry trace get --id 26e448c9dbe1c4a8d216a98c675079c9
  osapi client telemetry trace get --id 26e448c9dbe1c4a8d216a98c675079c9 --all
  osapi client telemetry trace get --id 26e448c9dbe1c4a8d216a98c675079c9 --json`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		traceID, _ := cmd.Flags().GetString("id")

		resp, err := sdkClient.Telemetry.Trace(ctx, traceID)
		if err != nil {
			cli.HandleError(err, logger)
			return
		}

		if jsonOutput {
			cli.PrintJSON(resp.RawJSON())
			return
		}

		trace := resp.Data
		start, end := trace.Spans[0].Start, trace.Spans[0].Start
		failed := 0
		for _, s := range trace.Spans {
			if s.Start.Before(start) {
				start = s.Start
			}
			if e := s.Start.Add(s.Duration); e.After(end) {
				end = e
			}
			if s.Failed {
				failed++
			}
		}

		cli.PrintHeading("Trace")
		cli.PrintKV("ID", trace.ID)
		cli.PrintKV("Started", cli.FormatTime(start.Format(time.RFC3339Nano)))
		cli.PrintKV("Duration", cli.FormatDuration(end.Sub(start).String()))
		cli.PrintKV("Services", strings.Join(cli.TraceServices(trace.Spans), ", "))
		if failed > 0 {
			cli.PrintKV("Failed spans", strconv.Itoa(failed))
		}

		all, _ := cmd.Flags().GetBool("all")
		rows := cli.TraceRows(trace.Spans, all)
		cli.PrintCompactTable([]cli.Section{{
			Title:     "Spans (" + strconv.Itoa(len(rows)) + ")",
			Headers:   []string{"SPAN", "HOSTNAME", "DURATION", "STATUS", "DETAIL"},
			Rows:      rows,
			Reference: true,
		}})
		if calls, took := cli.NATSCalls(trace.Spans); !all && calls > 0 {
			cli.PrintNote(fmt.Sprintf(
				"%s left out, %s in all; --all shows them.",
				cli.Plural(calls, "NATS call", "NATS calls"),
				cli.FormatDuration(took.String()),
			))
		}
	},
}

func init() {
	clientTelemetryTraceCmd.AddCommand(clientTelemetryTraceGetCmd)

	clientTelemetryTraceGetCmd.PersistentFlags().
		StringP("id", "", "", "The trace ID, which a command prints after its job ID")
	clientTelemetryTraceGetCmd.PersistentFlags().
		Bool("all", false, "Show the NATS calls too, the reads and writes of the job queue")

	_ = clientTelemetryTraceGetCmd.MarkPersistentFlagRequired("id")
}

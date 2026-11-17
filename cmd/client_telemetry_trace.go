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
	"github.com/osapi-io/osapi/internal/controller/endpoint/command"
	"github.com/osapi-io/osapi/internal/controller/endpoints/telemetry"
	"github.com/osapi-io/osapi/pkg/sdk/client"
)

var clientTelemetryTraceCmd = &cobra.Command{
	Use:   "trace",
	Short: "Read a trace from the tracing backend",
	Annotations: map[string]string{
		reference.Permission: "trace:read",
	},
}

func init() {
	clientTelemetryCmd.AddCommand(clientTelemetryTraceCmd)

	// osapi client telemetry trace get is generated from the telemetry
	// group's definition; this is how its answer prints.
	command.Render(telemetry.Group, "get", renderTrace)
}

// renderTrace prints a trace as its span tree, with the NATS calls left out
// and counted unless --all is given.
func renderTrace(
	cmd *cobra.Command,
	_ telemetry.TraceInput,
	out telemetry.TraceResponse,
) {
	spans := make([]client.TraceSpan, 0, len(out.Spans))
	for _, s := range out.Spans {
		spans = append(spans, client.TraceSpan{
			ID:            s.SpanID,
			ParentID:      s.ParentSpanID,
			Name:          s.Name,
			Service:       s.Service,
			Kind:          s.Kind,
			Start:         s.StartTime,
			Duration:      time.Duration(s.DurationMs * float64(time.Millisecond)),
			Failed:        s.Status == "error",
			StatusMessage: s.StatusMessage,
			Attributes:    s.Attributes,
		})
	}
	if len(spans) == 0 {
		cli.PrintHeading("Trace")
		cli.PrintKV("ID", out.TraceID)
		cli.PrintNote("No spans.")

		return
	}

	start, end := spans[0].Start, spans[0].Start
	failed := 0
	for _, s := range spans {
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
	cli.PrintKV("ID", out.TraceID)
	cli.PrintKV("Started", cli.FormatTime(start.Format(time.RFC3339Nano)))
	cli.PrintKV("Duration", cli.FormatDuration(end.Sub(start).String()))
	cli.PrintKV("Services", strings.Join(cli.TraceServices(spans), ", "))
	if failed > 0 {
		cli.PrintKV("Failed spans", strconv.Itoa(failed))
	}

	all, _ := cmd.Flags().GetBool("all")
	rows := cli.TraceRows(spans, all)
	cli.PrintCompactTable([]cli.Section{{
		Title:     "Spans (" + strconv.Itoa(len(rows)) + ")",
		Headers:   []string{"SPAN", "HOSTNAME", "DURATION", "STATUS", "DETAIL"},
		Rows:      rows,
		Reference: true,
	}})
	if calls, took := cli.NATSCalls(spans); !all && calls > 0 {
		cli.PrintNote(fmt.Sprintf(
			"%s left out, %s in all; --all shows them.",
			cli.Plural(calls, "NATS call", "NATS calls"),
			cli.FormatDuration(took.String()),
		))
	}
}

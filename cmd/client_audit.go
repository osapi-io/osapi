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
	"log/slog"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/osapi-io/osapi/internal/audit/export"
	"github.com/osapi-io/osapi/internal/cli"
	"github.com/osapi-io/osapi/internal/cli/reference"
	"github.com/osapi-io/osapi/internal/controller/endpoint/command"
	"github.com/osapi-io/osapi/internal/controller/endpoints/audit"
	"github.com/osapi-io/osapi/pkg/sdk/client"
)

// clientAuditCmd represents the clientAudit command.
var clientAuditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Read the audit log",
	Annotations: map[string]string{
		reference.Permission: "audit:read",
	},
}

func init() {
	clientCmd.AddCommand(clientAuditCmd)

	// osapi client audit list, get and export are generated from the audit
	// group's definition; this is how each answer prints.
	command.Render(audit.Group, "list", renderAuditList)
	command.Render(audit.Group, "get", renderAuditEntry)
	command.Acts(audit.Group, "export", exportAudit)
}

// renderAuditList prints a page of the audit log as a table.
func renderAuditList(
	_ *cobra.Command,
	_ audit.ListInput,
	out audit.ListAuditResponse,
) {
	rows := make([][]string, 0, len(out.Items))
	for _, entry := range out.Items {
		traceID := ""
		if len(entry.TraceID) > 16 {
			traceID = entry.TraceID[:16] + "…"
		}

		rows = append(rows, []string{
			entry.ID,
			entry.Timestamp.Format("2006-01-02 15:04:05"),
			entry.User,
			entry.Method,
			entry.Path,
			strconv.Itoa(entry.ResponseCode),
			strconv.FormatInt(entry.DurationMs, 10) + "ms",
			traceID,
		})
	}

	cli.PrintCompactTable([]cli.Section{
		{
			Title: fmt.Sprintf("Audit entries (%d of %d)", len(out.Items), out.TotalItems),
			Headers: []string{
				"ID",
				"TIMESTAMP",
				"USER",
				"METHOD",
				"PATH",
				"STATUS",
				"DURATION",
				"TRACE ID",
			},
			Rows: rows,
		},
	})
}

// renderAuditEntry prints one audit entry as its fields.
func renderAuditEntry(
	_ *cobra.Command,
	_ audit.GetInput,
	out audit.AuditEntryResponse,
) {
	entry := out.Entry

	cli.PrintHeading("Audit entry")
	cli.PrintKV("ID", entry.ID)
	cli.PrintKV("Timestamp", entry.Timestamp.Format("2006-01-02 15:04:05"))
	cli.PrintKV("User", entry.User)
	cli.PrintKV("Roles", strings.Join(entry.Roles, ", "))
	cli.PrintKV("Method", entry.Method, "Path", entry.Path)
	cli.PrintKV(
		"Status",
		strconv.Itoa(entry.ResponseCode),
		"Duration",
		strconv.FormatInt(entry.DurationMs, 10)+"ms",
	)
	cli.PrintKV("Source IP", entry.SourceIP)
	if entry.OperationID != "" {
		cli.PrintKV("Operation", entry.OperationID)
	}
	if entry.TraceID != "" {
		cli.PrintKV("Trace ID", entry.TraceID)
	}
}

// exportAudit writes each entry to the --output file as a JSON line, with
// or without --json, and prints how many it wrote.
func exportAudit(
	cmd *cobra.Command,
	_ audit.ExportInput,
	out audit.ListAuditResponse,
) {
	ctx := cmd.Context()
	output, _ := cmd.Flags().GetString("output")
	kind, _ := cmd.Flags().GetString("type")

	var exporter export.Exporter
	switch kind {
	case "file":
		exporter = export.NewFileExporter(output)
	default:
		cli.LogFatal(
			logger,
			"unsupported export type",
			fmt.Errorf("type %q is not supported, use \"file\"", kind),
		)
	}

	if err := exporter.Open(ctx); err != nil {
		cli.LogFatal(logger, "opening exporter", err)
	}

	defer func() {
		if closeErr := exporter.Close(ctx); closeErr != nil {
			logger.Error("closing exporter", slog.Any("error", closeErr))
		}
	}()

	for _, item := range out.Items {
		if err := exporter.Write(ctx, client.AuditEntry{
			ID:           item.ID,
			Timestamp:    item.Timestamp,
			User:         item.User,
			Roles:        item.Roles,
			Method:       item.Method,
			Path:         item.Path,
			ResponseCode: item.ResponseCode,
			DurationMs:   item.DurationMs,
			SourceIP:     item.SourceIP,
			OperationID:  item.OperationID,
			TraceID:      item.TraceID,
		}); err != nil {
			cli.LogFatal(logger, "writing entry", err)
		}
	}

	cli.PrintHeading("Export")
	cli.PrintKV(
		"Exported", strconv.Itoa(len(out.Items)),
		"Total", strconv.Itoa(out.TotalItems),
	)
	cli.PrintKV("Output", output)
}

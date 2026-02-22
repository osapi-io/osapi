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
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/retr0h/osapi/internal/audit/export"
	"github.com/retr0h/osapi/internal/cli"
	"github.com/retr0h/osapi/internal/client"
	gen "github.com/retr0h/osapi/internal/client/gen"
)

var (
	auditExportOutput string
	auditExportType   string
)

// clientAuditExportCmd represents the clientAuditExport command.
var clientAuditExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export audit log entries to a file",
	Long: `Export all audit log entries to a file for long-term retention.

Fetches all entries via the REST API export endpoint and writes each
entry as a JSON line (JSONL format). Requires audit:read permission.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		auditHandler := handler.(client.AuditHandler)

		resp, err := auditHandler.GetAuditExport(ctx)
		if err != nil {
			cli.LogFatal(logger, "API request failed", err)
		}

		switch resp.StatusCode() {
		case http.StatusOK:
			if resp.JSON200 == nil {
				cli.LogFatal(logger, "export failed", fmt.Errorf("response was nil"))
			}

			writeExport(ctx, resp.JSON200.Items, resp.JSON200.TotalItems)
		case http.StatusUnauthorized:
			cli.HandleAuthError(resp.JSON401, resp.StatusCode(), logger)
		case http.StatusForbidden:
			cli.HandleAuthError(resp.JSON403, resp.StatusCode(), logger)
		default:
			cli.HandleUnknownError(resp.JSON500, resp.StatusCode(), logger)
		}
	},
}

func writeExport(
	ctx context.Context,
	items []gen.AuditEntry,
	totalItems int,
) {
	var exporter export.Exporter
	switch auditExportType {
	case "file":
		exporter = export.NewFileExporter(auditExportOutput)
	default:
		cli.LogFatal(
			logger,
			"unsupported export type",
			fmt.Errorf("type %q is not supported, use \"file\"", auditExportType),
		)
	}

	if err := exporter.Open(ctx); err != nil {
		cli.LogFatal(logger, "opening exporter", err)
	}

	defer func() {
		if closeErr := exporter.Close(ctx); closeErr != nil {
			logger.Error("closing exporter", "error", closeErr)
		}
	}()

	for _, item := range items {
		if err := exporter.Write(ctx, item); err != nil {
			cli.LogFatal(logger, "writing entry", err)
		}
	}

	fmt.Println()
	cli.PrintKV(
		"Exported", strconv.Itoa(len(items)),
		"Total", strconv.Itoa(totalItems),
	)
	cli.PrintKV("Output", auditExportOutput)
}

func init() {
	clientAuditCmd.AddCommand(clientAuditExportCmd)
	clientAuditExportCmd.Flags().
		StringVar(&auditExportOutput, "output", "", "Output file path (required)")
	clientAuditExportCmd.Flags().
		StringVar(&auditExportType, "type", "file", "Export backend type")
	_ = clientAuditExportCmd.MarkFlagRequired("output")
}

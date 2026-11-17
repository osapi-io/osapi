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
package audit

import (
	"context"
	"errors"

	auditstore "github.com/osapi-io/osapi/internal/audit"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
)

// defaultLimit is the page size when the caller gives none.
const defaultLimit = 20

// List is a page of the audit log, newest first.
func List(
	ctx context.Context,
	d Deps,
	in ListInput,
) (ListAuditResponse, error) {
	if d.Store == nil {
		return ListAuditResponse{}, notConfigured()
	}
	limit := in.Limit
	if limit == 0 {
		limit = defaultLimit
	}

	entries, total, err := d.Store.List(ctx, limit, in.Offset)
	if err != nil {
		return ListAuditResponse{}, endpoint.Internal("failed to list audit entries", err)
	}

	return ListAuditResponse{TotalItems: total, Items: toEntries(entries)}, nil
}

// Export is each entry of the audit log, without pages.
func Export(
	ctx context.Context,
	d Deps,
	_ ExportInput,
) (ListAuditResponse, error) {
	if d.Store == nil {
		return ListAuditResponse{}, notConfigured()
	}

	entries, err := d.Store.ListAll(ctx)
	if err != nil {
		return ListAuditResponse{}, endpoint.Internal("failed to list all audit entries", err)
	}

	return ListAuditResponse{TotalItems: len(entries), Items: toEntries(entries)}, nil
}

// Get is one entry of the audit log.
func Get(
	ctx context.Context,
	d Deps,
	in GetInput,
) (AuditEntryResponse, error) {
	if d.Store == nil {
		return AuditEntryResponse{}, notConfigured()
	}

	entry, err := d.Store.Get(ctx, in.ID)
	if errors.Is(err, auditstore.ErrNotFound) {
		return AuditEntryResponse{}, endpoint.Errorf(ErrNotFound,
			"audit entry %s not found", in.ID)
	}
	if err != nil {
		return AuditEntryResponse{}, endpoint.Internal("failed to get audit entry", err)
	}

	return AuditEntryResponse{Entry: toEntry(*entry)}, nil
}

func notConfigured() error {
	return endpoint.Errorf(ErrNotConfigured,
		"the audit log is not configured: set nats.audit.stream on the controller")
}

func toEntries(
	entries []auditstore.Entry,
) []AuditEntry {
	items := make([]AuditEntry, 0, len(entries))
	for _, e := range entries {
		items = append(items, toEntry(e))
	}

	return items
}

func toEntry(
	e auditstore.Entry,
) AuditEntry {
	return AuditEntry{
		ID:             e.ID,
		Timestamp:      e.Timestamp,
		User:           e.User,
		Roles:          e.Roles,
		Method:         e.Method,
		Path:           e.Path,
		OperationID:    e.OperationID,
		SourceIP:       e.SourceIP,
		ResponseCode:   e.ResponseCode,
		DurationMs:     e.DurationMs,
		TraceID:        e.TraceID,
		JobID:          e.JobID,
		RequestSummary: e.RequestSummary,
	}
}

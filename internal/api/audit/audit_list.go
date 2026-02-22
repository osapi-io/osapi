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
	"log/slog"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/api/audit/gen"
	auditstore "github.com/retr0h/osapi/internal/audit"
	"github.com/retr0h/osapi/internal/validation"
)

// GetAuditLogs returns a paginated list of audit log entries.
func (a *Audit) GetAuditLogs(
	ctx context.Context,
	request gen.GetAuditLogsRequestObject,
) (gen.GetAuditLogsResponseObject, error) {
	if errMsg, ok := validation.Struct(request.Params); !ok {
		return gen.GetAuditLogs400JSONResponse{Error: &errMsg}, nil
	}

	limit := 20
	if request.Params.Limit != nil {
		limit = *request.Params.Limit
	}

	offset := 0
	if request.Params.Offset != nil {
		offset = *request.Params.Offset
	}

	entries, total, err := a.Store.List(ctx, limit, offset)
	if err != nil {
		a.logger.Error(
			"failed to list audit entries",
			slog.String("error", err.Error()),
		)
		errMsg := "failed to list audit entries"
		return gen.GetAuditLogs500JSONResponse{Error: &errMsg}, nil
	}

	items := make([]gen.AuditEntry, 0, len(entries))
	for _, e := range entries {
		items = append(items, mapEntryToGen(e))
	}

	return gen.GetAuditLogs200JSONResponse{
		TotalItems: total,
		Items:      items,
	}, nil
}

// mapEntryToGen converts an audit.Entry to the generated API type.
func mapEntryToGen(
	e auditstore.Entry,
) gen.AuditEntry {
	id := uuid.MustParse(e.ID)
	entry := gen.AuditEntry{
		Id:           id,
		Timestamp:    e.Timestamp,
		User:         e.User,
		Roles:        e.Roles,
		Method:       e.Method,
		Path:         e.Path,
		SourceIp:     e.SourceIP,
		ResponseCode: e.ResponseCode,
		DurationMs:   e.DurationMs,
	}
	if e.OperationID != "" {
		entry.OperationId = &e.OperationID
	}
	return entry
}

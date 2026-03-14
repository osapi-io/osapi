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

package client

import (
	"time"

	"github.com/retr0h/osapi/pkg/sdk/client/gen"
)

// AuditEntry represents a single audit log entry.
type AuditEntry struct {
	ID           string    `json:"id"`
	Timestamp    time.Time `json:"timestamp"`
	User         string    `json:"user"`
	Roles        []string  `json:"roles,omitempty"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	ResponseCode int       `json:"response_code"`
	DurationMs   int64     `json:"duration_ms"`
	SourceIP     string    `json:"source_ip"`
	OperationID  string    `json:"operation_id,omitempty"`
}

// AuditList is a paginated list of audit entries.
type AuditList struct {
	Items      []AuditEntry `json:"items"`
	TotalItems int          `json:"total_items"`
}

// auditEntryFromGen converts a gen.AuditEntry to an AuditEntry.
func auditEntryFromGen(
	g gen.AuditEntry,
) AuditEntry {
	a := AuditEntry{
		ID:           g.Id.String(),
		Timestamp:    g.Timestamp,
		User:         g.User,
		Roles:        g.Roles,
		Method:       g.Method,
		Path:         g.Path,
		ResponseCode: g.ResponseCode,
		DurationMs:   g.DurationMs,
		SourceIP:     g.SourceIp,
	}

	if g.OperationId != nil {
		a.OperationID = *g.OperationId
	}

	return a
}

// auditListFromGen converts a gen.ListAuditResponse to an AuditList.
func auditListFromGen(
	g *gen.ListAuditResponse,
) AuditList {
	items := make([]AuditEntry, 0, len(g.Items))
	for _, entry := range g.Items {
		items = append(items, auditEntryFromGen(entry))
	}

	return AuditList{
		Items:      items,
		TotalItems: g.TotalItems,
	}
}

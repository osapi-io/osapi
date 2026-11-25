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
	"net/http"
	"strings"

	auditstore "github.com/osapi-io/osapi/internal/audit"
	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/pkg/sdk/types"
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

	// A scoped token sees only the entries whose target is in its scope, so
	// its scope filters the log as a filter of the caller does.
	if !in.filtered() && len(authtoken.TargetsFrom(ctx)) == 0 {
		entries, total, err := d.Store.List(ctx, limit, in.Offset)
		if err != nil {
			return ListAuditResponse{}, endpoint.Internal("failed to list audit entries", err)
		}

		return ListAuditResponse{TotalItems: total, Items: toEntries(entries)}, nil
	}

	// A filter reads the whole log, newest first, as an export does, and
	// keeps the page of the entries that match. The store pages by
	// sequence, so it cannot skip the entries that do not match.
	all, err := d.Store.ListAll(ctx)
	if err != nil {
		return ListAuditResponse{}, endpoint.Internal("failed to list audit entries", err)
	}
	matched := make([]auditstore.Entry, 0, len(all))
	for _, e := range all {
		if in.matches(e) && authtoken.Sees(ctx, target(e)) {
			matched = append(matched, e)
		}
	}
	page := []auditstore.Entry{}
	if in.Offset < len(matched) {
		page = matched[in.Offset:min(in.Offset+limit, len(matched))]
	}

	return ListAuditResponse{TotalItems: len(matched), Items: toEntries(page)}, nil
}

// filtered reports whether in sets a filter.
func (in ListInput) filtered() bool {
	return in.Writes || in.User != "" || in.Operation != "" || in.Status != 0
}

// matches reports whether e matches each filter that in sets.
func (in ListInput) matches(
	e auditstore.Entry,
) bool {
	switch {
	case in.Writes && e.Method == http.MethodGet:
		return false
	case in.User != "" && e.User != in.User:
		return false
	case in.Operation != "" &&
		!strings.Contains(strings.ToLower(e.OperationID), strings.ToLower(in.Operation)) &&
		!strings.Contains(strings.ToLower(e.Path), strings.ToLower(in.Operation)):
		return false
	case in.Status != 0 && e.ResponseCode != in.Status:
		return false
	default:
		return true
	}
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

	all, err := d.Store.ListAll(ctx)
	if err != nil {
		return ListAuditResponse{}, endpoint.Internal("failed to list all audit entries", err)
	}
	entries := make([]auditstore.Entry, 0, len(all))
	for _, e := range all {
		if authtoken.Sees(ctx, target(e)) {
			entries = append(entries, e)
		}
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

	// An entry outside the scope of the token does not exist for it.
	entry, err := d.Store.Get(ctx, in.ID)
	if errors.Is(err, auditstore.ErrNotFound) ||
		(err == nil && !authtoken.Sees(ctx, target(*entry))) {
		return AuditEntryResponse{}, endpoint.Errorf(ErrNotFound,
			"audit entry %s not found", in.ID)
	}
	if err != nil {
		return AuditEntryResponse{}, endpoint.Internal("failed to get audit entry", err)
	}

	return AuditEntryResponse{Entry: toEntry(*entry)}, nil
}

// nodePrefix is the start of the path of each request that acts on a
// target.
const nodePrefix = "/api/node/"

// target is the target of the request that e records: the {hostname} of a
// /api/node/{hostname}/ path, such as group:web.dev or web-01. It is empty
// for a request that has no target, such as a read of the jobs.
func target(
	e auditstore.Entry,
) string {
	rest, ok := strings.CutPrefix(e.Path, nodePrefix)
	if !ok {
		return ""
	}
	t, _, _ := strings.Cut(rest, "/")

	return t
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
	out := AuditEntry{
		ID:             e.ID,
		Timestamp:      e.Timestamp,
		User:           e.User,
		Actor:          string(e.Actor),
		Action:         e.Action,
		Target:         e.Target,
		Roles:          rolesOf(e),
		Targets:        e.Targets,
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
	for _, h := range e.Steps {
		out.Steps = append(out.Steps, HostSteps{Hostname: h.Hostname, Steps: h.Steps})
		out.Chain = append(out.Chain, h.Hostname+": "+types.FormatSteps(h.Steps))
	}

	return out
}

// rolesOf is the roles of e, or none: an entry that no token made, such as
// one of a local command or of the system, has no roles, and the spec of
// the response has a list, never null.
func rolesOf(
	e auditstore.Entry,
) []types.Role {
	if e.Roles == nil {
		return []types.Role{}
	}

	return e.Roles
}

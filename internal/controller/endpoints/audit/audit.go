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
// Package audit is the endpoint group for the audit log: the record of each
// request the controller served. Until the group moves, the controller
// serves the hand-written spec in internal/controller/api/audit, and
// TestEndpointParity holds the two together.
package audit

import (
	"errors"

	"github.com/osapi-io/osapi/internal/controller/endpoint"
)

// ErrNotFound is an audit entry that is not there.
var ErrNotFound = errors.New("audit entry not found")

// Group is the audit group's definition.
var Group = endpoint.Register(endpoint.Define(
	"audit", "The audit log",
	endpoint.Title("Audit Log API"),
	endpoint.Tag("audit", "Audit", "Audit log endpoints for viewing API activity."),
	endpoint.Get[ListInput, ListAuditResponse]("list").
		ID("GetAuditLogs").
		Describe("List audit log entries").
		Explain("Returns a paginated list of audit log entries, newest first.").
		Returns("List of audit entries."),
	endpoint.Get[ExportInput, ListAuditResponse]("export").
		Sub("export").
		ID("GetAuditExport").
		Describe("Export all audit log entries").
		Explain("Returns all audit log entries without pagination for export.").
		Returns("All audit entries."),
	endpoint.Get[GetInput, AuditEntryResponse]("get").
		ID("GetAuditLogByID").
		Describe("Get a single audit log entry").
		Explain("Returns a single audit log entry by ID.").
		Returns("Audit entry found.").
		Errors(endpoint.NotFound(ErrNotFound, "Audit entry not found.")),
))

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
// request the controller served.
package audit

import (
	"errors"

	"github.com/osapi-io/osapi/internal/controller/endpoint"
)

var (
	// ErrNotFound is an audit entry that is not there.
	ErrNotFound = errors.New("audit entry not found")
	// ErrNotConfigured is a controller with no audit stream configured.
	ErrNotConfigured = errors.New("the audit log is not configured")
)

// notConfigured is the 503 of each method, when the controller keeps no
// audit log.
var notConfigured503 = endpoint.Unavailable(ErrNotConfigured,
	"The audit log is not configured on the controller.")

// Group is the audit group's definition.
var Group = endpoint.Register(endpoint.Define(
	"audit", "Read the audit log",
	endpoint.Title("Audit Log API"),
	endpoint.Tag("audit", "Audit", "Audit log endpoints for viewing API activity."),
	endpoint.Uses[Deps](),
	endpoint.Example[AuditEntry](
		"request_summary",
		`{"command":"systemctl","args":["restart","nginx"]}`,
	),
	endpoint.Get[ListInput, ListAuditResponse]("list").
		AdminOnly().
		Fails("a limit below 1", ListInput{Limit: -1}).
		Fails("a limit over 100", ListInput{Limit: 101}).
		Fails("a negative offset", ListInput{Offset: -1}).
		Handle(List).
		ID("GetAuditLogs").
		Describe("List audit log entries").
		Explain("Returns a paginated list of audit log entries, newest first.").
		Returns("List of audit entries.").
		Command("List audit log entries",
			"The list shows recent API activity: the user, the HTTP method, the path,\n"+
				"the response status and the duration.").
		Example("").
		Prints(sampleList).
		Example("--limit 10 --offset 20").
		Example("--json").
		Errors(notConfigured503),
	endpoint.Get[ExportInput, ListAuditResponse]("export").
		AdminOnly().
		Handle(Export).
		Sub("export").
		ID("GetAuditExport").
		Describe("Export all audit log entries").
		Explain("Returns all audit log entries without pagination for export.").
		Returns("All audit entries.").
		Command("Show all audit log entries, with no pages",
			"To save them, send the JSON to a file:\n"+
				"osapi client audit export --json > audit.json").
		Example("").
		Prints(sampleList).
		Example("--json").
		Errors(notConfigured503),
	endpoint.Get[GetInput, AuditEntryResponse]("get").
		AdminOnly().
		Fails("an ID that is not a UUID", GetInput{ID: "not-a-uuid"}).
		Handle(Get).
		ID("GetAuditLogByID").
		Describe("Get a single audit log entry").
		Explain("Returns a single audit log entry by ID.").
		Returns("Audit entry found.").
		Command("Get a single audit log entry", "").
		Example("--id 550e8400-e29b-41d4-a716-446655440000").
		Prints(sampleGet).
		Example("--id 550e8400-e29b-41d4-a716-446655440000 --json").
		Errors(endpoint.NotFound(ErrNotFound, "Audit entry not found."), notConfigured503),
))

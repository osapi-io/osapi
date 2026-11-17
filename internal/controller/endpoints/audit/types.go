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

import "time"

// ListInput is a page of the audit log.
type ListInput struct {
	Limit  int `json:"limit"  in:"query" validate:"omitempty,min=1,max=100" default:"20" doc:"Maximum number of entries to return."`
	Offset int `json:"offset" in:"query" validate:"omitempty,min=0"         default:"0"  doc:"Number of entries to skip."`
}

// ExportInput is the input of export, which has none.
type ExportInput struct{}

// GetInput names the audit entry to read.
type GetInput struct {
	ID string `json:"id" in:"path" validate:"uuid" doc:"The audit entry ID."`
}

// ListAuditResponse is a page of the audit log, or all of it.
type ListAuditResponse struct {
	TotalItems int          `json:"total_items" doc:"Total number of audit entries."   example:"150"`
	Items      []AuditEntry `json:"items"       doc:"The audit entries for this page."`
}

// AuditEntryResponse is one audit entry.
type AuditEntryResponse struct { //nolint:revive // the schema name is part of the API
	Entry AuditEntry `json:"entry"`
}

// AuditEntry is the record of one request.
type AuditEntry struct { //nolint:revive // the schema name is part of the API
	ID             string    `json:"id"                        doc:"Unique identifier for the audit entry."                                                                        example:"550e8400-e29b-41d4-a716-446655440000"               validate:"uuid"`
	Timestamp      time.Time `json:"timestamp"                 doc:"When the request was processed."                                                                               example:"2026-02-21T10:30:00Z"`
	User           string    `json:"user"                      doc:"Authenticated user (JWT subject)."                                                                             example:"ops@example.com"`
	Roles          []string  `json:"roles"                     doc:"Roles from the JWT token."`
	Method         string    `json:"method"                    doc:"HTTP method."                                                                                                  example:"GET"`
	Path           string    `json:"path"                      doc:"Request URL path."                                                                                             example:"/node/hostname"`
	OperationID    string    `json:"operation_id,omitempty"    doc:"OpenAPI operation ID."                                                                                         example:"GetNodeHostname"`
	SourceIP       string    `json:"source_ip"                 doc:"Client IP address."                                                                                            example:"192.168.1.100"`
	ResponseCode   int       `json:"response_code"             doc:"HTTP response status code."                                                                                    example:"200"`
	DurationMs     int64     `json:"duration_ms"               doc:"Request duration in milliseconds."                                                                             example:"42"`
	TraceID        string    `json:"trace_id,omitempty"        doc:"OpenTelemetry trace ID for correlation."                                                                       example:"4bf92f3577b34da6a3ce929d0e0e4736"`
	JobID          string    `json:"job_id,omitempty"          doc:"The job this request created, absent when it created none. Joins the entry to that job's own status timeline." example:"7c9e6679-7425-40de-944b-e07fc1f90ae7"`
	RequestSummary string    `json:"request_summary,omitempty" doc:"What the request asked for, with sensitive values replaced. Recorded for methods that change something."       example:"{\"command\":\"systemctl\",\"args\":[\"restart\",\"nginx\"]}"`
}

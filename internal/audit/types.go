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

// Package audit provides audit logging types and storage.
package audit

import "time"

// Entry represents a single audit log record.
type Entry struct {
	// ID is the unique identifier for this audit entry.
	ID string `json:"id"`
	// Timestamp is when the request was processed.
	Timestamp time.Time `json:"timestamp"`
	// User is the authenticated subject (from JWT sub claim).
	User string `json:"user"`
	// Roles are the roles from the JWT token.
	Roles []string `json:"roles"`
	// Method is the HTTP method (GET, POST, PUT, DELETE).
	Method string `json:"method"`
	// Path is the request URL path.
	Path string `json:"path"`
	// OperationID is the OpenAPI operation ID (if available).
	OperationID string `json:"operation_id,omitempty"`
	// SourceIP is the client's IP address.
	SourceIP string `json:"source_ip"`
	// ResponseCode is the HTTP response status code.
	ResponseCode int `json:"response_code"`
	// DurationMs is the request processing time in milliseconds.
	DurationMs int64 `json:"duration_ms"`
}

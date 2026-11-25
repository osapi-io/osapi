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

import (
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// Entry represents a single audit log record.
type Entry struct {
	// ID is the unique identifier for this audit entry.
	ID string `json:"id"`
	// Timestamp is when the request was processed.
	Timestamp time.Time `json:"timestamp"`
	// User is who did it: the authenticated subject (from JWT sub claim) of
	// a request, or the name of the actor of another kind.
	User string `json:"user"`
	// Actor is the kind of who did it: empty for a request through the API,
	// or system, join-token or local (osapi-io/osapi#1425).
	Actor Actor `json:"actor,omitempty"`
	// Action is what a change that did not go through the API did, such as
	// agent.admit or jointoken.create.
	Action string `json:"action,omitempty"`
	// Target is what the action was on, such as a machine ID or a token ID.
	Target string `json:"target,omitempty"`
	// Roles are the roles from the JWT token.
	Roles []types.Role `json:"roles"`
	// Targets are the target scope from the JWT token, empty when the token
	// acts on each host.
	Targets []string `json:"targets,omitempty"`
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
	// TraceID is the OpenTelemetry trace ID for correlation.
	TraceID string `json:"trace_id,omitempty"`
	// JobID is the job the request created, empty for a request that created
	// none. It is what joins an audit entry to the job's own status timeline.
	JobID string `json:"job_id,omitempty"`
	// Steps are the calls of other behaviors that the job's method made, for
	// each host that answered before the request returned, also when one
	// failed. Empty when no host made a call.
	Steps []HostSteps `json:"steps,omitempty"`
	// RequestSummary is what the request asked for, with sensitive values
	// replaced. Recorded for methods that change something, because "who ran
	// command/shell against which host" without the command is not an answer.
	RequestSummary string `json:"request_summary,omitempty"`
}

// HostSteps are the calls of other behaviors that the job's method made on one
// host.
type HostSteps struct {
	// Hostname is the host that answered the job.
	Hostname string `json:"hostname"`
	// Steps are the calls, in the order that the method made them.
	Steps []types.Step `json:"steps"`
}

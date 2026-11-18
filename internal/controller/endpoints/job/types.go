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
package job

// ListInput is a page of the job queue, and the status to filter by.
type ListInput struct {
	Status string `json:"status" in:"query" validate:"omitempty,oneof=submitted processing completed failed partial_failure" doc:"Filter jobs by status."`
	Limit  int    `json:"limit"  in:"query" validate:"omitempty,min=1,max=100"                                               doc:"Maximum number of jobs per page (1-100)." default:"10"`
	Offset int    `json:"offset" in:"query" validate:"omitempty,min=0"                                                       doc:"Number of jobs to skip for pagination."   default:"0"`
}

// GetInput names the job to read.
type GetInput struct {
	ID string `json:"id" in:"path" validate:"uuid" doc:"UUID of the job to retrieve."`
}

// DeleteInput names the job to delete.
type DeleteInput struct {
	ID string `json:"id" in:"path" validate:"uuid" doc:"UUID of the job to delete."`
}

// RetryInput names the job to retry, and where the new job goes.
type RetryInput struct {
	ID   string          `json:"id" in:"path" validate:"uuid"      doc:"UUID of the job to retry."`
	Body RetryJobRequest `          in:"body" validate:"omitempty"`
}

// RetryJobRequest is where a retried job goes.
type RetryJobRequest struct {
	TargetHostname string `json:"target_hostname,omitempty" validate:"omitempty,min=1,valid_target" doc:"The target of the retried job: _any, _all, a hostname, or a label such as group:web.dev." example:"_any" flag:"target" short:"T" default:"_any"`
}

// CreateJobResponse is the job that a request created.
type CreateJobResponse struct {
	JobID     string `json:"job_id"              validate:"uuid" doc:"Unique identifier for the created job." example:"550e8400-e29b-41d4-a716-446655440000"`
	Status    string `json:"status"                              doc:"Initial status of the job."             example:"unprocessed"`
	Revision  int64  `json:"revision,omitempty"                  doc:"The KV revision number."                example:"1"`
	Timestamp string `json:"timestamp,omitempty"                 doc:"Creation timestamp."                    example:"2025-06-14T10:00:00Z"`
}

// ListJobsResponse is a page of the job queue.
type ListJobsResponse struct {
	TotalItems   int                 `json:"total_items,omitempty"   doc:"Total number of jobs matching the filter."                                                                                                                     example:"42"`
	StatusCounts map[string]int      `json:"status_counts,omitempty" doc:"Count of all jobs by status (submitted, processing, completed, failed, partial_failure). Derived from key names during the listing pass, with no extra reads."`
	Items        []JobDetailResponse `json:"items,omitempty"`
}

// JobDetailResponse is one job, with what each agent did with it.
type JobDetailResponse struct { //nolint:revive // the schema name is part of the API
	ID          string                `json:"id,omitempty"           validate:"uuid" doc:"Unique identifier of the job."`
	Status      string                `json:"status,omitempty"                       doc:"Current status of the job."`
	Created     string                `json:"created,omitempty"                      doc:"Creation timestamp."`
	Operation   map[string]any        `json:"operation,omitempty"                    doc:"The operation data."`
	Result      any                   `json:"result,omitempty"                       doc:"The result data if completed."`
	Changed     *bool                 `json:"changed,omitempty"                      doc:"Whether the operation modified system state. Null for pending/failed jobs; true/false for completed jobs."`
	Error       string                `json:"error,omitempty"                        doc:"Error message if failed."`
	Hostname    string                `json:"hostname,omitempty"                     doc:"Agent hostname that processed the job."`
	Target      string                `json:"target,omitempty"                       doc:"The target the request gave: _any, _all, a hostname, or a label such as group:web.dev."`
	UpdatedAt   string                `json:"updated_at,omitempty"                   doc:"Last update timestamp."`
	Responses   map[string]Response   `json:"responses,omitempty"                    doc:"Per-agent response data for broadcast jobs."`
	AgentStates map[string]AgentState `json:"agent_states,omitempty"                 doc:"Per-agent processing state for broadcast jobs."`
	Timeline    []TimelineEvent       `json:"timeline,omitempty"                     doc:"Chronological sequence of job lifecycle events."`
}

// Response is what one agent answered, in place in the spec.
type Response = struct {
	Status   string `json:"status,omitempty"`
	Data     any    `json:"data,omitempty"     doc:"Agent result data."`
	Changed  *bool  `json:"changed,omitempty"`
	Error    string `json:"error,omitempty"`
	Hostname string `json:"hostname,omitempty"`
}

// AgentState is how far one agent got, in place in the spec.
type AgentState = struct {
	Status   string `json:"status,omitempty"`
	Error    string `json:"error,omitempty"`
	Duration string `json:"duration,omitempty"`
}

// TimelineEvent is one event of a job's life, in place in the spec.
type TimelineEvent = struct {
	Timestamp string `json:"timestamp,omitempty" doc:"ISO 8601 timestamp of the event."`
	Event     string `json:"event,omitempty"     doc:"Event type (submitted, acknowledged, started, completed, failed, retried)."`
	Hostname  string `json:"hostname,omitempty"  doc:"Agent or source that generated the event."`
	Message   string `json:"message,omitempty"   doc:"Human-readable description of the event."`
	Error     string `json:"error,omitempty"     doc:"Error details if applicable."`
}

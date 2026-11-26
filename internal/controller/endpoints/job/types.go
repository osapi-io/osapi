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

import (
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// ListInput is a page of the job queue, and the status to filter by.
type ListInput struct {
	Status types.JobStatus `json:"status" in:"query" validate:"omitempty,enum"          doc:"Filter jobs by status."`
	Limit  int             `json:"limit"  in:"query" validate:"omitempty,min=1,max=100" doc:"Maximum number of jobs per page (1-100)." default:"10"`
	Offset int             `json:"offset" in:"query" validate:"omitempty,min=0"         doc:"Number of jobs to skip for pagination."   default:"0"`
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
	JobID     string          `json:"job_id"              validate:"uuid" doc:"Unique identifier for the created job." example:"550e8400-e29b-41d4-a716-446655440000"`
	Status    types.JobStatus `json:"status"                              doc:"Initial status of the job."             example:"submitted"`
	Revision  int64           `json:"revision,omitempty"                  doc:"The KV revision number."                example:"1"`
	Timestamp string          `json:"timestamp,omitempty"                 doc:"Creation timestamp."                    example:"2025-06-14T10:00:00Z"                 render:"-"`
}

// ListJobsResponse is a page of the job queue.
type ListJobsResponse struct {
	_            struct{}                `title:"Jobs"`
	TotalItems   int                     `             json:"total_items,omitempty"   doc:"Total number of jobs matching the filter."                                                         example:"42"`
	StatusCounts map[types.JobStatus]int `             json:"status_counts,omitempty" doc:"Count of all jobs by status. Derived from key names during the listing pass, with no extra reads."`
	Items        []DetailResponse        `             json:"items,omitempty"`
}

// DetailResponse is one job, with what each agent did with it. The
// fields of a list come first, in the order of its columns.
type DetailResponse struct {
	_ struct{} `schema:"JobDetailResponse"`

	ID          string          `json:"id,omitempty"           validate:"uuid" doc:"Unique identifier of the job."                                                                                                                                    label:"Job ID"`
	Status      types.JobStatus `json:"status,omitempty"                       doc:"Current status of the job."                                                                                                                                                       render:"status"`
	Created     string          `json:"created,omitempty"                      doc:"Creation timestamp."                                                                                                                                                              render:"time"`
	Target      string          `json:"target,omitempty"                       doc:"The target the request gave: _any, _all, a hostname, or a label such as group:web.dev. A job stored before targets were kept has the host that answered."`
	Operation   *Operation      `json:"operation,omitempty"                    doc:"What the job asked for, with each secret masked."`
	Hostname    string          `json:"hostname,omitempty"                     doc:"Agent hostname that processed the job."                                                                                                                                                           table:"-"`
	UpdatedAt   string          `json:"updated_at,omitempty"                   doc:"Last update timestamp."                                                                                                                                           label:"Updated" render:"time"   table:"-"`
	Changed     *bool           `json:"changed,omitempty"                      doc:"Whether the operation modified system state. Null for pending/failed jobs; true/false for completed jobs."                                                                                        table:"-"`
	Error       string          `json:"error,omitempty"                        doc:"Error message if failed."                                                                                                                                                                         table:"-"`
	Result      any             `json:"result,omitempty"                       doc:"The result data if completed."                                                                                                                                                                    table:"-"`
	Responses   []Response      `json:"responses,omitempty"                    doc:"What each agent answered, by hostname."                                                                                                                                                           table:"-"`
	Chain       []string        `json:"chain,omitempty"                        doc:"The calls of other behaviors that the method made on each host, one line for each host, such as web-01: netplan update (changed) → managedfile update (changed)."`
	AgentStates []AgentState    `json:"agent_states,omitempty"                 doc:"How far each agent got, by hostname."                                                                                                                                                             table:"-"`
	Timeline    []TimelineEvent `json:"timeline,omitempty"                     doc:"Chronological sequence of job lifecycle events."                                                                                                                                                  table:"-"`
}

// Operation is what a job asked for: its type, such as node.hostname.get,
// and its data, in place in the spec.
type Operation = struct {
	Type string `json:"type"           doc:"The operation, such as node.hostname.get."`
	Data any    `json:"data,omitempty" doc:"The data of the operation, with each secret masked." table:"-" label:"Operation data"`
}

// Response is what one agent answered, in place in the spec.
type Response = struct {
	Hostname string               `json:"hostname"`
	Status   types.AgentJobStatus `json:"status,omitempty"  render:"status"`
	Changed  *bool                `json:"changed,omitempty"`
	Error    string               `json:"error,omitempty"`
	Data     any                  `json:"data,omitempty"                    doc:"Agent result data."`
	Steps    []types.Step         `json:"steps,omitempty"                   doc:"The calls of other behaviors that the method made on the host, also when one failed."`
}

// AgentState is how far one agent got, in place in the spec.
type AgentState = struct {
	Hostname string               `json:"hostname"`
	Status   types.AgentJobStatus `json:"status,omitempty"   render:"status"   doc:"How far the agent got."`
	Duration string               `json:"duration,omitempty" render:"duration"`
	Error    string               `json:"error,omitempty"`
}

// TimelineEvent is one event of a job's life, in place in the spec.
type TimelineEvent = struct {
	Timestamp string               `json:"timestamp,omitempty" doc:"ISO 8601 timestamp of the event."          render:"time"`
	Event     types.AgentJobStatus `json:"event,omitempty"     doc:"The status event."`
	Hostname  string               `json:"hostname,omitempty"  doc:"Agent or source that generated the event."`
	Message   string               `json:"message,omitempty"   doc:"Human-readable description of the event."`
	Error     string               `json:"error,omitempty"     doc:"Error details if applicable."`
}

// DeadLetterListInput limits the dead letters.
type DeadLetterListInput struct {
	Limit int `json:"limit" in:"query" validate:"omitempty,min=1,max=100" doc:"The most dead letters to list, the newest first (1-100)." default:"20"`
}

// DeadLetterList is the newest dead letters, with the count of them all.
type DeadLetterList struct {
	_          struct{}     `title:"Dead letters"`
	TotalItems int          `                     json:"total_items" doc:"The number of dead letters that the stream holds."`
	Items      []DeadLetter `                     json:"items"`
}

// DeadLetter is a job whose delivery to an agent used up its retries.
type DeadLetter struct {
	Seq        uint64     `json:"seq"                 doc:"The dead letter's sequence in the dead letter stream."                             example:"654"`
	Time       time.Time  `json:"time"                doc:"When NATS gave up on the delivery."`
	JobID      string     `json:"job_id,omitempty"    doc:"The job that the delivery carried, or empty when the job stream no longer has it." example:"550e8400-e29b-41d4-a716-446655440000"`
	Operation  *Operation `json:"operation,omitempty" doc:"What the job asked for, with each secret masked."                                                                                 table:"-"`
	Target     string     `json:"target,omitempty"    doc:"The target of the job."                                                            example:"_all"`
	Error      string     `json:"error,omitempty"     doc:"The error of the job, when it has one."`
	Deliveries uint64     `json:"deliveries"          doc:"How often NATS delivered the job before it gave up."                               example:"5"`
	Consumer   string     `json:"consumer"            doc:"The consumer of the agents that the delivery was for."                                                                            table:"-"`
}

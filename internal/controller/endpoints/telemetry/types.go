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

package telemetry

import (
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// TraceInput names the trace to read.
type TraceInput struct {
	ID  string `json:"id"  in:"path"  validate:"required"  pattern:"^[0-9a-f]{32}$" doc:"The trace ID, which a command prints after its job ID, 32 hexadecimal characters"`
	All bool   `json:"all" in:"query" validate:"omitempty"                          doc:"Show the NATS calls too, the reads and writes of the job queue"`
}

// TraceResponse is one trace: its summary, then its spans as a tree.
type TraceResponse struct {
	_           struct{}    `title:"Trace"`
	TraceID     string      `              json:"trace_id"               doc:"The trace ID."                                                                        label:"ID"`
	Started     time.Time   `              json:"started,omitzero"       doc:"When the first span started."                                                                                     render:"time"`
	Duration    string      `              json:"duration,omitempty"     doc:"From the start of the first span to the end of the last, as a Go duration."                                       render:"duration"`
	Services    []string    `              json:"services,omitempty"     doc:"The services that made the spans, in the order each first appears."`
	FailedSpans int         `              json:"failed_spans,omitempty" doc:"The number of spans that failed."`
	NATSCalls   int         `              json:"nats_calls,omitempty"   doc:"The number of NATS calls left out, the bookkeeping of the job queue. all shows them." label:"NATS calls left out"`
	NATSTime    string      `              json:"nats_time,omitempty"    doc:"The time that the NATS calls left out took, as a Go duration."                        label:"NATS call time"      render:"duration"`
	Spans       []TraceNode `              json:"spans"                  doc:"The spans that have no parent in the trace, each with its children, in start order."`
}

// TraceNode is one span of the tree of a trace: what ran it, its status
// and what it acted on, then the span itself and its children.
type TraceNode struct {
	Name      string      `json:"name"                 doc:"The name of the span, such as job.process."                                                                                                                                                    label:"Span"`
	On        string      `json:"on"                   doc:"What ran the span: the role of its service, such as controller, agent or cli."`
	Duration  string      `json:"duration"             doc:"How long the span took, as a Go duration."                                                                                                                                                                  render:"duration"`
	Status    string      `json:"status"               doc:"The job's status on a job.process, changed or unchanged on a behavior write, and ok or error on other spans."                                                                                               render:"status"`
	Detail    string      `json:"detail,omitempty"     doc:"What the span acted on, as key=value pairs, led by error= on a failed span."`
	ClockSkew string      `json:"clock_skew,omitempty" doc:"How far the span and the spans under it were moved forward, as a Go duration, because the span started before its parent: the clock of its host is behind. Empty when the span was not moved."                                table:"-"`
	Span      TraceSpan   `json:"span"                 doc:"The span as the backend has it."                                                                                                                                                                                              table:"-"`
	Children  []TraceNode `json:"children,omitempty"   doc:"The spans under this one, in start order."`
}

// TraceSpan is one span of a trace.
type TraceSpan struct {
	SpanID        string            `json:"span_id"                  doc:"The span ID."`
	ParentSpanID  string            `json:"parent_span_id,omitempty" doc:"The ID of the parent span. Empty for the root."`
	Name          string            `json:"name"                     doc:"The name of the span, such as job.process."`
	Service       string            `json:"service"                  doc:"The service that made the span, such as osapi-agent."`
	Kind          string            `json:"kind,omitempty"           doc:"The span kind, such as server, client, producer or consumer."`
	StartTime     time.Time         `json:"start_time"               doc:"When the span started."`
	DurationMs    float64           `json:"duration_ms"              doc:"How long the span took, in milliseconds."`
	Status        types.SpanStatus  `json:"status"                   doc:"Whether the work of the span failed."                         validate:"enum"`
	StatusMessage string            `json:"status_message,omitempty" doc:"The error of a failed span."`
	Attributes    map[string]string `json:"attributes"               doc:"The attributes of the span, each as text."`
}

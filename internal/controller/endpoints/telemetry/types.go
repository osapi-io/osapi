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

import "time"

// TraceInput names the trace to read.
type TraceInput struct {
	ID string `json:"id" in:"path" pattern:"^[0-9a-f]{32}$" doc:"The trace ID, which a command prints after its job ID, 32 hexadecimal characters"`
}

// TraceResponse is the spans of one trace.
type TraceResponse struct {
	TraceID string      `json:"trace_id" doc:"The trace ID."`
	Spans   []TraceSpan `json:"spans"    doc:"Each span of the trace, ordered by start time."`
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
	Status        string            `json:"status"                   doc:"Whether the work of the span failed."                         validate:"oneof=ok error"`
	StatusMessage string            `json:"status_message,omitempty" doc:"The error of a failed span."`
	Attributes    map[string]string `json:"attributes"               doc:"The attributes of the span, each as text."`
}

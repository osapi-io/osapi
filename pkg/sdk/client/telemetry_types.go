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

package client

import (
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/client/gen"
)

// Trace is the spans of one trace, ordered by start time.
type Trace struct {
	// ID is the trace ID.
	ID string `json:"trace_id"`
	// Spans are its spans, ordered by start time.
	Spans []TraceSpan `json:"spans"`
}

// TraceSpan is one span of a trace.
type TraceSpan struct {
	// ID is the span ID, and ParentID the ID of its parent, empty for the
	// root.
	ID       string `json:"span_id"`
	ParentID string `json:"parent_span_id,omitempty"`
	// Name is the name of the span, such as job.process, and Service the
	// service that made it, such as osapi-agent.
	Name    string `json:"name"`
	Service string `json:"service"`
	// Kind is the span kind, such as server or consumer.
	Kind string `json:"kind,omitempty"`
	// Start is when it started, and Duration how long it took.
	Start    time.Time     `json:"start_time"`
	Duration time.Duration `json:"duration"`
	// Failed is set when its work failed, with the error in StatusMessage.
	Failed        bool   `json:"failed"`
	StatusMessage string `json:"status_message,omitempty"`
	// Attributes are its attributes, each as text.
	Attributes map[string]string `json:"attributes"`
}

// traceFromGen converts the generated trace to the SDK type.
func traceFromGen(
	g *gen.TraceResponse,
) Trace {
	t := Trace{ID: g.TraceId, Spans: make([]TraceSpan, 0, len(g.Spans))}
	for _, s := range g.Spans {
		t.Spans = append(t.Spans, TraceSpan{
			ID:            s.SpanId,
			ParentID:      derefString(s.ParentSpanId),
			Name:          s.Name,
			Service:       s.Service,
			Kind:          derefString(s.Kind),
			Start:         s.StartTime,
			Duration:      time.Duration(s.DurationMs * float64(time.Millisecond)),
			Failed:        s.Status == gen.TraceSpanStatusError,
			StatusMessage: derefString(s.StatusMessage),
			Attributes:    s.Attributes,
		})
	}

	return t
}

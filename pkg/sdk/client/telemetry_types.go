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

// Trace is one trace: its summary, then its spans as a tree.
type Trace struct {
	// ID is the trace ID.
	ID string `json:"trace_id"`
	// Started is when the first span started, and Duration the time from
	// it to the end of the last span.
	Started  time.Time     `json:"started"`
	Duration time.Duration `json:"duration"`
	// Services are the services that made the spans, in the order each
	// first appears.
	Services []string `json:"services,omitempty"`
	// FailedSpans is the number of spans that failed.
	FailedSpans int `json:"failed_spans,omitempty"`
	// NATSCalls is the number of NATS calls left out, and NATSTime the time
	// they took. Asking for all of the trace leaves none out.
	NATSCalls int           `json:"nats_calls,omitempty"`
	NATSTime  time.Duration `json:"nats_time,omitempty"`
	// Spans are the spans with no parent in the trace, each with its
	// children, in start order.
	Spans []TraceNode `json:"spans"`
}

// TraceNode is one span of the tree of a trace.
type TraceNode struct {
	// Name is the name of the span, such as job.process, and On what ran
	// it: controller, agent or cli.
	Name string `json:"name"`
	On   string `json:"on"`
	// Status is the job's status on a job.process, changed or unchanged on
	// a behavior write, and ok or error on other spans.
	Status string `json:"status"`
	// Detail is what the span acted on, as key=value pairs.
	Detail string `json:"detail,omitempty"`
	// Span is the span as the backend has it.
	Span TraceSpan `json:"span"`
	// Children are the spans under this one, in start order.
	Children []TraceNode `json:"children,omitempty"`
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
	t := Trace{
		ID:          g.TraceId,
		Services:    derefStringSlice(g.Services),
		FailedSpans: derefInt(g.FailedSpans),
		NATSCalls:   derefInt(g.NatsCalls),
		Spans:       nodesFromGen(g.Spans),
	}
	if g.Started != nil {
		t.Started = *g.Started
	}
	t.Duration, _ = time.ParseDuration(derefString(g.Duration))
	t.NATSTime, _ = time.ParseDuration(derefString(g.NatsTime))

	return t
}

// nodesFromGen converts the generated span tree, with the children of each.
func nodesFromGen(
	in []gen.TraceNode,
) []TraceNode {
	out := make([]TraceNode, 0, len(in))
	for _, n := range in {
		s := n.Span
		node := TraceNode{
			Name:   n.Name,
			On:     n.On,
			Status: n.Status,
			Detail: derefString(n.Detail),
			Span: TraceSpan{
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
			},
		}
		if n.Children != nil {
			node.Children = nodesFromGen(*n.Children)
		}
		out = append(out, node)
	}

	return out
}

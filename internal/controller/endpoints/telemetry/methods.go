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
	"context"
	"errors"
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/internal/telemetry/traces"
)

// Get is the spans of a trace, read from the tracing backend.
func Get(
	ctx context.Context,
	d Deps,
	in TraceInput,
) (TraceResponse, error) {
	if d.Traces == nil {
		return TraceResponse{}, endpoint.Errorf(ErrNoBackend,
			"no tracing backend is configured: set telemetry.tracing.url on the controller")
	}

	trace, err := d.Traces.Trace(ctx, in.ID)
	// A trace outside the scope of the token does not exist for it.
	if err == nil && !sees(ctx, trace) {
		return TraceResponse{}, endpoint.Errorf(ErrNotFound, "trace %s not found", in.ID)
	}
	if errors.Is(err, traces.ErrNotFound) {
		return TraceResponse{}, endpoint.Errorf(ErrNotFound,
			"trace %s not found: the backend gets the spans a few seconds after the "+
				"request, so try again", in.ID)
	}
	if err != nil {
		return TraceResponse{}, endpoint.Internal(
			"failed to read the trace from the tracing backend", err)
	}

	spans := make([]TraceSpan, 0, len(trace.Spans))
	for _, s := range trace.Spans {
		status := types.SpanOK
		if s.Failed {
			status = types.SpanError
		}
		spans = append(spans, TraceSpan{
			SpanID:        s.ID,
			ParentSpanID:  s.ParentID,
			Name:          s.Name,
			Service:       s.Service,
			Kind:          s.Kind,
			StartTime:     s.Start,
			DurationMs:    float64(s.Duration) / float64(time.Millisecond),
			Status:        status,
			StatusMessage: s.StatusMessage,
			Attributes:    s.Attributes,
		})
	}

	return tree(trace.ID, spans, in.All), nil
}

// targetAttr is the span attribute that records the target of a request.
const targetAttr = "osapi.target"

// sees reports whether the token of ctx sees trace. A token with no scope
// sees each trace. A scoped token sees a trace only when a span of it records
// a target and each target that its spans record is in the scope. Thus it
// does not see a trace with no target, or one that reached _all or _any.
func sees(
	ctx context.Context,
	trace *traces.Trace,
) bool {
	if len(authtoken.TargetsFrom(ctx)) == 0 {
		return true
	}

	found := false
	for _, s := range trace.Spans {
		target, ok := s.Attributes[targetAttr]
		if !ok {
			continue
		}
		if !authtoken.Sees(ctx, target) {
			return false
		}
		found = true
	}

	return found
}

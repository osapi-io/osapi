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
	"log/slog"
	"regexp"
	"time"

	"github.com/osapi-io/osapi/internal/controller/api/telemetry/gen"
	"github.com/osapi-io/osapi/internal/telemetry/traces"
)

// traceID is a W3C trace ID: 32 lowercase hexadecimal characters.
var traceID = regexp.MustCompile(`^[0-9a-f]{32}$`)

// GetTrace returns the spans of a trace from the tracing backend.
func (t *Telemetry) GetTrace(
	ctx context.Context,
	request gen.GetTraceRequestObject,
) (gen.GetTraceResponseObject, error) {
	if !traceID.MatchString(request.Id) {
		msg := "a trace ID is 32 lowercase hexadecimal characters"
		return gen.GetTrace400JSONResponse{Error: &msg}, nil
	}
	if t.reader == nil {
		msg := "no tracing backend is configured: set telemetry.tracing.url on the controller"
		return gen.GetTrace503JSONResponse{Error: &msg}, nil
	}

	trace, err := t.reader.Trace(ctx, request.Id)
	if errors.Is(err, traces.ErrNotFound) {
		msg := "trace " + request.Id + " not found: the backend gets the spans a few " +
			"seconds after the request, so try again"
		return gen.GetTrace404JSONResponse{Error: &msg}, nil
	}
	if err != nil {
		t.logger.ErrorContext(ctx, "failed to read trace",
			slog.String("trace_id", request.Id),
			slog.String("error", err.Error()))
		msg := "failed to read the trace from the tracing backend"
		return gen.GetTrace500JSONResponse{Error: &msg}, nil
	}

	spans := make([]gen.TraceSpan, 0, len(trace.Spans))
	for _, s := range trace.Spans {
		status := gen.Ok
		if s.Failed {
			status = gen.Error
		}
		span := gen.TraceSpan{
			SpanId:     s.ID,
			Name:       s.Name,
			Service:    s.Service,
			StartTime:  s.Start,
			DurationMs: float64(s.Duration) / float64(time.Millisecond),
			Status:     status,
			Attributes: s.Attributes,
		}
		if s.ParentID != "" {
			span.ParentSpanId = &s.ParentID
		}
		if s.Kind != "" {
			span.Kind = &s.Kind
		}
		if s.StatusMessage != "" {
			span.StatusMessage = &s.StatusMessage
		}
		spans = append(spans, span)
	}

	return gen.GetTrace200JSONResponse{TraceId: trace.ID, Spans: spans}, nil
}

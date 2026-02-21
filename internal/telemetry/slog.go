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
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// traceHandler wraps a slog.Handler to add trace_id and span_id attributes
// when a valid span context is present.
type traceHandler struct {
	inner slog.Handler
}

// NewTraceHandler creates a slog.Handler that adds trace_id and span_id
// attributes from the context's span, delegating to the inner handler.
func NewTraceHandler(
	inner slog.Handler,
) slog.Handler {
	return &traceHandler{inner: inner}
}

// Enabled reports whether the inner handler handles records at the given level.
func (h *traceHandler) Enabled(
	ctx context.Context,
	level slog.Level,
) bool {
	return h.inner.Enabled(ctx, level)
}

// Handle adds trace_id and span_id attributes if a valid span context exists,
// then delegates to the inner handler.
func (h *traceHandler) Handle(
	ctx context.Context,
	record slog.Record,
) error {
	sc := trace.SpanContextFromContext(ctx)
	if sc.IsValid() {
		record.AddAttrs(
			slog.String("trace_id", sc.TraceID().String()),
			slog.String("span_id", sc.SpanID().String()),
		)
	}

	return h.inner.Handle(ctx, record)
}

// WithAttrs returns a new handler with the given attributes.
func (h *traceHandler) WithAttrs(
	attrs []slog.Attr,
) slog.Handler {
	return &traceHandler{inner: h.inner.WithAttrs(attrs)}
}

// WithGroup returns a new handler with the given group name.
func (h *traceHandler) WithGroup(
	name string,
) slog.Handler {
	return &traceHandler{inner: h.inner.WithGroup(name)}
}

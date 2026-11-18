// Copyright (c) 2025 John Dewey

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

// Package opspan opens the span of one operation of a store or a client,
// such as job.list or audit.get. The calls that the operation makes, such
// as the NATS calls, are its children. Thus a trace tells what the time of a
// request was for, where it would show only NATS calls.
package opspan

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Start opens the span name, from the tracer of the package that tracer
// names, with attrs.
func Start(
	ctx context.Context,
	tracer string,
	name string,
	attrs ...attribute.KeyValue,
) (context.Context, trace.Span) {
	return otel.Tracer(tracer).Start(ctx, name, trace.WithAttributes(attrs...))
}

// End ends span, with the error status when err is not nil, and returns
// err, so a call can end with return opspan.End(span, err).
func End(
	span trace.Span,
	err error,
) error {
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()

	return err
}

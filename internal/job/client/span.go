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

package client

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/osapi-io/osapi/internal/job"
)

// tracerName names the spans of the job client in a trace.
const tracerName = "osapi-job-client"

// startPublishSpan opens job.publish: the controller stores the job, finds
// the agents that a broadcast expects, and publishes the job. The trace
// context goes with the job from this span, so each agent's job.process is a
// child of it.
func startPublishSpan(
	ctx context.Context,
	jobID string,
	subject string,
	target string,
	req *job.Request,
) (context.Context, trace.Span) {
	return otel.Tracer(tracerName).Start(ctx, "job.publish",
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			attribute.String("job.id", jobID),
			attribute.String("job.type", req.Operation),
			attribute.String("job.kind", string(req.Type)),
			attribute.String("osapi.target", target),
			attribute.String("messaging.destination.name", subject),
		),
	)
}

// startCollectSpan opens job.collect: the controller waits for the answers of
// the agents. It is a sibling of job.publish, under the request.
func startCollectSpan(
	ctx context.Context,
	jobID string,
) (context.Context, trace.Span) {
	return otel.Tracer(tracerName).Start(ctx, "job.collect",
		trace.WithAttributes(attribute.String("job.id", jobID)),
	)
}

// endSpan marks span as failed with err, ends it, and returns err.
func endSpan(
	span trace.Span,
	err error,
) error {
	span.SetStatus(codes.Error, err.Error())
	span.End()

	return err
}

// endCollectSpan records the answers that a broadcast collected: how many
// agents the controller expected, how many answered, and how many timed out.
// A timed-out agent is not an error of the request, so it is an attribute
// and an event, not the error status.
func endCollectSpan(
	span trace.Span,
	expected int,
	responses map[string]*job.Response,
) {
	timedOut := 0
	for hostname, r := range responses {
		if r.Status == job.StatusTimeout {
			timedOut++
			span.AddEvent("job.timeout", trace.WithAttributes(
				attribute.String("osapi.agent.hostname", hostname),
			))
		}
	}
	span.SetAttributes(
		attribute.Int("job.expected", expected),
		attribute.Int("job.answered", len(responses)-timedOut),
		attribute.Int("job.timed_out", timedOut),
	)
	span.End()
}

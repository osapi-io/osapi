// Copyright (c) 2026 John Dewey
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package agent

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/osapi-io/osapi/internal/job"
)

// identityAttrs tell which agent ran a job: the hostname that it answers to,
// its machine ID, and its architecture from the facts that it collected. The
// agent resolved each of them before the job, so a span looks up nothing.
// In a broadcast, these are what make one job.process different from the
// next.
func (a *Agent) identityAttrs() []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.String("osapi.agent.hostname", a.hostname),
		attribute.String("osapi.agent.machine_id", a.machineID),
	}
	if facts := a.cachedFacts.Load(); facts != nil && facts.Architecture != "" {
		attrs = append(attrs, attribute.String("host.arch", facts.Architecture))
	}

	return attrs
}

// recordOutcome puts the result of a job on its span: job.status, and
// job.changed for a write that completed. A failed job gets the error status,
// so that Jaeger marks it. A skipped job is not an error, because the host
// cannot run the operation; job.status and an event say that it skipped.
func recordOutcome(
	span trace.Span,
	response job.Response,
	requestType job.Type,
) {
	span.SetAttributes(attribute.String("job.status", string(response.Status)))

	switch response.Status {
	case job.StatusFailed:
		span.SetStatus(codes.Error, response.Error)
	case job.StatusSkipped:
		span.AddEvent("job.skipped", trace.WithAttributes(
			attribute.String("job.skip_reason", response.Error),
		))
	case job.StatusCompleted:
		if requestType == job.TypeModify && response.Changed != nil {
			span.SetAttributes(attribute.Bool("job.changed", *response.Changed))
		}
	}
}

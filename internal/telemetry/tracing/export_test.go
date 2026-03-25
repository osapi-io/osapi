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

// Package tracing provides OpenTelemetry tracing initialization and helpers.
package tracing

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

// ExportNewMapCarrier exposes the private mapCarrier constructor for testing.
func ExportNewMapCarrier(
	data map[string]interface{},
) propagation.TextMapCarrier {
	return mapCarrier{data: data}
}

// noopClient implements otlptrace.Client for testing.
type noopClient struct{}

func (noopClient) Start(_ context.Context) error { return nil }
func (noopClient) Stop(_ context.Context) error  { return nil }
func (noopClient) UploadTraces(
	_ context.Context,
	_ []*tracepb.ResourceSpans,
) error {
	return nil
}

// SetResourceNewFn overrides the resourceNewFn injectable for testing.
func SetResourceNewFn(
	fn func(context.Context, ...resource.Option) (*resource.Resource, error),
) {
	resourceNewFn = fn
}

// ResetResourceNewFn restores the default resourceNewFn.
func ResetResourceNewFn() {
	resourceNewFn = resource.New
}

// SetStdouttraceNewFn overrides the stdouttraceNewFn injectable for testing.
func SetStdouttraceNewFn(
	fn func(...stdouttrace.Option) (*stdouttrace.Exporter, error),
) {
	stdouttraceNewFn = fn
}

// ResetStdouttraceNewFn restores the default stdouttraceNewFn.
func ResetStdouttraceNewFn() {
	stdouttraceNewFn = stdouttrace.New
}

// SetOtlptraceNewFn overrides the otlptraceNewFn injectable for testing.
func SetOtlptraceNewFn(
	fn func(context.Context, ...otlptracegrpc.Option) (*otlptrace.Exporter, error),
) {
	otlptraceNewFn = fn
}

// ResetOtlptraceNewFn restores the default otlptraceNewFn.
func ResetOtlptraceNewFn() {
	otlptraceNewFn = otlptracegrpc.New
}

// ExportNewNoopOTLPExporter creates an OTLP exporter backed by the noop client
// for testing.
func ExportNewNoopOTLPExporter() *otlptrace.Exporter {
	return otlptrace.NewUnstarted(noopClient{})
}

// ExportNewResourceError returns a resourceNewFn that always errors.
func ExportNewResourceError() func(context.Context, ...resource.Option) (*resource.Resource, error) {
	return func(_ context.Context, _ ...resource.Option) (*resource.Resource, error) {
		return nil, errors.New("resource creation failed")
	}
}

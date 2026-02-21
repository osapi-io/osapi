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

// Package telemetry provides OpenTelemetry tracing initialization and helpers.
package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/retr0h/osapi/internal/config"
)

// InitTracer initializes the OpenTelemetry tracer provider.
// It always sets the global W3C TraceContext propagator.
// When tracing is disabled, a noop provider is used.
// Returns a shutdown function that must be called on exit.
func InitTracer(
	ctx context.Context,
	serviceName string,
	cfg config.TracingConfig,
) (func(context.Context) error, error) {
	// Always set the W3C propagator so inject/extract works
	otel.SetTextMapPropagator(propagation.TraceContext{})

	if !cfg.Enabled {
		otel.SetTracerProvider(noop.NewTracerProvider())

		return func(_ context.Context) error { return nil }, nil
	}

	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("creating resource: %w", err)
	}

	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(res),
	}

	// Only add an exporter when explicitly configured.
	// With no exporter, spans are still created (so trace_id appears in logs)
	// but nothing is dumped to stdout or sent over the wire.
	switch cfg.Exporter {
	case "", "none":
		// No exporter — log-only trace correlation
	case "stdout":
		exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, fmt.Errorf("creating stdout exporter: %w", err)
		}

		opts = append(opts, sdktrace.WithBatcher(exp))
	case "otlp":
		exp, err := otlptracegrpc.New(
			ctx,
			otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
			otlptracegrpc.WithInsecure(),
		)
		if err != nil {
			return nil, fmt.Errorf("creating OTLP exporter: %w", err)
		}

		opts = append(opts, sdktrace.WithBatcher(exp))
	default:
		return nil, fmt.Errorf("unsupported tracing exporter: %q", cfg.Exporter)
	}

	tp := sdktrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(tp)

	return tp.Shutdown, nil
}

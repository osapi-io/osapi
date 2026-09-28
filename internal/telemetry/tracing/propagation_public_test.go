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

package tracing_test

import (
	"context"
	"net/http"
	"testing"

	"go.opentelemetry.io/otel/trace"

	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/osapi-io/osapi/internal/telemetry/tracing"
)

type PropagationPublicTestSuite struct {
	suite.Suite

	ctx context.Context
}

func (s *PropagationPublicTestSuite) SetupTest() {
	s.ctx = context.Background()

	// Set up a real tracer provider and propagator for tests
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
}

func (s *PropagationPublicTestSuite) TestMapCarrierGet() {
	tests := []struct {
		name         string
		data         map[string]interface{}
		key          string
		validateFunc func(string)
	}{
		{
			name: "when key exists returns value",
			data: map[string]interface{}{
				"traceparent": "00-abc123-def456-01",
			},
			key: "traceparent",
			validateFunc: func(result string) {
				s.Equal("00-abc123-def456-01", result)
			},
		},
		{
			name: "when key does not exist returns empty string",
			data: map[string]interface{}{},
			key:  "traceparent",
			validateFunc: func(result string) {
				s.Empty(result)
			},
		},
		{
			name: "when value is not a string returns empty string",
			data: map[string]interface{}{
				"count": 42,
			},
			key: "count",
			validateFunc: func(result string) {
				s.Empty(result)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			c := tracing.ExportNewMapCarrier(tc.data)

			tc.validateFunc(c.Get(tc.key))
		})
	}
}

func (s *PropagationPublicTestSuite) TestMapCarrierSet() {
	tests := []struct {
		name         string
		key          string
		value        string
		validateFunc func(map[string]interface{})
	}{
		{
			name:  "when setting a value stores it in data",
			key:   "traceparent",
			value: "00-abc123-def456-01",
			validateFunc: func(data map[string]interface{}) {
				s.Equal("00-abc123-def456-01", data["traceparent"])
			},
		},
		{
			name:  "when setting overwrites existing value",
			key:   "traceparent",
			value: "00-new-value-01",
			validateFunc: func(data map[string]interface{}) {
				s.Equal("00-new-value-01", data["traceparent"])
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			data := map[string]interface{}{
				"traceparent": "00-old-value-01",
			}
			c := tracing.ExportNewMapCarrier(data)
			c.Set(tc.key, tc.value)

			tc.validateFunc(data)
		})
	}
}

func (s *PropagationPublicTestSuite) TestMapCarrierKeys() {
	tests := []struct {
		name         string
		data         map[string]interface{}
		validateFunc func([]string)
	}{
		{
			name: "when data has entries returns all keys",
			data: map[string]interface{}{
				"traceparent":   "00-abc123-def456-01",
				"tracestate":    "vendor=opaque",
				"custom-header": "value",
			},
			validateFunc: func(keys []string) {
				s.Len(keys, 3)
				s.ElementsMatch(
					[]string{"traceparent", "tracestate", "custom-header"},
					keys,
				)
			},
		},
		{
			name: "when data is empty returns empty slice",
			data: map[string]interface{}{},
			validateFunc: func(keys []string) {
				s.Empty(keys)
			},
		},
		{
			name: "when data has single entry returns one key",
			data: map[string]interface{}{
				"traceparent": "00-abc123-def456-01",
			},
			validateFunc: func(keys []string) {
				s.Len(keys, 1)
				s.Equal("traceparent", keys[0])
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			c := tracing.ExportNewMapCarrier(tc.data)

			tc.validateFunc(c.Keys())
		})
	}
}

func TestPropagationPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(PropagationPublicTestSuite))
}

// TestExtractFromHeader covers the header path, which the NATS consumer uses on
// every job: JetStream delivers header keys in whatever casing the publisher used,
// and http.Header.Get only finds the canonical form.
func (s *PropagationPublicTestSuite) TestExtractFromHeader() {
	const traceID = "4bf92f3577b34da6a3ce929d0e0e4736"

	tests := []struct {
		name         string
		header       http.Header
		validateFunc func(context.Context)
	}{
		{
			name: "a canonical traceparent",
			header: http.Header{
				"Traceparent": []string{
					"00-" + traceID + "-00f067aa0ba902b7-01",
				},
			},
			validateFunc: func(ctx context.Context) {
				sc := trace.SpanContextFromContext(ctx)
				s.Require().True(sc.IsValid())
				s.Equal(traceID, sc.TraceID().String())
			},
		},
		{
			name: "a lowercase traceparent, which is what JetStream delivers",
			header: http.Header{
				"traceparent": []string{
					"00-" + traceID + "-00f067aa0ba902b7-01",
				},
			},
			validateFunc: func(ctx context.Context) {
				sc := trace.SpanContextFromContext(ctx)
				s.Require().True(sc.IsValid())
				s.Equal(traceID, sc.TraceID().String())
			},
		},
		{
			name:   "no trace context at all",
			header: http.Header{},
			validateFunc: func(ctx context.Context) {
				sc := trace.SpanContextFromContext(ctx)
				s.False(sc.IsValid())
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.validateFunc(tracing.ExtractTraceContextFromHeader(s.ctx, tt.header))
		})
	}
}

// TestExtractTraceContext covers the map path, which the agent uses when a job
// carries its trace context in the payload rather than in headers.
func (s *PropagationPublicTestSuite) TestExtractTraceContext() {
	const traceID = "4bf92f3577b34da6a3ce929d0e0e4736"

	tests := []struct {
		name         string
		data         map[string]interface{}
		validateFunc func(context.Context)
	}{
		{
			name: "a traceparent in the payload",
			data: map[string]interface{}{
				"traceparent": "00-" + traceID + "-00f067aa0ba902b7-01",
			},
			validateFunc: func(ctx context.Context) {
				sc := trace.SpanContextFromContext(ctx)
				s.Require().True(sc.IsValid())
				s.Equal(traceID, sc.TraceID().String())
			},
		},
		{
			name: "no trace context at all",
			data: map[string]interface{}{},
			validateFunc: func(ctx context.Context) {
				sc := trace.SpanContextFromContext(ctx)
				s.False(sc.IsValid())
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.validateFunc(tracing.ExtractTraceContext(s.ctx, tt.data))
		})
	}
}

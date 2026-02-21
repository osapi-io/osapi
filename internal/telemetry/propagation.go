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
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// Compile-time check that mapCarrier satisfies the TextMapCarrier interface.
var _ propagation.TextMapCarrier = mapCarrier{}

// mapCarrier implements propagation.TextMapCarrier for map[string]interface{}.
type mapCarrier struct {
	data map[string]interface{}
}

// Get returns the value for the key.
func (c mapCarrier) Get(
	key string,
) string {
	v, ok := c.data[key].(string)
	if !ok {
		return ""
	}

	return v
}

// Set stores a key-value pair.
func (c mapCarrier) Set(
	key string,
	value string,
) {
	c.data[key] = value
}

// Keys returns all keys in the carrier.
func (c mapCarrier) Keys() []string {
	keys := make([]string, 0, len(c.data))
	for k := range c.data {
		keys = append(keys, k)
	}

	return keys
}

// InjectTraceContext injects the current span's trace context into a data map.
// If there is no active span, this is a no-op.
func InjectTraceContext(
	ctx context.Context,
	data map[string]interface{},
) {
	otel.GetTextMapPropagator().Inject(ctx, mapCarrier{data: data})
}

// ExtractTraceContext extracts trace context from a data map and returns
// a new context with the extracted span context. If no trace context is
// present, the original context is returned.
func ExtractTraceContext(
	ctx context.Context,
	data map[string]interface{},
) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, mapCarrier{data: data})
}

// InjectTraceContextToHeader injects the current span's trace context into
// HTTP-compatible headers (usable with nats.Header via type conversion).
func InjectTraceContextToHeader(
	ctx context.Context,
	header http.Header,
) {
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(header))
}

// ExtractTraceContextFromHeader extracts trace context from HTTP-compatible
// headers and returns a new context. If no trace context is present, the
// original context is returned.
//
// Header keys are normalized to canonical MIME format before extraction.
// Some transports (e.g., NATS JetStream) deliver headers with non-canonical
// casing (lowercase "traceparent" instead of "Traceparent"), which breaks
// http.Header.Get lookups. Normalizing ensures reliable extraction.
func ExtractTraceContextFromHeader(
	ctx context.Context,
	header http.Header,
) context.Context {
	normalized := make(http.Header, len(header))
	for k, v := range header {
		normalized[http.CanonicalHeaderKey(k)] = v
	}

	return otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(normalized))
}

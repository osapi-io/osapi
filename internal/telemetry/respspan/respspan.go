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

// Package respspan times how long the controller takes to encode a response
// as JSON and write it, as a span of its own under the request. The code
// that oapi-codegen generates encodes after every middleware has returned,
// so no middleware can open a span around the encode. Two marks time it
// instead: Handled, when the handler has its answer, and the last write of
// the body, which echo reports through Response.After.
package respspan

import (
	"time"

	"github.com/labstack/echo/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// SpanName is the span of the encode and the write of a response.
const SpanName = "response.encode"

// AttrBodySize is the size of the body that the controller wrote, in bytes.
const AttrBodySize = attribute.Key("http.response.body.size")

const handledKey = "osapi.response.handled"

// Handled marks the time that the handler of c has its answer. What the
// controller does after it, until the last write, is the encode and the
// write of the response. ScopeMiddleware calls it for each strict handler;
// a handler that encodes itself calls it before c.JSON. The first mark wins:
// a handler that writes its own answer inside ScopeMiddleware has marked
// already, and the mark after its write would end the span before it starts.
func Handled(
	c *echo.Context,
) {
	if _, marked := c.Get(handledKey).(time.Time); marked {
		return
	}
	c.Set(handledKey, time.Now())
}

// Middleware opens the response.encode span of a request, from the time
// Handled marks to the last write of the body, under the span of the
// request. A request without a span, without the mark, or without a body
// gets none. It goes after the middleware that opens the request span.
func Middleware(
	next echo.HandlerFunc,
) echo.HandlerFunc {
	return func(c *echo.Context) error {
		resp, unwrapErr := echo.UnwrapResponse(c.Response())
		var last time.Time
		if unwrapErr == nil {
			resp.After(func() { last = time.Now() })
		}

		err := next(c)

		ctx := c.Request().Context()
		start, marked := c.Get(handledKey).(time.Time)
		if marked && !last.IsZero() && !last.Before(start) &&
			trace.SpanContextFromContext(ctx).IsValid() {
			_, span := otel.Tracer("osapi-controller").Start(ctx, SpanName,
				trace.WithTimestamp(start),
				trace.WithAttributes(AttrBodySize.Int64(resp.Size)),
			)
			span.End(trace.WithTimestamp(last))
		}

		return err
	}
}

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

// Package respspan times what the controller does after a handler has its
// answer, as two spans of their own under the request: response.encode, the
// encode of the answer as JSON, and response.write, the write of it to the
// client, which waits on the network. The code that oapi-codegen generates
// encodes after every middleware has returned, so no middleware can open a
// span around the encode. Three marks time them instead: Handled, when the
// handler has its answer; echo's Response.Before, when the status line goes
// out, after the encode; and echo's Response.After, the last write.
package respspan

import (
	"time"

	"github.com/labstack/echo/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// The spans after a handler has its answer.
const (
	// SpanEncode is the encode of the answer as JSON.
	SpanEncode = "response.encode"
	// SpanWrite is the write of the body to the client. It waits on the
	// network, so a slow client makes it long, not the controller.
	SpanWrite = "response.write"
)

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

// Middleware opens the response.encode and response.write spans of a
// request under the span of the request: response.encode from the time
// Handled marks to the status line, and response.write from there to the last
// write of the body, with its size. A request without a span, without the
// mark, or without a body gets neither, and no span ends before it starts.
// It goes after the middleware that opens the request span.
func Middleware(
	next echo.HandlerFunc,
) echo.HandlerFunc {
	return func(c *echo.Context) error {
		resp, unwrapErr := echo.UnwrapResponse(c.Response())
		var header, last time.Time
		if unwrapErr == nil {
			resp.Before(func() { header = time.Now() })
			resp.After(func() { last = time.Now() })
		}

		err := next(c)

		ctx := c.Request().Context()
		start, marked := c.Get(handledKey).(time.Time)
		if !marked || last.IsZero() || header.Before(start) || last.Before(header) ||
			!trace.SpanContextFromContext(ctx).IsValid() {
			return err
		}

		tracer := otel.Tracer("osapi-controller")
		_, encode := tracer.Start(ctx, SpanEncode, trace.WithTimestamp(start))
		encode.End(trace.WithTimestamp(header))
		_, write := tracer.Start(ctx, SpanWrite,
			trace.WithTimestamp(header),
			trace.WithAttributes(AttrBodySize.Int64(resp.Size)),
		)
		write.End(trace.WithTimestamp(last))

		return err
	}
}

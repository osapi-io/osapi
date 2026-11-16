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

package respspan_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	"github.com/osapi-io/osapi/internal/telemetry/respspan"
)

type RespspanPublicTestSuite struct {
	suite.Suite

	spans *tracetest.SpanRecorder
}

func (s *RespspanPublicTestSuite) SetupTest() {
	s.spans = tracetest.NewSpanRecorder()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(s.spans)))
}

func (s *RespspanPublicTestSuite) TestMiddleware() {
	tests := []struct {
		name     string
		parent   bool
		plain    bool
		handler  func(c *echo.Context) error
		wantSpan bool
	}{
		{
			name:   "when the handler marks its answer and writes a body there is a span",
			parent: true,
			handler: func(c *echo.Context) error {
				respspan.Handled(c)

				return c.JSON(http.StatusOK, map[string]string{"key": "value"})
			},
			wantSpan: true,
		},
		{
			name:   "when the handler does not mark its answer there is none",
			parent: true,
			handler: func(c *echo.Context) error {
				return c.JSON(http.StatusOK, map[string]string{"key": "value"})
			},
		},
		{
			name:   "when the response has no body there is none",
			parent: true,
			handler: func(c *echo.Context) error {
				respspan.Handled(c)

				return c.NoContent(http.StatusNoContent)
			},
		},
		{
			name: "when the request has no span there is none",
			handler: func(c *echo.Context) error {
				respspan.Handled(c)

				return c.JSON(http.StatusOK, map[string]string{"key": "value"})
			},
		},
		{
			name:   "when the response is not echo's there is none",
			parent: true,
			plain:  true,
			handler: func(c *echo.Context) error {
				respspan.Handled(c)
				_, err := c.Response().Write([]byte("{}"))

				return err
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()
			ctx := context.Background()
			if tc.parent {
				c, parent := otel.Tracer("test").Start(ctx, "GET /x")
				defer parent.End()
				ctx = c
			}

			req := httptest.NewRequest(http.MethodGet, "/x", nil).WithContext(ctx)
			rec := httptest.NewRecorder()
			c := echo.New().NewContext(req, rec)
			if tc.plain {
				c.SetResponse(httptest.NewRecorder())
			}

			s.Require().NoError(respspan.Middleware(tc.handler)(c))

			var encode []sdktrace.ReadOnlySpan
			for _, sp := range s.spans.Ended() {
				if sp.Name() == respspan.SpanName {
					encode = append(encode, sp)
				}
			}
			if !tc.wantSpan {
				s.Empty(encode)

				return
			}
			s.Require().Len(encode, 1)
			sp := encode[0]
			s.Equal(trace.SpanContextFromContext(ctx).SpanID(), sp.Parent().SpanID())
			s.False(sp.EndTime().Before(sp.StartTime()))
			var size int64
			for _, kv := range sp.Attributes() {
				if kv.Key == respspan.AttrBodySize {
					size = kv.Value.AsInt64()
				}
			}
			s.Equal(int64(len(rec.Body.Bytes())), size, "the size is the body written")
		})
	}
}

func TestRespspanPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(RespspanPublicTestSuite))
}

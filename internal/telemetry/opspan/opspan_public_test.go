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

package opspan_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	"github.com/osapi-io/osapi/internal/telemetry/opspan"
)

type OpSpanPublicTestSuite struct {
	suite.Suite

	recorder *tracetest.SpanRecorder
	previous trace.TracerProvider
}

func TestOpSpanPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(OpSpanPublicTestSuite))
}

func (s *OpSpanPublicTestSuite) SetupSubTest() {
	s.recorder = tracetest.NewSpanRecorder()
	s.previous = otel.GetTracerProvider()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(s.recorder)))
}

func (s *OpSpanPublicTestSuite) TearDownSubTest() {
	otel.SetTracerProvider(s.previous)
}

// TestOperation opens an operation's span under the caller's, with its
// attributes, and gives the caller a context whose calls are its children.
func (s *OpSpanPublicTestSuite) TestOperation() {
	tests := []struct {
		name string
		err  error
		code codes.Code
	}{
		{name: "when the operation succeeds leaves the status unset", code: codes.Unset},
		{
			name: "when the operation fails marks the span as an error",
			err:  errors.New("kv down"),
			code: codes.Error,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			ctx, request := otel.Tracer("test").Start(context.Background(), "GET /api/x")
			opCtx, span := opspan.Start(ctx, "osapi-test", "x.list", attribute.Int("x.count", 2))
			_, child := otel.Tracer("test").Start(opCtx, "kv.keys")
			child.End()
			s.Equal(tc.err, opspan.End(span, tc.err), "End returns the error it was given")
			request.End()

			byName := map[string]sdktrace.ReadOnlySpan{}
			for _, sp := range s.recorder.Ended() {
				byName[sp.Name()] = sp
			}
			op := byName["x.list"]
			s.Require().NotNil(op)
			s.Equal(request.SpanContext().SpanID(), op.Parent().SpanID())
			s.Equal(op.SpanContext().SpanID(), byName["kv.keys"].Parent().SpanID())
			s.Equal("osapi-test", op.InstrumentationScope().Name)
			s.Contains(op.Attributes(), attribute.Int("x.count", 2))
			s.Equal(tc.code, op.Status().Code)
		})
	}
}

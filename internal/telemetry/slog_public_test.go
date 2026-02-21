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

package telemetry_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/retr0h/osapi/internal/telemetry"
)

type SlogPublicTestSuite struct {
	suite.Suite

	ctx context.Context
}

func (s *SlogPublicTestSuite) SetupTest() {
	s.ctx = context.Background()

	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
}

func (s *SlogPublicTestSuite) TestNewTraceHandler() {
	tests := []struct {
		name         string
		setupCtx     func() context.Context
		validateFunc func(output string)
	}{
		{
			name: "when active span adds trace_id and span_id",
			setupCtx: func() context.Context {
				ctx, _ := otel.Tracer("test").Start(s.ctx, "test-span")

				return ctx
			},
			validateFunc: func(output string) {
				s.Contains(output, "trace_id=")
				s.Contains(output, "span_id=")
			},
		},
		{
			name: "when no active span does not add trace fields",
			setupCtx: func() context.Context {
				return context.Background()
			},
			validateFunc: func(output string) {
				s.NotContains(output, "trace_id=")
				s.NotContains(output, "span_id=")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var buf bytes.Buffer
			inner := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
			handler := telemetry.NewTraceHandler(inner)
			logger := slog.New(handler)

			ctx := tc.setupCtx()
			logger.InfoContext(ctx, "test message")

			tc.validateFunc(buf.String())
		})
	}
}

func (s *SlogPublicTestSuite) TestTraceHandlerWithAttrs() {
	var buf bytes.Buffer
	inner := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	handler := telemetry.NewTraceHandler(inner)

	withAttrs := handler.WithAttrs([]slog.Attr{slog.String("key", "value")})
	s.NotNil(withAttrs)

	logger := slog.New(withAttrs)
	ctx, _ := otel.Tracer("test").Start(s.ctx, "test-span")
	logger.InfoContext(ctx, "test")
	s.Contains(buf.String(), "key=value")
	s.Contains(buf.String(), "trace_id=")
}

func (s *SlogPublicTestSuite) TestTraceHandlerWithGroup() {
	var buf bytes.Buffer
	inner := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	handler := telemetry.NewTraceHandler(inner)

	withGroup := handler.WithGroup("mygroup")
	s.NotNil(withGroup)

	logger := slog.New(withGroup)
	ctx, _ := otel.Tracer("test").Start(s.ctx, "test-span")
	logger.InfoContext(ctx, "test", slog.String("field", "value"))
	s.Contains(buf.String(), "mygroup.field=value")
}

func (s *SlogPublicTestSuite) TestTraceHandlerEnabled() {
	inner := slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelWarn})
	handler := telemetry.NewTraceHandler(inner)

	s.False(handler.Enabled(s.ctx, slog.LevelDebug))
	s.True(handler.Enabled(s.ctx, slog.LevelWarn))
}

func (s *SlogPublicTestSuite) TestTraceHandlerPreservesTraceID() {
	var buf bytes.Buffer
	inner := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	handler := telemetry.NewTraceHandler(inner)
	logger := slog.New(handler)

	ctx, span := otel.Tracer("test").Start(s.ctx, "test-span")
	defer span.End()

	expectedTraceID := trace.SpanContextFromContext(ctx).TraceID().String()
	logger.InfoContext(ctx, "check trace id")

	s.Contains(buf.String(), expectedTraceID)
}

func TestSlogPublicTestSuite(t *testing.T) {
	suite.Run(t, new(SlogPublicTestSuite))
}

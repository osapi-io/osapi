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

package audit_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/audit"
	"github.com/osapi-io/osapi/internal/audit/mocks"
)

// SpanPublicTestSuite checks that each read of the audit store has its own
// span under the request, with the stream calls inside it.
type SpanPublicTestSuite struct {
	suite.Suite
}

func TestSpanPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(SpanPublicTestSuite))
}

func (s *SpanPublicTestSuite) TestReads() {
	tests := []struct {
		name string
		span string
		call func(st *audit.StreamStore, ctx context.Context) error
	}{
		{
			name: "when a page is listed opens audit.list",
			span: "audit.list",
			call: func(st *audit.StreamStore, ctx context.Context) error {
				_, _, err := st.List(ctx, 10, 0)
				return err
			},
		},
		{
			name: "when one entry is read opens audit.get",
			span: "audit.get",
			call: func(st *audit.StreamStore, ctx context.Context) error {
				_, err := st.Get(ctx, "e1")
				return err
			},
		},
		{
			name: "when every entry is exported opens audit.export",
			span: "audit.export",
			call: func(st *audit.StreamStore, ctx context.Context) error {
				_, err := st.ListAll(ctx)
				return err
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			recorder := tracetest.NewSpanRecorder()
			previous := otel.GetTracerProvider()
			otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder)))
			defer otel.SetTracerProvider(previous)

			stream := mocks.NewMockStream(gomock.NewController(s.T()))
			var inside []context.Context
			stream.EXPECT().StreamInfo(gomock.Any(), "AUDIT").
				DoAndReturn(func(ctx context.Context, _ string) (*jetstream.StreamInfo, error) {
					inside = append(inside, ctx)
					return nil, errors.New("stream down")
				}).AnyTimes()
			stream.EXPECT().StreamLastMsg(gomock.Any(), "AUDIT", gomock.Any()).
				DoAndReturn(func(ctx context.Context, _, _ string) (*jetstream.RawStreamMsg, error) {
					inside = append(inside, ctx)
					return nil, errors.New("stream down")
				}).
				AnyTimes()
			store := audit.NewStreamStore(slog.Default(), stream, "AUDIT", "audit.log")

			ctx, request := otel.Tracer("test").Start(context.Background(), "GET /api/audit")
			s.Error(tc.call(store, ctx))
			request.End()

			var span sdktrace.ReadOnlySpan
			for _, sp := range recorder.Ended() {
				if sp.Name() == tc.span {
					span = sp
				}
			}
			s.Require().NotNil(span)
			s.Equal(request.SpanContext().SpanID(), span.Parent().SpanID())
			s.Equal(codes.Error, span.Status().Code)
			s.Require().NotEmpty(inside)
			for _, c := range inside {
				s.Equal(span.SpanContext().SpanID(), trace.SpanContextFromContext(c).SpanID())
			}
		})
	}
}

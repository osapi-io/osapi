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

package enrollment_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/controller/enrollment"
	enrollMocks "github.com/osapi-io/osapi/internal/controller/enrollment/mocks"
)

// SpanPublicTestSuite checks that each enrollment operation the API calls
// has its own span under the request, with its KV calls inside it.
type SpanPublicTestSuite struct {
	suite.Suite
}

func TestSpanPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(SpanPublicTestSuite))
}

func (s *SpanPublicTestSuite) TestOperations() {
	tests := []struct {
		name string
		span string
		call func(w *enrollment.Watcher, ctx context.Context) error
	}{
		{
			name: "when the pending agents are listed opens enrollment.list",
			span: "enrollment.list",
			call: func(w *enrollment.Watcher, ctx context.Context) error {
				_, err := w.ListPending(ctx)
				return err
			},
		},
		{
			name: "when an agent is accepted by hostname opens enrollment.accept",
			span: "enrollment.accept",
			call: func(w *enrollment.Watcher, ctx context.Context) error {
				return w.AcceptByHostname(ctx, "web-01")
			},
		},
		{
			name: "when an agent is accepted by fingerprint opens enrollment.accept",
			span: "enrollment.accept",
			call: func(w *enrollment.Watcher, ctx context.Context) error {
				return w.AcceptByFingerprint(ctx, "SHA256:ab")
			},
		},
		{
			name: "when an agent is rejected opens enrollment.reject",
			span: "enrollment.reject",
			call: func(w *enrollment.Watcher, ctx context.Context) error {
				return w.RejectByHostname(ctx, "web-01", "no")
			},
		},
		{
			name: "when an enrolled agent is removed opens enrollment.remove",
			span: "enrollment.remove",
			call: func(w *enrollment.Watcher, ctx context.Context) error {
				_, err := w.RemoveByHostname(ctx, "web-01", "")
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

			ctrl := gomock.NewController(s.T())
			kv := enrollMocks.NewMockKVStore(ctrl)
			var inside []context.Context
			kv.EXPECT().KVKeys(gomock.Any(), "enrollment").
				DoAndReturn(func(ctx context.Context, _ string, _ ...string) ([]string, error) {
					inside = append(inside, ctx)
					return nil, errors.New("kv down")
				}).AnyTimes()
			w := enrollment.NewWatcher(slog.Default(), enrollMocks.NewMockNATSSubscriber(ctrl),
				kv, "enrollment", enrollMocks.NewMockPKIProvider(ctrl), false, "osapi")

			ctx, request := otel.Tracer("test").Start(context.Background(), "POST /api/agent")
			s.Error(tc.call(w, ctx))
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

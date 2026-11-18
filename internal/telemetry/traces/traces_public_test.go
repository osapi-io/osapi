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

package traces_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/osapi-io/osapi/internal/telemetry/traces"
)

// answer is OTLP JSON as Jaeger answers it: two services, kinds and status
// codes as names and as numbers, and an array attribute.
const answer = `{"result":{"resourceSpans":[
 {"resource":{"attributes":[{"key":"service.name","value":{"stringValue":"osapi-agent"}}]},
  "scopeSpans":[{"spans":[
   {"spanId":"b","parentSpanId":"a","name":"job.process","kind":5,
    "startTimeUnixNano":"2000000","endTimeUnixNano":"5000000",
    "attributes":[{"key":"osapi.agent.hostname","value":{"stringValue":"web-01"}},
                  {"key":"job.changed","value":{"boolValue":true}},
                  {"key":"count","value":{"intValue":"3"}},
                  {"key":"tags","value":{"arrayValue":{"values":[{"stringValue":"a"},{"stringValue":"b"}]}}}],
    "status":{"code":2,"message":"boom"}}]}]},
 {"resource":{"attributes":[{"key":"service.name","value":{"stringValue":"osapi-controller"}}]},
  "scopeSpans":[{"spans":[
   {"spanId":"a","name":"GET /api/x","kind":"SPAN_KIND_SERVER",
    "startTimeUnixNano":"1000000","endTimeUnixNano":"9000000",
    "status":{"code":"STATUS_CODE_UNSET"}},
   {"spanId":"c","parentSpanId":"a","name":"odd","kind":9,
    "startTimeUnixNano":"3000000","endTimeUnixNano":"3000000",
    "status":{"code":"STATUS_CODE_ERROR"}}]}]}]}}`

type TracesPublicTestSuite struct {
	suite.Suite
}

func (s *TracesPublicTestSuite) TestTrace() {
	tests := []struct {
		name     string
		status   int
		body     string
		validate func(*traces.Trace, error)
	}{
		{
			name:   "returns each span in start order, with its service and attributes",
			status: http.StatusOK,
			body:   answer,
			validate: func(t *traces.Trace, err error) {
				s.Require().NoError(err)
				s.Equal("0123456789abcdef0123456789abcdef", t.ID)
				s.Require().Len(t.Spans, 3)

				root, process, odd := t.Spans[0], t.Spans[1], t.Spans[2]
				s.Equal("GET /api/x", root.Name)
				s.Equal("osapi-controller", root.Service)
				s.Equal("server", root.Kind)
				s.Empty(root.ParentID)
				s.Equal(8*time.Millisecond, root.Duration)
				s.False(root.Failed)

				s.Equal("job.process", process.Name)
				s.Equal("a", process.ParentID)
				s.Equal("osapi-agent", process.Service)
				s.Equal("consumer", process.Kind)
				s.True(process.Failed)
				s.Equal("boom", process.StatusMessage)
				s.Equal(map[string]string{
					"osapi.agent.hostname": "web-01",
					"job.changed":          "true",
					"count":                "3",
					"tags":                 "a, b",
				}, process.Attributes)
				s.Equal(time.Unix(0, 2000000).UTC(), process.Start)

				s.Empty(odd.Kind, "a kind osapi does not know is empty")
				s.True(odd.Failed)
			},
		},
		{
			name:   "when Jaeger has no such trace returns ErrNotFound",
			status: http.StatusNotFound,
			body:   `{"error":{"httpCode":404,"message":"No traces found"}}`,
			validate: func(_ *traces.Trace, err error) {
				s.ErrorIs(err, traces.ErrNotFound)
			},
		},
		{
			name:   "when the answer has no spans returns ErrNotFound",
			status: http.StatusOK,
			body:   `{"result":{"resourceSpans":[]}}`,
			validate: func(_ *traces.Trace, err error) {
				s.ErrorIs(err, traces.ErrNotFound)
			},
		},
		{
			name:   "when Jaeger fails returns the status",
			status: http.StatusBadGateway,
			validate: func(_ *traces.Trace, err error) {
				s.EqualError(
					err,
					"read trace 0123456789abcdef0123456789abcdef: jaeger answered 502",
				)
			},
		},
		{
			name:   "when the answer is not JSON returns the error",
			status: http.StatusOK,
			body:   "not json",
			validate: func(_ *traces.Trace, err error) {
				s.ErrorContains(
					err,
					"read trace 0123456789abcdef0123456789abcdef: invalid character",
				)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			srv := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					s.Equal("/api/v3/traces/0123456789abcdef0123456789abcdef", r.URL.Path)
					w.WriteHeader(tc.status)
					_, _ = w.Write([]byte(tc.body))
				}),
			)
			defer srv.Close()

			tc.validate(traces.NewJaeger(srv.URL+"/", nil).
				Trace(context.Background(), "0123456789abcdef0123456789abcdef"))
		})
	}
}

// TestReadSpan opens traces.read for the HTTP call to Jaeger, which no
// other span shows: with the trace ID and the span count, and the error
// status when the read fails.
func (s *TracesPublicTestSuite) TestReadSpan() {
	tests := []struct {
		name     string
		status   int
		validate func(span sdktrace.ReadOnlySpan)
	}{
		{
			name:   "when Jaeger has the trace records how many spans it read",
			status: http.StatusOK,
			validate: func(span sdktrace.ReadOnlySpan) {
				s.Contains(span.Attributes(), attribute.Int("osapi.trace.spans", 3))
				s.Equal(codes.Unset, span.Status().Code)
			},
		},
		{
			name:   "when Jaeger fails marks traces.read as an error",
			status: http.StatusInternalServerError,
			validate: func(span sdktrace.ReadOnlySpan) {
				s.Equal(codes.Error, span.Status().Code)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			recorder := tracetest.NewSpanRecorder()
			previous := otel.GetTracerProvider()
			otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder)))
			defer otel.SetTracerProvider(previous)

			srv := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(tc.status)
					_, _ = w.Write([]byte(answer))
				}),
			)
			defer srv.Close()

			ctx, request := otel.Tracer("test").
				Start(context.Background(), "GET /api/telemetry/trace")
			_, _ = traces.NewJaeger(srv.URL, nil).Trace(ctx, "0123456789abcdef0123456789abcdef")
			request.End()

			var read sdktrace.ReadOnlySpan
			for _, sp := range recorder.Ended() {
				if sp.Name() == "traces.read" {
					read = sp
				}
			}
			s.Require().NotNil(read)
			s.Equal(request.SpanContext().SpanID(), read.Parent().SpanID())
			s.Contains(read.Attributes(),
				attribute.String("osapi.trace.id", "0123456789abcdef0123456789abcdef"))
			tc.validate(read)
		})
	}
}

func (s *TracesPublicTestSuite) TestTraceTransport() {
	s.Run("when the request cannot be made returns the error", func() {
		_, err := traces.NewJaeger("http://bad host", nil).Trace(context.Background(), "x")
		s.ErrorContains(err, "read trace x")
	})

	s.Run("when Jaeger cannot be reached returns the error", func() {
		srv := httptest.NewServer(http.NotFoundHandler())
		srv.Close()

		_, err := traces.NewJaeger(srv.URL, http.DefaultClient).Trace(context.Background(), "x")
		s.ErrorContains(err, "read trace x")
	})

	s.Run("when the answer ends early returns the error", func() {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Length", "100")
			_, _ = w.Write([]byte("{"))
		}))
		defer srv.Close()

		_, err := traces.NewJaeger(srv.URL, nil).Trace(context.Background(), "x")
		s.ErrorContains(err, "read trace x")
	})
}

func TestTracesPublicTestSuite(t *testing.T) {
	suite.Run(t, new(TracesPublicTestSuite))
}

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

package client_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/job/client"
	jobmocks "github.com/osapi-io/osapi/internal/job/mocks"
)

// SpanPublicTestSuite checks the controller half of the span tree of a
// request: under the span of the request, job.publish and job.collect, and
// the trace context that goes with the job is job.publish. The agent half,
// job.process under that context, is in internal/agent.
type SpanPublicTestSuite struct {
	suite.Suite

	ctrl     *gomock.Controller
	nats     *jobmocks.MockNATSClient
	recorder *tracetest.SpanRecorder
	previous trace.TracerProvider
}

func TestSpanPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(SpanPublicTestSuite))
}

func (s *SpanPublicTestSuite) SetupSubTest() {
	s.ctrl = gomock.NewController(s.T())
	s.nats = jobmocks.NewMockNATSClient(s.ctrl)
	s.recorder = tracetest.NewSpanRecorder()
	s.previous = otel.GetTracerProvider()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(s.recorder)))
}

func (s *SpanPublicTestSuite) TearDownSubTest() {
	otel.SetTracerProvider(s.previous)
}

// spans are the ended spans by name.
func (s *SpanPublicTestSuite) spans() map[string]sdktrace.ReadOnlySpan {
	out := map[string]sdktrace.ReadOnlySpan{}
	for _, sp := range s.recorder.Ended() {
		out[sp.Name()] = sp
	}

	return out
}

func (s *SpanPublicTestSuite) TestQuery() {
	tests := []struct {
		name     string
		response string
		timeout  bool
		validate func(spans map[string]sdktrace.ReadOnlySpan, request trace.SpanContext, sent trace.SpanContext)
	}{
		{
			name:     "when the agent answers makes publish and collect under the request",
			response: `{"status":"completed","hostname":"server1"}`,
			validate: func(spans map[string]sdktrace.ReadOnlySpan, request trace.SpanContext, sent trace.SpanContext) {
				pub, col := spans["job.publish"], spans["job.collect"]
				s.Require().NotNil(pub)
				s.Require().NotNil(col)

				s.Equal(request.SpanID(), pub.Parent().SpanID())
				s.Equal(request.SpanID(), col.Parent().SpanID())
				s.Equal(trace.SpanKindProducer, pub.SpanKind())
				s.Equal(pub.SpanContext().SpanID(), sent.SpanID(),
					"the job carries job.publish, so job.process is its child")

				s.Contains(attrMap(pub), "osapi.target=server1",
					"the target that the caller gave, not the machine ID")
				s.Contains(attrMap(pub), "messaging.destination.name=jobs.query.host.mid1")
				s.Contains(attrMap(col), "job.expected=1")
				s.Contains(attrMap(col), "job.answered=1")
				s.Equal(codes.Unset, col.Status().Code)
			},
		},
		{
			name:    "when no agent answers marks collect as an error",
			timeout: true,
			validate: func(spans map[string]sdktrace.ReadOnlySpan, _ trace.SpanContext, _ trace.SpanContext) {
				s.Require().NotNil(spans["job.collect"])
				s.Equal(codes.Error, spans["job.collect"].Status().Code)
				s.Equal(codes.Unset, spans["job.publish"].Status().Code)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			timeout := 30 * time.Second
			if tc.timeout {
				timeout = 50 * time.Millisecond
			}
			jobs, err := client.New(slog.Default(), s.nats, &client.Options{
				Timeout: timeout, JobsBucket: "jobs", StreamName: "JOBS",
				// The controller resolves a hostname to the agent's machine ID.
				TargetResolver: func(string) string { return "mid1" },
			})
			s.Require().NoError(err)

			var sent trace.SpanContext
			s.nats.EXPECT().
				KVPut(gomock.Any(), "jobs", gomock.Any(), gomock.Any()).
				Return(uint64(1), nil)
			s.nats.EXPECT().Publish(gomock.Any(), "jobs.query.host.mid1", gomock.Any()).
				DoAndReturn(func(ctx context.Context, _ string, _ []byte) error {
					sent = trace.SpanContextFromContext(ctx)

					return nil
				})
			ch := make(chan jetstream.KeyValueEntry, 1)
			if !tc.timeout {
				entry := jobmocks.NewMockKeyValueEntry(s.ctrl)
				entry.EXPECT().Value().Return([]byte(tc.response))
				ch <- entry
			}
			var watched trace.SpanContext
			s.nats.EXPECT().KVWatch(gomock.Any(), "jobs", gomock.Any()).
				DoAndReturn(func(ctx context.Context, _, _ string, _ ...jetstream.WatchOpt) (<-chan jetstream.KeyValueEntry, error) {
					watched = trace.SpanContextFromContext(ctx)
					return ch, nil
				})

			ctx, request := otel.Tracer("test").
				Start(context.Background(), "GET /api/node/:hostname")
			_, _, _ = jobs.Query(ctx, "server1", "node", job.OperationType("hostname.get"), nil)
			request.End()

			spans := s.spans()
			tc.validate(spans, request.SpanContext(), sent)
			s.Equal(spans["job.collect"].SpanContext().SpanID(), watched.SpanID(),
				"the watch for the answers is part of job.collect")
		})
	}
}

func (s *SpanPublicTestSuite) TestEndCollectSpan() {
	s.Run("when an agent times out counts it and names it in an event", func() {
		_, span := otel.Tracer("test").Start(context.Background(), "job.collect")

		client.ExportEndCollectSpan(span, 2, map[string]*job.Response{
			"web-01": {Status: job.StatusCompleted},
			"web-02": {Status: job.StatusTimeout},
		})

		got := s.spans()["job.collect"]
		s.Require().NotNil(got)
		s.Contains(attrMap(got), "job.expected=2")
		s.Contains(attrMap(got), "job.answered=1")
		s.Contains(attrMap(got), "job.timed_out=1")
		s.Equal(codes.Unset, got.Status().Code, "a timed-out agent is not an error of the request")
		s.Require().Len(got.Events(), 1)
		s.Equal("job.timeout", got.Events()[0].Name)
	})
}

// attrMap is the attributes of span as key=value strings.
func attrMap(
	span sdktrace.ReadOnlySpan,
) []string {
	out := make([]string, 0, len(span.Attributes()))
	for _, kv := range span.Attributes() {
		out = append(out, string(kv.Key)+"="+kv.Value.String())
	}

	return out
}

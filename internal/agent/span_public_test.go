// Copyright (c) 2026 John Dewey
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package agent_test

import (
	"context"
	"testing"

	natsclient "github.com/osapi-io/nats-client/pkg/client"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/agent"
	agentmocks "github.com/osapi-io/osapi/internal/agent/mocks"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/job/mocks"
	hostMocks "github.com/osapi-io/osapi/internal/provider/node/host/mocks"
)

type SpanPublicTestSuite struct {
	suite.Suite
}

func TestSpanPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(SpanPublicTestSuite))
}

// attrs are the attributes of a span as a map, for comparison.
func attrs(
	kvs []attribute.KeyValue,
) map[string]string {
	out := map[string]string{}
	for _, kv := range kvs {
		out[string(kv.Key)] = kv.Value.String()
	}

	return out
}

func (s *SpanPublicTestSuite) TestIdentityAttrs() {
	tests := []struct {
		name  string
		facts *job.FactsRegistration
		want  map[string]string
	}{
		{
			name:  "when facts are collected adds the architecture",
			facts: &job.FactsRegistration{Architecture: "arm64"},
			want: map[string]string{
				"osapi.agent.hostname":   "web-01",
				"osapi.agent.machine_id": "abc123",
				"host.arch":              "arm64",
			},
		},
		{
			name:  "when facts are not collected yet has the hostname and machine ID",
			facts: nil,
			want: map[string]string{
				"osapi.agent.hostname":   "web-01",
				"osapi.agent.machine_id": "abc123",
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			a := newTestAgent(newTestAgentParams{})

			got := agent.ExportIdentityAttrs(a, "web-01", "abc123", tc.facts)

			s.Equal(tc.want, attrs(got))
		})
	}
}

func (s *SpanPublicTestSuite) TestRecordOutcome() {
	changed := true
	tests := []struct {
		name        string
		response    job.Response
		requestType job.Type
		validate    func(span sdktrace.ReadOnlySpan)
	}{
		{
			name:        "when a write completes records the status and changed",
			response:    job.Response{Status: job.StatusCompleted, Changed: &changed},
			requestType: job.TypeModify,
			validate: func(span sdktrace.ReadOnlySpan) {
				s.Equal(
					map[string]string{"job.status": "completed", "job.changed": "true"},
					attrs(span.Attributes()),
				)
				s.Equal(codes.Unset, span.Status().Code)
			},
		},
		{
			name:        "when a read completes records only the status",
			response:    job.Response{Status: job.StatusCompleted},
			requestType: job.TypeQuery,
			validate: func(span sdktrace.ReadOnlySpan) {
				s.Equal(map[string]string{"job.status": "completed"}, attrs(span.Attributes()))
			},
		},
		{
			name:        "when the job fails marks the span as an error",
			response:    job.Response{Status: job.StatusFailed, Error: "permission denied"},
			requestType: job.TypeModify,
			validate: func(span sdktrace.ReadOnlySpan) {
				s.Equal(map[string]string{"job.status": "failed"}, attrs(span.Attributes()))
				s.Equal(codes.Error, span.Status().Code)
				s.Equal("permission denied", span.Status().Description)
			},
		},
		{
			name: "when the host skips adds an event and is not an error",
			response: job.Response{
				Status: job.StatusSkipped,
				Error:  "operation not supported on this OS family",
			},
			requestType: job.TypeQuery,
			validate: func(span sdktrace.ReadOnlySpan) {
				s.Equal(map[string]string{"job.status": "skipped"}, attrs(span.Attributes()))
				s.Equal(codes.Unset, span.Status().Code)
				s.Require().Len(span.Events(), 1)
				s.Equal("job.skipped", span.Events()[0].Name)
				s.Equal(
					map[string]string{
						"job.skip_reason": "operation not supported on this OS family",
					},
					attrs(span.Events()[0].Attributes),
				)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			recorder := tracetest.NewSpanRecorder()
			tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
			_, span := tp.Tracer("test").Start(context.Background(), "job.process")

			agent.ExportRecordOutcome(span, tc.response, tc.requestType)
			span.End()

			s.Require().Len(recorder.Ended(), 1)
			tc.validate(recorder.Ended()[0])
		})
	}
}

// TestJobProcessJoinsTheTrace checks the agent half of the span tree of a
// request: the job comes with the trace context of job.publish in its
// headers, nats-client gives the handler that context, and job.process is a
// child of it, with the agent and the result on it. The controller half is in
// internal/job/client.
func (s *SpanPublicTestSuite) TestJobProcessJoinsTheTrace() {
	recorder := tracetest.NewSpanRecorder()
	previous, previousProp := otel.GetTracerProvider(), otel.GetTextMapPropagator()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder)))
	otel.SetTextMapPropagator(propagation.TraceContext{})
	defer func() {
		otel.SetTracerProvider(previous)
		otel.SetTextMapPropagator(previousProp)
	}()

	// The span that the controller opened when it published the job.
	pubCtx, publish := otel.Tracer("test").Start(context.Background(), "job.publish")
	headers := nats.Header{}
	propagation.TraceContext{}.Inject(pubCtx, propagation.HeaderCarrier(headers))
	publish.End()

	ctrl := gomock.NewController(s.T())
	jobs := mocks.NewMockJobClient(ctrl)
	a := newTestAgent(newTestAgentParams{
		jobClient:    jobs,
		hostProvider: hostMocks.NewDefaultMockProvider(ctrl),
	})
	agent.ExportIdentityAttrs(a, "web-01", "abc123", nil)

	jobs.EXPECT().HasJobResponse(gomock.Any(), "trace-job", gomock.Any()).Return(false, nil)
	jobs.EXPECT().GetJobData(gomock.Any(), "jobs.trace-job").Return([]byte(`{
		"id": "trace-job",
		"operation": {"type": "node.hostname.get", "data": {}}
	}`), nil)
	jobs.EXPECT().
		WriteStatusEvent(gomock.Any(), "trace-job", gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()
	jobs.EXPECT().
		WriteJobResponse(gomock.Any(), "trace-job", gomock.Any(), gomock.Any(), "completed", "", gomock.Any()).
		Return(nil)

	msg := agentmocks.NewMockMsg(ctrl)
	msg.EXPECT().Subject().Return("jobs.query.web-01").AnyTimes()
	msg.EXPECT().Data().Return([]byte("trace-job")).AnyTimes()
	msg.EXPECT().InProgress().Return(nil).AnyTimes()

	ctx := natsclient.MessageContext(context.Background(), headers)
	s.Require().NoError(agent.ExportHandleJobMessage(ctx, a, msg))

	var process sdktrace.ReadOnlySpan
	for _, sp := range recorder.Ended() {
		if sp.Name() == "job.process" {
			process = sp
		}
	}
	s.Require().NotNil(process)
	s.Equal(publish.SpanContext().TraceID(), process.SpanContext().TraceID(), "one trace")
	s.Equal(publish.SpanContext().SpanID(), process.Parent().SpanID(), "a child of job.publish")
	got := attrs(process.Attributes())
	s.Equal("web-01", got["osapi.agent.hostname"])
	s.Equal("abc123", got["osapi.agent.machine_id"])
	s.Equal("completed", got["job.status"])
	s.Equal("hostname.get", got["job.operation"])
}

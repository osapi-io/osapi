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

package client_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/pkg/sdk/client"
)

type TelemetryPublicTestSuite struct {
	suite.Suite
}

func (suite *TelemetryPublicTestSuite) TestTrace() {
	const id = "0123456789abcdef0123456789abcdef"

	reply := func(status int, body string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			suite.Equal("/api/telemetry/trace/"+id, r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			_, _ = w.Write([]byte(body))
		}
	}

	tests := []struct {
		name      string
		handler   http.HandlerFunc
		serverURL string
		validate  func(*client.Response[client.Trace], error)
	}{
		{
			name: "returns the spans",
			handler: reply(http.StatusOK, `{"trace_id":"`+id+`",
				"started":"2026-11-15T12:00:00Z","duration":"8.5ms","services":["osapi-controller","osapi-agent"],
				"failed_spans":1,"nats_calls":3,"nats_time":"1.2ms","spans":[
				{"name":"GET /api/x","on":"controller","duration":"8.5ms","status":"ok","detail":"status=200",
				 "span":{"span_id":"a","name":"GET /api/x","service":"osapi-controller","kind":"server",
				 "start_time":"2026-11-15T12:00:00Z","duration_ms":8.5,"status":"ok","attributes":{}},
				 "children":[{"name":"job.process","on":"agent","duration":"1ms","status":"failed",
				 "span":{"span_id":"b","parent_span_id":"a","name":"job.process","service":"osapi-agent",
				 "start_time":"2026-11-15T12:00:00Z","duration_ms":1,"status":"error",
				 "status_message":"boom","attributes":{"job.status":"failed"}}}]}]}`),
			validate: func(resp *client.Response[client.Trace], err error) {
				suite.Require().NoError(err)
				t := resp.Data
				suite.Equal(id, t.ID)
				suite.Equal(8500*time.Microsecond, t.Duration)
				suite.Equal([]string{"osapi-controller", "osapi-agent"}, t.Services)
				suite.Equal(1, t.FailedSpans)
				suite.Equal(3, t.NATSCalls)
				suite.Equal(1200*time.Microsecond, t.NATSTime)
				suite.False(t.Started.IsZero())
				suite.Require().Len(t.Spans, 1)
				root := t.Spans[0]
				suite.Equal("controller", root.On)
				suite.Equal("status=200", root.Detail)
				suite.Equal("server", root.Span.Kind)
				suite.Empty(root.Span.ParentID)
				suite.Equal(8500*time.Microsecond, root.Span.Duration)
				suite.False(root.Span.Failed)
				suite.Require().Len(root.Children, 1)
				job := root.Children[0]
				suite.Equal("failed", job.Status)
				suite.Equal("a", job.Span.ParentID)
				suite.True(job.Span.Failed)
				suite.Equal("boom", job.Span.StatusMessage)
				suite.Equal("failed", job.Span.Attributes["job.status"])
				suite.NotEmpty(resp.RawJSON())
			},
		},
		{
			name:    "a trace that is not there yet is a NotFoundError",
			handler: reply(http.StatusNotFound, `{"error":"trace not found"}`),
			validate: func(resp *client.Response[client.Trace], err error) {
				var target *client.NotFoundError
				suite.True(errors.As(err, &target))
				suite.Nil(resp)
			},
		},
		{
			name: "no tracing backend is an error with its message",
			handler: reply(
				http.StatusServiceUnavailable,
				`{"error":"no tracing backend is configured"}`,
			),
			validate: func(_ *client.Response[client.Trace], err error) {
				suite.ErrorContains(err, "no tracing backend is configured")
			},
		},
		{
			name: "an answer that is not JSON is an error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusOK)
			},
			validate: func(_ *client.Response[client.Trace], err error) {
				var target *client.UnexpectedStatusError
				suite.True(errors.As(err, &target))
			},
		},
		{
			name:      "a controller that cannot be reached is an error",
			serverURL: "http://127.0.0.1:0",
			validate: func(_ *client.Response[client.Trace], err error) {
				suite.ErrorContains(err, "get trace "+id)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			url := tc.serverURL
			if tc.handler != nil {
				server := httptest.NewServer(tc.handler)
				defer server.Close()
				url = server.URL
			}

			sut := client.New(url, "test-token", client.WithLogger(slog.Default()))
			tc.validate(sut.Telemetry.Trace(context.Background(), id, false))
		})
	}
}

func TestTelemetryPublicTestSuite(t *testing.T) {
	suite.Run(t, new(TelemetryPublicTestSuite))
}

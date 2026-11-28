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
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/controller/endpoint/server"
	"github.com/osapi-io/osapi/internal/controller/endpoints/telemetry"
	"github.com/osapi-io/osapi/internal/controller/endpoints/telemetry/gen"
	"github.com/osapi-io/osapi/internal/controller/endpoints/telemetry/mocks"
	"github.com/osapi-io/osapi/internal/telemetry/traces"
)

const (
	traceID = "26e448c9dbe1c4a8d216a98c675079c9"
)

// signingKey signs every token a case sends.
var signingKey = authtoken.Key{ID: "test", Secret: "test-signing-key-for-the-telemetry-group"}

type MethodsPublicTestSuite struct {
	suite.Suite

	ctrl   *gomock.Controller
	reader *mocks.MockTraceReader
}

func (s *MethodsPublicTestSuite) SetupSubTest() {
	s.ctrl = gomock.NewController(s.T())
	s.reader = mocks.NewMockTraceReader(s.ctrl)
}

func (s *MethodsPublicTestSuite) TearDownSubTest() {
	s.ctrl.Finish()
}

func (s *MethodsPublicTestSuite) TestGet() {
	start := time.Date(2026, 11, 17, 8, 0, 0, 0, time.UTC)

	// spans is a trace whose spans record each target of targets.
	spans := func(targets ...string) *traces.Trace {
		t := &traces.Trace{ID: traceID, Spans: []traces.Span{
			{ID: "root", Name: "GET /api/node", Start: start},
		}}
		for _, target := range targets {
			t.Spans = append(t.Spans, traces.Span{
				ID: target, ParentID: "root", Name: "validate.target", Start: start,
				Attributes: map[string]string{"osapi.target": target},
			})
		}

		return t
	}

	tests := []struct {
		name         string
		noBackend    bool
		scope        []string
		setup        func()
		validateFunc func(telemetry.TraceResponse, error)
	}{
		{
			name:  "when a scoped token reads a trace whose targets are in its scope, it is the trace",
			scope: []string{"group:web"},
			setup: func() {
				s.reader.EXPECT().Trace(gomock.Any(), traceID).
					Return(spans("group:web.dev", "group:web"), nil)
			},
			validateFunc: func(got telemetry.TraceResponse, err error) {
				s.Require().NoError(err)
				s.Equal(traceID, got.TraceID)
			},
		},
		{
			name:  "when a scoped token reads a trace with a target outside its scope, it is ErrNotFound",
			scope: []string{"group:web"},
			setup: func() {
				s.reader.EXPECT().Trace(gomock.Any(), traceID).
					Return(spans("group:web.dev", "group:db"), nil)
			},
			validateFunc: func(_ telemetry.TraceResponse, err error) {
				s.ErrorIs(err, telemetry.ErrNotFound)
				s.ErrorContains(err, traceID)
			},
		},
		{
			name:  "when a scoped token reads a trace that reached _all, it is ErrNotFound",
			scope: []string{"group:web"},
			setup: func() {
				s.reader.EXPECT().Trace(gomock.Any(), traceID).Return(spans("_all"), nil)
			},
			validateFunc: func(_ telemetry.TraceResponse, err error) {
				s.ErrorIs(err, telemetry.ErrNotFound)
			},
		},
		{
			name:  "when a scoped token reads a trace with no target, it is ErrNotFound",
			scope: []string{"group:web"},
			setup: func() {
				s.reader.EXPECT().Trace(gomock.Any(), traceID).Return(spans(), nil)
			},
			validateFunc: func(_ telemetry.TraceResponse, err error) {
				s.ErrorIs(err, telemetry.ErrNotFound)
			},
		},
		{
			name: "when the backend has the trace, each span is in the answer",
			setup: func() {
				s.reader.EXPECT().Trace(gomock.Any(), traceID).Return(&traces.Trace{
					ID: traceID,
					Spans: []traces.Span{
						{
							ID: "a", Name: "GET /api/agent", Service: "osapi-controller",
							Kind: "server", Start: start, Duration: 1500 * time.Microsecond,
							Attributes: map[string]string{"http.status": "200"},
						},
						{
							ID: "b", ParentID: "a", Name: "job.process", Service: "osapi-agent",
							Start: start, Failed: true, StatusMessage: "boom",
						},
					},
				}, nil)
			},
			validateFunc: func(got telemetry.TraceResponse, err error) {
				s.Require().NoError(err)
				s.Equal(traceID, got.TraceID)
				s.Require().Len(got.Spans, 1, "the root, with the job under it")
				s.Equal(telemetry.TraceSpan{
					SpanID: "a", Name: "GET /api/agent", Service: "osapi-controller",
					Kind: "server", StartTime: start, DurationMs: 1.5, Status: "ok",
					Attributes: map[string]string{"http.status": "200"},
				}, got.Spans[0].Span)
				s.Equal(start, got.Started)
				s.Equal(1, got.FailedSpans)
				s.Require().Len(got.Spans[0].Children, 1)
				job := got.Spans[0].Children[0]
				s.Equal("a", job.Span.ParentSpanID)
				s.Equal("error", job.Status)
				s.Equal(`error=boom`, job.Detail)
			},
		},
		{
			name:      "when no tracing backend is configured, it is ErrNoBackend and names the setting",
			noBackend: true,
			setup:     func() {},
			validateFunc: func(_ telemetry.TraceResponse, err error) {
				s.ErrorIs(err, telemetry.ErrNoBackend)
				s.ErrorContains(err, "telemetry.tracing.url")
			},
		},
		{
			name: "when the backend does not have the trace yet, it is ErrNotFound and says to try again",
			setup: func() {
				s.reader.EXPECT().Trace(gomock.Any(), traceID).Return(nil, traces.ErrNotFound)
			},
			validateFunc: func(_ telemetry.TraceResponse, err error) {
				s.ErrorIs(err, telemetry.ErrNotFound)
				s.ErrorContains(err, "try again")
			},
		},
		{
			name: "when the backend fails, the caller gets a message without the backend's error",
			setup: func() {
				s.reader.EXPECT().Trace(gomock.Any(), traceID).
					Return(nil, errors.New("dial tcp 10.0.0.9:16686: refused"))
			},
			validateFunc: func(_ telemetry.TraceResponse, err error) {
				s.EqualError(err, "failed to read the trace from the tracing backend")
				s.ErrorContains(errors.Unwrap(err), "10.0.0.9")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()
			d := telemetry.Deps{Traces: s.reader}
			if tc.noBackend {
				d.Traces = nil
			}

			ctx := authtoken.WithTargets(context.Background(), tc.scope)
			tc.validateFunc(telemetry.Get(ctx, d, telemetry.TraceInput{ID: traceID}))
		})
	}
}

func token(
	perms ...string,
) string {
	t, _ := authtoken.New().
		Generate(signingKey, []types.Role{types.RoleRead}, "test", perms, nil)

	return t
}

// TestRoutes runs requests through the generated route, with the auth, the
// decoding and the status mapping of the endpoint runtime.
func (s *MethodsPublicTestSuite) TestRoutes() {
	tests := []struct {
		name      string
		path      string
		token     string
		noBackend bool
		setup     func()
		wantCode  int
		wantBody  string
	}{
		{
			name:  "when the read role has trace:read, it is a 200",
			path:  "/api/telemetry/trace/" + traceID,
			token: token("trace:read"),
			setup: func() {
				s.reader.EXPECT().Trace(gomock.Any(), traceID).
					Return(&traces.Trace{ID: traceID}, nil)
			},
			wantCode: http.StatusOK,
			wantBody: `{"trace_id":"` + traceID + `","spans":[]}`,
		},
		{
			name:     "when there is no token, it is a 401",
			path:     "/api/telemetry/trace/" + traceID,
			setup:    func() {},
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "when the token lacks trace:read, it is a 403",
			path:     "/api/telemetry/trace/" + traceID,
			token:    token("agent:read"),
			setup:    func() {},
			wantCode: http.StatusForbidden,
		},
		{
			name:     "when the ID is not 32 hexadecimal characters, it is a 400",
			path:     "/api/telemetry/trace/NOT-A-TRACE",
			token:    token("trace:read"),
			setup:    func() {},
			wantCode: http.StatusBadRequest,
		},
		{
			name:      "when no backend is configured, it is a 503",
			path:      "/api/telemetry/trace/" + traceID,
			token:     token("trace:read"),
			noBackend: true,
			setup:     func() {},
			wantCode:  http.StatusServiceUnavailable,
		},
		{
			name:  "when the backend does not have the trace, it is a 404",
			path:  "/api/telemetry/trace/" + traceID,
			token: token("trace:read"),
			setup: func() {
				s.reader.EXPECT().Trace(gomock.Any(), traceID).Return(nil, traces.ErrNotFound)
			},
			wantCode: http.StatusNotFound,
		},
		{
			name:  "when the backend fails, it is a 500 that does not name the backend",
			path:  "/api/telemetry/trace/" + traceID,
			token: token("trace:read"),
			setup: func() {
				s.reader.EXPECT().Trace(gomock.Any(), traceID).
					Return(nil, errors.New("dial tcp 10.0.0.9:16686: refused"))
			},
			wantCode: http.StatusInternalServerError,
			wantBody: `{"error":"failed to read the trace from the tracing backend","reason":"server_error"}`,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.setup()
			d := telemetry.Deps{Traces: s.reader}
			if tc.noBackend {
				d.Traces = nil
			}
			srv := server.New(slog.Default(), authtoken.Keys{signingKey}, nil, d)
			e := echo.New()
			srv.Mount(e, gen.Routes(srv))

			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			s.Equal(tc.wantCode, rec.Code, rec.Body.String())
			if tc.wantBody != "" {
				s.JSONEq(tc.wantBody, rec.Body.String())
			}
		})
	}
}

func TestMethodsPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(MethodsPublicTestSuite))
}

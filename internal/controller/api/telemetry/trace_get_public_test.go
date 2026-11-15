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
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/authtoken"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/controller/api"
	telemetryapi "github.com/osapi-io/osapi/internal/controller/api/telemetry"
	"github.com/osapi-io/osapi/internal/controller/api/telemetry/gen"
	"github.com/osapi-io/osapi/internal/controller/api/telemetry/mocks"
	"github.com/osapi-io/osapi/internal/telemetry/traces"
)

const (
	traceID    = "0123456789abcdef0123456789abcdef"
	signingKey = "test-signing-key-for-rbac-integration"
)

type TraceGetPublicTestSuite struct {
	suite.Suite

	ctrl   *gomock.Controller
	reader *mocks.MockTraceReader
}

func (s *TraceGetPublicTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.reader = mocks.NewMockTraceReader(s.ctrl)
}

func (s *TraceGetPublicTestSuite) SetupSubTest() {
	s.SetupTest()
}

func (s *TraceGetPublicTestSuite) TestGetTrace() {
	start := time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC)
	trace := &traces.Trace{ID: traceID, Spans: []traces.Span{
		{
			ID: "a", Name: "GET /api/x", Service: "osapi-controller", Kind: "server",
			Start: start, Duration: 8 * time.Millisecond,
			Attributes: map[string]string{"http.route": "/api/x"},
		},
		{
			ID: "b", ParentID: "a", Name: "job.process", Service: "osapi-agent",
			Start: start, Duration: 1500 * time.Microsecond, Failed: true, StatusMessage: "boom",
			Attributes: map[string]string{},
		},
	}}

	tests := []struct {
		name     string
		id       string
		reader   bool
		setup    func()
		validate func(gen.GetTraceResponseObject)
	}{
		{
			name:   "returns each span",
			id:     traceID,
			reader: true,
			setup: func() {
				s.reader.EXPECT().Trace(gomock.Any(), traceID).Return(trace, nil)
			},
			validate: func(r gen.GetTraceResponseObject) {
				got, ok := r.(gen.GetTrace200JSONResponse)
				s.Require().True(ok)
				s.Equal(traceID, got.TraceId)
				s.Require().Len(got.Spans, 2)

				root, child := got.Spans[0], got.Spans[1]
				s.Nil(root.ParentSpanId)
				s.Equal("server", *root.Kind)
				s.Equal(gen.Ok, root.Status)
				s.Nil(root.StatusMessage)
				s.InDelta(8.0, root.DurationMs, 0.0001)

				s.Equal("a", *child.ParentSpanId)
				s.Nil(child.Kind)
				s.Equal(gen.Error, child.Status)
				s.Equal("boom", *child.StatusMessage)
				s.InDelta(1.5, child.DurationMs, 0.0001)
			},
		},
		{
			name:   "an ID that is not 32 hexadecimal characters is a 400",
			id:     "not-a-trace",
			reader: true,
			validate: func(r gen.GetTraceResponseObject) {
				_, ok := r.(gen.GetTrace400JSONResponse)
				s.True(ok)
			},
		},
		{
			name: "with no tracing backend is a 503 that names the setting",
			id:   traceID,
			validate: func(r gen.GetTraceResponseObject) {
				got, ok := r.(gen.GetTrace503JSONResponse)
				s.Require().True(ok)
				s.Contains(*got.Error, "telemetry.tracing.url")
			},
		},
		{
			name:   "a trace the backend does not have yet is a 404 that says to try again",
			id:     traceID,
			reader: true,
			setup: func() {
				s.reader.EXPECT().Trace(gomock.Any(), traceID).Return(nil, traces.ErrNotFound)
			},
			validate: func(r gen.GetTraceResponseObject) {
				got, ok := r.(gen.GetTrace404JSONResponse)
				s.Require().True(ok)
				s.Contains(*got.Error, "try again")
			},
		},
		{
			name:   "a backend that fails is a 500",
			id:     traceID,
			reader: true,
			setup: func() {
				s.reader.EXPECT().Trace(gomock.Any(), traceID).Return(nil, errors.New("refused"))
			},
			validate: func(r gen.GetTraceResponseObject) {
				_, ok := r.(gen.GetTrace500JSONResponse)
				s.True(ok)
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.setup != nil {
				tc.setup()
			}
			var reader telemetryapi.TraceReader
			if tc.reader {
				reader = s.reader
			}

			resp, err := telemetryapi.New(slog.Default(), reader).
				GetTrace(context.Background(), gen.GetTraceRequestObject{Id: tc.id})

			s.Require().NoError(err)
			tc.validate(resp)
		})
	}
}

func (s *TraceGetPublicTestSuite) TestGetTraceRBACHTTP() {
	tokens := authtoken.New(slog.Default())

	tests := []struct {
		name  string
		roles []string
		perms []string
		want  int
	}{
		{name: "when no token returns 401", want: http.StatusUnauthorized},
		{
			name:  "when the token lacks trace:read returns 403",
			roles: []string{"read"},
			perms: []string{"job:read"},
			want:  http.StatusForbidden,
		},
		{
			name:  "when the read role has trace:read returns 200",
			roles: []string{"read"},
			want:  http.StatusOK,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.want == http.StatusOK {
				s.reader.EXPECT().Trace(gomock.Any(), traceID).
					Return(&traces.Trace{ID: traceID, Spans: []traces.Span{{ID: "a", Name: "x"}}}, nil)
			}
			cfg := config.Config{Controller: config.Controller{API: config.APIServer{
				Security: config.ServerSecurity{SigningKey: signingKey},
			}}}
			server := api.New(cfg, slog.Default())
			server.RegisterHandlers(telemetryapi.Handler(slog.Default(), s.reader, signingKey, nil))

			req := httptest.NewRequest(http.MethodGet, "/api/telemetry/trace/"+traceID, nil)
			if tc.roles != nil {
				token, err := tokens.Generate(signingKey, tc.roles, "test-user", tc.perms)
				s.Require().NoError(err)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			}
			rec := httptest.NewRecorder()
			server.Echo.ServeHTTP(rec, req)

			s.Equal(tc.want, rec.Code, rec.Body.String())
		})
	}
}

func TestTraceGetPublicTestSuite(t *testing.T) {
	suite.Run(t, new(TraceGetPublicTestSuite))
}

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

package api_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/trace"

	"github.com/osapi-io/osapi/internal/audit"
	"github.com/osapi-io/osapi/internal/audit/mocks"
	"github.com/osapi-io/osapi/internal/controller/api"
	"github.com/osapi-io/osapi/pkg/sdk/types"
	"go.uber.org/mock/gomock"
)

type AuditMiddlewarePublicTestSuite struct {
	suite.Suite
}

func (s *AuditMiddlewarePublicTestSuite) TestAuditMiddleware() {
	tests := []struct {
		name        string
		path        string
		method      string
		body        string
		contentType string
		jobID       string
		steps       map[string][]types.Step
		subject     string
		roles       []types.Role
		targets     []string
		storeErr    error
		wantWrite   bool
		setupReq    func(req *http.Request) *http.Request
		// respond is the handler's answer; nil writes 200 "ok".
		respond func(c *echo.Context) error
		// wantCode is the code the client gets; 0 is 200.
		wantCode int
		// validateResponse checks the answer and the log, when it is set.
		validateResponse func(body string, log string)
		validateFunc     func(entry audit.Entry)
		validateBody     func(seen string)
	}{
		{
			name:      "authenticated request is logged",
			path:      "/api/node/hostname",
			subject:   "user@example.com",
			roles:     []types.Role{types.RoleAdmin},
			wantWrite: true,
			validateFunc: func(entry audit.Entry) {
				s.Equal("user@example.com", entry.User)
				s.Equal("GET", entry.Method)
				s.Equal("/api/node/hostname", entry.Path)
				s.Equal(http.StatusOK, entry.ResponseCode)
				s.Equal([]types.Role{types.RoleAdmin}, entry.Roles)
				s.Empty(entry.Targets)
			},
		},
		{
			name:      "the target scope of the token is recorded",
			path:      "/api/node/web-01/hostname",
			subject:   "team-web",
			roles:     []types.Role{types.RoleOperator},
			targets:   []string{"group:web"},
			wantWrite: true,
			validateFunc: func(entry audit.Entry) {
				s.Equal([]string{"group:web"}, entry.Targets)
			},
		},
		{
			name:    "unauthenticated request is skipped",
			path:    "/api/node/hostname",
			subject: "",
		},
		{
			name:    "health path is excluded",
			path:    "/api/health",
			subject: "user@example.com",
		},
		{
			name:    "health ready path is excluded",
			path:    "/api/health/ready",
			subject: "user@example.com",
		},
		{
			name:    "metrics path is excluded",
			path:    "/metrics",
			subject: "user@example.com",
		},
		{
			name:      "authenticated request with trace context captures trace ID",
			path:      "/api/node/hostname",
			subject:   "user@example.com",
			roles:     []types.Role{types.RoleAdmin},
			wantWrite: true,
			setupReq: func(req *http.Request) *http.Request {
				traceID, _ := trace.TraceIDFromHex(
					"4bf92f3577b34da6a3ce929d0e0e4736",
				)
				spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
					TraceID:    traceID,
					SpanID:     trace.SpanID{1},
					TraceFlags: trace.FlagsSampled,
				})
				ctx := trace.ContextWithSpanContext(
					req.Context(), spanCtx,
				)
				return req.WithContext(ctx)
			},
			validateFunc: func(entry audit.Entry) {
				s.Equal(
					"4bf92f3577b34da6a3ce929d0e0e4736",
					entry.TraceID,
				)
			},
		},
		{
			name:      "store error is handled gracefully",
			path:      "/api/node/hostname",
			subject:   "user@example.com",
			roles:     []types.Role{types.RoleAdmin},
			storeErr:  fmt.Errorf("write failed"),
			wantWrite: true,
		},
		{
			name:        "a mutating request records what it asked for",
			path:        "/api/node/command/shell",
			method:      http.MethodPost,
			body:        `{"command":"systemctl restart nginx","password":"hunter2"}`,
			contentType: "application/json",
			jobID:       "7c9e6679-7425-40de-944b-e07fc1f90ae7",
			subject:     "user@example.com",
			roles:       []types.Role{types.RoleAdmin},
			wantWrite:   true,
			validateFunc: func(entry audit.Entry) {
				s.Contains(entry.RequestSummary, "systemctl restart nginx")
				s.NotContains(entry.RequestSummary, "hunter2")
				s.Equal("7c9e6679-7425-40de-944b-e07fc1f90ae7", entry.JobID)
				s.Empty(entry.Steps)
			},
			validateBody: func(seen string) {
				// The handler still gets the body the middleware read.
				s.Contains(seen, "systemctl restart nginx")
			},
		},
		{
			name:        "a request whose job made calls records the steps of each host",
			path:        "/api/node/_all/sysctl/net.ipv4.ip_forward",
			method:      http.MethodPost,
			body:        `{"value":"1"}`,
			contentType: "application/json",
			jobID:       "7c9e6679-7425-40de-944b-e07fc1f90ae7",
			steps: map[string][]types.Step{
				"web-01": {{Behavior: "managedfile", Method: "update"}},
				"web-02": {{Behavior: "managedfile", Method: "update", Error: "denied"}},
			},
			subject:   "user@example.com",
			wantWrite: true,
			validateFunc: func(entry audit.Entry) {
				s.Equal([]audit.HostSteps{
					{
						Hostname: "web-01",
						Steps:    []types.Step{{Behavior: "managedfile", Method: "update"}},
					},
					{
						Hostname: "web-02",
						Steps: []types.Step{
							{Behavior: "managedfile", Method: "update", Error: "denied"},
						},
					},
				}, entry.Steps)
			},
		},
		{
			name:        "a request that creates no job records none",
			path:        "/api/node/command/shell",
			method:      http.MethodPost,
			body:        `{"command":"ls"}`,
			contentType: "application/json",
			subject:     "user@example.com",
			wantWrite:   true,
			validateFunc: func(entry audit.Entry) {
				s.Empty(entry.JobID)
				s.Contains(entry.RequestSummary, "ls")
			},
		},
		{
			name:      "a read records no summary at all",
			path:      "/api/node/hostname",
			subject:   "user@example.com",
			wantWrite: true,
			validateFunc: func(entry audit.Entry) {
				s.Empty(entry.RequestSummary)
			},
		},
		{
			name:        "a body that cannot be read says so",
			path:        "/api/node/command/shell",
			method:      http.MethodPost,
			contentType: "application/json",
			subject:     "user@example.com",
			wantWrite:   true,
			setupReq: func(req *http.Request) *http.Request {
				req.Body = io.NopCloser(failingReader{})

				return req
			},
			validateFunc: func(entry audit.Entry) {
				// An entry that says nothing about what was asked reads the same
				// as one where nothing was asked, so it says this instead.
				s.Equal("[body unreadable]", entry.RequestSummary)
			},
		},
		{
			name:        "a request with no body at all",
			path:        "/api/node/command/shell",
			method:      http.MethodPost,
			contentType: "application/json",
			subject:     "user@example.com",
			wantWrite:   true,
			setupReq: func(req *http.Request) *http.Request {
				req.Body = nil

				return req
			},
			validateFunc: func(entry audit.Entry) {
				s.Empty(entry.RequestSummary)
			},
		},
		{
			name:        "a body too large to record is described instead",
			path:        "/api/node/command/shell",
			method:      http.MethodPost,
			body:        `{"note":"` + strings.Repeat("x", 70*1024) + `"}`,
			contentType: "application/json",
			subject:     "user@example.com",
			wantWrite:   true,
			validateFunc: func(entry audit.Entry) {
				s.Contains(entry.RequestSummary, "not recorded")
			},
			validateBody: func(seen string) {
				// The handler still gets all of it.
				s.Greater(len(seen), 70*1024)
			},
		},
		{
			name:    "a handler that writes its response records that status",
			method:  http.MethodPost,
			path:    "/api/file/f",
			subject: "ops@example.com",
			respond: func(c *echo.Context) error {
				return c.String(http.StatusCreated, "made")
			},
			wantCode:  http.StatusCreated,
			wantWrite: true,
			validateFunc: func(entry audit.Entry) {
				s.Equal(http.StatusCreated, entry.ResponseCode)
			},
		},
		{
			name:    "a handler that writes nothing records the default status",
			method:  http.MethodPost,
			path:    "/api/file/f",
			subject: "ops@example.com",
			respond: func(*echo.Context) error {
				return nil
			},
			wantWrite: true,
			validateFunc: func(entry audit.Entry) {
				s.Equal(http.StatusOK, entry.ResponseCode)
			},
		},
		{
			name:    "a handler that returns an HTTP error records its status",
			method:  http.MethodPost,
			path:    "/api/file/f",
			subject: "ops@example.com",
			respond: func(*echo.Context) error {
				return echo.NewHTTPError(http.StatusNotFound, "no such file")
			},
			wantCode:  http.StatusNotFound,
			wantWrite: true,
			validateFunc: func(entry audit.Entry) {
				s.Equal(http.StatusNotFound, entry.ResponseCode)
			},
		},
		{
			name:    "a handler that returns a plain error records 500",
			method:  http.MethodPost,
			path:    "/api/file/f",
			subject: "ops@example.com",
			respond: func(*echo.Context) error {
				return errors.New("store down")
			},
			wantCode:  http.StatusInternalServerError,
			wantWrite: true,
			validateFunc: func(entry audit.Entry) {
				s.Equal(http.StatusInternalServerError, entry.ResponseCode)
			},
		},
		{
			name:    "a handler that panics records 500, logs the panic and its stack, and answers with no panic text",
			method:  http.MethodPost,
			path:    "/api/file/f",
			subject: "ops@example.com",
			respond: func(*echo.Context) error {
				panic("half done")
			},
			wantCode:  http.StatusInternalServerError,
			wantWrite: true,
			validateFunc: func(entry audit.Entry) {
				s.Equal(http.StatusInternalServerError, entry.ResponseCode)
			},
			validateResponse: func(body string, log string) {
				s.NotContains(body, "half done")
				s.Contains(log, "level=ERROR")
				s.Contains(log, "handler panicked")
				s.Contains(log, "half done")
				s.Contains(log, "goroutine")
			},
		},
		{
			name:    "a handler that swaps the response writer records no status",
			method:  http.MethodPost,
			path:    "/api/file/f",
			subject: "ops@example.com",
			respond: func(c *echo.Context) error {
				c.SetResponse(httptest.NewRecorder())

				return nil
			},
			// The client's recorder gets nothing, which httptest reads as 200.
			wantWrite: true,
			validateFunc: func(entry audit.Entry) {
				s.Equal(0, entry.ResponseCode)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctrl := gomock.NewController(s.T())
			store := mocks.NewMockStore(ctrl)

			// The middleware writes from a goroutine it does not join. The mock
			// closes this channel from the write, so the test waits on the call
			// itself rather than on a duration.
			written := make(chan struct{})
			var got audit.Entry

			if tt.wantWrite {
				store.EXPECT().
					Write(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, entry audit.Entry) error {
						got = entry
						close(written)
						return tt.storeErr
					})
			} else {
				// Excluded and unauthenticated paths return before the
				// goroutine is started, so any write is a regression.
				store.EXPECT().Write(gomock.Any(), gomock.Any()).Times(0)
			}

			var logs bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&logs, nil))

			e := echo.New()
			// The order of the server: Recover outside the audit middleware.
			e.Use(middleware.Recover())
			mw, drain := api.ExportAuditMiddleware(store, logger)
			defer drain(context.Background())
			e.Use(mw)

			method := tt.method
			if method == "" {
				method = http.MethodGet
			}

			seenBody := ""

			handler := func(c *echo.Context) error {
				// Simulate scopeMiddleware setting context values.
				if tt.subject != "" {
					c.Set(api.ContextKeySubject, tt.subject)
					c.Set(api.ContextKeyRoles, tt.roles)
					c.Set(api.ContextKeyTargets, tt.targets)
				}

				if c.Request().Body != nil {
					read, _ := io.ReadAll(c.Request().Body)
					seenBody = string(read)
				}

				// Stand in for a handler dispatching a job, which is what the
				// job client does inside a real one.
				if tt.jobID != "" {
					audit.RecordJobID(c.Request().Context(), tt.jobID)
				}
				for hostname, steps := range tt.steps {
					audit.RecordSteps(c.Request().Context(), tt.jobID, hostname, steps)
				}

				if tt.respond != nil {
					return tt.respond(c)
				}

				return c.String(http.StatusOK, "ok")
			}

			switch method {
			case http.MethodPost:
				e.POST(tt.path, handler)
			default:
				e.GET(tt.path, handler)
			}

			var reqBody io.Reader
			if tt.body != "" {
				reqBody = strings.NewReader(tt.body)
			}

			req := httptest.NewRequest(method, tt.path, reqBody)
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			if tt.setupReq != nil {
				req = tt.setupReq(req)
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			wantCode := tt.wantCode
			if wantCode == 0 {
				wantCode = http.StatusOK
			}
			s.Equal(wantCode, rec.Code)
			if tt.validateResponse != nil {
				tt.validateResponse(rec.Body.String(), logs.String())
			}

			if tt.wantWrite {
				<-written

				if tt.validateFunc != nil {
					tt.validateFunc(got)
				}
			}

			if tt.validateBody != nil {
				tt.validateBody(seenBody)
			}
		})
	}
}

func (s *AuditMiddlewarePublicTestSuite) TestAuditWriterEnqueue() {
	tests := []struct {
		name string
		// entries is how many to enqueue.
		entries int
		// held makes the store hold each write until the entries are in.
		held bool
		// closed drains the writer before the entries go in.
		closed       bool
		validateFunc func(written int)
	}{
		{
			name:    "an entry goes to the store through the queue",
			entries: 1,
			validateFunc: func(written int) {
				s.Equal(1, written)
			},
		},
		{
			name:    "after a drain, a request writes its entry before enqueue returns",
			entries: 1,
			closed:  true,
			validateFunc: func(written int) {
				s.Equal(1, written)
			},
		},
		{
			// One entry is in the store's held write, the queue holds the
			// rest, and the last one finds no room.
			name:    "a full queue drops the entry after a wait, and the API goes on",
			entries: api.ExportAuditQueueSize + 2,
			held:    true,
			validateFunc: func(written int) {
				s.Equal(api.ExportAuditQueueSize+1, written)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctrl := gomock.NewController(s.T())
			store := mocks.NewMockStore(ctrl)

			release := make(chan struct{})
			if !tt.held {
				close(release)
			}
			var mu sync.Mutex
			written := 0
			store.EXPECT().
				Write(gomock.Any(), gomock.Any()).
				DoAndReturn(func(context.Context, audit.Entry) error {
					<-release
					mu.Lock()
					written++
					mu.Unlock()

					return nil
				}).
				AnyTimes()

			w := api.ExportNewAuditWriter(store, slog.Default())
			if tt.closed {
				w.Close(context.Background())
			}
			for i := range tt.entries {
				w.Enqueue(audit.Entry{ID: fmt.Sprintf("e%d", i)})
			}
			if tt.held {
				close(release)
			}
			w.Close(context.Background())

			mu.Lock()
			defer mu.Unlock()
			tt.validateFunc(written)
		})
	}
}

func (s *AuditMiddlewarePublicTestSuite) TestAuditWriterClose() {
	tests := []struct {
		name string
		// entries is how many are queued, behind a held write, when the
		// drain starts.
		entries int
		// deadline ends the drain's context; 0 is no deadline.
		deadline     time.Duration
		validateFunc func(written int, took time.Duration)
	}{
		{
			name:    "a drain writes each queued entry before it returns",
			entries: 50,
			validateFunc: func(written int, _ time.Duration) {
				s.Equal(50, written)
			},
		},
		{
			name:     "a drain gives up when its context ends, before the held write finishes",
			entries:  2,
			deadline: 20 * time.Millisecond,
			validateFunc: func(written int, took time.Duration) {
				s.Less(took, 2*time.Second)
				s.Less(written, 2)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctrl := gomock.NewController(s.T())
			store := mocks.NewMockStore(ctrl)

			release := make(chan struct{})
			var mu sync.Mutex
			written := 0
			store.EXPECT().
				Write(gomock.Any(), gomock.Any()).
				DoAndReturn(func(context.Context, audit.Entry) error {
					<-release
					mu.Lock()
					written++
					mu.Unlock()

					return nil
				}).
				AnyTimes()

			w := api.ExportNewAuditWriter(store, slog.Default())
			for i := range tt.entries {
				w.Enqueue(audit.Entry{ID: fmt.Sprintf("e%d", i)})
			}

			ctx := context.Background()
			if tt.deadline > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, tt.deadline)
				defer cancel()
			} else {
				go func() { close(release) }()
			}
			start := time.Now()
			w.Close(ctx)
			took := time.Since(start)

			mu.Lock()
			got := written
			mu.Unlock()
			tt.validateFunc(got, took)

			// A second drain does nothing.
			w.Close(context.Background())
			if tt.deadline > 0 {
				close(release)
			}
		})
	}
}

func TestAuditMiddlewarePublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(AuditMiddlewarePublicTestSuite))
}

// failingReader is a request body that cannot be read, which is what a client
// that disconnects mid-request leaves behind.
type failingReader struct{}

func (failingReader) Read(
	_ []byte,
) (int, error) {
	return 0, errors.New("connection reset")
}

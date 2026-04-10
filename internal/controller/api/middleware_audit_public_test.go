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
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/trace"

	"github.com/retr0h/osapi/internal/audit"
	"github.com/retr0h/osapi/internal/controller/api"
)

// captureStore is a concurrency-safe audit store spy that records Write calls.
// It is kept as a hand-written spy rather than a gomock mock because the
// auditMiddleware fires writes in a goroutine after the HTTP response is sent,
// making gomock's strict call-count semantics impractical.
type captureStore struct {
	mu      sync.Mutex
	entries []audit.Entry
	err     error
}

func (f *captureStore) Write(
	_ context.Context,
	entry audit.Entry,
) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.err != nil {
		return f.err
	}

	f.entries = append(f.entries, entry)
	return nil
}

func (f *captureStore) Get(
	_ context.Context,
	_ string,
) (*audit.Entry, error) {
	return nil, nil
}

func (f *captureStore) List(
	_ context.Context,
	_ int,
	_ int,
) ([]audit.Entry, int, error) {
	return nil, 0, nil
}

func (f *captureStore) ListAll(
	_ context.Context,
) ([]audit.Entry, error) {
	return nil, nil
}

func (f *captureStore) getEntries() []audit.Entry {
	f.mu.Lock()
	defer f.mu.Unlock()

	cp := make([]audit.Entry, len(f.entries))
	copy(cp, f.entries)
	return cp
}

type AuditMiddlewarePublicTestSuite struct {
	suite.Suite
}

func (s *AuditMiddlewarePublicTestSuite) TestAuditMiddleware() {
	tests := []struct {
		name         string
		path         string
		subject      string
		roles        []string
		storeErr     error
		setupReq     func(req *http.Request) *http.Request
		validateFunc func(store *captureStore)
	}{
		{
			name:    "authenticated request is logged",
			path:    "/api/node/hostname",
			subject: "user@example.com",
			roles:   []string{"admin"},
			validateFunc: func(store *captureStore) {
				// Give goroutine time to write
				time.Sleep(50 * time.Millisecond)
				entries := store.getEntries()
				s.Len(entries, 1)
				s.Equal("user@example.com", entries[0].User)
				s.Equal("GET", entries[0].Method)
				s.Equal("/api/node/hostname", entries[0].Path)
				s.Equal(http.StatusOK, entries[0].ResponseCode)
				s.Equal([]string{"admin"}, entries[0].Roles)
			},
		},
		{
			name:    "unauthenticated request is skipped",
			path:    "/api/node/hostname",
			subject: "",
			validateFunc: func(store *captureStore) {
				time.Sleep(50 * time.Millisecond)
				entries := store.getEntries()
				s.Empty(entries)
			},
		},
		{
			name:    "health path is excluded",
			path:    "/api/health",
			subject: "user@example.com",
			validateFunc: func(store *captureStore) {
				time.Sleep(50 * time.Millisecond)
				entries := store.getEntries()
				s.Empty(entries)
			},
		},
		{
			name:    "health ready path is excluded",
			path:    "/api/health/ready",
			subject: "user@example.com",
			validateFunc: func(store *captureStore) {
				time.Sleep(50 * time.Millisecond)
				entries := store.getEntries()
				s.Empty(entries)
			},
		},
		{
			name:    "metrics path is excluded",
			path:    "/metrics",
			subject: "user@example.com",
			validateFunc: func(store *captureStore) {
				time.Sleep(50 * time.Millisecond)
				entries := store.getEntries()
				s.Empty(entries)
			},
		},
		{
			name:    "authenticated request with trace context captures trace ID",
			path:    "/api/node/hostname",
			subject: "user@example.com",
			roles:   []string{"admin"},
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
			validateFunc: func(store *captureStore) {
				time.Sleep(50 * time.Millisecond)
				entries := store.getEntries()
				s.Len(entries, 1)
				s.Equal(
					"4bf92f3577b34da6a3ce929d0e0e4736",
					entries[0].TraceID,
				)
			},
		},
		{
			name:     "store error is handled gracefully",
			path:     "/api/node/hostname",
			subject:  "user@example.com",
			roles:    []string{"admin"},
			storeErr: fmt.Errorf("write failed"),
			validateFunc: func(store *captureStore) {
				// Should not panic; the middleware logs the error
				time.Sleep(50 * time.Millisecond)
				entries := store.getEntries()
				s.Empty(entries)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			store := &captureStore{err: tt.storeErr}
			logger := slog.Default()

			e := echo.New()
			e.Use(api.ExportAuditMiddleware(store, logger))
			e.GET(tt.path, func(c echo.Context) error {
				// Simulate scopeMiddleware setting context values.
				if tt.subject != "" {
					c.Set(api.ContextKeySubject, tt.subject)
					c.Set(api.ContextKeyRoles, tt.roles)
				}
				return c.String(http.StatusOK, "ok")
			})

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.setupReq != nil {
				req = tt.setupReq(req)
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			s.Equal(http.StatusOK, rec.Code)
			tt.validateFunc(store)
		})
	}
}

func TestAuditMiddlewarePublicTestSuite(t *testing.T) {
	suite.Run(t, new(AuditMiddlewarePublicTestSuite))
}

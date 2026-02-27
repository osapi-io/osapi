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

package api

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

	"github.com/retr0h/osapi/internal/audit"
)

// fakeAuditStore is a simple in-memory audit store for testing.
type fakeAuditStore struct {
	mu      sync.Mutex
	entries []audit.Entry
	err     error
}

func (f *fakeAuditStore) Write(
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

func (f *fakeAuditStore) Get(
	_ context.Context,
	_ string,
) (*audit.Entry, error) {
	return nil, nil
}

func (f *fakeAuditStore) List(
	_ context.Context,
	_ int,
	_ int,
) ([]audit.Entry, int, error) {
	return nil, 0, nil
}

func (f *fakeAuditStore) ListAll(
	_ context.Context,
) ([]audit.Entry, error) {
	return nil, nil
}

func (f *fakeAuditStore) getEntries() []audit.Entry {
	f.mu.Lock()
	defer f.mu.Unlock()

	cp := make([]audit.Entry, len(f.entries))
	copy(cp, f.entries)
	return cp
}

type AuditMiddlewareTestSuite struct {
	suite.Suite
}

func (s *AuditMiddlewareTestSuite) TestAuditMiddleware() {
	tests := []struct {
		name         string
		path         string
		subject      string
		roles        []string
		storeErr     error
		validateFunc func(store *fakeAuditStore)
	}{
		{
			name:    "authenticated request is logged",
			path:    "/node/hostname",
			subject: "user@example.com",
			roles:   []string{"admin"},
			validateFunc: func(store *fakeAuditStore) {
				// Give goroutine time to write
				time.Sleep(50 * time.Millisecond)
				entries := store.getEntries()
				s.Len(entries, 1)
				s.Equal("user@example.com", entries[0].User)
				s.Equal("GET", entries[0].Method)
				s.Equal("/node/hostname", entries[0].Path)
				s.Equal(http.StatusOK, entries[0].ResponseCode)
				s.Equal([]string{"admin"}, entries[0].Roles)
			},
		},
		{
			name:    "unauthenticated request is skipped",
			path:    "/node/hostname",
			subject: "",
			validateFunc: func(store *fakeAuditStore) {
				time.Sleep(50 * time.Millisecond)
				entries := store.getEntries()
				s.Empty(entries)
			},
		},
		{
			name:    "health path is excluded",
			path:    "/health",
			subject: "user@example.com",
			validateFunc: func(store *fakeAuditStore) {
				time.Sleep(50 * time.Millisecond)
				entries := store.getEntries()
				s.Empty(entries)
			},
		},
		{
			name:    "health ready path is excluded",
			path:    "/health/ready",
			subject: "user@example.com",
			validateFunc: func(store *fakeAuditStore) {
				time.Sleep(50 * time.Millisecond)
				entries := store.getEntries()
				s.Empty(entries)
			},
		},
		{
			name:    "metrics path is excluded",
			path:    "/metrics",
			subject: "user@example.com",
			validateFunc: func(store *fakeAuditStore) {
				time.Sleep(50 * time.Millisecond)
				entries := store.getEntries()
				s.Empty(entries)
			},
		},
		{
			name:     "store error is handled gracefully",
			path:     "/node/hostname",
			subject:  "user@example.com",
			roles:    []string{"admin"},
			storeErr: fmt.Errorf("write failed"),
			validateFunc: func(store *fakeAuditStore) {
				// Should not panic; the middleware logs the error
				time.Sleep(50 * time.Millisecond)
				entries := store.getEntries()
				s.Empty(entries)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			store := &fakeAuditStore{err: tt.storeErr}
			logger := slog.Default()

			e := echo.New()
			e.Use(auditMiddleware(store, logger))
			e.GET(tt.path, func(c echo.Context) error {
				// Simulate scopeMiddleware setting context values.
				if tt.subject != "" {
					c.Set(ContextKeySubject, tt.subject)
					c.Set(ContextKeyRoles, tt.roles)
				}
				return c.String(http.StatusOK, "ok")
			})

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			s.Equal(http.StatusOK, rec.Code)
			tt.validateFunc(store)
		})
	}
}

func TestAuditMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(AuditMiddlewareTestSuite))
}

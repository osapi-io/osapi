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
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"go.opentelemetry.io/otel/trace"

	"github.com/osapi-io/osapi/internal/audit"
)

// maxAuditBodyBytes bounds what the middleware will read of a request body. A
// file upload is a multipart stream that no audit entry should hold, and reading
// one into memory to describe it is worse than saying it was too large.
const maxAuditBodyBytes = 64 * 1024

// excludedAuditPaths lists path prefixes that should not generate audit entries.
var excludedAuditPaths = []string{
	"/api/health",
	"/metrics",
}

const (
	// auditQueueSize is how many entries wait for the writer. A slow store
	// fills it, and then a request waits for room.
	auditQueueSize = 1024
	// auditEnqueueWait is how long a request waits for room in a full
	// queue. Past it, the entry is dropped and the drop is logged, so a
	// store that does not answer cannot hold the API.
	auditEnqueueWait = time.Second
	// auditWriteTimeout bounds one write to the store.
	auditWriteTimeout = 5 * time.Second
)

// auditWriter writes audit entries from one goroutine, from a bounded
// queue, so a slow store neither adds latency to a request nor makes the
// count of goroutines grow. close drains the queue at shutdown.
type auditWriter struct {
	store  audit.Store
	logger *slog.Logger
	queue  chan audit.Entry
	done   chan struct{}

	// mu guards closed against a send on the closed queue: enqueue holds
	// it to read, close holds it to write.
	mu     sync.RWMutex
	closed bool
}

// newAuditWriter starts the writer of store.
func newAuditWriter(
	store audit.Store,
	logger *slog.Logger,
) *auditWriter {
	w := &auditWriter{
		store:  store,
		logger: logger,
		queue:  make(chan audit.Entry, auditQueueSize),
		done:   make(chan struct{}),
	}
	go w.run()

	return w
}

// run writes each queued entry until close closes the queue.
func (w *auditWriter) run() {
	defer close(w.done)

	for entry := range w.queue {
		w.write(entry)
	}
}

// write puts entry in the store, and logs a failure.
func (w *auditWriter) write(
	entry audit.Entry,
) {
	// Not the request's context: the response is sent by the time the
	// entry is written, and the request's context is cancelled then.
	ctx, cancel := context.WithTimeout(context.Background(), auditWriteTimeout)
	defer cancel()

	if err := w.store.Write(ctx, entry); err != nil {
		w.logger.Warn(
			"failed to write audit entry",
			slog.String("error", err.Error()),
			slog.String("entry_id", entry.ID),
		)
	}
}

// enqueue gives entry to the writer. After close, a request that the
// server is still draining writes its entry itself.
func (w *auditWriter) enqueue(
	entry audit.Entry,
) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if w.closed {
		w.write(entry)

		return
	}

	timer := time.NewTimer(auditEnqueueWait)
	defer timer.Stop()

	select {
	case w.queue <- entry:
	case <-timer.C:
		w.logger.Error(
			"audit queue full, entry dropped",
			slog.String("entry_id", entry.ID),
			slog.String("path", entry.Path),
		)
	}
}

// close stops the queue and waits until the writer has written each
// queued entry, or until ctx ends. A second close does nothing.
func (w *auditWriter) close(
	ctx context.Context,
) {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()

		return
	}
	w.closed = true
	close(w.queue)
	w.mu.Unlock()

	select {
	case <-w.done:
	case <-ctx.Done():
		w.logger.Warn(
			"audit entries not written at shutdown",
			slog.Int("count", len(w.queue)),
		)
	}
}

// auditMiddleware returns Echo middleware that records audit entries for
// authenticated requests, through w. A handler that panics gets an entry
// too, with status 500. The middleware logs the panic and its stack, and
// answers 500 with no detail of the panic.
func auditMiddleware(
	w *auditWriter,
) echo.MiddlewareFunc {
	logger := w.logger

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) (err error) {
			path := c.Request().URL.Path

			for _, prefix := range excludedAuditPaths {
				if strings.HasPrefix(path, prefix) {
					return next(c)
				}
			}

			start := time.Now()

			// Read the body before the handler consumes it, and put it back. Only
			// for methods that change something: a GET carries nothing worth
			// recording, and buffering every read would cost memory for nothing.
			summary := ""
			if isMutatingMethod(c.Request().Method) {
				summary = summarizeRequest(c, logger)
			}

			// The sink collects the job ID the handler's dispatch creates, which
			// the job client fills in as it publishes, and the steps of each
			// host that answers the job.
			c.SetRequest(c.Request().WithContext(
				audit.WithJobIDSink(c.Request().Context()),
			))

			defer func() {
				r := recover()
				if r == nil {
					return
				}
				record(w, c, start, summary, http.StatusInternalServerError)
				logger.Error(
					"handler panicked",
					slog.String("path", path),
					slog.Any("panic", r),
					slog.String("stack", string(debug.Stack())),
				)
				// Echo's error handler writes the response. The panic text
				// stays in the log, never in the answer.
				err = echo.NewHTTPError(http.StatusInternalServerError, "internal error")
			}()

			err = next(c)
			record(w, c, start, summary, statusOf(c, err))

			return err
		}
	}
}

// statusOf is the status of the response to c. A handler that returns an
// error has not written its response yet, because the error handler runs
// after the chain, so the status is the one the error handler will write.
func statusOf(
	c *echo.Context,
	err error,
) int {
	// v5 returns the bare http.ResponseWriter from Response().
	// UnwrapResponse recovers Echo's wrapper, which is what records the
	// status actually written.
	resp, uerr := echo.UnwrapResponse(c.Response())
	if uerr == nil && resp.Committed {
		return resp.Status
	}
	if err != nil {
		if code := echo.StatusCode(err); code != 0 {
			return code
		}

		return http.StatusInternalServerError
	}
	if uerr == nil {
		return resp.Status
	}

	return 0
}

// record queues the entry of the request c, when it is authenticated.
func record(
	w *auditWriter,
	c *echo.Context,
	start time.Time,
	summary string,
	responseCode int,
) {
	subject, _ := c.Get(ContextKeySubject).(string)
	if subject == "" {
		return
	}

	roles, _ := c.Get(ContextKeyRoles).([]types.Role)
	targets, _ := c.Get(ContextKeyTargets).([]string)

	entry := audit.Entry{
		ID:           uuid.New().String(),
		Timestamp:    start,
		User:         subject,
		Roles:        roles,
		Targets:      targets,
		Method:       c.Request().Method,
		Path:         c.Request().URL.Path,
		SourceIP:     c.RealIP(),
		ResponseCode: responseCode,
		DurationMs:   time.Since(start).Milliseconds(),
		JobID:        audit.JobIDFrom(c.Request().Context()),
		Steps:        audit.StepsFrom(c.Request().Context()),

		RequestSummary: summary,
	}

	spanCtx := trace.SpanContextFromContext(c.Request().Context())
	if spanCtx.HasTraceID() {
		entry.TraceID = spanCtx.TraceID().String()
	}

	w.enqueue(entry)
}

// isMutatingMethod reports whether a request can change something, which is what
// makes its body worth recording.
func isMutatingMethod(
	method string,
) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	}

	return false
}

// summarizeRequest reads the request body, restores it for the handler, and
// returns what the audit log should keep of it.
//
// A body that cannot be read is reported as such rather than silently omitted: an
// audit entry that says nothing about what was asked is indistinguishable from one
// where nothing was asked.
func summarizeRequest(
	c *echo.Context,
	logger *slog.Logger,
) string {
	req := c.Request()
	if req.Body == nil {
		return ""
	}

	body, err := io.ReadAll(io.LimitReader(req.Body, maxAuditBodyBytes+1))
	if err != nil {
		logger.Warn(
			"failed to read request body for audit",
			slog.String("error", err.Error()),
			slog.String("path", req.URL.Path),
		)

		return "[body unreadable]"
	}

	// Whatever was read has to go back, or the handler sees an empty body.
	rest := req.Body
	req.Body = struct {
		io.Reader
		io.Closer
	}{
		Reader: io.MultiReader(bytes.NewReader(body), rest),
		Closer: rest,
	}

	if len(body) > maxAuditBodyBytes {
		return fmt.Sprintf("[body over %d bytes, not recorded]", maxAuditBodyBytes)
	}

	return audit.Summarize(body, req.Header.Get("Content-Type"))
}

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
	"strings"
	"time"

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

// auditMiddleware returns Echo middleware that records audit entries for
// authenticated requests. Writes are asynchronous to avoid adding latency.
func auditMiddleware(
	store audit.Store,
	logger *slog.Logger,
) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
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
			// the job client fills in as it publishes.
			c.SetRequest(c.Request().WithContext(
				audit.WithJobIDSink(c.Request().Context()),
			))

			err := next(c)

			// Only audit authenticated requests.
			subject, _ := c.Get(ContextKeySubject).(string)
			if subject == "" {
				return err
			}

			roles, _ := c.Get(ContextKeyRoles).([]string)

			// v5 returns the bare http.ResponseWriter from Response().
			// UnwrapResponse recovers Echo's wrapper, which is what
			// records the status actually written.
			responseCode := 0
			if resp, uerr := echo.UnwrapResponse(c.Response()); uerr == nil {
				responseCode = resp.Status
			}

			entry := audit.Entry{
				ID:           uuid.New().String(),
				Timestamp:    start,
				User:         subject,
				Roles:        roles,
				Method:       c.Request().Method,
				Path:         path,
				SourceIP:     c.RealIP(),
				ResponseCode: responseCode,
				DurationMs:   time.Since(start).Milliseconds(),
				JobID:        audit.JobIDFrom(c.Request().Context()),

				RequestSummary: summary,
			}

			spanCtx := trace.SpanContextFromContext(
				c.Request().Context(),
			)
			if spanCtx.HasTraceID() {
				entry.TraceID = spanCtx.TraceID().String()
			}

			// Use Background context because this goroutine runs after the HTTP
			// response is sent, at which point the request context is canceled.
			go func() {
				if writeErr := store.Write(context.Background(), entry); writeErr != nil {
					logger.Warn(
						"failed to write audit entry",
						slog.String("error", writeErr.Error()),
						slog.String("entry_id", entry.ID),
					)
				}
			}()

			return err
		}
	}
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

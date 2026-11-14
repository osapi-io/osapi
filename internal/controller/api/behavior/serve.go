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

package behavior

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"reflect"

	"github.com/labstack/echo/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/controller/api/apierr"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/validation"
)

// Result statuses, as the hand-written handlers report them.
const (
	statusOK      = "ok"
	statusFailed  = "failed"
	statusSkipped = "skipped"
	statusTimeout = "timeout"
)

// Decoder builds a method's typed input from the request, or says why the
// request is a 400. Each gen/api_<method>.go has one.
type Decoder[In any] func(c *echo.Context) (In, string, bool)

// Serve returns the handler for one method: check the target, decode the
// input, queue the job on one host or every host the target names, and
// shape each host's answer into result rows.
func Serve[In any](
	s *Server,
	b *behavior.Behavior,
	method string,
	decode Decoder[In],
) echo.HandlerFunc {
	m := b.Method(method)

	return func(c *echo.Context) error {
		target := c.Param("hostname")
		if msg, ok := validation.Var(target, "required,min=1,valid_target"); !ok {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: msg})
		}

		input, msg, ok := decode(c)
		if !ok {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: msg})
		}

		s.logger.Debug(
			"behavior",
			slog.String("operation", b.Operation(m)),
			slog.String("target", target),
			slog.Bool("broadcast", job.IsBroadcastTarget(target)),
		)

		ctx := c.Request().Context()

		// The request's span says which behavior it ran and where; the
		// agent's span for the same job adds the outcome.
		trace.SpanFromContext(ctx).SetAttributes(
			behavior.AttrBehavior.String(b.Name),
			behavior.AttrMethod.String(m.Name),
			behavior.AttrVerb.String(string(m.Verb)),
			attribute.String("osapi.target", target),
		)

		// The agent reads the category from the operation's first segment.
		op := job.OperationType(job.CategoryBehavior + "." + b.Operation(m))

		if job.IsBroadcastTarget(target) {
			jobID, responses, err := s.broadcast(ctx, m, target, op, input)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			}

			results := make([]map[string]any, 0, len(responses))
			for host, resp := range responses {
				results = append(results, rows(m, host, resp)...)
			}

			return c.JSON(http.StatusOK, Response{JobID: jobID, Results: results})
		}

		jobID, resp, err := s.single(ctx, m, target, op, input)
		if err != nil {
			if apierr.IsMissing(err) {
				return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
			}

			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		}

		return c.JSON(
			http.StatusOK,
			Response{JobID: jobID, Results: rows(m, resp.Hostname, resp)},
		)
	}
}

func (s *Server) single(
	ctx context.Context,
	m *behavior.Method,
	target string,
	op job.OperationType,
	input any,
) (string, *job.Response, error) {
	if m.Verb == behavior.VerbWrite {
		return s.jobClient.Modify(ctx, target, job.CategoryBehavior, op, input)
	}

	return s.jobClient.Query(ctx, target, job.CategoryBehavior, op, input)
}

func (s *Server) broadcast(
	ctx context.Context,
	m *behavior.Method,
	target string,
	op job.OperationType,
	input any,
) (string, map[string]*job.Response, error) {
	if m.Verb == behavior.VerbWrite {
		return s.jobClient.ModifyBroadcast(ctx, target, job.CategoryBehavior, op, input)
	}

	return s.jobClient.QueryBroadcast(ctx, target, job.CategoryBehavior, op, input)
}

// rows turns one host's response into result rows. A failed, skipped or
// timed-out host is one row with its error. A slice output is one row per
// element; anything else is one row. A host whose slice is empty is still one
// row, with only its name and status, so a host that answered is never missing
// from the results. Every row carries the host's name and status, and a
// write's row carries changed.
func rows(
	m *behavior.Method,
	host string,
	resp *job.Response,
) []map[string]any {
	if s := failureStatus(resp.Status); s != "" {
		return []map[string]any{{"hostname": host, "status": s, "error": resp.Error}}
	}

	var outputs []map[string]any
	if m.Out.Kind() == reflect.Slice {
		_ = json.Unmarshal(resp.Data, &outputs)
	} else {
		var one map[string]any
		if err := json.Unmarshal(resp.Data, &one); err == nil && one != nil {
			outputs = []map[string]any{one}
		}
	}

	if len(outputs) == 0 {
		outputs = []map[string]any{{}}
	}

	for _, row := range outputs {
		row["hostname"] = host
		row["status"] = statusOK
		if m.Verb == behavior.VerbWrite {
			if _, ok := row["changed"]; !ok && resp.Changed != nil {
				row["changed"] = *resp.Changed
			}
		}
	}

	return outputs
}

func failureStatus(
	s job.Status,
) string {
	switch s {
	case job.StatusFailed:
		return statusFailed
	case job.StatusSkipped:
		return statusSkipped
	case job.StatusTimeout:
		return statusTimeout
	default:
		return ""
	}
}

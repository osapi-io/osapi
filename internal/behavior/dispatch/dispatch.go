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

// Package dispatch runs a behavior's method for the endpoint layer: it
// sends the job to the host, or to each host, that the target names, and
// makes the answer of each host a result row. The handler that the behavior
// generator writes for each method in gen/group.go calls Run, and the
// endpoint server serves it.
package dispatch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/controller/api/apierr"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/internal/job"
)

// Result statuses, as the hand-written handlers report them.
const (
	statusOK      = "ok"
	statusFailed  = "failed"
	statusSkipped = "skipped"
	statusTimeout = "timeout"
)

// ErrInvalid is an input whose template reads a var that its behavior does
// not have, which the route's validate tags cannot see. The group of a
// method with a template field maps it to 400.
var ErrInvalid = errors.New("invalid input")

// JobClient is the part of the job client a behavior's routes use: a read
// or a write, on one host or broadcast.
type JobClient interface {
	Query(
		ctx context.Context,
		target string,
		category string,
		operation job.OperationType,
		data any,
	) (string, *job.Response, error)
	QueryBroadcast(
		ctx context.Context,
		target string,
		category string,
		operation job.OperationType,
		data any,
	) (string, map[string]*job.Response, error)
	Modify(
		ctx context.Context,
		target string,
		category string,
		operation job.OperationType,
		data any,
	) (string, *job.Response, error)
	ModifyBroadcast(
		ctx context.Context,
		target string,
		category string,
		operation job.OperationType,
		data any,
	) (string, map[string]*job.Response, error)
}

// Deps is what every behavior's endpoint group uses: the job client. The
// controller gives it to the endpoint server.
type Deps struct {
	JobClient JobClient
}

// response is the answer of every behavior method: the job, and a row for
// each host.
type response struct {
	JobID   string           `json:"job_id"`
	Results []map[string]any `json:"results"`
}

// Run sends input to method of b on the host, or the hosts, that target
// names, and returns the answer as Out, the method's response type. A host
// that does not have the item is behavior.ErrNotFound, which the group maps
// to 404. A job that cannot be sent is a 500 that does not tell its cause.
func Run[Out any](
	ctx context.Context,
	d Deps,
	b *behavior.Behavior,
	method string,
	target string,
	input any,
) (Out, error) {
	var out Out
	m := b.Method(method)
	if msg, ok := behavior.CheckTemplates(input); !ok {
		return out, fmt.Errorf("%w: %s", ErrInvalid, msg)
	}

	// The request's span says which behavior it ran and where; the agent's
	// span for the same job adds the outcome.
	trace.SpanFromContext(ctx).SetAttributes(
		behavior.AttrBehavior.String(b.Name),
		behavior.AttrMethod.String(m.Name),
		behavior.AttrVerb.String(string(m.Verb)),
		attribute.String("osapi.target", target),
	)

	// The agent reads the category from the operation's first segment.
	op := job.OperationType(job.CategoryBehavior + "." + b.Operation(m))
	write := m.Verb == behavior.VerbWrite

	var answer response
	if job.IsBroadcastTarget(target) {
		send := d.JobClient.QueryBroadcast
		if write {
			send = d.JobClient.ModifyBroadcast
		}
		jobID, responses, err := send(ctx, target, job.CategoryBehavior, op, input)
		if err != nil {
			return out, endpoint.Internal("the job could not be sent", err)
		}
		answer.JobID = jobID
		for host, resp := range responses {
			answer.Results = append(answer.Results, rows(m, host, resp)...)
		}
	} else {
		send := d.JobClient.Query
		if write {
			send = d.JobClient.Modify
		}
		jobID, resp, err := send(ctx, target, job.CategoryBehavior, op, input)
		if err != nil {
			if apierr.IsMissing(err) {
				return out, fmt.Errorf("%w: %s", behavior.ErrNotFound, err)
			}

			return out, endpoint.Internal("the job could not be sent", err)
		}
		answer.JobID = jobID
		answer.Results = rows(m, resp.Hostname, resp)
	}
	if answer.Results == nil {
		answer.Results = []map[string]any{}
	}

	// The rows are maps, and Out is the generated type with the same
	// fields, so JSON carries one to the other.
	data, _ := json.Marshal(answer)
	_ = json.Unmarshal(data, &out)

	return out, nil
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

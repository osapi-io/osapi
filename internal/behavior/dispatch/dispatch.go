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
	"maps"
	"reflect"
	"slices"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/osapi-io/osapi/internal/behavior"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/internal/job"
)

// RowStatusDoc is the description of types.RowStatus in the spec.
const RowStatusDoc = "What happened on the host: ok, the operation ran and succeeded; " +
	"failed, it ran and did not succeed; skipped, the host's OS family does not support it; " +
	"timeout, the host did not answer, so nobody knows whether it ran."

// ErrInvalid is an input that the route's validate tags cannot see as
// wrong: a template that reads a var that its behavior does not have, or an
// input that the host refuses, such as a path that is a directory. The
// group of each method maps it to 400.
var ErrInvalid = behavior.ErrInvalid

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
// names, and returns the answer as Out, the method's response type. On one
// host, a provider's behavior.ErrNotFound is a 404, its ErrNotManaged
// and ErrNotInstalled are a 409, and its ErrInvalid is a 400, as the group
// maps them; on a broadcast each
// host's row says it. A job that cannot be sent is a 500 that does not tell
// its cause.
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
	// The controller step, such as a password hash, runs here, so what the
	// job keeps has no plain secret.
	input, err := b.RunBefore(ctx, m, input)
	if err != nil {
		return out, endpoint.Internal("the controller step failed", err)
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
	write := m.Verb == types.VerbWrite

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
		// The hosts answer in the order of a map, so the rows go in
		// hostname order: the same broadcast gives the same answer.
		for _, host := range slices.Sorted(maps.Keys(responses)) {
			answer.Results = append(answer.Results, rows(m, host, responses[host])...)
		}
	} else {
		send := d.JobClient.Query
		if write {
			send = d.JobClient.Modify
		}
		jobID, resp, err := send(ctx, target, job.CategoryBehavior, op, input)
		var failed *job.ResponseError
		switch {
		case errors.Is(err, job.ErrNotFound):
			return out, fmt.Errorf("%w: %s", behavior.ErrNotFound, err)
		case errors.Is(err, job.ErrNotManaged):
			return out, fmt.Errorf("%w: %s", behavior.ErrNotManaged, err)
		case errors.Is(err, job.ErrNotInstalled):
			return out, fmt.Errorf("%w: %s", behavior.ErrNotInstalled, err)
		case errors.Is(err, job.ErrInvalid):
			return out, fmt.Errorf("%w: %s", ErrInvalid, err)
		case errors.As(err, &failed):
			// The job ran and the provider refused it: the host's row
			// says why, as a broadcast's does.
			resp = &job.Response{
				Hostname: failed.Hostname,
				Status:   types.AgentJobStatusFailed,
				Error:    failed.Message,
			}
		case err != nil:
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
	// Each value in the rows came from JSON, so they encode again.
	data, _ := json.Marshal(answer)
	if err := json.Unmarshal(data, &out); err != nil {
		return out, endpoint.Internal("the answer is not the method's output", err)
	}

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

	outputs, err := decode(m, resp.Data)
	if err != nil {
		// The host answered, but not with the method's output: the row
		// says so, and is not an ok row with no data.
		return []map[string]any{{
			"hostname": host,
			"status":   types.RowStatusFailed,
			"error":    "the answer of the host could not be read: " + err.Error(),
		}}
	}

	if len(outputs) == 0 {
		outputs = []map[string]any{{}}
	}

	for _, row := range outputs {
		row["hostname"] = host
		row["status"] = types.RowStatusOK
		if m.Verb == types.VerbWrite {
			if _, ok := row["changed"]; !ok && resp.Changed != nil {
				row["changed"] = *resp.Changed
			}
		}
	}

	return outputs
}

// decode reads a host's answer as the method's output: one row for each
// element of a slice output, and one row for anything else. An empty answer
// is no rows.
func decode(
	m *behavior.Method,
	data json.RawMessage,
) ([]map[string]any, error) {
	if len(data) == 0 || string(data) == "null" {
		return nil, nil
	}
	if m.Out.Kind() == reflect.Slice {
		var outputs []map[string]any
		if err := json.Unmarshal(data, &outputs); err != nil {
			return nil, err
		}

		return outputs, nil
	}
	var one map[string]any
	if err := json.Unmarshal(data, &one); err != nil {
		return nil, err
	}

	return []map[string]any{one}, nil
}

// failureStatus is the row status of a host whose job did not succeed, or
// empty for a host whose job did.
func failureStatus(
	s types.AgentJobStatus,
) types.RowStatus {
	switch s {
	case types.AgentJobStatusFailed, types.AgentJobStatusRefused:
		// A refused host's result is not taken, so the row is not ok.
		return types.RowStatusFailed
	case types.AgentJobStatusSkipped:
		return types.RowStatusSkipped
	case types.AgentJobStatusTimeout:
		return types.RowStatusTimeout
	default:
		return ""
	}
}

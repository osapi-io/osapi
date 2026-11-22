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
package job

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"slices"
	"time"

	sdkclient "github.com/osapi-io/osapi/pkg/sdk/client"

	"github.com/osapi-io/osapi/internal/controller/api/apierr"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/job/client"
)

// List is a page of the job queue, newest first, with the count of each
// status.
func List(
	ctx context.Context,
	d Deps,
	in ListInput,
) (ListJobsResponse, error) {
	limit := in.Limit
	if limit == 0 {
		limit = client.DefaultPageSize
	}

	result, err := d.Jobs.ListJobs(ctx, in.Status, limit, in.Offset)
	if err != nil {
		return ListJobsResponse{}, endpoint.Internal("failed to list jobs", err)
	}

	items := make([]JobDetailResponse, 0, len(result.Jobs))
	for _, qj := range result.Jobs {
		items = append(items, summary(qj))
	}

	return ListJobsResponse{
		TotalItems:   result.TotalCount,
		StatusCounts: result.StatusCounts,
		Items:        items,
	}, nil
}

// Get is one job, with what each agent did with it.
func Get(
	ctx context.Context,
	d Deps,
	in GetInput,
) (JobDetailResponse, error) {
	qj, err := d.Jobs.GetJobStatus(ctx, in.ID)
	if apierr.IsMissing(err) {
		return JobDetailResponse{}, endpoint.Errorf(ErrNotFound, "job %s not found", in.ID)
	}
	if err != nil {
		return JobDetailResponse{}, endpoint.Internal("failed to get job", err)
	}

	out := summary(qj)
	for _, hostname := range slices.Sorted(maps.Keys(qj.Responses)) {
		r := qj.Responses[hostname]
		out.Responses = append(out.Responses, Response{
			Hostname: hostname,
			Status:   string(r.Status),
			Changed:  r.Changed,
			Error:    r.Error,
			Data:     decode(r.Data),
			Steps:    r.Steps,
		})
		if len(r.Steps) > 0 {
			out.Chain = append(out.Chain, hostname+": "+sdkclient.FormatSteps(r.Steps))
		}
	}
	for _, te := range qj.Timeline {
		out.Timeline = append(out.Timeline, TimelineEvent{
			Timestamp: te.Timestamp.Format(time.RFC3339),
			Event:     te.Event,
			Hostname:  te.Hostname,
			Message:   te.Message,
			Error:     te.Error,
		})
	}
	for _, hostname := range slices.Sorted(maps.Keys(qj.AgentStates)) {
		s := qj.AgentStates[hostname]
		out.AgentStates = append(out.AgentStates, AgentState{
			Hostname: hostname,
			Status:   s.Status,
			Duration: s.Duration,
			Error:    s.Error,
		})
	}

	return out, nil
}

// Delete removes a job from the queue.
func Delete(
	ctx context.Context,
	d Deps,
	in DeleteInput,
) (struct{}, error) {
	err := d.Jobs.DeleteJob(ctx, in.ID)
	if apierr.IsMissing(err) {
		return struct{}{}, endpoint.Errorf(ErrNotFound, "job %s not found", in.ID)
	}
	if err != nil {
		return struct{}{}, endpoint.Internal("failed to delete job", err)
	}

	return struct{}{}, nil
}

// Retry sends a new job with the operation of an old one. The old job
// stays.
func Retry(
	ctx context.Context,
	d Deps,
	in RetryInput,
) (CreateJobResponse, error) {
	result, err := d.Jobs.RetryJob(ctx, in.ID, in.Body.TargetHostname)
	if apierr.IsMissing(err) {
		return CreateJobResponse{}, endpoint.Errorf(ErrNotFound, "job %s not found", in.ID)
	}
	if errors.Is(err, job.ErrNoOperationData) {
		return CreateJobResponse{}, endpoint.Errorf(ErrNoOperation,
			"job %s has no operation data to run again", in.ID)
	}
	if err != nil {
		return CreateJobResponse{}, endpoint.Internal("failed to retry job", err)
	}

	return CreateJobResponse{
		JobID:     result.JobID,
		Status:    result.Status,
		Revision:  int64(result.Revision), //nolint:gosec // a KV revision fits an int64
		Timestamp: result.Timestamp,
	}, nil
}

// summary is the fields of a job that list and get both have, with the
// secrets of its operation masked.
func summary(
	qj *job.QueuedJob,
) JobDetailResponse {
	// A job stored before targets were kept has only the host that
	// answered.
	target := qj.Target
	if target == "" {
		target = qj.Hostname
	}

	return JobDetailResponse{
		ID:        qj.ID,
		Status:    qj.Status,
		Created:   qj.Created,
		Target:    target,
		Operation: operation(qj.Operation),
		Hostname:  qj.Hostname,
		UpdatedAt: qj.UpdatedAt,
		Changed:   qj.Changed,
		Error:     qj.Error,
		Result:    decode(qj.Result),
	}
}

// operation is the stored operation of a job, with its secrets masked, as
// its type and its data. It is nil for a job with no operation.
func operation(
	stored map[string]any,
) *Operation {
	redacted := redactOperation(stored)
	if redacted == nil {
		return nil
	}
	kind, _ := redacted["type"].(string)

	return &Operation{Type: kind, Data: redacted["data"]}
}

// decode is the value of raw JSON, or nil when there is none.
func decode(
	raw json.RawMessage,
) any {
	if raw == nil {
		return nil
	}
	var v any
	_ = json.Unmarshal(raw, &v)

	return v
}

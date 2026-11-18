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
	"time"

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
	if len(qj.Responses) > 0 {
		out.Responses = make(map[string]Response, len(qj.Responses))
		for hostname, r := range qj.Responses {
			out.Responses[hostname] = Response{
				Status:   string(r.Status),
				Data:     decode(r.Data),
				Changed:  r.Changed,
				Error:    r.Error,
				Hostname: r.Hostname,
			}
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
	if len(qj.AgentStates) > 0 {
		out.AgentStates = make(map[string]AgentState, len(qj.AgentStates))
		for hostname, s := range qj.AgentStates {
			out.AgentStates[hostname] = AgentState{
				Status:   s.Status,
				Error:    s.Error,
				Duration: s.Duration,
			}
		}
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
	return JobDetailResponse{
		ID:        qj.ID,
		Status:    qj.Status,
		Created:   qj.Created,
		Operation: redactOperation(qj.Operation),
		Result:    decode(qj.Result),
		Changed:   qj.Changed,
		Error:     qj.Error,
		Hostname:  qj.Hostname,
		UpdatedAt: qj.UpdatedAt,
	}
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

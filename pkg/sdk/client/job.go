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

package client

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/pkg/sdk/client/gen"
)

// JobService provides job queue operations.
type JobService struct {
	client *gen.ClientWithResponses
}

// Get retrieves a job by ID.
func (s *JobService) Get(
	ctx context.Context,
	id string,
) (*Response[JobDetail], error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid job ID: %w", err)
	}

	resp, err := s.client.GetJobByIDWithResponse(ctx, parsedID)
	if err != nil {
		return nil, fmt.Errorf("get job: %w", err)
	}

	if err := checkError(resp.StatusCode(), resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON404, resp.JSON500); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(jobDetailFromGen(resp.JSON200), resp.Body), nil
}

// Delete deletes a job by ID.
func (s *JobService) Delete(
	ctx context.Context,
	id string,
) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid job ID: %w", err)
	}

	resp, err := s.client.DeleteJobByIDWithResponse(ctx, parsedID)
	if err != nil {
		return fmt.Errorf("delete job: %w", err)
	}

	if err := checkError(resp.StatusCode(), resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON404, resp.JSON500); err != nil {
		return err
	}

	return nil
}

// ListParams contains optional filters for listing jobs.
type ListParams struct {
	// Status filters by job status (e.g., "pending", "completed").
	Status string

	// Limit is the maximum number of results. Zero uses server default.
	Limit int

	// Offset is the number of results to skip. Zero starts from the
	// beginning.
	Offset int
}

// List retrieves jobs, optionally filtered by status.
func (s *JobService) List(
	ctx context.Context,
	params ListParams,
) (*Response[JobList], error) {
	p := &gen.GetApiJobParams{}

	if params.Status != "" {
		status := gen.GetApiJobParamsStatus(params.Status)
		p.Status = &status
	}

	if params.Limit > 0 {
		p.Limit = &params.Limit
	}

	if params.Offset > 0 {
		p.Offset = &params.Offset
	}

	resp, err := s.client.GetApiJobWithResponse(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}

	if err := checkError(resp.StatusCode(), resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON500); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(jobListFromGen(resp.JSON200), resp.Body), nil
}

// Retry retries a failed job by ID, optionally on a different target.
func (s *JobService) Retry(
	ctx context.Context,
	id string,
	target string,
) (*Response[JobCreated], error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid job ID: %w", err)
	}

	body := gen.RetryJobByIDJSONRequestBody{}
	if target != "" {
		body.TargetHostname = &target
	}

	resp, err := s.client.RetryJobByIDWithResponse(ctx, parsedID, body)
	if err != nil {
		return nil, fmt.Errorf("retry job: %w", err)
	}

	if err := checkError(resp.StatusCode(), resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON404, resp.JSON500); err != nil {
		return nil, err
	}

	if resp.JSON201 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(jobCreatedFromGen(resp.JSON201), resp.Body), nil
}

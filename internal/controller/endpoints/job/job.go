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
// Package job is the endpoint group for the job queue: the jobs the
// controller sent to the agents. Until the group moves, the controller
// serves the hand-written spec in internal/controller/api/job, and
// TestEndpointParity holds the two together.
package job

import (
	"errors"
	"net/http"

	"github.com/osapi-io/osapi/internal/controller/endpoint"
)

// ErrNotFound is a job that is not there.
var ErrNotFound = errors.New("job not found")

// Group is the job group's definition.
var Group = endpoint.Register(endpoint.Define(
	"job", "The job queue",
	endpoint.Title("Job Management API"),
	endpoint.Tag("job_operations", "Job", "Operations related to the job queue."),
	endpoint.Example[ListJobsResponse]("status_counts",
		`{"submitted":5,"processing":2,"completed":30,"failed":3,"partial_failure":2}`),
	endpoint.Get[ListInput, ListJobsResponse]("list").
		ID("getJobs").
		Describe("List jobs").
		Explain("Retrieve jobs, optionally filtered by status.").
		Returns("A list of jobs."),
	endpoint.Get[GetInput, JobDetailResponse]("get").
		ID("GetJobByID").
		Describe("Get job detail").
		Explain("Retrieve details of a specific job by its ID.").
		Returns("The job details.").
		Errors(endpoint.NotFound(ErrNotFound, "Job not found.")),
	endpoint.Delete[DeleteInput, struct{}]("delete").
		ID("DeleteJobByID").
		Describe("Delete a job").
		Explain("Delete a specific job by its ID.").
		Returns("Job deleted successfully. No content returned.").
		Status(http.StatusNoContent).
		Errors(endpoint.NotFound(ErrNotFound, "Job not found.")),
	endpoint.Post[RetryInput, CreateJobResponse]("retry").
		Action("retry").
		ID("RetryJobByID").
		Describe("Retry a job").
		Explain("Create a new job using the same operation data as an existing job. "+
			"The original job is preserved. Returns the new job ID.").
		Returns("The retry job was created successfully.").
		Status(http.StatusCreated).
		Errors(endpoint.NotFound(ErrNotFound, "Job not found.")),
))

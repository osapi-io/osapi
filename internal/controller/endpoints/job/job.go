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
// controller sent to the agents.
package job

import (
	"errors"
	"net/http"

	"github.com/osapi-io/osapi/internal/area"
	"github.com/osapi-io/osapi/internal/controller/endpoint"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

var (
	// ErrNotFound is a job that is not there.
	ErrNotFound = errors.New("job not found")
	// ErrNoOperation is a job with no operation to run again.
	ErrNoOperation = errors.New("job has no operation data")
	// ErrOutOfScope is a retry to a target outside the token's target scope.
	ErrOutOfScope = errors.New("target is outside the token's target scope")
)

// Group is the job group's definition.
var Group = endpoint.Register(endpoint.Define(
	"job", "Manage the jobs in the queue",
	endpoint.Area(area.Fleet),
	endpoint.Title("Job Management API"),
	endpoint.Tag("job_operations", "Job", "Operations related to the job queue."),
	endpoint.Uses[Deps](),
	endpoint.Doc[types.JobStatus](
		"The status of a job, which the controller computes from what each agent did: "+
			"submitted, processing, completed, failed, partial_failure (completed on some "+
			"agents and failed on others) or skipped (each agent skipped it).",
	),
	endpoint.Doc[types.AgentJobStatus](
		"The status of a job on one host. The controller records submitted and retried; "+
			"an agent records acknowledged, started, completed, failed or skipped; and timeout "+
			"is a host whose agent did not answer in time.",
	),
	endpoint.Example[ListJobsResponse]("status_counts",
		`{"submitted":5,"processing":2,"completed":30,"failed":3,"partial_failure":2}`),
	endpoint.Get[ListInput, ListJobsResponse]("list").
		Fails("a status that is not a job status", ListInput{Status: "lost"}).
		Fails("a limit below 1", ListInput{Limit: -1}).
		Fails("a limit over 100", ListInput{Limit: 101}).
		Fails("a negative offset", ListInput{Offset: -1}).
		Handle(List).
		ID("GetJobs").
		Describe("List jobs").
		Explain("Retrieve jobs, optionally filtered by status. A token with a target scope gets "+
			"only the jobs whose target is in its scope, and the counts include only those jobs.").
		Returns("A list of jobs.").
		Command("List jobs", "").
		Example("").
		Prints(sampleList).
		Example("--status completed --limit 3"),
	endpoint.Get[GetInput, DetailResponse]("get").
		Fails("an ID that is not a UUID", GetInput{ID: "not-a-uuid"}).
		Handle(Get).
		ID("GetJobByID").
		Describe("Get job detail").
		Explain("Retrieve details of a specific job by its ID. A job whose target is outside the "+
			"target scope of the token is not found.").
		Returns("The job details.").
		Command("Get job details and status", "").
		Example("--id 550e8400-e29b-41d4-a716-446655440000").
		Prints(sampleDetail).
		Errors(endpoint.NotFound(ErrNotFound,
			"Job not found, or its target is outside the token's target scope.")),
	endpoint.Delete[DeleteInput, struct{}]("delete").
		Fails("an ID that is not a UUID", DeleteInput{ID: "not-a-uuid"}).
		Handle(Delete).
		ID("DeleteJobByID").
		Describe("Delete a job").
		Explain("Delete a specific job by its ID. A job whose target is outside the target "+
			"scope of the token is not found.").
		Returns("Job deleted successfully. No content returned.").
		Status(http.StatusNoContent).
		Command("Delete a job", "The job goes from storage permanently.").
		Example("--id 550e8400-e29b-41d4-a716-446655440000").
		Errors(endpoint.NotFound(ErrNotFound,
			"Job not found, or its target is outside the token's target scope.")),
	endpoint.Post[RetryInput, CreateJobResponse]("retry").
		Fails("an ID that is not a UUID", RetryInput{ID: "not-a-uuid"}).
		Fails("a target that is not a hostname, a label or _any or _all", RetryInput{ID: "7c9e6679-7425-40de-944b-e07fc1f90ae7", Body: RetryJobRequest{TargetHostname: "web 01!"}}).
		Handle(Retry).
		Action("retry").
		ID("RetryJobByID").
		Describe("Retry a job").
		Explain("Create a new job using the same operation data as an existing job. "+
			"The original job is preserved. Returns the new job ID.").
		Returns("The retry job was created successfully.").
		Status(http.StatusCreated).
		Location("job_id").
		Command("Retry a failed or stuck job",
			"The new job has the same operation as the old job. The original job stays in\n"+
				"the history. Its timeline gets a `retried` event that links to the new job.").
		Example("--id 550e8400-e29b-41d4-a716-446655440000").
		Prints(sampleRetry).
		Errors(
			endpoint.NotFound(ErrNotFound,
				"Job not found, or its target is outside the token's target scope."),
			endpoint.Invalid(ErrNoOperation,
				"Invalid request parameters, or a job with no operation data to run again."),
			endpoint.Forbidden(ErrOutOfScope,
				"Insufficient permissions, or a target outside the token's target scope."),
		),
))

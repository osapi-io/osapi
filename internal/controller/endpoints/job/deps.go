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

	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/job/client"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// Deps is what the job group's handlers use.
type Deps struct {
	// Jobs reads, deletes and retries the jobs in the queue.
	Jobs Jobs
}

// Jobs is the part of the job client that the job group uses.
type Jobs interface {
	ListJobs(
		ctx context.Context,
		statusFilter types.JobStatus,
		limit int,
		offset int,
		keep func(target string) bool,
	) (*client.ListJobsResult, error)
	GetJobStatus(ctx context.Context, jobID string) (*job.QueuedJob, error)
	DeleteJob(ctx context.Context, jobID string) error
	RetryJob(
		ctx context.Context,
		jobID string,
		targetHostname string,
	) (*client.CreateJobResult, error)
}

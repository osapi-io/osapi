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
	"log/slog"
	"strings"

	"github.com/retr0h/osapi/internal/api/job/gen"
	"github.com/retr0h/osapi/internal/validation"
)

// RetryJobByID creates a new job using the same operation data as an existing job.
func (j *Job) RetryJobByID(
	ctx context.Context,
	request gen.RetryJobByIDRequestObject,
) (gen.RetryJobByIDResponseObject, error) {
	jobID := request.Id.String()

	if request.Body != nil {
		if errMsg, ok := validation.Struct(request.Body); !ok {
			return gen.RetryJobByID400JSONResponse{Error: &errMsg}, nil
		}
	}

	var targetHostname string
	if request.Body != nil && request.Body.TargetHostname != nil {
		targetHostname = *request.Body.TargetHostname
	}

	j.logger.Debug("retrying job",
		slog.String("job_id", jobID),
		slog.String("target_hostname", targetHostname),
	)

	result, err := j.JobClient.RetryJob(ctx, jobID, targetHostname)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") {
			return gen.RetryJobByID404JSONResponse{
				Error: &errMsg,
			}, nil
		}
		if strings.Contains(errMsg, "no operation data") {
			return gen.RetryJobByID400JSONResponse{
				Error: &errMsg,
			}, nil
		}
		return gen.RetryJobByID500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	revision := int64(result.Revision)
	return gen.RetryJobByID201JSONResponse{
		JobId:     result.JobID,
		Status:    result.Status,
		Revision:  &revision,
		Timestamp: &result.Timestamp,
	}, nil
}

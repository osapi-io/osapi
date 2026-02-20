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

	"github.com/retr0h/osapi/internal/api/job/gen"
	"github.com/retr0h/osapi/internal/validation"
)

// PostJob creates a new job in the queue.
func (j *Job) PostJob(
	ctx context.Context,
	request gen.PostJobRequestObject,
) (gen.PostJobResponseObject, error) {
	if errMsg, ok := validation.Struct(request.Body); !ok {
		return gen.PostJob400JSONResponse{
			Error: &errMsg,
		}, nil
	}

	if opType, ok := request.Body.Operation["type"].(string); ok {
		j.logger.Debug("creating job",
			slog.String("operation_type", opType),
			slog.String("target", request.Body.TargetHostname),
		)
	}

	result, err := j.JobClient.CreateJob(ctx, request.Body.Operation, request.Body.TargetHostname)
	if err != nil {
		errMsg := err.Error()
		return gen.PostJob500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	revision := int64(result.Revision)
	return gen.PostJob201JSONResponse{
		JobId:     result.JobID,
		Status:    result.Status,
		Revision:  &revision,
		Timestamp: &result.Timestamp,
	}, nil
}

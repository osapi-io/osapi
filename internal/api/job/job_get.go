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
	"strings"

	"github.com/retr0h/osapi/internal/api/job/gen"
)

// GetJobByID retrieves details of a specific job by its ID.
func (j *Job) GetJobByID(
	ctx context.Context,
	request gen.GetJobByIDRequestObject,
) (gen.GetJobByIDResponseObject, error) {
	qj, err := j.JobClient.GetJobStatus(ctx, request.Id)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") {
			return gen.GetJobByID404JSONResponse{
				Error: &errMsg,
			}, nil
		}
		return gen.GetJobByID500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	resp := gen.GetJobByID200JSONResponse{
		Id:      &qj.ID,
		Status:  &qj.Status,
		Created: &qj.Created,
	}
	if qj.Operation != nil {
		op := map[string]interface{}(qj.Operation)
		resp.Operation = &op
	}
	if qj.Error != "" {
		resp.Error = &qj.Error
	}
	if qj.Hostname != "" {
		resp.Hostname = &qj.Hostname
	}
	if qj.UpdatedAt != "" {
		resp.UpdatedAt = &qj.UpdatedAt
	}
	if qj.Result != nil {
		var result interface{}
		_ = json.Unmarshal(qj.Result, &result)
		resp.Result = result
	}

	return resp, nil
}

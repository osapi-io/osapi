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

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/api/job/gen"
	"github.com/retr0h/osapi/internal/validation"
)

// GetJob lists jobs, optionally filtered by status.
func (j *Job) GetJob(
	ctx context.Context,
	request gen.GetJobRequestObject,
) (gen.GetJobResponseObject, error) {
	if errMsg, ok := validation.Struct(request.Params); !ok {
		return gen.GetJob400JSONResponse{Error: &errMsg}, nil
	}

	var statusFilter string
	if request.Params.Status != nil {
		statusFilter = string(*request.Params.Status)
	}

	limit := 10
	if request.Params.Limit != nil {
		limit = *request.Params.Limit
	}

	offset := 0
	if request.Params.Offset != nil {
		offset = *request.Params.Offset
	}

	result, err := j.JobClient.ListJobs(ctx, statusFilter, limit, offset)
	if err != nil {
		errMsg := err.Error()
		return gen.GetJob500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	items := make([]gen.JobDetailResponse, 0, len(result.Jobs))
	for _, qj := range result.Jobs {
		jobUUID := uuid.MustParse(qj.ID)
		item := gen.JobDetailResponse{
			Id:      &jobUUID,
			Status:  &qj.Status,
			Created: &qj.Created,
		}
		if qj.Operation != nil {
			op := map[string]interface{}(qj.Operation)
			item.Operation = &op
		}
		if qj.Error != "" {
			item.Error = &qj.Error
		}
		if qj.Hostname != "" {
			item.Hostname = &qj.Hostname
		}
		if qj.UpdatedAt != "" {
			item.UpdatedAt = &qj.UpdatedAt
		}
		if qj.Result != nil {
			var result interface{}
			_ = json.Unmarshal(qj.Result, &result)
			item.Result = result
		}
		items = append(items, item)
	}

	totalItems := result.TotalCount
	return gen.GetJob200JSONResponse{
		TotalItems: &totalItems,
		Items:      &items,
	}, nil
}

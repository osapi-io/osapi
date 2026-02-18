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
	"strings"

	"github.com/retr0h/osapi/internal/api/job/gen"
	"github.com/retr0h/osapi/internal/validation"
)

// DeleteJobByID deletes a specific job by its ID.
func (j *Job) DeleteJobByID(
	ctx context.Context,
	request gen.DeleteJobByIDRequestObject,
) (gen.DeleteJobByIDResponseObject, error) {
	id := struct {
		ID string `validate:"required,uuid"`
	}{ID: request.Id}
	if errMsg, ok := validation.Struct(id); !ok {
		return gen.DeleteJobByID400JSONResponse{Error: &errMsg}, nil
	}

	err := j.JobClient.DeleteJob(ctx, request.Id)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") {
			return gen.DeleteJobByID404JSONResponse{
				Error: &errMsg,
			}, nil
		}
		return gen.DeleteJobByID500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	return gen.DeleteJobByID204Response{}, nil
}

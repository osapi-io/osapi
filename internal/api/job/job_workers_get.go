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

	"github.com/retr0h/osapi/internal/api/job/gen"
)

// GetJobWorkers discovers all active workers in the fleet.
func (j *Job) GetJobWorkers(
	ctx context.Context,
	_ gen.GetJobWorkersRequestObject,
) (gen.GetJobWorkersResponseObject, error) {
	workers, err := j.JobClient.ListWorkers(ctx)
	if err != nil {
		errMsg := err.Error()
		return gen.GetJobWorkers500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	workerInfos := make([]gen.WorkerInfo, 0, len(workers))
	for _, w := range workers {
		workerInfos = append(workerInfos, gen.WorkerInfo{
			Hostname: w.Hostname,
		})
	}

	total := len(workerInfos)

	return gen.GetJobWorkers200JSONResponse{
		Workers: workerInfos,
		Total:   total,
	}, nil
}

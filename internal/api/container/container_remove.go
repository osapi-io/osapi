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

package container

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/api/container/gen"
	"github.com/retr0h/osapi/internal/job"
)

// DeleteNodeContainerById removes a container from a target node.
func (s *Container) DeleteNodeContainerById(
	ctx context.Context,
	request gen.DeleteNodeContainerByIdRequestObject,
) (gen.DeleteNodeContainerByIdResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.DeleteNodeContainerById400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname
	id := request.Id

	data := &job.ContainerRemoveData{}
	if request.Params.Force != nil {
		data.Force = *request.Params.Force
	}

	s.logger.Debug("container remove",
		slog.String("target", hostname),
		slog.String("id", id),
		slog.Bool("force", data.Force),
	)

	resp, err := s.JobClient.ModifyContainerRemove(ctx, hostname, id, data)
	if err != nil {
		errMsg := err.Error()
		return gen.DeleteNodeContainerById500JSONResponse{Error: &errMsg}, nil
	}

	jobUUID := uuid.MustParse(resp.JobID)
	changed := resp.Changed
	msg := "container removed"

	return gen.DeleteNodeContainerById202JSONResponse{
		JobId: &jobUUID,
		Results: []gen.ContainerActionResultItem{
			{
				Hostname: resp.Hostname,
				Id:       &id,
				Changed:  changed,
				Message:  &msg,
			},
		},
	}, nil
}

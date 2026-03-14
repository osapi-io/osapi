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

	"github.com/retr0h/osapi/internal/api/docker/gen"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/validation"
)

// DeleteNodeContainerDockerImage removes a container image from a target node.
func (s *Container) DeleteNodeContainerDockerImage(
	ctx context.Context,
	request gen.DeleteNodeContainerDockerImageRequestObject,
) (gen.DeleteNodeContainerDockerImageResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.DeleteNodeContainerDockerImage400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validation.Var(request.Image, "required,min=1"); !ok {
		return gen.DeleteNodeContainerDockerImage400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validation.Struct(request.Params); !ok {
		return gen.DeleteNodeContainerDockerImage400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname
	imageName := request.Image

	data := &job.DockerImageRemoveData{
		Image: imageName,
	}
	if request.Params.Force != nil {
		data.Force = *request.Params.Force
	}

	s.logger.Debug("container image remove",
		slog.String("target", hostname),
		slog.String("image", imageName),
		slog.Bool("force", data.Force),
	)

	resp, err := s.JobClient.ModifyDockerImageRemove(ctx, hostname, data)
	if err != nil {
		errMsg := err.Error()
		return gen.DeleteNodeContainerDockerImage500JSONResponse{Error: &errMsg}, nil
	}

	jobUUID := uuid.MustParse(resp.JobID)
	changed := resp.Changed
	msg := "image removed"

	return gen.DeleteNodeContainerDockerImage202JSONResponse{
		JobId: &jobUUID,
		Results: []gen.DockerActionResultItem{
			{
				Hostname: resp.Hostname,
				Id:       &imageName,
				Changed:  changed,
				Message:  &msg,
			},
		},
	}, nil
}

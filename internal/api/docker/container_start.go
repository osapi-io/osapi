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
)

// PostNodeContainerDockerStart starts a container on a target node.
func (s *Container) PostNodeContainerDockerStart(
	ctx context.Context,
	request gen.PostNodeContainerDockerStartRequestObject,
) (gen.PostNodeContainerDockerStartResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.PostNodeContainerDockerStart400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname
	id := request.Id

	s.logger.Debug("container start",
		slog.String("target", hostname),
		slog.String("id", id),
	)

	resp, err := s.JobClient.ModifyDockerStart(ctx, hostname, id)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeContainerDockerStart500JSONResponse{Error: &errMsg}, nil
	}

	jobUUID := uuid.MustParse(resp.JobID)
	changed := resp.Changed
	msg := "container started"

	return gen.PostNodeContainerDockerStart202JSONResponse{
		JobId: &jobUUID,
		Results: []gen.DockerActionResultItem{
			{
				Hostname: resp.Hostname,
				Id:       &id,
				Changed:  changed,
				Message:  &msg,
			},
		},
	}, nil
}

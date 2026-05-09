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

	"github.com/retr0h/osapi/internal/controller/api/node/docker/gen"
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

	// Defense in depth: Force *bool with omitempty cannot currently
	// fail validation, but guards against future parameter additions.
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

	s.logger.Debug(
		"container image remove",
		slog.String("target", hostname),
		slog.String("image", imageName),
		slog.Bool("force", data.Force),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.deleteNodeContainerDockerImageRemoveBroadcast(ctx, hostname, imageName, data)
	}

	jobID, resp, err := s.JobClient.Modify(
		ctx,
		hostname,
		"docker",
		job.OperationDockerImageRemove,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.DeleteNodeContainerDockerImage500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		jobUUID := uuid.MustParse(jobID)
		e := resp.Error
		return gen.DeleteNodeContainerDockerImage202JSONResponse{
			JobId: &jobUUID,
			Results: []gen.DockerActionResultItem{
				{
					Hostname: resp.Hostname,
					Status:   gen.DockerActionResultItemStatusSkipped,
					Id:       &imageName,
					Error:    &e,
				},
			},
		}, nil
	}

	jobUUID := uuid.MustParse(jobID)
	changed := resp.Changed
	msg := "image removed"

	return gen.DeleteNodeContainerDockerImage202JSONResponse{
		JobId: &jobUUID,
		Results: []gen.DockerActionResultItem{
			{
				Hostname: resp.Hostname,
				Status:   gen.DockerActionResultItemStatusOk,
				Id:       &imageName,
				Changed:  changed,
				Message:  &msg,
			},
		},
	}, nil
}

// deleteNodeContainerDockerImageRemoveBroadcast handles broadcast targets for image remove.
func (s *Container) deleteNodeContainerDockerImageRemoveBroadcast(
	ctx context.Context,
	target string,
	imageName string,
	data *job.DockerImageRemoveData,
) (gen.DeleteNodeContainerDockerImageResponseObject, error) {
	jobID, responses, err := s.JobClient.ModifyBroadcast(
		ctx,
		target,
		"docker",
		job.OperationDockerImageRemove,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.DeleteNodeContainerDockerImage500JSONResponse{Error: &errMsg}, nil
	}

	msg := "image removed"
	var items []gen.DockerActionResultItem
	for host, resp := range responses {
		item := gen.DockerActionResultItem{
			Hostname: host,
			Id:       &imageName,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.DockerActionResultItemStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.DockerActionResultItemStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.DockerActionResultItemStatusOk
			item.Changed = resp.Changed
			item.Message = &msg
		}
		items = append(items, item)
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.DeleteNodeContainerDockerImage202JSONResponse{
		JobId:   &jobUUID,
		Results: items,
	}, nil
}

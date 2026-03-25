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

	"github.com/retr0h/osapi/internal/controller/api/docker/gen"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/validation"
)

// DeleteNodeContainerDockerByID removes a container from a target node.
func (s *Container) DeleteNodeContainerDockerByID(
	ctx context.Context,
	request gen.DeleteNodeContainerDockerByIDRequestObject,
) (gen.DeleteNodeContainerDockerByIDResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.DeleteNodeContainerDockerByID400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validation.Var(request.Id, "required,min=1"); !ok {
		return gen.DeleteNodeContainerDockerByID400JSONResponse{Error: &errMsg}, nil
	}

	// Defense in depth: Force *bool with omitempty cannot currently
	// fail validation, but guards against future parameter additions.
	if errMsg, ok := validation.Struct(request.Params); !ok {
		return gen.DeleteNodeContainerDockerByID400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname
	id := request.Id

	data := &job.DockerRemoveData{}
	if request.Params.Force != nil {
		data.Force = *request.Params.Force
	}

	s.logger.Debug("container remove",
		slog.String("target", hostname),
		slog.String("id", id),
		slog.Bool("force", data.Force),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.deleteNodeContainerDockerRemoveBroadcast(ctx, hostname, id, data)
	}

	resp, err := s.JobClient.ModifyDockerRemove(ctx, hostname, id, data)
	if err != nil {
		errMsg := err.Error()
		return gen.DeleteNodeContainerDockerByID500JSONResponse{Error: &errMsg}, nil
	}

	jobUUID := uuid.MustParse(resp.JobID)
	changed := resp.Changed
	msg := "container removed"

	return gen.DeleteNodeContainerDockerByID202JSONResponse{
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

// deleteNodeContainerDockerRemoveBroadcast handles broadcast targets for container remove.
func (s *Container) deleteNodeContainerDockerRemoveBroadcast(
	ctx context.Context,
	target string,
	id string,
	data *job.DockerRemoveData,
) (gen.DeleteNodeContainerDockerByIDResponseObject, error) {
	jobID, results, errs, err := s.JobClient.ModifyDockerRemoveBroadcast(ctx, target, id, data)
	if err != nil {
		errMsg := err.Error()
		return gen.DeleteNodeContainerDockerByID500JSONResponse{Error: &errMsg}, nil
	}

	msg := "container removed"
	var responses []gen.DockerActionResultItem
	for _, resp := range results {
		responses = append(responses, gen.DockerActionResultItem{
			Hostname: resp.Hostname,
			Id:       &id,
			Changed:  resp.Changed,
			Message:  &msg,
		})
	}
	for hostname, errMsg := range errs {
		e := errMsg
		responses = append(responses, gen.DockerActionResultItem{
			Hostname: hostname,
			Error:    &e,
		})
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.DeleteNodeContainerDockerByID202JSONResponse{
		JobId:   &jobUUID,
		Results: responses,
	}, nil
}

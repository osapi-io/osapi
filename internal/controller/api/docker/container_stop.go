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

// PostNodeContainerDockerStop stops a container on a target node.
func (s *Container) PostNodeContainerDockerStop(
	ctx context.Context,
	request gen.PostNodeContainerDockerStopRequestObject,
) (gen.PostNodeContainerDockerStopResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.PostNodeContainerDockerStop400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validation.Var(request.Id, "required,min=1"); !ok {
		return gen.PostNodeContainerDockerStop400JSONResponse{Error: &errMsg}, nil
	}

	if request.Body != nil {
		if errMsg, ok := validation.Struct(request.Body); !ok {
			return gen.PostNodeContainerDockerStop400JSONResponse{Error: &errMsg}, nil
		}
	}

	hostname := request.Hostname
	id := request.Id

	data := &job.DockerStopData{}
	if request.Body != nil && request.Body.Timeout != nil {
		data.Timeout = request.Body.Timeout
	}

	s.logger.Debug("container stop",
		slog.String("target", hostname),
		slog.String("id", id),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.postNodeContainerDockerStopBroadcast(ctx, hostname, id, data)
	}

	stopData := struct {
		ID      string `json:"id"`
		Timeout *int   `json:"timeout,omitempty"`
	}{
		ID:      id,
		Timeout: data.Timeout,
	}
	jobID, resp, err := s.JobClient.Modify(
		ctx,
		hostname,
		"docker",
		job.OperationDockerStop,
		stopData,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeContainerDockerStop500JSONResponse{Error: &errMsg}, nil
	}

	jobUUID := uuid.MustParse(jobID)
	changed := resp.Changed
	msg := "container stopped"

	return gen.PostNodeContainerDockerStop202JSONResponse{
		JobId: &jobUUID,
		Results: []gen.DockerActionResultItem{
			{
				Hostname: resp.Hostname,
				Status:   gen.DockerActionResultItemStatusOk,
				Id:       &id,
				Changed:  changed,
				Message:  &msg,
			},
		},
	}, nil
}

// postNodeContainerDockerStopBroadcast handles broadcast targets for container stop.
func (s *Container) postNodeContainerDockerStopBroadcast(
	ctx context.Context,
	target string,
	id string,
	data *job.DockerStopData,
) (gen.PostNodeContainerDockerStopResponseObject, error) {
	stopData := struct {
		ID      string `json:"id"`
		Timeout *int   `json:"timeout,omitempty"`
	}{
		ID:      id,
		Timeout: data.Timeout,
	}
	jobID, responses, err := s.JobClient.ModifyBroadcast(
		ctx,
		target,
		"docker",
		job.OperationDockerStop,
		stopData,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeContainerDockerStop500JSONResponse{Error: &errMsg}, nil
	}

	msg := "container stopped"
	var items []gen.DockerActionResultItem
	for host, resp := range responses {
		item := gen.DockerActionResultItem{
			Hostname: host,
			Id:       &id,
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
	return gen.PostNodeContainerDockerStop202JSONResponse{
		JobId:   &jobUUID,
		Results: items,
	}, nil
}

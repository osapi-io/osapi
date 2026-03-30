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
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/node/docker/gen"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/validation"
)

// PostNodeContainerDocker creates a container on a target node.
func (s *Container) PostNodeContainerDocker(
	ctx context.Context,
	request gen.PostNodeContainerDockerRequestObject,
) (gen.PostNodeContainerDockerResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.PostNodeContainerDocker400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validation.Struct(request.Body); !ok {
		return gen.PostNodeContainerDocker400JSONResponse{Error: &errMsg}, nil
	}

	data := &job.DockerCreateData{
		Image:   request.Body.Image,
		Command: ptrToSlice(request.Body.Command),
		Env:     envSliceToMap(request.Body.Env),
		Ports:   parsePortMappings(request.Body.Ports),
		Volumes: parseVolumeMappings(request.Body.Volumes),
	}
	if request.Body.Name != nil {
		data.Name = *request.Body.Name
	}
	if request.Body.AutoStart != nil {
		data.AutoStart = *request.Body.AutoStart
	} else {
		data.AutoStart = true
	}

	hostname := request.Hostname

	s.logger.Debug("container create",
		slog.String("image", data.Image),
		slog.String("target", hostname),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.postNodeContainerDockerCreateBroadcast(ctx, hostname, data)
	}

	jobID, resp, err := s.JobClient.Modify(ctx, hostname, "docker", job.OperationDockerCreate, data)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeContainerDocker500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		jobUUID := uuid.MustParse(jobID)
		e := resp.Error
		return gen.PostNodeContainerDocker202JSONResponse{
			JobId: &jobUUID,
			Results: []gen.DockerResponse{
				{
					Hostname: resp.Hostname,
					Status:   gen.DockerResponseStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	var containerResp struct {
		ID string `json:"id"`
	}
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &containerResp)
	}

	jobUUID := uuid.MustParse(jobID)
	id := containerResp.ID
	changed := resp.Changed

	return gen.PostNodeContainerDocker202JSONResponse{
		JobId: &jobUUID,
		Results: []gen.DockerResponse{
			{
				Hostname: resp.Hostname,
				Status:   gen.DockerResponseStatusOk,
				Id:       stringPtrOrNil(id),
				Image:    &data.Image,
				Changed:  changed,
			},
		},
	}, nil
}

// postNodeContainerDockerCreateBroadcast handles broadcast targets for container create.
func (s *Container) postNodeContainerDockerCreateBroadcast(
	ctx context.Context,
	target string,
	data *job.DockerCreateData,
) (gen.PostNodeContainerDockerResponseObject, error) {
	jobID, responses, err := s.JobClient.ModifyBroadcast(
		ctx,
		target,
		"docker",
		job.OperationDockerCreate,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeContainerDocker500JSONResponse{Error: &errMsg}, nil
	}

	var items []gen.DockerResponse
	for host, resp := range responses {
		item := gen.DockerResponse{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.DockerResponseStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.DockerResponseStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.DockerResponseStatusOk
			var containerResp struct {
				ID string `json:"id"`
			}
			if resp.Data != nil {
				_ = json.Unmarshal(resp.Data, &containerResp)
			}
			id := containerResp.ID
			item.Id = stringPtrOrNil(id)
			item.Image = &data.Image
			item.Changed = resp.Changed
		}
		items = append(items, item)
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.PostNodeContainerDocker202JSONResponse{
		JobId:   &jobUUID,
		Results: items,
	}, nil
}

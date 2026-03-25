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

	"github.com/retr0h/osapi/internal/controller/api/docker/gen"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/validation"
)

// GetNodeContainerDockerByID inspects a container on a target node.
func (s *Container) GetNodeContainerDockerByID(
	ctx context.Context,
	request gen.GetNodeContainerDockerByIDRequestObject,
) (gen.GetNodeContainerDockerByIDResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeContainerDockerByID400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validation.Var(request.Id, "required,min=1"); !ok {
		return gen.GetNodeContainerDockerByID400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname
	id := request.Id

	s.logger.Debug("container inspect",
		slog.String("target", hostname),
		slog.String("id", id),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.getNodeContainerDockerInspectBroadcast(ctx, hostname, id)
	}

	resp, err := s.JobClient.QueryDockerInspect(ctx, hostname, id)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeContainerDockerByID500JSONResponse{Error: &errMsg}, nil
	}

	item := dockerDetailItemFromResponse(resp)
	jobUUID := uuid.MustParse(resp.JobID)

	return gen.GetNodeContainerDockerByID200JSONResponse{
		JobId:   &jobUUID,
		Results: []gen.DockerDetailResponse{item},
	}, nil
}

// dockerDetailItemFromResponse builds a DockerDetailResponse from a job response.
func dockerDetailItemFromResponse(
	resp *job.Response,
) gen.DockerDetailResponse {
	var detail struct {
		ID              string `json:"id"`
		Name            string `json:"name"`
		Image           string `json:"image"`
		State           string `json:"state"`
		Created         string `json:"created"`
		Health          string `json:"health,omitempty"`
		NetworkSettings *struct {
			IPAddress string `json:"ip_address,omitempty"`
			Gateway   string `json:"gateway,omitempty"`
		} `json:"network_settings,omitempty"`
		Ports []struct {
			Host      int `json:"host"`
			Container int `json:"container"`
		} `json:"ports,omitempty"`
		Mounts []struct {
			Host      string `json:"host"`
			Container string `json:"container"`
		} `json:"mounts,omitempty"`
	}
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &detail)
	}

	ports := portMappingsToStrings(detail.Ports)
	mounts := volumeMappingsToStrings(detail.Mounts)

	var networkSettings *map[string]string
	if detail.NetworkSettings != nil {
		ns := map[string]string{}
		if detail.NetworkSettings.IPAddress != "" {
			ns["ip_address"] = detail.NetworkSettings.IPAddress
		}
		if detail.NetworkSettings.Gateway != "" {
			ns["gateway"] = detail.NetworkSettings.Gateway
		}
		if len(ns) > 0 {
			networkSettings = &ns
		}
	}

	return gen.DockerDetailResponse{
		Hostname:        resp.Hostname,
		Id:              stringPtrOrNil(detail.ID),
		Name:            stringPtrOrNil(detail.Name),
		Image:           stringPtrOrNil(detail.Image),
		State:           stringPtrOrNil(detail.State),
		Created:         stringPtrOrNil(detail.Created),
		Health:          stringPtrOrNil(detail.Health),
		Ports:           nilIfEmptyStrSlice(ports),
		Mounts:          nilIfEmptyStrSlice(mounts),
		NetworkSettings: networkSettings,
		Changed:         resp.Changed,
	}
}

// getNodeContainerDockerInspectBroadcast handles broadcast targets for container inspect.
func (s *Container) getNodeContainerDockerInspectBroadcast(
	ctx context.Context,
	target string,
	id string,
) (gen.GetNodeContainerDockerByIDResponseObject, error) {
	jobID, results, errs, err := s.JobClient.QueryDockerInspectBroadcast(ctx, target, id)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeContainerDockerByID500JSONResponse{Error: &errMsg}, nil
	}

	var responses []gen.DockerDetailResponse
	for _, resp := range results {
		responses = append(responses, dockerDetailItemFromResponse(resp))
	}
	for hostname, errMsg := range errs {
		e := errMsg
		responses = append(responses, gen.DockerDetailResponse{
			Hostname: hostname,
			Error:    &e,
		})
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.GetNodeContainerDockerByID200JSONResponse{
		JobId:   &jobUUID,
		Results: responses,
	}, nil
}

// nilIfEmptyStrSlice returns nil if the slice is empty, otherwise a pointer.
func nilIfEmptyStrSlice(
	s []string,
) *[]string {
	if len(s) == 0 {
		return nil
	}
	return &s
}

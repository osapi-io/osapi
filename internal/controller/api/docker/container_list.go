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

// GetNodeContainerDocker lists containers on a target node.
func (s *Container) GetNodeContainerDocker(
	ctx context.Context,
	request gen.GetNodeContainerDockerRequestObject,
) (gen.GetNodeContainerDockerResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeContainerDocker400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validation.Struct(request.Params); !ok {
		return gen.GetNodeContainerDocker400JSONResponse{Error: &errMsg}, nil
	}

	data := &job.DockerListData{}
	if request.Params.State != nil {
		data.State = string(*request.Params.State)
	}
	if request.Params.Limit != nil {
		data.Limit = *request.Params.Limit
	}

	hostname := request.Hostname

	s.logger.Debug("container list",
		slog.String("target", hostname),
		slog.String("state", data.State),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.getNodeContainerDockerListBroadcast(ctx, hostname, data)
	}

	resp, err := s.JobClient.QueryDockerList(ctx, hostname, data)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeContainerDocker500JSONResponse{Error: &errMsg}, nil
	}

	summaries := dockerSummariesFromResponse(resp)
	jobUUID := uuid.MustParse(resp.JobID)
	changed := resp.Changed

	return gen.GetNodeContainerDocker200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.DockerListItem{
			{
				Hostname:   resp.Hostname,
				Containers: &summaries,
				Changed:    changed,
			},
		},
	}, nil
}

// dockerSummariesFromResponse extracts DockerSummary slice from a job response.
func dockerSummariesFromResponse(
	resp *job.Response,
) []gen.DockerSummary {
	var containers []struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Image   string `json:"image"`
		State   string `json:"state"`
		Created string `json:"created"`
	}
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &containers)
	}

	var summaries []gen.DockerSummary
	for _, c := range containers {
		id := c.ID
		name := c.Name
		image := c.Image
		state := c.State
		created := c.Created
		summaries = append(summaries, gen.DockerSummary{
			Id:      &id,
			Name:    &name,
			Image:   &image,
			State:   &state,
			Created: stringPtrOrNil(created),
		})
	}

	return summaries
}

// getNodeContainerDockerListBroadcast handles broadcast targets for container list.
func (s *Container) getNodeContainerDockerListBroadcast(
	ctx context.Context,
	target string,
	data *job.DockerListData,
) (gen.GetNodeContainerDockerResponseObject, error) {
	jobID, results, errs, err := s.JobClient.QueryDockerListBroadcast(ctx, target, data)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeContainerDocker500JSONResponse{Error: &errMsg}, nil
	}

	var responses []gen.DockerListItem
	for _, resp := range results {
		summaries := dockerSummariesFromResponse(resp)
		responses = append(responses, gen.DockerListItem{
			Hostname:   resp.Hostname,
			Containers: &summaries,
			Changed:    resp.Changed,
		})
	}
	for hostname, errMsg := range errs {
		e := errMsg
		responses = append(responses, gen.DockerListItem{
			Hostname: hostname,
			Error:    &e,
		})
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.GetNodeContainerDocker200JSONResponse{
		JobId:   &jobUUID,
		Results: responses,
	}, nil
}

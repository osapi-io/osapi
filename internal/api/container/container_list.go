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

	"github.com/retr0h/osapi/internal/api/container/gen"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/validation"
)

// GetNodeContainer lists containers on a target node.
func (s *Container) GetNodeContainer(
	ctx context.Context,
	request gen.GetNodeContainerRequestObject,
) (gen.GetNodeContainerResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeContainer400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validation.Struct(request.Params); !ok {
		return gen.GetNodeContainer400JSONResponse{Error: &errMsg}, nil
	}

	data := &job.ContainerListData{}
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
	)

	resp, err := s.JobClient.QueryContainerList(ctx, hostname, data)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeContainer500JSONResponse{Error: &errMsg}, nil
	}

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

	var summaries []gen.ContainerSummary
	for _, c := range containers {
		id := c.ID
		name := c.Name
		image := c.Image
		state := c.State
		created := c.Created
		summaries = append(summaries, gen.ContainerSummary{
			Id:      &id,
			Name:    &name,
			Image:   &image,
			State:   &state,
			Created: stringPtrOrNil(created),
		})
	}

	jobUUID := uuid.MustParse(resp.JobID)
	changed := resp.Changed

	return gen.GetNodeContainer200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.ContainerListItem{
			{
				Hostname:   resp.Hostname,
				Containers: &summaries,
				Changed:    changed,
			},
		},
	}, nil
}

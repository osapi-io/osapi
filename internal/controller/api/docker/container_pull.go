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

// PostNodeContainerDockerPull pulls a container image on a target node.
func (s *Container) PostNodeContainerDockerPull(
	ctx context.Context,
	request gen.PostNodeContainerDockerPullRequestObject,
) (gen.PostNodeContainerDockerPullResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.PostNodeContainerDockerPull400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validation.Struct(request.Body); !ok {
		return gen.PostNodeContainerDockerPull400JSONResponse{Error: &errMsg}, nil
	}

	data := &job.DockerPullData{
		Image: request.Body.Image,
	}

	hostname := request.Hostname

	s.logger.Debug("container pull",
		slog.String("target", hostname),
		slog.String("image", data.Image),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.postNodeContainerDockerPullBroadcast(ctx, hostname, data)
	}

	jobID, resp, err := s.JobClient.Modify(ctx, hostname, "docker", job.OperationDockerPull, data)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeContainerDockerPull500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		jobUUID := uuid.MustParse(jobID)
		e := resp.Error
		return gen.PostNodeContainerDockerPull202JSONResponse{
			JobId: &jobUUID,
			Results: []gen.DockerPullResultItem{
				{
					Hostname: resp.Hostname,
					Status:   gen.DockerPullResultItemStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	item := dockerPullItemFromResponse(resp)
	jobUUID := uuid.MustParse(jobID)

	return gen.PostNodeContainerDockerPull202JSONResponse{
		JobId:   &jobUUID,
		Results: []gen.DockerPullResultItem{item},
	}, nil
}

// dockerPullItemFromResponse builds a DockerPullResultItem from a job response.
func dockerPullItemFromResponse(
	resp *job.Response,
) gen.DockerPullResultItem {
	var pullResult struct {
		ImageID string `json:"image_id"`
		Tag     string `json:"tag"`
		Size    int64  `json:"size"`
	}
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &pullResult)
	}

	return gen.DockerPullResultItem{
		Hostname: resp.Hostname,
		Status:   gen.DockerPullResultItemStatusOk,
		ImageId:  stringPtrOrNil(pullResult.ImageID),
		Tag:      stringPtrOrNil(pullResult.Tag),
		Size:     int64PtrOrNil(pullResult.Size),
		Changed:  resp.Changed,
	}
}

// postNodeContainerDockerPullBroadcast handles broadcast targets for container pull.
func (s *Container) postNodeContainerDockerPullBroadcast(
	ctx context.Context,
	target string,
	data *job.DockerPullData,
) (gen.PostNodeContainerDockerPullResponseObject, error) {
	jobID, responses, err := s.JobClient.ModifyBroadcast(
		ctx,
		target,
		"docker",
		job.OperationDockerPull,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeContainerDockerPull500JSONResponse{Error: &errMsg}, nil
	}

	var items []gen.DockerPullResultItem
	for host, resp := range responses {
		item := gen.DockerPullResultItem{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.DockerPullResultItemStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.DockerPullResultItemStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.DockerPullResultItemStatusOk
			ok := dockerPullItemFromResponse(resp)
			item.ImageId = ok.ImageId
			item.Tag = ok.Tag
			item.Size = ok.Size
			item.Changed = resp.Changed
		}
		items = append(items, item)
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.PostNodeContainerDockerPull202JSONResponse{
		JobId:   &jobUUID,
		Results: items,
	}, nil
}

// int64PtrOrNil returns a pointer to v if non-zero, otherwise nil.
func int64PtrOrNil(
	v int64,
) *int64 {
	if v == 0 {
		return nil
	}
	return &v
}

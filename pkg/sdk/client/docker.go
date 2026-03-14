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

package client

import (
	"context"
	"fmt"

	"github.com/retr0h/osapi/pkg/sdk/client/gen"
)

// DockerService provides Docker container management operations.
type DockerService struct {
	client *gen.ClientWithResponses
}

// Create creates a new container on the target host.
func (s *DockerService) Create(
	ctx context.Context,
	hostname string,
	opts DockerCreateOpts,
) (*Response[Collection[DockerResult]], error) {
	body := gen.DockerCreateRequest{
		Image: opts.Image,
	}
	if opts.Name != "" {
		body.Name = &opts.Name
	}
	if len(opts.Command) > 0 {
		body.Command = &opts.Command
	}
	if len(opts.Env) > 0 {
		body.Env = &opts.Env
	}
	if len(opts.Ports) > 0 {
		body.Ports = &opts.Ports
	}
	if len(opts.Volumes) > 0 {
		body.Volumes = &opts.Volumes
	}
	if opts.AutoStart != nil {
		body.AutoStart = opts.AutoStart
	}

	resp, err := s.client.PostNodeContainerDockerWithResponse(ctx, hostname, body)
	if err != nil {
		return nil, fmt.Errorf("docker create: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON400,
		resp.JSON401,
		resp.JSON403,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON202 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(dockerResultCollectionFromGen(resp.JSON202), resp.Body), nil
}

// List lists containers on the target host, optionally filtered by state.
func (s *DockerService) List(
	ctx context.Context,
	hostname string,
	params *DockerListParams,
) (*Response[Collection[DockerListResult]], error) {
	var genParams *gen.GetNodeContainerDockerParams
	if params != nil {
		genParams = &gen.GetNodeContainerDockerParams{}
		if params.State != "" {
			state := gen.GetNodeContainerDockerParamsState(params.State)
			genParams.State = &state
		}
		if params.Limit > 0 {
			genParams.Limit = &params.Limit
		}
	}

	resp, err := s.client.GetNodeContainerDockerWithResponse(ctx, hostname, genParams)
	if err != nil {
		return nil, fmt.Errorf("docker list: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON400,
		resp.JSON401,
		resp.JSON403,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(dockerListCollectionFromGen(resp.JSON200), resp.Body), nil
}

// Inspect retrieves detailed information about a specific container.
func (s *DockerService) Inspect(
	ctx context.Context,
	hostname string,
	id string,
) (*Response[Collection[DockerDetailResult]], error) {
	resp, err := s.client.GetNodeContainerDockerByIDWithResponse(ctx, hostname, id)
	if err != nil {
		return nil, fmt.Errorf("docker inspect: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON400,
		resp.JSON401,
		resp.JSON403,
		resp.JSON404,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(dockerDetailCollectionFromGen(resp.JSON200), resp.Body), nil
}

// Start starts a stopped container on the target host.
func (s *DockerService) Start(
	ctx context.Context,
	hostname string,
	id string,
) (*Response[Collection[DockerActionResult]], error) {
	resp, err := s.client.PostNodeContainerDockerStartWithResponse(ctx, hostname, id)
	if err != nil {
		return nil, fmt.Errorf("docker start: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON400,
		resp.JSON401,
		resp.JSON403,
		resp.JSON404,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON202 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(dockerActionCollectionFromGen(resp.JSON202), resp.Body), nil
}

// Stop stops a running container on the target host.
func (s *DockerService) Stop(
	ctx context.Context,
	hostname string,
	id string,
	opts DockerStopOpts,
) (*Response[Collection[DockerActionResult]], error) {
	body := gen.DockerStopRequest{}
	if opts.Timeout > 0 {
		body.Timeout = &opts.Timeout
	}

	resp, err := s.client.PostNodeContainerDockerStopWithResponse(ctx, hostname, id, body)
	if err != nil {
		return nil, fmt.Errorf("docker stop: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON400,
		resp.JSON401,
		resp.JSON403,
		resp.JSON404,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON202 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(dockerActionCollectionFromGen(resp.JSON202), resp.Body), nil
}

// Remove removes a container from the target host.
func (s *DockerService) Remove(
	ctx context.Context,
	hostname string,
	id string,
	params *DockerRemoveParams,
) (*Response[Collection[DockerActionResult]], error) {
	var genParams *gen.DeleteNodeContainerDockerByIDParams
	if params != nil {
		genParams = &gen.DeleteNodeContainerDockerByIDParams{
			Force: &params.Force,
		}
	}

	resp, err := s.client.DeleteNodeContainerDockerByIDWithResponse(ctx, hostname, id, genParams)
	if err != nil {
		return nil, fmt.Errorf("docker remove: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON400,
		resp.JSON401,
		resp.JSON403,
		resp.JSON404,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON202 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(dockerActionCollectionFromGen(resp.JSON202), resp.Body), nil
}

// Exec executes a command inside a running container on the target host.
func (s *DockerService) Exec(
	ctx context.Context,
	hostname string,
	id string,
	opts DockerExecOpts,
) (*Response[Collection[DockerExecResult]], error) {
	body := gen.DockerExecRequest{
		Command: opts.Command,
	}
	if len(opts.Env) > 0 {
		body.Env = &opts.Env
	}
	if opts.WorkingDir != "" {
		body.WorkingDir = &opts.WorkingDir
	}

	resp, err := s.client.PostNodeContainerDockerExecWithResponse(ctx, hostname, id, body)
	if err != nil {
		return nil, fmt.Errorf("docker exec: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON400,
		resp.JSON401,
		resp.JSON403,
		resp.JSON404,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON202 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(dockerExecCollectionFromGen(resp.JSON202), resp.Body), nil
}

// ImageRemove removes a container image from the target host.
func (s *DockerService) ImageRemove(
	ctx context.Context,
	hostname string,
	imageName string,
	params *DockerImageRemoveParams,
) (*Response[Collection[DockerActionResult]], error) {
	var genParams *gen.DeleteNodeContainerDockerImageParams
	if params != nil {
		genParams = &gen.DeleteNodeContainerDockerImageParams{
			Force: &params.Force,
		}
	}

	resp, err := s.client.DeleteNodeContainerDockerImageWithResponse(
		ctx,
		hostname,
		imageName,
		genParams,
	)
	if err != nil {
		return nil, fmt.Errorf("docker image remove: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON400,
		resp.JSON401,
		resp.JSON403,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON202 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(dockerActionCollectionFromGen(resp.JSON202), resp.Body), nil
}

// Pull pulls a container image on the target host.
func (s *DockerService) Pull(
	ctx context.Context,
	hostname string,
	opts DockerPullOpts,
) (*Response[Collection[DockerPullResult]], error) {
	body := gen.DockerPullRequest{
		Image: opts.Image,
	}

	resp, err := s.client.PostNodeContainerDockerPullWithResponse(ctx, hostname, body)
	if err != nil {
		return nil, fmt.Errorf("docker pull: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON400,
		resp.JSON401,
		resp.JSON403,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON202 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(dockerPullCollectionFromGen(resp.JSON202), resp.Body), nil
}

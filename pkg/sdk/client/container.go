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

// ContainerService provides container management operations.
type ContainerService struct {
	client *gen.ClientWithResponses
}

// Create creates a new container on the target host.
func (s *ContainerService) Create(
	ctx context.Context,
	hostname string,
	body gen.ContainerCreateRequest,
) (*Response[Collection[ContainerResult]], error) {
	resp, err := s.client.PostNodeContainerWithResponse(ctx, hostname, body)
	if err != nil {
		return nil, fmt.Errorf("create container: %w", err)
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

	return NewResponse(containerResultCollectionFromGen(resp.JSON202), resp.Body), nil
}

// List lists containers on the target host, optionally filtered by state.
func (s *ContainerService) List(
	ctx context.Context,
	hostname string,
	params *gen.GetNodeContainerParams,
) (*Response[Collection[ContainerListResult]], error) {
	resp, err := s.client.GetNodeContainerWithResponse(ctx, hostname, params)
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
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

	return NewResponse(containerListCollectionFromGen(resp.JSON200), resp.Body), nil
}

// Inspect retrieves detailed information about a specific container.
func (s *ContainerService) Inspect(
	ctx context.Context,
	hostname string,
	id string,
) (*Response[Collection[ContainerDetailResult]], error) {
	resp, err := s.client.GetNodeContainerByIDWithResponse(ctx, hostname, id)
	if err != nil {
		return nil, fmt.Errorf("inspect container: %w", err)
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

	return NewResponse(containerDetailCollectionFromGen(resp.JSON200), resp.Body), nil
}

// Start starts a stopped container on the target host.
func (s *ContainerService) Start(
	ctx context.Context,
	hostname string,
	id string,
) (*Response[Collection[ContainerActionResult]], error) {
	resp, err := s.client.PostNodeContainerStartWithResponse(ctx, hostname, id)
	if err != nil {
		return nil, fmt.Errorf("start container: %w", err)
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

	return NewResponse(containerActionCollectionFromGen(resp.JSON202), resp.Body), nil
}

// Stop stops a running container on the target host.
func (s *ContainerService) Stop(
	ctx context.Context,
	hostname string,
	id string,
	body gen.ContainerStopRequest,
) (*Response[Collection[ContainerActionResult]], error) {
	resp, err := s.client.PostNodeContainerStopWithResponse(ctx, hostname, id, body)
	if err != nil {
		return nil, fmt.Errorf("stop container: %w", err)
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

	return NewResponse(containerActionCollectionFromGen(resp.JSON202), resp.Body), nil
}

// Remove removes a container from the target host.
func (s *ContainerService) Remove(
	ctx context.Context,
	hostname string,
	id string,
	params *gen.DeleteNodeContainerByIDParams,
) (*Response[Collection[ContainerActionResult]], error) {
	resp, err := s.client.DeleteNodeContainerByIDWithResponse(ctx, hostname, id, params)
	if err != nil {
		return nil, fmt.Errorf("remove container: %w", err)
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

	return NewResponse(containerActionCollectionFromGen(resp.JSON202), resp.Body), nil
}

// Exec executes a command inside a running container on the target host.
func (s *ContainerService) Exec(
	ctx context.Context,
	hostname string,
	id string,
	body gen.ContainerExecRequest,
) (*Response[Collection[ContainerExecResult]], error) {
	resp, err := s.client.PostNodeContainerExecWithResponse(ctx, hostname, id, body)
	if err != nil {
		return nil, fmt.Errorf("exec in container: %w", err)
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

	return NewResponse(containerExecCollectionFromGen(resp.JSON202), resp.Body), nil
}

// Pull pulls a container image on the target host.
func (s *ContainerService) Pull(
	ctx context.Context,
	hostname string,
	body gen.ContainerPullRequest,
) (*Response[Collection[ContainerPullResult]], error) {
	resp, err := s.client.PostNodeContainerPullWithResponse(ctx, hostname, body)
	if err != nil {
		return nil, fmt.Errorf("pull image: %w", err)
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

	return NewResponse(containerPullCollectionFromGen(resp.JSON202), resp.Body), nil
}

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

	"github.com/osapi-io/osapi/pkg/sdk/client/gen"
)

// DockerService manages the Docker containers and images of a host.
type DockerService struct {
	client *gen.ClientWithResponses
}

// checked is the error of a response: its status, or a body that is nil.
func checked(
	code int,
	noBody bool,
	errs ...*gen.ErrorResponse,
) error {
	if err := checkError(code, errs...); err != nil {
		return err
	}
	if noBody {
		return &UnexpectedStatusError{APIError{StatusCode: code, Message: "nil response body"}}
	}

	return nil
}

// ptr is a pointer to v, or nil for the zero value, so that an option
// that is not set is not sent.
func ptr[T comparable](
	v T,
) *T {
	var zero T
	if v == zero {
		return nil
	}

	return &v
}

// slicePtr is a pointer to s, or nil when it is empty.
func slicePtr[T any](
	s []T,
) *[]T {
	if len(s) == 0 {
		return nil
	}

	return &s
}

// mapPtr is a pointer to m, or nil when it is empty.
func mapPtr(
	m map[string]string,
) *map[string]string {
	if len(m) == 0 {
		return nil
	}

	return &m
}

// List lists the containers of the target. state is running, stopped,
// all, or empty for all.
func (s *DockerService) List(
	ctx context.Context,
	hostname string,
	state string,
) (*Response[Collection[DockerResult]], error) {
	params := &gen.GetNodeContainerDockerParams{}
	if state != "" {
		st := gen.GetNodeContainerDockerParamsState(state)
		params.State = &st
	}
	resp, err := s.client.GetNodeContainerDockerWithResponse(ctx, hostname, params)
	if err != nil {
		return nil, fmt.Errorf("docker list: %w", err)
	}
	if err := checked(resp.StatusCode(), resp.JSON200 == nil, resp.JSON400, resp.JSON401,
		resp.JSON403, resp.JSON404, resp.JSON409, resp.JSON500); err != nil {
		return nil, err
	}

	return newResponse(collect(resp.JSON200.Results, resp.JSON200.JobId, dockerFromGen),
		resp.Body, resp.HTTPResponse), nil
}

// Get reads one container by its name, its ID or the start of its ID.
func (s *DockerService) Get(
	ctx context.Context,
	hostname string,
	name string,
) (*Response[Collection[DockerResult]], error) {
	resp, err := s.client.GetNodeContainerDockerByKeyWithResponse(ctx, hostname, name)
	if err != nil {
		return nil, fmt.Errorf("docker get: %w", err)
	}
	if err := checked(resp.StatusCode(), resp.JSON200 == nil, resp.JSON400, resp.JSON401,
		resp.JSON403, resp.JSON404, resp.JSON409, resp.JSON500); err != nil {
		return nil, err
	}

	return newResponse(collect(resp.JSON200.Results, resp.JSON200.JobId, dockerFromGen),
		resp.Body, resp.HTTPResponse), nil
}

// Create creates a container. A container of that name that the host has
// already is not made again, and reports changed: false.
func (s *DockerService) Create(
	ctx context.Context,
	hostname string,
	opts DockerCreateOpts,
) (*Response[Collection[DockerActionResult]], error) {
	body := gen.ContainerDockerCreateRequest{
		Name:              opts.Name,
		Image:             opts.Image,
		ContainerHostname: ptr(opts.Hostname),
		Dns:               slicePtr(opts.DNS),
		Command:           slicePtr(opts.Command),
		Env:               mapPtr(opts.Env),
		Ports:             slicePtr(opts.Ports),
		Volumes:           slicePtr(opts.Volumes),
		AutoStart:         ptr(opts.AutoStart),
	}
	resp, err := s.client.PostNodeContainerDockerWithResponse(ctx, hostname, body)
	if err != nil {
		return nil, fmt.Errorf("docker create: %w", err)
	}
	if err := checked(resp.StatusCode(), resp.JSON200 == nil, resp.JSON400, resp.JSON401,
		resp.JSON403, resp.JSON404, resp.JSON409, resp.JSON500); err != nil {
		return nil, err
	}

	return newResponse(collect(resp.JSON200.Results, resp.JSON200.JobId, dockerActionFromGen),
		resp.Body, resp.HTTPResponse), nil
}

// Remove removes a container, also one that runs.
func (s *DockerService) Remove(
	ctx context.Context,
	hostname string,
	name string,
) (*Response[Collection[DockerActionResult]], error) {
	resp, err := s.client.DeleteNodeContainerDockerWithResponse(ctx, hostname, name)
	if err != nil {
		return nil, fmt.Errorf("docker remove: %w", err)
	}
	if err := checked(resp.StatusCode(), resp.JSON200 == nil, resp.JSON400, resp.JSON401,
		resp.JSON403, resp.JSON404, resp.JSON409, resp.JSON500); err != nil {
		return nil, err
	}

	return newResponse(collect(resp.JSON200.Results, resp.JSON200.JobId, dockerActionFromGen),
		resp.Body, resp.HTTPResponse), nil
}

// Start starts a container.
func (s *DockerService) Start(
	ctx context.Context,
	hostname string,
	name string,
) (*Response[Collection[DockerActionResult]], error) {
	resp, err := s.client.PostNodeContainerDockerStartWithResponse(ctx, hostname, name)
	if err != nil {
		return nil, fmt.Errorf("docker start: %w", err)
	}
	if err := checked(resp.StatusCode(), resp.JSON200 == nil, resp.JSON400, resp.JSON401,
		resp.JSON403, resp.JSON404, resp.JSON409, resp.JSON500); err != nil {
		return nil, err
	}

	return newResponse(collect(resp.JSON200.Results, resp.JSON200.JobId, dockerActionFromGen),
		resp.Body, resp.HTTPResponse), nil
}

// Stop stops a container, waiting timeout seconds before docker kills it;
// 0 is docker's default.
func (s *DockerService) Stop(
	ctx context.Context,
	hostname string,
	name string,
	timeout int,
) (*Response[Collection[DockerActionResult]], error) {
	resp, err := s.client.PostNodeContainerDockerStopWithResponse(ctx, hostname, name,
		gen.ContainerDockerStopRequest{Timeout: ptr(timeout)})
	if err != nil {
		return nil, fmt.Errorf("docker stop: %w", err)
	}
	if err := checked(resp.StatusCode(), resp.JSON200 == nil, resp.JSON400, resp.JSON401,
		resp.JSON403, resp.JSON404, resp.JSON409, resp.JSON500); err != nil {
		return nil, err
	}

	return newResponse(collect(resp.JSON200.Results, resp.JSON200.JobId, dockerActionFromGen),
		resp.Body, resp.HTTPResponse), nil
}

// Exec runs a command in a container, and answers what it wrote and its
// exit code.
func (s *DockerService) Exec(
	ctx context.Context,
	hostname string,
	name string,
	opts DockerExecOpts,
) (*Response[Collection[DockerExecResult]], error) {
	body := gen.ContainerDockerExecRequest{
		Command:    opts.Command,
		Env:        mapPtr(opts.Env),
		WorkingDir: ptr(opts.WorkingDir),
	}
	resp, err := s.client.PostNodeContainerDockerExecWithResponse(ctx, hostname, name, body)
	if err != nil {
		return nil, fmt.Errorf("docker exec: %w", err)
	}
	if err := checked(resp.StatusCode(), resp.JSON200 == nil, resp.JSON400, resp.JSON401,
		resp.JSON403, resp.JSON404, resp.JSON409, resp.JSON500); err != nil {
		return nil, err
	}

	return newResponse(collect(resp.JSON200.Results, resp.JSON200.JobId, dockerExecFromGen),
		resp.Body, resp.HTTPResponse), nil
}

// ListImages lists the images of the target.
func (s *DockerService) ListImages(
	ctx context.Context,
	hostname string,
) (*Response[Collection[DockerImageResult]], error) {
	resp, err := s.client.GetNodeContainerImageWithResponse(ctx, hostname)
	if err != nil {
		return nil, fmt.Errorf("docker image list: %w", err)
	}
	if err := checked(resp.StatusCode(), resp.JSON200 == nil, resp.JSON400, resp.JSON401,
		resp.JSON403, resp.JSON404, resp.JSON409, resp.JSON500); err != nil {
		return nil, err
	}

	return newResponse(collect(resp.JSON200.Results, resp.JSON200.JobId, dockerImageFromGen),
		resp.Body, resp.HTTPResponse), nil
}

// Pull pulls an image, and answers the ID and the size of the image that
// the host has after it.
func (s *DockerService) Pull(
	ctx context.Context,
	hostname string,
	image string,
) (*Response[Collection[DockerPullResult]], error) {
	resp, err := s.client.PostNodeContainerImagePullWithResponse(ctx, hostname,
		gen.ContainerImagePullRequest{Image: image})
	if err != nil {
		return nil, fmt.Errorf("docker pull: %w", err)
	}
	if err := checked(resp.StatusCode(), resp.JSON200 == nil, resp.JSON400, resp.JSON401,
		resp.JSON403, resp.JSON404, resp.JSON409, resp.JSON500); err != nil {
		return nil, err
	}

	return newResponse(collect(resp.JSON200.Results, resp.JSON200.JobId, dockerPullFromGen),
		resp.Body, resp.HTTPResponse), nil
}

// RemoveImage removes an image; force removes one that a container uses.
func (s *DockerService) RemoveImage(
	ctx context.Context,
	hostname string,
	image string,
	force bool,
) (*Response[Collection[DockerImageRemoveResult]], error) {
	resp, err := s.client.DeleteNodeContainerImageWithResponse(ctx, hostname, image,
		gen.ContainerImageDeleteRequest{Force: ptr(force)})
	if err != nil {
		return nil, fmt.Errorf("docker image remove: %w", err)
	}
	if err := checked(resp.StatusCode(), resp.JSON200 == nil, resp.JSON400, resp.JSON401,
		resp.JSON403, resp.JSON404, resp.JSON409, resp.JSON500); err != nil {
		return nil, err
	}

	return newResponse(collect(resp.JSON200.Results, resp.JSON200.JobId, dockerImageRemoveFromGen),
		resp.Body, resp.HTTPResponse), nil
}

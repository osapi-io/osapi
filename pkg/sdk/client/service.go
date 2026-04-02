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

// ServiceService provides systemd service management operations.
type ServiceService struct {
	client *gen.ClientWithResponses
}

// List returns all services on the target host.
func (s *ServiceService) List(
	ctx context.Context,
	hostname string,
) (*Response[Collection[ServiceInfoResult]], error) {
	resp, err := s.client.GetNodeServiceWithResponse(ctx, hostname)
	if err != nil {
		return nil, fmt.Errorf("service list: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
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

	return NewResponse(serviceListCollectionFromGen(resp.JSON200), resp.Body), nil
}

// Get returns information about a specific service on the target host.
func (s *ServiceService) Get(
	ctx context.Context,
	hostname string,
	name string,
) (*Response[Collection[ServiceGetResult]], error) {
	resp, err := s.client.GetNodeServiceByNameWithResponse(ctx, hostname, name)
	if err != nil {
		return nil, fmt.Errorf("service get: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
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

	return NewResponse(serviceGetCollectionFromGen(resp.JSON200), resp.Body), nil
}

// Create creates a new service unit file on the target host.
func (s *ServiceService) Create(
	ctx context.Context,
	hostname string,
	opts ServiceCreateOpts,
) (*Response[Collection[ServiceMutationResult]], error) {
	body := gen.ServiceCreateRequest{
		Name:   opts.Name,
		Object: opts.Object,
	}

	resp, err := s.client.PostNodeServiceWithResponse(ctx, hostname, body)
	if err != nil {
		return nil, fmt.Errorf("service create: %w", err)
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

	return NewResponse(serviceMutationCollectionFromGen(resp.JSON200), resp.Body), nil
}

// Update updates an existing service unit file on the target host.
func (s *ServiceService) Update(
	ctx context.Context,
	hostname string,
	name string,
	opts ServiceUpdateOpts,
) (*Response[Collection[ServiceMutationResult]], error) {
	body := gen.ServiceUpdateRequest{
		Object: opts.Object,
	}

	resp, err := s.client.PutNodeServiceWithResponse(ctx, hostname, name, body)
	if err != nil {
		return nil, fmt.Errorf("service update: %w", err)
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

	return NewResponse(serviceMutationCollectionFromGen(resp.JSON200), resp.Body), nil
}

// Delete deletes a service unit file on the target host.
func (s *ServiceService) Delete(
	ctx context.Context,
	hostname string,
	name string,
) (*Response[Collection[ServiceMutationResult]], error) {
	resp, err := s.client.DeleteNodeServiceWithResponse(ctx, hostname, name)
	if err != nil {
		return nil, fmt.Errorf("service delete: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
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

	return NewResponse(serviceMutationCollectionFromGen(resp.JSON200), resp.Body), nil
}

// Start starts a service on the target host.
func (s *ServiceService) Start(
	ctx context.Context,
	hostname string,
	name string,
) (*Response[Collection[ServiceMutationResult]], error) {
	resp, err := s.client.PostNodeServiceStartWithResponse(ctx, hostname, name)
	if err != nil {
		return nil, fmt.Errorf("service start: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
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

	return NewResponse(serviceMutationCollectionFromGen(resp.JSON200), resp.Body), nil
}

// Stop stops a service on the target host.
func (s *ServiceService) Stop(
	ctx context.Context,
	hostname string,
	name string,
) (*Response[Collection[ServiceMutationResult]], error) {
	resp, err := s.client.PostNodeServiceStopWithResponse(ctx, hostname, name)
	if err != nil {
		return nil, fmt.Errorf("service stop: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
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

	return NewResponse(serviceMutationCollectionFromGen(resp.JSON200), resp.Body), nil
}

// Restart restarts a service on the target host.
func (s *ServiceService) Restart(
	ctx context.Context,
	hostname string,
	name string,
) (*Response[Collection[ServiceMutationResult]], error) {
	resp, err := s.client.PostNodeServiceRestartWithResponse(ctx, hostname, name)
	if err != nil {
		return nil, fmt.Errorf("service restart: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
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

	return NewResponse(serviceMutationCollectionFromGen(resp.JSON200), resp.Body), nil
}

// Enable enables a service to start on boot on the target host.
func (s *ServiceService) Enable(
	ctx context.Context,
	hostname string,
	name string,
) (*Response[Collection[ServiceMutationResult]], error) {
	resp, err := s.client.PostNodeServiceEnableWithResponse(ctx, hostname, name)
	if err != nil {
		return nil, fmt.Errorf("service enable: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
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

	return NewResponse(serviceMutationCollectionFromGen(resp.JSON200), resp.Body), nil
}

// Disable disables a service from starting on boot on the target host.
func (s *ServiceService) Disable(
	ctx context.Context,
	hostname string,
	name string,
) (*Response[Collection[ServiceMutationResult]], error) {
	resp, err := s.client.PostNodeServiceDisableWithResponse(ctx, hostname, name)
	if err != nil {
		return nil, fmt.Errorf("service disable: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
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

	return NewResponse(serviceMutationCollectionFromGen(resp.JSON200), resp.Body), nil
}

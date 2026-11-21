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

// InterfaceService reads the network interfaces of a host and writes the
// config osapi has of one.
type InterfaceService struct {
	client *gen.ClientWithResponses
}

// List is the network interfaces of each target host, one result per host.
func (s *InterfaceService) List(
	ctx context.Context,
	target string,
) (*Response[Collection[InterfaceListResult]], error) {
	resp, err := s.client.GetNodeNetworkInterfaceWithResponse(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("interface list: %w", err)
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
		return nil, &UnexpectedStatusError{
			APIError{StatusCode: resp.StatusCode(), Message: "nil response body"},
		}
	}

	return newResponse(interfaceListFromGen(resp.JSON200), resp.Body, resp.HTTPResponse), nil
}

// Get is one network interface of each target host.
func (s *InterfaceService) Get(
	ctx context.Context,
	target string,
	name string,
) (*Response[Collection[InterfaceGetResult]], error) {
	resp, err := s.client.GetNodeNetworkInterfaceByKeyWithResponse(ctx, target, name)
	if err != nil {
		return nil, fmt.Errorf("interface get: %w", err)
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
		return nil, &UnexpectedStatusError{
			APIError{StatusCode: resp.StatusCode(), Message: "nil response body"},
		}
	}

	return newResponse(interfaceGetFromGen(resp.JSON200), resp.Body, resp.HTTPResponse), nil
}

// Create writes osapi's config of an interface on the target. An interface
// osapi already configures does not change.
func (s *InterfaceService) Create(
	ctx context.Context,
	target string,
	name string,
	opts InterfaceConfigOpts,
) (*Response[Collection[InterfaceMutationResult]], error) {
	resp, err := s.client.PostNodeNetworkInterfaceWithResponse(
		ctx, target, name, gen.NetworkInterfaceCreateRequest(interfaceRequest(opts)))
	if err != nil {
		return nil, fmt.Errorf("interface create: %w", err)
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
		return nil, &UnexpectedStatusError{
			APIError{StatusCode: resp.StatusCode(), Message: "nil response body"},
		}
	}

	return newResponse(interfaceMutationsFromGen(resp.JSON200.Results, resp.JSON200.JobId),
		resp.Body, resp.HTTPResponse), nil
}

// Update writes osapi's config of an interface over the one it has on the
// target.
func (s *InterfaceService) Update(
	ctx context.Context,
	target string,
	name string,
	opts InterfaceConfigOpts,
) (*Response[Collection[InterfaceMutationResult]], error) {
	resp, err := s.client.PutNodeNetworkInterfaceWithResponse(
		ctx,
		target,
		name,
		interfaceRequest(opts),
	)
	if err != nil {
		return nil, fmt.Errorf("interface update: %w", err)
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
		return nil, &UnexpectedStatusError{
			APIError{StatusCode: resp.StatusCode(), Message: "nil response body"},
		}
	}

	return newResponse(interfaceMutationsFromGen(resp.JSON200.Results, resp.JSON200.JobId),
		resp.Body, resp.HTTPResponse), nil
}

// Delete drops osapi's config of an interface on the target.
func (s *InterfaceService) Delete(
	ctx context.Context,
	target string,
	name string,
) (*Response[Collection[InterfaceMutationResult]], error) {
	resp, err := s.client.DeleteNodeNetworkInterfaceWithResponse(ctx, target, name)
	if err != nil {
		return nil, fmt.Errorf("interface delete: %w", err)
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
		return nil, &UnexpectedStatusError{
			APIError{StatusCode: resp.StatusCode(), Message: "nil response body"},
		}
	}

	return newResponse(interfaceMutationsFromGen(resp.JSON200.Results, resp.JSON200.JobId),
		resp.Body, resp.HTTPResponse), nil
}

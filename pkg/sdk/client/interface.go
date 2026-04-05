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

// InterfaceService provides network interface management operations.
type InterfaceService struct {
	client *gen.ClientWithResponses
}

// List lists all network interfaces on the target host.
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
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(interfaceListCollectionFromGen(resp.JSON200), resp.Body), nil
}

// Get retrieves a single network interface by name on the target host.
func (s *InterfaceService) Get(
	ctx context.Context,
	target string,
	name string,
) (*Response[Collection[InterfaceGetResult]], error) {
	resp, err := s.client.GetNodeNetworkInterfaceByNameWithResponse(ctx, target, name)
	if err != nil {
		return nil, fmt.Errorf("interface get: %w", err)
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

	return NewResponse(interfaceGetCollectionFromGen(resp.JSON200), resp.Body), nil
}

// Create creates a network interface configuration on the target host.
func (s *InterfaceService) Create(
	ctx context.Context,
	target string,
	name string,
	opts InterfaceConfigOpts,
) (*Response[Collection[InterfaceMutationResult]], error) {
	body := interfaceConfigRequestFromOpts(opts)

	resp, err := s.client.PostNodeNetworkInterfaceWithResponse(ctx, target, name, body)
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
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(interfaceMutationCollectionFromCreate(resp.JSON200), resp.Body), nil
}

// Update updates a network interface configuration on the target host.
func (s *InterfaceService) Update(
	ctx context.Context,
	target string,
	name string,
	opts InterfaceConfigOpts,
) (*Response[Collection[InterfaceMutationResult]], error) {
	body := interfaceConfigRequestFromOpts(opts)

	resp, err := s.client.PutNodeNetworkInterfaceWithResponse(ctx, target, name, body)
	if err != nil {
		return nil, fmt.Errorf("interface update: %w", err)
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

	return NewResponse(interfaceMutationCollectionFromUpdate(resp.JSON200), resp.Body), nil
}

// Delete removes a network interface configuration on the target host.
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
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(interfaceMutationCollectionFromDelete(resp.JSON200), resp.Body), nil
}

// interfaceConfigRequestFromOpts builds a gen.InterfaceConfigRequest
// from SDK options.
func interfaceConfigRequestFromOpts(
	opts InterfaceConfigOpts,
) gen.InterfaceConfigRequest {
	body := gen.InterfaceConfigRequest{}

	if opts.DHCP4 != nil {
		body.Dhcp4 = opts.DHCP4
	}

	if opts.DHCP6 != nil {
		body.Dhcp6 = opts.DHCP6
	}

	if len(opts.Addresses) > 0 {
		body.Addresses = &opts.Addresses
	}

	if opts.Gateway4 != "" {
		body.Gateway4 = &opts.Gateway4
	}

	if opts.Gateway6 != "" {
		body.Gateway6 = &opts.Gateway6
	}

	if opts.MTU != nil {
		body.Mtu = opts.MTU
	}

	if opts.MACAddress != "" {
		body.MacAddress = &opts.MACAddress
	}

	if opts.WakeOnLAN != nil {
		body.Wakeonlan = opts.WakeOnLAN
	}

	return body
}

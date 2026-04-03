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

// RouteService provides network route management operations.
type RouteService struct {
	client *gen.ClientWithResponses
}

// List lists all network routes on the target host.
func (s *RouteService) List(
	ctx context.Context,
	target string,
) (*Response[Collection[RouteListResult]], error) {
	resp, err := s.client.GetNodeNetworkRouteWithResponse(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("route list: %w", err)
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

	return NewResponse(routeListCollectionFromGen(resp.JSON200), resp.Body), nil
}

// Get retrieves routes for a specific interface on the target host.
func (s *RouteService) Get(
	ctx context.Context,
	target string,
	interfaceName string,
) (*Response[Collection[RouteGetResult]], error) {
	resp, err := s.client.GetNodeNetworkRouteByInterfaceWithResponse(ctx, target, interfaceName)
	if err != nil {
		return nil, fmt.Errorf("route get: %w", err)
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

	return NewResponse(routeGetCollectionFromGen(resp.JSON200), resp.Body), nil
}

// Create creates route configuration for an interface on the target host.
func (s *RouteService) Create(
	ctx context.Context,
	target string,
	interfaceName string,
	opts RouteConfigOpts,
) (*Response[Collection[RouteMutationResult]], error) {
	body := routeConfigRequestFromOpts(opts)

	resp, err := s.client.PostNodeNetworkRouteWithResponse(ctx, target, interfaceName, body)
	if err != nil {
		return nil, fmt.Errorf("route create: %w", err)
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

	return NewResponse(routeMutationCollectionFromCreate(resp.JSON200), resp.Body), nil
}

// Update updates route configuration for an interface on the target host.
func (s *RouteService) Update(
	ctx context.Context,
	target string,
	interfaceName string,
	opts RouteConfigOpts,
) (*Response[Collection[RouteMutationResult]], error) {
	body := routeConfigRequestFromOpts(opts)

	resp, err := s.client.PutNodeNetworkRouteWithResponse(ctx, target, interfaceName, body)
	if err != nil {
		return nil, fmt.Errorf("route update: %w", err)
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

	return NewResponse(routeMutationCollectionFromUpdate(resp.JSON200), resp.Body), nil
}

// Delete removes route configuration for an interface on the target host.
func (s *RouteService) Delete(
	ctx context.Context,
	target string,
	interfaceName string,
) (*Response[Collection[RouteMutationResult]], error) {
	resp, err := s.client.DeleteNodeNetworkRouteWithResponse(ctx, target, interfaceName)
	if err != nil {
		return nil, fmt.Errorf("route delete: %w", err)
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

	return NewResponse(routeMutationCollectionFromDelete(resp.JSON200), resp.Body), nil
}

// routeConfigRequestFromOpts builds a gen.RouteConfigRequest from SDK options.
func routeConfigRequestFromOpts(
	opts RouteConfigOpts,
) gen.RouteConfigRequest {
	items := make([]gen.RouteItem, 0, len(opts.Routes))
	for _, r := range opts.Routes {
		item := gen.RouteItem{
			To:  r.To,
			Via: r.Via,
		}

		if r.Metric != nil {
			item.Metric = r.Metric
		}

		items = append(items, item)
	}

	return gen.RouteConfigRequest{
		Routes: items,
	}
}

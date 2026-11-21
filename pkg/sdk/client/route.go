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

// RouteService reads the routes of a host and writes the routes osapi has
// of an interface.
type RouteService struct {
	client *gen.ClientWithResponses
}

// List is the routes of each target host, one result per host.
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
		return nil, &UnexpectedStatusError{
			APIError{StatusCode: resp.StatusCode(), Message: "nil response body"},
		}
	}
	rows := routeRowsFromGen(resp.JSON200.Results)
	results := make([]RouteListResult, 0, len(rows))
	for _, r := range rows {
		results = append(results, RouteListResult(r))
	}

	return newResponse(Collection[RouteListResult]{
		Results: results, JobID: jobIDFromGen(resp.JSON200.JobId),
	}, resp.Body, resp.HTTPResponse), nil
}

// Get is the routes of one interface of each target host.
func (s *RouteService) Get(
	ctx context.Context,
	target string,
	interfaceName string,
) (*Response[Collection[RouteGetResult]], error) {
	resp, err := s.client.GetNodeNetworkRouteByKeyWithResponse(ctx, target, interfaceName)
	if err != nil {
		return nil, fmt.Errorf("route get: %w", err)
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

	return newResponse(Collection[RouteGetResult]{
		Results: routeRowsFromGen(resp.JSON200.Results), JobID: jobIDFromGen(resp.JSON200.JobId),
	}, resp.Body, resp.HTTPResponse), nil
}

// Create writes osapi's routes of an interface on the target. An interface
// osapi already has routes of does not change.
func (s *RouteService) Create(
	ctx context.Context,
	target string,
	interfaceName string,
	opts RouteConfigOpts,
) (*Response[Collection[RouteMutationResult]], error) {
	resp, err := s.client.PostNodeNetworkRouteWithResponse(ctx, target, interfaceName,
		gen.NetworkRouteCreateRequest{Routes: routeStrings(opts.Routes)})
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
		return nil, &UnexpectedStatusError{
			APIError{StatusCode: resp.StatusCode(), Message: "nil response body"},
		}
	}

	return newResponse(routeMutationsFromGen(resp.JSON200.Results, resp.JSON200.JobId),
		resp.Body, resp.HTTPResponse), nil
}

// Update writes osapi's routes of an interface over the ones it has on the
// target.
func (s *RouteService) Update(
	ctx context.Context,
	target string,
	interfaceName string,
	opts RouteConfigOpts,
) (*Response[Collection[RouteMutationResult]], error) {
	resp, err := s.client.PutNodeNetworkRouteWithResponse(ctx, target, interfaceName,
		gen.NetworkRouteUpdateRequest{Routes: routeStrings(opts.Routes)})
	if err != nil {
		return nil, fmt.Errorf("route update: %w", err)
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

	return newResponse(routeMutationsFromGen(resp.JSON200.Results, resp.JSON200.JobId),
		resp.Body, resp.HTTPResponse), nil
}

// Delete drops osapi's routes of an interface on the target.
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
		return nil, &UnexpectedStatusError{
			APIError{StatusCode: resp.StatusCode(), Message: "nil response body"},
		}
	}

	return newResponse(routeMutationsFromGen(resp.JSON200.Results, resp.JSON200.JobId),
		resp.Body, resp.HTTPResponse), nil
}

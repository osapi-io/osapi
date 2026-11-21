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
	"net/http"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/osapi-io/osapi/pkg/sdk/client/gen"
)

// RouteService reads the routes of an interface of a host, and adds,
// changes and drops the static routes that osapi writes, one at a time.
type RouteService struct {
	client *gen.ClientWithResponses
}

// RouteInfo is one route of an interface on one host.
type RouteInfo struct {
	Hostname    string `json:"hostname"`
	Status      string `json:"status"`
	Error       string `json:"error,omitempty"`
	Destination string `json:"destination,omitempty"`
	Gateway     string `json:"gateway,omitempty"`
	Interface   string `json:"interface,omitempty"`
	Family      string `json:"family,omitempty"`
	Metric      int    `json:"metric,omitempty"`
	Proto       string `json:"proto,omitempty"`
	Scope       string `json:"scope,omitempty"`
}

// RouteOpts is one route: where it goes, its gateway, and its metric, none
// when it is 0.
type RouteOpts struct {
	Destination string
	Via         string
	Metric      int
}

// RouteMutationResult is a route after a write or a removal.
type RouteMutationResult struct {
	Hostname    string `json:"hostname"`
	Status      string `json:"status"`
	Destination string `json:"destination,omitempty"`
	Changed     bool   `json:"changed"`
	Error       string `json:"error,omitempty"`
}

// List lists the routes of an interface on the target, one row per route
// per host.
func (s *RouteService) List(
	ctx context.Context,
	target string,
	interfaceName string,
) (*Response[Collection[RouteInfo]], error) {
	resp, err := s.client.GetNodeNetworkInterfaceRouteWithResponse(ctx, target, interfaceName)
	if err != nil {
		return nil, fmt.Errorf("route list: %w", err)
	}
	if resp.JSON200 == nil {
		return nil, failed(resp.StatusCode(),
			resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON404, resp.JSON500)
	}
	rows := make([]RouteInfo, 0, len(resp.JSON200.Results))
	for _, r := range resp.JSON200.Results {
		rows = append(rows, RouteInfo{
			Hostname:    r.Hostname,
			Status:      string(r.Status),
			Error:       derefString(r.Error),
			Destination: derefString(r.Destination),
			Gateway:     derefString(r.Gateway),
			Interface:   derefString(r.Interface),
			Family:      derefString(r.Family),
			Metric:      derefInt(r.Metric),
			Proto:       derefString(r.Proto),
			Scope:       derefString(r.Scope),
		})
	}

	return newResponse(
		Collection[RouteInfo]{Results: rows, JobID: jobIDFromGen(resp.JSON200.JobId)},
		resp.Body,
		resp.HTTPResponse,
	), nil
}

// Create adds a route to an interface on the target. A route osapi already
// has, the same, does not change.
func (s *RouteService) Create(
	ctx context.Context,
	target string,
	interfaceName string,
	opts RouteOpts,
) (*Response[Collection[RouteMutationResult]], error) {
	resp, err := s.client.PostNodeNetworkInterfaceRouteWithResponse(ctx, target, interfaceName,
		gen.NetworkInterfaceRouteCreateRequest{
			Destination: opts.Destination, Via: opts.Via, Metric: metric(opts.Metric),
		})
	if err != nil {
		return nil, fmt.Errorf("route create: %w", err)
	}
	if resp.JSON200 == nil {
		return nil, failed(resp.StatusCode(),
			resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON404, resp.JSON409, resp.JSON500)
	}

	return routeMutation(
		resp.JSON200.JobId,
		resp.JSON200.Results,
		resp.Body,
		resp.HTTPResponse,
	), nil
}

// Update changes the gateway or the metric of a route that osapi has on the
// target.
func (s *RouteService) Update(
	ctx context.Context,
	target string,
	interfaceName string,
	opts RouteOpts,
) (*Response[Collection[RouteMutationResult]], error) {
	resp, err := s.client.PutNodeNetworkInterfaceRouteWithResponse(ctx, target, interfaceName,
		gen.NetworkInterfaceRouteUpdateRequest{
			Destination: opts.Destination, Via: opts.Via, Metric: metric(opts.Metric),
		})
	if err != nil {
		return nil, fmt.Errorf("route update: %w", err)
	}
	if resp.JSON200 == nil {
		return nil, failed(resp.StatusCode(),
			resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON404, resp.JSON409, resp.JSON500)
	}

	return routeMutation(
		resp.JSON200.JobId,
		resp.JSON200.Results,
		resp.Body,
		resp.HTTPResponse,
	), nil
}

// Delete drops the route to destination that osapi wrote on an interface of
// the target.
func (s *RouteService) Delete(
	ctx context.Context,
	target string,
	interfaceName string,
	destination string,
) (*Response[Collection[RouteMutationResult]], error) {
	resp, err := s.client.DeleteNodeNetworkInterfaceRouteWithResponse(ctx, target, interfaceName,
		gen.NetworkInterfaceRouteDeleteRequest{Destination: destination})
	if err != nil {
		return nil, fmt.Errorf("route delete: %w", err)
	}
	if resp.JSON200 == nil {
		return nil, failed(resp.StatusCode(),
			resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON404, resp.JSON409, resp.JSON500)
	}

	return routeMutation(
		resp.JSON200.JobId,
		resp.JSON200.Results,
		resp.Body,
		resp.HTTPResponse,
	), nil
}

// metric is the metric of a request: none when it is 0.
func metric(
	m int,
) *int {
	if m == 0 {
		return nil
	}

	return &m
}

// routeMutation is the response of a route write.
func routeMutation(
	jobID *openapi_types.UUID,
	rows []gen.NetworkInterfaceRouteResult,
	body []byte,
	httpResp *http.Response,
) *Response[Collection[RouteMutationResult]] {
	results := make([]RouteMutationResult, 0, len(rows))
	for _, r := range rows {
		results = append(results, RouteMutationResult{
			Hostname:    r.Hostname,
			Status:      string(r.Status),
			Destination: derefString(r.Destination),
			Changed:     derefBool(r.Changed),
			Error:       derefString(r.Error),
		})
	}

	return newResponse(
		Collection[RouteMutationResult]{Results: results, JobID: jobIDFromGen(jobID)},
		body,
		httpResp,
	)
}

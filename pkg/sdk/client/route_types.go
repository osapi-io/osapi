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
	"github.com/retr0h/osapi/pkg/sdk/client/gen"
)

// RouteInfo represents a network route entry.
type RouteInfo struct {
	Destination string `json:"destination,omitempty"`
	Gateway     string `json:"gateway,omitempty"`
	Interface   string `json:"interface,omitempty"`
	Metric      int    `json:"metric,omitempty"`
	Scope       string `json:"scope,omitempty"`
}

// RouteListResult represents a route list result from a single agent.
type RouteListResult struct {
	Hostname string      `json:"hostname"`
	Status   string      `json:"status"`
	Error    string      `json:"error,omitempty"`
	Routes   []RouteInfo `json:"routes,omitempty"`
}

// RouteGetResult represents a route get result from a single agent.
type RouteGetResult struct {
	Hostname string      `json:"hostname"`
	Status   string      `json:"status"`
	Error    string      `json:"error,omitempty"`
	Routes   []RouteInfo `json:"routes,omitempty"`
}

// RouteMutationResult represents the result of a route create, update,
// or delete.
type RouteMutationResult struct {
	Hostname  string `json:"hostname"`
	Status    string `json:"status"`
	Interface string `json:"interface,omitempty"`
	Changed   bool   `json:"changed"`
	Error     string `json:"error,omitempty"`
}

// RouteItem represents a single route entry for create/update operations.
type RouteItem struct {
	// To is the destination in CIDR notation. Required.
	To string
	// Via is the gateway IP address. Required.
	Via string
	// Metric is the route metric (priority). Optional.
	Metric *int
}

// RouteConfigOpts contains options for creating or updating routes.
type RouteConfigOpts struct {
	// Routes is the list of route entries. Required.
	Routes []RouteItem
}

// routeInfoFromGen converts a gen.RouteInfo to a RouteInfo.
func routeInfoFromGen(
	g gen.RouteInfo,
) RouteInfo {
	return RouteInfo{
		Destination: derefString(g.Destination),
		Gateway:     derefString(g.Gateway),
		Interface:   derefString(g.Interface),
		Metric:      derefInt(g.Metric),
		Scope:       derefString(g.Scope),
	}
}

// routeListCollectionFromGen converts a gen.RouteListResponse
// to a Collection[RouteListResult].
func routeListCollectionFromGen(
	g *gen.RouteListResponse,
) Collection[RouteListResult] {
	results := make([]RouteListResult, 0, len(g.Results))
	for _, r := range g.Results {
		entry := RouteListResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Error:    derefString(r.Error),
		}

		if r.Routes != nil {
			routes := make([]RouteInfo, 0, len(*r.Routes))
			for _, rt := range *r.Routes {
				routes = append(routes, routeInfoFromGen(rt))
			}

			entry.Routes = routes
		}

		results = append(results, entry)
	}

	return Collection[RouteListResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// routeGetCollectionFromGen converts a gen.RouteGetResponse
// to a Collection[RouteGetResult].
func routeGetCollectionFromGen(
	g *gen.RouteGetResponse,
) Collection[RouteGetResult] {
	results := make([]RouteGetResult, 0, len(g.Results))
	for _, r := range g.Results {
		entry := RouteGetResult{
			Hostname: r.Hostname,
			Status:   string(r.Status),
			Error:    derefString(r.Error),
		}

		if r.Routes != nil {
			routes := make([]RouteInfo, 0, len(*r.Routes))
			for _, rt := range *r.Routes {
				routes = append(routes, routeInfoFromGen(rt))
			}

			entry.Routes = routes
		}

		results = append(results, entry)
	}

	return Collection[RouteGetResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

// routeMutationCollectionFromCreate converts a gen.RouteMutationResponse
// to a Collection[RouteMutationResult].
func routeMutationCollectionFromCreate(
	g *gen.RouteMutationResponse,
) Collection[RouteMutationResult] {
	return routeMutationCollectionFromGen(g)
}

// routeMutationCollectionFromUpdate converts a gen.RouteMutationResponse
// to a Collection[RouteMutationResult].
func routeMutationCollectionFromUpdate(
	g *gen.RouteMutationResponse,
) Collection[RouteMutationResult] {
	return routeMutationCollectionFromGen(g)
}

// routeMutationCollectionFromDelete converts a gen.RouteMutationResponse
// to a Collection[RouteMutationResult].
func routeMutationCollectionFromDelete(
	g *gen.RouteMutationResponse,
) Collection[RouteMutationResult] {
	return routeMutationCollectionFromGen(g)
}

// routeMutationCollectionFromGen converts a gen.RouteMutationResponse
// to a Collection[RouteMutationResult].
func routeMutationCollectionFromGen(
	g *gen.RouteMutationResponse,
) Collection[RouteMutationResult] {
	results := make([]RouteMutationResult, 0, len(g.Results))
	for _, r := range g.Results {
		results = append(results, RouteMutationResult{
			Hostname:  r.Hostname,
			Status:    string(r.Status),
			Interface: derefString(r.Interface),
			Changed:   derefBool(r.Changed),
			Error:     derefString(r.Error),
		})
	}

	return Collection[RouteMutationResult]{
		Results: results,
		JobID:   jobIDFromGen(g.JobId),
	}
}

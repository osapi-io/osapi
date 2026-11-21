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
	"fmt"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/osapi-io/osapi/pkg/sdk/client/gen"
)

// RouteInfo is a route of a host's main routing table.
type RouteInfo struct {
	Destination string `json:"destination,omitempty"`
	Gateway     string `json:"gateway,omitempty"`
	Interface   string `json:"interface,omitempty"`
	Family      string `json:"family,omitempty"`
	Metric      int    `json:"metric,omitempty"`
	Proto       string `json:"proto,omitempty"`
	Scope       string `json:"scope,omitempty"`
}

// RouteListResult is the routes of one host.
type RouteListResult struct {
	Hostname string      `json:"hostname"`
	Status   string      `json:"status"`
	Error    string      `json:"error,omitempty"`
	Routes   []RouteInfo `json:"routes,omitempty"`
}

// RouteGetResult is the routes of one interface of one host.
type RouteGetResult struct {
	Hostname string      `json:"hostname"`
	Status   string      `json:"status"`
	Error    string      `json:"error,omitempty"`
	Routes   []RouteInfo `json:"routes,omitempty"`
}

// RouteMutationResult is an interface's routes after a create, an update or
// a delete on one host.
type RouteMutationResult struct {
	Hostname  string `json:"hostname"`
	Status    string `json:"status"`
	Interface string `json:"interface,omitempty"`
	Changed   bool   `json:"changed"`
	Error     string `json:"error,omitempty"`
}

// RouteItem is a static route.
type RouteItem struct {
	// To is the destination in CIDR notation. Required.
	To string
	// Via is the gateway, of the destination's family. Required.
	Via string
	// Metric is the route metric (priority). Optional.
	Metric *int
}

// RouteConfigOpts is the routes osapi writes for an interface.
type RouteConfigOpts struct {
	// Routes is the routes, one or more. A default route is refused: the
	// interface's gateway is set with InterfaceService.
	Routes []RouteItem
}

// routeStrings is each route as the API takes it: <to> via <via>, with
// metric <n> after it when set.
func routeStrings(
	items []RouteItem,
) []string {
	out := make([]string, 0, len(items))
	for _, r := range items {
		s := r.To + " via " + r.Via
		if r.Metric != nil {
			s += fmt.Sprintf(" metric %d", *r.Metric)
		}
		out = append(out, s)
	}

	return out
}

// routeRowsFromGen groups the rows of a list or a get, one per route, into
// one result per host, in the order each host first answers. A host that
// failed or was skipped is a result without routes.
func routeRowsFromGen(
	rows []gen.NetworkRouteEntry,
) []RouteGetResult {
	results := []RouteGetResult{}
	index := map[string]int{}
	for _, r := range rows {
		i, ok := index[r.Hostname]
		if !ok {
			i = len(results)
			index[r.Hostname] = i
			results = append(results, RouteGetResult{
				Hostname: r.Hostname,
				Status:   string(r.Status),
				Error:    derefString(r.Error),
			})
		}
		if r.Destination != nil {
			results[i].Routes = append(results[i].Routes, RouteInfo{
				Destination: *r.Destination,
				Gateway:     derefString(r.Gateway),
				Interface:   derefString(r.Interface),
				Family:      derefString(r.Family),
				Metric:      derefInt(r.Metric),
				Proto:       derefString(r.Proto),
				Scope:       derefString(r.Scope),
			})
		}
	}

	return results
}

// routeMutationsFromGen is the rows of a create, an update or a delete.
func routeMutationsFromGen(
	rows []gen.NetworkRouteResult,
	jobID *openapi_types.UUID,
) Collection[RouteMutationResult] {
	results := make([]RouteMutationResult, 0, len(rows))
	for _, r := range rows {
		results = append(results, RouteMutationResult{
			Hostname:  r.Hostname,
			Status:    string(r.Status),
			Interface: derefString(r.InterfaceName),
			Changed:   derefBool(r.Changed),
			Error:     derefString(r.Error),
		})
	}

	return Collection[RouteMutationResult]{Results: results, JobID: jobIDFromGen(jobID)}
}

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

package network

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/node/network/gen"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/validation"
)

// PostNodeNetworkRoute post the node network route create API endpoint.
func (s *Network) PostNodeNetworkRoute(
	ctx context.Context,
	request gen.PostNodeNetworkRouteRequestObject,
) (gen.PostNodeNetworkRouteResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.PostNodeNetworkRoute400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validateInterfaceName(request.InterfaceName); !ok {
		return gen.PostNodeNetworkRoute400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validation.Struct(request.Body); !ok {
		return gen.PostNodeNetworkRoute400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	s.logger.Debug("route create",
		slog.String("interface", request.InterfaceName),
		slog.Int("routes", len(request.Body.Routes)),
		slog.String("target", hostname),
	)

	data := buildRouteData(request.InterfaceName, request.Body)

	if job.IsBroadcastTarget(hostname) {
		return s.postNodeNetworkRouteBroadcast(ctx, hostname, request.InterfaceName, data)
	}

	jobID, rawResp, err := s.JobClient.Modify(
		ctx,
		hostname,
		"network",
		job.OperationNetworkRouteCreate,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeNetworkRoute500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	if rawResp.Status == job.StatusSkipped {
		jobUUID := uuid.MustParse(jobID)
		e := rawResp.Error
		falseVal := false
		return gen.PostNodeNetworkRoute200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.RouteMutationEntry{
				{
					Hostname: rawResp.Hostname,
					Status:   gen.RouteMutationEntryStatusSkipped,
					Error:    &e,
					Changed:  &falseVal,
				},
			},
		}, nil
	}

	changed := rawResp.Changed == nil || *rawResp.Changed
	iface := request.InterfaceName
	jobUUID := uuid.MustParse(jobID)
	return gen.PostNodeNetworkRoute200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.RouteMutationEntry{
			{
				Hostname:  rawResp.Hostname,
				Status:    gen.RouteMutationEntryStatusOk,
				Changed:   &changed,
				Interface: &iface,
			},
		},
	}, nil
}

// postNodeNetworkRouteBroadcast handles broadcast targets for route create.
func (s *Network) postNodeNetworkRouteBroadcast(
	ctx context.Context,
	target string,
	iface string,
	data map[string]any,
) (gen.PostNodeNetworkRouteResponseObject, error) {
	jobID, responses, err := s.JobClient.ModifyBroadcast(
		ctx,
		target,
		"network",
		job.OperationNetworkRouteCreate,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeNetworkRoute500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	apiResponses := buildRouteMutationResults(responses, iface)
	jobUUID := uuid.MustParse(jobID)
	return gen.PostNodeNetworkRoute200JSONResponse{
		JobId:   &jobUUID,
		Results: apiResponses,
	}, nil
}

// buildRouteData constructs the job payload from the route request body.
func buildRouteData(
	interfaceName string,
	body *gen.RouteConfigRequest,
) map[string]any {
	var routes []map[string]any
	for _, r := range body.Routes {
		route := map[string]any{
			"to":  r.To,
			"via": r.Via,
		}
		if r.Metric != nil {
			route["metric"] = *r.Metric
		}
		routes = append(routes, route)
	}
	return map[string]any{
		"interface": interfaceName,
		"routes":    routes,
	}
}

// buildRouteMutationResults converts broadcast responses to RouteMutationEntry slice.
func buildRouteMutationResults(
	responses map[string]*job.Response,
	iface string,
) []gen.RouteMutationEntry {
	var apiResponses []gen.RouteMutationEntry
	for host, resp := range responses {
		item := gen.RouteMutationEntry{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.RouteMutationEntryStatusFailed
			e := resp.Error
			falseVal := false
			item.Error = &e
			item.Changed = &falseVal
		case job.StatusSkipped:
			item.Status = gen.RouteMutationEntryStatusSkipped
			e := resp.Error
			falseVal := false
			item.Error = &e
			item.Changed = &falseVal
		default:
			item.Status = gen.RouteMutationEntryStatusOk
			changed := resp.Changed == nil || *resp.Changed
			item.Changed = &changed
			item.Interface = &iface
		}
		apiResponses = append(apiResponses, item)
	}
	return apiResponses
}

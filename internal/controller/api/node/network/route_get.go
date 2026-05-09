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
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/node/network/gen"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider/network/netplan/route"
)

// GetNodeNetworkRouteByInterface get the node network routes by interface API endpoint.
func (s *Network) GetNodeNetworkRouteByInterface(
	ctx context.Context,
	request gen.GetNodeNetworkRouteByInterfaceRequestObject,
) (gen.GetNodeNetworkRouteByInterfaceResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeNetworkRouteByInterface400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validateInterfaceName(request.InterfaceName); !ok {
		return gen.GetNodeNetworkRouteByInterface400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	s.logger.Debug(
		"route get",
		slog.String("interface", request.InterfaceName),
		slog.String("target", hostname),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.getNodeNetworkRouteByInterfaceBroadcast(ctx, hostname, request.InterfaceName)
	}

	data := map[string]any{"interface": request.InterfaceName}
	jobID, rawResp, err := s.JobClient.Query(
		ctx,
		hostname,
		"network",
		job.OperationNetworkRouteGet,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeNetworkRouteByInterface500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	if rawResp.Status == job.StatusSkipped {
		e := rawResp.Error
		jobUUID := uuid.MustParse(jobID)
		return gen.GetNodeNetworkRouteByInterface200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.RouteGetEntry{
				{
					Hostname: rawResp.Hostname,
					Status:   gen.RouteGetEntryStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	var entry route.Entry
	if rawResp.Data != nil {
		_ = json.Unmarshal(rawResp.Data, &entry)
	}

	routes := convertRouteEntryRoutes(entry.Routes)
	jobUUID := uuid.MustParse(jobID)
	return gen.GetNodeNetworkRouteByInterface200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.RouteGetEntry{
			{
				Hostname: rawResp.Hostname,
				Status:   gen.RouteGetEntryStatusOk,
				Routes:   &routes,
			},
		},
	}, nil
}

// getNodeNetworkRouteByInterfaceBroadcast handles broadcast targets for route get.
func (s *Network) getNodeNetworkRouteByInterfaceBroadcast(
	ctx context.Context,
	target string,
	iface string,
) (gen.GetNodeNetworkRouteByInterfaceResponseObject, error) {
	data := map[string]any{"interface": iface}
	jobID, responses, err := s.JobClient.QueryBroadcast(
		ctx,
		target,
		"network",
		job.OperationNetworkRouteGet,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeNetworkRouteByInterface500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	apiResponses := make([]gen.RouteGetEntry, 0, len(responses))
	for host, resp := range responses {
		item := gen.RouteGetEntry{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.RouteGetEntryStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.RouteGetEntryStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.RouteGetEntryStatusOk
			var entry route.Entry
			if resp.Data != nil {
				_ = json.Unmarshal(resp.Data, &entry)
			}
			routes := convertRouteEntryRoutes(entry.Routes)
			item.Routes = &routes
		}
		apiResponses = append(apiResponses, item)
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.GetNodeNetworkRouteByInterface200JSONResponse{
		JobId:   &jobUUID,
		Results: apiResponses,
	}, nil
}

// convertRouteEntryRoutes converts provider routes to API RouteInfo.
func convertRouteEntryRoutes(
	routes []route.Route,
) []gen.RouteInfo {
	result := make([]gen.RouteInfo, 0, len(routes))
	for _, r := range routes {
		info := gen.RouteInfo{
			Destination: strPtrOrNil(r.To),
			Gateway:     strPtrOrNil(r.Via),
		}
		if r.Metric > 0 {
			metric := r.Metric
			info.Metric = &metric
		}
		result = append(result, info)
	}
	return result
}

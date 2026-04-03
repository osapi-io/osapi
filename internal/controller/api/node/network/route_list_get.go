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
	"github.com/retr0h/osapi/internal/provider/network/netplan"
)

// GetNodeNetworkRoute get the node network route list API endpoint.
func (s *Network) GetNodeNetworkRoute(
	ctx context.Context,
	request gen.GetNodeNetworkRouteRequestObject,
) (gen.GetNodeNetworkRouteResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeNetworkRoute400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	s.logger.Debug("route list",
		slog.String("target", hostname),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.getNodeNetworkRouteListBroadcast(ctx, hostname)
	}

	jobID, rawResp, err := s.JobClient.Query(
		ctx,
		hostname,
		"network",
		job.OperationNetworkRouteList,
		nil,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeNetworkRoute500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	if rawResp.Status == job.StatusSkipped {
		e := rawResp.Error
		jobUUID := uuid.MustParse(jobID)
		return gen.GetNodeNetworkRoute200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.RouteListEntry{
				{
					Hostname: rawResp.Hostname,
					Status:   gen.RouteListEntryStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	var entries []netplan.RouteListEntry
	if rawResp.Data != nil {
		_ = json.Unmarshal(rawResp.Data, &entries)
	}

	routes := convertRouteListEntries(entries)
	jobUUID := uuid.MustParse(jobID)
	return gen.GetNodeNetworkRoute200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.RouteListEntry{
			{
				Hostname: rawResp.Hostname,
				Status:   gen.RouteListEntryStatusOk,
				Routes:   &routes,
			},
		},
	}, nil
}

// getNodeNetworkRouteListBroadcast handles broadcast targets for route list.
func (s *Network) getNodeNetworkRouteListBroadcast(
	ctx context.Context,
	target string,
) (gen.GetNodeNetworkRouteResponseObject, error) {
	jobID, responses, err := s.JobClient.QueryBroadcast(
		ctx,
		target,
		"network",
		job.OperationNetworkRouteList,
		nil,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeNetworkRoute500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	apiResponses := make([]gen.RouteListEntry, 0, len(responses))
	for host, resp := range responses {
		item := gen.RouteListEntry{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.RouteListEntryStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.RouteListEntryStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.RouteListEntryStatusOk
			var entries []netplan.RouteListEntry
			if resp.Data != nil {
				_ = json.Unmarshal(resp.Data, &entries)
			}
			routes := convertRouteListEntries(entries)
			item.Routes = &routes
		}
		apiResponses = append(apiResponses, item)
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.GetNodeNetworkRoute200JSONResponse{
		JobId:   &jobUUID,
		Results: apiResponses,
	}, nil
}

// convertRouteListEntries converts provider route list entries to API RouteInfo.
func convertRouteListEntries(
	entries []netplan.RouteListEntry,
) []gen.RouteInfo {
	result := make([]gen.RouteInfo, 0, len(entries))
	for _, e := range entries {
		info := gen.RouteInfo{
			Destination: strPtrOrNil(e.Destination),
			Gateway:     strPtrOrNil(e.Gateway),
			Interface:   strPtrOrNil(e.Interface),
			Scope:       strPtrOrNil(e.Flags),
		}
		if e.Metric > 0 {
			metric := e.Metric
			info.Metric = &metric
		}
		result = append(result, info)
	}
	return result
}

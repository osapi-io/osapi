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
)

// DeleteNodeNetworkRoute delete the node network routes by interface API endpoint.
func (s *Network) DeleteNodeNetworkRoute(
	ctx context.Context,
	request gen.DeleteNodeNetworkRouteRequestObject,
) (gen.DeleteNodeNetworkRouteResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.DeleteNodeNetworkRoute400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validateInterfaceName(request.InterfaceName); !ok {
		return gen.DeleteNodeNetworkRoute400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	s.logger.Debug("route delete",
		slog.String("interface", request.InterfaceName),
		slog.String("target", hostname),
	)

	data := map[string]any{"interface": request.InterfaceName}

	if job.IsBroadcastTarget(hostname) {
		return s.deleteNodeNetworkRouteBroadcast(ctx, hostname, request.InterfaceName, data)
	}

	jobID, rawResp, err := s.JobClient.Modify(
		ctx,
		hostname,
		"network",
		job.OperationNetworkRouteDelete,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.DeleteNodeNetworkRoute500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	if rawResp.Status == job.StatusSkipped {
		jobUUID := uuid.MustParse(jobID)
		e := rawResp.Error
		falseVal := false
		return gen.DeleteNodeNetworkRoute200JSONResponse{
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
	return gen.DeleteNodeNetworkRoute200JSONResponse{
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

// deleteNodeNetworkRouteBroadcast handles broadcast targets for route delete.
func (s *Network) deleteNodeNetworkRouteBroadcast(
	ctx context.Context,
	target string,
	iface string,
	data map[string]any,
) (gen.DeleteNodeNetworkRouteResponseObject, error) {
	jobID, responses, err := s.JobClient.ModifyBroadcast(
		ctx,
		target,
		"network",
		job.OperationNetworkRouteDelete,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.DeleteNodeNetworkRoute500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	apiResponses := buildRouteMutationResults(responses, iface)
	jobUUID := uuid.MustParse(jobID)
	return gen.DeleteNodeNetworkRoute200JSONResponse{
		JobId:   &jobUUID,
		Results: apiResponses,
	}, nil
}

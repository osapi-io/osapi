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

// PutNodeNetworkRoute put the node network route update API endpoint.
func (s *Network) PutNodeNetworkRoute(
	ctx context.Context,
	request gen.PutNodeNetworkRouteRequestObject,
) (gen.PutNodeNetworkRouteResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.PutNodeNetworkRoute400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validateInterfaceName(request.InterfaceName); !ok {
		return gen.PutNodeNetworkRoute400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validation.Struct(request.Body); !ok {
		return gen.PutNodeNetworkRoute400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	s.logger.Debug(
		"route update",
		slog.String("interface", request.InterfaceName),
		slog.Int("routes", len(request.Body.Routes)),
		slog.String("target", hostname),
	)

	data := buildRouteData(request.InterfaceName, request.Body)

	if job.IsBroadcastTarget(hostname) {
		return s.putNodeNetworkRouteBroadcast(ctx, hostname, request.InterfaceName, data)
	}

	jobID, rawResp, err := s.JobClient.Modify(
		ctx,
		hostname,
		"network",
		job.OperationNetworkRouteUpdate,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PutNodeNetworkRoute500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	if rawResp.Status == job.StatusSkipped {
		jobUUID := uuid.MustParse(jobID)
		e := rawResp.Error
		falseVal := false
		return gen.PutNodeNetworkRoute200JSONResponse{
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
	return gen.PutNodeNetworkRoute200JSONResponse{
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

// putNodeNetworkRouteBroadcast handles broadcast targets for route update.
func (s *Network) putNodeNetworkRouteBroadcast(
	ctx context.Context,
	target string,
	iface string,
	data map[string]any,
) (gen.PutNodeNetworkRouteResponseObject, error) {
	jobID, responses, err := s.JobClient.ModifyBroadcast(
		ctx,
		target,
		"network",
		job.OperationNetworkRouteUpdate,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PutNodeNetworkRoute500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	apiResponses := buildRouteMutationResults(responses, iface)
	jobUUID := uuid.MustParse(jobID)
	return gen.PutNodeNetworkRoute200JSONResponse{
		JobId:   &jobUUID,
		Results: apiResponses,
	}, nil
}

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

// DeleteNodeNetworkInterface delete the node network interface API endpoint.
func (s *Network) DeleteNodeNetworkInterface(
	ctx context.Context,
	request gen.DeleteNodeNetworkInterfaceRequestObject,
) (gen.DeleteNodeNetworkInterfaceResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.DeleteNodeNetworkInterface400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validateInterfaceName(request.Name); !ok {
		return gen.DeleteNodeNetworkInterface400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	s.logger.Debug("interface delete",
		slog.String("name", request.Name),
		slog.String("target", hostname),
	)

	data := map[string]any{"name": request.Name}

	if job.IsBroadcastTarget(hostname) {
		return s.deleteNodeNetworkInterfaceBroadcast(ctx, hostname, request.Name, data)
	}

	jobID, rawResp, err := s.JobClient.Modify(
		ctx,
		hostname,
		"network",
		job.OperationNetworkInterfaceDelete,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.DeleteNodeNetworkInterface500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	if rawResp.Status == job.StatusSkipped {
		jobUUID := uuid.MustParse(jobID)
		e := rawResp.Error
		falseVal := false
		return gen.DeleteNodeNetworkInterface200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.InterfaceMutationEntry{
				{
					Hostname: rawResp.Hostname,
					Status:   gen.InterfaceMutationEntryStatusSkipped,
					Error:    &e,
					Changed:  &falseVal,
				},
			},
		}, nil
	}

	changed := rawResp.Changed == nil || *rawResp.Changed
	name := request.Name
	jobUUID := uuid.MustParse(jobID)
	return gen.DeleteNodeNetworkInterface200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.InterfaceMutationEntry{
			{
				Hostname: rawResp.Hostname,
				Status:   gen.InterfaceMutationEntryStatusOk,
				Changed:  &changed,
				Name:     &name,
			},
		},
	}, nil
}

// deleteNodeNetworkInterfaceBroadcast handles broadcast targets for interface delete.
func (s *Network) deleteNodeNetworkInterfaceBroadcast(
	ctx context.Context,
	target string,
	name string,
	data map[string]any,
) (gen.DeleteNodeNetworkInterfaceResponseObject, error) {
	jobID, responses, err := s.JobClient.ModifyBroadcast(
		ctx,
		target,
		"network",
		job.OperationNetworkInterfaceDelete,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.DeleteNodeNetworkInterface500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	apiResponses := buildInterfaceMutationResults(responses, name)
	jobUUID := uuid.MustParse(jobID)
	return gen.DeleteNodeNetworkInterface200JSONResponse{
		JobId:   &jobUUID,
		Results: apiResponses,
	}, nil
}

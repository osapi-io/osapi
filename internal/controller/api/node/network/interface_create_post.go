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

// PostNodeNetworkInterface post the node network interface create API endpoint.
func (s *Network) PostNodeNetworkInterface(
	ctx context.Context,
	request gen.PostNodeNetworkInterfaceRequestObject,
) (gen.PostNodeNetworkInterfaceResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.PostNodeNetworkInterface400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validateInterfaceName(request.Name); !ok {
		return gen.PostNodeNetworkInterface400JSONResponse{Error: &errMsg}, nil
	}

	// Defense in depth: current fields use omitempty so validation
	// always passes, but guards against future field additions.
	if errMsg, ok := validation.Struct(request.Body); !ok {
		return gen.PostNodeNetworkInterface400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	s.logger.Debug("interface create",
		slog.String("name", request.Name),
		slog.String("target", hostname),
	)

	data := buildInterfaceData(request.Name, request.Body)

	if job.IsBroadcastTarget(hostname) {
		return s.postNodeNetworkInterfaceBroadcast(ctx, hostname, request.Name, data)
	}

	jobID, rawResp, err := s.JobClient.Modify(
		ctx,
		hostname,
		"network",
		job.OperationNetworkInterfaceCreate,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeNetworkInterface500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	if rawResp.Status == job.StatusSkipped {
		jobUUID := uuid.MustParse(jobID)
		e := rawResp.Error
		falseVal := false
		return gen.PostNodeNetworkInterface200JSONResponse{
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
	return gen.PostNodeNetworkInterface200JSONResponse{
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

// postNodeNetworkInterfaceBroadcast handles broadcast targets for interface create.
func (s *Network) postNodeNetworkInterfaceBroadcast(
	ctx context.Context,
	target string,
	name string,
	data map[string]any,
) (gen.PostNodeNetworkInterfaceResponseObject, error) {
	jobID, responses, err := s.JobClient.ModifyBroadcast(
		ctx,
		target,
		"network",
		job.OperationNetworkInterfaceCreate,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeNetworkInterface500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	apiResponses := buildInterfaceMutationResults(responses, name)
	jobUUID := uuid.MustParse(jobID)
	return gen.PostNodeNetworkInterface200JSONResponse{
		JobId:   &jobUUID,
		Results: apiResponses,
	}, nil
}

// buildInterfaceData constructs the job payload from the request body.
func buildInterfaceData(
	name string,
	body *gen.InterfaceConfigRequest,
) map[string]any {
	data := map[string]any{
		"name": name,
	}
	if body.Dhcp4 != nil {
		data["dhcp4"] = *body.Dhcp4
	}
	if body.Dhcp6 != nil {
		data["dhcp6"] = *body.Dhcp6
	}
	if body.Addresses != nil {
		data["addresses"] = *body.Addresses
	}
	if body.Gateway4 != nil {
		data["gateway4"] = *body.Gateway4
	}
	if body.Gateway6 != nil {
		data["gateway6"] = *body.Gateway6
	}
	if body.Mtu != nil {
		data["mtu"] = *body.Mtu
	}
	if body.MacAddress != nil {
		data["mac_address"] = *body.MacAddress
	}
	if body.Wakeonlan != nil {
		data["wakeonlan"] = *body.Wakeonlan
	}
	return data
}

// buildInterfaceMutationResults converts broadcast responses to InterfaceMutationEntry slice.
func buildInterfaceMutationResults(
	responses map[string]*job.Response,
	name string,
) []gen.InterfaceMutationEntry {
	var apiResponses []gen.InterfaceMutationEntry
	for host, resp := range responses {
		item := gen.InterfaceMutationEntry{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.InterfaceMutationEntryStatusFailed
			e := resp.Error
			falseVal := false
			item.Error = &e
			item.Changed = &falseVal
		case job.StatusSkipped:
			item.Status = gen.InterfaceMutationEntryStatusSkipped
			e := resp.Error
			falseVal := false
			item.Error = &e
			item.Changed = &falseVal
		default:
			item.Status = gen.InterfaceMutationEntryStatusOk
			changed := resp.Changed == nil || *resp.Changed
			item.Changed = &changed
			item.Name = &name
		}
		apiResponses = append(apiResponses, item)
	}
	return apiResponses
}

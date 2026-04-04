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
	"github.com/retr0h/osapi/internal/provider/network/netif"
)

// GetNodeNetworkInterfaceByName get the node network interface by name API endpoint.
func (s *Network) GetNodeNetworkInterfaceByName(
	ctx context.Context,
	request gen.GetNodeNetworkInterfaceByNameRequestObject,
) (gen.GetNodeNetworkInterfaceByNameResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeNetworkInterfaceByName400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validateInterfaceName(request.Name); !ok {
		return gen.GetNodeNetworkInterfaceByName400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	s.logger.Debug("interface get",
		slog.String("name", request.Name),
		slog.String("target", hostname),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.getNodeNetworkInterfaceByNameBroadcast(ctx, hostname, request.Name)
	}

	data := map[string]any{"name": request.Name}
	jobID, rawResp, err := s.JobClient.Query(
		ctx,
		hostname,
		"network",
		job.OperationNetworkInterfaceGet,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeNetworkInterfaceByName500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	if rawResp.Status == job.StatusSkipped {
		e := rawResp.Error
		jobUUID := uuid.MustParse(jobID)
		return gen.GetNodeNetworkInterfaceByName200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.InterfaceGetEntry{
				{
					Hostname: rawResp.Hostname,
					Status:   gen.InterfaceGetEntryStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	var entry netif.InterfaceEntry
	if rawResp.Data != nil {
		_ = json.Unmarshal(rawResp.Data, &entry)
	}

	info := convertInterfaceEntry(entry)
	jobUUID := uuid.MustParse(jobID)
	return gen.GetNodeNetworkInterfaceByName200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.InterfaceGetEntry{
			{
				Hostname:  rawResp.Hostname,
				Status:    gen.InterfaceGetEntryStatusOk,
				Interface: &info,
			},
		},
	}, nil
}

// getNodeNetworkInterfaceByNameBroadcast handles broadcast targets for interface get.
func (s *Network) getNodeNetworkInterfaceByNameBroadcast(
	ctx context.Context,
	target string,
	name string,
) (gen.GetNodeNetworkInterfaceByNameResponseObject, error) {
	data := map[string]any{"name": name}
	jobID, responses, err := s.JobClient.QueryBroadcast(
		ctx,
		target,
		"network",
		job.OperationNetworkInterfaceGet,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeNetworkInterfaceByName500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	apiResponses := make([]gen.InterfaceGetEntry, 0, len(responses))
	for host, resp := range responses {
		item := gen.InterfaceGetEntry{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.InterfaceGetEntryStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.InterfaceGetEntryStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.InterfaceGetEntryStatusOk
			var entry netif.InterfaceEntry
			if resp.Data != nil {
				_ = json.Unmarshal(resp.Data, &entry)
			}
			info := convertInterfaceEntry(entry)
			item.Interface = &info
		}
		apiResponses = append(apiResponses, item)
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.GetNodeNetworkInterfaceByName200JSONResponse{
		JobId:   &jobUUID,
		Results: apiResponses,
	}, nil
}

// convertInterfaceEntry converts a provider entry to API InterfaceInfo.
func convertInterfaceEntry(
	e netif.InterfaceEntry,
) gen.InterfaceInfo {
	info := gen.InterfaceInfo{
		Name:       &e.Name,
		Dhcp4:      e.DHCP4,
		Dhcp6:      e.DHCP6,
		Wakeonlan:  e.WakeOnLAN,
		Managed:    &e.Managed,
		MacAddress: strPtrOrNil(e.MACAddress),
		Gateway4:   strPtrOrNil(e.Gateway4),
		Gateway6:   strPtrOrNil(e.Gateway6),
	}
	if len(e.Addresses) > 0 {
		addrs := e.Addresses
		info.Addresses = &addrs
	}
	if e.MTU > 0 {
		mtu := e.MTU
		info.Mtu = &mtu
	}
	return info
}

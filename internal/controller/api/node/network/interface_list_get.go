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

// GetNodeNetworkInterface get the node network interface list API endpoint.
func (s *Network) GetNodeNetworkInterface(
	ctx context.Context,
	request gen.GetNodeNetworkInterfaceRequestObject,
) (gen.GetNodeNetworkInterfaceResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeNetworkInterface400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	s.logger.Debug("interface list",
		slog.String("target", hostname),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.getNodeNetworkInterfaceListBroadcast(ctx, hostname)
	}

	jobID, rawResp, err := s.JobClient.Query(
		ctx,
		hostname,
		"network",
		job.OperationNetworkInterfaceList,
		nil,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeNetworkInterface500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	if rawResp.Status == job.StatusSkipped {
		e := rawResp.Error
		jobUUID := uuid.MustParse(jobID)
		return gen.GetNodeNetworkInterface200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.InterfaceListEntry{
				{
					Hostname: rawResp.Hostname,
					Status:   gen.InterfaceListEntryStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	var entries []netplan.InterfaceEntry
	if rawResp.Data != nil {
		_ = json.Unmarshal(rawResp.Data, &entries)
	}

	interfaces := convertInterfaceEntries(entries)
	jobUUID := uuid.MustParse(jobID)
	return gen.GetNodeNetworkInterface200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.InterfaceListEntry{
			{
				Hostname:   rawResp.Hostname,
				Status:     gen.InterfaceListEntryStatusOk,
				Interfaces: &interfaces,
			},
		},
	}, nil
}

// getNodeNetworkInterfaceListBroadcast handles broadcast targets for interface list.
func (s *Network) getNodeNetworkInterfaceListBroadcast(
	ctx context.Context,
	target string,
) (gen.GetNodeNetworkInterfaceResponseObject, error) {
	jobID, responses, err := s.JobClient.QueryBroadcast(
		ctx,
		target,
		"network",
		job.OperationNetworkInterfaceList,
		nil,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeNetworkInterface500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	var apiResponses []gen.InterfaceListEntry
	for host, resp := range responses {
		item := gen.InterfaceListEntry{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.InterfaceListEntryStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.InterfaceListEntryStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.InterfaceListEntryStatusOk
			var entries []netplan.InterfaceEntry
			if resp.Data != nil {
				_ = json.Unmarshal(resp.Data, &entries)
			}
			interfaces := convertInterfaceEntries(entries)
			item.Interfaces = &interfaces
		}
		apiResponses = append(apiResponses, item)
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.GetNodeNetworkInterface200JSONResponse{
		JobId:   &jobUUID,
		Results: apiResponses,
	}, nil
}

// convertInterfaceEntries converts provider entries to API InterfaceInfo.
func convertInterfaceEntries(
	entries []netplan.InterfaceEntry,
) []gen.InterfaceInfo {
	var result []gen.InterfaceInfo
	for _, e := range entries {
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
		result = append(result, info)
	}
	return result
}

// strPtrOrNil returns a pointer to s if non-empty, otherwise nil.
func strPtrOrNil(
	s string,
) *string {
	if s == "" {
		return nil
	}
	return &s
}

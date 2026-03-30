// Copyright (c) 2024 John Dewey

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
	"github.com/retr0h/osapi/internal/provider/network/dns"
)

// GetNodeNetworkDNSByInterface get the node network dns get API endpoint.
func (s *Network) GetNodeNetworkDNSByInterface(
	ctx context.Context,
	request gen.GetNodeNetworkDNSByInterfaceRequestObject,
) (gen.GetNodeNetworkDNSByInterfaceResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeNetworkDNSByInterface400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validateInterfaceName(request.InterfaceName); !ok {
		return gen.GetNodeNetworkDNSByInterface400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	s.logger.Debug("dns get",
		slog.String("interface", request.InterfaceName),
		slog.String("target", hostname),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.getNodeNetworkDNSBroadcast(ctx, hostname, request.InterfaceName)
	}

	data := map[string]any{"interface": request.InterfaceName}
	jobID, rawResp, err := s.JobClient.Query(
		ctx,
		hostname,
		"network",
		job.OperationNetworkDNSGet,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeNetworkDNSByInterface500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	if rawResp.Status == job.StatusSkipped {
		e := rawResp.Error
		jobUUID := uuid.MustParse(jobID)
		return gen.GetNodeNetworkDNSByInterface200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.DNSConfigResponse{
				{
					Hostname: rawResp.Hostname,
					Status:   gen.DNSConfigResponseStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	var dnsConfig dns.GetResult
	if rawResp.Data != nil {
		_ = json.Unmarshal(rawResp.Data, &dnsConfig)
	}

	searchDomains := dnsConfig.SearchDomains
	servers := dnsConfig.DNSServers
	changed := false

	jobUUID := uuid.MustParse(jobID)
	return gen.GetNodeNetworkDNSByInterface200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.DNSConfigResponse{
			{
				Hostname:      rawResp.Hostname,
				Status:        gen.DNSConfigResponseStatusOk,
				Servers:       &servers,
				SearchDomains: &searchDomains,
				Changed:       &changed,
			},
		},
	}, nil
}

// getNodeNetworkDNSBroadcast handles broadcast targets (_all or label) for DNS config.
func (s *Network) getNodeNetworkDNSBroadcast(
	ctx context.Context,
	target string,
	iface string,
) (gen.GetNodeNetworkDNSByInterfaceResponseObject, error) {
	data := map[string]any{"interface": iface}
	jobID, responses, err := s.JobClient.QueryBroadcast(
		ctx,
		target,
		"network",
		job.OperationNetworkDNSGet,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeNetworkDNSByInterface500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	var apiResponses []gen.DNSConfigResponse
	for host, resp := range responses {
		item := gen.DNSConfigResponse{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.DNSConfigResponseStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.DNSConfigResponseStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.DNSConfigResponseStatusOk
			var cfg dns.GetResult
			if resp.Data != nil {
				_ = json.Unmarshal(resp.Data, &cfg)
			}
			servers := cfg.DNSServers
			searchDomains := cfg.SearchDomains
			changed := false
			item.Servers = &servers
			item.SearchDomains = &searchDomains
			item.Changed = &changed
		}
		apiResponses = append(apiResponses, item)
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.GetNodeNetworkDNSByInterface200JSONResponse{
		JobId:   &jobUUID,
		Results: apiResponses,
	}, nil
}

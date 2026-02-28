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

package node

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/api/node/gen"
	"github.com/retr0h/osapi/internal/job"
)

// GetNodeNetworkDNSByInterface get the node network dns get API endpoint.
func (s *Node) GetNodeNetworkDNSByInterface(
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

	jobID, dnsConfig, agentHostname, err := s.JobClient.QueryNetworkDNS(
		ctx,
		hostname,
		request.InterfaceName,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeNetworkDNSByInterface500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	searchDomains := dnsConfig.SearchDomains
	servers := dnsConfig.DNSServers

	jobUUID := uuid.MustParse(jobID)
	return gen.GetNodeNetworkDNSByInterface200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.DNSConfigResponse{
			{
				Hostname:      agentHostname,
				Servers:       &servers,
				SearchDomains: &searchDomains,
			},
		},
	}, nil
}

// getNodeNetworkDNSBroadcast handles broadcast targets (_all or label) for DNS config.
func (s *Node) getNodeNetworkDNSBroadcast(
	ctx context.Context,
	target string,
	iface string,
) (gen.GetNodeNetworkDNSByInterfaceResponseObject, error) {
	jobID, results, errs, err := s.JobClient.QueryNetworkDNSBroadcast(ctx, target, iface)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeNetworkDNSByInterface500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	var responses []gen.DNSConfigResponse
	for host, cfg := range results {
		servers := cfg.DNSServers
		searchDomains := cfg.SearchDomains
		responses = append(responses, gen.DNSConfigResponse{
			Hostname:      host,
			Servers:       &servers,
			SearchDomains: &searchDomains,
		})
	}
	for host, errMsg := range errs {
		e := errMsg
		responses = append(responses, gen.DNSConfigResponse{
			Hostname: host,
			Error:    &e,
		})
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.GetNodeNetworkDNSByInterface200JSONResponse{
		JobId:   &jobUUID,
		Results: responses,
	}, nil
}

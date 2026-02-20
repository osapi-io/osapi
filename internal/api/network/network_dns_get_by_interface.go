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
	"log/slog"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/api/network/gen"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/validation"
)

// GetNetworkDNSByInterface get the network dns get API endpoint.
func (n Network) GetNetworkDNSByInterface(
	ctx context.Context,
	request gen.GetNetworkDNSByInterfaceRequestObject,
) (gen.GetNetworkDNSByInterfaceResponseObject, error) {
	iface := struct {
		InterfaceName string `validate:"required,alphanum"`
	}{InterfaceName: request.InterfaceName}
	if errMsg, ok := validation.Struct(iface); !ok {
		return gen.GetNetworkDNSByInterface400JSONResponse{Error: &errMsg}, nil
	}

	if request.Params.TargetHostname != nil {
		th := struct {
			TargetHostname string `validate:"min=1"`
		}{TargetHostname: *request.Params.TargetHostname}
		if errMsg, ok := validation.Struct(th); !ok {
			return gen.GetNetworkDNSByInterface400JSONResponse{Error: &errMsg}, nil
		}
	}

	hostname := job.AnyHost
	if request.Params.TargetHostname != nil {
		hostname = *request.Params.TargetHostname
	}

	n.logger.Debug("dns get",
		slog.String("interface", request.InterfaceName),
		slog.String("target", hostname),
	)

	if job.IsBroadcastTarget(hostname) {
		return n.getNetworkDNSBroadcast(ctx, hostname, request.InterfaceName)
	}

	jobID, dnsConfig, workerHostname, err := n.JobClient.QueryNetworkDNS(
		ctx,
		hostname,
		request.InterfaceName,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNetworkDNSByInterface500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	searchDomains := dnsConfig.SearchDomains
	servers := dnsConfig.DNSServers

	jobUUID := uuid.MustParse(jobID)
	return gen.GetNetworkDNSByInterface200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.DNSConfigResponse{
			{
				Hostname:      workerHostname,
				Servers:       &servers,
				SearchDomains: &searchDomains,
			},
		},
	}, nil
}

// getNetworkDNSBroadcast handles broadcast targets (_all or label) for DNS config.
func (n Network) getNetworkDNSBroadcast(
	ctx context.Context,
	target string,
	iface string,
) (gen.GetNetworkDNSByInterfaceResponseObject, error) {
	jobID, results, err := n.JobClient.QueryNetworkDNSBroadcast(ctx, target, iface)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNetworkDNSByInterface500JSONResponse{
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

	jobUUID := uuid.MustParse(jobID)
	return gen.GetNetworkDNSByInterface200JSONResponse{
		JobId:   &jobUUID,
		Results: responses,
	}, nil
}

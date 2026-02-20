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
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/api/network/gen"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/validation"
)

// PutNetworkDNS put the network dns API endpoint.
func (n Network) PutNetworkDNS(
	ctx context.Context,
	request gen.PutNetworkDNSRequestObject,
) (gen.PutNetworkDNSResponseObject, error) {
	if errMsg, ok := validation.Struct(request.Body); !ok {
		return gen.PutNetworkDNS400JSONResponse{
			Error: &errMsg,
		}, nil
	}

	if request.Params.TargetHostname != nil {
		th := struct {
			TargetHostname string `validate:"min=1"`
		}{TargetHostname: *request.Params.TargetHostname}
		if errMsg, ok := validation.Struct(th); !ok {
			return gen.PutNetworkDNS400JSONResponse{Error: &errMsg}, nil
		}
	}

	var servers []string
	if request.Body.Servers != nil {
		servers = *request.Body.Servers
	}

	var searchDomains []string
	if request.Body.SearchDomains != nil {
		searchDomains = *request.Body.SearchDomains
	}

	interfaceName := request.Body.InterfaceName

	hostname := job.AnyHost
	if request.Params.TargetHostname != nil {
		hostname = *request.Params.TargetHostname
	}

	n.logger.Debug("dns put",
		slog.String("interface", interfaceName),
		slog.Int("servers", len(servers)),
		slog.String("target", hostname),
	)

	if job.IsBroadcastTarget(hostname) {
		return n.putNetworkDNSBroadcast(ctx, hostname, servers, searchDomains, interfaceName)
	}

	jobID, workerHostname, err := n.JobClient.ModifyNetworkDNS(
		ctx,
		hostname,
		servers,
		searchDomains,
		interfaceName,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PutNetworkDNS500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.PutNetworkDNS202JSONResponse{
		JobId: &jobUUID,
		Results: []gen.DNSUpdateResultItem{
			{
				Hostname: workerHostname,
				Status:   gen.Ok,
			},
		},
	}, nil
}

// putNetworkDNSBroadcast handles broadcast targets (_all or label) for DNS modification.
func (n Network) putNetworkDNSBroadcast(
	ctx context.Context,
	target string,
	servers []string,
	searchDomains []string,
	interfaceName string,
) (gen.PutNetworkDNSResponseObject, error) {
	jobID, results, err := n.JobClient.ModifyNetworkDNSBroadcast(
		ctx,
		target,
		servers,
		searchDomains,
		interfaceName,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PutNetworkDNS500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	var responses []gen.DNSUpdateResultItem
	for host, hostErr := range results {
		item := gen.DNSUpdateResultItem{
			Hostname: host,
			Status:   gen.Ok,
		}
		if hostErr != nil {
			item.Status = gen.Failed
			errStr := fmt.Sprintf("%v", hostErr)
			item.Error = &errStr
		}
		responses = append(responses, item)
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.PutNetworkDNS202JSONResponse{
		JobId:   &jobUUID,
		Results: responses,
	}, nil
}

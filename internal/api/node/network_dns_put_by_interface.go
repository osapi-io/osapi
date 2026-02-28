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
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/api/node/gen"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/validation"
)

// PutNodeNetworkDNS put the node network dns API endpoint.
func (s *Node) PutNodeNetworkDNS(
	ctx context.Context,
	request gen.PutNodeNetworkDNSRequestObject,
) (gen.PutNodeNetworkDNSResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.PutNodeNetworkDNS400JSONResponse{
			Error: &errMsg,
		}, nil
	}

	if errMsg, ok := validation.Struct(request.Body); !ok {
		return gen.PutNodeNetworkDNS400JSONResponse{
			Error: &errMsg,
		}, nil
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

	hostname := request.Hostname

	s.logger.Debug("dns put",
		slog.String("interface", interfaceName),
		slog.Int("servers", len(servers)),
		slog.String("target", hostname),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.putNodeNetworkDNSBroadcast(ctx, hostname, servers, searchDomains, interfaceName)
	}

	jobID, agentHostname, changed, err := s.JobClient.ModifyNetworkDNS(
		ctx,
		hostname,
		servers,
		searchDomains,
		interfaceName,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PutNodeNetworkDNS500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.PutNodeNetworkDNS202JSONResponse{
		JobId: &jobUUID,
		Results: []gen.DNSUpdateResultItem{
			{
				Hostname: agentHostname,
				Status:   gen.Ok,
				Changed:  &changed,
			},
		},
	}, nil
}

// putNodeNetworkDNSBroadcast handles broadcast targets (_all or label) for DNS modification.
func (s *Node) putNodeNetworkDNSBroadcast(
	ctx context.Context,
	target string,
	servers []string,
	searchDomains []string,
	interfaceName string,
) (gen.PutNodeNetworkDNSResponseObject, error) {
	jobID, results, changedMap, err := s.JobClient.ModifyNetworkDNSBroadcast(
		ctx,
		target,
		servers,
		searchDomains,
		interfaceName,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PutNodeNetworkDNS500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	var responses []gen.DNSUpdateResultItem
	for host, hostErr := range results {
		c := changedMap[host]
		item := gen.DNSUpdateResultItem{
			Hostname: host,
			Status:   gen.Ok,
			Changed:  &c,
		}
		if hostErr != nil {
			item.Status = gen.Failed
			errStr := fmt.Sprintf("%v", hostErr)
			item.Error = &errStr
		}
		responses = append(responses, item)
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.PutNodeNetworkDNS202JSONResponse{
		JobId:   &jobUUID,
		Results: responses,
	}, nil
}

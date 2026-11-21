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

package client

import (
	"context"
	"fmt"

	"github.com/osapi-io/osapi/pkg/sdk/client/gen"
)

// DNSService reads and sets the DNS resolvers of a host's interfaces.
type DNSService struct {
	client *gen.ClientWithResponses
}

// Get is the DNS resolvers of an interface on the target.
func (s *DNSService) Get(
	ctx context.Context,
	target string,
	interfaceName string,
) (*Response[Collection[DNSConfig]], error) {
	resp, err := s.client.GetNodeNetworkDnsByKeyWithResponse(ctx, target, interfaceName)
	if err != nil {
		return nil, fmt.Errorf("get dns: %w", err)
	}
	if err := checkError(
		resp.StatusCode(),
		resp.JSON400,
		resp.JSON401,
		resp.JSON403,
		resp.JSON500,
	); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{
			APIError{StatusCode: resp.StatusCode(), Message: "nil response body"},
		}
	}

	return newResponse(dnsConfigsFromGen(resp.JSON200), resp.Body, resp.HTTPResponse), nil
}

// Update sets the DNS resolvers of an interface on the target. An empty
// list keeps the interface's own servers or domains.
func (s *DNSService) Update(
	ctx context.Context,
	target string,
	interfaceName string,
	servers []string,
	searchDomains []string,
	overrideDHCP bool,
) (*Response[Collection[DNSUpdateResult]], error) {
	body := gen.NetworkDnsUpdateRequest{}
	if len(servers) > 0 {
		body.Servers = &servers
	}
	if len(searchDomains) > 0 {
		body.SearchDomains = &searchDomains
	}
	if overrideDHCP {
		body.OverrideDhcp = &overrideDHCP
	}
	resp, err := s.client.PutNodeNetworkDnsWithResponse(ctx, target, interfaceName, body)
	if err != nil {
		return nil, fmt.Errorf("update dns: %w", err)
	}
	if err := checkError(
		resp.StatusCode(),
		resp.JSON400,
		resp.JSON401,
		resp.JSON403,
		resp.JSON500,
	); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{
			APIError{StatusCode: resp.StatusCode(), Message: "nil response body"},
		}
	}

	return newResponse(Collection[DNSUpdateResult]{
		Results: dnsResults(resp.JSON200.Results), JobID: jobIDFromGen(resp.JSON200.JobId),
	}, resp.Body, resp.HTTPResponse), nil
}

// Delete drops the DNS resolvers osapi set on an interface on the target.
func (s *DNSService) Delete(
	ctx context.Context,
	target string,
	interfaceName string,
) (*Response[Collection[DNSDeleteResult]], error) {
	resp, err := s.client.DeleteNodeNetworkDnsWithResponse(ctx, target, interfaceName)
	if err != nil {
		return nil, fmt.Errorf("delete dns: %w", err)
	}
	if err := checkError(
		resp.StatusCode(),
		resp.JSON400,
		resp.JSON401,
		resp.JSON403,
		resp.JSON500,
	); err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{
			APIError{StatusCode: resp.StatusCode(), Message: "nil response body"},
		}
	}
	rows := dnsResults(resp.JSON200.Results)
	results := make([]DNSDeleteResult, 0, len(rows))
	for _, r := range rows {
		results = append(results, DNSDeleteResult(r))
	}

	return newResponse(Collection[DNSDeleteResult]{
		Results: results, JobID: jobIDFromGen(resp.JSON200.JobId),
	}, resp.Body, resp.HTTPResponse), nil
}

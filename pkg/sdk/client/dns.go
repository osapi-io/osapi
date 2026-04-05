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

	"github.com/retr0h/osapi/pkg/sdk/client/gen"
)

// DNSService provides DNS configuration query and update operations.
type DNSService struct {
	client *gen.ClientWithResponses
}

// Get retrieves DNS configuration for a network interface on the
// target host.
func (s *DNSService) Get(
	ctx context.Context,
	target string,
	interfaceName string,
) (*Response[Collection[DNSConfig]], error) {
	resp, err := s.client.GetNodeNetworkDNSByInterfaceWithResponse(ctx, target, interfaceName)
	if err != nil {
		return nil, fmt.Errorf("get dns: %w", err)
	}

	if err := checkError(resp.StatusCode(), resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON500); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(dnsConfigCollectionFromGen(resp.JSON200), resp.Body), nil
}

// Update updates DNS configuration for a network interface on the
// target host.
func (s *DNSService) Update(
	ctx context.Context,
	target string,
	interfaceName string,
	servers []string,
	searchDomains []string,
	overrideDHCP bool,
) (*Response[Collection[DNSUpdateResult]], error) {
	body := gen.DNSConfigUpdateRequest{
		InterfaceName: interfaceName,
	}

	if len(servers) > 0 {
		body.Servers = &servers
	}

	if len(searchDomains) > 0 {
		body.SearchDomains = &searchDomains
	}

	if overrideDHCP {
		body.OverrideDhcp = &overrideDHCP
	}

	resp, err := s.client.PutNodeNetworkDNSWithResponse(ctx, target, body)
	if err != nil {
		return nil, fmt.Errorf("update dns: %w", err)
	}

	if err := checkError(resp.StatusCode(), resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON500); err != nil {
		return nil, err
	}

	if resp.JSON202 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(dnsUpdateCollectionFromGen(resp.JSON202), resp.Body), nil
}

// Delete removes DNS configuration for a network interface on the
// target host.
func (s *DNSService) Delete(
	ctx context.Context,
	target string,
	interfaceName string,
) (*Response[Collection[DNSDeleteResult]], error) {
	body := gen.DNSDeleteRequest{
		InterfaceName: interfaceName,
	}

	resp, err := s.client.DeleteNodeNetworkDNSWithResponse(ctx, target, body)
	if err != nil {
		return nil, fmt.Errorf("delete dns: %w", err)
	}

	if err := checkError(resp.StatusCode(), resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON500); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(dnsDeleteCollectionFromGen(resp.JSON200), resp.Body), nil
}

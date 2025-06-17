// Copyright (c) 2025 John Dewey

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
	"encoding/json"
	"fmt"

	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider/network/ping"
)

// ModifyNetworkDNS modifies DNS configuration on a specific hostname.
func (c *Client) ModifyNetworkDNS(
	ctx context.Context,
	hostname string,
	servers []string,
	searchDomains []string,
	iface string,
) error {
	data, _ := json.Marshal(map[string]interface{}{
		"servers":        servers,
		"search_domains": searchDomains,
		"interface":      iface,
	})
	req := &job.Request{
		Type:      job.TypeModify,
		Category:  "network",
		Operation: "dns.set",
		Data:      json.RawMessage(data),
	}

	subject := job.BuildModifySubject(hostname)
	resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == "failed" {
		return fmt.Errorf("job failed: %s", resp.Error)
	}

	return nil
}

// ModifyNetworkDNSAny modifies DNS configuration on any available host.
func (c *Client) ModifyNetworkDNSAny(
	ctx context.Context,
	servers []string,
	searchDomains []string,
	iface string,
) error {
	return c.ModifyNetworkDNS(ctx, job.AnyHost, servers, searchDomains, iface)
}

// ModifyNetworkPing pings a host from a specific hostname.
func (c *Client) ModifyNetworkPing(
	ctx context.Context,
	hostname string,
	address string,
) (*ping.Result, error) {
	data, _ := json.Marshal(map[string]interface{}{
		"address": address,
	})
	req := &job.Request{
		Type:      job.TypeModify,
		Category:  "network",
		Operation: "ping.do",
		Data:      json.RawMessage(data),
	}

	subject := job.BuildModifySubject(hostname)
	resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return nil, fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == "failed" {
		return nil, fmt.Errorf("job failed: %s", resp.Error)
	}

	var result ping.Result
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ping response: %w", err)
	}

	return &result, nil
}

// ModifyNetworkPingAny pings a host from any available hostname.
func (c *Client) ModifyNetworkPingAny(
	ctx context.Context,
	address string,
) (*ping.Result, error) {
	return c.ModifyNetworkPing(ctx, job.AnyHost, address)
}

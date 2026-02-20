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
)

// ModifyNetworkDNS modifies DNS configuration on a specific hostname.
func (c *Client) ModifyNetworkDNS(
	ctx context.Context,
	hostname string,
	servers []string,
	searchDomains []string,
	iface string,
) (string, string, error) {
	data, _ := json.Marshal(map[string]interface{}{
		"servers":        servers,
		"search_domains": searchDomains,
		"interface":      iface,
	})
	req := &job.Request{
		Type:      job.TypeModify,
		Category:  "network",
		Operation: "dns.update",
		Data:      json.RawMessage(data),
	}

	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, hostname)
	jobID, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return "", "", fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == "failed" {
		return "", "", fmt.Errorf("job failed: %s", resp.Error)
	}

	return jobID, resp.Hostname, nil
}

// ModifyNetworkDNSAny modifies DNS configuration on any available host.
func (c *Client) ModifyNetworkDNSAny(
	ctx context.Context,
	servers []string,
	searchDomains []string,
	iface string,
) (string, string, error) {
	return c.ModifyNetworkDNS(ctx, job.AnyHost, servers, searchDomains, iface)
}

// ModifyNetworkDNSBroadcast modifies DNS configuration on a broadcast target
// (_all or a label target like role:web).
func (c *Client) ModifyNetworkDNSBroadcast(
	ctx context.Context,
	target string,
	servers []string,
	searchDomains []string,
	iface string,
) (string, map[string]error, error) {
	data, _ := json.Marshal(map[string]interface{}{
		"servers":        servers,
		"search_domains": searchDomains,
		"interface":      iface,
	})
	req := &job.Request{
		Type:      job.TypeModify,
		Category:  "network",
		Operation: "dns.update",
		Data:      json.RawMessage(data),
	}

	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, target)
	jobID, responses, err := c.publishAndCollect(ctx, subject, req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to collect broadcast responses: %w", err)
	}

	results := make(map[string]error)
	for hostname, resp := range responses {
		if resp.Status == "failed" {
			results[hostname] = fmt.Errorf("job failed: %s", resp.Error)
		} else {
			results[hostname] = nil
		}
	}

	return jobID, results, nil
}

// ModifyNetworkDNSAll modifies DNS configuration on all hosts.
func (c *Client) ModifyNetworkDNSAll(
	ctx context.Context,
	servers []string,
	searchDomains []string,
	iface string,
) (string, map[string]error, error) {
	return c.ModifyNetworkDNSBroadcast(ctx, job.BroadcastHost, servers, searchDomains, iface)
}

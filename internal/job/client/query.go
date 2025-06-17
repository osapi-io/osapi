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
	"github.com/retr0h/osapi/internal/provider/network/dns"
)

// QuerySystemStatus queries system status from a specific hostname.
func (c *Client) QuerySystemStatus(
	ctx context.Context,
	hostname string,
) (*job.SystemStatusResponse, error) {
	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  "system",
		Operation: "status.get",
		Data:      json.RawMessage(`{}`),
	}

	subject := job.BuildQuerySubject(hostname)
	resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return nil, fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == "failed" {
		return nil, fmt.Errorf("job failed: %s", resp.Error)
	}

	var result job.SystemStatusResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal status response: %w", err)
	}

	return &result, nil
}

// QuerySystemHostname queries hostname from a specific hostname.
func (c *Client) QuerySystemHostname(
	ctx context.Context,
	hostname string,
) (string, error) {
	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  "system",
		Operation: "hostname.get",
		Data:      json.RawMessage(`{}`),
	}

	subject := job.BuildQuerySubject(hostname)
	resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return "", fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == "failed" {
		return "", fmt.Errorf("job failed: %s", resp.Error)
	}

	var result struct {
		Hostname string `json:"hostname"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal hostname response: %w", err)
	}

	return result.Hostname, nil
}

// QueryNetworkDNS queries DNS configuration from a specific hostname.
func (c *Client) QueryNetworkDNS(
	ctx context.Context,
	hostname string,
	iface string,
) (*dns.Config, error) {
	data, _ := json.Marshal(map[string]interface{}{
		"interface": iface,
	})
	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  "network",
		Operation: "dns.get",
		Data:      json.RawMessage(data),
	}

	subject := job.BuildQuerySubject(hostname)
	resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return nil, fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == "failed" {
		return nil, fmt.Errorf("job failed: %s", resp.Error)
	}

	var result dns.Config
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal DNS response: %w", err)
	}

	return &result, nil
}

// QuerySystemStatusAny queries system status from any available host.
func (c *Client) QuerySystemStatusAny(
	ctx context.Context,
) (*job.SystemStatusResponse, error) {
	return c.QuerySystemStatus(ctx, job.AnyHost)
}

// QuerySystemStatusAll queries system status from all hosts.
func (c *Client) QuerySystemStatusAll(
	_ context.Context,
) ([]*job.SystemStatusResponse, error) {
	return nil, fmt.Errorf("broadcast queries not yet implemented")
}

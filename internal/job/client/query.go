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

// QuerySystemStatus queries the system status of a specific host.
func (c *Client) QuerySystemStatus(
	ctx context.Context,
	hostname string,
) (*job.SystemStatusResponse, error) {
	subject := job.BuildQuerySubject(
		hostname,
		job.SubjectCategorySystem,
		job.SystemOperationStatus,
	)

	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  job.SubjectCategorySystem,
		Operation: job.SystemOperationStatus,
		Data:      json.RawMessage(`{}`), // No parameters needed for status
	}

	resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return nil, err
	}

	if resp.Status == job.StatusFailed {
		return nil, fmt.Errorf("job failed: %s", resp.Error)
	}

	// Unmarshal the response data
	var statusResp job.SystemStatusResponse
	if err := json.Unmarshal(resp.Data, &statusResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal status response: %w", err)
	}

	return &statusResp, nil
}

// QuerySystemHostname queries the hostname of a specific host.
func (c *Client) QuerySystemHostname(
	ctx context.Context,
	hostname string,
) (string, error) {
	subject := job.BuildQuerySubject(
		hostname,
		job.SubjectCategorySystem,
		job.SystemOperationHostname,
	)

	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  job.SubjectCategorySystem,
		Operation: job.SystemOperationHostname,
		Data:      json.RawMessage(`{}`),
	}

	resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return "", err
	}

	if resp.Status == job.StatusFailed {
		return "", fmt.Errorf("job failed: %s", resp.Error)
	}

	// For hostname, we expect a simple string response
	var hostnameResp struct {
		Hostname string `json:"hostname"`
	}
	if err := json.Unmarshal(resp.Data, &hostnameResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal hostname response: %w", err)
	}

	return hostnameResp.Hostname, nil
}

// QueryNetworkDNS queries the DNS configuration of a specific host and interface.
func (c *Client) QueryNetworkDNS(
	ctx context.Context,
	hostname string,
	iface string,
) (*dns.Config, error) {
	subject := job.BuildQuerySubject(
		hostname,
		job.SubjectCategoryNetwork,
		job.NetworkOperationDNS,
	)

	reqData := map[string]interface{}{
		"interface": iface,
	}
	data, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal DNS request: %w", err)
	}

	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  job.SubjectCategoryNetwork,
		Operation: job.NetworkOperationDNS,
		Data:      data,
	}

	resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return nil, err
	}

	if resp.Status == job.StatusFailed {
		return nil, fmt.Errorf("job failed: %s", resp.Error)
	}

	// Unmarshal the response data
	var dnsResp dns.Config
	if err := json.Unmarshal(resp.Data, &dnsResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal DNS response: %w", err)
	}

	return &dnsResp, nil
}

// QuerySystemStatusAny queries system status from any available host.
func (c *Client) QuerySystemStatusAny(
	ctx context.Context,
) (*job.SystemStatusResponse, error) {
	return c.QuerySystemStatus(ctx, job.AnyHost)
}

// QuerySystemStatusAll queries system status from all hosts.
// Returns a map of hostname to status response.
func (c *Client) QuerySystemStatusAll(
	_ context.Context,
) (map[string]interface{}, error) {
	// This would require a different implementation that collects multiple responses
	// For now, return an error indicating it's not implemented
	return nil, fmt.Errorf("broadcast queries not yet implemented")
}

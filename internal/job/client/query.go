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
	"github.com/retr0h/osapi/internal/provider/network/ping"
)

// QuerySystemStatus queries system status from a specific hostname.
func (c *Client) QuerySystemStatus(
	ctx context.Context,
	hostname string,
) (string, *job.SystemStatusResponse, error) {
	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  "system",
		Operation: "status.get",
		Data:      json.RawMessage(`{}`),
	}

	subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, hostname)
	jobID, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == "failed" {
		return "", nil, fmt.Errorf("job failed: %s", resp.Error)
	}

	var result job.SystemStatusResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return "", nil, fmt.Errorf("failed to unmarshal status response: %w", err)
	}

	return jobID, &result, nil
}

// QuerySystemHostname queries hostname from a specific hostname.
func (c *Client) QuerySystemHostname(
	ctx context.Context,
	hostname string,
) (string, string, *job.WorkerInfo, error) {
	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  "system",
		Operation: "hostname.get",
		Data:      json.RawMessage(`{}`),
	}

	subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, hostname)
	jobID, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == "failed" {
		return "", "", nil, fmt.Errorf("job failed: %s", resp.Error)
	}

	var result struct {
		Hostname string            `json:"hostname"`
		Labels   map[string]string `json:"labels,omitempty"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return "", "", nil, fmt.Errorf("failed to unmarshal hostname response: %w", err)
	}

	worker := &job.WorkerInfo{
		Hostname: resp.Hostname,
		Labels:   result.Labels,
	}

	return jobID, result.Hostname, worker, nil
}

// QueryNetworkDNS queries DNS configuration from a specific hostname.
func (c *Client) QueryNetworkDNS(
	ctx context.Context,
	hostname string,
	iface string,
) (string, *dns.Config, string, error) {
	data, _ := json.Marshal(map[string]interface{}{
		"interface": iface,
	})
	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  "network",
		Operation: "dns.get",
		Data:      json.RawMessage(data),
	}

	subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, hostname)
	jobID, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return "", nil, "", fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == "failed" {
		return "", nil, "", fmt.Errorf("job failed: %s", resp.Error)
	}

	var result dns.Config
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return "", nil, "", fmt.Errorf("failed to unmarshal DNS response: %w", err)
	}

	return jobID, &result, resp.Hostname, nil
}

// QuerySystemStatusAny queries system status from any available host.
func (c *Client) QuerySystemStatusAny(
	ctx context.Context,
) (string, *job.SystemStatusResponse, error) {
	return c.QuerySystemStatus(ctx, job.AnyHost)
}

// QuerySystemStatusBroadcast queries system status from a broadcast target
// (_all or a label target like role:web).
func (c *Client) QuerySystemStatusBroadcast(
	ctx context.Context,
	target string,
) (string, []*job.SystemStatusResponse, map[string]string, error) {
	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  "system",
		Operation: "status.get",
		Data:      json.RawMessage(`{}`),
	}

	subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, target)
	jobID, responses, err := c.publishAndCollect(ctx, subject, req)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to collect broadcast responses: %w", err)
	}

	var results []*job.SystemStatusResponse
	errs := make(map[string]string)
	for hostname, resp := range responses {
		if resp.Status == "failed" {
			errs[hostname] = resp.Error
			continue
		}

		var result job.SystemStatusResponse
		if err := json.Unmarshal(resp.Data, &result); err != nil {
			continue
		}

		if result.Hostname == "" {
			result.Hostname = hostname
		}

		results = append(results, &result)
	}

	return jobID, results, errs, nil
}

// QuerySystemStatusAll queries system status from all hosts.
func (c *Client) QuerySystemStatusAll(
	ctx context.Context,
) (string, []*job.SystemStatusResponse, map[string]string, error) {
	return c.QuerySystemStatusBroadcast(ctx, job.BroadcastHost)
}

// QueryNetworkPing pings a host from a specific hostname.
func (c *Client) QueryNetworkPing(
	ctx context.Context,
	hostname string,
	address string,
) (string, *ping.Result, string, error) {
	data, _ := json.Marshal(map[string]interface{}{
		"address": address,
	})
	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  "network",
		Operation: "ping.do",
		Data:      json.RawMessage(data),
	}

	subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, hostname)
	jobID, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return "", nil, "", fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == "failed" {
		return "", nil, "", fmt.Errorf("job failed: %s", resp.Error)
	}

	var result ping.Result
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return "", nil, "", fmt.Errorf("failed to unmarshal ping response: %w", err)
	}

	return jobID, &result, resp.Hostname, nil
}

// QueryNetworkPingAny pings a host from any available hostname.
func (c *Client) QueryNetworkPingAny(
	ctx context.Context,
	address string,
) (string, *ping.Result, string, error) {
	return c.QueryNetworkPing(ctx, job.AnyHost, address)
}

// QuerySystemHostnameBroadcast queries hostname from a broadcast target
// (_all or a label target like role:web).
func (c *Client) QuerySystemHostnameBroadcast(
	ctx context.Context,
	target string,
) (string, map[string]*job.WorkerInfo, map[string]string, error) {
	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  "system",
		Operation: "hostname.get",
		Data:      json.RawMessage(`{}`),
	}

	subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, target)
	jobID, responses, err := c.publishAndCollect(ctx, subject, req)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to collect broadcast responses: %w", err)
	}

	results := make(map[string]*job.WorkerInfo)
	errs := make(map[string]string)
	for hostname, resp := range responses {
		if resp.Status == "failed" {
			errs[hostname] = resp.Error
			continue
		}

		var result struct {
			Hostname string            `json:"hostname"`
			Labels   map[string]string `json:"labels,omitempty"`
		}
		if err := json.Unmarshal(resp.Data, &result); err != nil {
			continue
		}

		results[hostname] = &job.WorkerInfo{
			Hostname: result.Hostname,
			Labels:   result.Labels,
		}
	}

	return jobID, results, errs, nil
}

// QuerySystemHostnameAll queries hostname from all hosts.
func (c *Client) QuerySystemHostnameAll(
	ctx context.Context,
) (string, map[string]*job.WorkerInfo, map[string]string, error) {
	return c.QuerySystemHostnameBroadcast(ctx, job.BroadcastHost)
}

// QueryNetworkDNSBroadcast queries DNS configuration from a broadcast target
// (_all or a label target like role:web).
func (c *Client) QueryNetworkDNSBroadcast(
	ctx context.Context,
	target string,
	iface string,
) (string, map[string]*dns.Config, map[string]string, error) {
	data, _ := json.Marshal(map[string]interface{}{
		"interface": iface,
	})
	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  "network",
		Operation: "dns.get",
		Data:      json.RawMessage(data),
	}

	subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, target)
	jobID, responses, err := c.publishAndCollect(ctx, subject, req)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to collect broadcast responses: %w", err)
	}

	results := make(map[string]*dns.Config)
	errs := make(map[string]string)
	for hostname, resp := range responses {
		if resp.Status == "failed" {
			errs[hostname] = resp.Error
			continue
		}

		var result dns.Config
		if err := json.Unmarshal(resp.Data, &result); err != nil {
			continue
		}

		results[hostname] = &result
	}

	return jobID, results, errs, nil
}

// QueryNetworkDNSAll queries DNS configuration from all hosts.
func (c *Client) QueryNetworkDNSAll(
	ctx context.Context,
	iface string,
) (string, map[string]*dns.Config, map[string]string, error) {
	return c.QueryNetworkDNSBroadcast(ctx, job.BroadcastHost, iface)
}

// QueryNetworkPingBroadcast pings a host from a broadcast target
// (_all or a label target like role:web).
func (c *Client) QueryNetworkPingBroadcast(
	ctx context.Context,
	target string,
	address string,
) (string, map[string]*ping.Result, map[string]string, error) {
	data, _ := json.Marshal(map[string]interface{}{
		"address": address,
	})
	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  "network",
		Operation: "ping.do",
		Data:      json.RawMessage(data),
	}

	subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, target)
	jobID, responses, err := c.publishAndCollect(ctx, subject, req)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to collect broadcast responses: %w", err)
	}

	results := make(map[string]*ping.Result)
	errs := make(map[string]string)
	for hostname, resp := range responses {
		if resp.Status == "failed" {
			errs[hostname] = resp.Error
			continue
		}

		var result ping.Result
		if err := json.Unmarshal(resp.Data, &result); err != nil {
			continue
		}

		results[hostname] = &result
	}

	return jobID, results, errs, nil
}

// QueryNetworkPingAll pings a host from all hosts.
func (c *Client) QueryNetworkPingAll(
	ctx context.Context,
	address string,
) (string, map[string]*ping.Result, map[string]string, error) {
	return c.QueryNetworkPingBroadcast(ctx, job.BroadcastHost, address)
}

// ListWorkers reads the worker registry KV bucket and returns all registered
// workers. Workers register via heartbeat, so only live workers appear.
func (c *Client) ListWorkers(
	ctx context.Context,
) ([]job.WorkerInfo, error) {
	if c.registryKV == nil {
		return nil, fmt.Errorf("worker registry not configured")
	}

	keys, err := c.registryKV.Keys(ctx)
	if err != nil {
		// Keys returns jetstream.ErrNoKeysFound when the bucket is empty.
		if err.Error() == "nats: no keys found" {
			return []job.WorkerInfo{}, nil
		}
		return nil, fmt.Errorf("failed to list registry keys: %w", err)
	}

	workers := make([]job.WorkerInfo, 0, len(keys))
	for _, key := range keys {
		entry, err := c.registryKV.Get(ctx, key)
		if err != nil {
			continue
		}

		var reg job.WorkerRegistration
		if err := json.Unmarshal(entry.Value(), &reg); err != nil {
			continue
		}

		workers = append(workers, job.WorkerInfo{
			Hostname: reg.Hostname,
			Labels:   reg.Labels,
		})
	}

	return workers, nil
}

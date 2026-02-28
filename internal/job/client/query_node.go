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
	"encoding/json"
	"fmt"

	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider/node/host"
	"github.com/retr0h/osapi/internal/provider/node/load"
	"github.com/retr0h/osapi/internal/provider/node/mem"
)

// QueryNodeDisk queries disk usage from a specific hostname.
func (c *Client) QueryNodeDisk(
	ctx context.Context,
	hostname string,
) (string, *job.NodeDiskResponse, string, error) {
	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  "node",
		Operation: "disk.get",
		Data:      json.RawMessage(`{}`),
	}

	subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, hostname)
	jobID, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return "", nil, "", fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == "failed" {
		return "", nil, "", fmt.Errorf("job failed: %s", resp.Error)
	}

	var result job.NodeDiskResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return "", nil, "", fmt.Errorf("failed to unmarshal disk response: %w", err)
	}

	return jobID, &result, resp.Hostname, nil
}

// QueryNodeDiskBroadcast queries disk usage from a broadcast target.
func (c *Client) QueryNodeDiskBroadcast(
	ctx context.Context,
	target string,
) (string, map[string]*job.NodeDiskResponse, map[string]string, error) {
	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  "node",
		Operation: "disk.get",
		Data:      json.RawMessage(`{}`),
	}

	subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, target)
	jobID, responses, err := c.publishAndCollect(ctx, subject, req)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to collect broadcast responses: %w", err)
	}

	results := make(map[string]*job.NodeDiskResponse)
	errs := make(map[string]string)
	for hostname, resp := range responses {
		if resp.Status == "failed" {
			errs[hostname] = resp.Error
			continue
		}

		var result job.NodeDiskResponse
		if err := json.Unmarshal(resp.Data, &result); err != nil {
			continue
		}

		results[hostname] = &result
	}

	return jobID, results, errs, nil
}

// QueryNodeMemory queries memory stats from a specific hostname.
func (c *Client) QueryNodeMemory(
	ctx context.Context,
	hostname string,
) (string, *mem.Stats, string, error) {
	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  "node",
		Operation: "memory.get",
		Data:      json.RawMessage(`{}`),
	}

	subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, hostname)
	jobID, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return "", nil, "", fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == "failed" {
		return "", nil, "", fmt.Errorf("job failed: %s", resp.Error)
	}

	var result mem.Stats
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return "", nil, "", fmt.Errorf("failed to unmarshal memory response: %w", err)
	}

	return jobID, &result, resp.Hostname, nil
}

// QueryNodeMemoryBroadcast queries memory stats from a broadcast target.
func (c *Client) QueryNodeMemoryBroadcast(
	ctx context.Context,
	target string,
) (string, map[string]*mem.Stats, map[string]string, error) {
	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  "node",
		Operation: "memory.get",
		Data:      json.RawMessage(`{}`),
	}

	subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, target)
	jobID, responses, err := c.publishAndCollect(ctx, subject, req)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to collect broadcast responses: %w", err)
	}

	results := make(map[string]*mem.Stats)
	errs := make(map[string]string)
	for hostname, resp := range responses {
		if resp.Status == "failed" {
			errs[hostname] = resp.Error
			continue
		}

		var result mem.Stats
		if err := json.Unmarshal(resp.Data, &result); err != nil {
			continue
		}

		results[hostname] = &result
	}

	return jobID, results, errs, nil
}

// QueryNodeLoad queries load averages from a specific hostname.
func (c *Client) QueryNodeLoad(
	ctx context.Context,
	hostname string,
) (string, *load.AverageStats, string, error) {
	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  "node",
		Operation: "load.get",
		Data:      json.RawMessage(`{}`),
	}

	subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, hostname)
	jobID, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return "", nil, "", fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == "failed" {
		return "", nil, "", fmt.Errorf("job failed: %s", resp.Error)
	}

	var result load.AverageStats
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return "", nil, "", fmt.Errorf("failed to unmarshal load response: %w", err)
	}

	return jobID, &result, resp.Hostname, nil
}

// QueryNodeLoadBroadcast queries load averages from a broadcast target.
func (c *Client) QueryNodeLoadBroadcast(
	ctx context.Context,
	target string,
) (string, map[string]*load.AverageStats, map[string]string, error) {
	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  "node",
		Operation: "load.get",
		Data:      json.RawMessage(`{}`),
	}

	subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, target)
	jobID, responses, err := c.publishAndCollect(ctx, subject, req)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to collect broadcast responses: %w", err)
	}

	results := make(map[string]*load.AverageStats)
	errs := make(map[string]string)
	for hostname, resp := range responses {
		if resp.Status == "failed" {
			errs[hostname] = resp.Error
			continue
		}

		var result load.AverageStats
		if err := json.Unmarshal(resp.Data, &result); err != nil {
			continue
		}

		results[hostname] = &result
	}

	return jobID, results, errs, nil
}

// QueryNodeOS queries OS information from a specific hostname.
func (c *Client) QueryNodeOS(
	ctx context.Context,
	hostname string,
) (string, *host.OSInfo, string, error) {
	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  "node",
		Operation: "os.get",
		Data:      json.RawMessage(`{}`),
	}

	subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, hostname)
	jobID, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return "", nil, "", fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == "failed" {
		return "", nil, "", fmt.Errorf("job failed: %s", resp.Error)
	}

	var result host.OSInfo
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return "", nil, "", fmt.Errorf("failed to unmarshal OS info response: %w", err)
	}

	return jobID, &result, resp.Hostname, nil
}

// QueryNodeOSBroadcast queries OS information from a broadcast target.
func (c *Client) QueryNodeOSBroadcast(
	ctx context.Context,
	target string,
) (string, map[string]*host.OSInfo, map[string]string, error) {
	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  "node",
		Operation: "os.get",
		Data:      json.RawMessage(`{}`),
	}

	subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, target)
	jobID, responses, err := c.publishAndCollect(ctx, subject, req)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to collect broadcast responses: %w", err)
	}

	results := make(map[string]*host.OSInfo)
	errs := make(map[string]string)
	for hostname, resp := range responses {
		if resp.Status == "failed" {
			errs[hostname] = resp.Error
			continue
		}

		var result host.OSInfo
		if err := json.Unmarshal(resp.Data, &result); err != nil {
			continue
		}

		results[hostname] = &result
	}

	return jobID, results, errs, nil
}

// QueryNodeUptime queries system uptime from a specific hostname.
func (c *Client) QueryNodeUptime(
	ctx context.Context,
	hostname string,
) (string, *job.NodeUptimeResponse, string, error) {
	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  "node",
		Operation: "uptime.get",
		Data:      json.RawMessage(`{}`),
	}

	subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, hostname)
	jobID, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return "", nil, "", fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == "failed" {
		return "", nil, "", fmt.Errorf("job failed: %s", resp.Error)
	}

	var result job.NodeUptimeResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return "", nil, "", fmt.Errorf("failed to unmarshal uptime response: %w", err)
	}

	return jobID, &result, resp.Hostname, nil
}

// QueryNodeUptimeBroadcast queries system uptime from a broadcast target.
func (c *Client) QueryNodeUptimeBroadcast(
	ctx context.Context,
	target string,
) (string, map[string]*job.NodeUptimeResponse, map[string]string, error) {
	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  "node",
		Operation: "uptime.get",
		Data:      json.RawMessage(`{}`),
	}

	subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, target)
	jobID, responses, err := c.publishAndCollect(ctx, subject, req)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to collect broadcast responses: %w", err)
	}

	results := make(map[string]*job.NodeUptimeResponse)
	errs := make(map[string]string)
	for hostname, resp := range responses {
		if resp.Status == "failed" {
			errs[hostname] = resp.Error
			continue
		}

		var result job.NodeUptimeResponse
		if err := json.Unmarshal(resp.Data, &result); err != nil {
			continue
		}

		results[hostname] = &result
	}

	return jobID, results, errs, nil
}

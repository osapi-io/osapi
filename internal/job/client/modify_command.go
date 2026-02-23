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
	"github.com/retr0h/osapi/internal/provider/command"
)

// ModifyCommandExec executes a command on a specific hostname.
func (c *Client) ModifyCommandExec(
	ctx context.Context,
	hostname string,
	cmdName string,
	args []string,
	cwd string,
	timeout int,
) (string, *command.Result, string, error) {
	data, _ := json.Marshal(job.CommandExecData{
		Command: cmdName,
		Args:    args,
		Cwd:     cwd,
		Timeout: timeout,
	})
	req := &job.Request{
		Type:      job.TypeModify,
		Category:  "command",
		Operation: "exec.execute",
		Data:      json.RawMessage(data),
	}

	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, hostname)
	jobID, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return "", nil, "", fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == "failed" {
		return "", nil, "", fmt.Errorf("job failed: %s", resp.Error)
	}

	var result command.Result
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return "", nil, "", fmt.Errorf("failed to unmarshal command result: %w", err)
	}

	return jobID, &result, resp.Hostname, nil
}

// ModifyCommandExecBroadcast executes a command on a broadcast target.
func (c *Client) ModifyCommandExecBroadcast(
	ctx context.Context,
	target string,
	cmdName string,
	args []string,
	cwd string,
	timeout int,
) (string, map[string]*command.Result, map[string]string, error) {
	data, _ := json.Marshal(job.CommandExecData{
		Command: cmdName,
		Args:    args,
		Cwd:     cwd,
		Timeout: timeout,
	})
	req := &job.Request{
		Type:      job.TypeModify,
		Category:  "command",
		Operation: "exec.execute",
		Data:      json.RawMessage(data),
	}

	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, target)
	jobID, responses, err := c.publishAndCollect(ctx, subject, req)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to collect broadcast responses: %w", err)
	}

	results := make(map[string]*command.Result)
	errs := make(map[string]string)
	for hostname, resp := range responses {
		if resp.Status == "failed" {
			errs[hostname] = resp.Error
		} else {
			var result command.Result
			if unmarshalErr := json.Unmarshal(resp.Data, &result); unmarshalErr != nil {
				errs[hostname] = fmt.Sprintf("failed to unmarshal result: %v", unmarshalErr)
			} else {
				results[hostname] = &result
			}
		}
	}

	return jobID, results, errs, nil
}

// ModifyCommandShell executes a shell command on a specific hostname.
func (c *Client) ModifyCommandShell(
	ctx context.Context,
	hostname string,
	cmdStr string,
	cwd string,
	timeout int,
) (string, *command.Result, string, error) {
	data, _ := json.Marshal(job.CommandShellData{
		Command: cmdStr,
		Cwd:     cwd,
		Timeout: timeout,
	})
	req := &job.Request{
		Type:      job.TypeModify,
		Category:  "command",
		Operation: "shell.execute",
		Data:      json.RawMessage(data),
	}

	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, hostname)
	jobID, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return "", nil, "", fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == "failed" {
		return "", nil, "", fmt.Errorf("job failed: %s", resp.Error)
	}

	var result command.Result
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return "", nil, "", fmt.Errorf("failed to unmarshal command result: %w", err)
	}

	return jobID, &result, resp.Hostname, nil
}

// ModifyCommandShellBroadcast executes a shell command on a broadcast target.
func (c *Client) ModifyCommandShellBroadcast(
	ctx context.Context,
	target string,
	cmdStr string,
	cwd string,
	timeout int,
) (string, map[string]*command.Result, map[string]string, error) {
	data, _ := json.Marshal(job.CommandShellData{
		Command: cmdStr,
		Cwd:     cwd,
		Timeout: timeout,
	})
	req := &job.Request{
		Type:      job.TypeModify,
		Category:  "command",
		Operation: "shell.execute",
		Data:      json.RawMessage(data),
	}

	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, target)
	jobID, responses, err := c.publishAndCollect(ctx, subject, req)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to collect broadcast responses: %w", err)
	}

	results := make(map[string]*command.Result)
	errs := make(map[string]string)
	for hostname, resp := range responses {
		if resp.Status == "failed" {
			errs[hostname] = resp.Error
		} else {
			var result command.Result
			if unmarshalErr := json.Unmarshal(resp.Data, &result); unmarshalErr != nil {
				errs[hostname] = fmt.Sprintf("failed to unmarshal result: %v", unmarshalErr)
			} else {
				results[hostname] = &result
			}
		}
	}

	return jobID, results, errs, nil
}

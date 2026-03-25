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
	"github.com/retr0h/osapi/internal/provider/file"
)

// ModifyFileDeployBroadcast deploys a file to a broadcast target.
func (c *Client) ModifyFileDeployBroadcast(
	ctx context.Context,
	target string,
	objectName string,
	path string,
	contentType string,
	mode string,
	owner string,
	group string,
	vars map[string]any,
) (string, map[string]bool, map[string]string, error) {
	data, _ := json.Marshal(file.DeployRequest{
		ObjectName:  objectName,
		Path:        path,
		Mode:        mode,
		Owner:       owner,
		Group:       group,
		ContentType: contentType,
		Vars:        vars,
	})

	req := &job.Request{
		Type:      job.TypeModify,
		Category:  "file",
		Operation: job.OperationFileDeployExecute,
		Data:      json.RawMessage(data),
	}

	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, target)
	jobID, responses, err := c.publishAndCollect(ctx, subject, target, req)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to collect broadcast responses: %w", err)
	}

	changed := make(map[string]bool)
	errs := make(map[string]string)
	for hostname, resp := range responses {
		if resp.Status == job.StatusFailed || resp.Status == job.StatusSkipped {
			errs[hostname] = resp.Error
		} else {
			changed[hostname] = resp.Changed != nil && *resp.Changed
		}
	}

	return jobID, changed, errs, nil
}

// ModifyFileUndeployBroadcast removes a deployed file from a broadcast target.
func (c *Client) ModifyFileUndeployBroadcast(
	ctx context.Context,
	target string,
	path string,
) (string, map[string]bool, map[string]string, error) {
	data, _ := json.Marshal(file.UndeployRequest{
		Path: path,
	})

	req := &job.Request{
		Type:      job.TypeModify,
		Category:  "file",
		Operation: job.OperationFileUndeployExecute,
		Data:      json.RawMessage(data),
	}

	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, target)
	jobID, responses, err := c.publishAndCollect(ctx, subject, target, req)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to collect broadcast responses: %w", err)
	}

	changed := make(map[string]bool)
	errs := make(map[string]string)
	for hostname, resp := range responses {
		if resp.Status == job.StatusFailed || resp.Status == job.StatusSkipped {
			errs[hostname] = resp.Error
		} else {
			changed[hostname] = resp.Changed != nil && *resp.Changed
		}
	}

	return jobID, changed, errs, nil
}

// QueryFileStatusBroadcast queries the status of a deployed file on a broadcast target.
func (c *Client) QueryFileStatusBroadcast(
	ctx context.Context,
	target string,
	path string,
) (string, map[string]*file.StatusResult, map[string]string, error) {
	data, _ := json.Marshal(file.StatusRequest{
		Path: path,
	})

	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  "file",
		Operation: job.OperationFileStatusGet,
		Data:      json.RawMessage(data),
	}

	subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, target)
	jobID, responses, err := c.publishAndCollect(ctx, subject, target, req)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to collect broadcast responses: %w", err)
	}

	results := make(map[string]*file.StatusResult)
	errs := make(map[string]string)
	for hostname, resp := range responses {
		if resp.Status == job.StatusFailed || resp.Status == job.StatusSkipped {
			errs[hostname] = resp.Error
		} else {
			var result file.StatusResult
			if unmarshalErr := json.Unmarshal(resp.Data, &result); unmarshalErr != nil {
				errs[hostname] = fmt.Sprintf("failed to unmarshal result: %v", unmarshalErr)
			} else {
				results[hostname] = &result
			}
		}
	}

	return jobID, results, errs, nil
}

// ModifyFileDeploy deploys a file to a specific hostname.
func (c *Client) ModifyFileDeploy(
	ctx context.Context,
	hostname string,
	objectName string,
	path string,
	contentType string,
	mode string,
	owner string,
	group string,
	vars map[string]any,
) (string, string, bool, error) {
	data, _ := json.Marshal(file.DeployRequest{
		ObjectName:  objectName,
		Path:        path,
		Mode:        mode,
		Owner:       owner,
		Group:       group,
		ContentType: contentType,
		Vars:        vars,
	})

	req := &job.Request{
		Type:      job.TypeModify,
		Category:  "file",
		Operation: job.OperationFileDeployExecute,
		Data:      json.RawMessage(data),
	}

	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, hostname)
	jobID, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return "", "", false, fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == job.StatusFailed || resp.Status == job.StatusSkipped {
		return "", "", false, fmt.Errorf("job failed: %s", resp.Error)
	}

	changed := resp.Changed != nil && *resp.Changed
	return jobID, resp.Hostname, changed, nil
}

// ModifyFileUndeploy removes a deployed file from disk on a specific hostname.
func (c *Client) ModifyFileUndeploy(
	ctx context.Context,
	hostname string,
	path string,
) (string, string, bool, error) {
	data, _ := json.Marshal(file.UndeployRequest{
		Path: path,
	})

	req := &job.Request{
		Type:      job.TypeModify,
		Category:  "file",
		Operation: job.OperationFileUndeployExecute,
		Data:      json.RawMessage(data),
	}

	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, hostname)
	jobID, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return "", "", false, fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == job.StatusFailed || resp.Status == job.StatusSkipped {
		return "", "", false, fmt.Errorf("job failed: %s", resp.Error)
	}

	changed := resp.Changed != nil && *resp.Changed
	return jobID, resp.Hostname, changed, nil
}

// QueryFileStatus queries the status of a deployed file on a specific hostname.
func (c *Client) QueryFileStatus(
	ctx context.Context,
	hostname string,
	path string,
) (string, *file.StatusResult, string, error) {
	data, _ := json.Marshal(file.StatusRequest{
		Path: path,
	})

	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  "file",
		Operation: job.OperationFileStatusGet,
		Data:      json.RawMessage(data),
	}

	subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, hostname)
	jobID, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return "", nil, "", fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == job.StatusFailed || resp.Status == job.StatusSkipped {
		return "", nil, "", fmt.Errorf("job failed: %s", resp.Error)
	}

	var result file.StatusResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return "", nil, "", fmt.Errorf("failed to unmarshal file status response: %w", err)
	}

	return jobID, &result, resp.Hostname, nil
}

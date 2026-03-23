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
	"github.com/retr0h/osapi/internal/provider/scheduled/cron"
)

// QueryScheduleCronList lists cron entries on a target.
func (c *Client) QueryScheduleCronList(
	ctx context.Context,
	target string,
) (*job.Response, error) {
	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  "schedule",
		Operation: job.OperationCronList,
	}

	subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, target)
	jobID, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return nil, fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == job.StatusFailed {
		return nil, fmt.Errorf("job failed: %s", resp.Error)
	}

	resp.JobID = jobID

	return resp, nil
}

// QueryScheduleCronListBroadcast lists cron entries from multiple agents.
func (c *Client) QueryScheduleCronListBroadcast(
	ctx context.Context,
	target string,
) (string, map[string]*job.Response, error) {
	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  "schedule",
		Operation: job.OperationCronList,
	}

	subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, target)
	jobID, responses, err := c.publishAndCollect(ctx, subject, target, req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to collect broadcast responses: %w", err)
	}

	return jobID, responses, nil
}

// QueryScheduleCronGet gets a single cron entry by name on a target.
func (c *Client) QueryScheduleCronGet(
	ctx context.Context,
	target string,
	name string,
) (*job.Response, error) {
	data := map[string]string{"name": name}
	dataBytes, _ := json.Marshal(data)
	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  "schedule",
		Operation: job.OperationCronGet,
		Data:      json.RawMessage(dataBytes),
	}

	subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, target)
	jobID, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return nil, fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == "failed" {
		return nil, fmt.Errorf("job failed: %s", resp.Error)
	}

	resp.JobID = jobID

	return resp, nil
}

// ModifyScheduleCronCreate creates a cron entry on a target.
func (c *Client) ModifyScheduleCronCreate(
	ctx context.Context,
	target string,
	entry cron.Entry,
) (*job.Response, error) {
	dataBytes, _ := json.Marshal(entry)
	req := &job.Request{
		Type:      job.TypeModify,
		Category:  "schedule",
		Operation: job.OperationCronCreate,
		Data:      json.RawMessage(dataBytes),
	}

	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, target)
	jobID, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return nil, fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == "failed" {
		return nil, fmt.Errorf("job failed: %s", resp.Error)
	}

	resp.JobID = jobID

	return resp, nil
}

// ModifyScheduleCronUpdate updates a cron entry on a target.
func (c *Client) ModifyScheduleCronUpdate(
	ctx context.Context,
	target string,
	entry cron.Entry,
) (*job.Response, error) {
	dataBytes, _ := json.Marshal(entry)
	req := &job.Request{
		Type:      job.TypeModify,
		Category:  "schedule",
		Operation: job.OperationCronUpdate,
		Data:      json.RawMessage(dataBytes),
	}

	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, target)
	jobID, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return nil, fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == "failed" {
		return nil, fmt.Errorf("job failed: %s", resp.Error)
	}

	resp.JobID = jobID

	return resp, nil
}

// ModifyScheduleCronDelete deletes a cron entry on a target.
func (c *Client) ModifyScheduleCronDelete(
	ctx context.Context,
	target string,
	name string,
) (*job.Response, error) {
	data := map[string]string{"name": name}
	dataBytes, _ := json.Marshal(data)
	req := &job.Request{
		Type:      job.TypeModify,
		Category:  "schedule",
		Operation: job.OperationCronDelete,
		Data:      json.RawMessage(dataBytes),
	}

	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, target)
	jobID, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return nil, fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == "failed" {
		return nil, fmt.Errorf("job failed: %s", resp.Error)
	}

	resp.JobID = jobID

	return resp, nil
}

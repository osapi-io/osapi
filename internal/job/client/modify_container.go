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
)

// ModifyContainerCreate creates a container on a target.
func (c *Client) ModifyContainerCreate(
	ctx context.Context,
	target string,
	data *job.ContainerCreateData,
) (*job.Response, error) {
	dataBytes, _ := json.Marshal(data)
	req := &job.Request{
		Type:      job.TypeModify,
		Category:  "container",
		Operation: job.OperationContainerCreate,
		Data:      json.RawMessage(dataBytes),
	}

	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, target)
	_, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return nil, fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == "failed" {
		return nil, fmt.Errorf("job failed: %s", resp.Error)
	}

	return resp, nil
}

// ModifyContainerStart starts a container on a target.
func (c *Client) ModifyContainerStart(
	ctx context.Context,
	target string,
	id string,
) (*job.Response, error) {
	data := map[string]string{"id": id}
	dataBytes, _ := json.Marshal(data)
	req := &job.Request{
		Type:      job.TypeModify,
		Category:  "container",
		Operation: job.OperationContainerStart,
		Data:      json.RawMessage(dataBytes),
	}

	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, target)
	_, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return nil, fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == "failed" {
		return nil, fmt.Errorf("job failed: %s", resp.Error)
	}

	return resp, nil
}

// ModifyContainerStop stops a container on a target.
func (c *Client) ModifyContainerStop(
	ctx context.Context,
	target string,
	id string,
	data *job.ContainerStopData,
) (*job.Response, error) {
	// Merge id into the data
	stopData := struct {
		ID      string `json:"id"`
		Timeout *int   `json:"timeout,omitempty"`
	}{
		ID:      id,
		Timeout: data.Timeout,
	}
	dataBytes, _ := json.Marshal(stopData)
	req := &job.Request{
		Type:      job.TypeModify,
		Category:  "container",
		Operation: job.OperationContainerStop,
		Data:      json.RawMessage(dataBytes),
	}

	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, target)
	_, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return nil, fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == "failed" {
		return nil, fmt.Errorf("job failed: %s", resp.Error)
	}

	return resp, nil
}

// ModifyContainerRemove removes a container on a target.
func (c *Client) ModifyContainerRemove(
	ctx context.Context,
	target string,
	id string,
	data *job.ContainerRemoveData,
) (*job.Response, error) {
	// Merge id into the data
	removeData := struct {
		ID    string `json:"id"`
		Force bool   `json:"force,omitempty"`
	}{
		ID:    id,
		Force: data.Force,
	}
	dataBytes, _ := json.Marshal(removeData)
	req := &job.Request{
		Type:      job.TypeModify,
		Category:  "container",
		Operation: job.OperationContainerRemove,
		Data:      json.RawMessage(dataBytes),
	}

	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, target)
	_, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return nil, fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == "failed" {
		return nil, fmt.Errorf("job failed: %s", resp.Error)
	}

	return resp, nil
}

// QueryContainerList lists containers on a target.
func (c *Client) QueryContainerList(
	ctx context.Context,
	target string,
	data *job.ContainerListData,
) (*job.Response, error) {
	dataBytes, _ := json.Marshal(data)
	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  "container",
		Operation: job.OperationContainerList,
		Data:      json.RawMessage(dataBytes),
	}

	subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, target)
	_, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return nil, fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == "failed" {
		return nil, fmt.Errorf("job failed: %s", resp.Error)
	}

	return resp, nil
}

// QueryContainerInspect inspects a container on a target.
func (c *Client) QueryContainerInspect(
	ctx context.Context,
	target string,
	id string,
) (*job.Response, error) {
	data := map[string]string{"id": id}
	dataBytes, _ := json.Marshal(data)
	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  "container",
		Operation: job.OperationContainerInspect,
		Data:      json.RawMessage(dataBytes),
	}

	subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, target)
	_, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return nil, fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == "failed" {
		return nil, fmt.Errorf("job failed: %s", resp.Error)
	}

	return resp, nil
}

// ModifyContainerExec executes a command in a container on a target.
func (c *Client) ModifyContainerExec(
	ctx context.Context,
	target string,
	id string,
	data *job.ContainerExecData,
) (*job.Response, error) {
	// Merge id into the data
	execData := struct {
		ID         string            `json:"id"`
		Command    []string          `json:"command"`
		Env        map[string]string `json:"env,omitempty"`
		WorkingDir string            `json:"working_dir,omitempty"`
	}{
		ID:         id,
		Command:    data.Command,
		Env:        data.Env,
		WorkingDir: data.WorkingDir,
	}
	dataBytes, _ := json.Marshal(execData)
	req := &job.Request{
		Type:      job.TypeModify,
		Category:  "container",
		Operation: job.OperationContainerExec,
		Data:      json.RawMessage(dataBytes),
	}

	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, target)
	_, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return nil, fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == "failed" {
		return nil, fmt.Errorf("job failed: %s", resp.Error)
	}

	return resp, nil
}

// ModifyContainerPull pulls an image on a target.
func (c *Client) ModifyContainerPull(
	ctx context.Context,
	target string,
	data *job.ContainerPullData,
) (*job.Response, error) {
	dataBytes, _ := json.Marshal(data)
	req := &job.Request{
		Type:      job.TypeModify,
		Category:  "container",
		Operation: job.OperationContainerPull,
		Data:      json.RawMessage(dataBytes),
	}

	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, target)
	_, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return nil, fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == "failed" {
		return nil, fmt.Errorf("job failed: %s", resp.Error)
	}

	return resp, nil
}

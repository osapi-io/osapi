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

// ModifyDockerCreate creates a docker container on a target.
func (c *Client) ModifyDockerCreate(
	ctx context.Context,
	target string,
	data *job.DockerCreateData,
) (*job.Response, error) {
	dataBytes, _ := json.Marshal(data)
	req := &job.Request{
		Type:      job.TypeModify,
		Category:  "docker",
		Operation: job.OperationDockerCreate,
		Data:      json.RawMessage(dataBytes),
	}

	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, target)
	jobID, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return nil, fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == job.StatusFailed || resp.Status == job.StatusSkipped {
		return nil, fmt.Errorf("job failed: %s", resp.Error)
	}

	resp.JobID = jobID

	return resp, nil
}

// ModifyDockerStart starts a docker container on a target.
func (c *Client) ModifyDockerStart(
	ctx context.Context,
	target string,
	id string,
) (*job.Response, error) {
	data := map[string]string{"id": id}
	dataBytes, _ := json.Marshal(data)
	req := &job.Request{
		Type:      job.TypeModify,
		Category:  "docker",
		Operation: job.OperationDockerStart,
		Data:      json.RawMessage(dataBytes),
	}

	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, target)
	jobID, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return nil, fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == job.StatusFailed || resp.Status == job.StatusSkipped {
		return nil, fmt.Errorf("job failed: %s", resp.Error)
	}

	resp.JobID = jobID

	return resp, nil
}

// ModifyDockerStop stops a docker container on a target.
func (c *Client) ModifyDockerStop(
	ctx context.Context,
	target string,
	id string,
	data *job.DockerStopData,
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
		Category:  "docker",
		Operation: job.OperationDockerStop,
		Data:      json.RawMessage(dataBytes),
	}

	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, target)
	jobID, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return nil, fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == job.StatusFailed || resp.Status == job.StatusSkipped {
		return nil, fmt.Errorf("job failed: %s", resp.Error)
	}

	resp.JobID = jobID

	return resp, nil
}

// ModifyDockerRemove removes a docker container on a target.
func (c *Client) ModifyDockerRemove(
	ctx context.Context,
	target string,
	id string,
	data *job.DockerRemoveData,
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
		Category:  "docker",
		Operation: job.OperationDockerRemove,
		Data:      json.RawMessage(dataBytes),
	}

	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, target)
	jobID, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return nil, fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == job.StatusFailed || resp.Status == job.StatusSkipped {
		return nil, fmt.Errorf("job failed: %s", resp.Error)
	}

	resp.JobID = jobID

	return resp, nil
}

// QueryDockerList lists docker containers on a target.
func (c *Client) QueryDockerList(
	ctx context.Context,
	target string,
	data *job.DockerListData,
) (*job.Response, error) {
	dataBytes, _ := json.Marshal(data)
	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  "docker",
		Operation: job.OperationDockerList,
		Data:      json.RawMessage(dataBytes),
	}

	subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, target)
	jobID, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return nil, fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == job.StatusFailed || resp.Status == job.StatusSkipped {
		return nil, fmt.Errorf("job failed: %s", resp.Error)
	}

	resp.JobID = jobID

	return resp, nil
}

// QueryDockerInspect inspects a docker container on a target.
func (c *Client) QueryDockerInspect(
	ctx context.Context,
	target string,
	id string,
) (*job.Response, error) {
	data := map[string]string{"id": id}
	dataBytes, _ := json.Marshal(data)
	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  "docker",
		Operation: job.OperationDockerInspect,
		Data:      json.RawMessage(dataBytes),
	}

	subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, target)
	jobID, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return nil, fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == job.StatusFailed || resp.Status == job.StatusSkipped {
		return nil, fmt.Errorf("job failed: %s", resp.Error)
	}

	resp.JobID = jobID

	return resp, nil
}

// ModifyDockerExec executes a command in a docker container on a target.
func (c *Client) ModifyDockerExec(
	ctx context.Context,
	target string,
	id string,
	data *job.DockerExecData,
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
		Category:  "docker",
		Operation: job.OperationDockerExec,
		Data:      json.RawMessage(dataBytes),
	}

	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, target)
	jobID, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return nil, fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == job.StatusFailed || resp.Status == job.StatusSkipped {
		return nil, fmt.Errorf("job failed: %s", resp.Error)
	}

	resp.JobID = jobID

	return resp, nil
}

// ModifyDockerImageRemove removes a docker image on a target.
func (c *Client) ModifyDockerImageRemove(
	ctx context.Context,
	target string,
	data *job.DockerImageRemoveData,
) (*job.Response, error) {
	dataBytes, _ := json.Marshal(data)
	req := &job.Request{
		Type:      job.TypeModify,
		Category:  "docker",
		Operation: job.OperationDockerImageRemove,
		Data:      json.RawMessage(dataBytes),
	}

	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, target)
	jobID, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return nil, fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == job.StatusFailed || resp.Status == job.StatusSkipped {
		return nil, fmt.Errorf("job failed: %s", resp.Error)
	}

	resp.JobID = jobID

	return resp, nil
}

// ModifyDockerPull pulls a docker image on a target.
func (c *Client) ModifyDockerPull(
	ctx context.Context,
	target string,
	data *job.DockerPullData,
) (*job.Response, error) {
	dataBytes, _ := json.Marshal(data)
	req := &job.Request{
		Type:      job.TypeModify,
		Category:  "docker",
		Operation: job.OperationDockerPull,
		Data:      json.RawMessage(dataBytes),
	}

	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, target)
	jobID, resp, err := c.publishAndWait(ctx, subject, req)
	if err != nil {
		return nil, fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == job.StatusFailed || resp.Status == job.StatusSkipped {
		return nil, fmt.Errorf("job failed: %s", resp.Error)
	}

	resp.JobID = jobID

	return resp, nil
}

// ModifyDockerCreateBroadcast creates a docker container on a broadcast target
// (_all or a label target like role:web).
func (c *Client) ModifyDockerCreateBroadcast(
	ctx context.Context,
	target string,
	data *job.DockerCreateData,
) (string, map[string]*job.Response, map[string]string, error) {
	dataBytes, _ := json.Marshal(data)
	req := &job.Request{
		Type:      job.TypeModify,
		Category:  "docker",
		Operation: job.OperationDockerCreate,
		Data:      json.RawMessage(dataBytes),
	}

	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, target)
	jobID, responses, err := c.publishAndCollect(ctx, subject, target, req)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to collect broadcast responses: %w", err)
	}

	results := make(map[string]*job.Response)
	errs := make(map[string]string)
	for hostname, resp := range responses {
		if resp.Status == job.StatusFailed || resp.Status == job.StatusSkipped {
			errs[hostname] = resp.Error
			continue
		}
		results[hostname] = resp
	}

	return jobID, results, errs, nil
}

// ModifyDockerStartBroadcast starts a docker container on a broadcast target
// (_all or a label target like role:web).
func (c *Client) ModifyDockerStartBroadcast(
	ctx context.Context,
	target string,
	id string,
) (string, map[string]*job.Response, map[string]string, error) {
	data := map[string]string{"id": id}
	dataBytes, _ := json.Marshal(data)
	req := &job.Request{
		Type:      job.TypeModify,
		Category:  "docker",
		Operation: job.OperationDockerStart,
		Data:      json.RawMessage(dataBytes),
	}

	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, target)
	jobID, responses, err := c.publishAndCollect(ctx, subject, target, req)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to collect broadcast responses: %w", err)
	}

	results := make(map[string]*job.Response)
	errs := make(map[string]string)
	for hostname, resp := range responses {
		if resp.Status == job.StatusFailed || resp.Status == job.StatusSkipped {
			errs[hostname] = resp.Error
			continue
		}
		results[hostname] = resp
	}

	return jobID, results, errs, nil
}

// ModifyDockerStopBroadcast stops a docker container on a broadcast target
// (_all or a label target like role:web).
func (c *Client) ModifyDockerStopBroadcast(
	ctx context.Context,
	target string,
	id string,
	data *job.DockerStopData,
) (string, map[string]*job.Response, map[string]string, error) {
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
		Category:  "docker",
		Operation: job.OperationDockerStop,
		Data:      json.RawMessage(dataBytes),
	}

	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, target)
	jobID, responses, err := c.publishAndCollect(ctx, subject, target, req)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to collect broadcast responses: %w", err)
	}

	results := make(map[string]*job.Response)
	errs := make(map[string]string)
	for hostname, resp := range responses {
		if resp.Status == job.StatusFailed || resp.Status == job.StatusSkipped {
			errs[hostname] = resp.Error
			continue
		}
		results[hostname] = resp
	}

	return jobID, results, errs, nil
}

// ModifyDockerRemoveBroadcast removes a docker container on a broadcast target
// (_all or a label target like role:web).
func (c *Client) ModifyDockerRemoveBroadcast(
	ctx context.Context,
	target string,
	id string,
	data *job.DockerRemoveData,
) (string, map[string]*job.Response, map[string]string, error) {
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
		Category:  "docker",
		Operation: job.OperationDockerRemove,
		Data:      json.RawMessage(dataBytes),
	}

	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, target)
	jobID, responses, err := c.publishAndCollect(ctx, subject, target, req)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to collect broadcast responses: %w", err)
	}

	results := make(map[string]*job.Response)
	errs := make(map[string]string)
	for hostname, resp := range responses {
		if resp.Status == job.StatusFailed || resp.Status == job.StatusSkipped {
			errs[hostname] = resp.Error
			continue
		}
		results[hostname] = resp
	}

	return jobID, results, errs, nil
}

// QueryDockerListBroadcast lists docker containers on a broadcast target
// (_all or a label target like role:web).
func (c *Client) QueryDockerListBroadcast(
	ctx context.Context,
	target string,
	data *job.DockerListData,
) (string, map[string]*job.Response, map[string]string, error) {
	dataBytes, _ := json.Marshal(data)
	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  "docker",
		Operation: job.OperationDockerList,
		Data:      json.RawMessage(dataBytes),
	}

	subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, target)
	jobID, responses, err := c.publishAndCollect(ctx, subject, target, req)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to collect broadcast responses: %w", err)
	}

	results := make(map[string]*job.Response)
	errs := make(map[string]string)
	for hostname, resp := range responses {
		if resp.Status == job.StatusFailed || resp.Status == job.StatusSkipped {
			errs[hostname] = resp.Error
			continue
		}
		results[hostname] = resp
	}

	return jobID, results, errs, nil
}

// QueryDockerInspectBroadcast inspects a docker container on a broadcast target
// (_all or a label target like role:web).
func (c *Client) QueryDockerInspectBroadcast(
	ctx context.Context,
	target string,
	id string,
) (string, map[string]*job.Response, map[string]string, error) {
	data := map[string]string{"id": id}
	dataBytes, _ := json.Marshal(data)
	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  "docker",
		Operation: job.OperationDockerInspect,
		Data:      json.RawMessage(dataBytes),
	}

	subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, target)
	jobID, responses, err := c.publishAndCollect(ctx, subject, target, req)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to collect broadcast responses: %w", err)
	}

	results := make(map[string]*job.Response)
	errs := make(map[string]string)
	for hostname, resp := range responses {
		if resp.Status == job.StatusFailed || resp.Status == job.StatusSkipped {
			errs[hostname] = resp.Error
			continue
		}
		results[hostname] = resp
	}

	return jobID, results, errs, nil
}

// ModifyDockerExecBroadcast executes a command in a docker container on a
// broadcast target (_all or a label target like role:web).
func (c *Client) ModifyDockerExecBroadcast(
	ctx context.Context,
	target string,
	id string,
	data *job.DockerExecData,
) (string, map[string]*job.Response, map[string]string, error) {
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
		Category:  "docker",
		Operation: job.OperationDockerExec,
		Data:      json.RawMessage(dataBytes),
	}

	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, target)
	jobID, responses, err := c.publishAndCollect(ctx, subject, target, req)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to collect broadcast responses: %w", err)
	}

	results := make(map[string]*job.Response)
	errs := make(map[string]string)
	for hostname, resp := range responses {
		if resp.Status == job.StatusFailed || resp.Status == job.StatusSkipped {
			errs[hostname] = resp.Error
			continue
		}
		results[hostname] = resp
	}

	return jobID, results, errs, nil
}

// ModifyDockerPullBroadcast pulls a docker image on a broadcast target
// (_all or a label target like role:web).
func (c *Client) ModifyDockerPullBroadcast(
	ctx context.Context,
	target string,
	data *job.DockerPullData,
) (string, map[string]*job.Response, map[string]string, error) {
	dataBytes, _ := json.Marshal(data)
	req := &job.Request{
		Type:      job.TypeModify,
		Category:  "docker",
		Operation: job.OperationDockerPull,
		Data:      json.RawMessage(dataBytes),
	}

	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, target)
	jobID, responses, err := c.publishAndCollect(ctx, subject, target, req)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to collect broadcast responses: %w", err)
	}

	results := make(map[string]*job.Response)
	errs := make(map[string]string)
	for hostname, resp := range responses {
		if resp.Status == job.StatusFailed || resp.Status == job.StatusSkipped {
			errs[hostname] = resp.Error
			continue
		}
		results[hostname] = resp
	}

	return jobID, results, errs, nil
}

// ModifyDockerImageRemoveBroadcast removes a docker image on a broadcast target
// (_all or a label target like role:web).
func (c *Client) ModifyDockerImageRemoveBroadcast(
	ctx context.Context,
	target string,
	data *job.DockerImageRemoveData,
) (string, map[string]*job.Response, map[string]string, error) {
	dataBytes, _ := json.Marshal(data)
	req := &job.Request{
		Type:      job.TypeModify,
		Category:  "docker",
		Operation: job.OperationDockerImageRemove,
		Data:      json.RawMessage(dataBytes),
	}

	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, target)
	jobID, responses, err := c.publishAndCollect(ctx, subject, target, req)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to collect broadcast responses: %w", err)
	}

	results := make(map[string]*job.Response)
	errs := make(map[string]string)
	for hostname, resp := range responses {
		if resp.Status == job.StatusFailed || resp.Status == job.StatusSkipped {
			errs[hostname] = resp.Error
			continue
		}
		results[hostname] = resp
	}

	return jobID, results, errs, nil
}

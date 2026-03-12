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

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider/container/runtime"
)

// processContainerOperation handles container-related operations.
func (a *Agent) processContainerOperation(
	jobRequest job.Request,
) (json.RawMessage, error) {
	if a.containerProvider == nil {
		return nil, fmt.Errorf("container runtime not available")
	}

	ctx := context.Background()

	// Extract base operation from dotted operation (e.g., "create.execute" -> "create")
	baseOperation := strings.Split(jobRequest.Operation, ".")[0]

	switch baseOperation {
	case "create":
		return a.processContainerCreate(ctx, jobRequest)
	case "start":
		return a.processContainerStart(ctx, jobRequest)
	case "stop":
		return a.processContainerStop(ctx, jobRequest)
	case "remove":
		return a.processContainerRemove(ctx, jobRequest)
	case "list":
		return a.processContainerList(ctx, jobRequest)
	case "inspect":
		return a.processContainerInspect(ctx, jobRequest)
	case "exec":
		return a.processContainerExec(ctx, jobRequest)
	case "pull":
		return a.processContainerPull(ctx, jobRequest)
	default:
		return nil, fmt.Errorf("unsupported container operation: %s", jobRequest.Operation)
	}
}

// processContainerCreate handles container creation.
func (a *Agent) processContainerCreate(
	ctx context.Context,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data job.ContainerCreateData
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal create data: %w", err)
	}

	// Map ports and volumes from job types to runtime types
	var ports []runtime.PortMapping
	for _, p := range data.Ports {
		ports = append(ports, runtime.PortMapping{Host: p.Host, Container: p.Container})
	}

	var volumes []runtime.VolumeMapping
	for _, v := range data.Volumes {
		volumes = append(volumes, runtime.VolumeMapping{Host: v.Host, Container: v.Container})
	}

	result, err := a.containerProvider.Create(ctx, runtime.CreateParams{
		Image:     data.Image,
		Name:      data.Name,
		Command:   data.Command,
		Env:       data.Env,
		Ports:     ports,
		Volumes:   volumes,
		AutoStart: data.AutoStart,
	})
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processContainerStart handles starting a container.
func (a *Agent) processContainerStart(
	ctx context.Context,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal start data: %w", err)
	}

	if err := a.containerProvider.Start(ctx, data.ID); err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"message": "Container started successfully",
	}
	return json.Marshal(result)
}

// processContainerStop handles stopping a container.
func (a *Agent) processContainerStop(
	ctx context.Context,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		ID      string `json:"id"`
		Timeout *int   `json:"timeout,omitempty"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal stop data: %w", err)
	}

	var timeout *time.Duration
	if data.Timeout != nil {
		d := time.Duration(*data.Timeout) * time.Second
		timeout = &d
	}

	if err := a.containerProvider.Stop(ctx, data.ID, timeout); err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"message": "Container stopped successfully",
	}
	return json.Marshal(result)
}

// processContainerRemove handles removing a container.
func (a *Agent) processContainerRemove(
	ctx context.Context,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		ID    string `json:"id"`
		Force bool   `json:"force,omitempty"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal remove data: %w", err)
	}

	if err := a.containerProvider.Remove(ctx, data.ID, data.Force); err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"message": "Container removed successfully",
	}
	return json.Marshal(result)
}

// processContainerList handles listing containers.
func (a *Agent) processContainerList(
	ctx context.Context,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data job.ContainerListData
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal list data: %w", err)
	}

	result, err := a.containerProvider.List(ctx, runtime.ListParams{
		State: data.State,
		Limit: data.Limit,
	})
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processContainerInspect handles inspecting a container.
func (a *Agent) processContainerInspect(
	ctx context.Context,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal inspect data: %w", err)
	}

	result, err := a.containerProvider.Inspect(ctx, data.ID)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processContainerExec handles executing a command in a container.
func (a *Agent) processContainerExec(
	ctx context.Context,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		ID         string            `json:"id"`
		Command    []string          `json:"command"`
		Env        map[string]string `json:"env,omitempty"`
		WorkingDir string            `json:"working_dir,omitempty"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal exec data: %w", err)
	}

	result, err := a.containerProvider.Exec(ctx, data.ID, runtime.ExecParams{
		Command:    data.Command,
		Env:        data.Env,
		WorkingDir: data.WorkingDir,
	})
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processContainerPull handles pulling a container image.
func (a *Agent) processContainerPull(
	ctx context.Context,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		Image string `json:"image"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal pull data: %w", err)
	}

	result, err := a.containerProvider.Pull(ctx, data.Image)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

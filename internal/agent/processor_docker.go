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
	dockerProv "github.com/retr0h/osapi/internal/provider/container/docker"
)

// processDockerOperation handles docker-related operations.
func (a *Agent) processDockerOperation(
	jobRequest job.Request,
) (json.RawMessage, error) {
	if a.dockerProvider == nil {
		return nil, fmt.Errorf("docker runtime not available")
	}

	ctx := context.Background()

	// Extract base operation from dotted operation (e.g., "create.execute" -> "create")
	baseOperation := strings.Split(jobRequest.Operation, ".")[0]

	switch baseOperation {
	case "create":
		return a.processDockerCreate(ctx, jobRequest)
	case "start":
		return a.processDockerStart(ctx, jobRequest)
	case "stop":
		return a.processDockerStop(ctx, jobRequest)
	case "remove":
		return a.processDockerRemove(ctx, jobRequest)
	case "list":
		return a.processDockerList(ctx, jobRequest)
	case "inspect":
		return a.processDockerInspect(ctx, jobRequest)
	case "exec":
		return a.processDockerExec(ctx, jobRequest)
	case "pull":
		return a.processDockerPull(ctx, jobRequest)
	case "image-remove":
		return a.processDockerImageRemove(ctx, jobRequest)
	default:
		return nil, fmt.Errorf("unsupported docker operation: %s", jobRequest.Operation)
	}
}

// processDockerCreate handles docker container creation.
func (a *Agent) processDockerCreate(
	ctx context.Context,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data job.DockerCreateData
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal create data: %w", err)
	}

	// Map ports and volumes from job types to runtime types
	var ports []dockerProv.PortMapping
	for _, p := range data.Ports {
		ports = append(ports, dockerProv.PortMapping{Host: p.Host, Container: p.Container})
	}

	var volumes []dockerProv.VolumeMapping
	for _, v := range data.Volumes {
		volumes = append(volumes, dockerProv.VolumeMapping{Host: v.Host, Container: v.Container})
	}

	result, err := a.dockerProvider.Create(ctx, dockerProv.CreateParams{
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

// processDockerStart handles starting a docker container.
func (a *Agent) processDockerStart(
	ctx context.Context,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal start data: %w", err)
	}

	result, err := a.dockerProvider.Start(ctx, data.ID)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processDockerStop handles stopping a docker container.
func (a *Agent) processDockerStop(
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

	result, err := a.dockerProvider.Stop(ctx, data.ID, timeout)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processDockerRemove handles removing a docker container.
func (a *Agent) processDockerRemove(
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

	result, err := a.dockerProvider.Remove(ctx, data.ID, data.Force)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processDockerList handles listing docker containers.
func (a *Agent) processDockerList(
	ctx context.Context,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data job.DockerListData
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal list data: %w", err)
	}

	result, err := a.dockerProvider.List(ctx, dockerProv.ListParams{
		State: data.State,
		Limit: data.Limit,
	})
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processDockerInspect handles inspecting a docker container.
func (a *Agent) processDockerInspect(
	ctx context.Context,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal inspect data: %w", err)
	}

	result, err := a.dockerProvider.Inspect(ctx, data.ID)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processDockerExec handles executing a command in a docker container.
func (a *Agent) processDockerExec(
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

	result, err := a.dockerProvider.Exec(ctx, data.ID, dockerProv.ExecParams{
		Command:    data.Command,
		Env:        data.Env,
		WorkingDir: data.WorkingDir,
	})
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processDockerPull handles pulling a docker image.
func (a *Agent) processDockerPull(
	ctx context.Context,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		Image string `json:"image"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal pull data: %w", err)
	}

	result, err := a.dockerProvider.Pull(ctx, data.Image)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processDockerImageRemove handles removing a docker image.
func (a *Agent) processDockerImageRemove(
	ctx context.Context,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data job.DockerImageRemoveData
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal image-remove data: %w", err)
	}

	result, err := a.dockerProvider.ImageRemove(ctx, data.Image, data.Force)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

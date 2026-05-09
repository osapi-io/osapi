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

package container

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/node/docker/gen"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/validation"
)

// PostNodeContainerDockerExec executes a command in a container on a target node.
func (s *Container) PostNodeContainerDockerExec(
	ctx context.Context,
	request gen.PostNodeContainerDockerExecRequestObject,
) (gen.PostNodeContainerDockerExecResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.PostNodeContainerDockerExec400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validation.Var(request.Id, "required,min=1"); !ok {
		return gen.PostNodeContainerDockerExec400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validation.Struct(request.Body); !ok {
		return gen.PostNodeContainerDockerExec400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname
	id := request.Id

	data := &job.DockerExecData{
		Command: request.Body.Command,
		Env:     envSliceToMap(request.Body.Env),
	}
	if request.Body.WorkingDir != nil {
		data.WorkingDir = *request.Body.WorkingDir
	}

	s.logger.Debug(
		"container exec",
		slog.String("target", hostname),
		slog.String("id", id),
		slog.Any("command", data.Command),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.postNodeContainerDockerExecBroadcast(ctx, hostname, id, data)
	}

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
	jobID, resp, err := s.JobClient.Modify(
		ctx,
		hostname,
		"docker",
		job.OperationDockerExec,
		execData,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeContainerDockerExec500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		jobUUID := uuid.MustParse(jobID)
		e := resp.Error
		return gen.PostNodeContainerDockerExec202JSONResponse{
			JobId: &jobUUID,
			Results: []gen.DockerExecResultItem{
				{
					Hostname: resp.Hostname,
					Status:   gen.DockerExecResultItemStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	item := dockerExecItemFromResponse(resp)
	jobUUID := uuid.MustParse(jobID)

	return gen.PostNodeContainerDockerExec202JSONResponse{
		JobId:   &jobUUID,
		Results: []gen.DockerExecResultItem{item},
	}, nil
}

// dockerExecItemFromResponse builds a DockerExecResultItem from a job response.
func dockerExecItemFromResponse(
	resp *job.Response,
) gen.DockerExecResultItem {
	var execResult struct {
		Stdout   string `json:"stdout"`
		Stderr   string `json:"stderr"`
		ExitCode int    `json:"exit_code"`
	}
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &execResult)
	}

	stdout := execResult.Stdout
	stderr := execResult.Stderr
	exitCode := execResult.ExitCode

	return gen.DockerExecResultItem{
		Hostname: resp.Hostname,
		Status:   gen.DockerExecResultItemStatusOk,
		Stdout:   &stdout,
		Stderr:   &stderr,
		ExitCode: &exitCode,
		Changed:  resp.Changed,
	}
}

// postNodeContainerDockerExecBroadcast handles broadcast targets for container exec.
func (s *Container) postNodeContainerDockerExecBroadcast(
	ctx context.Context,
	target string,
	id string,
	data *job.DockerExecData,
) (gen.PostNodeContainerDockerExecResponseObject, error) {
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
	jobID, responses, err := s.JobClient.ModifyBroadcast(
		ctx,
		target,
		"docker",
		job.OperationDockerExec,
		execData,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeContainerDockerExec500JSONResponse{Error: &errMsg}, nil
	}

	var items []gen.DockerExecResultItem
	for host, resp := range responses {
		item := gen.DockerExecResultItem{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.DockerExecResultItemStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.DockerExecResultItemStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.DockerExecResultItemStatusOk
			ok := dockerExecItemFromResponse(resp)
			item.Stdout = ok.Stdout
			item.Stderr = ok.Stderr
			item.ExitCode = ok.ExitCode
			item.Changed = resp.Changed
		}
		items = append(items, item)
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.PostNodeContainerDockerExec202JSONResponse{
		JobId:   &jobUUID,
		Results: items,
	}, nil
}

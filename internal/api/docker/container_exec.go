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

	"github.com/retr0h/osapi/internal/api/docker/gen"
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

	s.logger.Debug("container exec",
		slog.String("target", hostname),
		slog.String("id", id),
		slog.Any("command", data.Command),
	)

	resp, err := s.JobClient.ModifyDockerExec(ctx, hostname, id, data)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeContainerDockerExec500JSONResponse{Error: &errMsg}, nil
	}

	var execResult struct {
		Stdout   string `json:"stdout"`
		Stderr   string `json:"stderr"`
		ExitCode int    `json:"exit_code"`
	}
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &execResult)
	}

	jobUUID := uuid.MustParse(resp.JobID)
	changed := resp.Changed
	stdout := execResult.Stdout
	stderr := execResult.Stderr
	exitCode := execResult.ExitCode

	return gen.PostNodeContainerDockerExec202JSONResponse{
		JobId: &jobUUID,
		Results: []gen.DockerExecResultItem{
			{
				Hostname: resp.Hostname,
				Stdout:   &stdout,
				Stderr:   &stderr,
				ExitCode: &exitCode,
				Changed:  changed,
			},
		},
	}, nil
}

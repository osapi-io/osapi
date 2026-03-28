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

package node

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/node/gen"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider/command"
	"github.com/retr0h/osapi/internal/validation"
)

// PostNodeCommandShell post the node command shell API endpoint.
func (s *Node) PostNodeCommandShell(
	ctx context.Context,
	request gen.PostNodeCommandShellRequestObject,
) (gen.PostNodeCommandShellResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.PostNodeCommandShell400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validation.Struct(request.Body); !ok {
		return gen.PostNodeCommandShell400JSONResponse{
			Error: &errMsg,
		}, nil
	}

	cmdStr := request.Body.Command

	var cwd string
	if request.Body.Cwd != nil {
		cwd = *request.Body.Cwd
	}

	var timeout int
	if request.Body.Timeout != nil {
		timeout = *request.Body.Timeout
	}

	hostname := request.Hostname

	s.logger.Debug("command shell",
		slog.String("command", cmdStr),
		slog.String("target", hostname),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.postNodeCommandShellBroadcast(ctx, hostname, cmdStr, cwd, timeout)
	}

	data := job.CommandShellData{
		Command: cmdStr,
		Cwd:     cwd,
		Timeout: timeout,
	}
	jobID, rawResp, err := s.JobClient.Modify(
		ctx,
		hostname,
		"command",
		job.OperationCommandShellExecute,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeCommandShell500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	var result command.Result
	if rawResp.Data != nil {
		_ = json.Unmarshal(rawResp.Data, &result)
	}

	jobUUID := uuid.MustParse(jobID)
	stdout := result.Stdout
	stderr := result.Stderr
	exitCode := result.ExitCode
	durationMs := result.DurationMs
	changed := result.Changed

	return gen.PostNodeCommandShell202JSONResponse{
		JobId: &jobUUID,
		Results: []gen.CommandResultItem{
			{
				Hostname:   rawResp.Hostname,
				Status:     gen.CommandResultItemStatusOk,
				Stdout:     &stdout,
				Stderr:     &stderr,
				ExitCode:   &exitCode,
				DurationMs: &durationMs,
				Changed:    &changed,
			},
		},
	}, nil
}

// postNodeCommandShellBroadcast handles broadcast targets for command shell.
func (s *Node) postNodeCommandShellBroadcast(
	ctx context.Context,
	target string,
	cmdStr string,
	cwd string,
	timeout int,
) (gen.PostNodeCommandShellResponseObject, error) {
	data := job.CommandShellData{
		Command: cmdStr,
		Cwd:     cwd,
		Timeout: timeout,
	}
	jobID, responses, err := s.JobClient.ModifyBroadcast(
		ctx,
		target,
		"command",
		job.OperationCommandShellExecute,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeCommandShell500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	var apiResponses []gen.CommandResultItem
	for host, resp := range responses {
		item := gen.CommandResultItem{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.CommandResultItemStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.CommandResultItemStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.CommandResultItemStatusOk
			var result command.Result
			if resp.Data != nil {
				_ = json.Unmarshal(resp.Data, &result)
			}
			stdout := result.Stdout
			stderr := result.Stderr
			exitCode := result.ExitCode
			durationMs := result.DurationMs
			changed := result.Changed
			item.Stdout = &stdout
			item.Stderr = &stderr
			item.ExitCode = &exitCode
			item.DurationMs = &durationMs
			item.Changed = &changed
		}
		apiResponses = append(apiResponses, item)
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.PostNodeCommandShell202JSONResponse{
		JobId:   &jobUUID,
		Results: apiResponses,
	}, nil
}

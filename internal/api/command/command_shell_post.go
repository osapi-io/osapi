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

package command

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/api/command/gen"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/validation"
)

// PostCommandShell post the command shell API endpoint.
func (c Command) PostCommandShell(
	ctx context.Context,
	request gen.PostCommandShellRequestObject,
) (gen.PostCommandShellResponseObject, error) {
	if errMsg, ok := validation.Struct(request.Body); !ok {
		return gen.PostCommandShell400JSONResponse{
			Error: &errMsg,
		}, nil
	}

	if errMsg, ok := validation.Struct(request.Params); !ok {
		return gen.PostCommandShell400JSONResponse{Error: &errMsg}, nil
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

	hostname := job.AnyHost
	if request.Params.TargetHostname != nil {
		hostname = *request.Params.TargetHostname
	}

	c.logger.Debug("command shell",
		slog.String("command", cmdStr),
		slog.String("target", hostname),
	)

	if job.IsBroadcastTarget(hostname) {
		return c.postCommandShellBroadcast(ctx, hostname, cmdStr, cwd, timeout)
	}

	jobID, result, workerHostname, err := c.JobClient.ModifyCommandShell(
		ctx,
		hostname,
		cmdStr,
		cwd,
		timeout,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostCommandShell500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	jobUUID := uuid.MustParse(jobID)
	stdout := result.Stdout
	stderr := result.Stderr
	exitCode := result.ExitCode
	durationMs := result.DurationMs

	return gen.PostCommandShell202JSONResponse{
		JobId: &jobUUID,
		Results: []gen.CommandResultItem{
			{
				Hostname:   workerHostname,
				Stdout:     &stdout,
				Stderr:     &stderr,
				ExitCode:   &exitCode,
				DurationMs: &durationMs,
			},
		},
	}, nil
}

// postCommandShellBroadcast handles broadcast targets for command shell.
func (c Command) postCommandShellBroadcast(
	ctx context.Context,
	target string,
	cmdStr string,
	cwd string,
	timeout int,
) (gen.PostCommandShellResponseObject, error) {
	jobID, results, errs, err := c.JobClient.ModifyCommandShellBroadcast(
		ctx,
		target,
		cmdStr,
		cwd,
		timeout,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostCommandShell500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	var responses []gen.CommandResultItem
	for host, result := range results {
		stdout := result.Stdout
		stderr := result.Stderr
		exitCode := result.ExitCode
		durationMs := result.DurationMs
		responses = append(responses, gen.CommandResultItem{
			Hostname:   host,
			Stdout:     &stdout,
			Stderr:     &stderr,
			ExitCode:   &exitCode,
			DurationMs: &durationMs,
		})
	}
	for host, errMsg := range errs {
		e := errMsg
		responses = append(responses, gen.CommandResultItem{
			Hostname: host,
			Error:    &e,
		})
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.PostCommandShell202JSONResponse{
		JobId:   &jobUUID,
		Results: responses,
	}, nil
}

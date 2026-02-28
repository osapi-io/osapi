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
	"log/slog"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/api/node/gen"
	"github.com/retr0h/osapi/internal/job"
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

	jobID, result, agentHostname, err := s.JobClient.ModifyCommandShell(
		ctx,
		hostname,
		cmdStr,
		cwd,
		timeout,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeCommandShell500JSONResponse{
			Error: &errMsg,
		}, nil
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
				Hostname:   agentHostname,
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
	jobID, results, errs, err := s.JobClient.ModifyCommandShellBroadcast(
		ctx,
		target,
		cmdStr,
		cwd,
		timeout,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeCommandShell500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	var responses []gen.CommandResultItem
	for host, result := range results {
		stdout := result.Stdout
		stderr := result.Stderr
		exitCode := result.ExitCode
		durationMs := result.DurationMs
		changed := result.Changed
		responses = append(responses, gen.CommandResultItem{
			Hostname:   host,
			Stdout:     &stdout,
			Stderr:     &stderr,
			ExitCode:   &exitCode,
			DurationMs: &durationMs,
			Changed:    &changed,
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
	return gen.PostNodeCommandShell202JSONResponse{
		JobId:   &jobUUID,
		Results: responses,
	}, nil
}

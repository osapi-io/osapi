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

package process

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/node/process/gen"
	"github.com/retr0h/osapi/internal/job"
	processProv "github.com/retr0h/osapi/internal/provider/node/process"
)

// GetNodeProcessByPid gets a process by PID on a target node.
func (s *Process) GetNodeProcessByPid(
	ctx context.Context,
	request gen.GetNodeProcessByPidRequestObject,
) (gen.GetNodeProcessByPidResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeProcessByPid400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname
	pid := request.Pid

	s.logger.Debug(
		"process get",
		slog.String("target", hostname),
		slog.Int("pid", pid),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.getNodeProcessByPidBroadcast(ctx, hostname, pid)
	}

	jobID, resp, err := s.JobClient.Query(
		ctx,
		hostname,
		"node",
		job.OperationProcessGet,
		map[string]int{"pid": pid},
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeProcessByPid500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		e := resp.Error
		jobUUID := uuid.MustParse(jobID)
		return gen.GetNodeProcessByPid200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.ProcessGetEntry{
				{
					Hostname: resp.Hostname,
					Status:   gen.ProcessGetEntryStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	if resp.Status == job.StatusFailed && strings.Contains(resp.Error, "not found") {
		errMsg := resp.Error
		return gen.GetNodeProcessByPid404JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusFailed {
		errMsg := resp.Error
		return gen.GetNodeProcessByPid500JSONResponse{Error: &errMsg}, nil
	}

	var info processProv.Info
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &info)
	}

	genInfo := processInfoToGen(info)
	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeProcessByPid200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.ProcessGetEntry{
			{
				Hostname: resp.Hostname,
				Status:   gen.ProcessGetEntryStatusOk,
				Process:  &genInfo,
			},
		},
	}, nil
}

// getNodeProcessByPidBroadcast handles broadcast targets for process get by PID.
func (s *Process) getNodeProcessByPidBroadcast(
	ctx context.Context,
	target string,
	pid int,
) (gen.GetNodeProcessByPidResponseObject, error) {
	jobID, responses, err := s.JobClient.QueryBroadcast(
		ctx,
		target,
		"node",
		job.OperationProcessGet,
		map[string]int{"pid": pid},
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeProcessByPid500JSONResponse{Error: &errMsg}, nil
	}

	var items []gen.ProcessGetEntry
	for host, resp := range responses {
		item := gen.ProcessGetEntry{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.ProcessGetEntryStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.ProcessGetEntryStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.ProcessGetEntryStatusOk
			var info processProv.Info
			if resp.Data != nil {
				_ = json.Unmarshal(resp.Data, &info)
			}
			genInfo := processInfoToGen(info)
			item.Process = &genInfo
		}
		items = append(items, item)
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.GetNodeProcessByPid200JSONResponse{
		JobId:   &jobUUID,
		Results: items,
	}, nil
}

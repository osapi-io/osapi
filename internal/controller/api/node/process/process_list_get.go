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

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/node/process/gen"
	"github.com/retr0h/osapi/internal/job"
	processProv "github.com/retr0h/osapi/internal/provider/node/process"
)

// GetNodeProcess lists processes on a target node.
func (s *Process) GetNodeProcess(
	ctx context.Context,
	request gen.GetNodeProcessRequestObject,
) (gen.GetNodeProcessResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeProcess500JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	s.logger.Debug("process list",
		slog.String("target", hostname),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.getNodeProcessListBroadcast(ctx, hostname)
	}

	jobID, resp, err := s.JobClient.Query(ctx, hostname, "node", job.OperationProcessList, nil)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeProcess500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		e := resp.Error
		jobUUID := uuid.MustParse(jobID)
		return gen.GetNodeProcess200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.ProcessEntry{
				{
					Hostname: resp.Hostname,
					Status:   gen.ProcessEntryStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	processes := processInfoListFromResponse(resp)
	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeProcess200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.ProcessEntry{
			{
				Hostname:  resp.Hostname,
				Status:    gen.ProcessEntryStatusOk,
				Processes: &processes,
			},
		},
	}, nil
}

// processInfoListFromResponse extracts ProcessInfo slice from a job response.
func processInfoListFromResponse(
	resp *job.Response,
) []gen.ProcessInfo {
	var infos []processProv.Info
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &infos)
	}

	result := make([]gen.ProcessInfo, 0, len(infos))
	for _, info := range infos {
		result = append(result, processInfoToGen(info))
	}

	return result
}

// processInfoToGen converts a provider process.Info to a gen.ProcessInfo.
func processInfoToGen(
	info processProv.Info,
) gen.ProcessInfo {
	pid := info.PID
	name := info.Name
	user := info.User
	state := info.State
	cpuPercent := info.CPUPercent
	memPercent := info.MemPercent
	memRSS := info.MemRSS
	command := info.Command
	startTime := info.StartTime

	return gen.ProcessInfo{
		Pid:        &pid,
		Name:       &name,
		User:       &user,
		State:      &state,
		CpuPercent: &cpuPercent,
		MemPercent: &memPercent,
		MemRss:     &memRSS,
		Command:    &command,
		StartTime:  stringPtrOrNil(startTime),
	}
}

// stringPtrOrNil returns nil if the string is empty, otherwise a pointer.
func stringPtrOrNil(
	s string,
) *string {
	if s == "" {
		return nil
	}
	return &s
}

// getNodeProcessListBroadcast handles broadcast targets for process list.
func (s *Process) getNodeProcessListBroadcast(
	ctx context.Context,
	target string,
) (gen.GetNodeProcessResponseObject, error) {
	jobID, responses, err := s.JobClient.QueryBroadcast(
		ctx,
		target,
		"node",
		job.OperationProcessList,
		nil,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeProcess500JSONResponse{Error: &errMsg}, nil
	}

	var items []gen.ProcessEntry
	for host, resp := range responses {
		item := gen.ProcessEntry{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.ProcessEntryStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.ProcessEntryStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.ProcessEntryStatusOk
			processes := processInfoListFromResponse(resp)
			item.Processes = &processes
		}
		items = append(items, item)
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.GetNodeProcess200JSONResponse{
		JobId:   &jobUUID,
		Results: items,
	}, nil
}

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
	"github.com/retr0h/osapi/internal/validation"
)

// PostNodeProcessSignal sends a signal to a process on a target node.
func (s *Process) PostNodeProcessSignal(
	ctx context.Context,
	request gen.PostNodeProcessSignalRequestObject,
) (gen.PostNodeProcessSignalResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.PostNodeProcessSignal400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validation.Struct(request.Body); !ok {
		return gen.PostNodeProcessSignal400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname
	pid := request.Pid
	signal := string(request.Body.Signal)

	s.logger.Debug(
		"process signal",
		slog.String("target", hostname),
		slog.Int("pid", pid),
		slog.String("signal", signal),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	data := struct {
		PID    int    `json:"pid"`
		Signal string `json:"signal"`
	}{
		PID:    pid,
		Signal: signal,
	}

	if job.IsBroadcastTarget(hostname) {
		return s.postNodeProcessSignalBroadcast(ctx, hostname, data)
	}

	jobID, resp, err := s.JobClient.Modify(
		ctx,
		hostname,
		"node",
		job.OperationProcessSignal,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeProcessSignal500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		jobUUID := uuid.MustParse(jobID)
		e := resp.Error
		return gen.PostNodeProcessSignal200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.ProcessSignalResult{
				{
					Hostname: resp.Hostname,
					Status:   gen.Skipped,
					Pid:      &pid,
					Signal:   &signal,
					Error:    &e,
				},
			},
		}, nil
	}

	if resp.Status == job.StatusFailed && strings.Contains(resp.Error, "not found") {
		errMsg := resp.Error
		return gen.PostNodeProcessSignal404JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusFailed {
		errMsg := resp.Error
		return gen.PostNodeProcessSignal500JSONResponse{Error: &errMsg}, nil
	}

	var result processProv.SignalResult
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &result)
	}

	jobUUID := uuid.MustParse(jobID)
	changed := resp.Changed
	resPID := result.PID
	resSignal := result.Signal

	return gen.PostNodeProcessSignal200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.ProcessSignalResult{
			{
				Hostname: resp.Hostname,
				Status:   gen.Ok,
				Pid:      &resPID,
				Signal:   &resSignal,
				Changed:  changed,
			},
		},
	}, nil
}

// postNodeProcessSignalBroadcast handles broadcast targets for process signal.
func (s *Process) postNodeProcessSignalBroadcast(
	ctx context.Context,
	target string,
	data interface{},
) (gen.PostNodeProcessSignalResponseObject, error) {
	jobID, responses, err := s.JobClient.ModifyBroadcast(
		ctx,
		target,
		"node",
		job.OperationProcessSignal,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeProcessSignal500JSONResponse{Error: &errMsg}, nil
	}

	var items []gen.ProcessSignalResult
	for host, resp := range responses {
		item := gen.ProcessSignalResult{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.Failed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.Skipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.Ok
			item.Changed = resp.Changed

			var result processProv.SignalResult
			if resp.Data != nil {
				_ = json.Unmarshal(resp.Data, &result)
			}
			resPID := result.PID
			resSignal := result.Signal
			item.Pid = &resPID
			item.Signal = &resSignal
		}
		items = append(items, item)
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.PostNodeProcessSignal200JSONResponse{
		JobId:   &jobUUID,
		Results: items,
	}, nil
}

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

package log

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/node/log/gen"
	"github.com/retr0h/osapi/internal/job"
	logProv "github.com/retr0h/osapi/internal/provider/node/log"
)

// unitQueryPayload is the JSON payload sent to the agent for unit log queries.
type unitQueryPayload struct {
	Unit     string `json:"unit"`
	Lines    int    `json:"lines,omitempty"`
	Since    string `json:"since,omitempty"`
	Priority string `json:"priority,omitempty"`
}

// GetNodeLogUnit returns log entries for a specific systemd unit from a target node.
func (s *Log) GetNodeLogUnit(
	ctx context.Context,
	request gen.GetNodeLogUnitRequestObject,
) (gen.GetNodeLogUnitResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeLogUnit500JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	s.logger.Debug("log unit query",
		slog.String("target", hostname),
		slog.String("unit", request.Name),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	payload := unitQueryPayload{
		Unit: request.Name,
	}
	if request.Params.Lines != nil {
		payload.Lines = *request.Params.Lines
	}
	if request.Params.Since != nil {
		payload.Since = *request.Params.Since
	}
	if request.Params.Priority != nil {
		payload.Priority = *request.Params.Priority
	}

	data, err := json.Marshal(payload)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeLogUnit500JSONResponse{Error: &errMsg}, nil
	}

	if job.IsBroadcastTarget(hostname) {
		return s.getNodeLogUnitBroadcast(ctx, hostname, data)
	}

	jobID, resp, err := s.JobClient.Query(ctx, hostname, "node", job.OperationLogQueryUnit, data)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeLogUnit500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		e := resp.Error
		jobUUID := uuid.MustParse(jobID)
		return gen.GetNodeLogUnit200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.LogResultEntry{
				{
					Hostname: resp.Hostname,
					Status:   gen.Skipped,
					Error:    &e,
				},
			},
		}, nil
	}

	entries := logEntriesFromUnitResponse(resp)
	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeLogUnit200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.LogResultEntry{
			{
				Hostname: resp.Hostname,
				Status:   gen.Ok,
				Entries:  &entries,
			},
		},
	}, nil
}

// logEntriesFromUnitResponse extracts LogEntryInfo slice from a unit job response.
func logEntriesFromUnitResponse(
	resp *job.Response,
) []gen.LogEntryInfo {
	var entries []logProv.Entry
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &entries)
	}

	result := make([]gen.LogEntryInfo, 0, len(entries))
	for _, e := range entries {
		result = append(result, logEntryToGen(e))
	}

	return result
}

// getNodeLogUnitBroadcast handles broadcast targets for unit log query.
func (s *Log) getNodeLogUnitBroadcast(
	ctx context.Context,
	target string,
	data []byte,
) (gen.GetNodeLogUnitResponseObject, error) {
	jobID, responses, err := s.JobClient.QueryBroadcast(
		ctx,
		target,
		"node",
		job.OperationLogQueryUnit,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeLogUnit500JSONResponse{Error: &errMsg}, nil
	}

	var items []gen.LogResultEntry
	for host, resp := range responses {
		item := gen.LogResultEntry{
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
			entries := logEntriesFromUnitResponse(resp)
			item.Entries = &entries
		}
		items = append(items, item)
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.GetNodeLogUnit200JSONResponse{
		JobId:   &jobUUID,
		Results: items,
	}, nil
}

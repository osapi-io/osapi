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


// GetNodeLog returns system log entries from a target node.
func (s *Log) GetNodeLog(
	ctx context.Context,
	request gen.GetNodeLogRequestObject,
) (gen.GetNodeLogResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeLog500JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	s.logger.Debug("log query",
		slog.String("target", hostname),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	opts := logProv.QueryOpts{}
	if request.Params.Lines != nil {
		opts.Lines = *request.Params.Lines
	}
	if request.Params.Since != nil {
		opts.Since = *request.Params.Since
	}
	if request.Params.Priority != nil {
		opts.Priority = *request.Params.Priority
	}

	if job.IsBroadcastTarget(hostname) {
		return s.getNodeLogBroadcast(ctx, hostname, opts)
	}

	jobID, resp, err := s.JobClient.Query(ctx, hostname, "node", job.OperationLogQuery, opts)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeLog500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		e := resp.Error
		jobUUID := uuid.MustParse(jobID)
		return gen.GetNodeLog200JSONResponse{
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

	entries := logEntriesFromResponse(resp)
	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeLog200JSONResponse{
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

// logEntriesFromResponse extracts LogEntryInfo slice from a job response.
func logEntriesFromResponse(
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

// logEntryToGen converts a provider log.Entry to a gen.LogEntryInfo.
func logEntryToGen(
	e logProv.Entry,
) gen.LogEntryInfo {
	return gen.LogEntryInfo{
		Timestamp: stringPtrOrNil(e.Timestamp),
		Unit:      stringPtrOrNil(e.Unit),
		Priority:  stringPtrOrNil(e.Priority),
		Message:   stringPtrOrNil(e.Message),
		Pid:       intPtrOrNil(e.PID),
		Hostname:  stringPtrOrNil(e.Hostname),
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

// intPtrOrNil returns nil if the int is zero, otherwise a pointer.
func intPtrOrNil(
	i int,
) *int {
	if i == 0 {
		return nil
	}
	return &i
}

// getNodeLogBroadcast handles broadcast targets for log query.
func (s *Log) getNodeLogBroadcast(
	ctx context.Context,
	target string,
	opts logProv.QueryOpts,
) (gen.GetNodeLogResponseObject, error) {
	jobID, responses, err := s.JobClient.QueryBroadcast(
		ctx,
		target,
		"node",
		job.OperationLogQuery,
		opts,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeLog500JSONResponse{Error: &errMsg}, nil
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
			entries := logEntriesFromResponse(resp)
			item.Entries = &entries
		}
		items = append(items, item)
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.GetNodeLog200JSONResponse{
		JobId:   &jobUUID,
		Results: items,
	}, nil
}

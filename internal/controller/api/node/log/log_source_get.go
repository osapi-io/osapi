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
)

// GetNodeLogSource returns unique syslog identifiers from a target node.
func (s *Log) GetNodeLogSource(
	ctx context.Context,
	request gen.GetNodeLogSourceRequestObject,
) (gen.GetNodeLogSourceResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeLogSource500JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	s.logger.Debug("log list sources",
		slog.String("target", hostname),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.getNodeLogSourceBroadcast(ctx, hostname)
	}

	jobID, resp, err := s.JobClient.Query(ctx, hostname, "node", job.OperationLogSources, nil)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeLogSource500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		e := resp.Error
		jobUUID := uuid.MustParse(jobID)

		return gen.GetNodeLogSource200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.LogSourceEntry{
				{
					Hostname: resp.Hostname,
					Status:   gen.LogSourceEntryStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	sources := logSourcesFromResponse(resp)
	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeLogSource200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.LogSourceEntry{
			{
				Hostname: resp.Hostname,
				Status:   gen.LogSourceEntryStatusOk,
				Sources:  &sources,
			},
		},
	}, nil
}

// logSourcesFromResponse extracts a []string of syslog identifiers from a job
// response.
func logSourcesFromResponse(
	resp *job.Response,
) []string {
	var sources []string
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &sources)
	}

	if sources == nil {
		sources = []string{}
	}

	return sources
}

// getNodeLogSourceBroadcast handles broadcast targets for log source listing.
func (s *Log) getNodeLogSourceBroadcast(
	ctx context.Context,
	target string,
) (gen.GetNodeLogSourceResponseObject, error) {
	jobID, responses, err := s.JobClient.QueryBroadcast(
		ctx,
		target,
		"node",
		job.OperationLogSources,
		nil,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeLogSource500JSONResponse{Error: &errMsg}, nil
	}

	var items []gen.LogSourceEntry
	for host, resp := range responses {
		item := gen.LogSourceEntry{
			Hostname: host,
		}

		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.LogSourceEntryStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.LogSourceEntryStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.LogSourceEntryStatusOk
			sources := logSourcesFromResponse(resp)
			item.Sources = &sources
		}

		items = append(items, item)
	}

	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeLogSource200JSONResponse{
		JobId:   &jobUUID,
		Results: items,
	}, nil
}

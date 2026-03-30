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

package sysctl

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/sysctl/gen"
	"github.com/retr0h/osapi/internal/job"
	sysctlProv "github.com/retr0h/osapi/internal/provider/node/sysctl"
)

// GetNodeSysctl lists all managed sysctl entries on a target node.
func (s *Sysctl) GetNodeSysctl(
	ctx context.Context,
	request gen.GetNodeSysctlRequestObject,
) (gen.GetNodeSysctlResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeSysctl500JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	s.logger.Debug("sysctl list",
		slog.String("target", hostname),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.getNodeSysctlBroadcast(ctx, hostname)
	}

	jobID, resp, err := s.JobClient.Query(ctx, hostname, "node", job.OperationSysctlList, nil)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeSysctl500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		e := resp.Error
		jobUUID := uuid.MustParse(jobID)
		return gen.GetNodeSysctl200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.SysctlEntry{
				{
					Hostname: resp.Hostname,
					Status:   gen.SysctlEntryStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	results := responseToSysctlEntries(resp)
	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeSysctl200JSONResponse{
		JobId:   &jobUUID,
		Results: results,
	}, nil
}

// getNodeSysctlBroadcast handles broadcast targets for sysctl list.
func (s *Sysctl) getNodeSysctlBroadcast(
	ctx context.Context,
	target string,
) (gen.GetNodeSysctlResponseObject, error) {
	jobID, responses, err := s.JobClient.QueryBroadcast(
		ctx,
		target,
		"node",
		job.OperationSysctlList,
		nil,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeSysctl500JSONResponse{Error: &errMsg}, nil
	}

	allResults := make([]gen.SysctlEntry, 0)
	for host, resp := range responses {
		switch resp.Status {
		case job.StatusFailed:
			e := resp.Error
			h := host
			allResults = append(allResults, gen.SysctlEntry{
				Hostname: h,
				Status:   gen.SysctlEntryStatusFailed,
				Error:    &e,
			})
		case job.StatusSkipped:
			e := resp.Error
			h := host
			allResults = append(allResults, gen.SysctlEntry{
				Hostname: h,
				Status:   gen.SysctlEntryStatusSkipped,
				Error:    &e,
			})
		default:
			allResults = append(allResults, responseToSysctlEntries(resp)...)
		}
	}

	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeSysctl200JSONResponse{
		JobId:   &jobUUID,
		Results: allResults,
	}, nil
}

// responseToSysctlEntries converts a job response to gen SysctlEntry slice.
func responseToSysctlEntries(
	resp *job.Response,
) []gen.SysctlEntry {
	var entries []sysctlProv.Entry
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &entries)
	}

	hostname := resp.Hostname

	results := make([]gen.SysctlEntry, 0, len(entries))
	for _, e := range entries {
		key := e.Key
		value := e.Value

		entry := gen.SysctlEntry{
			Hostname: hostname,
			Status:   gen.SysctlEntryStatusOk,
			Key:      &key,
			Value:    &value,
		}

		results = append(results, entry)
	}

	return results
}

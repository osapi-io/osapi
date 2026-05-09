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
	"strings"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/node/sysctl/gen"
	"github.com/retr0h/osapi/internal/job"
	sysctlProv "github.com/retr0h/osapi/internal/provider/node/sysctl"
)

// GetNodeSysctlByKey gets a single sysctl entry by key on a target node.
func (s *Sysctl) GetNodeSysctlByKey(
	ctx context.Context,
	request gen.GetNodeSysctlByKeyRequestObject,
) (gen.GetNodeSysctlByKeyResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeSysctlByKey400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname
	key := request.Key

	s.logger.Debug(
		"sysctl get",
		slog.String("target", hostname),
		slog.String("key", key),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.getNodeSysctlByKeyBroadcast(ctx, hostname, key)
	}

	jobID, resp, err := s.JobClient.Query(
		ctx,
		hostname,
		"node",
		job.OperationSysctlGet,
		map[string]string{"key": key},
	)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "does not exist") ||
			strings.Contains(errMsg, "not managed") {
			return gen.GetNodeSysctlByKey404JSONResponse{Error: &errMsg}, nil
		}
		return gen.GetNodeSysctlByKey500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		e := resp.Error
		jobUUID := uuid.MustParse(jobID)
		return gen.GetNodeSysctlByKey200JSONResponse{
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

	var entry sysctlProv.Entry
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &entry)
	}

	jobUUID := uuid.MustParse(jobID)
	entryKey := entry.Key
	entryValue := entry.Value
	agentHostname := resp.Hostname

	return gen.GetNodeSysctlByKey200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.SysctlEntry{
			{
				Hostname: agentHostname,
				Status:   gen.SysctlEntryStatusOk,
				Key:      &entryKey,
				Value:    &entryValue,
			},
		},
	}, nil
}

// getNodeSysctlByKeyBroadcast handles broadcast targets for sysctl get.
func (s *Sysctl) getNodeSysctlByKeyBroadcast(
	ctx context.Context,
	target string,
	key string,
) (gen.GetNodeSysctlByKeyResponseObject, error) {
	jobID, responses, err := s.JobClient.QueryBroadcast(
		ctx,
		target,
		"node",
		job.OperationSysctlGet,
		map[string]string{"key": key},
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeSysctlByKey500JSONResponse{Error: &errMsg}, nil
	}

	allResults := make([]gen.SysctlEntry, 0)
	for host, resp := range responses {
		item := gen.SysctlEntry{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.SysctlEntryStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.SysctlEntryStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.SysctlEntryStatusOk
			var entry sysctlProv.Entry
			if resp.Data != nil {
				_ = json.Unmarshal(resp.Data, &entry)
			}
			entryKey := entry.Key
			entryValue := entry.Value
			item.Key = &entryKey
			item.Value = &entryValue
		}
		allResults = append(allResults, item)
	}

	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeSysctlByKey200JSONResponse{
		JobId:   &jobUUID,
		Results: allResults,
	}, nil
}

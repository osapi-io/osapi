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

	"github.com/retr0h/osapi/internal/controller/api/node/sysctl/gen"
	"github.com/retr0h/osapi/internal/job"
	sysctlProv "github.com/retr0h/osapi/internal/provider/node/sysctl"
	"github.com/retr0h/osapi/internal/validation"
)

// PostNodeSysctl creates a sysctl parameter on a target node.
func (s *Sysctl) PostNodeSysctl(
	ctx context.Context,
	request gen.PostNodeSysctlRequestObject,
) (gen.PostNodeSysctlResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.PostNodeSysctl400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validation.Struct(request.Body); !ok {
		return gen.PostNodeSysctl400JSONResponse{Error: &errMsg}, nil
	}

	entry := sysctlProv.Entry{
		Key:   request.Body.Key,
		Value: request.Body.Value,
	}

	hostname := request.Hostname

	s.logger.Debug("sysctl create",
		slog.String("target", hostname),
		slog.String("key", entry.Key),
		slog.String("value", entry.Value),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.postNodeSysctlBroadcast(ctx, hostname, entry)
	}

	jobID, resp, err := s.JobClient.Modify(
		ctx,
		hostname,
		"node",
		job.OperationSysctlCreate,
		entry,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeSysctl500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		jobUUID := uuid.MustParse(jobID)
		e := resp.Error
		return gen.PostNodeSysctl200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.SysctlMutationResult{
				{
					Hostname: resp.Hostname,
					Status:   gen.SysctlMutationResultStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	var result sysctlProv.CreateResult
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &result)
	}

	jobUUID := uuid.MustParse(jobID)
	changed := resp.Changed
	resultKey := result.Key
	agentHostname := resp.Hostname

	return gen.PostNodeSysctl200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.SysctlMutationResult{
			{
				Hostname: agentHostname,
				Status:   gen.SysctlMutationResultStatusOk,
				Key:      &resultKey,
				Changed:  changed,
			},
		},
	}, nil
}

// postNodeSysctlBroadcast handles broadcast targets for sysctl create.
func (s *Sysctl) postNodeSysctlBroadcast(
	ctx context.Context,
	target string,
	entry sysctlProv.Entry,
) (gen.PostNodeSysctlResponseObject, error) {
	jobID, responses, err := s.JobClient.ModifyBroadcast(
		ctx,
		target,
		"node",
		job.OperationSysctlCreate,
		entry,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeSysctl500JSONResponse{Error: &errMsg}, nil
	}

	var apiResponses []gen.SysctlMutationResult
	for host, resp := range responses {
		item := gen.SysctlMutationResult{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.SysctlMutationResultStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.SysctlMutationResultStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.SysctlMutationResultStatusOk
			var result sysctlProv.CreateResult
			if resp.Data != nil {
				_ = json.Unmarshal(resp.Data, &result)
			}
			resultKey := result.Key
			item.Key = &resultKey
			item.Changed = resp.Changed
		}
		apiResponses = append(apiResponses, item)
	}

	jobUUID := uuid.MustParse(jobID)

	return gen.PostNodeSysctl200JSONResponse{
		JobId:   &jobUUID,
		Results: apiResponses,
	}, nil
}

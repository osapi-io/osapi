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

package service

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/node/service/gen"
	"github.com/retr0h/osapi/internal/job"
	serviceProv "github.com/retr0h/osapi/internal/provider/node/service"
)

// PostNodeServiceEnable enables a service on a target node.
func (s *Service) PostNodeServiceEnable(
	ctx context.Context,
	request gen.PostNodeServiceEnableRequestObject,
) (gen.PostNodeServiceEnableResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.PostNodeServiceEnable400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname
	name := request.Name

	s.logger.Debug("service enable",
		slog.String("target", hostname),
		slog.String("name", name),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.postNodeServiceEnableBroadcast(ctx, hostname, name)
	}

	jobID, resp, err := s.JobClient.Modify(
		ctx,
		hostname,
		"node",
		job.OperationServiceEnable,
		map[string]string{"name": name},
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeServiceEnable500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		jobUUID := uuid.MustParse(jobID)
		e := resp.Error
		return gen.PostNodeServiceEnable200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.ServiceMutationEntry{
				{
					Hostname: resp.Hostname,
					Status:   gen.Skipped,
					Error:    &e,
				},
			},
		}, nil
	}

	var result serviceProv.ActionResult
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &result)
	}

	jobUUID := uuid.MustParse(jobID)
	changed := resp.Changed
	resultName := result.Name
	agentHostname := resp.Hostname

	return gen.PostNodeServiceEnable200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.ServiceMutationEntry{
			{
				Hostname: agentHostname,
				Status:   gen.Ok,
				Name:     &resultName,
				Changed:  changed,
			},
		},
	}, nil
}

// postNodeServiceEnableBroadcast handles broadcast targets for service enable.
func (s *Service) postNodeServiceEnableBroadcast(
	ctx context.Context,
	target string,
	name string,
) (gen.PostNodeServiceEnableResponseObject, error) {
	jobID, responses, err := s.JobClient.ModifyBroadcast(
		ctx,
		target,
		"node",
		job.OperationServiceEnable,
		map[string]string{"name": name},
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeServiceEnable500JSONResponse{Error: &errMsg}, nil
	}

	var apiResponses []gen.ServiceMutationEntry
	for host, resp := range responses {
		item := gen.ServiceMutationEntry{
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
			var result serviceProv.ActionResult
			if resp.Data != nil {
				_ = json.Unmarshal(resp.Data, &result)
			}
			resultName := result.Name
			item.Name = &resultName
			item.Changed = resp.Changed
		}
		apiResponses = append(apiResponses, item)
	}

	jobUUID := uuid.MustParse(jobID)

	return gen.PostNodeServiceEnable200JSONResponse{
		JobId:   &jobUUID,
		Results: apiResponses,
	}, nil
}

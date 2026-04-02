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

// GetNodeService lists all services on a target node.
func (s *Service) GetNodeService(
	ctx context.Context,
	request gen.GetNodeServiceRequestObject,
) (gen.GetNodeServiceResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeService400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	s.logger.Debug("service list",
		slog.String("target", hostname),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.getNodeServiceBroadcast(ctx, hostname)
	}

	jobID, resp, err := s.JobClient.Query(
		ctx,
		hostname,
		"node",
		job.OperationServiceList,
		nil,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeService500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		e := resp.Error
		jobUUID := uuid.MustParse(jobID)
		return gen.GetNodeService200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.ServiceListEntry{
				{
					Hostname: resp.Hostname,
					Status:   gen.ServiceListEntryStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	results := responseToServiceListEntries(resp)
	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeService200JSONResponse{
		JobId:   &jobUUID,
		Results: results,
	}, nil
}

// getNodeServiceBroadcast handles broadcast targets for service list.
func (s *Service) getNodeServiceBroadcast(
	ctx context.Context,
	target string,
) (gen.GetNodeServiceResponseObject, error) {
	jobID, responses, err := s.JobClient.QueryBroadcast(
		ctx,
		target,
		"node",
		job.OperationServiceList,
		nil,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeService500JSONResponse{Error: &errMsg}, nil
	}

	allResults := make([]gen.ServiceListEntry, 0)
	for host, resp := range responses {
		switch resp.Status {
		case job.StatusFailed:
			e := resp.Error
			h := host
			allResults = append(allResults, gen.ServiceListEntry{
				Hostname: h,
				Status:   gen.ServiceListEntryStatusFailed,
				Error:    &e,
			})
		case job.StatusSkipped:
			e := resp.Error
			h := host
			allResults = append(allResults, gen.ServiceListEntry{
				Hostname: h,
				Status:   gen.ServiceListEntryStatusSkipped,
				Error:    &e,
			})
		default:
			allResults = append(allResults, responseToServiceListEntries(resp)...)
		}
	}

	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeService200JSONResponse{
		JobId:   &jobUUID,
		Results: allResults,
	}, nil
}

// responseToServiceListEntries converts a job response to gen ServiceListEntry slice.
func responseToServiceListEntries(
	resp *job.Response,
) []gen.ServiceListEntry {
	var infos []serviceProv.Info
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &infos)
	}

	hostname := resp.Hostname

	services := make([]gen.ServiceInfo, 0, len(infos))
	for _, info := range infos {
		services = append(services, serviceInfoToGen(info))
	}

	return []gen.ServiceListEntry{
		{
			Hostname: hostname,
			Status:   gen.ServiceListEntryStatusOk,
			Services: &services,
		},
	}
}

// serviceInfoToGen converts a provider Info to a gen ServiceInfo.
func serviceInfoToGen(
	info serviceProv.Info,
) gen.ServiceInfo {
	result := gen.ServiceInfo{}

	if info.Name != "" {
		name := info.Name
		result.Name = &name
	}
	if info.Status != "" {
		status := info.Status
		result.Status = &status
	}
	if info.Description != "" {
		desc := info.Description
		result.Description = &desc
	}

	enabled := info.Enabled
	result.Enabled = &enabled

	if info.PID != 0 {
		pid := info.PID
		result.Pid = &pid
	}

	return result
}

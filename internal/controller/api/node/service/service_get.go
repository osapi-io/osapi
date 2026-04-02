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
	"strings"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/node/service/gen"
	"github.com/retr0h/osapi/internal/job"
	serviceProv "github.com/retr0h/osapi/internal/provider/node/service"
)

// GetNodeServiceByName gets details for a single service on a target node.
func (s *Service) GetNodeServiceByName(
	ctx context.Context,
	request gen.GetNodeServiceByNameRequestObject,
) (gen.GetNodeServiceByNameResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeServiceByName500JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname
	name := request.Name

	s.logger.Debug("service get",
		slog.String("target", hostname),
		slog.String("name", name),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.getNodeServiceByNameBroadcast(ctx, hostname, name)
	}

	jobID, resp, err := s.JobClient.Query(
		ctx,
		hostname,
		"node",
		job.OperationServiceGet,
		map[string]string{"name": name},
	)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "does not exist") {
			return gen.GetNodeServiceByName404JSONResponse{Error: &errMsg}, nil
		}
		return gen.GetNodeServiceByName500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		e := resp.Error
		jobUUID := uuid.MustParse(jobID)
		return gen.GetNodeServiceByName200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.ServiceGetEntry{
				{
					Hostname: resp.Hostname,
					Status:   gen.ServiceGetEntryStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	results := responseToServiceGetEntries(resp)
	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeServiceByName200JSONResponse{
		JobId:   &jobUUID,
		Results: results,
	}, nil
}

// getNodeServiceByNameBroadcast handles broadcast targets for service get.
func (s *Service) getNodeServiceByNameBroadcast(
	ctx context.Context,
	target string,
	name string,
) (gen.GetNodeServiceByNameResponseObject, error) {
	jobID, responses, err := s.JobClient.QueryBroadcast(
		ctx,
		target,
		"node",
		job.OperationServiceGet,
		map[string]string{"name": name},
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeServiceByName500JSONResponse{Error: &errMsg}, nil
	}

	allResults := make([]gen.ServiceGetEntry, 0)
	for host, resp := range responses {
		switch resp.Status {
		case job.StatusFailed:
			e := resp.Error
			h := host
			allResults = append(allResults, gen.ServiceGetEntry{
				Hostname: h,
				Status:   gen.ServiceGetEntryStatusFailed,
				Error:    &e,
			})
		case job.StatusSkipped:
			e := resp.Error
			h := host
			allResults = append(allResults, gen.ServiceGetEntry{
				Hostname: h,
				Status:   gen.ServiceGetEntryStatusSkipped,
				Error:    &e,
			})
		default:
			allResults = append(allResults, responseToServiceGetEntries(resp)...)
		}
	}

	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeServiceByName200JSONResponse{
		JobId:   &jobUUID,
		Results: allResults,
	}, nil
}

// responseToServiceGetEntries converts a job response to gen ServiceGetEntry slice.
func responseToServiceGetEntries(
	resp *job.Response,
) []gen.ServiceGetEntry {
	var info serviceProv.Info
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &info)
	}

	hostname := resp.Hostname
	svc := serviceInfoToGen(info)

	return []gen.ServiceGetEntry{
		{
			Hostname: hostname,
			Status:   gen.ServiceGetEntryStatusOk,
			Service:  &svc,
		},
	}
}

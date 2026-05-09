// Copyright (c) 2024 John Dewey

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

package node

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/node/gen"
	"github.com/retr0h/osapi/internal/job"
)

// GetNodeUptime get the node uptime API endpoint.
func (s *Node) GetNodeUptime(
	ctx context.Context,
	request gen.GetNodeUptimeRequestObject,
) (gen.GetNodeUptimeResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeUptime400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	s.logger.Debug(
		"uptime get",
		slog.String("target", hostname),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.getNodeUptimeBroadcast(ctx, hostname)
	}

	jobID, rawResp, err := s.JobClient.Query(
		ctx,
		hostname,
		"node",
		job.OperationNodeUptimeGet,
		struct{}{},
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeUptime500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	if rawResp.Status == job.StatusSkipped {
		e := rawResp.Error
		jobUUID := uuid.MustParse(jobID)
		return gen.GetNodeUptime200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.UptimeResponse{
				{
					Hostname: rawResp.Hostname,
					Status:   gen.UptimeResponseStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	var uptimeResp job.NodeUptimeResponse
	if rawResp.Data != nil {
		_ = json.Unmarshal(rawResp.Data, &uptimeResp)
	}

	resp := buildUptimeResponse(rawResp.Hostname, &uptimeResp)
	resp.Status = gen.UptimeResponseStatusOk
	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeUptime200JSONResponse{
		JobId:   &jobUUID,
		Results: []gen.UptimeResponse{*resp},
	}, nil
}

// getNodeUptimeBroadcast handles broadcast targets (_all or label) for node uptime.
func (s *Node) getNodeUptimeBroadcast(
	ctx context.Context,
	target string,
) (gen.GetNodeUptimeResponseObject, error) {
	jobID, responses, err := s.JobClient.QueryBroadcast(
		ctx,
		target,
		"node",
		job.OperationNodeUptimeGet,
		struct{}{},
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeUptime500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	var apiResponses []gen.UptimeResponse
	for host, resp := range responses {
		item := gen.UptimeResponse{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.UptimeResponseStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.UptimeResponseStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.UptimeResponseStatusOk
			var uptimeResp job.NodeUptimeResponse
			if resp.Data != nil {
				_ = json.Unmarshal(resp.Data, &uptimeResp)
			}
			built := buildUptimeResponse(host, &uptimeResp)
			item.Uptime = built.Uptime
			item.Changed = built.Changed
		}
		apiResponses = append(apiResponses, item)
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.GetNodeUptime200JSONResponse{
		JobId:   &jobUUID,
		Results: apiResponses,
	}, nil
}

// buildUptimeResponse converts a NodeUptimeResponse to an UptimeResponse.
func buildUptimeResponse(
	hostname string,
	uptimeResp *job.NodeUptimeResponse,
) *gen.UptimeResponse {
	uptime := uptimeResp.Uptime
	changed := false

	return &gen.UptimeResponse{
		Hostname: hostname,
		Uptime:   &uptime,
		Changed:  &changed,
	}
}

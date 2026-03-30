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

package hostname

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/node/hostname/gen"
	"github.com/retr0h/osapi/internal/job"
)

// GetNodeHostname get the node hostname API endpoint.
func (s *Hostname) GetNodeHostname(
	ctx context.Context,
	request gen.GetNodeHostnameRequestObject,
) (gen.GetNodeHostnameResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeHostname400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	s.logger.Debug("routing",
		slog.String("target", hostname),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.getNodeHostnameBroadcast(ctx, hostname)
	}

	jobID, resp, err := s.JobClient.Query(
		ctx,
		hostname,
		"node",
		job.OperationNodeHostnameGet,
		struct{}{},
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeHostname500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	if resp.Status == job.StatusSkipped {
		e := resp.Error
		jobUUID := uuid.MustParse(jobID)
		return gen.GetNodeHostname200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.HostnameResponse{
				{
					Hostname: resp.Hostname,
					Status:   gen.HostnameResponseStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	var result struct {
		Hostname string            `json:"hostname"`
		Labels   map[string]string `json:"labels,omitempty"`
	}
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &result)
	}

	displayHostname := result.Hostname
	if displayHostname == "" {
		displayHostname = resp.Hostname
	}

	changed := false
	apiResp := gen.HostnameResponse{
		Hostname: displayHostname,
		Changed:  &changed,
		Status:   gen.HostnameResponseStatusOk,
	}
	if len(result.Labels) > 0 {
		apiResp.Labels = &result.Labels
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.GetNodeHostname200JSONResponse{
		JobId:   &jobUUID,
		Results: []gen.HostnameResponse{apiResp},
	}, nil
}

// getNodeHostnameBroadcast handles broadcast targets (_all or label) for node hostname.
func (s *Hostname) getNodeHostnameBroadcast(
	ctx context.Context,
	target string,
) (gen.GetNodeHostnameResponseObject, error) {
	jobID, responses, err := s.JobClient.QueryBroadcast(
		ctx,
		target,
		"node",
		job.OperationNodeHostnameGet,
		struct{}{},
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeHostname500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	var apiResponses []gen.HostnameResponse
	for host, resp := range responses {
		item := gen.HostnameResponse{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.HostnameResponseStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.HostnameResponseStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.HostnameResponseStatusOk
			var result struct {
				Hostname string            `json:"hostname"`
				Labels   map[string]string `json:"labels,omitempty"`
			}
			if resp.Data != nil {
				_ = json.Unmarshal(resp.Data, &result)
			}
			item.Hostname = result.Hostname
			changed := false
			item.Changed = &changed
			if len(result.Labels) > 0 {
				item.Labels = &result.Labels
			}
		}
		apiResponses = append(apiResponses, item)
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.GetNodeHostname200JSONResponse{
		JobId:   &jobUUID,
		Results: apiResponses,
	}, nil
}

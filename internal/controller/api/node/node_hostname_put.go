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

package node

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/node/gen"
	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/validation"
)

// PutNodeHostname put the node hostname API endpoint.
func (s *Node) PutNodeHostname(
	ctx context.Context,
	request gen.PutNodeHostnameRequestObject,
) (gen.PutNodeHostnameResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.PutNodeHostname400JSONResponse{
			Error: &errMsg,
		}, nil
	}

	if errMsg, ok := validation.Struct(request.Body); !ok {
		return gen.PutNodeHostname400JSONResponse{
			Error: &errMsg,
		}, nil
	}

	hostname := request.Hostname

	s.logger.Debug("hostname put",
		slog.String("new_hostname", request.Body.Hostname),
		slog.String("target", hostname),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.putNodeHostnameBroadcast(ctx, hostname, request.Body.Hostname)
	}

	data := map[string]any{
		"hostname": request.Body.Hostname,
	}
	jobID, rawResp, err := s.JobClient.Modify(
		ctx,
		hostname,
		"node",
		job.OperationNodeHostnameUpdate,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PutNodeHostname500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	if rawResp.Status == job.StatusSkipped {
		jobUUID := uuid.MustParse(jobID)
		e := rawResp.Error
		falseVal := false
		return gen.PutNodeHostname202JSONResponse{
			JobId: &jobUUID,
			Results: []gen.HostnameUpdateResultItem{
				{
					Hostname: rawResp.Hostname,
					Status:   gen.HostnameUpdateResultItemStatusSkipped,
					Error:    &e,
					Changed:  &falseVal,
				},
			},
		}, nil
	}

	changed := rawResp.Changed == nil || *rawResp.Changed
	jobUUID := uuid.MustParse(jobID)
	return gen.PutNodeHostname202JSONResponse{
		JobId: &jobUUID,
		Results: []gen.HostnameUpdateResultItem{
			{
				Hostname: rawResp.Hostname,
				Status:   gen.HostnameUpdateResultItemStatusOk,
				Changed:  &changed,
			},
		},
	}, nil
}

// putNodeHostnameBroadcast handles broadcast targets (_all or label) for hostname modification.
func (s *Node) putNodeHostnameBroadcast(
	ctx context.Context,
	target string,
	newHostname string,
) (gen.PutNodeHostnameResponseObject, error) {
	data := map[string]any{
		"hostname": newHostname,
	}
	jobID, responses, err := s.JobClient.ModifyBroadcast(
		ctx,
		target,
		"node",
		job.OperationNodeHostnameUpdate,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PutNodeHostname500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	var apiResponses []gen.HostnameUpdateResultItem
	for host, resp := range responses {
		item := gen.HostnameUpdateResultItem{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.HostnameUpdateResultItemStatusFailed
			e := resp.Error
			falseVal := false
			item.Error = &e
			item.Changed = &falseVal
		case job.StatusSkipped:
			item.Status = gen.HostnameUpdateResultItemStatusSkipped
			e := resp.Error
			falseVal := false
			item.Error = &e
			item.Changed = &falseVal
		default:
			item.Status = gen.HostnameUpdateResultItemStatusOk
			changed := resp.Changed == nil || *resp.Changed
			item.Changed = &changed
		}
		apiResponses = append(apiResponses, item)
	}

	jobUUID := uuid.MustParse(jobID)
	return gen.PutNodeHostname202JSONResponse{
		JobId:   &jobUUID,
		Results: apiResponses,
	}, nil
}

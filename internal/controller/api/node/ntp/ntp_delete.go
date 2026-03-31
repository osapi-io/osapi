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

package ntp

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/node/ntp/gen"
	"github.com/retr0h/osapi/internal/job"
	ntpProv "github.com/retr0h/osapi/internal/provider/node/ntp"
)

// DeleteNodeNtp deletes a managed NTP configuration on a target node.
func (s *Ntp) DeleteNodeNtp(
	ctx context.Context,
	request gen.DeleteNodeNtpRequestObject,
) (gen.DeleteNodeNtpResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.DeleteNodeNtp500JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	s.logger.Debug("ntp delete",
		slog.String("target", hostname),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.deleteNodeNtpBroadcast(ctx, hostname)
	}

	jobID, resp, err := s.JobClient.Modify(
		ctx,
		hostname,
		"node",
		job.OperationNtpDelete,
		nil,
	)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "does not exist") {
			return gen.DeleteNodeNtp404JSONResponse{Error: &errMsg}, nil
		}
		return gen.DeleteNodeNtp500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		jobUUID := uuid.MustParse(jobID)
		e := resp.Error
		return gen.DeleteNodeNtp200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.NtpMutationResult{
				{
					Hostname: resp.Hostname,
					Status:   gen.NtpMutationResultStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	var result ntpProv.DeleteResult
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &result)
	}

	jobUUID := uuid.MustParse(jobID)
	changed := resp.Changed
	agentHostname := resp.Hostname

	return gen.DeleteNodeNtp200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.NtpMutationResult{
			{
				Hostname: agentHostname,
				Status:   gen.NtpMutationResultStatusOk,
				Changed:  changed,
			},
		},
	}, nil
}

// deleteNodeNtpBroadcast handles broadcast targets for NTP delete.
func (s *Ntp) deleteNodeNtpBroadcast(
	ctx context.Context,
	target string,
) (gen.DeleteNodeNtpResponseObject, error) {
	jobID, responses, err := s.JobClient.ModifyBroadcast(
		ctx,
		target,
		"node",
		job.OperationNtpDelete,
		nil,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.DeleteNodeNtp500JSONResponse{Error: &errMsg}, nil
	}

	var apiResponses []gen.NtpMutationResult
	for host, resp := range responses {
		item := gen.NtpMutationResult{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.NtpMutationResultStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.NtpMutationResultStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.NtpMutationResultStatusOk
			item.Changed = resp.Changed
		}
		apiResponses = append(apiResponses, item)
	}

	jobUUID := uuid.MustParse(jobID)

	return gen.DeleteNodeNtp200JSONResponse{
		JobId:   &jobUUID,
		Results: apiResponses,
	}, nil
}

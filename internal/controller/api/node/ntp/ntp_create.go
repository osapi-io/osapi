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

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/node/ntp/gen"
	"github.com/retr0h/osapi/internal/job"
	ntpProv "github.com/retr0h/osapi/internal/provider/node/ntp"
	"github.com/retr0h/osapi/internal/validation"
)

// PostNodeNtp creates an NTP configuration on a target node.
func (s *Ntp) PostNodeNtp(
	ctx context.Context,
	request gen.PostNodeNtpRequestObject,
) (gen.PostNodeNtpResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.PostNodeNtp400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validation.Struct(request.Body); !ok {
		return gen.PostNodeNtp400JSONResponse{Error: &errMsg}, nil
	}

	config := ntpProv.Config{
		Servers: request.Body.Servers,
	}

	hostname := request.Hostname

	s.logger.Debug("ntp create",
		slog.String("target", hostname),
		slog.Any("servers", config.Servers),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.postNodeNtpBroadcast(ctx, hostname, config)
	}

	jobID, resp, err := s.JobClient.Modify(
		ctx,
		hostname,
		"node",
		job.OperationNtpCreate,
		config,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeNtp500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		jobUUID := uuid.MustParse(jobID)
		e := resp.Error
		return gen.PostNodeNtp200JSONResponse{
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

	var result ntpProv.CreateResult
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &result)
	}

	jobUUID := uuid.MustParse(jobID)
	changed := resp.Changed
	agentHostname := resp.Hostname

	return gen.PostNodeNtp200JSONResponse{
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

// postNodeNtpBroadcast handles broadcast targets for NTP create.
func (s *Ntp) postNodeNtpBroadcast(
	ctx context.Context,
	target string,
	config ntpProv.Config,
) (gen.PostNodeNtpResponseObject, error) {
	jobID, responses, err := s.JobClient.ModifyBroadcast(
		ctx,
		target,
		"node",
		job.OperationNtpCreate,
		config,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeNtp500JSONResponse{Error: &errMsg}, nil
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

	return gen.PostNodeNtp200JSONResponse{
		JobId:   &jobUUID,
		Results: apiResponses,
	}, nil
}

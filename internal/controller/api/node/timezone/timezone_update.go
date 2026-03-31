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

package timezone

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/node/timezone/gen"
	"github.com/retr0h/osapi/internal/job"
	tzProv "github.com/retr0h/osapi/internal/provider/node/timezone"
	"github.com/retr0h/osapi/internal/validation"
)

// PutNodeTimezone updates the system timezone on a target node.
func (s *Timezone) PutNodeTimezone(
	ctx context.Context,
	request gen.PutNodeTimezoneRequestObject,
) (gen.PutNodeTimezoneResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.PutNodeTimezone400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validation.Struct(request.Body); !ok {
		return gen.PutNodeTimezone400JSONResponse{Error: &errMsg}, nil
	}

	data := map[string]string{
		"timezone": request.Body.Timezone,
	}

	hostname := request.Hostname

	s.logger.Debug("timezone update",
		slog.String("target", hostname),
		slog.String("timezone", request.Body.Timezone),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.putNodeTimezoneBroadcast(ctx, hostname, data)
	}

	jobID, resp, err := s.JobClient.Modify(
		ctx,
		hostname,
		"node",
		job.OperationTimezoneUpdate,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PutNodeTimezone500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		jobUUID := uuid.MustParse(jobID)
		e := resp.Error
		return gen.PutNodeTimezone200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.TimezoneMutationResult{
				{
					Hostname: resp.Hostname,
					Status:   gen.TimezoneMutationResultStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	var result tzProv.UpdateResult
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &result)
	}

	jobUUID := uuid.MustParse(jobID)
	changed := resp.Changed
	agentHostname := resp.Hostname

	var tzPtr *string
	if result.Timezone != "" {
		tz := result.Timezone
		tzPtr = &tz
	}

	return gen.PutNodeTimezone200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.TimezoneMutationResult{
			{
				Hostname: agentHostname,
				Status:   gen.TimezoneMutationResultStatusOk,
				Changed:  changed,
				Timezone: tzPtr,
			},
		},
	}, nil
}

// putNodeTimezoneBroadcast handles broadcast targets for timezone update.
func (s *Timezone) putNodeTimezoneBroadcast(
	ctx context.Context,
	target string,
	data map[string]string,
) (gen.PutNodeTimezoneResponseObject, error) {
	jobID, responses, err := s.JobClient.ModifyBroadcast(
		ctx,
		target,
		"node",
		job.OperationTimezoneUpdate,
		data,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PutNodeTimezone500JSONResponse{Error: &errMsg}, nil
	}

	var apiResponses []gen.TimezoneMutationResult
	for host, resp := range responses {
		item := gen.TimezoneMutationResult{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.TimezoneMutationResultStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.TimezoneMutationResultStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.TimezoneMutationResultStatusOk
			item.Changed = resp.Changed
		}
		apiResponses = append(apiResponses, item)
	}

	jobUUID := uuid.MustParse(jobID)

	return gen.PutNodeTimezone200JSONResponse{
		JobId:   &jobUUID,
		Results: apiResponses,
	}, nil
}

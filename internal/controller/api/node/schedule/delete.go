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

package schedule

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"

	"github.com/osapi-io/osapi/internal/controller/api/node/schedule/gen"
	"github.com/osapi-io/osapi/internal/job"
	cronProv "github.com/osapi-io/osapi/internal/provider/schedule/cron"

	"github.com/osapi-io/osapi/internal/controller/api/apierr"
)

// DeleteNodeSchedule deletes a cron entry on a target node.
func (s *Schedule) DeleteNodeSchedule(
	ctx context.Context,
	request gen.DeleteNodeScheduleRequestObject,
) (gen.DeleteNodeScheduleResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.DeleteNodeSchedule400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validateName(request.Name); !ok {
		return gen.DeleteNodeSchedule400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname
	name := request.Name

	s.logger.Debug(
		"cron delete",
		slog.String("target", hostname),
		slog.String("name", name),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.deleteNodeScheduleBroadcast(ctx, hostname, name)
	}

	jobID, resp, err := s.JobClient.Modify(
		ctx,
		hostname,
		"schedule",
		job.OperationCronDelete,
		map[string]string{"name": name},
	)
	if err != nil {
		errMsg := err.Error()
		if apierr.IsMissing(err) {
			return gen.DeleteNodeSchedule404JSONResponse{Error: &errMsg}, nil
		}
		return gen.DeleteNodeSchedule500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		jobUUID := uuid.MustParse(jobID)
		e := resp.Error
		return gen.DeleteNodeSchedule200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.ScheduleMutationResult{
				{
					Hostname: resp.Hostname,
					Status:   gen.ScheduleMutationResultStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	var result cronProv.DeleteResult
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &result)
	}

	jobUUID := uuid.MustParse(jobID)
	changed := resp.Changed
	resultName := result.Name
	agentHostname := resp.Hostname

	return gen.DeleteNodeSchedule200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.ScheduleMutationResult{
			{
				Hostname: agentHostname,
				Status:   gen.ScheduleMutationResultStatusOk,
				Name:     &resultName,
				Changed:  changed,
			},
		},
	}, nil
}

// deleteNodeScheduleBroadcast handles broadcast targets for cron delete.
func (s *Schedule) deleteNodeScheduleBroadcast(
	ctx context.Context,
	target string,
	name string,
) (gen.DeleteNodeScheduleResponseObject, error) {
	jobID, responses, err := s.JobClient.ModifyBroadcast(
		ctx,
		target,
		"schedule",
		job.OperationCronDelete,
		map[string]string{"name": name},
	)
	if err != nil {
		errMsg := err.Error()
		return gen.DeleteNodeSchedule500JSONResponse{Error: &errMsg}, nil
	}

	var apiResponses []gen.ScheduleMutationResult
	for host, resp := range responses {
		item := gen.ScheduleMutationResult{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.ScheduleMutationResultStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.ScheduleMutationResultStatusSkipped
			e := resp.Error
			item.Error = &e
		case job.StatusTimeout:
			item.Status = gen.ScheduleMutationResultStatusTimeout
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.ScheduleMutationResultStatusOk
			var result cronProv.DeleteResult
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

	return gen.DeleteNodeSchedule200JSONResponse{
		JobId:   &jobUUID,
		Results: apiResponses,
	}, nil
}

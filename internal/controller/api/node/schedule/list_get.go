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
)

// GetNodeSchedule lists all cron entries on a target node.
func (s *Schedule) GetNodeSchedule(
	ctx context.Context,
	request gen.GetNodeScheduleRequestObject,
) (gen.GetNodeScheduleResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeSchedule400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	s.logger.Debug(
		"cron list",
		slog.String("target", hostname),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.getNodeScheduleBroadcast(ctx, hostname)
	}

	jobID, resp, err := s.JobClient.Query(ctx, hostname, "schedule", job.OperationCronList, nil)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeSchedule500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		e := resp.Error
		jobUUID := uuid.MustParse(jobID)
		return gen.GetNodeSchedule200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.ScheduleEntry{
				{
					Hostname: resp.Hostname,
					Status:   gen.ScheduleEntryStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	results := responseToCronEntries(resp)
	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeSchedule200JSONResponse{
		JobId:   &jobUUID,
		Results: results,
	}, nil
}

// getNodeScheduleBroadcast handles broadcast targets for cron list.
func (s *Schedule) getNodeScheduleBroadcast(
	ctx context.Context,
	target string,
) (gen.GetNodeScheduleResponseObject, error) {
	jobID, responses, err := s.JobClient.QueryBroadcast(
		ctx,
		target,
		"schedule",
		job.OperationCronList,
		nil,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeSchedule500JSONResponse{Error: &errMsg}, nil
	}

	allResults := make([]gen.ScheduleEntry, 0)
	for host, resp := range responses {
		switch resp.Status {
		case job.StatusFailed:
			e := resp.Error
			h := host
			allResults = append(allResults, gen.ScheduleEntry{
				Hostname: h,
				Status:   gen.ScheduleEntryStatusFailed,
				Error:    &e,
			})
		case job.StatusSkipped:
			e := resp.Error
			h := host
			allResults = append(allResults, gen.ScheduleEntry{
				Hostname: h,
				Status:   gen.ScheduleEntryStatusSkipped,
				Error:    &e,
			})
		case job.StatusTimeout:
			e := resp.Error
			h := host
			allResults = append(allResults, gen.ScheduleEntry{
				Hostname: h,
				Status:   gen.ScheduleEntryStatusTimeout,
				Error:    &e,
			})
		default:
			allResults = append(allResults, responseToCronEntries(resp)...)
		}
	}

	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeSchedule200JSONResponse{
		JobId:   &jobUUID,
		Results: allResults,
	}, nil
}

// responseToCronEntries converts a job response to gen ScheduleEntry slice.
func responseToCronEntries(
	resp *job.Response,
) []gen.ScheduleEntry {
	var entries []cronProv.Entry
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &entries)
	}

	hostname := resp.Hostname

	results := make([]gen.ScheduleEntry, 0, len(entries))
	for _, e := range entries {
		name := e.Name
		object := e.Object
		source := e.Source

		entry := gen.ScheduleEntry{
			Hostname: hostname,
			Status:   gen.ScheduleEntryStatusOk,
			Name:     &name,
			Object:   &object,
			Source:   &source,
		}

		if e.Schedule != "" {
			schedule := e.Schedule
			entry.Schedule = &schedule
		}
		if e.User != "" {
			user := e.User
			entry.User = &user
		}
		if e.Interval != "" {
			interval := gen.ScheduleEntryInterval(e.Interval)
			entry.Interval = &interval
		}

		results = append(results, entry)
	}

	return results
}

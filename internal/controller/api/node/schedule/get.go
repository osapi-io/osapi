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

// GetNodeScheduleByName gets a single cron entry by name on a target node.
func (s *Schedule) GetNodeScheduleByName(
	ctx context.Context,
	request gen.GetNodeScheduleByNameRequestObject,
) (gen.GetNodeScheduleByNameResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeScheduleByName400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname
	name := request.Name

	s.logger.Debug(
		"cron get",
		slog.String("target", hostname),
		slog.String("name", name),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.getNodeScheduleByNameBroadcast(ctx, hostname, name)
	}

	jobID, resp, err := s.JobClient.Query(
		ctx,
		hostname,
		"schedule",
		job.OperationCronGet,
		map[string]string{"name": name},
	)
	if err != nil {
		errMsg := err.Error()
		if apierr.IsMissing(err) {
			return gen.GetNodeScheduleByName404JSONResponse{Error: &errMsg}, nil
		}
		return gen.GetNodeScheduleByName500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		e := resp.Error
		jobUUID := uuid.MustParse(jobID)
		return gen.GetNodeScheduleByName200JSONResponse{
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

	var entry cronProv.Entry
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &entry)
	}

	jobUUID := uuid.MustParse(jobID)
	entryName := entry.Name
	object := entry.Object
	schedule := entry.Schedule
	user := entry.User
	agentHostname := resp.Hostname

	return gen.GetNodeScheduleByName200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.ScheduleEntry{
			{
				Hostname: agentHostname,
				Status:   gen.ScheduleEntryStatusOk,
				Name:     &entryName,
				Object:   &object,
				Schedule: &schedule,
				User:     &user,
			},
		},
	}, nil
}

// getNodeScheduleByNameBroadcast handles broadcast targets for cron get.
func (s *Schedule) getNodeScheduleByNameBroadcast(
	ctx context.Context,
	target string,
	name string,
) (gen.GetNodeScheduleByNameResponseObject, error) {
	jobID, responses, err := s.JobClient.QueryBroadcast(
		ctx,
		target,
		"schedule",
		job.OperationCronGet,
		map[string]string{"name": name},
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeScheduleByName500JSONResponse{Error: &errMsg}, nil
	}

	allResults := make([]gen.ScheduleEntry, 0)
	for host, resp := range responses {
		item := gen.ScheduleEntry{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.ScheduleEntryStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.ScheduleEntryStatusSkipped
			e := resp.Error
			item.Error = &e
		case job.StatusTimeout:
			item.Status = gen.ScheduleEntryStatusTimeout
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.ScheduleEntryStatusOk
			var entry cronProv.Entry
			if resp.Data != nil {
				_ = json.Unmarshal(resp.Data, &entry)
			}
			entryName := entry.Name
			object := entry.Object
			schedule := entry.Schedule
			user := entry.User
			item.Name = &entryName
			item.Object = &object
			item.Schedule = &schedule
			item.User = &user
		}
		allResults = append(allResults, item)
	}

	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeScheduleByName200JSONResponse{
		JobId:   &jobUUID,
		Results: allResults,
	}, nil
}

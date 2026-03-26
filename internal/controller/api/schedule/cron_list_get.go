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

	"github.com/retr0h/osapi/internal/controller/api/schedule/gen"
	"github.com/retr0h/osapi/internal/job"
	cronProv "github.com/retr0h/osapi/internal/provider/scheduled/cron"
)

// GetNodeScheduleCron lists all cron entries on a target node.
func (s *Schedule) GetNodeScheduleCron(
	ctx context.Context,
	request gen.GetNodeScheduleCronRequestObject,
) (gen.GetNodeScheduleCronResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeScheduleCron500JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	s.logger.Debug("cron list",
		slog.String("target", hostname),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.getNodeScheduleCronBroadcast(ctx, hostname)
	}

	jobID, resp, err := s.JobClient.Query(ctx, hostname, "schedule", job.OperationCronList, nil)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeScheduleCron500JSONResponse{Error: &errMsg}, nil
	}

	results := responseToCronEntries(resp)
	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeScheduleCron200JSONResponse{
		JobId:   &jobUUID,
		Results: results,
	}, nil
}

// getNodeScheduleCronBroadcast handles broadcast targets for cron list.
func (s *Schedule) getNodeScheduleCronBroadcast(
	ctx context.Context,
	target string,
) (gen.GetNodeScheduleCronResponseObject, error) {
	jobID, responses, errs, err := s.JobClient.QueryBroadcast(
		ctx,
		target,
		"schedule",
		job.OperationCronList,
		nil,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeScheduleCron500JSONResponse{Error: &errMsg}, nil
	}

	allResults := make([]gen.CronEntry, 0)
	for _, resp := range responses {
		allResults = append(allResults, responseToCronEntries(resp)...)
	}
	for hostname, errMsg := range errs {
		h := hostname
		e := errMsg
		allResults = append(allResults, gen.CronEntry{
			Hostname: &h,
			Error:    &e,
		})
	}

	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeScheduleCron200JSONResponse{
		JobId:   &jobUUID,
		Results: allResults,
	}, nil
}

// responseToCronEntries converts a job response to gen CronEntry slice.
func responseToCronEntries(
	resp *job.Response,
) []gen.CronEntry {
	var entries []cronProv.Entry
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &entries)
	}

	hostname := resp.Hostname

	results := make([]gen.CronEntry, 0, len(entries))
	for _, e := range entries {
		name := e.Name
		object := e.Object
		source := e.Source

		entry := gen.CronEntry{
			Hostname: &hostname,
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
			interval := gen.CronEntryInterval(e.Interval)
			entry.Interval = &interval
		}

		results = append(results, entry)
	}

	return results
}

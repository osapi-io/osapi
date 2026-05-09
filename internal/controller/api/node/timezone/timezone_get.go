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
)

// GetNodeTimezone gets the current system timezone on a target node.
func (s *Timezone) GetNodeTimezone(
	ctx context.Context,
	request gen.GetNodeTimezoneRequestObject,
) (gen.GetNodeTimezoneResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeTimezone400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	s.logger.Debug(
		"timezone get",
		slog.String("target", hostname),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.getNodeTimezoneBroadcast(ctx, hostname)
	}

	jobID, resp, err := s.JobClient.Query(
		ctx,
		hostname,
		"node",
		job.OperationTimezoneGet,
		nil,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeTimezone500JSONResponse{Error: &errMsg}, nil
	}

	if resp.Status == job.StatusSkipped {
		e := resp.Error
		jobUUID := uuid.MustParse(jobID)
		return gen.GetNodeTimezone200JSONResponse{
			JobId: &jobUUID,
			Results: []gen.TimezoneEntry{
				{
					Hostname: resp.Hostname,
					Status:   gen.TimezoneEntryStatusSkipped,
					Error:    &e,
				},
			},
		}, nil
	}

	var info tzProv.Info
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &info)
	}

	jobUUID := uuid.MustParse(jobID)
	agentHostname := resp.Hostname

	return gen.GetNodeTimezone200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.TimezoneEntry{
			infoToEntry(agentHostname, &info),
		},
	}, nil
}

// getNodeTimezoneBroadcast handles broadcast targets for timezone get.
func (s *Timezone) getNodeTimezoneBroadcast(
	ctx context.Context,
	target string,
) (gen.GetNodeTimezoneResponseObject, error) {
	jobID, responses, err := s.JobClient.QueryBroadcast(
		ctx,
		target,
		"node",
		job.OperationTimezoneGet,
		nil,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNodeTimezone500JSONResponse{Error: &errMsg}, nil
	}

	allResults := make([]gen.TimezoneEntry, 0)
	for host, resp := range responses {
		item := gen.TimezoneEntry{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.TimezoneEntryStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.TimezoneEntryStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			var info tzProv.Info
			if resp.Data != nil {
				_ = json.Unmarshal(resp.Data, &info)
			}
			item = infoToEntry(host, &info)
		}
		allResults = append(allResults, item)
	}

	jobUUID := uuid.MustParse(jobID)

	return gen.GetNodeTimezone200JSONResponse{
		JobId:   &jobUUID,
		Results: allResults,
	}, nil
}

// infoToEntry converts a timezone Info to a gen TimezoneEntry.
func infoToEntry(
	hostname string,
	info *tzProv.Info,
) gen.TimezoneEntry {
	entry := gen.TimezoneEntry{
		Hostname: hostname,
		Status:   gen.TimezoneEntryStatusOk,
	}

	if info.Timezone != "" {
		tz := info.Timezone
		entry.Timezone = &tz
	}

	if info.UTCOffset != "" {
		offset := info.UTCOffset
		entry.UtcOffset = &offset
	}

	return entry
}

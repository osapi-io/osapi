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
	"strings"

	"github.com/google/uuid"

	"github.com/retr0h/osapi/internal/controller/api/schedule/gen"
	"github.com/retr0h/osapi/internal/job"
	cronProv "github.com/retr0h/osapi/internal/provider/scheduled/cron"
	"github.com/retr0h/osapi/internal/validation"
)

// PutNodeScheduleCron updates a cron entry on a target node.
func (s *Schedule) PutNodeScheduleCron(
	ctx context.Context,
	request gen.PutNodeScheduleCronRequestObject,
) (gen.PutNodeScheduleCronResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.PutNodeScheduleCron400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validation.Struct(request.Body); !ok {
		return gen.PutNodeScheduleCron400JSONResponse{Error: &errMsg}, nil
	}

	entry := cronProv.Entry{
		Name: request.Name,
	}
	if request.Body.Object != nil {
		entry.Object = *request.Body.Object
	}
	if request.Body.Schedule != nil {
		entry.Schedule = *request.Body.Schedule
	}
	if request.Body.User != nil {
		entry.User = *request.Body.User
	}
	if request.Body.ContentType != nil {
		entry.ContentType = string(*request.Body.ContentType)
	}
	if request.Body.Vars != nil {
		entry.Vars = *request.Body.Vars
	}

	hostname := request.Hostname

	s.logger.Debug("cron update",
		slog.String("target", hostname),
		slog.String("name", entry.Name),
		slog.Bool("broadcast", job.IsBroadcastTarget(hostname)),
	)

	if job.IsBroadcastTarget(hostname) {
		return s.putNodeScheduleCronUpdateBroadcast(ctx, hostname, entry)
	}

	jobID, resp, err := s.JobClient.Modify(
		ctx,
		hostname,
		"schedule",
		job.OperationCronUpdate,
		entry,
	)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "does not exist") {
			return gen.PutNodeScheduleCron404JSONResponse{Error: &errMsg}, nil
		}
		return gen.PutNodeScheduleCron500JSONResponse{Error: &errMsg}, nil
	}

	var result cronProv.UpdateResult
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &result)
	}

	jobUUID := uuid.MustParse(jobID)
	changed := resp.Changed
	name := result.Name
	agentHostname := resp.Hostname

	return gen.PutNodeScheduleCron200JSONResponse{
		JobId: &jobUUID,
		Results: []gen.CronMutationResult{
			{
				Hostname: agentHostname,
				Status:   gen.CronMutationResultStatusOk,
				Name:     &name,
				Changed:  changed,
			},
		},
	}, nil
}

// putNodeScheduleCronUpdateBroadcast handles broadcast targets for cron update.
func (s *Schedule) putNodeScheduleCronUpdateBroadcast(
	ctx context.Context,
	target string,
	entry cronProv.Entry,
) (gen.PutNodeScheduleCronResponseObject, error) {
	jobID, responses, err := s.JobClient.ModifyBroadcast(
		ctx,
		target,
		"schedule",
		job.OperationCronUpdate,
		entry,
	)
	if err != nil {
		errMsg := err.Error()
		return gen.PutNodeScheduleCron500JSONResponse{Error: &errMsg}, nil
	}

	var apiResponses []gen.CronMutationResult
	for host, resp := range responses {
		item := gen.CronMutationResult{
			Hostname: host,
		}
		switch resp.Status {
		case job.StatusFailed:
			item.Status = gen.CronMutationResultStatusFailed
			e := resp.Error
			item.Error = &e
		case job.StatusSkipped:
			item.Status = gen.CronMutationResultStatusSkipped
			e := resp.Error
			item.Error = &e
		default:
			item.Status = gen.CronMutationResultStatusOk
			var result cronProv.UpdateResult
			if resp.Data != nil {
				_ = json.Unmarshal(resp.Data, &result)
			}
			name := result.Name
			item.Name = &name
			item.Changed = resp.Changed
		}
		apiResponses = append(apiResponses, item)
	}

	jobUUID := uuid.MustParse(jobID)

	return gen.PutNodeScheduleCron200JSONResponse{
		JobId:   &jobUUID,
		Results: apiResponses,
	}, nil
}

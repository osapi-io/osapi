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
	cronProv "github.com/retr0h/osapi/internal/provider/scheduled/cron"
	"github.com/retr0h/osapi/internal/validation"
)

// PostNodeScheduleCron creates a cron entry on a target node.
func (s *Schedule) PostNodeScheduleCron(
	ctx context.Context,
	request gen.PostNodeScheduleCronRequestObject,
) (gen.PostNodeScheduleCronResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.PostNodeScheduleCron400JSONResponse{Error: &errMsg}, nil
	}

	if errMsg, ok := validation.Struct(request.Body); !ok {
		return gen.PostNodeScheduleCron400JSONResponse{Error: &errMsg}, nil
	}

	entry := cronProv.Entry{
		Name:    request.Body.Name,
		Command: request.Body.Command,
	}
	if request.Body.Schedule != nil {
		entry.Schedule = *request.Body.Schedule
	}
	if request.Body.Interval != nil {
		entry.Interval = string(*request.Body.Interval)
	}
	if request.Body.User != nil {
		entry.User = *request.Body.User
	}

	hostname := request.Hostname

	s.logger.Debug("cron create",
		slog.String("target", hostname),
		slog.String("name", entry.Name),
	)

	resp, err := s.JobClient.ModifyScheduleCronCreate(ctx, hostname, entry)
	if err != nil {
		errMsg := err.Error()
		return gen.PostNodeScheduleCron500JSONResponse{Error: &errMsg}, nil
	}

	var result cronProv.CreateResult
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &result)
	}

	jobUUID := uuid.MustParse(resp.JobID)
	changed := resp.Changed
	name := result.Name

	return gen.PostNodeScheduleCron200JSONResponse{
		JobId:   &jobUUID,
		Changed: changed,
		Name:    &name,
	}, nil
}

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
	cronProv "github.com/retr0h/osapi/internal/provider/scheduled/cron"
)

// GetNodeScheduleCronByName gets a single cron entry by name on a target node.
func (s *Schedule) GetNodeScheduleCronByName(
	ctx context.Context,
	request gen.GetNodeScheduleCronByNameRequestObject,
) (gen.GetNodeScheduleCronByNameResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.GetNodeScheduleCronByName500JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname
	name := request.Name

	s.logger.Debug("cron get",
		slog.String("target", hostname),
		slog.String("name", name),
	)

	resp, err := s.JobClient.QueryScheduleCronGet(ctx, hostname, name)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "does not exist") ||
			strings.Contains(errMsg, "not managed") {
			return gen.GetNodeScheduleCronByName404JSONResponse{Error: &errMsg}, nil
		}
		return gen.GetNodeScheduleCronByName500JSONResponse{Error: &errMsg}, nil
	}

	var entry cronProv.Entry
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &entry)
	}

	jobUUID := uuid.MustParse(resp.JobID)
	entryName := entry.Name
	schedule := entry.Schedule
	user := entry.User
	command := entry.Command

	return gen.GetNodeScheduleCronByName200JSONResponse{
		JobId:    &jobUUID,
		Name:     &entryName,
		Schedule: &schedule,
		User:     &user,
		Command:  &command,
	}, nil
}

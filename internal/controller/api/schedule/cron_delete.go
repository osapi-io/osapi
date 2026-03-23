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

// DeleteNodeScheduleCron deletes a cron entry on a target node.
func (s *Schedule) DeleteNodeScheduleCron(
	ctx context.Context,
	request gen.DeleteNodeScheduleCronRequestObject,
) (gen.DeleteNodeScheduleCronResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.DeleteNodeScheduleCron500JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname
	name := request.Name

	s.logger.Debug("cron delete",
		slog.String("target", hostname),
		slog.String("name", name),
	)

	resp, err := s.JobClient.ModifyScheduleCronDelete(ctx, hostname, name)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "does not exist") {
			return gen.DeleteNodeScheduleCron404JSONResponse{Error: &errMsg}, nil
		}
		return gen.DeleteNodeScheduleCron500JSONResponse{Error: &errMsg}, nil
	}

	var result cronProv.DeleteResult
	if resp.Data != nil {
		_ = json.Unmarshal(resp.Data, &result)
	}

	jobUUID := uuid.MustParse(resp.JobID)
	changed := resp.Changed
	resultName := result.Name

	return gen.DeleteNodeScheduleCron200JSONResponse{
		JobId:   &jobUUID,
		Changed: changed,
		Name:    &resultName,
	}, nil
}

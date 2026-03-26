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

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider/scheduled/cron"
)

// NewScheduleProcessor returns a ProcessorFunc that handles schedule-related operations.
func NewScheduleProcessor(
	cronProvider cron.Provider,
	logger *slog.Logger,
) ProcessorFunc {
	return func(req job.Request) (json.RawMessage, error) {
		if cronProvider == nil {
			return nil, fmt.Errorf("cron provider not available")
		}

		// Extract base operation from dotted operation (e.g., "cron.list" -> "cron")
		baseOperation := strings.Split(req.Operation, ".")[0]

		switch baseOperation {
		case "cron":
			return processCronOperation(cronProvider, logger, req)
		default:
			return nil, fmt.Errorf("unsupported schedule operation: %s", req.Operation)
		}
	}
}

// processCronOperation dispatches cron sub-operations.
func processCronOperation(
	cronProvider cron.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	// Extract sub-operation: "cron.list" -> "list"
	parts := strings.Split(jobRequest.Operation, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid cron operation: %s", jobRequest.Operation)
	}
	subOp := parts[1]

	ctx := context.Background()

	switch subOp {
	case "list":
		return processCronList(ctx, cronProvider, logger)
	case "get":
		return processCronGet(ctx, cronProvider, logger, jobRequest)
	case "create":
		return processCronCreate(ctx, cronProvider, logger, jobRequest)
	case "update":
		return processCronUpdate(ctx, cronProvider, logger, jobRequest)
	case "delete":
		return processCronDelete(ctx, cronProvider, logger, jobRequest)
	default:
		return nil, fmt.Errorf("unsupported cron operation: %s", jobRequest.Operation)
	}
}

// processCronList lists all cron entries.
func processCronList(
	ctx context.Context,
	cronProvider cron.Provider,
	logger *slog.Logger,
) (json.RawMessage, error) {
	logger.Debug("executing cron.List")

	entries, err := cronProvider.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list cron entries: %w", err)
	}

	return json.Marshal(entries)
}

// processCronGet gets a single cron entry by name.
func processCronGet(
	ctx context.Context,
	cronProvider cron.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal cron get data: %w", err)
	}

	logger.Debug("executing cron.Get",
		slog.String("name", data.Name),
	)

	entry, err := cronProvider.Get(ctx, data.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get cron entry: %w", err)
	}

	return json.Marshal(entry)
}

// processCronCreate creates a new cron entry.
func processCronCreate(
	ctx context.Context,
	cronProvider cron.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var entry cron.Entry
	if err := json.Unmarshal(jobRequest.Data, &entry); err != nil {
		return nil, fmt.Errorf("unmarshal cron create data: %w", err)
	}

	logger.Debug("executing cron.Create",
		slog.String("name", entry.Name),
	)

	result, err := cronProvider.Create(ctx, entry)
	if err != nil {
		return nil, fmt.Errorf("failed to create cron entry: %w", err)
	}

	return json.Marshal(result)
}

// processCronUpdate updates an existing cron entry.
func processCronUpdate(
	ctx context.Context,
	cronProvider cron.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var entry cron.Entry
	if err := json.Unmarshal(jobRequest.Data, &entry); err != nil {
		return nil, fmt.Errorf("unmarshal cron update data: %w", err)
	}

	logger.Debug("executing cron.Update",
		slog.String("name", entry.Name),
	)

	result, err := cronProvider.Update(ctx, entry)
	if err != nil {
		return nil, fmt.Errorf("failed to update cron entry: %w", err)
	}

	return json.Marshal(result)
}

// processCronDelete deletes a cron entry.
func processCronDelete(
	ctx context.Context,
	cronProvider cron.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal cron delete data: %w", err)
	}

	logger.Debug("executing cron.Delete",
		slog.String("name", data.Name),
	)

	result, err := cronProvider.Delete(ctx, data.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to delete cron entry: %w", err)
	}

	return json.Marshal(result)
}

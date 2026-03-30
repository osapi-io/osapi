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
	"github.com/retr0h/osapi/internal/provider/node/sysctl"
)

// processSysctlOperation dispatches sysctl sub-operations.
func processSysctlOperation(
	sysctlProvider sysctl.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	if sysctlProvider == nil {
		return nil, fmt.Errorf("sysctl provider not available")
	}

	// Extract sub-operation: "sysctl.list" -> "list"
	parts := strings.Split(jobRequest.Operation, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid sysctl operation: %s", jobRequest.Operation)
	}
	subOp := parts[1]

	ctx := context.Background()

	switch subOp {
	case "list":
		return processSysctlList(ctx, sysctlProvider, logger)
	case "get":
		return processSysctlGet(ctx, sysctlProvider, logger, jobRequest)
	case "create":
		return processSysctlCreate(ctx, sysctlProvider, logger, jobRequest)
	case "update":
		return processSysctlUpdate(ctx, sysctlProvider, logger, jobRequest)
	case "delete":
		return processSysctlDelete(ctx, sysctlProvider, logger, jobRequest)
	default:
		return nil, fmt.Errorf("unsupported sysctl operation: %s", jobRequest.Operation)
	}
}

// processSysctlList lists all sysctl entries.
func processSysctlList(
	ctx context.Context,
	sysctlProvider sysctl.Provider,
	logger *slog.Logger,
) (json.RawMessage, error) {
	logger.Debug("executing sysctl.List")

	entries, err := sysctlProvider.List(ctx)
	if err != nil {
		return nil, err
	}

	return json.Marshal(entries)
}

// processSysctlGet gets a single sysctl entry by key.
func processSysctlGet(
	ctx context.Context,
	sysctlProvider sysctl.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		Key string `json:"key"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal sysctl get data: %w", err)
	}

	logger.Debug("executing sysctl.Get",
		slog.String("key", data.Key),
	)

	entry, err := sysctlProvider.Get(ctx, data.Key)
	if err != nil {
		return nil, err
	}

	return json.Marshal(entry)
}

// processSysctlCreate creates a sysctl entry.
func processSysctlCreate(
	ctx context.Context,
	sysctlProvider sysctl.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var entry sysctl.Entry
	if err := json.Unmarshal(jobRequest.Data, &entry); err != nil {
		return nil, fmt.Errorf("unmarshal sysctl create data: %w", err)
	}

	logger.Debug("executing sysctl.Create",
		slog.String("key", entry.Key),
	)

	result, err := sysctlProvider.Create(ctx, entry)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processSysctlUpdate updates a sysctl entry.
func processSysctlUpdate(
	ctx context.Context,
	sysctlProvider sysctl.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var entry sysctl.Entry
	if err := json.Unmarshal(jobRequest.Data, &entry); err != nil {
		return nil, fmt.Errorf("unmarshal sysctl update data: %w", err)
	}

	logger.Debug("executing sysctl.Update",
		slog.String("key", entry.Key),
	)

	result, err := sysctlProvider.Update(ctx, entry)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processSysctlDelete deletes a sysctl entry by key.
func processSysctlDelete(
	ctx context.Context,
	sysctlProvider sysctl.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		Key string `json:"key"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal sysctl delete data: %w", err)
	}

	logger.Debug("executing sysctl.Delete",
		slog.String("key", data.Key),
	)

	result, err := sysctlProvider.Delete(ctx, data.Key)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

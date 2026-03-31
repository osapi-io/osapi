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
	"github.com/retr0h/osapi/internal/provider/node/process"
)

// processProcessOperation dispatches process management sub-operations.
func processProcessOperation(
	processProvider process.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	if processProvider == nil {
		return nil, fmt.Errorf("process provider not available")
	}

	// Extract sub-operation: "process.list" -> "list"
	parts := strings.Split(jobRequest.Operation, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid process operation: %s", jobRequest.Operation)
	}
	subOp := parts[1]

	ctx := context.Background()

	switch subOp {
	case "list":
		return processProcessList(ctx, processProvider, logger)
	case "get":
		return processProcessGet(ctx, processProvider, logger, jobRequest)
	case "signal":
		return processProcessSignal(ctx, processProvider, logger, jobRequest)
	default:
		return nil, fmt.Errorf("unsupported process operation: %s", jobRequest.Operation)
	}
}

// processProcessList retrieves all running processes.
func processProcessList(
	ctx context.Context,
	processProvider process.Provider,
	logger *slog.Logger,
) (json.RawMessage, error) {
	logger.Debug("executing process.List")

	result, err := processProvider.List(ctx)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processProcessGet retrieves details for a specific process by PID.
func processProcessGet(
	ctx context.Context,
	processProvider process.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	logger.Debug("executing process.Get")

	var data struct {
		PID int `json:"pid"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal process get data: %w", err)
	}

	result, err := processProvider.Get(ctx, data.PID)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processProcessSignal sends a signal to a process by PID.
func processProcessSignal(
	ctx context.Context,
	processProvider process.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	logger.Debug("executing process.Signal")

	var data struct {
		PID    int    `json:"pid"`
		Signal string `json:"signal"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal process signal data: %w", err)
	}

	result, err := processProvider.Signal(ctx, data.PID, data.Signal)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

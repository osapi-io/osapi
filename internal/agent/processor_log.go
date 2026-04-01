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
	logProv "github.com/retr0h/osapi/internal/provider/node/log"
)

// processLogOperation dispatches log management sub-operations.
func processLogOperation(
	logProvider logProv.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	if logProvider == nil {
		return nil, fmt.Errorf("log provider not available")
	}

	// Extract sub-operation: "log.query" -> "query"
	parts := strings.Split(jobRequest.Operation, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid log operation: %s", jobRequest.Operation)
	}
	subOp := parts[1]

	ctx := context.Background()

	switch subOp {
	case "query":
		return processLogQuery(ctx, logProvider, logger, jobRequest)
	case "queryUnit":
		return processLogQueryUnit(ctx, logProvider, logger, jobRequest)
	default:
		return nil, fmt.Errorf("unsupported log operation: %s", jobRequest.Operation)
	}
}

// processLogQuery retrieves journal entries with optional filtering.
func processLogQuery(
	ctx context.Context,
	logProvider logProv.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	logger.Debug("executing log.Query")

	var opts logProv.QueryOpts
	if len(jobRequest.Data) > 0 {
		if err := json.Unmarshal(jobRequest.Data, &opts); err != nil {
			return nil, fmt.Errorf("unmarshal log query data: %w", err)
		}
	}

	result, err := logProvider.Query(ctx, opts)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processLogQueryUnit retrieves journal entries for a specific systemd unit.
func processLogQueryUnit(
	ctx context.Context,
	logProvider logProv.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	logger.Debug("executing log.QueryUnit")

	var data struct {
		Unit string `json:"unit"`
		logProv.QueryOpts
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal log query unit data: %w", err)
	}

	result, err := logProvider.QueryUnit(ctx, data.Unit, data.QueryOpts)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

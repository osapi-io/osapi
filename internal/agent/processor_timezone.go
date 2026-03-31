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
	"github.com/retr0h/osapi/internal/provider/node/timezone"
)

// processTimezoneOperation dispatches timezone sub-operations.
func processTimezoneOperation(
	timezoneProvider timezone.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	if timezoneProvider == nil {
		return nil, fmt.Errorf("timezone provider not available")
	}

	// Extract sub-operation: "timezone.get" -> "get"
	parts := strings.Split(jobRequest.Operation, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid timezone operation: %s", jobRequest.Operation)
	}
	subOp := parts[1]

	ctx := context.Background()

	switch subOp {
	case "get":
		return processTimezoneGet(ctx, timezoneProvider, logger)
	case "update":
		return processTimezoneUpdate(ctx, timezoneProvider, logger, jobRequest)
	default:
		return nil, fmt.Errorf("unsupported timezone operation: %s", jobRequest.Operation)
	}
}

// processTimezoneGet retrieves the current system timezone.
func processTimezoneGet(
	ctx context.Context,
	timezoneProvider timezone.Provider,
	logger *slog.Logger,
) (json.RawMessage, error) {
	logger.Debug("executing timezone.Get")

	info, err := timezoneProvider.Get(ctx)
	if err != nil {
		return nil, err
	}

	return json.Marshal(info)
}

// processTimezoneUpdate sets the system timezone.
func processTimezoneUpdate(
	ctx context.Context,
	timezoneProvider timezone.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		Timezone string `json:"timezone"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal timezone update data: %w", err)
	}

	logger.Debug("executing timezone.Update",
		slog.String("timezone", data.Timezone),
	)

	result, err := timezoneProvider.Update(ctx, data.Timezone)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

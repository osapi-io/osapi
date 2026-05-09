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
	"github.com/retr0h/osapi/internal/provider/node/ntp"
)

// processNtpOperation dispatches ntp sub-operations.
func processNtpOperation(
	ntpProvider ntp.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	if ntpProvider == nil {
		return nil, fmt.Errorf("ntp provider not available")
	}

	// Extract sub-operation: "ntp.get" -> "get"
	parts := strings.Split(jobRequest.Operation, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid ntp operation: %s", jobRequest.Operation)
	}
	subOp := parts[1]

	ctx := context.Background()

	switch subOp {
	case "get":
		return processNtpGet(ctx, ntpProvider, logger)
	case "create":
		return processNtpCreate(ctx, ntpProvider, logger, jobRequest)
	case "update":
		return processNtpUpdate(ctx, ntpProvider, logger, jobRequest)
	case "delete":
		return processNtpDelete(ctx, ntpProvider, logger)
	default:
		return nil, fmt.Errorf("unsupported ntp operation: %s", jobRequest.Operation)
	}
}

// processNtpGet retrieves the current NTP status and configured servers.
func processNtpGet(
	ctx context.Context,
	ntpProvider ntp.Provider,
	logger *slog.Logger,
) (json.RawMessage, error) {
	logger.Debug("executing ntp.Get")

	status, err := ntpProvider.Get(ctx)
	if err != nil {
		return nil, err
	}

	return json.Marshal(status)
}

// processNtpCreate deploys a managed NTP server configuration.
func processNtpCreate(
	ctx context.Context,
	ntpProvider ntp.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var config ntp.Config
	if err := json.Unmarshal(jobRequest.Data, &config); err != nil {
		return nil, fmt.Errorf("unmarshal ntp create data: %w", err)
	}

	logger.Debug(
		"executing ntp.Create",
		slog.Int("servers", len(config.Servers)),
	)

	result, err := ntpProvider.Create(ctx, config)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processNtpUpdate replaces the managed NTP server configuration.
func processNtpUpdate(
	ctx context.Context,
	ntpProvider ntp.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var config ntp.Config
	if err := json.Unmarshal(jobRequest.Data, &config); err != nil {
		return nil, fmt.Errorf("unmarshal ntp update data: %w", err)
	}

	logger.Debug(
		"executing ntp.Update",
		slog.Int("servers", len(config.Servers)),
	)

	result, err := ntpProvider.Update(ctx, config)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processNtpDelete removes the managed NTP server configuration.
func processNtpDelete(
	ctx context.Context,
	ntpProvider ntp.Provider,
	logger *slog.Logger,
) (json.RawMessage, error) {
	logger.Debug("executing ntp.Delete")

	result, err := ntpProvider.Delete(ctx)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

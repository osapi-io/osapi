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
	"github.com/retr0h/osapi/internal/provider/node/power"
)

// processPowerOperation dispatches power management sub-operations.
func processPowerOperation(
	powerProvider power.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	if powerProvider == nil {
		return nil, fmt.Errorf("power provider not available")
	}

	// Extract sub-operation: "power.reboot" -> "reboot"
	parts := strings.Split(jobRequest.Operation, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid power operation: %s", jobRequest.Operation)
	}
	subOp := parts[1]

	var opts power.Opts

	if len(jobRequest.Data) > 0 {
		if err := json.Unmarshal(jobRequest.Data, &opts); err != nil {
			return nil, fmt.Errorf("unmarshal power opts: %w", err)
		}
	}

	ctx := context.Background()

	switch subOp {
	case "reboot":
		return processPowerReboot(ctx, powerProvider, logger, opts)
	case "shutdown":
		return processPowerShutdown(ctx, powerProvider, logger, opts)
	default:
		return nil, fmt.Errorf("unsupported power operation: %s", jobRequest.Operation)
	}
}

// processPowerReboot triggers a system reboot.
func processPowerReboot(
	ctx context.Context,
	powerProvider power.Provider,
	logger *slog.Logger,
	opts power.Opts,
) (json.RawMessage, error) {
	logger.Debug("executing power.Reboot")

	result, err := powerProvider.Reboot(ctx, opts)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processPowerShutdown triggers a system shutdown.
func processPowerShutdown(
	ctx context.Context,
	powerProvider power.Provider,
	logger *slog.Logger,
	opts power.Opts,
) (json.RawMessage, error) {
	logger.Debug("executing power.Shutdown")

	result, err := powerProvider.Shutdown(ctx, opts)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

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
	"github.com/retr0h/osapi/internal/provider/node/service"
)

// processServiceOperation dispatches service management sub-operations.
func processServiceOperation(
	serviceProvider service.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	if serviceProvider == nil {
		return nil, fmt.Errorf("service provider not available")
	}

	// Extract sub-operation: "service.list" -> "list"
	parts := strings.Split(jobRequest.Operation, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid service operation: %s", jobRequest.Operation)
	}
	subOp := parts[1]

	ctx := context.Background()

	switch subOp {
	case "list":
		return processServiceList(ctx, serviceProvider, logger)
	case "get":
		return processServiceGet(ctx, serviceProvider, logger, jobRequest)
	case "create":
		return processServiceCreate(ctx, serviceProvider, logger, jobRequest)
	case "update":
		return processServiceUpdate(ctx, serviceProvider, logger, jobRequest)
	case "delete":
		return processServiceDelete(ctx, serviceProvider, logger, jobRequest)
	case "start":
		return processServiceStart(ctx, serviceProvider, logger, jobRequest)
	case "stop":
		return processServiceStop(ctx, serviceProvider, logger, jobRequest)
	case "restart":
		return processServiceRestart(ctx, serviceProvider, logger, jobRequest)
	case "enable":
		return processServiceEnable(ctx, serviceProvider, logger, jobRequest)
	case "disable":
		return processServiceDisable(ctx, serviceProvider, logger, jobRequest)
	default:
		return nil, fmt.Errorf("unsupported service operation: %s", jobRequest.Operation)
	}
}

// processServiceList lists all systemd services.
func processServiceList(
	ctx context.Context,
	serviceProvider service.Provider,
	logger *slog.Logger,
) (json.RawMessage, error) {
	logger.Debug("executing service.List")

	entries, err := serviceProvider.List(ctx)
	if err != nil {
		return nil, err
	}

	return json.Marshal(entries)
}

// processServiceGet gets a single systemd service by name.
func processServiceGet(
	ctx context.Context,
	serviceProvider service.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal service get data: %w", err)
	}

	logger.Debug(
		"executing service.Get",
		slog.String("name", data.Name),
	)

	entry, err := serviceProvider.Get(ctx, data.Name)
	if err != nil {
		return nil, err
	}

	return json.Marshal(entry)
}

// processServiceCreate creates a new unit file via the file provider.
func processServiceCreate(
	ctx context.Context,
	serviceProvider service.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var entry service.Entry
	if err := json.Unmarshal(jobRequest.Data, &entry); err != nil {
		return nil, fmt.Errorf("unmarshal service create data: %w", err)
	}

	logger.Debug(
		"executing service.Create",
		slog.String("name", entry.Name),
	)

	result, err := serviceProvider.Create(ctx, entry)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processServiceUpdate redeploys an existing unit file via the file provider.
func processServiceUpdate(
	ctx context.Context,
	serviceProvider service.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var entry service.Entry
	if err := json.Unmarshal(jobRequest.Data, &entry); err != nil {
		return nil, fmt.Errorf("unmarshal service update data: %w", err)
	}

	logger.Debug(
		"executing service.Update",
		slog.String("name", entry.Name),
	)

	result, err := serviceProvider.Update(ctx, entry)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processServiceDelete undeploys a unit file via the file provider.
func processServiceDelete(
	ctx context.Context,
	serviceProvider service.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal service delete data: %w", err)
	}

	logger.Debug(
		"executing service.Delete",
		slog.String("name", data.Name),
	)

	result, err := serviceProvider.Delete(ctx, data.Name)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processServiceStart starts a systemd service.
func processServiceStart(
	ctx context.Context,
	serviceProvider service.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal service start data: %w", err)
	}

	logger.Debug(
		"executing service.Start",
		slog.String("name", data.Name),
	)

	result, err := serviceProvider.Start(ctx, data.Name)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processServiceStop stops a systemd service.
func processServiceStop(
	ctx context.Context,
	serviceProvider service.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal service stop data: %w", err)
	}

	logger.Debug(
		"executing service.Stop",
		slog.String("name", data.Name),
	)

	result, err := serviceProvider.Stop(ctx, data.Name)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processServiceRestart restarts a systemd service.
func processServiceRestart(
	ctx context.Context,
	serviceProvider service.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal service restart data: %w", err)
	}

	logger.Debug(
		"executing service.Restart",
		slog.String("name", data.Name),
	)

	result, err := serviceProvider.Restart(ctx, data.Name)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processServiceEnable enables a systemd service.
func processServiceEnable(
	ctx context.Context,
	serviceProvider service.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal service enable data: %w", err)
	}

	logger.Debug(
		"executing service.Enable",
		slog.String("name", data.Name),
	)

	result, err := serviceProvider.Enable(ctx, data.Name)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processServiceDisable disables a systemd service.
func processServiceDisable(
	ctx context.Context,
	serviceProvider service.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal service disable data: %w", err)
	}

	logger.Debug(
		"executing service.Disable",
		slog.String("name", data.Name),
	)

	result, err := serviceProvider.Disable(ctx, data.Name)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

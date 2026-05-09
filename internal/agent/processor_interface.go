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
	"github.com/retr0h/osapi/internal/provider/network/netplan/iface"
)

// processInterfaceOperation dispatches network interface sub-operations.
func processInterfaceOperation(
	provider iface.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	if provider == nil {
		return nil, fmt.Errorf("interface provider not available")
	}

	// Extract sub-operation: "interface.list" -> "list"
	parts := strings.Split(jobRequest.Operation, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid interface operation: %s", jobRequest.Operation)
	}
	subOp := parts[1]

	ctx := context.Background()

	switch subOp {
	case "list":
		return processInterfaceList(ctx, provider, logger)
	case "get":
		return processInterfaceGet(ctx, provider, logger, jobRequest)
	case "create":
		return processInterfaceCreate(ctx, provider, logger, jobRequest)
	case "update":
		return processInterfaceUpdate(ctx, provider, logger, jobRequest)
	case "delete":
		return processInterfaceDelete(ctx, provider, logger, jobRequest)
	default:
		return nil, fmt.Errorf("unsupported interface operation: %s", jobRequest.Operation)
	}
}

// processInterfaceList lists all managed network interface configurations.
func processInterfaceList(
	ctx context.Context,
	provider iface.Provider,
	logger *slog.Logger,
) (json.RawMessage, error) {
	logger.Debug("executing interface.List")

	entries, err := provider.List(ctx)
	if err != nil {
		return nil, err
	}

	return json.Marshal(entries)
}

// processInterfaceGet gets a single interface configuration by name.
func processInterfaceGet(
	ctx context.Context,
	provider iface.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal interface get data: %w", err)
	}

	logger.Debug(
		"executing interface.Get",
		slog.String("name", data.Name),
	)

	entry, err := provider.Get(ctx, data.Name)
	if err != nil {
		return nil, err
	}

	return json.Marshal(entry)
}

// processInterfaceCreate creates a new interface configuration via Netplan.
func processInterfaceCreate(
	ctx context.Context,
	provider iface.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var entry iface.InterfaceEntry
	if err := json.Unmarshal(jobRequest.Data, &entry); err != nil {
		return nil, fmt.Errorf("unmarshal interface create data: %w", err)
	}

	logger.Debug(
		"executing interface.Create",
		slog.String("name", entry.Name),
	)

	result, err := provider.Create(ctx, entry)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processInterfaceUpdate redeploys an existing interface configuration via Netplan.
func processInterfaceUpdate(
	ctx context.Context,
	provider iface.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var entry iface.InterfaceEntry
	if err := json.Unmarshal(jobRequest.Data, &entry); err != nil {
		return nil, fmt.Errorf("unmarshal interface update data: %w", err)
	}

	logger.Debug(
		"executing interface.Update",
		slog.String("name", entry.Name),
	)

	result, err := provider.Update(ctx, entry)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processInterfaceDelete removes an interface configuration via Netplan.
func processInterfaceDelete(
	ctx context.Context,
	provider iface.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal interface delete data: %w", err)
	}

	logger.Debug(
		"executing interface.Delete",
		slog.String("name", data.Name),
	)

	result, err := provider.Delete(ctx, data.Name)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

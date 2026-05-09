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
	"github.com/retr0h/osapi/internal/provider/network/netplan/route"
)

// processRouteOperation dispatches network route sub-operations.
func processRouteOperation(
	provider route.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	if provider == nil {
		return nil, fmt.Errorf("route provider not available")
	}

	// Extract sub-operation: "route.list" -> "list"
	parts := strings.Split(jobRequest.Operation, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid route operation: %s", jobRequest.Operation)
	}
	subOp := parts[1]

	ctx := context.Background()

	switch subOp {
	case "list":
		return processRouteList(ctx, provider, logger)
	case "get":
		return processRouteGet(ctx, provider, logger, jobRequest)
	case "create":
		return processRouteCreate(ctx, provider, logger, jobRequest)
	case "update":
		return processRouteUpdate(ctx, provider, logger, jobRequest)
	case "delete":
		return processRouteDelete(ctx, provider, logger, jobRequest)
	default:
		return nil, fmt.Errorf("unsupported route operation: %s", jobRequest.Operation)
	}
}

// processRouteList lists all routes from the system routing table.
func processRouteList(
	ctx context.Context,
	provider route.Provider,
	logger *slog.Logger,
) (json.RawMessage, error) {
	logger.Debug("executing route.List")

	entries, err := provider.List(ctx)
	if err != nil {
		return nil, err
	}

	return json.Marshal(entries)
}

// processRouteGet gets the managed routes for a specific interface.
func processRouteGet(
	ctx context.Context,
	provider route.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		Interface string `json:"interface"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal route get data: %w", err)
	}

	logger.Debug(
		"executing route.Get",
		slog.String("interface", data.Interface),
	)

	entry, err := provider.Get(ctx, data.Interface)
	if err != nil {
		return nil, err
	}

	return json.Marshal(entry)
}

// processRouteCreate deploys new routes for an interface via Netplan.
func processRouteCreate(
	ctx context.Context,
	provider route.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var entry route.Entry
	if err := json.Unmarshal(jobRequest.Data, &entry); err != nil {
		return nil, fmt.Errorf("unmarshal route create data: %w", err)
	}

	logger.Debug(
		"executing route.Create",
		slog.String("interface", entry.Interface),
	)

	result, err := provider.Create(ctx, entry)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processRouteUpdate redeploys routes for an existing interface via Netplan.
func processRouteUpdate(
	ctx context.Context,
	provider route.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var entry route.Entry
	if err := json.Unmarshal(jobRequest.Data, &entry); err != nil {
		return nil, fmt.Errorf("unmarshal route update data: %w", err)
	}

	logger.Debug(
		"executing route.Update",
		slog.String("interface", entry.Interface),
	)

	result, err := provider.Update(ctx, entry)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processRouteDelete removes managed routes for an interface via Netplan.
func processRouteDelete(
	ctx context.Context,
	provider route.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		Interface string `json:"interface"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal route delete data: %w", err)
	}

	logger.Debug(
		"executing route.Delete",
		slog.String("interface", data.Interface),
	)

	result, err := provider.Delete(ctx, data.Interface)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

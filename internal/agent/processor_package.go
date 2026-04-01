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
	"github.com/retr0h/osapi/internal/provider/node/apt"
)

// processPackageOperation dispatches package management sub-operations.
func processPackageOperation(
	packageProvider apt.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	if packageProvider == nil {
		return nil, fmt.Errorf("package provider not available")
	}

	// Extract sub-operation: "package.list" -> "list"
	parts := strings.Split(jobRequest.Operation, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid package operation: %s", jobRequest.Operation)
	}
	subOp := parts[1]

	ctx := context.Background()

	switch subOp {
	case "list":
		return processPackageList(ctx, packageProvider, logger)
	case "get":
		return processPackageGet(ctx, packageProvider, logger, jobRequest)
	case "install":
		return processPackageInstall(ctx, packageProvider, logger, jobRequest)
	case "remove":
		return processPackageRemove(ctx, packageProvider, logger, jobRequest)
	case "update":
		return processPackageUpdate(ctx, packageProvider, logger)
	case "listUpdates":
		return processPackageListUpdates(ctx, packageProvider, logger)
	default:
		return nil, fmt.Errorf("unsupported package operation: %s", jobRequest.Operation)
	}
}

// processPackageList retrieves all installed packages.
func processPackageList(
	ctx context.Context,
	packageProvider apt.Provider,
	logger *slog.Logger,
) (json.RawMessage, error) {
	logger.Debug("executing package.List")

	result, err := packageProvider.List(ctx)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processPackageGet retrieves details for a single installed package.
func processPackageGet(
	ctx context.Context,
	packageProvider apt.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	logger.Debug("executing package.Get")

	var data struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal package get data: %w", err)
	}

	result, err := packageProvider.Get(ctx, data.Name)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processPackageInstall installs a package by name.
func processPackageInstall(
	ctx context.Context,
	packageProvider apt.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	logger.Debug("executing package.Install")

	var data struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal package install data: %w", err)
	}

	result, err := packageProvider.Install(ctx, data.Name)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processPackageRemove removes a package by name.
func processPackageRemove(
	ctx context.Context,
	packageProvider apt.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	logger.Debug("executing package.Remove")

	var data struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal package remove data: %w", err)
	}

	result, err := packageProvider.Remove(ctx, data.Name)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processPackageUpdate refreshes the package index.
func processPackageUpdate(
	ctx context.Context,
	packageProvider apt.Provider,
	logger *slog.Logger,
) (json.RawMessage, error) {
	logger.Debug("executing package.Update")

	result, err := packageProvider.Update(ctx)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processPackageListUpdates returns packages with available updates.
func processPackageListUpdates(
	ctx context.Context,
	packageProvider apt.Provider,
	logger *slog.Logger,
) (json.RawMessage, error) {
	logger.Debug("executing package.ListUpdates")

	result, err := packageProvider.ListUpdates(ctx)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

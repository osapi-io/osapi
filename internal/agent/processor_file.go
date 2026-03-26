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
	fileProv "github.com/retr0h/osapi/internal/provider/file"
)

// NewFileProcessor returns a ProcessorFunc that handles file-related operations.
func NewFileProcessor(
	fileProvider fileProv.Provider,
	_ *slog.Logger,
) ProcessorFunc {
	return func(req job.Request) (json.RawMessage, error) {
		if fileProvider == nil {
			return nil, fmt.Errorf("file provider not configured")
		}

		// Extract base operation from dotted operation (e.g., "deploy.execute" -> "deploy")
		baseOperation := strings.Split(req.Operation, ".")[0]

		switch baseOperation {
		case "deploy":
			return processFileDeploy(fileProvider, req)
		case "undeploy":
			return processFileUndeploy(fileProvider, req)
		case "status":
			return processFileStatus(fileProvider, req)
		default:
			return nil, fmt.Errorf("unsupported file operation: %s", req.Operation)
		}
	}
}

// processFileDeploy handles file deploy operations.
func processFileDeploy(
	fileProvider fileProv.Provider,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var req fileProv.DeployRequest
	if err := json.Unmarshal(jobRequest.Data, &req); err != nil {
		return nil, fmt.Errorf("failed to parse file deploy data: %w", err)
	}

	result, err := fileProvider.Deploy(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("file deploy failed: %w", err)
	}

	return json.Marshal(result)
}

// processFileStatus handles file status operations.
func processFileStatus(
	fileProvider fileProv.Provider,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var req fileProv.StatusRequest
	if err := json.Unmarshal(jobRequest.Data, &req); err != nil {
		return nil, fmt.Errorf("failed to parse file status data: %w", err)
	}

	result, err := fileProvider.Status(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("file status failed: %w", err)
	}

	return json.Marshal(result)
}

// processFileUndeploy handles file undeploy operations.
func processFileUndeploy(
	fileProvider fileProv.Provider,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var req fileProv.UndeployRequest
	if err := json.Unmarshal(jobRequest.Data, &req); err != nil {
		return nil, fmt.Errorf("failed to parse file undeploy data: %w", err)
	}

	result, err := fileProvider.Undeploy(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("file undeploy failed: %w", err)
	}

	return json.Marshal(result)
}

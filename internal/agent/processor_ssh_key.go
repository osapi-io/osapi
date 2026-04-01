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
	"github.com/retr0h/osapi/internal/provider/node/user"
)

// processSshKeyOperation dispatches SSH key sub-operations.
func processSshKeyOperation(
	userProvider user.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	if userProvider == nil {
		return nil, fmt.Errorf("user provider not available")
	}

	// Extract sub-operation: "sshKey.list" -> "list"
	parts := strings.Split(jobRequest.Operation, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid sshKey operation: %s", jobRequest.Operation)
	}
	subOp := parts[1]

	ctx := context.Background()

	switch subOp {
	case "list":
		return processSshKeyList(ctx, userProvider, logger, jobRequest)
	case "add":
		return processSshKeyAdd(ctx, userProvider, logger, jobRequest)
	case "remove":
		return processSshKeyRemove(ctx, userProvider, logger, jobRequest)
	default:
		return nil, fmt.Errorf("unsupported sshKey operation: %s", jobRequest.Operation)
	}
}

// processSshKeyList lists SSH keys for a user.
func processSshKeyList(
	ctx context.Context,
	userProvider user.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		Username string `json:"username"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal sshKey list data: %w", err)
	}

	logger.Debug("executing sshKey.List",
		slog.String("username", data.Username),
	)

	keys, err := userProvider.ListKeys(ctx, data.Username)
	if err != nil {
		return nil, err
	}

	return json.Marshal(keys)
}

// processSshKeyAdd adds an SSH key for a user.
func processSshKeyAdd(
	ctx context.Context,
	userProvider user.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		Username string   `json:"username"`
		Key      user.SSHKey `json:"key"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal sshKey add data: %w", err)
	}

	logger.Debug("executing sshKey.Add",
		slog.String("username", data.Username),
	)

	result, err := userProvider.AddKey(ctx, data.Username, data.Key)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processSshKeyRemove removes an SSH key for a user.
func processSshKeyRemove(
	ctx context.Context,
	userProvider user.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		Username    string `json:"username"`
		Fingerprint string `json:"fingerprint"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal sshKey remove data: %w", err)
	}

	logger.Debug("executing sshKey.Remove",
		slog.String("username", data.Username),
		slog.String("fingerprint", data.Fingerprint),
	)

	result, err := userProvider.RemoveKey(ctx, data.Username, data.Fingerprint)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

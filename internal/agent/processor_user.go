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

// processUserOperation dispatches user sub-operations.
func processUserOperation(
	userProvider user.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	if userProvider == nil {
		return nil, fmt.Errorf("user provider not available")
	}

	// Extract sub-operation: "user.list" -> "list"
	parts := strings.Split(jobRequest.Operation, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid user operation: %s", jobRequest.Operation)
	}
	subOp := parts[1]

	ctx := context.Background()

	switch subOp {
	case "list":
		return processUserList(ctx, userProvider, logger)
	case "get":
		return processUserGet(ctx, userProvider, logger, jobRequest)
	case "create":
		return processUserCreate(ctx, userProvider, logger, jobRequest)
	case "update":
		return processUserUpdate(ctx, userProvider, logger, jobRequest)
	case "delete":
		return processUserDelete(ctx, userProvider, logger, jobRequest)
	case "password":
		return processUserPassword(ctx, userProvider, logger, jobRequest)
	default:
		return nil, fmt.Errorf("unsupported user operation: %s", jobRequest.Operation)
	}
}

// processGroupOperation dispatches group sub-operations.
func processGroupOperation(
	userProvider user.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	if userProvider == nil {
		return nil, fmt.Errorf("user provider not available")
	}

	// Extract sub-operation: "group.list" -> "list"
	parts := strings.Split(jobRequest.Operation, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid group operation: %s", jobRequest.Operation)
	}
	subOp := parts[1]

	ctx := context.Background()

	switch subOp {
	case "list":
		return processGroupList(ctx, userProvider, logger)
	case "get":
		return processGroupGet(ctx, userProvider, logger, jobRequest)
	case "create":
		return processGroupCreate(ctx, userProvider, logger, jobRequest)
	case "update":
		return processGroupUpdate(ctx, userProvider, logger, jobRequest)
	case "delete":
		return processGroupDelete(ctx, userProvider, logger, jobRequest)
	default:
		return nil, fmt.Errorf("unsupported group operation: %s", jobRequest.Operation)
	}
}

// processUserList lists all system users.
func processUserList(
	ctx context.Context,
	userProvider user.Provider,
	logger *slog.Logger,
) (json.RawMessage, error) {
	logger.Debug("executing user.ListUsers")

	users, err := userProvider.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	return json.Marshal(users)
}

// processUserGet retrieves a single user by name.
func processUserGet(
	ctx context.Context,
	userProvider user.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal user get data: %w", err)
	}

	logger.Debug(
		"executing user.GetUser",
		slog.String("name", data.Name),
	)

	result, err := userProvider.GetUser(ctx, data.Name)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processUserCreate creates a new system user.
func processUserCreate(
	ctx context.Context,
	userProvider user.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var opts user.CreateUserOpts
	if err := json.Unmarshal(jobRequest.Data, &opts); err != nil {
		return nil, fmt.Errorf("unmarshal user create data: %w", err)
	}

	logger.Debug(
		"executing user.CreateUser",
		slog.String("name", opts.Name),
	)

	result, err := userProvider.CreateUser(ctx, opts)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processUserUpdate updates an existing system user.
func processUserUpdate(
	ctx context.Context,
	userProvider user.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		Name string              `json:"name"`
		Opts user.UpdateUserOpts `json:"opts"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal user update data: %w", err)
	}

	logger.Debug(
		"executing user.UpdateUser",
		slog.String("name", data.Name),
	)

	result, err := userProvider.UpdateUser(ctx, data.Name, data.Opts)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processUserDelete deletes a system user by name.
func processUserDelete(
	ctx context.Context,
	userProvider user.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal user delete data: %w", err)
	}

	logger.Debug(
		"executing user.DeleteUser",
		slog.String("name", data.Name),
	)

	result, err := userProvider.DeleteUser(ctx, data.Name)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processUserPassword changes a user's password.
func processUserPassword(
	ctx context.Context,
	userProvider user.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal user password data: %w", err)
	}

	logger.Debug(
		"executing user.ChangePassword",
		slog.String("name", data.Name),
	)

	result, err := userProvider.ChangePassword(ctx, data.Name, data.Password)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processGroupList lists all system groups.
func processGroupList(
	ctx context.Context,
	userProvider user.Provider,
	logger *slog.Logger,
) (json.RawMessage, error) {
	logger.Debug("executing user.ListGroups")

	groups, err := userProvider.ListGroups(ctx)
	if err != nil {
		return nil, err
	}

	return json.Marshal(groups)
}

// processGroupGet retrieves a single group by name.
func processGroupGet(
	ctx context.Context,
	userProvider user.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal group get data: %w", err)
	}

	logger.Debug(
		"executing user.GetGroup",
		slog.String("name", data.Name),
	)

	result, err := userProvider.GetGroup(ctx, data.Name)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processGroupCreate creates a new system group.
func processGroupCreate(
	ctx context.Context,
	userProvider user.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var opts user.CreateGroupOpts
	if err := json.Unmarshal(jobRequest.Data, &opts); err != nil {
		return nil, fmt.Errorf("unmarshal group create data: %w", err)
	}

	logger.Debug(
		"executing user.CreateGroup",
		slog.String("name", opts.Name),
	)

	result, err := userProvider.CreateGroup(ctx, opts)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processGroupUpdate updates an existing system group.
func processGroupUpdate(
	ctx context.Context,
	userProvider user.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		Name string               `json:"name"`
		Opts user.UpdateGroupOpts `json:"opts"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal group update data: %w", err)
	}

	logger.Debug(
		"executing user.UpdateGroup",
		slog.String("name", data.Name),
	)

	result, err := userProvider.UpdateGroup(ctx, data.Name, data.Opts)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// processGroupDelete deletes a system group by name.
func processGroupDelete(
	ctx context.Context,
	userProvider user.Provider,
	logger *slog.Logger,
	jobRequest job.Request,
) (json.RawMessage, error) {
	var data struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(jobRequest.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal group delete data: %w", err)
	}

	logger.Debug(
		"executing user.DeleteGroup",
		slog.String("name", data.Name),
	)

	result, err := userProvider.DeleteGroup(ctx, data.Name)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

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

package client

import (
	"context"
	"fmt"

	"github.com/retr0h/osapi/pkg/sdk/client/gen"
)

// GroupService provides group management operations.
type GroupService struct {
	client *gen.ClientWithResponses
}

// List lists all groups on the target host.
func (s *GroupService) List(
	ctx context.Context,
	hostname string,
) (*Response[Collection[GroupInfoResult]], error) {
	resp, err := s.client.GetNodeGroupWithResponse(ctx, hostname)
	if err != nil {
		return nil, fmt.Errorf("group list: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON400,
		resp.JSON401,
		resp.JSON403,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(groupInfoCollectionFromList(resp.JSON200), resp.Body), nil
}

// Get retrieves a single group by name on the target host.
func (s *GroupService) Get(
	ctx context.Context,
	hostname string,
	name string,
) (*Response[Collection[GroupInfoResult]], error) {
	resp, err := s.client.GetNodeGroupByNameWithResponse(ctx, hostname, name)
	if err != nil {
		return nil, fmt.Errorf("group get: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON400,
		resp.JSON401,
		resp.JSON403,
		resp.JSON404,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(groupInfoCollectionFromGet(resp.JSON200), resp.Body), nil
}

// Create creates a group on the target host.
func (s *GroupService) Create(
	ctx context.Context,
	hostname string,
	opts GroupCreateOpts,
) (*Response[Collection[GroupMutationResult]], error) {
	body := gen.GroupCreateRequest{
		Name: opts.Name,
	}

	if opts.GID != 0 {
		body.Gid = &opts.GID
	}

	if opts.System {
		body.System = &opts.System
	}

	resp, err := s.client.PostNodeGroupWithResponse(ctx, hostname, body)
	if err != nil {
		return nil, fmt.Errorf("group create: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON400,
		resp.JSON401,
		resp.JSON403,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(groupMutationCollectionFromCreate(resp.JSON200), resp.Body), nil
}

// Update updates a group on the target host.
func (s *GroupService) Update(
	ctx context.Context,
	hostname string,
	name string,
	opts GroupUpdateOpts,
) (*Response[Collection[GroupMutationResult]], error) {
	body := gen.GroupUpdateRequest{}

	if opts.Members != nil {
		body.Members = &opts.Members
	}

	resp, err := s.client.PutNodeGroupWithResponse(ctx, hostname, name, body)
	if err != nil {
		return nil, fmt.Errorf("group update: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON400,
		resp.JSON401,
		resp.JSON403,
		resp.JSON404,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(groupMutationCollectionFromUpdate(resp.JSON200), resp.Body), nil
}

// Delete removes a group on the target host.
func (s *GroupService) Delete(
	ctx context.Context,
	hostname string,
	name string,
) (*Response[Collection[GroupMutationResult]], error) {
	resp, err := s.client.DeleteNodeGroupWithResponse(ctx, hostname, name)
	if err != nil {
		return nil, fmt.Errorf("group delete: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON400,
		resp.JSON401,
		resp.JSON403,
		resp.JSON404,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(groupMutationCollectionFromDelete(resp.JSON200), resp.Body), nil
}

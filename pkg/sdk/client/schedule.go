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

	"github.com/osapi-io/osapi/pkg/sdk/client/gen"
)

// ScheduleService provides cron schedule management operations.
type ScheduleService struct {
	client *gen.ClientWithResponses
}

// List lists all cron entries on the target host.
func (s *ScheduleService) List(
	ctx context.Context,
	hostname string,
) (*Response[Collection[ScheduleEntryResult]], error) {
	resp, err := s.client.GetNodeScheduleWithResponse(ctx, hostname)
	if err != nil {
		return nil, fmt.Errorf("cron list: %w", err)
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

	return NewResponse(cronEntryCollectionFromGen(resp.JSON200), resp.Body), nil
}

// Get retrieves a single cron entry by name on the target host.
func (s *ScheduleService) Get(
	ctx context.Context,
	hostname string,
	name string,
) (*Response[Collection[ScheduleEntryResult]], error) {
	resp, err := s.client.GetNodeScheduleByNameWithResponse(ctx, hostname, name)
	if err != nil {
		return nil, fmt.Errorf("cron get: %w", err)
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

	return NewResponse(cronGetCollectionFromGen(resp.JSON200), resp.Body), nil
}

// Create creates a new cron entry on the target host.
func (s *ScheduleService) Create(
	ctx context.Context,
	hostname string,
	opts ScheduleCreateOpts,
) (*Response[Collection[ScheduleMutationResult]], error) {
	body := gen.ScheduleCreateRequest{
		Name:   opts.Name,
		Object: opts.Object,
	}
	if opts.Schedule != "" {
		body.Schedule = &opts.Schedule
	}
	if opts.Interval != "" {
		interval := gen.ScheduleCreateRequestInterval(opts.Interval)
		body.Interval = &interval
	}
	if opts.User != "" {
		body.User = &opts.User
	}
	if opts.ContentType != "" {
		ct := gen.ScheduleCreateRequestContentType(opts.ContentType)
		body.ContentType = &ct
	}
	if opts.Vars != nil {
		body.Vars = &opts.Vars
	}

	resp, err := s.client.PostNodeScheduleWithResponse(ctx, hostname, body)
	if err != nil {
		return nil, fmt.Errorf("cron create: %w", err)
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

	return NewResponse(cronMutationCollectionFromCreate(resp.JSON200), resp.Body), nil
}

// Update updates an existing cron entry on the target host.
func (s *ScheduleService) Update(
	ctx context.Context,
	hostname string,
	name string,
	opts ScheduleUpdateOpts,
) (*Response[Collection[ScheduleMutationResult]], error) {
	body := gen.ScheduleUpdateRequest{}
	if opts.Object != "" {
		body.Object = &opts.Object
	}
	if opts.Schedule != "" {
		body.Schedule = &opts.Schedule
	}
	if opts.User != "" {
		body.User = &opts.User
	}
	if opts.ContentType != "" {
		ct := gen.ScheduleUpdateRequestContentType(opts.ContentType)
		body.ContentType = &ct
	}
	if opts.Vars != nil {
		body.Vars = &opts.Vars
	}

	resp, err := s.client.PutNodeScheduleWithResponse(ctx, hostname, name, body)
	if err != nil {
		return nil, fmt.Errorf("cron update: %w", err)
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

	return NewResponse(cronMutationCollectionFromUpdate(resp.JSON200), resp.Body), nil
}

// Delete deletes a cron entry on the target host.
func (s *ScheduleService) Delete(
	ctx context.Context,
	hostname string,
	name string,
) (*Response[Collection[ScheduleMutationResult]], error) {
	resp, err := s.client.DeleteNodeScheduleWithResponse(ctx, hostname, name)
	if err != nil {
		return nil, fmt.Errorf("cron delete: %w", err)
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

	return NewResponse(cronMutationCollectionFromDelete(resp.JSON200), resp.Body), nil
}

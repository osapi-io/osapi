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

// ScheduleService provides schedule management operations.
type ScheduleService struct {
	client *gen.ClientWithResponses
}

// CronList lists all cron entries on the target host.
func (s *ScheduleService) CronList(
	ctx context.Context,
	hostname string,
) (*Response[Collection[CronEntryResult]], error) {
	resp, err := s.client.GetNodeScheduleCronWithResponse(ctx, hostname)
	if err != nil {
		return nil, fmt.Errorf("cron list: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
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

// CronGet retrieves a single cron entry by name on the target host.
func (s *ScheduleService) CronGet(
	ctx context.Context,
	hostname string,
	name string,
) (*Response[CronEntryResult], error) {
	resp, err := s.client.GetNodeScheduleCronByNameWithResponse(ctx, hostname, name)
	if err != nil {
		return nil, fmt.Errorf("cron get: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
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

	return NewResponse(cronEntryFromGen(resp.JSON200), resp.Body), nil
}

// CronCreate creates a new cron entry on the target host.
func (s *ScheduleService) CronCreate(
	ctx context.Context,
	hostname string,
	opts CronCreateOpts,
) (*Response[CronMutationResult], error) {
	body := gen.CronCreateRequest{
		Name:   opts.Name,
		Object: opts.Object,
	}
	if opts.Schedule != "" {
		body.Schedule = &opts.Schedule
	}
	if opts.Interval != "" {
		interval := gen.CronCreateRequestInterval(opts.Interval)
		body.Interval = &interval
	}
	if opts.User != "" {
		body.User = &opts.User
	}
	if opts.ContentType != "" {
		ct := gen.CronCreateRequestContentType(opts.ContentType)
		body.ContentType = &ct
	}
	if opts.Vars != nil {
		body.Vars = &opts.Vars
	}

	resp, err := s.client.PostNodeScheduleCronWithResponse(ctx, hostname, body)
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

	return NewResponse(cronMutationFromCreate(resp.JSON200), resp.Body), nil
}

// CronUpdate updates an existing cron entry on the target host.
func (s *ScheduleService) CronUpdate(
	ctx context.Context,
	hostname string,
	name string,
	opts CronUpdateOpts,
) (*Response[CronMutationResult], error) {
	body := gen.CronUpdateRequest{}
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
		ct := gen.CronUpdateRequestContentType(opts.ContentType)
		body.ContentType = &ct
	}
	if opts.Vars != nil {
		body.Vars = &opts.Vars
	}

	resp, err := s.client.PutNodeScheduleCronWithResponse(ctx, hostname, name, body)
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

	return NewResponse(cronMutationFromUpdate(resp.JSON200), resp.Body), nil
}

// CronDelete deletes a cron entry on the target host.
func (s *ScheduleService) CronDelete(
	ctx context.Context,
	hostname string,
	name string,
) (*Response[CronMutationResult], error) {
	resp, err := s.client.DeleteNodeScheduleCronWithResponse(ctx, hostname, name)
	if err != nil {
		return nil, fmt.Errorf("cron delete: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
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

	return NewResponse(cronMutationFromDelete(resp.JSON200), resp.Body), nil
}

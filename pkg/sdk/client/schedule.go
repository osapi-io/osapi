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
	// TODO: Wire to generated client after combined spec regeneration.
	// Expected method: s.client.GetNodeScheduleCronWithResponse(ctx, hostname)
	_ = ctx
	_ = hostname

	return nil, fmt.Errorf("cron list: %w", fmt.Errorf("SDK client not yet generated — run `just generate` to regenerate the combined spec"))
}

// CronGet retrieves a single cron entry by name on the target host.
func (s *ScheduleService) CronGet(
	ctx context.Context,
	hostname string,
	name string,
) (*Response[CronEntryResult], error) {
	// TODO: Wire to generated client after combined spec regeneration.
	// Expected method: s.client.GetNodeScheduleCronByNameWithResponse(ctx, hostname, name)
	_ = ctx
	_ = hostname
	_ = name

	return nil, fmt.Errorf("cron get: %w", fmt.Errorf("SDK client not yet generated — run `just generate` to regenerate the combined spec"))
}

// CronCreate creates a new cron entry on the target host.
func (s *ScheduleService) CronCreate(
	ctx context.Context,
	hostname string,
	opts CronCreateOpts,
) (*Response[CronMutationResult], error) {
	// TODO: Wire to generated client after combined spec regeneration.
	// Expected method: s.client.PostNodeScheduleCronWithResponse(ctx, hostname, body)
	//
	// body := gen.CronCreateRequest{
	//     Name:     opts.Name,
	//     Schedule: opts.Schedule,
	//     Command:  opts.Command,
	// }
	// if opts.User != "" {
	//     body.User = &opts.User
	// }
	_ = ctx
	_ = hostname
	_ = opts

	return nil, fmt.Errorf("cron create: %w", fmt.Errorf("SDK client not yet generated — run `just generate` to regenerate the combined spec"))
}

// CronUpdate updates an existing cron entry on the target host.
func (s *ScheduleService) CronUpdate(
	ctx context.Context,
	hostname string,
	name string,
	opts CronUpdateOpts,
) (*Response[CronMutationResult], error) {
	// TODO: Wire to generated client after combined spec regeneration.
	// Expected method: s.client.PutNodeScheduleCronWithResponse(ctx, hostname, name, body)
	//
	// body := gen.CronUpdateRequest{}
	// if opts.Schedule != "" {
	//     body.Schedule = &opts.Schedule
	// }
	// if opts.Command != "" {
	//     body.Command = &opts.Command
	// }
	// if opts.User != "" {
	//     body.User = &opts.User
	// }
	_ = ctx
	_ = hostname
	_ = name
	_ = opts

	return nil, fmt.Errorf("cron update: %w", fmt.Errorf("SDK client not yet generated — run `just generate` to regenerate the combined spec"))
}

// CronDelete deletes a cron entry on the target host.
func (s *ScheduleService) CronDelete(
	ctx context.Context,
	hostname string,
	name string,
) (*Response[CronMutationResult], error) {
	// TODO: Wire to generated client after combined spec regeneration.
	// Expected method: s.client.DeleteNodeScheduleCronWithResponse(ctx, hostname, name)
	_ = ctx
	_ = hostname
	_ = name

	return nil, fmt.Errorf("cron delete: %w", fmt.Errorf("SDK client not yet generated — run `just generate` to regenerate the combined spec"))
}

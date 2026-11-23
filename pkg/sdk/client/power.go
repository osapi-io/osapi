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

// PowerService provides power management operations.
type PowerService struct {
	client *gen.ClientWithResponses
}

// Reboot initiates a reboot on the target host.
func (s *PowerService) Reboot(
	ctx context.Context,
	hostname string,
	opts PowerOpts,
) (*Response[Collection[PowerResult]], error) {
	delay, message := powerFields(opts)

	resp, err := s.client.PostNodePowerRebootWithResponse(
		ctx, hostname, gen.PowerRebootRequest{Delay: delay, Message: message})
	if err != nil {
		return nil, fmt.Errorf("power reboot: %w", err)
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

	return newResponse(powerCollectionFromReboot(resp.JSON200), resp.Body, resp.HTTPResponse), nil
}

// Shutdown initiates a shutdown on the target host.
func (s *PowerService) Shutdown(
	ctx context.Context,
	hostname string,
	opts PowerOpts,
) (*Response[Collection[PowerResult]], error) {
	delay, message := powerFields(opts)

	resp, err := s.client.PostNodePowerShutdownWithResponse(
		ctx, hostname, gen.PowerShutdownRequest{Delay: delay, Message: message})
	if err != nil {
		return nil, fmt.Errorf("power shutdown: %w", err)
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

	return newResponse(powerCollectionFromShutdown(resp.JSON200), resp.Body, resp.HTTPResponse), nil
}

// powerFields are the delay and the message of opts, each nil when it is
// zero, so the request leaves it out.
func powerFields(
	opts PowerOpts,
) (*int, *string) {
	var delay *int
	var message *string
	if opts.Delay != 0 {
		delay = &opts.Delay
	}
	if opts.Message != "" {
		message = &opts.Message
	}

	return delay, message
}

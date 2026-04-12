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

// TimezoneService provides timezone management operations.
type TimezoneService struct {
	client *gen.ClientWithResponses
}

// Get retrieves the system timezone from the target host.
func (s *TimezoneService) Get(
	ctx context.Context,
	hostname string,
) (*Response[Collection[TimezoneResult]], error) {
	resp, err := s.client.GetNodeTimezoneWithResponse(ctx, hostname)
	if err != nil {
		return nil, fmt.Errorf("timezone get: %w", err)
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

	return NewResponse(timezoneCollectionFromGen(resp.JSON200), resp.Body), nil
}

// Update sets the system timezone on the target host.
func (s *TimezoneService) Update(
	ctx context.Context,
	hostname string,
	opts TimezoneUpdateOpts,
) (*Response[Collection[TimezoneMutationResult]], error) {
	body := gen.TimezoneUpdateRequest{
		Timezone: opts.Timezone,
	}

	resp, err := s.client.PutNodeTimezoneWithResponse(ctx, hostname, body)
	if err != nil {
		return nil, fmt.Errorf("timezone update: %w", err)
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

	return NewResponse(timezoneMutationCollectionFromUpdate(resp.JSON200), resp.Body), nil
}

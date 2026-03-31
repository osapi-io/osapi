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

// NTPService provides NTP management operations.
type NTPService struct {
	client *gen.ClientWithResponses
}

// Get retrieves NTP status from the target host.
func (s *NTPService) Get(
	ctx context.Context,
	hostname string,
) (*Response[Collection[NtpStatusResult]], error) {
	resp, err := s.client.GetNodeNtpWithResponse(ctx, hostname)
	if err != nil {
		return nil, fmt.Errorf("ntp get: %w", err)
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

	return NewResponse(ntpStatusCollectionFromGen(resp.JSON200), resp.Body), nil
}

// Create creates NTP configuration on the target host.
func (s *NTPService) Create(
	ctx context.Context,
	hostname string,
	opts NtpCreateOpts,
) (*Response[Collection[NtpMutationResult]], error) {
	body := gen.NtpCreateRequest{
		Servers: opts.Servers,
	}

	resp, err := s.client.PostNodeNtpWithResponse(ctx, hostname, body)
	if err != nil {
		return nil, fmt.Errorf("ntp create: %w", err)
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

	return NewResponse(ntpMutationCollectionFromCreate(resp.JSON200), resp.Body), nil
}

// Update updates NTP configuration on the target host.
func (s *NTPService) Update(
	ctx context.Context,
	hostname string,
	opts NtpUpdateOpts,
) (*Response[Collection[NtpMutationResult]], error) {
	body := gen.NtpUpdateRequest{
		Servers: opts.Servers,
	}

	resp, err := s.client.PutNodeNtpWithResponse(ctx, hostname, body)
	if err != nil {
		return nil, fmt.Errorf("ntp update: %w", err)
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

	return NewResponse(ntpMutationCollectionFromUpdate(resp.JSON200), resp.Body), nil
}

// Delete removes NTP configuration from the target host.
func (s *NTPService) Delete(
	ctx context.Context,
	hostname string,
) (*Response[Collection[NtpMutationResult]], error) {
	resp, err := s.client.DeleteNodeNtpWithResponse(ctx, hostname)
	if err != nil {
		return nil, fmt.Errorf("ntp delete: %w", err)
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

	return NewResponse(ntpMutationCollectionFromDelete(resp.JSON200), resp.Body), nil
}

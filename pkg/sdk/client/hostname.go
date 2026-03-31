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

// HostnameService provides hostname query and update operations.
type HostnameService struct {
	client *gen.ClientWithResponses
}

// Get retrieves the hostname from the target host.
func (s *HostnameService) Get(
	ctx context.Context,
	target string,
) (*Response[Collection[HostnameResult]], error) {
	resp, err := s.client.GetNodeHostnameWithResponse(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("get hostname: %w", err)
	}

	if err := checkError(resp.StatusCode(), resp.JSON400, resp.JSON401, resp.JSON403, resp.JSON500); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(hostnameCollectionFromGen(resp.JSON200), resp.Body), nil
}

// Update updates the hostname on the target node.
func (s *HostnameService) Update(
	ctx context.Context,
	target string,
	name string,
) (*Response[Collection[HostnameUpdateResult]], error) {
	body := gen.HostnameUpdateRequest{
		Hostname: name,
	}

	resp, err := s.client.PutNodeHostnameWithResponse(ctx, target, body)
	if err != nil {
		return nil, fmt.Errorf("set hostname: %w", err)
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

	if resp.JSON202 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(
		hostnameUpdateCollectionFromGen(resp.JSON202),
		resp.Body,
	), nil
}

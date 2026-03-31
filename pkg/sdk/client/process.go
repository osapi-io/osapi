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

// ProcessService provides process management operations.
type ProcessService struct {
	client *gen.ClientWithResponses
}

// List returns all running processes on the target host.
func (s *ProcessService) List(
	ctx context.Context,
	hostname string,
) (*Response[Collection[ProcessInfoResult]], error) {
	resp, err := s.client.GetNodeProcessWithResponse(ctx, hostname)
	if err != nil {
		return nil, fmt.Errorf("process list: %w", err)
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

	return NewResponse(processInfoCollectionFromList(resp.JSON200), resp.Body), nil
}

// Get returns information about a specific process by PID on the target host.
func (s *ProcessService) Get(
	ctx context.Context,
	hostname string,
	pid int,
) (*Response[Collection[ProcessInfoResult]], error) {
	resp, err := s.client.GetNodeProcessByPidWithResponse(ctx, hostname, pid)
	if err != nil {
		return nil, fmt.Errorf("process get: %w", err)
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

	return NewResponse(processInfoCollectionFromGet(resp.JSON200), resp.Body), nil
}

// Signal sends a signal to a specific process by PID on the target host.
func (s *ProcessService) Signal(
	ctx context.Context,
	hostname string,
	pid int,
	opts ProcessSignalOpts,
) (*Response[Collection[ProcessSignalResult]], error) {
	body := gen.ProcessSignalRequest{
		Signal: gen.ProcessSignalRequestSignal(opts.Signal),
	}

	resp, err := s.client.PostNodeProcessSignalWithResponse(ctx, hostname, pid, body)
	if err != nil {
		return nil, fmt.Errorf("process signal: %w", err)
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

	return NewResponse(processSignalCollectionFromGen(resp.JSON200), resp.Body), nil
}

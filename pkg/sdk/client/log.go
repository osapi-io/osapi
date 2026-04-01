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

// LogService provides log viewing operations.
type LogService struct {
	client *gen.ClientWithResponses
}

// Query returns journal log entries for the target host.
func (s *LogService) Query(
	ctx context.Context,
	hostname string,
	opts LogQueryOpts,
) (*Response[Collection[LogEntryResult]], error) {
	params := &gen.GetNodeLogParams{
		Lines:    opts.Lines,
		Since:    opts.Since,
		Priority: opts.Priority,
	}

	resp, err := s.client.GetNodeLogWithResponse(ctx, hostname, params)
	if err != nil {
		return nil, fmt.Errorf("log query: %w", err)
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

	return NewResponse(logCollectionFromGen(resp.JSON200), resp.Body), nil
}

// Sources returns unique syslog identifiers available in the journal on the
// target host.
func (s *LogService) Sources(
	ctx context.Context,
	hostname string,
) (*Response[Collection[LogSourceResult]], error) {
	resp, err := s.client.GetNodeLogSourceWithResponse(ctx, hostname)
	if err != nil {
		return nil, fmt.Errorf("log sources: %w", err)
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

	return NewResponse(logSourceCollectionFromGen(resp.JSON200), resp.Body), nil
}

// QueryUnit returns journal log entries for a specific systemd unit on the
// target host.
func (s *LogService) QueryUnit(
	ctx context.Context,
	hostname string,
	unit string,
	opts LogQueryOpts,
) (*Response[Collection[LogEntryResult]], error) {
	params := &gen.GetNodeLogUnitParams{
		Lines:    opts.Lines,
		Since:    opts.Since,
		Priority: opts.Priority,
	}

	resp, err := s.client.GetNodeLogUnitWithResponse(ctx, hostname, unit, params)
	if err != nil {
		return nil, fmt.Errorf("log query unit: %w", err)
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

	return NewResponse(logCollectionFromGen(resp.JSON200), resp.Body), nil
}

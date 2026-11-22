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

// LogService reads the systemd journal of target hosts: its entries and
// the sources that wrote them. It calls the log entry and log source
// behaviors.
type LogService struct {
	client *gen.ClientWithResponses
}

// Entries is the journal entries of the target that opts selects, one row
// for each entry of each host.
func (s *LogService) Entries(
	ctx context.Context,
	hostname string,
	opts LogEntryOpts,
) (*Response[Collection[LogEntryResult]], error) {
	params := &gen.GetNodeLogEntryParams{}
	if opts.Unit != "" {
		params.Unit = &opts.Unit
	}
	if opts.Lines != 0 {
		params.Lines = &opts.Lines
	}
	if opts.Since != "" {
		params.Since = &opts.Since
	}
	if opts.Priority != "" {
		p := gen.GetNodeLogEntryParamsPriority(opts.Priority)
		params.Priority = &p
	}

	resp, err := s.client.GetNodeLogEntryWithResponse(ctx, hostname, params)
	if err != nil {
		return nil, fmt.Errorf("log entries: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON400,
		resp.JSON401,
		resp.JSON403,
		resp.JSON404,
		resp.JSON409,
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

	return newResponse(
		logEntryCollectionFromGen(resp.JSON200),
		resp.Body,
		resp.HTTPResponse,
	), nil
}

// Sources is each syslog identifier of the target's journal, one row for
// each source of each host.
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
		resp.JSON400,
		resp.JSON401,
		resp.JSON403,
		resp.JSON404,
		resp.JSON409,
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

	return newResponse(
		logSourceCollectionFromGen(resp.JSON200),
		resp.Body,
		resp.HTTPResponse,
	), nil
}

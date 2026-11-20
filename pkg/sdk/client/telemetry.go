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

// TelemetryService reads the telemetry the controller can reach, such as the
// spans of a trace from its tracing backend.
type TelemetryService struct {
	client *gen.ClientWithResponses
}

// Trace returns the trace id, which a command prints after its job ID: its
// summary and its span tree. Without all, the tree leaves out the NATS
// calls, the bookkeeping of the job queue, and counts them. The tracing
// backend gets the spans a few seconds after the request, so a trace of a
// request that just ran can be a NotFoundError until then.
func (s *TelemetryService) Trace(
	ctx context.Context,
	id string,
	all bool,
) (*Response[Trace], error) {
	resp, err := s.client.GetTraceWithResponse(ctx, id, &gen.GetTraceParams{All: &all})
	if err != nil {
		return nil, fmt.Errorf("get trace %s: %w", id, err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON400,
		resp.JSON401,
		resp.JSON403,
		resp.JSON404,
		resp.JSON500,
		resp.JSON503,
	); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return newResponse(traceFromGen(resp.JSON200), resp.Body, resp.HTTPResponse), nil
}

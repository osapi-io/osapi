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

package osapi

import (
	"context"
	"fmt"

	"github.com/retr0h/osapi/pkg/sdk/osapi/gen"
)

// HealthService provides health check operations.
type HealthService struct {
	client *gen.ClientWithResponses
}

// Liveness checks if the API server process is alive.
func (s *HealthService) Liveness(
	ctx context.Context,
) (*Response[HealthStatus], error) {
	resp, err := s.client.GetHealthWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("health liveness: %w", err)
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(healthStatusFromGen(resp.JSON200), resp.Body), nil
}

// Ready checks if the API server and its dependencies are ready to
// serve traffic. A 503 response is treated as success with the
// ServiceUnavailable flag set.
func (s *HealthService) Ready(
	ctx context.Context,
) (*Response[ReadyStatus], error) {
	resp, err := s.client.GetHealthReadyWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("health ready: %w", err)
	}

	switch resp.StatusCode() {
	case 200:
		if resp.JSON200 == nil {
			return nil, &UnexpectedStatusError{APIError{
				StatusCode: 200,
				Message:    "nil response body",
			}}
		}

		return NewResponse(readyStatusFromGen(resp.JSON200, false), resp.Body), nil
	case 503:
		if resp.JSON503 == nil {
			return nil, &UnexpectedStatusError{APIError{
				StatusCode: 503,
				Message:    "nil response body",
			}}
		}

		return NewResponse(readyStatusFromGen(resp.JSON503, true), resp.Body), nil
	default:
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "unexpected status",
		}}
	}
}

// Status returns detailed system status including component health,
// NATS info, stream stats, and job queue counts. Requires authentication.
// A 503 response is treated as success with the ServiceUnavailable flag set.
func (s *HealthService) Status(
	ctx context.Context,
) (*Response[SystemStatus], error) {
	resp, err := s.client.GetHealthStatusWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("health status: %w", err)
	}

	// Auth errors take precedence.
	if resp.StatusCode() == 401 || resp.StatusCode() == 403 {
		return nil, checkError(resp.StatusCode(), resp.JSON401, resp.JSON403)
	}

	switch resp.StatusCode() {
	case 200:
		if resp.JSON200 == nil {
			return nil, &UnexpectedStatusError{APIError{
				StatusCode: 200,
				Message:    "nil response body",
			}}
		}

		return NewResponse(systemStatusFromGen(resp.JSON200, false), resp.Body), nil
	case 503:
		if resp.JSON503 == nil {
			return nil, &UnexpectedStatusError{APIError{
				StatusCode: 503,
				Message:    "nil response body",
			}}
		}

		return NewResponse(systemStatusFromGen(resp.JSON503, true), resp.Body), nil
	default:
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "unexpected status",
		}}
	}
}

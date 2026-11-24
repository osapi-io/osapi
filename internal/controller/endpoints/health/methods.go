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
package health

import (
	"context"
	"net/http"

	"github.com/osapi-io/osapi/internal/controller/endpoint"
)

// Liveness says the process is alive. It checks nothing.
func Liveness(
	_ context.Context,
	_ Deps,
	_ LivenessInput,
) (HealthResponse, error) {
	return HealthResponse{Status: "ok"}, nil
}

// Ready says whether the controller can take traffic, and why not with a
// 503.
func Ready(
	ctx context.Context,
	d Deps,
	_ ReadyInput,
) (ReadyResponse, error) {
	if err := d.Health.Ready(ctx); err != nil {
		return ReadyResponse{Status: "not_ready", Error: err.Error()},
			endpoint.Answer(http.StatusServiceUnavailable)
	}

	return ReadyResponse{Status: "ready"}, nil
}

// Status is the health of each component. A component that is not healthy
// makes the status degraded; the request still succeeds, because it is a
// report and not a probe. Ready is the probe.
func Status(
	ctx context.Context,
	d Deps,
	in StatusInput,
) (StatusResponse, error) {
	out := d.Health.Status(ctx)
	if in.Component != "" {
		kept := []ComponentEntry{}
		for _, e := range out.Registry {
			if e.Type == in.Component {
				kept = append(kept, e)
			}
		}
		out.Registry = kept
	}

	return out, nil
}

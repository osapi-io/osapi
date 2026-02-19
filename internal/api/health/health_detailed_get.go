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
	"time"

	"github.com/retr0h/osapi/internal/api/health/gen"
)

// GetHealthDetailed returns per-component health status (authenticated).
func (h *Health) GetHealthDetailed(
	_ context.Context,
	_ gen.GetHealthDetailedRequestObject,
) (gen.GetHealthDetailedResponseObject, error) {
	checker, ok := h.Checker.(*NATSChecker)
	if !ok {
		return h.buildDetailedResponse(nil, nil), nil
	}

	natsErr := checker.CheckNATS()
	kvErr := checker.CheckKV()

	return h.buildDetailedResponse(natsErr, kvErr), nil
}

// buildDetailedResponse constructs the detailed health response from component checks.
func (h *Health) buildDetailedResponse(
	natsErr error,
	kvErr error,
) gen.GetHealthDetailedResponseObject {
	natsComponent := gen.ComponentHealth{Status: "ok"}
	if natsErr != nil {
		errMsg := natsErr.Error()
		natsComponent = gen.ComponentHealth{Status: "error", Error: &errMsg}
	}

	kvComponent := gen.ComponentHealth{Status: "ok"}
	if kvErr != nil {
		errMsg := kvErr.Error()
		kvComponent = gen.ComponentHealth{Status: "error", Error: &errMsg}
	}

	uptime := time.Since(h.StartTime).Round(time.Second).String()

	components := map[string]gen.ComponentHealth{
		"nats": natsComponent,
		"kv":   kvComponent,
	}

	overall := "ok"
	if natsErr != nil || kvErr != nil {
		overall = "degraded"
	}

	resp := gen.DetailedHealthResponse{
		Status:     overall,
		Components: components,
		Version:    h.Version,
		Uptime:     uptime,
	}

	if overall != "ok" {
		return gen.GetHealthDetailed503JSONResponse(resp)
	}

	return gen.GetHealthDetailed200JSONResponse(resp)
}

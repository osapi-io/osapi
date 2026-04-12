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

package agent

import (
	"context"

	"github.com/retr0h/osapi/internal/controller/api/agent/gen"
)

// GetAgentsPending handles GET /agent/pending.
func (a *Agent) GetAgentsPending(
	ctx context.Context,
	_ gen.GetAgentsPendingRequestObject,
) (gen.GetAgentsPendingResponseObject, error) {
	if a.enrollment == nil {
		errMsg := "enrollment not enabled"
		return gen.GetAgentsPending500JSONResponse{Error: &errMsg}, nil
	}

	pending, err := a.enrollment.ListPending(ctx)
	if err != nil {
		errMsg := err.Error()
		return gen.GetAgentsPending500JSONResponse{Error: &errMsg}, nil
	}

	agents := make([]gen.PendingAgentInfo, 0, len(pending))
	for _, p := range pending {
		agents = append(agents, gen.PendingAgentInfo{
			MachineId:   p.MachineID,
			Hostname:    p.Hostname,
			Fingerprint: p.Fingerprint,
			RequestedAt: p.RequestedAt,
		})
	}

	return gen.GetAgentsPending200JSONResponse{
		Agents: agents,
		Total:  len(agents),
	}, nil
}

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
	"fmt"
	"strings"

	"github.com/retr0h/osapi/internal/controller/api/agent/gen"
)

// RejectAgent handles POST /agent/{hostname}/reject.
func (a *Agent) RejectAgent(
	ctx context.Context,
	request gen.RejectAgentRequestObject,
) (gen.RejectAgentResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.RejectAgent400JSONResponse{Error: &errMsg}, nil
	}

	if a.enrollment == nil {
		errMsg := "enrollment not enabled"
		return gen.RejectAgent500JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname
	if err := a.enrollment.RejectByHostname(ctx, hostname, "rejected via API"); err != nil {
		if strings.Contains(err.Error(), "no pending agent") {
			errMsg := fmt.Sprintf("no pending agent with hostname %q", hostname)
			return gen.RejectAgent404JSONResponse{Error: &errMsg}, nil
		}

		errMsg := err.Error()
		return gen.RejectAgent500JSONResponse{Error: &errMsg}, nil
	}

	msg := fmt.Sprintf("agent %s rejected", hostname)

	return gen.RejectAgent200JSONResponse{Message: msg}, nil
}

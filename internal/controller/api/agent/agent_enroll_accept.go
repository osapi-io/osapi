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

// AcceptAgent handles POST /agent/{hostname}/accept.
func (a *Agent) AcceptAgent(
	ctx context.Context,
	request gen.AcceptAgentRequestObject,
) (gen.AcceptAgentResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.AcceptAgent400JSONResponse{Error: &errMsg}, nil
	}

	if a.enrollment == nil {
		errMsg := "enrollment not enabled"
		return gen.AcceptAgent500JSONResponse{Error: &errMsg}, nil
	}

	// If fingerprint query param is provided, accept by fingerprint.
	if request.Params.Fingerprint != nil && *request.Params.Fingerprint != "" {
		fingerprint := *request.Params.Fingerprint

		if err := a.enrollment.AcceptByFingerprint(ctx, fingerprint); err != nil {
			if strings.Contains(err.Error(), "no pending agent") {
				errMsg := fmt.Sprintf("no pending agent with fingerprint %q", fingerprint)
				return gen.AcceptAgent404JSONResponse{Error: &errMsg}, nil
			}

			errMsg := err.Error()
			return gen.AcceptAgent500JSONResponse{Error: &errMsg}, nil
		}

		msg := fmt.Sprintf("agent with fingerprint %s accepted", fingerprint)

		return gen.AcceptAgent200JSONResponse{Message: msg}, nil
	}

	// Accept by hostname.
	hostname := request.Hostname
	if err := a.enrollment.AcceptByHostname(ctx, hostname); err != nil {
		if strings.Contains(err.Error(), "no pending agent") {
			errMsg := fmt.Sprintf("no pending agent with hostname %q", hostname)
			return gen.AcceptAgent404JSONResponse{Error: &errMsg}, nil
		}

		errMsg := err.Error()
		return gen.AcceptAgent500JSONResponse{Error: &errMsg}, nil
	}

	msg := fmt.Sprintf("agent %s accepted", hostname)

	return gen.AcceptAgent200JSONResponse{Message: msg}, nil
}

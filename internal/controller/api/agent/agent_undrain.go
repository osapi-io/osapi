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
	"github.com/retr0h/osapi/internal/job"
)

// UndrainAgent handles POST /agent/{hostname}/undrain.
func (a *Agent) UndrainAgent(
	ctx context.Context,
	request gen.UndrainAgentRequestObject,
) (gen.UndrainAgentResponseObject, error) {
	if errMsg, ok := validateHostname(request.Hostname); !ok {
		return gen.UndrainAgent400JSONResponse{Error: &errMsg}, nil
	}

	hostname := request.Hostname

	agentInfo, err := a.JobClient.GetAgent(ctx, hostname)
	if err != nil {
		errMsg := fmt.Sprintf("agent not found: %s", hostname)
		return gen.UndrainAgent404JSONResponse{Error: &errMsg}, nil
	}

	if agentInfo.State != job.AgentStateDraining && agentInfo.State != job.AgentStateCordoned {
		errMsg := fmt.Sprintf(
			"agent %s is not in draining or cordoned state (current: %s)",
			hostname,
			agentInfo.State,
		)
		return gen.UndrainAgent409JSONResponse{Error: &errMsg}, nil
	}

	if err := a.JobClient.DeleteDrainFlag(ctx, hostname); err != nil {
		errMsg := fmt.Sprintf("failed to delete drain flag: %s", err.Error())
		return gen.UndrainAgent409JSONResponse{Error: &errMsg}, nil
	}

	if err := a.JobClient.WriteAgentTimelineEvent(ctx, hostname, "undrain", "Undrain initiated via API"); err != nil {
		if strings.Contains(err.Error(), "not found") {
			errMsg := fmt.Sprintf("agent not found: %s", hostname)
			return gen.UndrainAgent404JSONResponse{Error: &errMsg}, nil
		}

		errMsg := err.Error()
		return gen.UndrainAgent409JSONResponse{Error: &errMsg}, nil
	}

	msg := fmt.Sprintf("undrain initiated for agent %s", hostname)

	return gen.UndrainAgent200JSONResponse{Message: msg}, nil
}

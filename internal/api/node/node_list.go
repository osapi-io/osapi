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

package node

import (
	"context"

	"github.com/retr0h/osapi/internal/api/node/gen"
)

// GetNode discovers all active agents in the fleet.
func (s *Node) GetNode(
	ctx context.Context,
	_ gen.GetNodeRequestObject,
) (gen.GetNodeResponseObject, error) {
	agents, err := s.JobClient.ListAgents(ctx)
	if err != nil {
		errMsg := err.Error()
		return gen.GetNode500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	agentInfos := make([]gen.AgentInfo, 0, len(agents))
	for _, a := range agents {
		agentInfos = append(agentInfos, gen.AgentInfo{
			Hostname: a.Hostname,
		})
	}

	total := len(agentInfos)

	return gen.GetNode200JSONResponse{
		Agents: &agentInfos,
		Total:  total,
	}, nil
}

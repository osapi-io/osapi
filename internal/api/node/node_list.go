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
	"github.com/retr0h/osapi/internal/job"
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
		agentInfos = append(agentInfos, buildAgentInfo(&a))
	}

	total := len(agentInfos)

	return gen.GetNode200JSONResponse{
		Agents: agentInfos,
		Total:  total,
	}, nil
}

// buildAgentInfo maps a job.AgentInfo to the API gen.AgentInfo response.
func buildAgentInfo(
	a *job.AgentInfo,
) gen.AgentInfo {
	status := gen.Ready
	info := gen.AgentInfo{
		Hostname: a.Hostname,
		Status:   status,
	}

	if len(a.Labels) > 0 {
		labels := a.Labels
		info.Labels = &labels
	}

	if !a.RegisteredAt.IsZero() {
		info.RegisteredAt = &a.RegisteredAt
	}

	if !a.StartedAt.IsZero() {
		info.StartedAt = &a.StartedAt
	}

	if a.Uptime > 0 {
		uptime := formatDuration(a.Uptime)
		info.Uptime = &uptime
	}

	if a.OSInfo != nil {
		info.OsInfo = &gen.OSInfoResponse{
			Distribution: a.OSInfo.Distribution,
			Version:      a.OSInfo.Version,
		}
	}

	if a.LoadAverages != nil {
		info.LoadAverage = &gen.LoadAverageResponse{
			N1min:  a.LoadAverages.Load1,
			N5min:  a.LoadAverages.Load5,
			N15min: a.LoadAverages.Load15,
		}
	}

	if a.MemoryStats != nil {
		info.Memory = &gen.MemoryResponse{
			Total: uint64ToInt(a.MemoryStats.Total),
			Free:  uint64ToInt(a.MemoryStats.Free),
			Used:  uint64ToInt(a.MemoryStats.Cached),
		}
	}

	return info
}

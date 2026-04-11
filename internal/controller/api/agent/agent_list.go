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
	"time"

	"github.com/retr0h/osapi/internal/controller/api/agent/gen"
	"github.com/retr0h/osapi/internal/job"
)

// GetAgents discovers all active agents in the fleet.
func (a *Agent) GetAgents(
	ctx context.Context,
	_ gen.GetAgentsRequestObject,
) (gen.GetAgentsResponseObject, error) {
	agents, err := a.JobClient.ListAgents(ctx)
	if err != nil {
		errMsg := err.Error()
		return gen.GetAgents500JSONResponse{
			Error: &errMsg,
		}, nil
	}

	agentInfos := make([]gen.AgentInfo, 0, len(agents))
	for _, ag := range agents {
		agentInfos = append(agentInfos, buildAgentInfo(&ag))
	}

	total := len(agentInfos)

	return gen.GetAgents200JSONResponse{
		Agents: agentInfos,
		Total:  total,
	}, nil
}

// buildAgentInfo maps a job.AgentInfo to the API gen.AgentInfo response.
func buildAgentInfo(
	a *job.AgentInfo,
) gen.AgentInfo {
	status := gen.AgentInfoStatusReady
	info := gen.AgentInfo{
		Hostname: a.Hostname,
		Status:   status,
	}

	setIdentity(a, &info)
	setSystem(a, &info)
	setNetwork(a, &info)
	setScheduling(a, &info)

	return info
}

// setIdentity populates labels, timestamps, and basic identification fields.
func setIdentity(
	a *job.AgentInfo,
	info *gen.AgentInfo,
) {
	if a.MachineID != "" {
		machineID := a.MachineID
		info.MachineId = &machineID
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

	if len(a.Facts) > 0 {
		facts := a.Facts
		info.Facts = &facts
	}
}

// setSystem populates OS, CPU, memory, and package/service manager fields.
func setSystem(
	a *job.AgentInfo,
	info *gen.AgentInfo,
) {
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

	if a.Architecture != "" {
		arch := a.Architecture
		info.Architecture = &arch
	}

	if a.KernelVersion != "" {
		kv := a.KernelVersion
		info.KernelVersion = &kv
	}

	if a.CPUCount > 0 {
		cpuCount := a.CPUCount
		info.CpuCount = &cpuCount
	}

	if a.FQDN != "" {
		fqdn := a.FQDN
		info.Fqdn = &fqdn
	}

	if a.ServiceMgr != "" {
		svcMgr := a.ServiceMgr
		info.ServiceMgr = &svcMgr
	}

	if a.PackageMgr != "" {
		pkgMgr := a.PackageMgr
		info.PackageMgr = &pkgMgr
	}
}

// setNetwork populates interfaces, primary interface, and routes.
func setNetwork(
	a *job.AgentInfo,
	info *gen.AgentInfo,
) {
	if len(a.Interfaces) > 0 {
		ifaces := make([]gen.NetworkInterfaceResponse, len(a.Interfaces))
		for i, ni := range a.Interfaces {
			ifaces[i] = gen.NetworkInterfaceResponse{
				Name: ni.Name,
			}
			if ni.IPv4 != "" {
				ipv4 := ni.IPv4
				ifaces[i].Ipv4 = &ipv4
			}
			if ni.IPv6 != "" {
				ipv6 := ni.IPv6
				ifaces[i].Ipv6 = &ipv6
			}
			if ni.MAC != "" {
				mac := ni.MAC
				ifaces[i].Mac = &mac
			}
			if ni.Family != "" {
				family := gen.NetworkInterfaceResponseFamily(ni.Family)
				ifaces[i].Family = &family
			}
		}
		info.Interfaces = &ifaces
	}

	if a.PrimaryInterface != "" {
		pi := a.PrimaryInterface
		info.PrimaryInterface = &pi
	}

	if len(a.Routes) > 0 {
		routes := make([]gen.RouteResponse, len(a.Routes))
		for i, r := range a.Routes {
			routes[i] = gen.RouteResponse{
				Destination: r.Destination,
				Gateway:     r.Gateway,
				Interface:   r.Interface,
			}
			if r.Mask != "" {
				mask := r.Mask
				routes[i].Mask = &mask
			}
			if r.Metric != 0 {
				metric := r.Metric
				routes[i].Metric = &metric
			}
			if r.Flags != "" {
				flags := r.Flags
				routes[i].Flags = &flags
			}
		}
		info.Routes = &routes
	}
}

// setScheduling populates state, conditions, and timeline.
func setScheduling(
	a *job.AgentInfo,
	info *gen.AgentInfo,
) {
	if a.State != "" {
		state := gen.AgentInfoState(a.State)
		info.State = &state
	}

	if len(a.Conditions) > 0 {
		conditions := make([]gen.NodeCondition, len(a.Conditions))
		for i, c := range a.Conditions {
			conditions[i] = gen.NodeCondition{
				Type:               gen.NodeConditionType(c.Type),
				Status:             c.Status,
				LastTransitionTime: c.LastTransitionTime,
			}
			if c.Reason != "" {
				reason := c.Reason
				conditions[i].Reason = &reason
			}
		}
		info.Conditions = &conditions
	}

	if len(a.Timeline) > 0 {
		timeline := make([]gen.TimelineEvent, len(a.Timeline))
		for i, te := range a.Timeline {
			timeline[i] = gen.TimelineEvent{
				Timestamp: te.Timestamp,
				Event:     te.Event,
			}
			if te.Hostname != "" {
				hostname := te.Hostname
				timeline[i].Hostname = &hostname
			}
			if te.Message != "" {
				message := te.Message
				timeline[i].Message = &message
			}
			if te.Error != "" {
				errStr := te.Error
				timeline[i].Error = &errStr
			}
		}
		info.Timeline = &timeline
	}
}

func formatDuration(
	d time.Duration,
) string {
	totalMinutes := int(d.Minutes())
	days := totalMinutes / (24 * 60)
	hours := (totalMinutes % (24 * 60)) / 60
	minutes := totalMinutes % 60

	dayStr := "day"
	if days != 1 {
		dayStr = "days"
	}

	hourStr := "hour"
	if hours != 1 {
		hourStr = "hours"
	}

	minuteStr := "minute"
	if minutes != 1 {
		minuteStr = "minutes"
	}

	return fmt.Sprintf("%d %s, %d %s, %d %s", days, dayStr, hours, hourStr, minutes, minuteStr)
}

// uint64ToInt convert uint64 to int, with overflow protection.
func uint64ToInt(
	value uint64,
) int {
	maxInt := int(^uint(0) >> 1)
	if value > uint64(maxInt) {
		return maxInt
	}
	return int(value)
}

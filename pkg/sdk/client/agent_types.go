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

package client

import (
	"time"

	"github.com/retr0h/osapi/pkg/sdk/client/gen"
)

// Agent represents a registered OSAPI agent.
type Agent struct {
	Hostname         string
	Status           string
	State            string
	Labels           map[string]string
	Architecture     string
	CPUCount         int
	Fqdn             string
	KernelVersion    string
	PackageMgr       string
	ServiceMgr       string
	LoadAverage      *LoadAverage
	Memory           *Memory
	OSInfo           *OSInfo
	PrimaryInterface string
	Interfaces       []NetworkInterface
	Routes           []Route
	Conditions       []Condition
	Timeline         []TimelineEvent
	Uptime           string
	StartedAt        time.Time
	RegisteredAt     time.Time
	Facts            map[string]any
}

// Condition represents a node condition evaluated agent-side.
type Condition struct {
	Type               string
	Status             bool
	Reason             string
	LastTransitionTime time.Time
}

// AgentList is a collection of agents.
type AgentList struct {
	Agents []Agent
	Total  int
}

// NetworkInterface represents a network interface on an agent.
type NetworkInterface struct {
	Name   string
	Family string
	IPv4   string
	IPv6   string
	MAC    string
}

// Route represents a network routing table entry.
type Route struct {
	Destination string
	Gateway     string
	Interface   string
	Mask        string
	Flags       string
	Metric      int
}

// LoadAverage represents system load averages.
type LoadAverage struct {
	OneMin     float32
	FiveMin    float32
	FifteenMin float32
}

// Memory represents memory usage information.
type Memory struct {
	Total int
	Used  int
	Free  int
}

// OSInfo represents operating system information.
type OSInfo struct {
	Distribution string
	Version      string
}

// agentFromGen converts a gen.AgentInfo to an Agent.
func agentFromGen(
	g *gen.AgentInfo,
) Agent {
	a := Agent{
		Hostname: g.Hostname,
		Status:   string(g.Status),
	}

	if g.Labels != nil {
		a.Labels = *g.Labels
	}

	if g.Architecture != nil {
		a.Architecture = *g.Architecture
	}

	if g.CpuCount != nil {
		a.CPUCount = *g.CpuCount
	}

	if g.Fqdn != nil {
		a.Fqdn = *g.Fqdn
	}

	if g.KernelVersion != nil {
		a.KernelVersion = *g.KernelVersion
	}

	if g.PackageMgr != nil {
		a.PackageMgr = *g.PackageMgr
	}

	if g.ServiceMgr != nil {
		a.ServiceMgr = *g.ServiceMgr
	}

	a.LoadAverage = loadAverageFromGen(g.LoadAverage)
	a.Memory = memoryFromGen(g.Memory)
	a.OSInfo = osInfoFromGen(g.OsInfo)

	if g.PrimaryInterface != nil {
		a.PrimaryInterface = *g.PrimaryInterface
	}

	if g.Routes != nil {
		routes := make([]Route, 0, len(*g.Routes))
		for _, r := range *g.Routes {
			route := Route{
				Destination: r.Destination,
				Gateway:     r.Gateway,
				Interface:   r.Interface,
			}

			if r.Mask != nil {
				route.Mask = *r.Mask
			}

			if r.Flags != nil {
				route.Flags = *r.Flags
			}

			if r.Metric != nil {
				route.Metric = *r.Metric
			}

			routes = append(routes, route)
		}

		a.Routes = routes
	}

	if g.Interfaces != nil {
		ifaces := make([]NetworkInterface, 0, len(*g.Interfaces))
		for _, iface := range *g.Interfaces {
			ni := NetworkInterface{
				Name: iface.Name,
			}

			if iface.Family != nil {
				ni.Family = string(*iface.Family)
			}

			if iface.Ipv4 != nil {
				ni.IPv4 = *iface.Ipv4
			}

			if iface.Ipv6 != nil {
				ni.IPv6 = *iface.Ipv6
			}

			if iface.Mac != nil {
				ni.MAC = *iface.Mac
			}

			ifaces = append(ifaces, ni)
		}

		a.Interfaces = ifaces
	}

	if g.Uptime != nil {
		a.Uptime = *g.Uptime
	}

	if g.StartedAt != nil {
		a.StartedAt = *g.StartedAt
	}

	if g.RegisteredAt != nil {
		a.RegisteredAt = *g.RegisteredAt
	}

	if g.Facts != nil {
		a.Facts = *g.Facts
	}

	if g.State != nil {
		a.State = string(*g.State)
	}

	if g.Conditions != nil {
		conditions := make([]Condition, 0, len(*g.Conditions))
		for _, c := range *g.Conditions {
			cond := Condition{
				Type:               string(c.Type),
				Status:             c.Status,
				LastTransitionTime: c.LastTransitionTime,
			}

			if c.Reason != nil {
				cond.Reason = *c.Reason
			}

			conditions = append(conditions, cond)
		}

		a.Conditions = conditions
	}

	if g.Timeline != nil {
		timeline := make([]TimelineEvent, 0, len(*g.Timeline))
		for _, t := range *g.Timeline {
			te := TimelineEvent{
				Event:     t.Event,
				Timestamp: t.Timestamp.Format(time.RFC3339),
			}

			if t.Hostname != nil {
				te.Hostname = *t.Hostname
			}

			if t.Message != nil {
				te.Message = *t.Message
			}

			if t.Error != nil {
				te.Error = *t.Error
			}

			timeline = append(timeline, te)
		}

		a.Timeline = timeline
	}

	return a
}

// agentListFromGen converts a gen.ListAgentsResponse to an AgentList.
func agentListFromGen(
	g *gen.ListAgentsResponse,
) AgentList {
	agents := make([]Agent, 0, len(g.Agents))
	for i := range g.Agents {
		agents = append(agents, agentFromGen(&g.Agents[i]))
	}

	return AgentList{
		Agents: agents,
		Total:  g.Total,
	}
}

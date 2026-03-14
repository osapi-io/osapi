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
	Hostname         string             `json:"hostname"`
	Status           string             `json:"status"`
	State            string             `json:"state,omitempty"`
	Labels           map[string]string  `json:"labels,omitempty"`
	Architecture     string             `json:"architecture,omitempty"`
	CPUCount         int                `json:"cpu_count"`
	Fqdn             string             `json:"fqdn,omitempty"`
	KernelVersion    string             `json:"kernel_version,omitempty"`
	PackageMgr       string             `json:"package_mgr,omitempty"`
	ServiceMgr       string             `json:"service_mgr,omitempty"`
	LoadAverage      *LoadAverage       `json:"load_average,omitempty"`
	Memory           *Memory            `json:"memory,omitempty"`
	OSInfo           *OSInfo            `json:"os_info,omitempty"`
	PrimaryInterface string             `json:"primary_interface,omitempty"`
	Interfaces       []NetworkInterface `json:"interfaces,omitempty"`
	Routes           []Route            `json:"routes,omitempty"`
	Conditions       []Condition        `json:"conditions,omitempty"`
	Timeline         []TimelineEvent    `json:"timeline,omitempty"`
	Uptime           string             `json:"uptime,omitempty"`
	StartedAt        time.Time          `json:"started_at"`
	RegisteredAt     time.Time          `json:"registered_at"`
	Facts            map[string]any     `json:"facts,omitempty"`
}

// Condition represents a node condition evaluated agent-side.
type Condition struct {
	Type               string    `json:"type"`
	Status             bool      `json:"status"`
	Reason             string    `json:"reason,omitempty"`
	LastTransitionTime time.Time `json:"last_transition_time"`
}

// AgentList is a collection of agents.
type AgentList struct {
	Agents []Agent `json:"agents"`
	Total  int     `json:"total"`
}

// NetworkInterface represents a network interface on an agent.
type NetworkInterface struct {
	Name   string `json:"name"`
	Family string `json:"family,omitempty"`
	IPv4   string `json:"ipv4,omitempty"`
	IPv6   string `json:"ipv6,omitempty"`
	MAC    string `json:"mac,omitempty"`
}

// Route represents a network routing table entry.
type Route struct {
	Destination string `json:"destination"`
	Gateway     string `json:"gateway"`
	Interface   string `json:"interface"`
	Mask        string `json:"mask,omitempty"`
	Flags       string `json:"flags,omitempty"`
	Metric      int    `json:"metric"`
}

// LoadAverage represents system load averages.
type LoadAverage struct {
	OneMin     float32 `json:"one_min"`
	FiveMin    float32 `json:"five_min"`
	FifteenMin float32 `json:"fifteen_min"`
}

// Memory represents memory usage information.
type Memory struct {
	Total int `json:"total"`
	Used  int `json:"used"`
	Free  int `json:"free"`
}

// OSInfo represents operating system information.
type OSInfo struct {
	Distribution string `json:"distribution"`
	Version      string `json:"version"`
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

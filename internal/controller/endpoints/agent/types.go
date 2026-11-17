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

import "time"

// ListInput is the input of list and pending, which have none.
type ListInput struct{}

// GetInput names the agent to read.
type GetInput struct {
	Hostname string `json:"hostname" in:"path" validate:"required,min=1,max=255" doc:"The hostname of the agent to retrieve."`
}

// RemoveInput names the accepted agent to remove.
type RemoveInput struct {
	Hostname  string `json:"hostname"   in:"path"  validate:"required,min=1,max=255"  doc:"The hostname of the accepted agent to remove."`
	MachineID string `json:"machine_id" in:"query" validate:"omitempty,min=1,max=255" doc:"The machine ID to remove, when more than one accepted agent has the hostname."`
}

// AcceptInput names the pending agent to accept.
type AcceptInput struct {
	Hostname    string `json:"hostname"    in:"path"  validate:"required,min=1,max=255"  doc:"The hostname of the pending agent to accept."`
	Fingerprint string `json:"fingerprint" in:"query" validate:"omitempty,min=1,max=255" doc:"Accept by fingerprint instead of hostname."`
}

// RejectInput names the pending agent to reject.
type RejectInput struct {
	Hostname string `json:"hostname" in:"path" validate:"required,min=1,max=255" doc:"The hostname of the pending agent to reject."`
}

// DrainInput names the agent to drain.
type DrainInput struct {
	Hostname string `json:"hostname" in:"path" validate:"required,min=1,max=255" doc:"The hostname of the agent to drain."`
}

// UndrainInput names the agent to undrain.
type UndrainInput struct {
	Hostname string `json:"hostname" in:"path" validate:"required,min=1,max=255" doc:"The hostname of the agent to undrain."`
}

// Message is the answer of an action: a sentence that says what occurred.
// It is an alias of a struct with no name, so the spec has it in place, as
// the hand-written spec does.
type Message = struct {
	Message string `json:"message"`
}

// ListAgentsResponse is the agents that report.
type ListAgentsResponse struct {
	Agents []AgentInfo `json:"agents"`
	Total  int         `json:"total"  doc:"Total number of active agents."`
}

// ListPendingAgentsResponse is the agents that wait for enrollment.
type ListPendingAgentsResponse struct {
	Agents []PendingAgentInfo `json:"agents"`
	Total  int                `json:"total"  doc:"Total number of pending agents."`
}

// PendingAgentInfo is an agent that asked to enroll.
type PendingAgentInfo struct {
	MachineID   string    `json:"machine_id"   doc:"The permanent machine identifier."`
	Hostname    string    `json:"hostname"     doc:"The hostname of the agent."`
	Fingerprint string    `json:"fingerprint"  doc:"The SHA256 fingerprint of the agent's public key."`
	RequestedAt time.Time `json:"requested_at" doc:"When the enrollment request was received."`
}

// AgentInfo is one agent: its identity, its host's facts and its state. The
// spec names its schema for the type, so the name stays AgentInfo, as the
// hand-written spec and the clients made from it have it.
type AgentInfo struct { //nolint:revive // the schema name is part of the API
	MachineID        string                     `json:"machine_id,omitempty"        doc:"The permanent machine identifier."`
	Fingerprint      string                     `json:"fingerprint,omitempty"       doc:"The SHA256 fingerprint of the agent's PKI public key. Empty when PKI is disabled."`
	Version          string                     `json:"version,omitempty"           doc:"The build version of the agent binary. Absent for an agent that does not report it."                                                                                                                                           example:"v1.4.0"`
	Hostname         string                     `json:"hostname"                    doc:"The hostname of the agent."`
	Status           string                     `json:"status"                      doc:"The current status of the agent."                                                                                                                                                                                                                           validate:"oneof=Ready NotReady"`
	Labels           map[string]string          `json:"labels,omitempty"            doc:"Key-value labels configured on the agent."`
	RegisteredAt     time.Time                  `json:"registered_at,omitzero"      doc:"When the agent last refreshed its heartbeat."`
	StartedAt        time.Time                  `json:"started_at,omitzero"         doc:"When the agent process started."`
	OSInfo           *OSInfoResponse            `json:"os_info,omitempty"`
	Uptime           string                     `json:"uptime,omitempty"            doc:"The system uptime."`
	LoadAverage      *LoadAverageResponse       `json:"load_average,omitempty"`
	Memory           *MemoryResponse            `json:"memory,omitempty"`
	Architecture     string                     `json:"architecture,omitempty"      doc:"CPU architecture."                                                                                                                                                                                                             example:"amd64"`
	KernelVersion    string                     `json:"kernel_version,omitempty"    doc:"OS kernel version."                                                                                                                                                                                                            example:"5.15.0-91-generic"`
	CPUCount         int                        `json:"cpu_count,omitempty"         doc:"Number of logical CPUs."                                                                                                                                                                                                       example:"4"`
	FQDN             string                     `json:"fqdn,omitempty"              doc:"Fully qualified domain name."                                                                                                                                                                                                  example:"web-01.example.com"`
	ServiceMgr       string                     `json:"service_mgr,omitempty"       doc:"Init system."                                                                                                                                                                                                                  example:"systemd"`
	PackageMgr       string                     `json:"package_mgr,omitempty"       doc:"Package manager."                                                                                                                                                                                                              example:"apt"`
	Interfaces       []NetworkInterfaceResponse `json:"interfaces,omitempty"`
	PrimaryInterface string                     `json:"primary_interface,omitempty" doc:"Name of the interface used for the default route."                                                                                                                                                                             example:"eth0"`
	Routes           []RouteResponse            `json:"routes,omitempty"            doc:"Network routing table entries."`
	Facts            map[string]any             `json:"facts,omitempty"             doc:"Extended facts from additional providers."`
	State            string                     `json:"state,omitempty"             doc:"Agent scheduling state."                                                                                                                                                                                                                                    validate:"oneof=Ready Draining Cordoned Pending"`
	Conditions       []NodeCondition            `json:"conditions,omitempty"        doc:"Evaluated node conditions."`
	Timeline         []TimelineEvent            `json:"timeline,omitempty"          doc:"Agent state transition history."`
	KeyStored        *bool                      `json:"key_stored,omitempty"        doc:"Whether the controller holds this agent's public key, recorded when its enrollment was accepted. False for an agent that has not enrolled since the key store existed, which is what to re-enrol before enabling enforcement."`
	Verified         *bool                      `json:"verified,omitempty"          doc:"Whether this agent's registration was signed by its stored key and claims the hostname it enrolled under. Only a verified registration decides where work is routed. True when the controller is not enforcing verification."`
}

// OSInfoResponse is the host's operating system.
type OSInfoResponse struct {
	Distribution string `json:"distribution" doc:"The name of the Linux distribution."    example:"Ubuntu"`
	Version      string `json:"version"      doc:"The version of the Linux distribution." example:"20.04"`
}

// LoadAverageResponse is the host's load over 1, 5 and 15 minutes.
type LoadAverageResponse struct {
	OneMin     float32 `json:"1min"  doc:"Load average for the last 1 minute."   example:"0.32"`
	FiveMin    float32 `json:"5min"  doc:"Load average for the last 5 minutes."  example:"0.28"`
	FifteenMin float32 `json:"15min" doc:"Load average for the last 15 minutes." example:"0.25"`
}

// MemoryResponse is the host's memory, in bytes.
type MemoryResponse struct {
	Total int `json:"total" doc:"Total memory in bytes." example:"8388608"`
	Free  int `json:"free"  doc:"Free memory in bytes."  example:"2097152"`
	Used  int `json:"used"  doc:"Used memory in bytes."  example:"4194304"`
}

// NetworkInterfaceResponse is one network interface of the host.
type NetworkInterfaceResponse struct {
	Name   string `json:"name"             example:"eth0"`
	IPv4   string `json:"ipv4,omitempty"   example:"192.168.1.10"`
	IPv6   string `json:"ipv6,omitempty"   example:"fe80::1"`
	MAC    string `json:"mac,omitempty"    example:"00:11:22:33:44:55"`
	Family string `json:"family,omitempty" example:"dual"              doc:"IP address family." validate:"oneof=inet inet6 dual"`
}

// NodeCondition is a condition the controller evaluated for the host.
type NodeCondition struct {
	Type               string    `json:"type"                 validate:"oneof=MemoryPressure HighLoad DiskPressure"`
	Status             bool      `json:"status"`
	Reason             string    `json:"reason,omitempty"`
	LastTransitionTime time.Time `json:"last_transition_time"`
}

// RouteResponse is one entry of the host's routing table.
type RouteResponse struct {
	Destination string `json:"destination"      doc:"Destination network address."   example:"0.0.0.0"`
	Gateway     string `json:"gateway"          doc:"Gateway address."               example:"192.168.1.1"`
	Interface   string `json:"interface"        doc:"Network interface name."        example:"eth0"`
	Mask        string `json:"mask,omitempty"   doc:"Network mask in CIDR notation." example:"/0"`
	Metric      int    `json:"metric,omitempty" doc:"Route metric."                  example:"100"`
	Flags       string `json:"flags,omitempty"  doc:"Route flags."                   example:"0003"`
}

// TimelineEvent is one change of the agent's state.
type TimelineEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Event     string    `json:"event"`
	Hostname  string    `json:"hostname,omitempty"`
	Message   string    `json:"message,omitempty"`
	Error     string    `json:"error,omitempty"`
}

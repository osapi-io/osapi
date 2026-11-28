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
	"time"

	"github.com/osapi-io/osapi/internal/cryptoreport"
	"github.com/osapi-io/osapi/internal/secret"

	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// ListInput is the input of list, which has none.
type ListInput struct{}

// GetInput names the agent to read.
type GetInput struct {
	Hostname string `json:"hostname" in:"path" validate:"required,min=1,max=255" doc:"The hostname of the agent to retrieve."`
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
	Items      []Info `json:"items"       doc:"The active agents."             empty:"An agent appears here once it starts and connects."`
	TotalItems int    `json:"total_items" doc:"Total number of active agents."`
}

// Info is one agent: its identity, its host's facts and its state. The
// spec names its schema for the type, so the name stays Info, as the
// hand-written spec and the clients made from it have it.
type Info struct {
	_ struct{} `schema:"AgentInfo"`

	Hostname          string                     `json:"hostname"                      doc:"The hostname of the agent."`
	MachineID         string                     `json:"machine_id,omitempty"          doc:"The permanent machine identifier."`
	State             types.AgentState           `json:"state"                         doc:"The scheduling state of the agent: Ready, Draining, Cordoned, Pending, or Refused when the controller refuses its results because its keys are not the admitted keys. A drain that the agent has not applied yet is Draining." render:"status"`
	Conditions        []NodeCondition            `json:"conditions,omitempty"          doc:"The conditions that hold now, such as HighLoad. A condition that does not hold is not in the list."                                                                                                                                            none:"-"`
	Labels            map[string]string          `json:"labels,omitempty"              doc:"Key-value labels configured on the agent."`
	StartedAt         time.Time                  `json:"started_at,omitzero"           doc:"When the agent process started."                                                                                                                                                                                               render:"age"             label:"Age"`
	LoadAverage       *LoadAverageResponse       `json:"load_average,omitempty"                                                                                                                                                                                                                                                                     label:"Load"`
	OSInfo            *OSInfoResponse            `json:"os_info,omitempty"                                                                                                                                                                                                                                                                          label:"OS"`
	Version           string                     `json:"version,omitempty"             doc:"The build version of the agent binary. Absent for an agent that does not report it."                                                                                                                                                                                                example:"v1.4.0"`
	ConfigVersion     int                        `json:"config_version,omitempty"      doc:"The version of the controller's config of the agent that the agent applied. Absent when it applied none."                                                                                                                                                                           example:"2"`
	ConfigWanted      int                        `json:"config_wanted,omitempty"       doc:"The version of the config that the controller holds for the agent. Absent when it holds none."                                                                                                                                                                                      example:"3"`
	ConfigBehind      bool                       `json:"config_behind,omitempty"       doc:"The agent has not applied the newest config that the controller holds for it yet: it applies it at its next heartbeat, or it cannot reach the controller's config."`
	Fingerprint       string                     `json:"fingerprint,omitempty"         doc:"The SHA256 fingerprint of the agent's PKI public key. Empty until the agent loads its keys."                                                                                                                                                                                                                     table:"-"`
	ControllerKey     string                     `json:"controller_key,omitempty"      doc:"The fingerprint of the controller key that the agent verifies jobs with."                                                                                                                                                                               label:"Controller key"                                   table:"-"`
	ControllerKeyNext string                     `json:"controller_key_next,omitempty" doc:"The fingerprint of the key that the controller switches to, which the agent has: empty when no rotation of the controller's key is in progress, or the agent has not taken it yet."                                                                     label:"Controller key next"                              table:"-"`
	NATSCertNext      string                     `json:"nats_cert_next,omitempty"      doc:"The fingerprint of the NATS server's next certificate, which the agent trusts during a rotation of it."                                                                                                                                                 label:"Controller key next"                              table:"-"`
	FIPS              string                     `json:"fips,omitempty"                doc:"The FIPS 140-3 mode of the agent's process, then what is not approved, by kind, such as off · 1 TLS listener not approved, or approved. pki tls has the section, with each reason."                                                                     label:"FIPS"`
	Crypto            *cryptoreport.Report       `json:"crypto,omitempty"              doc:"The crypto of the agent's process: the Go Cryptographic Module, its FIPS 140-3 mode, each TLS listener, and each algorithm with whether it is approved."                                                                                                                                                         table:"-"`
	RegisteredAt      time.Time                  `json:"registered_at,omitzero"        doc:"When the agent last refreshed its heartbeat."                                                                                                                                                                                  render:"age"             label:"Last seen"                                        table:"-"`
	Uptime            string                     `json:"uptime,omitempty"              doc:"The system uptime."                                                                                                                                                                                                                                                                                              table:"-"`
	Memory            *MemoryResponse            `json:"memory,omitempty"                                                                                                                                                                                                                                                                                                                                    table:"-"`
	Architecture      string                     `json:"architecture,omitempty"        doc:"CPU architecture."                                                                                                                                                                                                                                                                  example:"amd64"              table:"-"`
	KernelVersion     string                     `json:"kernel_version,omitempty"      doc:"OS kernel version."                                                                                                                                                                                                                                                                 example:"5.15.0-91-generic"  table:"-"`
	CPUCount          int                        `json:"cpu_count,omitempty"           doc:"Number of logical CPUs."                                                                                                                                                                                                                                                            example:"4"                  table:"-"`
	FQDN              string                     `json:"fqdn,omitempty"                doc:"Fully qualified domain name."                                                                                                                                                                                                                                                       example:"web-01.example.com" table:"-"`
	ServiceMgr        string                     `json:"service_mgr,omitempty"         doc:"Init system."                                                                                                                                                                                                                                           label:"Service manager"     example:"systemd"            table:"-"`
	PackageMgr        string                     `json:"package_mgr,omitempty"         doc:"Package manager."                                                                                                                                                                                                                                       label:"Package manager"     example:"apt"                table:"-"`
	PrimaryInterface  string                     `json:"primary_interface,omitempty"   doc:"Name of the interface used for the default route."                                                                                                                                                                                                                                  example:"eth0"               table:"-"`
	Interfaces        []NetworkInterfaceResponse `json:"interfaces,omitempty"                                                                                                                                                                                                                                                                                                                                table:"-"`
	Routes            []RouteResponse            `json:"routes,omitempty"              doc:"Network routing table entries."                                                                                                                                                                                                                                                                                  table:"-"`
	MissingPrograms   []string                   `json:"missing_programs,omitempty"    doc:"The programs that the agent's behaviors declare and that the host lacks. A behavior fails until its program is installed."                                                                                                     render:"-"               label:"Missing programs"`
	Facts             map[string]any             `json:"facts,omitempty"               doc:"Extended facts from additional providers."                                                                                                                                                                                     render:"-"`
	Timeline          []TimelineEvent            `json:"timeline,omitempty"            doc:"Agent state transition history."                                                                                                                                                                                                                                                                                 table:"-"`
	KeyStored         *bool                      `json:"key_stored,omitempty"          doc:"Whether the controller holds this agent's public key, recorded when its enrollment was accepted. False for an agent that has not enrolled since the key store existed, which is what to re-enrol before enabling enforcement." render:"-"`
	Verified          *bool                      `json:"verified,omitempty"            doc:"Whether this agent's registration was signed by its stored key and claims the hostname it enrolled under. Only a verified registration decides where work is routed. True when the controller is not enforcing verification."  render:"-"`
}

// OSInfoResponse is the host's operating system.
type OSInfoResponse struct {
	Distribution string `json:"distribution"     doc:"The name of the Linux distribution."                                                                                                                                     example:"Ubuntu"`
	Version      string `json:"version"          doc:"The version of the Linux distribution."                                                                                                                                  example:"20.04"`
	Family       string `json:"family,omitempty" doc:"The platform family that the agent picks providers by, as gohai names it: debian, suse, darwin and so on. A method runs on a host whose family is one of its platforms." example:"debian"`
}

// LoadAverageResponse is the host's load over 1, 5 and 15 minutes.
type LoadAverageResponse struct {
	OneMin     float32 `json:"1min"  doc:"Load average for the last 1 minute."   example:"0.32" label:"1 min"`
	FiveMin    float32 `json:"5min"  doc:"Load average for the last 5 minutes."  example:"0.28" label:"5 min"`
	FifteenMin float32 `json:"15min" doc:"Load average for the last 15 minutes." example:"0.25" label:"15 min"`
}

// MemoryResponse is the host's memory, in bytes.
type MemoryResponse struct {
	Total int `json:"total" doc:"Total memory in bytes." example:"8388608" render:"bytes"`
	Free  int `json:"free"  doc:"Free memory in bytes."  example:"2097152" render:"bytes"`
	Used  int `json:"used"  doc:"Used memory in bytes."  example:"4194304" render:"bytes"`
}

// NetworkInterfaceResponse is one network interface of the host.
type NetworkInterfaceResponse struct {
	Name   string              `json:"name"             example:"eth0"`
	IPv4   string              `json:"ipv4,omitempty"   example:"192.168.1.10"`
	IPv6   string              `json:"ipv6,omitempty"   example:"fe80::1"`
	MAC    string              `json:"mac,omitempty"    example:"00:11:22:33:44:55"`
	Family types.AddressFamily `json:"family,omitempty" example:"dual"              doc:"IP address family." validate:"enum"`
}

// NodeCondition is a condition the controller evaluated for the host.
type NodeCondition struct {
	Type               types.ConditionType `json:"type"`
	Reason             string              `json:"reason,omitempty"`
	LastTransitionTime time.Time           `json:"last_transition_time" label:"Since" render:"age"`
}

// RouteResponse is one entry of the host's routing table.
type RouteResponse struct {
	Destination string `json:"destination"      doc:"Destination network address."   example:"0.0.0.0"`
	Gateway     string `json:"gateway"          doc:"Gateway address."               example:"192.168.1.1"`
	Interface   string `json:"interface"        doc:"Network interface name."        example:"eth0"`
	Mask        string `json:"mask,omitempty"   doc:"Network mask in CIDR notation." example:"/0"`
	Metric      int    `json:"metric,omitempty" doc:"Route metric."                  example:"100"`
}

// TimelineEvent is one change of the agent's state.
type TimelineEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Event     string    `json:"event"`
	Hostname  string    `json:"hostname,omitempty"`
	Message   string    `json:"message,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// JoinInput is an agent's join: its request, in the body.
type JoinInput struct {
	Body JoinRequest `in:"body" validate:"required"`
}

// JoinRequest is what an agent sends to join: the join token, who it is, its
// two public keys, and its signature over them with its signing key.
type JoinRequest struct {
	Token         secret.Value `json:"token"          validate:"required,max=256"        mask:"password" doc:"The join token, osapi_jt_<id>.<secret>, from osapi controller join token create."`
	MachineID     string       `json:"machine_id"     validate:"required,min=1,max=255"                  doc:"The machine ID of the host."`
	Hostname      string       `json:"hostname"       validate:"required,min=1,max=255"                  doc:"The hostname of the host."`
	PublicKey     string       `json:"public_key"     validate:"required,max=128,base64"                 doc:"The agent's Ed25519 signing key, in base64."`
	EncryptionKey string       `json:"encryption_key" validate:"required,max=128"                        doc:"The agent's public P-256 encryption key: base64url, with no padding, of the 65-byte uncompressed point."`
	Signature     string       `json:"signature"      validate:"required,max=256,base64"                 doc:"The signature of the request by the signing key, in base64."`
}

// JoinResponse is the controller's answer to a join that admits.
type JoinResponse struct {
	_                   struct{} `title:"Join"`
	ControllerPublicKey string   `             json:"controller_public_key" doc:"The controller's Ed25519 public key, in base64, which the agent keeps and checks each job with."`
}

// ConfigGetInput names the agent whose config is read.
type ConfigGetInput struct {
	Hostname string `json:"hostname" in:"path" validate:"required,min=1,max=255" doc:"The hostname of the agent."`
}

// ConfigSetInput names the agent and the labels of its config.
type ConfigSetInput struct {
	Hostname string        `json:"hostname" in:"path" validate:"required,min=1,max=255" doc:"The hostname of the agent."`
	Body     ConfigRequest `json:"body"     in:"body" validate:"required"`
}

// ConfigRequest is the config that the controller gives an agent.
type ConfigRequest struct {
	_ struct{} `schema:"AgentConfigRequest"`

	Labels        map[string]string `json:"labels,omitempty"         validate:"omitempty"                doc:"The labels of the agent, as key=value. They replace the labels of the agent's config file with the same keys. Empty removes each."`
	MaxJobs       int               `json:"max_jobs,omitempty"       validate:"omitempty,min=1,max=1000" doc:"The most jobs that the agent runs at the same time. It replaces max_jobs of the agent's config file; empty keeps the file's."`
	LogLevel      types.LogLevel    `json:"log_level,omitempty"      validate:"omitempty,enum"           doc:"The level that the agent logs at: debug, info, warn or error. Empty keeps the level of the agent's config file."`
	FactsInterval string            `json:"facts_interval,omitempty" validate:"omitempty,go_duration"    doc:"How often the agent collects its facts, as a Go duration such as 5m. Empty keeps the interval of the agent's config file."`
}

// ConfigResponse is the config that the controller holds for an
// agent.
type ConfigResponse struct {
	_             struct{}          `title:"Agent config" schema:"AgentConfigResponse"`
	Version       int               `                                                  json:"version"                  doc:"One more at each change; 0 when nothing was set."`
	Labels        map[string]string `                                                  json:"labels,omitempty"         doc:"The labels that the controller gives the agent."`
	MaxJobs       int               `                                                  json:"max_jobs,omitempty"       doc:"The most jobs that the agent runs at the same time; empty keeps the agent's config file." label:"Max jobs"`
	LogLevel      types.LogLevel    `                                                  json:"log_level,omitempty"      doc:"The level that the agent logs at; empty keeps the agent's config file."                   label:"Log level"`
	FactsInterval string            `                                                  json:"facts_interval,omitempty" doc:"How often the agent collects its facts; empty keeps the agent's config file."             label:"Facts interval"`
	UpdatedAt     *time.Time        `                                                  json:"updated_at,omitempty"     doc:"When the config last changed."                                                            label:"Updated"        render:"time"`
}

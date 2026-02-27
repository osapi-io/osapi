// Copyright (c) 2025 John Dewey

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

package job

import (
	"encoding/json"
	"time"

	"github.com/retr0h/osapi/internal/provider/node/disk"
	"github.com/retr0h/osapi/internal/provider/node/host"
	"github.com/retr0h/osapi/internal/provider/node/load"
	"github.com/retr0h/osapi/internal/provider/node/mem"
)

// Type represents the type of job operation.
type Type string

const (
	// TypeQuery represents read operations that query system state.
	TypeQuery Type = "query"
	// TypeModify represents write operations that modify system state.
	TypeModify Type = "modify"
)

// Status represents the current status of a job.
type Status string

const (
	// StatusPending indicates the job is queued but not yet processed.
	StatusPending Status = "pending"
	// StatusProcessing indicates the job is currently being processed.
	StatusProcessing Status = "processing"
	// StatusCompleted indicates the job completed successfully.
	StatusCompleted Status = "completed"
	// StatusFailed indicates the job failed during processing.
	StatusFailed Status = "failed"
)

// Request represents a request to perform a job operation.
type Request struct {
	// JobID is a unique identifier for this job.
	JobID string `json:"job_id"`
	// Type specifies whether this is a query or modify operation.
	Type Type `json:"type"`
	// Category specifies the operation category (node, network, etc.).
	Category string `json:"category"`
	// Operation specifies the specific operation to perform.
	Operation string `json:"operation"`
	// Data contains operation-specific parameters as raw JSON.
	Data json.RawMessage `json:"data,omitempty"`
	// Timestamp indicates when the request was created.
	Timestamp time.Time `json:"timestamp"`
}

// Response represents the response from a job operation.
type Response struct {
	// JobID matches the original job ID.
	JobID string `json:"job_id"`
	// Status indicates the job completion status.
	Status Status `json:"status"`
	// Data contains the operation results as raw JSON.
	Data json.RawMessage `json:"data,omitempty"`
	// Error contains error information if the job failed.
	Error string `json:"error,omitempty"`
	// Changed indicates whether the operation modified system state.
	// Nil for query operations; set for mutation operations.
	Changed *bool `json:"changed,omitempty"`
	// Hostname identifies which agent processed this job.
	Hostname string `json:"hostname"`
	// Timestamp indicates when the response was created.
	Timestamp time.Time `json:"timestamp"`
}

// Operation type definitions for hierarchical job routing
// These support the new dot-notation format used by the jobs CLI

// OperationType represents the specific operation using hierarchical format.
// This complements the existing JobType (query/modify) with specific operations.
type OperationType string

// Node operations - read-only operations that query node state
const (
	OperationNodeHostnameGet = "node.hostname.get"
	OperationNodeStatusGet   = "node.status.get"
	OperationNodeUptimeGet   = "node.uptime.get"
	OperationNodeLoadGet     = "node.load.get"
	OperationNodeMemoryGet   = "node.memory.get"
	OperationNodeDiskGet     = "node.disk.get"
)

// Network operations - operations that can modify network configuration
const (
	OperationNetworkDNSGet    = "network.dns.get"
	OperationNetworkDNSUpdate = "network.dns.update"
	OperationNetworkPingDo    = "network.ping.do"
)

// Node operations - operations that can modify node state
const (
	OperationNodeShutdown = "node.shutdown.execute"
	OperationNodeReboot   = "node.reboot.execute"
)

// Command operations - execute arbitrary commands on agents
const (
	OperationCommandExecExecute  = "command.exec.execute"
	OperationCommandShellExecute = "command.shell.execute"
)

// Operation represents an operation in the new hierarchical format
type Operation struct {
	// Type specifies the type of operation using hierarchical format
	// (e.g., "node.hostname.get", "network.dns.update")
	Type OperationType `json:"type"`
	// Data contains the operation-specific data as raw JSON
	Data json.RawMessage `json:"data"`
}

// QueuedJob represents a job stored in the KV queue with metadata
type QueuedJob struct {
	// ID is the unique identifier for this job
	ID string `json:"id"`
	// Status tracks the current state of the job
	Status string `json:"status"` // "unprocessed", "processing", "completed", "failed"
	// Created is the timestamp when the job was created
	Created string `json:"created"`
	// Subject is the NATS subject for this job (optional)
	Subject string `json:"subject,omitempty"`
	// Operation contains the actual work to be performed (stored as flexible JSON)
	Operation map[string]interface{} `json:"operation"`
	// StatusHistory tracks status transitions (optional)
	StatusHistory []interface{} `json:"status_history,omitempty"`
	// Result contains the output when the job is completed (optional)
	Result json.RawMessage `json:"result,omitempty"`
	// Error contains error details if the job failed (optional)
	Error string `json:"error,omitempty"`
	// Hostname identifies which agent processed this job (optional)
	Hostname string `json:"hostname,omitempty"`
	// UpdatedAt is the timestamp when the job was last updated (optional)
	UpdatedAt string `json:"updated_at,omitempty"`
	// AgentStates contains detailed state for each agent that processed this job
	AgentStates map[string]AgentState `json:"agent_states,omitempty"`
	// Timeline contains the chronological sequence of events for this job
	Timeline []TimelineEvent `json:"timeline,omitempty"`
	// Responses contains the actual response data from each agent
	Responses map[string]Response `json:"responses,omitempty"`
}

// AgentState represents the state of a specific agent processing a job
type AgentState struct {
	Status    string    `json:"status"`
	Error     string    `json:"error,omitempty"`
	Duration  string    `json:"duration,omitempty"`
	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`
}

// TimelineEvent represents a single event in the job timeline
type TimelineEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Event     string    `json:"event"`
	Hostname  string    `json:"hostname"`
	Message   string    `json:"message"`
	Error     string    `json:"error,omitempty"`
}

// QueueStats represents statistics about the job queue.
type QueueStats struct {
	TotalJobs       int            `json:"total_jobs"`
	StatusCounts    map[string]int `json:"status_counts"`
	OperationCounts map[string]int `json:"operation_counts"`
	DLQCount        int            `json:"dlq_count"`
}

// Operation data structures for specific operations

// NodeHostnameGetData represents data for hostname retrieval
type NodeHostnameGetData struct {
	// No additional data needed for hostname retrieval
}

// NetworkDNSUpdateData represents data for DNS configuration changes
type NetworkDNSUpdateData struct {
	// DNSServers is a list of DNS server IP addresses (IPv4 or IPv6)
	DNSServers []string `json:"dns_servers"`
	// SearchDomains is a list of search domains for DNS resolution
	SearchDomains []string `json:"search_domains"`
	// InterfaceName is the name of the network interface to apply DNS settings to
	InterfaceName string `json:"interface_name"`
}

// NetworkPingExecuteData represents data for ping operations
type NetworkPingExecuteData struct {
	// Target is the hostname or IP address to ping
	Target string `json:"target"`
	// Count is the number of ping packets to send (optional, default: 4)
	Count int `json:"count,omitempty"`
	// Timeout is the timeout duration in seconds (optional, default: 5)
	Timeout int `json:"timeout,omitempty"`
}

// CommandExecData represents data for direct command execution
type CommandExecData struct {
	// Command is the executable name or path
	Command string `json:"command"`
	// Args are the command arguments
	Args []string `json:"args,omitempty"`
	// Cwd is the optional working directory
	Cwd string `json:"cwd,omitempty"`
	// Timeout is the timeout in seconds
	Timeout int `json:"timeout,omitempty"`
}

// CommandShellData represents data for shell command execution
type CommandShellData struct {
	// Command is the full shell command string
	Command string `json:"command"`
	// Cwd is the optional working directory
	Cwd string `json:"cwd,omitempty"`
	// Timeout is the timeout in seconds
	Timeout int `json:"timeout,omitempty"`
}

// NodeShutdownData represents data for node shutdown/reboot operations
type NodeShutdownData struct {
	// Action specifies whether to reboot or shutdown the system
	Action string `json:"action"` // "reboot" or "shutdown"
	// DelaySeconds is an optional field to specify a delay in seconds before reboot/shutdown
	DelaySeconds int32 `json:"delay_seconds,omitempty"`
	// Message is an optional message to log or display before reboot/shutdown
	Message string `json:"message,omitempty"`
}

// AgentRegistration represents an agent's registration entry in the KV registry.
type AgentRegistration struct {
	// Hostname is the hostname of the agent.
	Hostname string `json:"hostname"`
	// Labels are the key-value labels configured on the agent.
	Labels map[string]string `json:"labels,omitempty"`
	// RegisteredAt is the timestamp when the agent last registered.
	RegisteredAt time.Time `json:"registered_at"`
	// StartedAt is the timestamp when the agent process started.
	StartedAt time.Time `json:"started_at"`
	// OSInfo contains operating system information.
	OSInfo *host.OSInfo `json:"os_info,omitempty"`
	// Uptime is the system uptime.
	Uptime time.Duration `json:"uptime,omitempty"`
	// LoadAverages contains the system load averages.
	LoadAverages *load.AverageStats `json:"load_averages,omitempty"`
	// MemoryStats contains memory usage information.
	MemoryStats *mem.Stats `json:"memory_stats,omitempty"`
	// AgentVersion is the version of the agent binary.
	AgentVersion string `json:"agent_version,omitempty"`
}

// AgentInfo represents information about an active agent.
type AgentInfo struct {
	// Hostname is the hostname of the agent.
	Hostname string `json:"hostname"`
	// Labels are the key-value labels configured on the agent.
	Labels map[string]string `json:"labels,omitempty"`
	// RegisteredAt is the timestamp when the agent last registered (heartbeat).
	RegisteredAt time.Time `json:"registered_at"`
	// StartedAt is the timestamp when the agent process started.
	StartedAt time.Time `json:"started_at"`
	// OSInfo contains operating system information.
	OSInfo *host.OSInfo `json:"os_info,omitempty"`
	// Uptime is the system uptime.
	Uptime time.Duration `json:"uptime,omitempty"`
	// LoadAverages contains the system load averages.
	LoadAverages *load.AverageStats `json:"load_averages,omitempty"`
	// MemoryStats contains memory usage information.
	MemoryStats *mem.Stats `json:"memory_stats,omitempty"`
	// AgentVersion is the version of the agent binary.
	AgentVersion string `json:"agent_version,omitempty"`
}

// NodeStatusResponse aggregates node status information from multiple providers.
// This represents the response for node.status.get operations in the job queue.
type NodeStatusResponse struct {
	// Hostname from the host provider
	Hostname string `json:"hostname"`
	// Uptime from the host provider
	Uptime time.Duration `json:"uptime"`
	// OSInfo from the host provider
	OSInfo *host.OSInfo `json:"os_info"`
	// LoadAverages from the load provider
	LoadAverages *load.AverageStats `json:"load_averages"`
	// MemoryStats from the memory provider
	MemoryStats *mem.Stats `json:"memory_stats"`
	// DiskUsage from the disk provider
	DiskUsage []disk.UsageStats `json:"disk_usage"`
}

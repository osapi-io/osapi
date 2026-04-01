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
	"github.com/retr0h/osapi/pkg/sdk/client"
)

// Type is a type alias for client.JobType.
type Type = client.JobType

// Job type constants re-exported from the SDK.
const (
	TypeQuery  = client.JobTypeQuery
	TypeModify = client.JobTypeModify
)

// Status represents the current status of a job.
// Status is a type alias for client.JobStatus so internal code and the SDK
// share the same type. All status constants are defined in pkg/sdk/client/.
type Status = client.JobStatus

// Job status constants re-exported from the SDK.
const (
	StatusSubmitted      = client.JobStatusSubmitted
	StatusAcknowledged   = client.JobStatusAcknowledged
	StatusStarted        = client.JobStatusStarted
	StatusPending        = client.JobStatusPending
	StatusProcessing     = client.JobStatusProcessing
	StatusCompleted      = client.JobStatusCompleted
	StatusFailed         = client.JobStatusFailed
	StatusSkipped        = client.JobStatusSkipped
	StatusPartialFailure = client.JobStatusPartialFailure
	StatusRetried        = client.JobStatusRetried
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
	Operation OperationType `json:"operation"`
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

// OperationType is a type alias for client.JobOperation.
type OperationType = client.JobOperation

// Node operations — read-only operations that query node state.
const (
	OperationNodeHostnameGet    = client.OpNodeHostnameGet
	OperationNodeHostnameUpdate = client.OpNodeHostnameUpdate
	OperationNodeStatusGet      = client.OpNodeStatusGet
	OperationNodeUptimeGet      = client.OpNodeUptimeGet
	OperationNodeLoadGet        = client.OpNodeLoadGet
	OperationNodeMemoryGet      = client.OpNodeMemoryGet
	OperationNodeDiskGet        = client.OpNodeDiskGet
	OperationNodeOSGet          = client.OpNodeOSGet
)

// Network operations.
const (
	OperationNetworkDNSGet    = client.OpNetworkDNSGet
	OperationNetworkDNSUpdate = client.OpNetworkDNSUpdate
	OperationNetworkPingDo    = client.OpNetworkPingDo
)

// Command operations — execute arbitrary commands on agents.
const (
	OperationCommandExecExecute  = client.OpCommandExec
	OperationCommandShellExecute = client.OpCommandShell
)

// File operations — manage file deployments and status.
const (
	OperationFileDeployExecute   = client.OpFileDeploy
	OperationFileUndeployExecute = client.OpFileUndeploy
	OperationFileStatusGet       = client.OpFileStatusGet
)

// Docker operations.
const (
	OperationDockerCreate      = client.OpDockerCreate
	OperationDockerStart       = client.OpDockerStart
	OperationDockerStop        = client.OpDockerStop
	OperationDockerRemove      = client.OpDockerRemove
	OperationDockerList        = client.OpDockerList
	OperationDockerInspect     = client.OpDockerInspect
	OperationDockerExec        = client.OpDockerExec
	OperationDockerPull        = client.OpDockerPull
	OperationDockerImageRemove = client.OpDockerImageRemove
)

// Schedule/Cron operations.
const (
	OperationCronList   = client.OpCronList
	OperationCronGet    = client.OpCronGet
	OperationCronCreate = client.OpCronCreate
	OperationCronUpdate = client.OpCronUpdate
	OperationCronDelete = client.OpCronDelete
)

// Sysctl operations.
const (
	OperationSysctlList   = client.OpSysctlList
	OperationSysctlGet    = client.OpSysctlGet
	OperationSysctlCreate = client.OpSysctlCreate
	OperationSysctlUpdate = client.OpSysctlUpdate
	OperationSysctlDelete = client.OpSysctlDelete
)

// NTP operations.
const (
	OperationNtpGet    = client.OpNtpGet
	OperationNtpCreate = client.OpNtpCreate
	OperationNtpUpdate = client.OpNtpUpdate
	OperationNtpDelete = client.OpNtpDelete
)

// Timezone operations.
const (
	OperationTimezoneGet    = client.OpTimezoneGet
	OperationTimezoneUpdate = client.OpTimezoneUpdate
)

// Power operations.
const (
	OperationPowerReboot   = client.OpPowerReboot
	OperationPowerShutdown = client.OpPowerShutdown
)

// Process operations.
const (
	OperationProcessList   = client.OpProcessList
	OperationProcessGet    = client.OpProcessGet
	OperationProcessSignal = client.OpProcessSignal
)

// User operations.
const (
	OperationUserList           = client.OpUserList
	OperationUserGet            = client.OpUserGet
	OperationUserCreate         = client.OpUserCreate
	OperationUserUpdate         = client.OpUserUpdate
	OperationUserDelete         = client.OpUserDelete
	OperationUserChangePassword = client.OpUserChangePassword
)

// Group operations.
const (
	OperationGroupList   = client.OpGroupList
	OperationGroupGet    = client.OpGroupGet
	OperationGroupCreate = client.OpGroupCreate
	OperationGroupUpdate = client.OpGroupUpdate
	OperationGroupDelete = client.OpGroupDelete
)

// Package operations.
const (
	OperationPackageList        = client.OpPackageList
	OperationPackageGet         = client.OpPackageGet
	OperationPackageInstall     = client.OpPackageInstall
	OperationPackageRemove      = client.OpPackageRemove
	OperationPackageUpdate      = client.OpPackageUpdate
	OperationPackageListUpdates = client.OpPackageListUpdates
)

// Log operations.
const (
	OperationLogQuery     = client.OpLogQuery
	OperationLogQueryUnit = client.OpLogQueryUnit
	OperationLogSources   = client.OpLogSources
)

// Certificate operations.
const (
	OperationCertificateCAList   = client.OpCertificateCAList
	OperationCertificateCACreate = client.OpCertificateCACreate
	OperationCertificateCAUpdate = client.OpCertificateCAUpdate
	OperationCertificateCADelete = client.OpCertificateCADelete
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
	// Changed indicates whether the operation modified system state.
	// Nil for query operations; set for mutation operations.
	Changed *bool `json:"changed,omitempty"`
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
	TotalJobs    int            `json:"total_jobs"`
	StatusCounts map[string]int `json:"status_counts"`
	DLQCount     int            `json:"dlq_count"`
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

// DockerCreateData represents data for docker container creation.
type DockerCreateData struct {
	Image     string            `json:"image"`
	Name      string            `json:"name,omitempty"`
	Command   []string          `json:"command,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
	Ports     []PortMapping     `json:"ports,omitempty"`
	Volumes   []VolumeMapping   `json:"volumes,omitempty"`
	AutoStart bool              `json:"auto_start,omitempty"`
}

// PortMapping maps a host port to a container port (job layer).
// Intentionally duplicated from runtime.PortMapping to keep the job
// layer decoupled from the provider layer. Both have the same shape.
type PortMapping struct {
	Host      int `json:"host"`
	Container int `json:"container"`
}

// VolumeMapping maps a host path to a container path (job layer).
// Intentionally duplicated from runtime.VolumeMapping for the same reason.
type VolumeMapping struct {
	Host      string `json:"host"`
	Container string `json:"container"`
}

// DockerStopData represents data for stopping a docker container.
type DockerStopData struct {
	Timeout *int `json:"timeout,omitempty"`
}

// DockerRemoveData represents data for removing a docker container.
type DockerRemoveData struct {
	Force bool `json:"force,omitempty"`
}

// DockerListData represents data for listing docker containers.
type DockerListData struct {
	State string `json:"state,omitempty"`
	Limit int    `json:"limit,omitempty"`
}

// DockerExecData represents data for executing a command in a docker container.
type DockerExecData struct {
	Command    []string          `json:"command"`
	Env        map[string]string `json:"env,omitempty"`
	WorkingDir string            `json:"working_dir,omitempty"`
}

// DockerPullData represents data for pulling a docker image.
type DockerPullData struct {
	Image string `json:"image"`
}

// DockerImageRemoveData represents data for removing a docker image.
type DockerImageRemoveData struct {
	Image string `json:"image"`
	Force bool   `json:"force,omitempty"`
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

// FileState represents a deployed file's state in the file-state KV.
// Keyed by <hostname>.<sha256-of-path>.
type FileState struct {
	ObjectName   string            `json:"object_name"`
	Path         string            `json:"path"`
	SHA256       string            `json:"sha256"`
	Mode         string            `json:"mode,omitempty"`
	Owner        string            `json:"owner,omitempty"`
	Group        string            `json:"group,omitempty"`
	DeployedAt   string            `json:"deployed_at"`
	ContentType  string            `json:"content_type"`
	UndeployedAt string            `json:"undeployed_at,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// NetworkInterface represents a network interface with its address.
type NetworkInterface struct {
	Name   string `json:"name"`
	IPv4   string `json:"ipv4,omitempty"`
	IPv6   string `json:"ipv6,omitempty"`
	MAC    string `json:"mac,omitempty"`
	Family string `json:"family,omitempty"`
}

// Route represents a network routing table entry.
type Route struct {
	Destination string `json:"destination"`
	Gateway     string `json:"gateway"`
	Interface   string `json:"interface"`
	Mask        string `json:"mask,omitempty"`
	Metric      int    `json:"metric,omitempty"`
	Flags       string `json:"flags,omitempty"`
}

// FactsRegistration represents an agent's facts entry in the facts KV bucket.
type FactsRegistration struct {
	Architecture     string             `json:"architecture,omitempty"`
	KernelVersion    string             `json:"kernel_version,omitempty"`
	CPUCount         int                `json:"cpu_count,omitempty"`
	FQDN             string             `json:"fqdn,omitempty"`
	ServiceMgr       string             `json:"service_mgr,omitempty"`
	PackageMgr       string             `json:"package_mgr,omitempty"`
	Containerized    bool               `json:"containerized"`
	Interfaces       []NetworkInterface `json:"interfaces,omitempty"`
	PrimaryInterface string             `json:"primary_interface,omitempty"`
	Routes           []Route            `json:"routes,omitempty"`
	Facts            map[string]any     `json:"facts,omitempty"`
}

// Condition type constants re-exported from the SDK.
const (
	ConditionMemoryPressure = client.ConditionMemoryPressure
	ConditionHighLoad       = client.ConditionHighLoad
	ConditionDiskPressure   = client.ConditionDiskPressure
)

// Agent state constants re-exported from the SDK.
const (
	AgentStateReady    = client.AgentReady
	AgentStateDraining = client.AgentDraining
	AgentStateCordoned = client.AgentCordoned
)

// Condition represents a node condition evaluated agent-side.
type Condition struct {
	Type               string    `json:"type"`
	Status             bool      `json:"status"`
	Reason             string    `json:"reason,omitempty"`
	LastTransitionTime time.Time `json:"last_transition_time"`
}

// ProcessMetrics holds process-level resource usage.
type ProcessMetrics struct {
	// CPUPercent is the process CPU usage as a percentage.
	CPUPercent float64 `json:"cpu_percent"`
	// RSSBytes is the resident set size in bytes.
	RSSBytes int64 `json:"rss_bytes"`
	// Goroutines is the number of active goroutines.
	Goroutines int `json:"goroutines"`
}

// SubComponentInfo holds the status and optional address of a sub-component.
type SubComponentInfo struct {
	// Status is the sub-component status (e.g., "ok", "disabled").
	Status string `json:"status"`
	// Address is the optional network endpoint (e.g., "http://0.0.0.0:9090").
	Address string `json:"address,omitempty"`
}

// ComponentRegistration represents a non-agent component's heartbeat
// entry in the KV registry. Used by API server and NATS server.
type ComponentRegistration struct {
	// Type is the component type: "controller" or "nats".
	Type string `json:"type"`
	// Hostname is the hostname of the component.
	Hostname string `json:"hostname"`
	// StartedAt is the timestamp when the component process started.
	StartedAt time.Time `json:"started_at"`
	// RegisteredAt is the timestamp of the last heartbeat.
	RegisteredAt time.Time `json:"registered_at"`
	// Process holds process-level resource usage.
	Process *ProcessMetrics `json:"process,omitempty"`
	// Conditions contains evaluated process conditions.
	Conditions []Condition `json:"conditions,omitempty"`
	// Version is the component binary version.
	Version string `json:"version,omitempty"`
	// SubComponents reports the status of internal services.
	SubComponents map[string]SubComponentInfo `json:"sub_components,omitempty"`
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
	OSInfo *host.Result `json:"os_info,omitempty"`
	// Uptime is the system uptime.
	Uptime time.Duration `json:"uptime,omitempty"`
	// LoadAverages contains the system load averages.
	LoadAverages *load.Result `json:"load_averages,omitempty"`
	// MemoryStats contains memory usage information.
	MemoryStats *mem.Result `json:"memory_stats,omitempty"`
	// AgentVersion is the version of the agent binary.
	AgentVersion string `json:"agent_version,omitempty"`
	// Process holds process-level resource usage.
	Process *ProcessMetrics `json:"process,omitempty"`
	// Conditions contains the evaluated node conditions.
	Conditions []Condition `json:"conditions,omitempty"`
	// State is the agent's scheduling state (Ready, Draining, Cordoned).
	State string `json:"state,omitempty"`
	// SubComponents reports the status of internal services.
	SubComponents map[string]SubComponentInfo `json:"sub_components,omitempty"`
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
	OSInfo *host.Result `json:"os_info,omitempty"`
	// Uptime is the system uptime.
	Uptime time.Duration `json:"uptime,omitempty"`
	// LoadAverages contains the system load averages.
	LoadAverages *load.Result `json:"load_averages,omitempty"`
	// MemoryStats contains memory usage information.
	MemoryStats *mem.Result `json:"memory_stats,omitempty"`
	// AgentVersion is the version of the agent binary.
	AgentVersion string `json:"agent_version,omitempty"`
	// Architecture is the CPU architecture (e.g., x86_64, aarch64).
	Architecture string `json:"architecture,omitempty"`
	// KernelVersion is the kernel version string.
	KernelVersion string `json:"kernel_version,omitempty"`
	// CPUCount is the number of logical CPUs.
	CPUCount int `json:"cpu_count,omitempty"`
	// FQDN is the fully qualified domain name.
	FQDN string `json:"fqdn,omitempty"`
	// ServiceMgr is the init/service manager (e.g., systemd).
	ServiceMgr string `json:"service_mgr,omitempty"`
	// PackageMgr is the package manager (e.g., apt, yum).
	PackageMgr string `json:"package_mgr,omitempty"`
	// Interfaces contains network interface information.
	Interfaces []NetworkInterface `json:"interfaces,omitempty"`
	// PrimaryInterface is the name of the interface used for the default route.
	PrimaryInterface string `json:"primary_interface,omitempty"`
	// Routes contains the network routing table.
	Routes []Route `json:"routes,omitempty"`
	// Facts contains arbitrary key-value facts collected by the agent.
	Facts map[string]any `json:"facts,omitempty"`
	// Conditions contains the evaluated node conditions.
	Conditions []Condition `json:"conditions,omitempty"`
	// State is the agent's scheduling state (Ready, Draining, Cordoned).
	State string `json:"state,omitempty"`
	// Timeline contains the chronological sequence of state transition events.
	Timeline []TimelineEvent `json:"timeline,omitempty"`
}

// NodeDiskResponse represents the response for node.disk.get operations.
type NodeDiskResponse struct {
	Disks []disk.Result `json:"disks"`
}

// NodeUptimeResponse represents the response for node.uptime.get operations.
type NodeUptimeResponse struct {
	UptimeSeconds float64 `json:"uptime_seconds"`
	Uptime        string  `json:"uptime"`
}

// NodeStatusResponse aggregates node status information from multiple providers.
// This represents the response for node.status.get operations in the job queue.
type NodeStatusResponse struct {
	// Hostname from the host provider
	Hostname string `json:"hostname"`
	// Uptime from the host provider
	Uptime time.Duration `json:"uptime"`
	// OSInfo from the host provider
	OSInfo *host.Result `json:"os_info"`
	// LoadAverages from the load provider
	LoadAverages *load.Result `json:"load_averages"`
	// MemoryStats from the memory provider
	MemoryStats *mem.Result `json:"memory_stats"`
	// DiskUsage from the disk provider
	DiskUsage []disk.Result `json:"disk_usage"`
}

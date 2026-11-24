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
	"reflect"
	"strings"
	"time"

	"github.com/osapi-io/osapi/internal/provider/node/disk"
	"github.com/osapi-io/osapi/internal/provider/node/host"
	"github.com/osapi-io/osapi/internal/provider/node/load"
	"github.com/osapi-io/osapi/internal/provider/node/memory"
	"github.com/osapi-io/osapi/pkg/sdk/client"
)

// Type is a type alias for client.JobType.
type Type = client.JobType

// Job type constants re-exported from the SDK.
const (
	TypeQuery  = client.JobTypeQuery
	TypeModify = client.JobTypeModify
)

// CategoryBehavior is the job category every registered behavior runs under.
// Its operation is "<behavior>.<method>", such as "sysctl.get".
const CategoryBehavior = "behavior"

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
	StatusTimeout        = client.JobStatusTimeout
	StatusPartialFailure = client.JobStatusPartialFailure
	StatusRetried        = client.JobStatusRetried
)

// StatusPriority maps each job status to a numeric priority used when
// computing the overall status from append-only events. Higher values
// take precedence — a "completed" event (4) is never overwritten by a
// "started" event (2) that happens to sort later alphabetically in KV.
var StatusPriority = map[string]int{
	string(StatusSubmitted):    0,
	string(StatusAcknowledged): 1,
	string(StatusStarted):      2,
	string(StatusFailed):       3,
	string(StatusSkipped):      3,
	string(StatusTimeout):      3,
	string(StatusCompleted):    4,
	string(StatusRetried):      4,
}

// SignedEnvelope wraps a job or response payload with an Ed25519 signature.
type SignedEnvelope struct {
	// Payload is the raw JSON payload (job data or response data).
	Payload []byte `json:"payload"`
	// Signature is the Ed25519 signature of the payload.
	Signature []byte `json:"signature"`
	// Fingerprint is the SHA256 fingerprint of the signer's public key.
	Fingerprint string `json:"fingerprint"`
	// MachineID identifies the signer so a verifier can find the key to
	// check the signature against. It is self-reported and therefore only
	// an index: it selects which stored record to verify against, and a
	// signature that does not match that record's key is rejected. Empty
	// for controller-signed payloads, which agents verify against the
	// controller key they were given at enrollment.
	MachineID string `json:"machine_id,omitempty"`
}

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
	// SealedKeys is the job key, sealed to each target agent, by machine
	// ID, when Data has sealed secrets. Empty when it has none.
	SealedKeys map[string]string `json:"sealed_keys,omitempty"`
	// SealedFrom is the public X25519 key that sealed each job key, which
	// an agent needs to open its own.
	SealedFrom string `json:"sealed_from,omitempty"`
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
	// ErrorCode names the cause of the failure, so the controller can answer
	// differently without reading the message. Empty when the failure has no
	// cause the API distinguishes.
	ErrorCode ErrorCode `json:"error_code,omitempty"`
	// Changed indicates whether the operation modified system state.
	// Nil for query operations; set for mutation operations.
	Changed *bool `json:"changed,omitempty"`
	// Steps are the calls of other behaviors that the job's method made on
	// the host, also when one failed. Empty for a method that calls none.
	Steps []client.Step `json:"steps,omitempty"`
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
const ()

// User operations.
const ()

// SSH Key operations.
const ()

// Network interface operations.
const (
	OperationNetworkInterfaceList   = client.OpNetworkInterfaceList
	OperationNetworkInterfaceGet    = client.OpNetworkInterfaceGet
	OperationNetworkInterfaceCreate = client.OpNetworkInterfaceCreate
	OperationNetworkInterfaceUpdate = client.OpNetworkInterfaceUpdate
	OperationNetworkInterfaceDelete = client.OpNetworkInterfaceDelete
)

// Network route operations.
const (
	OperationNetworkRouteList   = client.OpNetworkRouteList
	OperationNetworkRouteGet    = client.OpNetworkRouteGet
	OperationNetworkRouteCreate = client.OpNetworkRouteCreate
	OperationNetworkRouteUpdate = client.OpNetworkRouteUpdate
	OperationNetworkRouteDelete = client.OpNetworkRouteDelete
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
	// Target is the target the request gave, such as _all, a hostname or a
	// label, as the controller stored it with the job.
	Target string `json:"target,omitempty"`
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

// FactKeys are the keys that a template reads under .Facts: the JSON names of
// FactsRegistration, as the agent hands its facts to a provider. facts holds
// what each collector adds, under its own keys.
func FactKeys() []string {
	t := reflect.TypeFor[FactsRegistration]()
	keys := make([]string, 0, t.NumField())
	for i := range t.NumField() {
		name, _, _ := strings.Cut(t.Field(i).Tag.Get("json"), ",")
		keys = append(keys, name)
	}

	return keys
}

// Condition type constants re-exported from the SDK.
const (
	ConditionMemoryPressure = client.ConditionMemoryPressure
	ConditionHighLoad       = client.ConditionHighLoad
	ConditionDiskPressure   = client.ConditionDiskPressure
	ConditionDrainStuck     = client.ConditionDrainStuck
	// ConditionConsumersUnavailable is set while a job consumer of the
	// agent cannot be created or bound.
	ConditionConsumersUnavailable = client.ConditionConsumersUnavailable
)

// Agent state constants re-exported from the SDK.
const (
	AgentStateReady    = client.AgentReady
	AgentStateDraining = client.AgentDraining
	AgentStateCordoned = client.AgentCordoned
	AgentStatePending  = client.AgentPending
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
	// URL is where a person opens the sub-component, from its url config key.
	// An agent of an older version sends none.
	URL string `json:"url,omitempty"`
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
	// MachineID is the permanent host identifier.
	MachineID string `json:"machine_id"`
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
	MemoryStats *memory.Result `json:"memory_stats,omitempty"`
	// AgentVersion is the version of the agent binary.
	AgentVersion string `json:"agent_version,omitempty"`
	// MissingPrograms are the programs that the agent's behaviors declare
	// and that the host lacks.
	MissingPrograms []string `json:"missing_programs,omitempty"`
	// Process holds process-level resource usage.
	Process *ProcessMetrics `json:"process,omitempty"`
	// Conditions contains the evaluated node conditions.
	Conditions []Condition `json:"conditions,omitempty"`
	// State is the agent's scheduling state (Ready, Draining, Cordoned, Pending).
	State string `json:"state,omitempty"`
	// Fingerprint is the SHA256 fingerprint of the agent's PKI public key.
	// Empty until the agent loads its keys.
	Fingerprint string `json:"fingerprint,omitempty"`
	// SubComponents reports the status of internal services.
	SubComponents map[string]SubComponentInfo `json:"sub_components,omitempty"`
	// Signature is the agent's Ed25519 signature over the routing fields of
	// this registration, as RegistrationSigningBytes encodes them. Empty before
	// the agent loads its keys, and then unverifiable rather than trusted.
	Signature []byte `json:"signature,omitempty"`
}

// AgentInfo represents information about an active agent.
type AgentInfo struct {
	// MachineID is the permanent host identifier.
	MachineID string `json:"machine_id"`
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
	MemoryStats *memory.Result `json:"memory_stats,omitempty"`
	// AgentVersion is the version of the agent binary.
	AgentVersion string `json:"agent_version,omitempty"`
	// MissingPrograms are the programs that the agent's behaviors declare
	// and that the host lacks.
	MissingPrograms []string `json:"missing_programs,omitempty"`
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
	// State is the agent's scheduling state (Ready, Draining, Cordoned, Pending).
	State string `json:"state,omitempty"`
	// Fingerprint is the SHA256 fingerprint of the agent's PKI public key.
	// Empty until the agent loads its keys.
	Fingerprint string `json:"fingerprint,omitempty"`
	// Timeline contains the chronological sequence of state transition events.
	Timeline []TimelineEvent `json:"timeline,omitempty"`
	// Verified reports whether this registration's signature checked out
	// against the key stored for its machine ID. Only a verified registration
	// may decide where work goes. True when the controller is not enforcing,
	// which is the behaviour before verification existed.
	Verified bool `json:"verified"`
	// KeyStored reports whether the controller holds a key for this agent at
	// all, so an operator can tell an agent that has not re-enrolled from one
	// whose registration is being forged.
	KeyStored bool `json:"key_stored"`
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

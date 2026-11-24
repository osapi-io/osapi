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

package types

import "slices"

// JobStatus is the status of a job: what the controller computes from the
// status events of each agent that the job went to. It is one of Values.
type JobStatus string

// The statuses of a job.
const (
	// JobStatusSubmitted is a job in the queue that no agent has taken.
	JobStatusSubmitted JobStatus = "submitted"
	// JobStatusProcessing is a job that an agent acknowledged or started.
	JobStatusProcessing JobStatus = "processing"
	// JobStatusCompleted is a job that each agent completed, or completed
	// or skipped.
	JobStatusCompleted JobStatus = "completed"
	// JobStatusFailed is a job that failed on each agent that did not skip
	// it.
	JobStatusFailed JobStatus = "failed"
	// JobStatusPartialFailure is a job that completed on some agents and
	// failed on others.
	JobStatusPartialFailure JobStatus = "partial_failure"
	// JobStatusSkipped is a job that each agent skipped, because the
	// operation is not supported on its OS family.
	JobStatusSkipped JobStatus = "skipped"
)

// Values are the statuses of a job, in the order of their life.
func (JobStatus) Values() []JobStatus {
	return []JobStatus{
		JobStatusSubmitted,
		JobStatusProcessing,
		JobStatusCompleted,
		JobStatusFailed,
		JobStatusPartialFailure,
		JobStatusSkipped,
	}
}

// Valid is whether s is one of Values.
func (s JobStatus) Valid() bool {
	return slices.Contains(s.Values(), s)
}

// AgentJobStatus is the status of a job on one host: the event that a
// status key of the job records, the state of an agent, and the status of an
// agent's response. The controller records submitted and retried for the
// host _api, and an agent records the others. It is one of Values.
type AgentJobStatus string

// The statuses of a job on a host.
const (
	// AgentJobStatusSubmitted is the event of the controller that put the
	// job in the queue.
	AgentJobStatusSubmitted AgentJobStatus = "submitted"
	// AgentJobStatusAcknowledged is an agent that received the job.
	AgentJobStatusAcknowledged AgentJobStatus = "acknowledged"
	// AgentJobStatusStarted is an agent that started the operation.
	AgentJobStatusStarted AgentJobStatus = "started"
	// AgentJobStatusCompleted is an agent that completed the operation.
	AgentJobStatusCompleted AgentJobStatus = "completed"
	// AgentJobStatusFailed is an agent on which the operation failed.
	AgentJobStatusFailed AgentJobStatus = "failed"
	// AgentJobStatusSkipped is an agent that skipped the operation, because
	// it is not supported on the OS family of the host.
	AgentJobStatusSkipped AgentJobStatus = "skipped"
	// AgentJobStatusRetried is the event of the controller that sent the
	// job again as a new job.
	AgentJobStatusRetried AgentJobStatus = "retried"
	// AgentJobStatusTimeout is a host whose agent did not answer in time.
	// The controller sets it on the response, and no event records it.
	// Distinct from failed, which means the operation ran and did not
	// succeed: a timeout says nothing about whether it ran.
	AgentJobStatusTimeout AgentJobStatus = "timeout"
)

// Values are the statuses of a job on a host, in the order of their life.
func (AgentJobStatus) Values() []AgentJobStatus {
	return []AgentJobStatus{
		AgentJobStatusSubmitted,
		AgentJobStatusAcknowledged,
		AgentJobStatusStarted,
		AgentJobStatusCompleted,
		AgentJobStatusFailed,
		AgentJobStatusSkipped,
		AgentJobStatusRetried,
		AgentJobStatusTimeout,
	}
}

// Valid is whether s is one of Values.
func (s AgentJobStatus) Valid() bool {
	return slices.Contains(s.Values(), s)
}

// AgentState is the scheduling state of an agent, one of AgentStates.
type AgentState string

// The scheduling states of an agent.
const (
	// AgentReady indicates the agent is accepting and processing jobs.
	AgentReady AgentState = "Ready"
	// AgentDraining indicates the agent is finishing in-flight jobs
	// but not accepting new ones.
	AgentDraining AgentState = "Draining"
	// AgentCordoned indicates the agent is blocked from receiving
	// new jobs until manually uncordoned.
	AgentCordoned AgentState = "Cordoned"
	// AgentPending indicates the agent is awaiting PKI enrollment
	// acceptance and will not process jobs until accepted.
	AgentPending AgentState = "Pending"
)

// ConditionType is a node or process condition that the agent evaluates,
// one of ConditionTypes.
type ConditionType string

// The conditions that an agent evaluates.
const (
	ConditionMemoryPressure ConditionType = "MemoryPressure"
	ConditionHighLoad       ConditionType = "HighLoad"
	ConditionDiskPressure   ConditionType = "DiskPressure"
	// ConditionDrainStuck is set when a drain's consumers have not stopped
	// within the agent's limit, so the agent cannot reach Cordoned.
	ConditionDrainStuck ConditionType = "DrainStuck"
	// ConditionConsumersUnavailable is set while a job consumer of the
	// agent cannot be created or bound, so the agent does not get all of
	// its jobs.
	ConditionConsumersUnavailable ConditionType = "ConsumersUnavailable"
	// ConditionProcessMemoryPressure is set while the RSS of the osapi
	// process is over its limit.
	ConditionProcessMemoryPressure ConditionType = "ProcessMemoryPressure"
	// ConditionProcessHighCPU is set while the CPU use of the osapi process
	// is over its limit.
	ConditionProcessHighCPU ConditionType = "ProcessHighCPU"
)

// AgentStates are each scheduling state of an agent.
var AgentStates = []AgentState{AgentReady, AgentDraining, AgentCordoned, AgentPending}

// ConditionTypes are each condition that an agent evaluates.
var ConditionTypes = []ConditionType{
	ConditionMemoryPressure, ConditionHighLoad, ConditionDiskPressure,
	ConditionDrainStuck, ConditionConsumersUnavailable,
	ConditionProcessMemoryPressure, ConditionProcessHighCPU,
}

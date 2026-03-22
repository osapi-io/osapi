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

// JobStatus represents the status of a job in the OSAPI system.
// These values match the status strings returned by the REST API.
type JobStatus string

const (
	// JobStatusSubmitted indicates the job is queued but no agent has
	// acknowledged it yet.
	JobStatusSubmitted JobStatus = "submitted"
	// JobStatusAcknowledged indicates an agent received the job.
	JobStatusAcknowledged JobStatus = "acknowledged"
	// JobStatusStarted indicates the agent has begun processing.
	JobStatusStarted JobStatus = "started"
	// JobStatusProcessing indicates the job is currently being processed.
	JobStatusProcessing JobStatus = "processing"
	// JobStatusCompleted indicates the job completed successfully.
	JobStatusCompleted JobStatus = "completed"
	// JobStatusFailed indicates the job failed during processing.
	JobStatusFailed JobStatus = "failed"
	// JobStatusSkipped indicates the job was skipped because the operation
	// is not supported on the target OS family.
	JobStatusSkipped JobStatus = "skipped"
	// JobStatusPartialFailure indicates some agents completed and some
	// failed in a broadcast operation.
	JobStatusPartialFailure JobStatus = "partial_failure"
	// JobStatusRetried indicates the job was retried after a failure.
	JobStatusRetried JobStatus = "retried"
	// JobStatusPending indicates the job is queued but not yet processed.
	JobStatusPending JobStatus = "pending"
)

// AgentSchedulingState represents the scheduling state of an agent.
type AgentSchedulingState = string

// Agent scheduling state constants.
const (
	// AgentReady indicates the agent is accepting and processing jobs.
	AgentReady AgentSchedulingState = "Ready"
	// AgentDraining indicates the agent is finishing in-flight jobs
	// but not accepting new ones.
	AgentDraining AgentSchedulingState = "Draining"
	// AgentCordoned indicates the agent is blocked from receiving
	// new jobs until manually uncordoned.
	AgentCordoned AgentSchedulingState = "Cordoned"
)

// ConditionType represents a node or process condition evaluated agent-side.
type ConditionType = string

// Condition type constants.
const (
	ConditionMemoryPressure ConditionType = "MemoryPressure"
	ConditionHighLoad       ConditionType = "HighLoad"
	ConditionDiskPressure   ConditionType = "DiskPressure"
)

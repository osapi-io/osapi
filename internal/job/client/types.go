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

package client

import (
	"context"
	"crypto/ed25519"
	"errors"

	"github.com/nats-io/nats.go/jetstream"
	natsclient "github.com/osapi-io/nats-client/pkg/client"

	"github.com/osapi-io/osapi/internal/job"
)

// PKISigner signs and verifies payloads. Nil when PKI is disabled.
type PKISigner interface {
	// Sign signs the given data with the signer's private key.
	Sign(data []byte) []byte
	// Fingerprint returns the SHA256 fingerprint of the signer's public key.
	Fingerprint() string
	// ControllerPublicKey returns the controller's public key for verifying
	// controller-signed messages. Returns nil if not set.
	ControllerPublicKey() ed25519.PublicKey
}

// AgentKey is what the controller holds for an accepted agent: the key its
// messages are verified against, and the hostname it enrolled under.
type AgentKey struct {
	// MachineID is the permanent host identifier the record is keyed by.
	MachineID string
	// Hostname is the hostname recorded at acceptance. A response claiming
	// a different hostname is not this agent's to answer for.
	Hostname string
	// PublicKey is the key recorded at acceptance.
	PublicKey ed25519.PublicKey
}

// AgentKeyStore looks up an accepted agent's key by machine ID. Nil when the
// controller is not enforcing response verification, which leaves behaviour
// exactly as it was before verification existed.
//
// Implemented on the controller by the enrollment watcher, which is the only
// thing allowed to write the underlying records.
type AgentKeyStore interface {
	LookupAgentKey(
		ctx context.Context,
		machineID string,
	) (*AgentKey, error)
}

// The causes of a rejected response stay distinct so an operator can tell an
// agent that has not re-enrolled yet from one whose messages are being
// forged, and both from a store that cannot be read. None of them ever means
// "treat as verified".
var (
	// ErrResponseNotSigned means the response carried no signed envelope.
	// Expected from an agent that has not been upgraded; never accepted
	// while enforcing.
	ErrResponseNotSigned = errors.New("response is not a signed envelope")

	// ErrResponseKeyUnknown means no key is stored for the machine ID the
	// response claims. Expected during a rollout, before that agent
	// re-enrolls.
	ErrResponseKeyUnknown = errors.New("no stored key for responding agent")

	// ErrResponseStoreUnavailable means the store could not be read. It is
	// never mistaken for "no stored key".
	ErrResponseStoreUnavailable = errors.New("agent key store unavailable")

	// ErrResponseSignatureInvalid means a key that is not the stored key
	// for that agent signed the response.
	ErrResponseSignatureInvalid = errors.New("invalid agent signature on response")

	// ErrResponseHostnameMismatch means the signature verified, but the
	// response claims a hostname the signing agent did not enroll under —
	// an accepted agent answering for a host that is not its own.
	ErrResponseHostnameMismatch = errors.New(
		"response hostname does not match the enrolled hostname",
	)
)

const (
	// DefaultPageSize is the default number of jobs per page.
	DefaultPageSize = 10
	// MaxPageSize is the maximum allowed page size.
	MaxPageSize = 100
)

// JobClient defines the interface for interacting with the jobs system.
type JobClient interface {
	// Generic dispatch operations — used by API handlers to submit jobs
	// without importing typed wrapper methods.
	Query(
		ctx context.Context,
		target string,
		category string,
		operation job.OperationType,
		data any,
	) (string, *job.Response, error)
	QueryBroadcast(
		ctx context.Context,
		target string,
		category string,
		operation job.OperationType,
		data any,
	) (string, map[string]*job.Response, error)
	Modify(
		ctx context.Context,
		target string,
		category string,
		operation job.OperationType,
		data any,
	) (string, *job.Response, error)
	ModifyBroadcast(
		ctx context.Context,
		target string,
		category string,
		operation job.OperationType,
		data any,
	) (string, map[string]*job.Response, error)

	// Job queue management operations
	GetQueueSummary(
		ctx context.Context,
	) (*job.QueueStats, error)
	GetJobStatus(
		ctx context.Context,
		jobID string,
	) (*job.QueuedJob, error)
	ListJobs(
		ctx context.Context,
		statusFilter string,
		limit int,
		offset int,
	) (*ListJobsResult, error)

	// Agent discovery
	ListAgents(
		ctx context.Context,
	) ([]job.AgentInfo, error)
	GetAgent(
		ctx context.Context,
		hostname string,
	) (*job.AgentInfo, error)

	// Agent timeline
	WriteAgentTimelineEvent(
		ctx context.Context,
		hostname, event, message string,
	) error
	GetAgentTimeline(
		ctx context.Context,
		hostname string,
	) ([]job.TimelineEvent, error)

	// Agent drain flag
	CheckDrainFlag(
		ctx context.Context,
		hostname string,
	) bool
	SetDrainFlag(
		ctx context.Context,
		hostname string,
	) error
	DeleteDrainFlag(
		ctx context.Context,
		hostname string,
	) error

	// Job deletion
	DeleteJob(
		ctx context.Context,
		jobID string,
	) error

	// Job retry
	RetryJob(
		ctx context.Context,
		jobID string,
		targetHostname string,
	) (*CreateJobResult, error)

	// Agent operations - used by agents for processing
	// HasJobResponse reports whether this agent has already recorded a
	// response for the given job. Used to detect a redelivered message
	// for a job this agent has already executed, so the operation is not
	// run a second time.
	HasJobResponse(
		ctx context.Context,
		jobID string,
		hostname string,
	) (bool, error)
	WriteStatusEvent(
		ctx context.Context,
		jobID string,
		event string,
		hostname string,
		data map[string]interface{},
	) error
	WriteJobResponse(
		ctx context.Context,
		jobID string,
		hostname string,
		responseData []byte,
		status string,
		errorMsg string,
		changed *bool,
	) error
	ConsumeJobs(
		ctx context.Context,
		streamName string,
		consumerName string,
		handler func(jetstream.Msg) error,
		opts *natsclient.ConsumeOptions,
	) error
	GetJobData(
		ctx context.Context,
		jobKey string,
	) ([]byte, error)
	CreateOrUpdateConsumer(
		ctx context.Context,
		streamName string,
		consumerConfig jetstream.ConsumerConfig,
	) error

	// SetPKISigner wires the signer used to sign outgoing job and response
	// payloads and to verify incoming ones. The keypair a signer needs is
	// not always available at client construction time, so this is called
	// once one is: on the controller when the enrollment watcher loads the
	// controller keypair, and on the agent when PKI enrollment loads the
	// agent keypair. Passing nil disables signing and verification.
	SetPKISigner(
		signer PKISigner,
	)

	// SetAgentKeyStore wires the store used to verify agent responses. Like
	// the signer, it is not available at construction: on the controller it
	// arrives with the enrollment watcher. Passing nil leaves responses
	// unverified, which is the behaviour before this feature.
	SetAgentKeyStore(
		store AgentKeyStore,
	)

	// SetMachineID records the identity stamped on payloads this client
	// signs, so the verifier knows which stored key to check them against.
	// Set on the agent once its identity is resolved; empty on the
	// controller, whose payloads agents verify against the controller key.
	SetMachineID(
		machineID string,
	)
}

// CreateJobResult represents the result of creating a job.
type CreateJobResult struct {
	JobID     string `json:"job_id"`
	Status    string `json:"status"`
	Revision  uint64 `json:"revision"`
	Timestamp string `json:"timestamp"`
}

// ListJobsResult represents the result of listing jobs with pagination.
type ListJobsResult struct {
	Jobs         []*job.QueuedJob
	TotalCount   int
	StatusCounts map[string]int
}

// computedJobStatus represents the computed status from events
type computedJobStatus struct {
	Status      string
	Error       string
	Hostname    string
	UpdatedAt   string
	AgentStates map[string]job.AgentState
	Timeline    []job.TimelineEvent
}

// lightJobInfo holds status derived from KV key names only (no reads).
type lightJobInfo struct {
	Status string
}

// Ensure Client implements JobClient interface
var _ JobClient = (*Client)(nil)

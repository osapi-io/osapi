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

// Package client provides job client operations for NATS JetStream.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/metric"

	"github.com/osapi-io/osapi/internal/job"

	"github.com/osapi-io/osapi/internal/audit"
	"github.com/osapi-io/osapi/internal/secret"
	"github.com/osapi-io/osapi/internal/secret/seal"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// Client provides methods for publishing job requests and retrieving responses.
type Client struct {
	logger     *slog.Logger
	natsClient NATSClient
	// jobsBucket, registryBucket, factsBucket and stateBucket are the KV
	// buckets of the jobs and their responses, the registry, the facts and
	// the agent state. A bucket that the config does not name is "".
	jobsBucket     string
	registryBucket string
	factsBucket    string
	stateBucket    string
	timeout        time.Duration
	streamName     string
	// JSONMarshalFn is the function used to marshal JSON. Defaults to json.Marshal.
	// Exported for testing.
	JSONMarshalFn func(v any) ([]byte, error)
	jobsCreated   metric.Int64Counter
	// pkiSigner signs job payloads. Nil until SetPKISigner.
	pkiSigner PKISigner
	// targetResolver resolves a target (hostname or machine ID) to the
	// value used for NATS subject routing. When nil, the target is used as-is.
	targetResolver func(ctx context.Context, target string) string
	// consumers looks up and removes a consumer whose deliver policy
	// changed. Nil when the client does not make consumers.
	consumers ConsumerStore
	// agentKeyStore verifies agent responses against the key recorded when
	// the agent was accepted. Nil when the controller is not enforcing, in
	// which case responses are handled exactly as they were before
	// verification existed.
	agentKeyStore AgentKeyStore
	// auditLog is where a change that no request made is recorded, or nil.
	auditLog audit.Writer
	// refusedSeen holds each machine ID whose result this client refused,
	// so the first refusal of each is on its timeline and in the audit log
	// once, not on each job.
	refusedSeen sync.Map
	// machineID identifies this client's host on payloads it signs, so the
	// verifier knows which stored key to check them against. Empty on the
	// controller.
	machineID string
}

// Options configures the jobs client.
type Options struct {
	// Timeout for waiting for job responses. Zero is DefaultTimeout.
	Timeout time.Duration
	// JobsBucket is the KV bucket of the jobs and their responses (required).
	JobsBucket string
	// RegistryBucket is the KV bucket of the registry.
	RegistryBucket string
	// FactsBucket is the KV bucket of the agent facts.
	FactsBucket string
	// StateBucket is the KV bucket of the agent state.
	StateBucket string
	// StreamName is the JetStream stream name (used to derive DLQ name).
	StreamName string
	// PKISigner signs job payloads. Nil until SetPKISigner.
	PKISigner PKISigner
	// TargetResolver resolves hostname targets to machine IDs for NATS
	// subject routing. When nil, targets are used as-is.
	TargetResolver func(ctx context.Context, target string) string
	// Consumers looks up and removes a consumer, so CreateOrUpdateConsumer
	// can make again a consumer whose deliver policy changed. When nil, a
	// refused update is returned as is.
	Consumers ConsumerStore
}

// DefaultTimeout is how long a job client waits for job responses when its
// Options name no timeout. With no default, a zero timeout would end each
// wait at once, and each job would time out.
const DefaultTimeout = 30 * time.Second

// New creates a new jobs client using an existing NATS client.
func New(
	logger *slog.Logger,
	natsClient NATSClient,
	opts *Options,
) (*Client, error) {
	if opts == nil {
		return nil, fmt.Errorf("options cannot be nil")
	}
	if opts.JobsBucket == "" {
		return nil, fmt.Errorf("jobs bucket cannot be empty")
	}
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	return &Client{
		logger:         logger.With(slog.String("subsystem", "job.client")),
		natsClient:     natsClient,
		jobsBucket:     opts.JobsBucket,
		registryBucket: opts.RegistryBucket,
		factsBucket:    opts.FactsBucket,
		stateBucket:    opts.StateBucket,
		streamName:     opts.StreamName,
		timeout:        timeout,
		JSONMarshalFn:  json.Marshal,
		pkiSigner:      opts.PKISigner,
		targetResolver: opts.TargetResolver,
		consumers:      opts.Consumers,
	}, nil
}

// SetPKISigner wires the signer used to sign outgoing job and response
// payloads and to verify incoming ones. See the JobClient interface for why
// this is separate from construction. With nil, the job client writes no
// job and no response, and reads no job record.
func (c *Client) SetPKISigner(
	signer PKISigner,
) {
	c.pkiSigner = signer
}

// SetAgentKeyStore wires the store used to verify agent responses. See the
// JobClient interface for why this is separate from construction. With nil,
// the job client refuses each response of an agent and trusts no
// registration.
func (c *Client) SetAgentKeyStore(
	store AgentKeyStore,
) {
	c.agentKeyStore = store
}

// SetAudit wires the audit log. See the JobClient interface.
func (c *Client) SetAudit(
	w audit.Writer,
) {
	c.auditLog = w
}

// SetMachineID records the identity stamped on payloads this client signs.
func (c *Client) SetMachineID(
	machineID string,
) {
	c.machineID = machineID
}

// resolveTarget resolves a target (hostname or machine ID) to the routing
// value for NATS subjects. When no resolver is configured, returns the
// target unchanged.
func (c *Client) resolveTarget(
	ctx context.Context,
	target string,
) string {
	if c.targetResolver != nil {
		return c.targetResolver(ctx, target)
	}
	return target
}

// Query publishes a query job to a single target and waits for the response.
func (c *Client) Query(
	ctx context.Context,
	target string,
	category string,
	operation job.OperationType,
	data any,
) (string, *job.Response, error) {
	resolved := c.resolveTarget(ctx, target)
	req, err := c.request(ctx, types.JobTypeQuery, target, resolved, category, operation, data)
	if err != nil {
		return "", nil, err
	}
	subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, resolved)
	// The span records the target that the caller gave, such as debian-1;
	// the subject has the machine ID that it resolves to.
	jobID, resp, err := c.publishAndWait(ctx, subject, target, req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == types.AgentJobStatusFailed {
		return jobID, nil, job.NewResponseError(resp)
	}

	resp.JobID = jobID
	return jobID, resp, nil
}

// QueryBroadcast publishes a query job to a broadcast target and collects all responses.
func (c *Client) QueryBroadcast(
	ctx context.Context,
	target string,
	category string,
	operation job.OperationType,
	data any,
) (string, map[string]*job.Response, error) {
	resolved := c.resolveTarget(ctx, target)
	req, err := c.request(ctx, types.JobTypeQuery, target, resolved, category, operation, data)
	if err != nil {
		return "", nil, err
	}
	subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, resolved)
	jobID, responses, err := c.publishAndCollect(ctx, subject, target, req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to collect broadcast responses: %w", err)
	}

	return jobID, responses, nil
}

// Modify publishes a modify job to a single target and waits for the response.
func (c *Client) Modify(
	ctx context.Context,
	target string,
	category string,
	operation job.OperationType,
	data any,
) (string, *job.Response, error) {
	resolved := c.resolveTarget(ctx, target)
	req, err := c.request(ctx, types.JobTypeModify, target, resolved, category, operation, data)
	if err != nil {
		return "", nil, err
	}
	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, resolved)
	// The span records the target that the caller gave, such as debian-1;
	// the subject has the machine ID that it resolves to.
	jobID, resp, err := c.publishAndWait(ctx, subject, target, req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == types.AgentJobStatusFailed {
		return jobID, nil, job.NewResponseError(resp)
	}

	resp.JobID = jobID
	return jobID, resp, nil
}

// ModifyBroadcast publishes a modify job to a broadcast target and collects all responses.
func (c *Client) ModifyBroadcast(
	ctx context.Context,
	target string,
	category string,
	operation job.OperationType,
	data any,
) (string, map[string]*job.Response, error) {
	resolved := c.resolveTarget(ctx, target)
	req, err := c.request(ctx, types.JobTypeModify, target, resolved, category, operation, data)
	if err != nil {
		return "", nil, err
	}
	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, resolved)
	jobID, responses, err := c.publishAndCollect(ctx, subject, target, req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to collect broadcast responses: %w", err)
	}

	return jobID, responses, nil
}

// publishAndWait stores a job in KV, publishes a notification, and waits for the agent response.
// Returns the job ID, response, and any error.
func (c *Client) publishAndWait(
	ctx context.Context,
	subject string,
	target string,
	req *job.Request,
) (string, *job.Response, error) {
	req.Timestamp = time.Now()

	jobID := req.JobID
	pubCtx, pubSpan := startPublishSpan(ctx, jobID, subject, target, req)

	// Tell the audit trail which job answered this request, so an entry can be
	// joined to the job's own status timeline. Outside a request — an agent, a
	// background task, a test — there is no sink and this does nothing.
	audit.RecordJobID(ctx, jobID)
	createdTime := req.Timestamp.Format(time.RFC3339)

	operationData := operation(req)

	// Store immutable job data in KV (same structure as CreateJob)
	jobData := map[string]interface{}{
		"id":        jobID,
		"status":    types.JobStatusSubmitted,
		"created":   createdTime,
		"subject":   subject,
		"target":    target,
		"operation": operationData,
	}

	jobJSON, _ := json.Marshal(jobData)

	// Sign the job data with the key of the controller.
	kvPayload, err := c.sign(jobJSON)
	if err != nil {
		return "", nil, endSpan(pubSpan, fmt.Errorf("failed to sign job data: %w", err))
	}

	kvKey := "jobs." + jobID

	c.logger.DebugContext(
		ctx, "kv.put",
		slog.String("key", kvKey),
		slog.String("job_id", jobID),
	)

	if _, err := c.natsClient.KVPut(pubCtx, c.jobsBucket, kvKey, kvPayload); err != nil {
		return "", nil, endSpan(pubSpan, fmt.Errorf("failed to store job in KV: %w", err))
	}

	c.logger.InfoContext(
		ctx, "publishing job request",
		slog.String("job_id", jobID),
		slog.String("subject", subject),
		slog.String("type", string(req.Type)),
	)

	// Publish just the job ID to the stream as a notification.
	// Trace context is propagated via NATS message headers automatically by nats-client.
	if err := c.natsClient.Publish(pubCtx, subject, []byte(jobID)); err != nil {
		return "", nil, endSpan(pubSpan, fmt.Errorf("failed to publish notification: %w", err))
	}
	pubSpan.End()

	if c.jobsCreated != nil {
		c.jobsCreated.Add(ctx, 1)
	}

	// Watch for agent response in KV
	colCtx, colSpan := startCollectSpan(ctx, jobID)
	responsePattern := "responses." + jobID + ".>"
	// The watch ends when this function returns.
	watchCtx, stopWatch := context.WithCancel(colCtx)
	defer stopWatch()
	updates, err := c.natsClient.KVWatch(watchCtx, c.jobsBucket, responsePattern)
	if err != nil {
		return "", nil, endSpan(colSpan, fmt.Errorf("failed to create response watcher: %w", err))
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, job.Wait(ctx, c.timeout))
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			// The caller's ctx ended first: the caller stopped waiting,
			// which is not an agent that did not answer.
			if err := ctx.Err(); err != nil {
				return "", nil, endSpan(colSpan, fmt.Errorf("wait for job response: %w", err))
			}

			return "", nil, endSpan(
				colSpan,
				fmt.Errorf("timeout waiting for job response: %w", timeoutCtx.Err()),
			)
		case entry, ok := <-updates:
			if !ok {
				// The watch ended before the timeout. A nil channel
				// blocks, so the loop waits for the timeout and does
				// not spin on the closed channel.
				updates = nil
				continue
			}

			// A response that does not verify cannot be trusted, so it is
			// a job failure, not data.
			response, verifyErr := c.verifyResponse(ctx, entry.Value())
			if verifyErr != nil {
				c.refused(ctx, jobID, verifyErr)

				return "", nil, endSpan(colSpan, fmt.Errorf(
					"response verification failed for job %s: %w",
					jobID, verifyErr,
				))
			}

			c.logger.InfoContext(
				ctx, "received job response",
				slog.String("job_id", jobID),
				slog.String("status", string(response.Status)),
			)
			c.followRename(ctx, req, response)

			endCollectSpan(colSpan, 1, map[string]*job.Response{response.Hostname: response})

			// The audit entry of the request records the chain of the
			// job, also when the job failed on the host.
			audit.RecordSteps(ctx, jobID, response.Hostname, response.Steps)

			return jobID, response, nil
		}
	}
}

// publishAndCollect stores a job in KV, publishes a notification, and collects
// agent responses. It uses the agent registry to determine how many agents are
// expected to respond and returns immediately once all have replied. The overall
// timeout acts as a safety net if an agent doesn't respond.
// Returns the job ID, collected responses keyed by hostname, and any error.
func (c *Client) publishAndCollect(
	ctx context.Context,
	subject string,
	target string,
	req *job.Request,
) (string, map[string]*job.Response, error) {
	req.Timestamp = time.Now()

	jobID := req.JobID
	pubCtx, pubSpan := startPublishSpan(ctx, jobID, subject, target, req)

	// Tell the audit trail which job answered this request, so an entry can be
	// joined to the job's own status timeline. Outside a request — an agent, a
	// background task, a test — there is no sink and this does nothing.
	audit.RecordJobID(ctx, jobID)
	createdTime := req.Timestamp.Format(time.RFC3339)

	operationData := operation(req)

	// Store immutable job data in KV
	jobData := map[string]interface{}{
		"id":        jobID,
		"status":    types.JobStatusSubmitted,
		"created":   createdTime,
		"subject":   subject,
		"target":    target,
		"operation": operationData,
	}

	jobJSON, _ := json.Marshal(jobData)

	// Sign the job data with the key of the controller.
	kvPayload, err := c.sign(jobJSON)
	if err != nil {
		return "", nil, endSpan(pubSpan, fmt.Errorf("failed to sign job data: %w", err))
	}

	kvKey := "jobs." + jobID

	c.logger.DebugContext(
		ctx, "kv.put",
		slog.String("key", kvKey),
		slog.String("job_id", jobID),
	)

	if _, err := c.natsClient.KVPut(pubCtx, c.jobsBucket, kvKey, kvPayload); err != nil {
		return "", nil, endSpan(pubSpan, fmt.Errorf("failed to store job in KV: %w", err))
	}

	// Determine expected agents from registry for early completion and
	// timeout reporting. If ListAgents fails, fall back to full timeout.
	var expectedHostnames []string
	agents, err := c.ListAgents(pubCtx)
	if err != nil {
		c.logger.WarnContext(
			ctx, "failed to list agents for broadcast count, using full timeout",
			slog.String("error", err.Error()),
		)
	} else {
		// A snapshot, taken before the job is published. An agent that joins in
		// the window between here and the first response is not expected, so it
		// is not waited for and gets no timeout row: its own response still
		// arrives and is collected if it answers in time. Re-listing mid-collect
		// would trade that for a broadcast whose expected set changes while it
		// is being satisfied.
		expectedHostnames = job.ExpectedAgentHostnames(agents, target)
		c.logger.DebugContext(
			ctx, "broadcast expected agent count",
			slog.String("target", target),
			slog.Int("expected_count", len(expectedHostnames)),
		)
	}

	c.logger.InfoContext(
		ctx, "publishing broadcast job request",
		slog.String("job_id", jobID),
		slog.String("subject", subject),
		slog.String("type", string(req.Type)),
	)

	// Publish just the job ID to the stream as a notification.
	// Trace context is propagated via NATS message headers automatically by nats-client.
	if err := c.natsClient.Publish(pubCtx, subject, []byte(jobID)); err != nil {
		return "", nil, endSpan(pubSpan, fmt.Errorf("failed to publish notification: %w", err))
	}
	pubSpan.End()

	if c.jobsCreated != nil {
		c.jobsCreated.Add(ctx, 1)
	}

	// Watch for agent responses in KV
	colCtx, colSpan := startCollectSpan(ctx, jobID)
	responsePattern := "responses." + jobID + ".>"
	// The watch ends when this function returns.
	watchCtx, stopWatch := context.WithCancel(colCtx)
	defer stopWatch()
	updates, err := c.natsClient.KVWatch(watchCtx, c.jobsBucket, responsePattern)
	if err != nil {
		return "", nil, endSpan(colSpan, fmt.Errorf("failed to create response watcher: %w", err))
	}

	responses := make(map[string]*job.Response)

	// Overall timeout as safety net
	timeoutCtx, cancel := context.WithTimeout(ctx, job.Wait(ctx, c.timeout))
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			// The caller's ctx ended first: no host timed out, so the
			// answer has no timeout rows.
			if err := ctx.Err(); err != nil {
				return "", nil, endSpan(
					colSpan,
					fmt.Errorf("wait for broadcast responses: %w", err),
				)
			}
			if len(responses) == 0 {
				return "", nil, endSpan(colSpan, fmt.Errorf(
					"timeout waiting for broadcast responses: no agents responded",
				))
			}

			// Inject timeout entries for agents that didn't respond.
			for _, hostname := range expectedHostnames {
				if _, ok := responses[hostname]; !ok {
					// Not failed: this agent never answered, and reporting
					// that as a failure tells an operator the operation ran
					// and did not succeed, which nobody knows.
					responses[hostname] = &job.Response{
						JobID:     jobID,
						Hostname:  hostname,
						Status:    types.AgentJobStatusTimeout,
						ErrorCode: job.ErrorCodeTimeout,
						Error:     "timeout: agent did not respond",
					}
				}
			}
			endCollectSpan(colSpan, len(expectedHostnames), responses)

			return jobID, responses, nil
		case entry, ok := <-updates:
			if !ok {
				// The watch ended before the timeout. A nil channel
				// blocks, so the loop waits for the timeout and does
				// not spin on the closed channel.
				updates = nil
				continue
			}

			// A response that does not verify cannot be trusted, so it is
			// dropped. The agent that sent it then shows as missing rather
			// than answered, and surfaces as a timeout for that hostname if
			// it never sends a good response.
			response, verifyErr := c.verifyResponse(ctx, entry.Value())
			if verifyErr != nil {
				c.logger.WarnContext(
					ctx, "broadcast response verification failed",
					slog.String("job_id", jobID),
					slog.String("error", verifyErr.Error()),
				)
				// A refusal is the host's answer: it is a row, not a
				// timeout later.
				if row, ok := c.refused(ctx, jobID, verifyErr); ok {
					responses[row.Hostname] = row
				}

				continue
			}

			hostname := response.Hostname
			if hostname == "" {
				hostname = "unknown"
			}

			c.logger.InfoContext(
				ctx, "received broadcast response",
				slog.String("job_id", jobID),
				slog.String("hostname", hostname),
				slog.String("status", string(response.Status)),
			)
			c.followRename(ctx, req, response)

			responses[hostname] = response
			audit.RecordSteps(ctx, jobID, hostname, response.Steps)

			// Return early when each expected agent has responded. A
			// count is not enough: an agent that joined after the
			// snapshot, or an "unknown" or refused row, is a response
			// too, and counting it would return before an expected agent
			// answered and drop that agent's timeout row.
			if len(expectedHostnames) > 0 && allAnswered(expectedHostnames, responses) {
				endCollectSpan(colSpan, len(expectedHostnames), responses)

				return jobID, responses, nil
			}
		}
	}
}

// allAnswered reports whether responses holds a row for each expected
// hostname.
func allAnswered(
	expected []string,
	responses map[string]*job.Response,
) bool {
	for _, hostname := range expected {
		if _, ok := responses[hostname]; !ok {
			return false
		}
	}

	return true
}

// operation is the operation of the job data in the KV: its type, its
// data, and the sealed job keys when the data has sealed secrets.
func operation(
	req *job.Request,
) map[string]interface{} {
	op := map[string]interface{}{
		"type": req.Operation,
		"data": req.Data,
	}
	if len(req.SealedKeys) > 0 {
		op["sealed_keys"] = req.SealedKeys
		op["sealed_from"] = req.SealedFrom
	}

	return op
}

// newSealer makes the Sealer of a job. A test replaces it.
var newSealer = seal.New

// request is the job of data for target, which resolves to resolved, with
// a new job ID. When data holds a secret, each secret is sealed with a new
// job key, and the job key is sealed to each agent that the target can
// reach and that has an encryption key: JetStream and the job KV never see
// the value. A job client with no agent keys, which only the controller
// has, refuses a job with a secret, so no secret goes in plain text.
func (c *Client) request(
	ctx context.Context,
	typ types.JobType,
	target string,
	resolved string,
	category string,
	operation job.OperationType,
	data any,
) (*job.Request, error) {
	req := &job.Request{
		JobID:     uuid.New().String(),
		Type:      typ,
		Category:  category,
		Operation: operation,
	}
	if !secret.Has(data) {
		dataBytes, err := json.Marshal(secret.Plain(data))
		if err != nil {
			return nil, fmt.Errorf("marshal data: %w", err)
		}
		req.Data = dataBytes

		return req, nil
	}

	if c.agentKeyStore == nil {
		return nil, ErrNoAgentKeys
	}

	sealer, err := newSealer(req.JobID)
	if err != nil {
		return nil, err
	}
	dataBytes, err := json.Marshal(secret.PlainFunc(data, sealer.Value))
	if err != nil {
		return nil, fmt.Errorf("marshal data: %w", err)
	}
	req.Data = dataBytes
	req.SealedFrom = sealer.From()
	req.SealedKeys = map[string]string{}

	agents, err := c.ListAgents(ctx)
	if err != nil {
		return nil, fmt.Errorf("list the agents to seal the job to: %w", err)
	}
	for _, a := range job.Recipients(agents, target, resolved) {
		key, err := c.agentKeyStore.LookupAgentKey(ctx, a.MachineID)
		if err != nil || key.EncryptionKey == "" {
			// The agent reports the job as not sealed to it, and does
			// not run it.
			c.logger.WarnContext(ctx, "the job is not sealed to an agent with no encryption key",
				slog.String("job_id", req.JobID),
				slog.String("machine_id", a.MachineID))

			continue
		}
		sealed, err := sealer.KeyFor(key.EncryptionKey)
		if err != nil {
			return nil, err
		}
		req.SealedKeys[a.MachineID] = sealed
	}

	return req, nil
}

// sign is payload in a signed envelope, with the key of this job client.
// A job client with no signer writes nothing: PKI is always on, and a
// reader refuses an unsigned value.
func (c *Client) sign(
	payload []byte,
) ([]byte, error) {
	if c.pkiSigner == nil {
		return nil, ErrNoSigner
	}

	return wrapInSignedEnvelope(c.pkiSigner, c.machineID, payload)
}

// verifyResponse is the response of an agent, verified against the key
// stored for the agent that claims to have sent it. A job client with no
// agent keys refuses each response, so no unsigned response is read as an
// answer.
func (c *Client) verifyResponse(
	ctx context.Context,
	data []byte,
) (*job.Response, error) {
	if c.agentKeyStore == nil {
		return nil, ErrNoAgentKeys
	}

	return verifyAgentResponse(ctx, c.agentKeyStore, data)
}

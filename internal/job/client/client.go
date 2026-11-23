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
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/metric"

	"github.com/osapi-io/osapi/internal/job"

	"github.com/osapi-io/osapi/internal/audit"
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
	// pkiSigner signs job payloads when PKI is enabled. Nil when disabled.
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
	// machineID identifies this client's host on payloads it signs, so the
	// verifier knows which stored key to check them against. Empty on the
	// controller.
	machineID string
}

// Options configures the jobs client.
type Options struct {
	// Timeout for waiting for job responses (default: 30s)
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
	// PKISigner signs job payloads when PKI is enabled. Nil when disabled.
	PKISigner PKISigner
	// TargetResolver resolves hostname targets to machine IDs for NATS
	// subject routing. When nil, targets are used as-is.
	TargetResolver func(ctx context.Context, target string) string
	// Consumers looks up and removes a consumer, so CreateOrUpdateConsumer
	// can make again a consumer whose deliver policy changed. When nil, a
	// refused update is returned as is.
	Consumers ConsumerStore
}

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

	return &Client{
		logger:         logger.With(slog.String("subsystem", "job.client")),
		natsClient:     natsClient,
		jobsBucket:     opts.JobsBucket,
		registryBucket: opts.RegistryBucket,
		factsBucket:    opts.FactsBucket,
		stateBucket:    opts.StateBucket,
		streamName:     opts.StreamName,
		timeout:        opts.Timeout,
		JSONMarshalFn:  json.Marshal,
		pkiSigner:      opts.PKISigner,
		targetResolver: opts.TargetResolver,
		consumers:      opts.Consumers,
	}, nil
}

// SetPKISigner wires the signer used to sign outgoing job and response
// payloads and to verify incoming ones. See the JobClient interface for why
// this is separate from construction. Passing nil disables signing and
// verification.
func (c *Client) SetPKISigner(
	signer PKISigner,
) {
	c.pkiSigner = signer
}

// SetAgentKeyStore wires the store used to verify agent responses. See the
// JobClient interface for why this is separate from construction. Passing nil
// leaves responses unverified.
func (c *Client) SetAgentKeyStore(
	store AgentKeyStore,
) {
	c.agentKeyStore = store
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
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return "", nil, fmt.Errorf("marshal data: %w", err)
	}

	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  category,
		Operation: operation,
		Data:      json.RawMessage(dataBytes),
	}

	resolved := c.resolveTarget(ctx, target)
	subject := job.BuildSubjectFromTarget(job.JobsQueryPrefix, resolved)
	// The span records the target that the caller gave, such as debian-1;
	// the subject has the machine ID that it resolves to.
	jobID, resp, err := c.publishAndWait(ctx, subject, target, req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == job.StatusFailed {
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
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return "", nil, fmt.Errorf("marshal data: %w", err)
	}

	req := &job.Request{
		Type:      job.TypeQuery,
		Category:  category,
		Operation: operation,
		Data:      json.RawMessage(dataBytes),
	}

	resolved := c.resolveTarget(ctx, target)
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
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return "", nil, fmt.Errorf("marshal data: %w", err)
	}

	req := &job.Request{
		Type:      job.TypeModify,
		Category:  category,
		Operation: operation,
		Data:      json.RawMessage(dataBytes),
	}

	resolved := c.resolveTarget(ctx, target)
	subject := job.BuildSubjectFromTarget(job.JobsModifyPrefix, resolved)
	// The span records the target that the caller gave, such as debian-1;
	// the subject has the machine ID that it resolves to.
	jobID, resp, err := c.publishAndWait(ctx, subject, target, req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to publish and wait: %w", err)
	}

	if resp.Status == job.StatusFailed {
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
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return "", nil, fmt.Errorf("marshal data: %w", err)
	}

	req := &job.Request{
		Type:      job.TypeModify,
		Category:  category,
		Operation: operation,
		Data:      json.RawMessage(dataBytes),
	}

	resolved := c.resolveTarget(ctx, target)
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
	// Generate job ID if not provided
	if req.JobID == "" {
		req.JobID = uuid.New().String()
	}
	req.Timestamp = time.Now()

	jobID := req.JobID
	pubCtx, pubSpan := startPublishSpan(ctx, jobID, subject, target, req)

	// Tell the audit trail which job answered this request, so an entry can be
	// joined to the job's own status timeline. Outside a request — an agent, a
	// background task, a test — there is no sink and this does nothing.
	audit.RecordJobID(ctx, jobID)
	createdTime := req.Timestamp.Format(time.RFC3339)

	operationData := map[string]interface{}{
		"type": req.Operation,
		"data": req.Data,
	}

	// Store immutable job data in KV (same structure as CreateJob)
	jobData := map[string]interface{}{
		"id":        jobID,
		"status":    "unprocessed",
		"created":   createdTime,
		"subject":   subject,
		"target":    target,
		"operation": operationData,
	}

	jobJSON, _ := json.Marshal(jobData)

	// Sign the job data when PKI is enabled.
	kvPayload := jobJSON
	if c.pkiSigner != nil {
		signed, signErr := wrapInSignedEnvelope(c.pkiSigner, c.machineID, jobJSON)
		if signErr != nil {
			return "", nil, endSpan(pubSpan, fmt.Errorf("failed to sign job data: %w", signErr))
		}
		kvPayload = signed
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

	timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
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

			// Unwrap signed envelope if present. A verification failure
			// means the response cannot be trusted, so it is treated as a
			// job failure rather than fallen through as unverified data
			// that happens to fail to unmarshal into a useful response.
			responseData := entry.Value()

			switch {
			case c.agentKeyStore != nil:
				// Enforcing: the response is verified against the key
				// stored for the agent that claims to have sent it, and
				// the job fails rather than returning unverified data.
				verified, verifyErr := verifyAgentResponse(
					ctx,
					c.agentKeyStore,
					responseData,
				)
				if verifyErr != nil {
					return "", nil, endSpan(colSpan, fmt.Errorf(
						"response verification failed for job %s: %w",
						jobID, verifyErr,
					))
				}
				responseData = verified
			case c.pkiSigner != nil:
				unwrapped, _, unwrapErr := unwrapSignedEnvelope(
					responseData,
					c.pkiSigner.ControllerPublicKey(),
				)
				if unwrapErr != nil {
					return "", nil, endSpan(colSpan, fmt.Errorf(
						"response signature verification failed for job %s: %w",
						jobID, unwrapErr,
					))
				}
				responseData = unwrapped
			}

			var response job.Response
			if err := json.Unmarshal(responseData, &response); err != nil {
				return "", nil, endSpan(
					colSpan,
					fmt.Errorf("failed to unmarshal response: %w", err),
				)
			}

			c.logger.InfoContext(
				ctx, "received job response",
				slog.String("job_id", jobID),
				slog.String("status", string(response.Status)),
			)

			endCollectSpan(colSpan, 1, map[string]*job.Response{response.Hostname: &response})

			return jobID, &response, nil
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
	// Generate job ID if not provided
	if req.JobID == "" {
		req.JobID = uuid.New().String()
	}
	req.Timestamp = time.Now()

	jobID := req.JobID
	pubCtx, pubSpan := startPublishSpan(ctx, jobID, subject, target, req)

	// Tell the audit trail which job answered this request, so an entry can be
	// joined to the job's own status timeline. Outside a request — an agent, a
	// background task, a test — there is no sink and this does nothing.
	audit.RecordJobID(ctx, jobID)
	createdTime := req.Timestamp.Format(time.RFC3339)

	operationData := map[string]interface{}{
		"type": req.Operation,
		"data": req.Data,
	}

	// Store immutable job data in KV
	jobData := map[string]interface{}{
		"id":        jobID,
		"status":    "unprocessed",
		"created":   createdTime,
		"subject":   subject,
		"target":    target,
		"operation": operationData,
	}

	jobJSON, _ := json.Marshal(jobData)

	// Sign the job data when PKI is enabled.
	kvPayload := jobJSON
	if c.pkiSigner != nil {
		signed, signErr := wrapInSignedEnvelope(c.pkiSigner, c.machineID, jobJSON)
		if signErr != nil {
			return "", nil, endSpan(pubSpan, fmt.Errorf("failed to sign job data: %w", signErr))
		}
		kvPayload = signed
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
	timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
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
						Status:    job.StatusTimeout,
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

			// Unwrap signed envelope if present. A verification failure
			// means this response cannot be trusted, so it is dropped
			// rather than kept as unverified data — same treatment as an
			// unmarshalable response below. The agent that sent it then
			// shows as missing rather than answered, and surfaces as a
			// timeout for that hostname if it never sends a good response.
			responseData := entry.Value()

			switch {
			case c.agentKeyStore != nil:
				// Enforcing: a response that does not verify is not this
				// agent's reply. It is dropped, so the agent shows as not
				// having answered rather than as answered by whoever sent
				// it.
				verified, verifyErr := verifyAgentResponse(
					ctx,
					c.agentKeyStore,
					responseData,
				)
				if verifyErr != nil {
					c.logger.WarnContext(
						ctx, "broadcast response verification failed",
						slog.String("job_id", jobID),
						slog.String("error", verifyErr.Error()),
					)

					continue
				}
				responseData = verified
			case c.pkiSigner != nil:
				unwrapped, _, unwrapErr := unwrapSignedEnvelope(
					responseData,
					c.pkiSigner.ControllerPublicKey(),
				)
				if unwrapErr != nil {
					c.logger.WarnContext(
						ctx, "broadcast response signature verification failed",
						slog.String("job_id", jobID),
						slog.String("error", unwrapErr.Error()),
					)

					continue
				}
				responseData = unwrapped
			}

			var response job.Response
			if err := json.Unmarshal(responseData, &response); err != nil {
				c.logger.WarnContext(
					ctx, "failed to unmarshal broadcast response",
					slog.String("job_id", jobID),
					slog.String("error", err.Error()),
				)
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

			responses[hostname] = &response

			// Return early when all expected agents have responded.
			if len(expectedHostnames) > 0 && len(responses) >= len(expectedHostnames) {
				endCollectSpan(colSpan, len(expectedHostnames), responses)

				return jobID, responses, nil
			}
		}
	}
}

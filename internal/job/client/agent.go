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
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/osapi-io/osapi/internal/telemetry/opspan"
	"go.opentelemetry.io/otel/attribute"

	"github.com/nats-io/nats.go/jetstream"
	natsclient "github.com/osapi-io/nats-client/pkg/client"

	"github.com/osapi-io/osapi/internal/job"
)

// HasJobResponse reports whether this agent has already written a response
// for the given job. A redelivered message must not re-execute an operation
// this agent has already answered, so the handler checks this before running
// the operation. The check is scoped to hostname so broadcast jobs, where
// every targeted agent responds independently, remain unaffected.
func (c *Client) HasJobResponse(
	ctx context.Context,
	jobID string,
	hostname string,
) (bool, error) {
	filter := fmt.Sprintf("responses.%s.%s.>", jobID, sanitizeKeyForNATS(hostname))
	keys, err := c.natsClient.KVKeys(ctx, c.jobsBucket, filter)
	if err != nil {
		return false, fmt.Errorf("failed to check for existing job response: %w", err)
	}
	return len(keys) > 0, nil
}

// WriteStatusEvent writes an append-only status event for a job.
// This eliminates race conditions by never updating existing keys.
func (c *Client) WriteStatusEvent(
	ctx context.Context,
	jobID, event, hostname string,
	data map[string]interface{},
) error {
	// Capture time once for consistency across key and payload
	now := time.Now()

	// Create unique key with nanosecond timestamp to ensure uniqueness
	statusKey := fmt.Sprintf("status.%s.%s.%s.%d",
		jobID, event, sanitizeKeyForNATS(hostname), now.UnixNano())

	// Build event data
	eventData := map[string]interface{}{
		"job_id":    jobID,
		"event":     event,
		"hostname":  hostname,
		"timestamp": now.Format(time.RFC3339Nano),
		"unix_nano": now.UnixNano(),
	}

	// Add any additional data
	if data != nil {
		eventData["data"] = data
	}

	// Marshal event data
	eventJSON, err := json.Marshal(eventData)
	if err != nil {
		return fmt.Errorf("failed to marshal status event: %w", err)
	}

	// Write to KV - this always succeeds due to unique key
	_, err = c.natsClient.KVPut(ctx, c.jobsBucket, statusKey, eventJSON)
	if err != nil {
		return fmt.Errorf("failed to write status event: %w", err)
	}

	c.logger.Debug(
		"wrote status event",
		slog.String("job_id", jobID),
		slog.String("event", event),
		slog.String("hostname", hostname),
		slog.String("key", statusKey),
	)

	return nil
}

// WriteJobResponse stores agent response data for a job.
func (c *Client) WriteJobResponse(
	ctx context.Context,
	jobID, hostname string,
	responseData []byte,
	status string,
	errorMsg string,
	errorCode job.ErrorCode,
	changed *bool,
	steps []types.Step,
) error {
	// Create response key with timestamp to ensure uniqueness
	responseKey := fmt.Sprintf("responses.%s.%s.%d",
		jobID, sanitizeKeyForNATS(hostname), time.Now().UnixNano())

	// Build response structure
	response := job.Response{
		Status:    job.Status(status),
		Data:      responseData,
		Error:     errorMsg,
		ErrorCode: errorCode,
		Changed:   changed,
		Steps:     steps,
		Hostname:  hostname,
		Timestamp: time.Now(),
	}

	// Marshal response (Response fields are always marshalable)
	responseJSON, _ := json.Marshal(response)

	// Sign the response with the key of the agent.
	kvPayload, err := c.sign(responseJSON)
	if err != nil {
		return fmt.Errorf("failed to sign response: %w", err)
	}

	// Store in KV
	_, err = c.natsClient.KVPut(ctx, c.jobsBucket, responseKey, kvPayload)
	if err != nil {
		return fmt.Errorf("failed to store job response: %w", err)
	}

	c.logger.Debug(
		"wrote job response",
		slog.String("job_id", jobID),
		slog.String("hostname", hostname),
		slog.String("status", status),
		slog.String("key", responseKey),
	)

	return nil
}

// ConsumeJobs sets up message consumption for job processing.
func (c *Client) ConsumeJobs(
	ctx context.Context,
	streamName, consumerName string,
	handler natsclient.JetStreamMessageHandler,
	opts *natsclient.ConsumeOptions,
) error {
	// Use the natsClient to consume messages with the provided consumer name
	return c.natsClient.Consume(ctx, streamName, consumerName, handler, opts)
}

// GetJobData retrieves raw job data from the KV store.
func (c *Client) GetJobData(
	ctx context.Context,
	jobKey string,
) ([]byte, error) {
	c.logger.Debug(
		"kv.get",
		slog.String("key", jobKey),
	)
	entry, err := c.natsClient.KVGet(ctx, c.jobsBucket, jobKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get job data for key %s: %w", jobKey, err)
	}
	return entry.Value(), nil
}

// CreateOrUpdateConsumer creates or updates a JetStream consumer.
//
// NATS refuses to change the deliver policy of a consumer that exists. When
// the update fails and the existing consumer has a different deliver policy,
// such as a consumer that an earlier build made, the client removes it and
// makes it again with consumerConfig. Thus an upgraded agent gets its
// consumers without a person. The new consumer starts at its own deliver
// policy, not at the position of the old one.
func (c *Client) CreateOrUpdateConsumer(
	ctx context.Context,
	streamName string,
	consumerConfig jetstream.ConsumerConfig,
) error {
	err := c.natsClient.ConsumerCreateOrUpdate(ctx, streamName, consumerConfig)
	if err == nil || c.consumers == nil {
		return err
	}

	name := consumerConfig.Durable
	existing, lookupErr := c.consumers.Consumer(ctx, streamName, name)
	if lookupErr != nil {
		return err
	}
	old := existing.CachedInfo().Config.DeliverPolicy
	if old == consumerConfig.DeliverPolicy {
		return err
	}

	if err := c.consumers.DeleteConsumer(ctx, streamName, name); err != nil {
		return fmt.Errorf("delete consumer %s to change its deliver policy: %w", name, err)
	}

	c.logger.Warn(
		"made the consumer again with a new deliver policy",
		slog.String("stream", streamName),
		slog.String("consumer", name),
		slog.String("old_deliver_policy", old.String()),
		slog.String("new_deliver_policy", consumerConfig.DeliverPolicy.String()),
	)

	return c.natsClient.ConsumerCreateOrUpdate(ctx, streamName, consumerConfig)
}

// ListAgents reads the agent registry KV bucket and returns all registered
// agents. Agents register via heartbeat, so only live agents appear.
func (c *Client) listAgents(
	ctx context.Context,
) ([]job.AgentInfo, error) {
	if c.registryBucket == "" {
		return nil, fmt.Errorf("agent registry not configured")
	}

	keys, err := c.natsClient.KVKeys(ctx, c.registryBucket)
	if err != nil {
		return nil, fmt.Errorf("failed to list registry keys: %w", err)
	}

	agents := make([]job.AgentInfo, 0, len(keys))
	for _, key := range keys {
		if !strings.HasPrefix(key, "agents.") {
			continue
		}

		entry, err := c.natsClient.KVGet(ctx, c.registryBucket, key)
		if err != nil {
			continue
		}

		var reg job.AgentRegistration
		if err := json.Unmarshal(entry.Value(), &reg); err != nil {
			continue
		}

		info := agentInfoFromRegistration(&reg)
		c.applyRegistrationTrust(ctx, &reg, &info)
		c.mergeFacts(ctx, &info)
		c.overlayDrainState(ctx, &info)

		agents = append(agents, info)
	}

	return agents, nil
}

// GetAgent reads a single agent's registration from the KV registry.
// It tries a direct machine ID lookup first, then falls back to scanning
// all agents by hostname.
func (c *Client) getAgent(
	ctx context.Context,
	target string,
) (*job.AgentInfo, error) {
	if c.registryBucket == "" {
		return nil, fmt.Errorf("agent registry not configured")
	}

	// Try as machine ID first (direct KV lookup).
	key := "agents." + target
	entry, err := c.natsClient.KVGet(ctx, c.registryBucket, key)
	if err == nil {
		var reg job.AgentRegistration
		if err := json.Unmarshal(entry.Value(), &reg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal agent registration: %w", err)
		}

		info := agentInfoFromRegistration(&reg)
		c.applyRegistrationTrust(ctx, &reg, &info)
		c.mergeFacts(ctx, &info)
		c.overlayDrainState(ctx, &info)

		timeline, _ := c.GetAgentTimeline(ctx, reg.Hostname)
		if len(timeline) > 0 {
			info.Timeline = timeline
		}

		return &info, nil
	}

	// Fall back to hostname scan.
	agents, listErr := c.ListAgents(ctx)
	if listErr != nil {
		return nil, fmt.Errorf("agent %s: %w", target, job.ErrNotFound)
	}

	for i := range agents {
		if agents[i].Hostname == target {
			timeline, _ := c.GetAgentTimeline(ctx, target)
			if len(timeline) > 0 {
				agents[i].Timeline = timeline
			}

			return &agents[i], nil
		}
	}

	return nil, fmt.Errorf("agent %s: %w", target, job.ErrNotFound)
}

// overlayDrainState sets the agent's state by job.EffectiveState, from the
// state the agent reported at its last heartbeat and its drain flag. A
// drain the agent has not applied yet is Draining. Drain flags are stored
// in the agent-state KV bucket (no TTL).
func (c *Client) overlayDrainState(
	ctx context.Context,
	info *job.AgentInfo,
) {
	if c.stateBucket == "" {
		return
	}

	key := "drain." + info.MachineID
	_, err := c.natsClient.KVGet(ctx, c.stateBucket, key)
	info.State = job.EffectiveState(info.State, err == nil)
}

// mergeFacts reads facts from the facts KV bucket and merges them into the
// AgentInfo. If factsKV is nil or the read fails, this is a no-op.
func (c *Client) mergeFacts(
	ctx context.Context,
	info *job.AgentInfo,
) {
	if c.factsBucket == "" {
		return
	}

	key := "facts." + info.MachineID
	entry, err := c.natsClient.KVGet(ctx, c.factsBucket, key)
	if err != nil {
		return
	}

	var facts job.FactsRegistration
	if err := json.Unmarshal(entry.Value(), &facts); err != nil {
		return
	}

	info.Architecture = facts.Architecture
	info.KernelVersion = facts.KernelVersion
	info.CPUCount = facts.CPUCount
	info.FQDN = facts.FQDN
	info.ServiceMgr = facts.ServiceMgr
	info.PackageMgr = facts.PackageMgr
	info.Interfaces = facts.Interfaces
	info.PrimaryInterface = facts.PrimaryInterface
	info.Routes = facts.Routes
	info.Facts = facts.Facts
}

// agentInfoFromRegistration maps an AgentRegistration to an AgentInfo.
func agentInfoFromRegistration(
	reg *job.AgentRegistration,
) job.AgentInfo {
	return job.AgentInfo{
		MachineID:       reg.MachineID,
		Hostname:        reg.Hostname,
		Labels:          reg.Labels,
		RegisteredAt:    reg.RegisteredAt,
		StartedAt:       reg.StartedAt,
		OSInfo:          reg.OSInfo,
		Uptime:          reg.Uptime,
		LoadAverages:    reg.LoadAverages,
		MemoryStats:     reg.MemoryStats,
		AgentVersion:    reg.AgentVersion,
		MissingPrograms: reg.MissingPrograms,
		Conditions:      reg.Conditions,
		State:           reg.State,
		Fingerprint:     reg.Fingerprint,
	}
}

// applyRegistrationTrust records whether this registration may decide where
// work goes.
//
// The machine ID selects which stored record to check against, so a
// registration naming another agent is verified against that agent's key and
// fails unless it was actually signed by it. The hostname must also match the
// one recorded at acceptance: an accepted agent is not entitled to answer for
// a host it did not enrol as, which is the claim GHSA-j73r describes.
//
// The fingerprint needs no separate check. It is inside the signed bytes, so a
// wrong one would have to be signed by the right key.
//
// A job client with no store, which is not the controller's, trusts no
// registration: PKI is always on.
func (c *Client) applyRegistrationTrust(
	ctx context.Context,
	reg *job.AgentRegistration,
	info *job.AgentInfo,
) {
	if c.agentKeyStore == nil {
		info.Verified = false
		info.KeyStored = false

		return
	}

	record, err := c.agentKeyStore.LookupAgentKey(ctx, reg.MachineID)
	if err != nil {
		// No stored key, or the store could not be read. Either way this
		// registration is not authoritative, and the fleet view shows why.
		info.Verified = false
		info.KeyStored = false

		return
	}

	info.KeyStored = true

	signedByStoredKey := false

	for _, key := range record.AcceptableKeys(nowFn()) {
		if job.VerifyRegistration(reg, key) {
			signedByStoredKey = true

			break
		}
	}

	if !signedByStoredKey {
		info.Verified = false

		return
	}

	if reg.Hostname != record.Hostname {
		c.logger.WarnContext(
			ctx, "registration claims a hostname the agent did not enrol as",
			slog.String("machine_id", reg.MachineID),
			slog.String("enrolled_as", record.Hostname),
			slog.String("claimed", reg.Hostname),
		)

		info.Verified = false

		return
	}

	info.Verified = true
}

// sanitizeKeyForNATS sanitizes a string for use as a NATS key.
func sanitizeKeyForNATS(
	input string,
) string {
	// Replace invalid characters with underscores
	reg := regexp.MustCompile(`[^a-zA-Z0-9_\-]`)
	return reg.ReplaceAllString(input, "_")
}

// ListAgents returns each agent in the registry, in an agent.list span.
func (c *Client) ListAgents(
	ctx context.Context,
) ([]job.AgentInfo, error) {
	ctx, span := opspan.Start(ctx, tracerName, "agent.list")
	agents, err := c.listAgents(ctx)
	if err == nil {
		span.SetAttributes(attribute.Int("osapi.agent.count", len(agents)))
	}

	return agents, opspan.End(span, err)
}

// GetAgent returns one agent by its hostname or machine ID, in an agent.get
// span.
func (c *Client) GetAgent(
	ctx context.Context,
	target string,
) (*job.AgentInfo, error) {
	ctx, span := opspan.Start(ctx, tracerName, "agent.get",
		attribute.String("osapi.target", target),
	)
	info, err := c.getAgent(ctx, target)

	return info, opspan.End(span, err)
}

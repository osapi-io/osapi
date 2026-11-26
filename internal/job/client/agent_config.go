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

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"go.opentelemetry.io/otel/attribute"

	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/telemetry/opspan"
)

// GetAgentConfig is the config that the controller holds for the agent
// with machineID, in an agent.config.get span. An agent that has none has
// the zero config, version 0.
func (c *Client) GetAgentConfig(
	ctx context.Context,
	machineID string,
) (job.AgentConfig, error) {
	ctx, span := opspan.Start(ctx, tracerName, "agent.config.get",
		attribute.String("osapi.agent.machine_id", machineID),
	)
	cfg, err := c.agentConfig(ctx, machineID)

	return cfg, opspan.End(span, err)
}

// agentConfig reads the config of machineID.
func (c *Client) agentConfig(
	ctx context.Context,
	machineID string,
) (job.AgentConfig, error) {
	if c.stateBucket == "" {
		return job.AgentConfig{}, fmt.Errorf("agent state bucket not configured")
	}
	entry, err := c.natsClient.KVGet(ctx, c.stateBucket, job.AgentConfigKey(machineID))
	if errors.Is(err, jetstream.ErrKeyNotFound) {
		return job.AgentConfig{}, nil
	}
	if err != nil {
		return job.AgentConfig{}, fmt.Errorf("read the config of %s: %w", machineID, err)
	}
	var cfg job.AgentConfig
	if err := json.Unmarshal(entry.Value(), &cfg); err != nil {
		return job.AgentConfig{}, fmt.Errorf("the config of %s: %w", machineID, err)
	}

	return cfg, nil
}

// SetAgentConfig sets the labels and the max jobs of set as the config of
// the agent with machineID, as its next version, in an agent.config.set
// span.
func (c *Client) SetAgentConfig(
	ctx context.Context,
	machineID string,
	set job.AgentConfig,
) (job.AgentConfig, error) {
	ctx, span := opspan.Start(ctx, tracerName, "agent.config.set",
		attribute.String("osapi.agent.machine_id", machineID),
	)
	cfg, err := c.setAgentConfig(ctx, machineID, set)
	span.SetAttributes(attribute.Int("osapi.agent.config_version", cfg.Version))

	return cfg, opspan.End(span, err)
}

// setAgentConfig writes the next version of the config of machineID.
func (c *Client) setAgentConfig(
	ctx context.Context,
	machineID string,
	set job.AgentConfig,
) (job.AgentConfig, error) {
	cfg, err := c.agentConfig(ctx, machineID)
	if err != nil {
		return job.AgentConfig{}, err
	}
	cfg.Version++
	cfg.Labels = set.Labels
	cfg.MaxJobs = set.MaxJobs
	cfg.UpdatedAt = time.Now().UTC()
	data, err := c.JSONMarshalFn(cfg)
	if err != nil {
		return job.AgentConfig{}, fmt.Errorf("marshal the config of %s: %w", machineID, err)
	}
	if _, err := c.natsClient.KVPut(
		ctx,
		c.stateBucket,
		job.AgentConfigKey(machineID),
		data,
	); err != nil {
		return job.AgentConfig{}, fmt.Errorf("write the config of %s: %w", machineID, err)
	}

	return cfg, nil
}

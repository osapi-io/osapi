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
	"sort"
	"time"

	"github.com/osapi-io/osapi/internal/telemetry/opspan"
	"go.opentelemetry.io/otel/attribute"

	"github.com/osapi-io/osapi/internal/job"
)

// WriteAgentTimelineEvent writes an append-only timeline event for an
// agent state transition, under the agent's machine ID, which a rename does
// not change. The event keeps the hostname that the agent had then.
func (c *Client) WriteAgentTimelineEvent(
	ctx context.Context,
	machineID, hostname, event, message string,
) error {
	if c.stateBucket == "" {
		return fmt.Errorf("agent state bucket not configured")
	}

	now := time.Now()
	key := fmt.Sprintf(
		"timeline.%s.%s.%d",
		job.SanitizeHostname(machineID),
		event,
		now.UnixNano(),
	)

	data, err := c.JSONMarshalFn(job.TimelineEvent{
		Timestamp: now,
		Event:     event,
		Hostname:  hostname,
		Message:   message,
	})
	if err != nil {
		return fmt.Errorf("marshal timeline event: %w", err)
	}

	_, err = c.natsClient.KVPut(ctx, c.stateBucket, key, data)
	if err != nil {
		return fmt.Errorf("write timeline event: %w", err)
	}

	c.logger.Debug(
		"wrote agent timeline event",
		slog.String("hostname", hostname),
		slog.String("event", event),
		slog.String("key", key),
	)

	return nil
}

// agentTimeline returns the sorted timeline events of the agent with
// machineID. hostname names the agent in an error.
func (c *Client) agentTimeline(
	ctx context.Context,
	machineID string,
	hostname string,
) ([]job.TimelineEvent, error) {
	if c.stateBucket == "" {
		return nil, fmt.Errorf("agent state bucket not configured")
	}

	// Ask the server for this agent's timeline keys only: the bucket holds
	// each agent's drain, config, seen and timeline keys, and the timeline
	// keys are never removed.
	filter := "timeline." + job.SanitizeHostname(machineID) + ".>"

	keys, err := c.natsClient.KVKeys(ctx, c.stateBucket, filter)
	if err != nil {
		// An empty bucket has no keys and no error, so it is an empty
		// timeline. An error is a bucket that could not be read, and
		// reporting that as "no events" tells an operator the agent has no
		// history when the truth is that nobody knows.
		return nil, fmt.Errorf("list timeline keys for %q: %w", hostname, err)
	}

	events := []job.TimelineEvent{}
	for _, key := range keys {
		entry, err := c.natsClient.KVGet(ctx, c.stateBucket, key)
		if err != nil {
			continue
		}

		var te job.TimelineEvent
		if err := json.Unmarshal(entry.Value(), &te); err != nil {
			continue
		}

		events = append(events, te)
	}

	// Sort by timestamp
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.Before(events[j].Timestamp)
	})

	return events, nil
}

// GetAgentTimeline returns the timeline of an agent, in an agent.timeline
// span.
func (c *Client) GetAgentTimeline(
	ctx context.Context,
	machineID string,
	hostname string,
) ([]job.TimelineEvent, error) {
	ctx, span := opspan.Start(ctx, tracerName, "agent.timeline",
		attribute.String("osapi.agent.machine_id", machineID),
		attribute.String("osapi.agent.hostname", hostname),
	)
	events, err := c.agentTimeline(ctx, machineID, hostname)

	return events, opspan.End(span, err)
}

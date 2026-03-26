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
	"strings"
	"time"

	"github.com/retr0h/osapi/internal/job"
)

// WriteAgentTimelineEvent writes an append-only timeline event
// for an agent state transition.
func (c *Client) WriteAgentTimelineEvent(
	ctx context.Context,
	hostname, event, message string,
) error {
	if c.stateKV == nil {
		return fmt.Errorf("agent state bucket not configured")
	}

	now := time.Now()
	key := fmt.Sprintf(
		"timeline.%s.%s.%d",
		job.SanitizeHostname(hostname),
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

	_, err = c.stateKV.Put(ctx, key, data)
	if err != nil {
		return fmt.Errorf("write timeline event: %w", err)
	}

	c.logger.Debug("wrote agent timeline event",
		slog.String("hostname", hostname),
		slog.String("event", event),
		slog.String("key", key),
	)

	return nil
}

// GetAgentTimeline returns sorted timeline events for a hostname.
func (c *Client) GetAgentTimeline(
	ctx context.Context,
	hostname string,
) ([]job.TimelineEvent, error) {
	if c.stateKV == nil {
		return nil, fmt.Errorf("agent state bucket not configured")
	}

	prefix := "timeline." + job.SanitizeHostname(hostname) + "."

	keys, err := c.stateKV.Keys(ctx)
	if err != nil {
		// No keys found is not an error for timeline
		return []job.TimelineEvent{}, nil
	}

	var events []job.TimelineEvent
	for _, key := range keys {
		if !strings.HasPrefix(key, prefix) {
			continue
		}

		entry, err := c.stateKV.Get(ctx, key)
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

// ComputeAgentState returns the current state from timeline events.
func ComputeAgentState(
	events []job.TimelineEvent,
) string {
	if len(events) == 0 {
		return job.AgentStateReady
	}

	latest := events[len(events)-1]
	switch latest.Event {
	case "drain":
		return job.AgentStateDraining
	case "cordoned":
		return job.AgentStateCordoned
	case "undrain", "ready":
		return job.AgentStateReady
	default:
		return job.AgentStateReady
	}
}

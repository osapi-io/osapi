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
package admission

import (
	"context"
	"log/slog"
)

// Timeline is where the watcher writes the lifecycle events of a machine:
// the timeline of the agent, under its machine ID. The job client is one.
type Timeline interface {
	WriteAgentTimelineEvent(ctx context.Context, machineID, hostname, event, message string) error
}

// SetTimeline gives the watcher the timeline of the agents, which each
// admission decision writes an event to. Without it, a decision is only
// logged.
func (w *Watcher) SetTimeline(
	t Timeline,
) {
	w.timeline = t
}

// event writes the lifecycle event of machineID to its timeline. A failed
// write is logged: the decision stands without it.
func (w *Watcher) event(
	ctx context.Context,
	machineID string,
	hostname string,
	event string,
	message string,
) {
	if w.timeline == nil {
		return
	}
	if err := w.timeline.WriteAgentTimelineEvent(
		context.WithoutCancel(ctx), machineID, hostname, event, message,
	); err != nil {
		w.logger.Warn("failed to write a timeline event",
			slog.String("machine_id", machineID),
			slog.String("event", event),
			slog.String("error", err.Error()),
		)
	}
}

// hostnameOf is the hostname of machineID's record, or "".
func (w *Watcher) hostnameOf(
	ctx context.Context,
	machineID string,
) string {
	r, err := w.get(ctx, machineID)
	if err != nil {
		return ""
	}

	return r.Hostname
}

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

package agent

import (
	"context"

	"github.com/retr0h/osapi/internal/job"
)

// checkDrainFlag checks the drain flag via the job client. Drain flags
// are stored in the main KV bucket (longer TTL than registry).
func (a *Agent) checkDrainFlag(
	ctx context.Context,
	hostname string,
) bool {
	return a.jobClient.CheckDrainFlag(ctx, hostname)
}

// handleDrainDetection checks drain flag on each heartbeat tick.
// When drain is requested and agent is Ready, it transitions to Draining
// and stops accepting new jobs by stopping consumer message handlers.
// When drain flag is removed and agent is Cordoned, it transitions back
// to Ready and resumes accepting jobs.
func (a *Agent) handleDrainDetection(
	ctx context.Context,
	hostname string,
) {
	drainRequested := a.checkDrainFlag(ctx, hostname)

	switch {
	case drainRequested && a.state == job.AgentStateReady:
		a.logger.Info("drain detected, stopping job consumption")
		a.stopConsumers()
		a.state = job.AgentStateCordoned
		a.logger.Info("all consumers stopped, agent cordoned")
		_ = a.jobClient.WriteAgentTimelineEvent(
			ctx, hostname, "drain", "Drain initiated",
		)
		_ = a.jobClient.WriteAgentTimelineEvent(
			ctx, hostname, "cordoned", "All jobs completed",
		)

	case !drainRequested && (a.state == job.AgentStateDraining || a.state == job.AgentStateCordoned):
		a.logger.Info("undrain detected, resuming job consumption")
		a.startConsumers()
		a.state = job.AgentStateReady
		_ = a.jobClient.WriteAgentTimelineEvent(
			ctx, hostname, "undrain", "Resumed accepting jobs",
		)
	}
}

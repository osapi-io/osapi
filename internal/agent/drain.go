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
	"log/slog"
	"sync"
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/osapi-io/osapi/internal/job"
)

// checkDrainFlag checks the drain flag via the job client. Drain flags
// are stored in the main KV bucket (longer TTL than registry).
func (a *Agent) checkDrainFlag(
	ctx context.Context,
	machineID string,
) (bool, error) {
	return a.jobClient.CheckDrainFlag(ctx, machineID)
}

// drainStuckAfter is how long a drain waits for its consumers before it
// reports the DrainStuck condition. A consumer stops once its fetch ends,
// so a healthy drain takes well under this.
var drainStuckAfter = time.Minute

// handleDrainDetection checks the drain flag on each heartbeat tick. On a
// drain of a Ready agent it cancels the consumers and moves to Draining at
// once; finishDrain waits for them in its own goroutine and moves to
// Cordoned. The heartbeat never waits for a consumer, so a slow or stuck
// consumer cannot stop the agent from heartbeating (#876). On an undrain of
// a Cordoned agent it starts the consumers and moves to Ready. An undrain
// that comes while the agent is still Draining waits for the drain to
// finish, and applies at the next tick.
func (a *Agent) handleDrainDetection(
	ctx context.Context,
	machineID string,
	hostname string,
) {
	drainRequested, err := a.checkDrainFlag(ctx, machineID)
	if err != nil {
		// The flag is unknown, not absent: keep the state until a read
		// answers, so an outage does not undrain a cordoned agent.
		a.logger.Warn(
			"drain flag not read, the state stays",
			slog.String("error", err.Error()),
		)

		return
	}

	// The transition is taken under the lock: reading the state, moving the
	// consumers, and recording the new state are one step, so an enrollment
	// acceptance arriving now waits rather than interleaving with it.
	a.lifecycleMu.Lock()
	defer a.lifecycleMu.Unlock()

	switch {
	case drainRequested && a.state == types.AgentReady:
		a.logger.Info("drain detected, stopping job consumption")
		set := a.cancelConsumersLocked()
		a.state = types.AgentDraining
		_ = a.jobClient.WriteAgentTimelineEvent(
			ctx, a.machineID, hostname, "drain", "Drain initiated",
		)
		a.wg.Add(1)
		go a.finishDrain(ctx, set, hostname)

	case !drainRequested && a.state == types.AgentCordoned:
		a.logger.Info("undrain detected, resuming job consumption")
		a.startConsumersLocked()
		a.state = types.AgentReady
		_ = a.jobClient.WriteAgentTimelineEvent(
			ctx, a.machineID, hostname, "undrain", "Resumed accepting jobs",
		)
	}
}

// finishDrain waits for the consumers of a drained set to stop, then moves
// the agent from Draining to Cordoned. When they take longer than
// drainStuckAfter it logs a warning and sets drainStuck, which the
// heartbeat reports as the DrainStuck condition, and goes on waiting.
func (a *Agent) finishDrain(
	ctx context.Context,
	set *sync.WaitGroup,
	hostname string,
) {
	defer a.wg.Done()

	if set != nil {
		done := make(chan struct{})
		go func() {
			set.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(drainStuckAfter):
			a.logger.Warn(
				"drain stuck: job consumers have not stopped",
				slog.Duration("after", drainStuckAfter),
			)
			a.lifecycleMu.Lock()
			a.drainStuck = true
			a.lifecycleMu.Unlock()
			<-done
		}
	}

	a.lifecycleMu.Lock()
	a.drainStuck = false
	cordoned := a.state == types.AgentDraining
	if cordoned {
		a.state = types.AgentCordoned
	}
	a.lifecycleMu.Unlock()

	if cordoned {
		a.logger.Info("all consumers stopped, agent cordoned")
		_ = a.jobClient.WriteAgentTimelineEvent(
			ctx, a.machineID, hostname, "cordoned", "All jobs completed",
		)
	}
}

// drainCondition is the DrainStuck condition of the agent, for the
// heartbeat.
func (a *Agent) drainCondition() job.Condition {
	a.lifecycleMu.RLock()
	stuck := a.drainStuck
	a.lifecycleMu.RUnlock()

	c := job.Condition{
		Type:               types.ConditionDrainStuck,
		Status:             stuck,
		LastTransitionTime: transitionTime(types.ConditionDrainStuck, stuck, a.prevConditions),
	}
	if stuck {
		c.Reason = "job consumers have not stopped since the drain"
	}

	return c
}

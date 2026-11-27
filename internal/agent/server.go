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

package agent

import (
	"context"
	"log/slog"
	"time"

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/osapi-io/osapi/internal/agent/identity"
)

// getIdentityFn dispatches to the identity resolution function.
// Overridable in tests via export_test.go.
var getIdentityFn = identity.GetIdentity

// Start starts the agent without blocking. Call Stop to shut down.
func (a *Agent) Start() {
	a.ctx, a.cancel = context.WithCancel(context.Background())
	a.startedAt = time.Now()
	a.setState(types.AgentReady)

	a.logger.Info("starting node agent")

	// Resolve agent identity (machine ID + hostname).
	ident, err := getIdentityFn(a.appFs, a.appConfig.Agent.Hostname)
	if err != nil {
		a.logger.Error(
			"failed to resolve agent identity",
			slog.String("error", err.Error()),
		)
		a.cancel()

		return
	}
	a.machineID = ident.MachineID
	a.hostname = ident.Hostname

	a.logger.Info(
		"agent configuration",
		slog.String("machine_id", a.machineID),
		slog.String("hostname", a.hostname),
		slog.Int("max_jobs", a.appConfig.Agent.MaxJobs),
		slog.Any("labels", a.appConfig.Agent.Labels),
	)

	// PKI enrollment: the agent is always admitted before it takes a job.
	if err := a.handlePKIEnrollment(a.ctx); err != nil {
		a.logger.Error(
			"PKI enrollment failed",
			slog.String("error", err.Error()),
		)
		a.cancel()

		return
	}

	a.refreshPrograms()

	// A ping needs the agent's group in net.ipv4.ping_group_range, whatever
	// the privileges of the agent, so this runs on each start.
	if problem := PingProblem(); problem != "" {
		a.logger.Warn("a ping from this agent fails", slog.String("error", problem))
	}

	// Run preflight checks when privilege escalation is enabled.
	pe := a.appConfig.Agent.PrivilegeEscalation
	if pe.Enabled {
		results, ok := RunPreflight(a.ctx, a.logger, a.execManager, detectFn())
		for _, r := range results {
			if r.Passed {
				continue
			}
			level := slog.LevelWarn
			if r.Fatal {
				level = slog.LevelError
			}
			a.logger.Log(
				a.ctx,
				level,
				"preflight check failed",
				slog.String("check", r.Name),
				slog.String("error", r.Error),
			)
		}

		if !ok {
			a.logger.Error("preflight failed, agent cannot start")
			a.cancel()

			return
		}

		a.logger.Info("preflight checks done")
	}

	// Start consuming messages only when not pending enrollment, and
	// before the first heartbeat: the consumers deliver only the jobs sent
	// after they exist, so an agent that shows Ready must have them
	// (osapi-io/osapi#1560). Pending agents are visible (heartbeat) but
	// don't process jobs.
	pending := a.State() == types.AgentPending
	if !pending {
		a.startConsumers()
	}

	// Register in agent registry and start heartbeat keepalive.
	// Heartbeats run even in Pending state so the agent is visible.
	a.startHeartbeat(a.ctx, a.machineID, a.hostname)
	a.lifecycle(a.ctx, "started", "started, version "+a.versionOrUnknown())

	// In operator mode, renew the JWT of the agent's NATS user.
	a.startRenewal(a.ctx)

	// Collect and publish system facts.
	a.startFacts(a.ctx, a.machineID, a.hostname)

	if pending {
		a.logger.Info(
			"waiting for admission; run osapi pki key get --name agent on this machine, and give its fingerprint to osapi client admission admit",
			slog.String("machine_id", a.machineID),
			slog.String("fingerprint", a.pkiManager.Fingerprint()),
		)
	}

	a.logger.Info("node agent started successfully")
}

// Stop gracefully shuts down the agent, waiting for in-flight jobs
// to finish or the context deadline to expire.
func (a *Agent) Stop(
	ctx context.Context,
) {
	a.logger.Info("node agent shutting down")
	if a.machineID != "" {
		a.lifecycle(ctx, "stopped", "stopped")
	}
	a.cancel()

	done := make(chan struct{})
	go func() {
		a.consumerWg.Wait()
		a.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		a.logger.Info("node agent stopped gracefully")
	case <-ctx.Done():
		a.logger.Warn("node agent shutdown timed out")
	}
}

// lifecycle writes an event of the agent's own lifecycle to its timeline. A
// failed write is logged: the agent runs without it.
func (a *Agent) lifecycle(
	ctx context.Context,
	event string,
	message string,
) {
	if err := a.jobClient.WriteAgentTimelineEvent(
		context.WithoutCancel(ctx), a.machineID, a.hostname, event, message,
	); err != nil {
		a.logger.Warn("failed to write a timeline event",
			slog.String("event", event),
			slog.String("error", err.Error()),
		)
	}
}

// versionOrUnknown is the agent's version, or "unknown" when SetVersion did
// not set one.
func (a *Agent) versionOrUnknown() string {
	if a.version == "" {
		return "unknown"
	}

	return a.version
}

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

	"github.com/retr0h/osapi/internal/agent/identity"
	"github.com/retr0h/osapi/internal/job"
)

// getIdentityFn dispatches to the identity resolution function.
// Overridable in tests via export_test.go.
var getIdentityFn = identity.GetIdentity

// Start starts the agent without blocking. Call Stop to shut down.
func (a *Agent) Start() {
	a.ctx, a.cancel = context.WithCancel(context.Background())
	a.startedAt = time.Now()
	a.state = job.AgentStateReady

	a.logger.Info("starting node agent")

	// Resolve agent identity (machine ID + hostname).
	ident, err := getIdentityFn(a.appFs, a.appConfig.Agent.Hostname)
	if err != nil {
		a.logger.Error("failed to resolve agent identity",
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
		slog.String("queue_group", a.appConfig.Agent.QueueGroup),
		slog.Int("max_jobs", a.appConfig.Agent.MaxJobs),
		slog.Any("labels", a.appConfig.Agent.Labels),
	)

	// PKI enrollment (when enabled).
	if a.appConfig.Agent.PKI.Enabled {
		if err := a.handlePKIEnrollment(a.ctx); err != nil {
			a.logger.Error("PKI enrollment failed",
				slog.String("error", err.Error()),
			)
			a.cancel()

			return
		}
	}

	// Run preflight checks when privilege escalation is enabled.
	pe := a.appConfig.Agent.PrivilegeEscalation
	if pe.Enabled {
		results, ok := RunPreflight(a.logger, a.execManager)
		if !ok {
			for _, r := range results {
				if !r.Passed {
					a.logger.Error("preflight check failed",
						slog.String("check", r.Name),
						slog.String("error", r.Error),
					)
				}
			}

			a.logger.Error("preflight failed, agent cannot start")
			a.cancel()

			return
		}

		a.logger.Info("preflight checks passed")
	}

	// Register in agent registry and start heartbeat keepalive.
	a.startHeartbeat(a.ctx, a.machineID, a.hostname)

	// Collect and publish system facts.
	a.startFacts(a.ctx, a.machineID, a.hostname)

	// Start consuming messages for different job types.
	a.startConsumers()

	a.logger.Info("node agent started successfully")
}

// Stop gracefully shuts down the agent, waiting for in-flight jobs
// to finish or the context deadline to expire.
func (a *Agent) Stop(
	ctx context.Context,
) {
	a.logger.Info("node agent shutting down")
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

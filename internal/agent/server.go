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

	"github.com/retr0h/osapi/internal/job"
)

// Start starts the agent without blocking. Call Stop to shut down.
func (a *Agent) Start() {
	a.ctx, a.cancel = context.WithCancel(context.Background())
	a.startedAt = time.Now()

	a.logger.Info("starting node agent")

	// Determine agent hostname (GetAgentHostname always succeeds)
	hostname, _ := job.GetAgentHostname(a.appConfig.Agent.Hostname)

	a.logger.Info(
		"agent configuration",
		slog.String("hostname", hostname),
		slog.String("queue_group", a.appConfig.Agent.QueueGroup),
		slog.Int("max_jobs", a.appConfig.Agent.MaxJobs),
		slog.Any("labels", a.appConfig.Agent.Labels),
	)

	// Register in agent registry and start heartbeat keepalive.
	a.startHeartbeat(a.ctx, hostname)

	// Start consuming messages for different job types.
	// Each consume function spawns goroutines tracked by a.wg.
	_ = a.consumeQueryJobs(a.ctx, hostname)
	_ = a.consumeModifyJobs(a.ctx, hostname)

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

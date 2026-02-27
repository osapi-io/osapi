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
	"encoding/json"
	"log/slog"
	"time"

	"github.com/retr0h/osapi/internal/job"
)

// heartbeatInterval is the interval between heartbeat refreshes.
var heartbeatInterval = 10 * time.Second

// marshalJSON is a package-level variable for testing the marshal error path.
var marshalJSON = json.Marshal

// startHeartbeat writes the initial registration, spawns a goroutine that
// refreshes the entry on a ticker, and deregisters on ctx.Done().
func (a *Agent) startHeartbeat(
	ctx context.Context,
	hostname string,
) {
	if a.registryKV == nil {
		return
	}

	ttl := a.appConfig.NATS.Registry.TTL
	key := registryKey(hostname)

	a.writeRegistration(ctx, hostname)

	a.logger.Info(
		"registered in agent registry",
		slog.String("hostname", hostname),
		slog.String("key", key),
		slog.String("ttl", ttl),
	)

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()

		ticker := time.NewTicker(heartbeatInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				a.deregister(hostname)
				return
			case <-ticker.C:
				a.writeRegistration(ctx, hostname)

				a.logger.Info(
					"heartbeat refreshed",
					slog.String("hostname", hostname),
					slog.String("key", key),
					slog.String("next_in", heartbeatInterval.String()),
				)
			}
		}
	}()
}

// writeRegistration marshals an agent registration and puts it to the registry KV.
func (a *Agent) writeRegistration(
	ctx context.Context,
	hostname string,
) {
	reg := job.AgentRegistration{
		Hostname:     hostname,
		Labels:       a.appConfig.Node.Agent.Labels,
		RegisteredAt: time.Now(),
	}

	data, err := marshalJSON(reg)
	if err != nil {
		a.logger.Warn(
			"failed to marshal agent registration",
			slog.String("hostname", hostname),
			slog.String("error", err.Error()),
		)
		return
	}

	key := registryKey(hostname)
	if _, err := a.registryKV.Put(ctx, key, data); err != nil {
		a.logger.Warn(
			"failed to write agent registration",
			slog.String("hostname", hostname),
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
	}
}

// deregister deletes the agent's registration key on clean shutdown.
func (a *Agent) deregister(
	hostname string,
) {
	key := registryKey(hostname)
	if err := a.registryKV.Delete(context.Background(), key); err != nil {
		a.logger.Warn(
			"failed to deregister agent",
			slog.String("hostname", hostname),
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		return
	}

	a.logger.Info(
		"agent deregistered",
		slog.String("hostname", hostname),
		slog.String("key", key),
	)
}

// registryKey returns the KV key for an agent's registration entry.
func registryKey(
	hostname string,
) string {
	return "workers." + job.SanitizeHostname(hostname)
}

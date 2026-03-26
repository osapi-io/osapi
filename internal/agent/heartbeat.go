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
	"github.com/retr0h/osapi/internal/provider/node/disk"
	"github.com/retr0h/osapi/internal/telemetry/process"
	"github.com/retr0h/osapi/internal/provider/node/load"
	"github.com/retr0h/osapi/internal/provider/node/mem"
)

// heartbeatInterval is the interval between heartbeat refreshes.
var heartbeatInterval = 10 * time.Second

// marshalJSON is a package-level variable for testing the marshal error path.
var marshalJSON = json.Marshal

// unmarshalJSON is a package-level variable for testing the unmarshal error path.
var unmarshalJSON = json.Unmarshal

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

	a.heartbeatLogger.Info(
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

				a.heartbeatLogger.Info(
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
// Provider errors are non-fatal; the heartbeat still writes with whatever data it gathered.
func (a *Agent) writeRegistration(
	ctx context.Context,
	hostname string,
) {
	a.handleDrainDetection(ctx, hostname)

	reg := job.AgentRegistration{
		Hostname:      hostname,
		Labels:        a.appConfig.Agent.Labels,
		RegisteredAt:  time.Now(),
		StartedAt:     a.startedAt,
		State:         a.state,
		SubComponents: a.subComponents,
	}

	if info, err := a.hostProvider.GetOSInfo(); err == nil {
		reg.OSInfo = info
	}

	if uptime, err := a.hostProvider.GetUptime(); err == nil {
		reg.Uptime = uptime
	}

	var loadAvg *load.Result
	if avg, err := a.loadProvider.GetAverageStats(); err == nil {
		loadAvg = avg
		reg.LoadAverages = avg
	}

	var memStats *mem.Result
	if stats, err := a.memProvider.GetStats(); err == nil {
		memStats = stats
		reg.MemoryStats = stats
	}

	var diskStats []disk.Result
	if stats, err := a.diskProvider.GetLocalUsageStats(); err == nil {
		diskStats = stats
	}

	conditions := []job.Condition{
		evaluateMemoryPressure(
			memStats,
			a.appConfig.Agent.Conditions.MemoryPressureThreshold,
			a.prevConditions,
		),
		evaluateHighLoad(
			loadAvg,
			a.cpuCount,
			a.appConfig.Agent.Conditions.HighLoadMultiplier,
			a.prevConditions,
		),
		evaluateDiskPressure(
			diskStats,
			a.appConfig.Agent.Conditions.DiskPressureThreshold,
			a.prevConditions,
		),
	}
	a.prevConditions = conditions

	if a.processProvider != nil {
		if pm, err := a.processProvider.GetMetrics(); err == nil {
			reg.Process = &job.ProcessMetrics{
				CPUPercent: pm.CPUPercent,
				RSSBytes:   pm.RSSBytes,
				Goroutines: pm.Goroutines,
			}

			processConditions := process.EvaluateProcessConditions(
				pm,
				process.ConditionThresholds{
					MemoryPressureBytes: a.appConfig.Agent.ProcessConditions.MemoryPressureBytes,
					HighCPUPercent:      a.appConfig.Agent.ProcessConditions.HighCPUPercent,
				},
				a.prevProcessConditions,
			)
			a.prevProcessConditions = processConditions
			conditions = append(conditions, processConditions...)
		}
	}

	reg.Conditions = conditions

	data, err := marshalJSON(reg)
	if err != nil {
		a.heartbeatLogger.Warn(
			"failed to marshal agent registration",
			slog.String("hostname", hostname),
			slog.String("error", err.Error()),
		)
		return
	}

	key := registryKey(hostname)
	if _, err := a.registryKV.Put(ctx, key, data); err != nil {
		a.heartbeatLogger.Warn(
			"failed to write agent registration",
			slog.String("hostname", hostname),
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
	} else {
		a.lastHeartbeatTime.Store(time.Now())
	}
}

// deregister deletes the agent's registration key on clean shutdown.
func (a *Agent) deregister(
	hostname string,
) {
	key := registryKey(hostname)
	if err := a.registryKV.Delete(context.Background(), key); err != nil {
		a.heartbeatLogger.Warn(
			"failed to deregister agent",
			slog.String("hostname", hostname),
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		return
	}

	a.heartbeatLogger.Info(
		"agent deregistered",
		slog.String("hostname", hostname),
		slog.String("key", key),
	)
}

// registryKey returns the KV key for an agent's registration entry.
func registryKey(
	hostname string,
) string {
	return "agents." + job.SanitizeHostname(hostname)
}

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

package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/retr0h/osapi/internal/job"
	"github.com/retr0h/osapi/internal/provider/process"
)

// heartbeatMarshalFn is the JSON marshal function (injectable for testing).
var heartbeatMarshalFn = json.Marshal

// ComponentHeartbeat writes ComponentRegistration to the registry KV
// on a configurable interval.
type ComponentHeartbeat struct {
	logger          *slog.Logger
	registryKV      jetstream.KeyValue
	hostname        string
	version         string
	componentType   string
	processProvider process.Provider
	interval        time.Duration
	startedAt       time.Time
	thresholds      process.ConditionThresholds
	prevConditions  []job.Condition
}

// NewComponentHeartbeat creates a heartbeat writer for a non-agent component.
func NewComponentHeartbeat(
	logger *slog.Logger,
	registryKV jetstream.KeyValue,
	hostname string,
	version string,
	componentType string,
	processProvider process.Provider,
	interval time.Duration,
	thresholds process.ConditionThresholds,
) *ComponentHeartbeat {
	return &ComponentHeartbeat{
		logger:          logger,
		registryKV:      registryKV,
		hostname:        hostname,
		version:         version,
		componentType:   componentType,
		processProvider: processProvider,
		interval:        interval,
		startedAt:       time.Now(),
		thresholds:      thresholds,
	}
}

// Start begins writing heartbeats. It writes an initial registration
// immediately, then refreshes on each tick. On context cancellation it
// deregisters and returns.
func (h *ComponentHeartbeat) Start(
	ctx context.Context,
) {
	key := h.registryKey()

	h.writeRegistration(ctx)

	h.logger.Info(
		"registered in component registry",
		slog.String("hostname", h.hostname),
		slog.String("type", h.componentType),
		slog.String("key", key),
	)

	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.deregister()
			return
		case <-ticker.C:
			h.writeRegistration(ctx)

			h.logger.Info(
				"component heartbeat refreshed",
				slog.String("hostname", h.hostname),
				slog.String("type", h.componentType),
				slog.String("key", key),
				slog.String("next_in", h.interval.String()),
			)
		}
	}
}

// writeRegistration marshals a ComponentRegistration and puts it to the
// registry KV. Provider errors are non-fatal; the heartbeat still writes
// with whatever data it gathered.
func (h *ComponentHeartbeat) writeRegistration(
	ctx context.Context,
) {
	reg := job.ComponentRegistration{
		Type:         h.componentType,
		Hostname:     h.hostname,
		RegisteredAt: time.Now(),
		StartedAt:    h.startedAt,
		Version:      h.version,
	}

	if pm, err := h.processProvider.GetMetrics(); err == nil {
		reg.Process = &job.ProcessMetrics{
			CPUPercent: pm.CPUPercent,
			RSSBytes:   pm.RSSBytes,
			Goroutines: pm.Goroutines,
		}

		conditions := process.EvaluateProcessConditions(pm, h.thresholds, h.prevConditions)
		h.prevConditions = conditions
		reg.Conditions = conditions
	}

	data, err := heartbeatMarshalFn(reg)
	if err != nil {
		h.logger.Warn(
			"failed to marshal component registration",
			slog.String("hostname", h.hostname),
			slog.String("type", h.componentType),
			slog.String("error", err.Error()),
		)
		return
	}

	key := h.registryKey()
	if _, err := h.registryKV.Put(ctx, key, data); err != nil {
		h.logger.Warn(
			"failed to write component registration",
			slog.String("hostname", h.hostname),
			slog.String("type", h.componentType),
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
	}
}

// deregister deletes the component's registration key on clean shutdown.
func (h *ComponentHeartbeat) deregister() {
	key := h.registryKey()
	if err := h.registryKV.Delete(context.Background(), key); err != nil {
		h.logger.Warn(
			"failed to deregister component",
			slog.String("hostname", h.hostname),
			slog.String("type", h.componentType),
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		return
	}

	h.logger.Info(
		"component deregistered",
		slog.String("hostname", h.hostname),
		slog.String("type", h.componentType),
		slog.String("key", key),
	)
}

// registryKey returns the KV key for this component's registration entry.
func (h *ComponentHeartbeat) registryKey() string {
	return h.componentType + "." + job.SanitizeHostname(h.hostname)
}

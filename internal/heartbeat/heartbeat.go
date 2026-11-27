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

// Package heartbeat writes the registration of a component that is not an agent,
// for example the controller or the NATS server, to the registry KV bucket.
package heartbeat

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/osapi-io/osapi/internal/cryptoreport"
	"github.com/osapi-io/osapi/internal/job"
	"github.com/osapi-io/osapi/internal/telemetry/process"
)

// heartbeatMarshalFn is the JSON marshal function (injectable for testing).
var heartbeatMarshalFn = json.Marshal

// Heartbeat writes ComponentRegistration to the registry KV
// on a configurable interval.
type Heartbeat struct {
	logger          *slog.Logger
	kv              RegistryStore
	registryBucket  string
	hostname        string
	version         string
	componentType   string
	processProvider process.Provider
	interval        time.Duration
	startedAt       time.Time
	thresholds      process.ConditionThresholds
	prevConditions  []job.Condition
	subComponents   map[string]job.SubComponentInfo
	crypto          *cryptoreport.Report
}

// RegistryStore writes and removes a registration in a KV bucket. A
// *natsclient.Client is one.
type RegistryStore interface {
	KVPut(ctx context.Context, bucket string, key string, value []byte) (uint64, error)
	KVDelete(ctx context.Context, bucket string, key string) error
}

// Options describe the component that a Heartbeat registers.
type Options struct {
	// KV writes the registration.
	KV RegistryStore
	// RegistryBucket is the KV bucket that gets the registration.
	RegistryBucket string
	// Hostname is the host that the component runs on.
	Hostname string
	// Version is the version of the component.
	Version string
	// Component is the kind of component, for example controller or nats.
	Component string
	// Process reads the CPU and the memory of the process.
	Process process.Provider
	// Interval is the time between two registrations.
	Interval time.Duration
	// Thresholds tell when the process reports a condition.
	Thresholds process.ConditionThresholds
	// SubComponents are the parts of the component that the registration names.
	SubComponents map[string]job.SubComponentInfo
	// Crypto is the crypto of the component's process.
	Crypto *cryptoreport.Report
}

// New creates a heartbeat writer for a component that is not an agent.
func New(
	logger *slog.Logger,
	opts Options,
) *Heartbeat {
	return &Heartbeat{
		logger:          logger.With(slog.String("subsystem", opts.Component+".heartbeat")),
		kv:              opts.KV,
		registryBucket:  opts.RegistryBucket,
		hostname:        opts.Hostname,
		version:         opts.Version,
		componentType:   opts.Component,
		processProvider: opts.Process,
		interval:        opts.Interval,
		startedAt:       time.Now(),
		thresholds:      opts.Thresholds,
		subComponents:   opts.SubComponents,
		crypto:          opts.Crypto,
	}
}

// Start begins writing heartbeats. It writes an initial registration
// immediately, then refreshes on each tick. On context cancellation it
// deregisters and returns.
func (h *Heartbeat) Start(
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
func (h *Heartbeat) writeRegistration(
	ctx context.Context,
) {
	reg := job.ComponentRegistration{
		Type:          h.componentType,
		Hostname:      h.hostname,
		RegisteredAt:  time.Now(),
		StartedAt:     h.startedAt,
		Version:       h.version,
		SubComponents: h.subComponents,
		Crypto:        h.crypto,
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
	if _, err := h.kv.KVPut(ctx, h.registryBucket, key, data); err != nil {
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
func (h *Heartbeat) deregister() {
	key := h.registryKey()
	if err := h.kv.KVDelete(context.Background(), h.registryBucket, key); err != nil {
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
func (h *Heartbeat) registryKey() string {
	return h.componentType + "." + job.SanitizeHostname(h.hostname)
}

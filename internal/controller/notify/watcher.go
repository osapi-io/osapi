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

package notify

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/retr0h/osapi/internal/job"
)

// conditionState tracks the state of a single condition.
type conditionState struct {
	active       bool
	lastNotified time.Time
}

// Watcher monitors the registry KV bucket for condition transitions and
// emits ConditionEvents via the configured Notifier. Active conditions
// are re-notified after the configured renotify interval.
type Watcher struct {
	kv               jetstream.KeyValue
	notifier         Notifier
	logger           *slog.Logger
	prev             map[string]map[string]*conditionState // key → condition → state
	renotifyInterval time.Duration
	mu               sync.Mutex
}

// NewWatcher creates a new Watcher that monitors registry KV entries and
// notifies on condition transitions. Active conditions are re-notified
// after renotifyInterval elapses (0 disables re-notification).
func NewWatcher(
	kv jetstream.KeyValue,
	notifier Notifier,
	logger *slog.Logger,
	renotifyInterval time.Duration,
) *Watcher {
	return &Watcher{
		kv:               kv,
		notifier:         notifier,
		logger:           logger.With(slog.String("subsystem", "controller.notify")),
		prev:             make(map[string]map[string]*conditionState),
		renotifyInterval: renotifyInterval,
	}
}

// Start watches the registry KV for changes. Blocks until ctx is cancelled.
// Returns nil when the context is cancelled, or an error if WatchAll fails.
func (w *Watcher) Start(
	ctx context.Context,
) error {
	w.logger.Info("condition watcher started")

	watcher, err := w.kv.WatchAll(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if stopErr := watcher.Stop(); stopErr != nil {
			w.logger.Warn(
				"failed to stop KV watcher",
				slog.String("error", stopErr.Error()),
			)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case entry, ok := <-watcher.Updates():
			if !ok {
				return nil
			}
			if entry == nil {
				// nil entry signals end of initial values
				continue
			}
			w.handleEntry(ctx, entry)
		}
	}
}

// handleEntry processes a single KV entry update.
func (w *Watcher) handleEntry(
	ctx context.Context,
	entry jetstream.KeyValueEntry,
) {
	key := entry.Key()
	componentType, identifier, ok := parseRegistryKey(key)
	if !ok {
		return
	}

	if entry.Operation() == jetstream.KeyValueDelete ||
		entry.Operation() == jetstream.KeyValuePurge {
		w.handleDelete(ctx, key, componentType, identifier)
		return
	}

	// For agents, resolve the display name from the registration value.
	// The identifier is the machine ID, but we want the hostname for display.
	displayName := resolveDisplayName(componentType, identifier, entry.Value())

	conditions, ok := w.extractConditions(key, componentType, entry.Value())
	if !ok {
		return
	}

	w.detectTransitions(key, componentType, displayName, conditions)
}

// resolveDisplayName returns the hostname from an agent registration value,
// or the identifier as-is for non-agent components.
func resolveDisplayName(
	componentType string,
	identifier string,
	value []byte,
) string {
	if componentType != "agent" {
		return identifier
	}

	var reg struct {
		Hostname string `json:"hostname"`
	}
	if err := json.Unmarshal(value, &reg); err == nil && reg.Hostname != "" {
		return reg.Hostname
	}

	return identifier
}

// handleDelete emits a ComponentUnreachable event when a key expires or is
// deleted, and cleans up internal state.
func (w *Watcher) handleDelete(
	ctx context.Context,
	key string,
	componentType string,
	hostname string,
) {
	_ = w.notifier.Notify(ctx, ConditionEvent{
		ComponentType: componentType,
		Hostname:      hostname,
		Condition:     "ComponentUnreachable",
		Active:        true,
		Reason:        "heartbeat expired",
		Timestamp:     time.Now(),
	})

	w.mu.Lock()
	delete(w.prev, key)
	w.mu.Unlock()
}

// extractConditions parses a KV value into a slice of conditions based on
// the component type prefix.
func (w *Watcher) extractConditions(
	key string,
	componentType string,
	value []byte,
) ([]job.Condition, bool) {
	switch componentType {
	case "agent":
		var reg job.AgentRegistration
		if err := json.Unmarshal(value, &reg); err != nil {
			w.logger.Warn(
				"failed to unmarshal agent registration",
				slog.String("key", key),
				slog.String("error", err.Error()),
			)
			return nil, false
		}
		return reg.Conditions, true

	case "api", "nats":
		var reg job.ComponentRegistration
		if err := json.Unmarshal(value, &reg); err != nil {
			w.logger.Warn(
				"failed to unmarshal component registration",
				slog.String("key", key),
				slog.String("error", err.Error()),
			)
			return nil, false
		}
		return reg.Conditions, true

	default:
		return nil, false
	}
}

// detectTransitions compares the current conditions against the previously
// seen conditions for a key and calls Notify for any transitions.
func (w *Watcher) detectTransitions(
	key string,
	componentType string,
	hostname string,
	conditions []job.Condition,
) {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()

	current := make(map[string]bool)
	for _, c := range conditions {
		if c.Status {
			current[c.Type] = true
		}
	}

	prev := w.prev[key]
	if prev == nil {
		prev = make(map[string]*conditionState)
	}

	// New or re-notifiable conditions.
	for condType := range current {
		state := prev[condType]

		if state == nil {
			// New condition — fire.
			_ = w.notifier.Notify(context.Background(), ConditionEvent{
				ComponentType: componentType,
				Hostname:      hostname,
				Condition:     condType,
				Active:        true,
				Timestamp:     now,
			})
			prev[condType] = &conditionState{
				active:       true,
				lastNotified: now,
			}
		} else if state.active && w.renotifyInterval > 0 &&
			now.Sub(state.lastNotified) >= w.renotifyInterval {
			// Still active and renotify interval elapsed — re-fire.
			_ = w.notifier.Notify(context.Background(), ConditionEvent{
				ComponentType: componentType,
				Hostname:      hostname,
				Condition:     condType,
				Active:        true,
				Timestamp:     now,
			})
			state.lastNotified = now
		}
	}

	// Resolved conditions (in prev, not in current).
	for condType, state := range prev {
		if !current[condType] && state.active {
			_ = w.notifier.Notify(context.Background(), ConditionEvent{
				ComponentType: componentType,
				Hostname:      hostname,
				Condition:     condType,
				Active:        false,
				Timestamp:     now,
			})

			delete(prev, condType)
		}
	}

	w.prev[key] = prev
}

// parseRegistryKey splits a registry key like "agents.web-01" into its
// component type and hostname. Returns ok=false for unrecognized prefixes.
func parseRegistryKey(
	key string,
) (componentType string, identifier string, ok bool) {
	parts := strings.SplitN(key, ".", 2)
	if len(parts) != 2 {
		return "", "", false
	}

	prefix := parts[0]
	identifier = parts[1]

	switch prefix {
	case "agents":
		// For agents, the identifier is the machine ID (not hostname).
		// The hostname is extracted from the registration value by the caller.
		return "agent", identifier, true
	case "api":
		return "api", identifier, true
	case "controller":
		return "controller", identifier, true
	case "nats":
		return "nats", identifier, true
	default:
		return "", "", false
	}
}

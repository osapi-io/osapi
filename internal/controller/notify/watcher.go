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

	"github.com/osapi-io/osapi/pkg/sdk/types"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/osapi-io/osapi/internal/job"
)

// conditionState tracks the state of a single condition.
type conditionState struct {
	active       bool
	lastNotified time.Time
}

// Registry watches the registry KV bucket. A *natsclient.Client is one.
type Registry interface {
	KVWatch(
		ctx context.Context,
		bucket string,
		pattern string,
		opts ...jetstream.WatchOpt,
	) (<-chan jetstream.KeyValueEntry, error)
}

// Watcher monitors the registry KV bucket for condition transitions and
// emits ConditionEvents via the configured Notifier. Active conditions
// are re-notified after the configured renotify interval.
type Watcher struct {
	kv       Registry
	bucket   string
	notifier Notifier
	logger   *slog.Logger
	prev     map[string]map[types.ConditionType]*conditionState // key → condition → state
	// names is the display name of each key, so a delete, which carries
	// no value, names the hostname of an agent and not its machine ID.
	names map[string]string
	// unreachable is each key whose heartbeat expired, so its next
	// heartbeat resolves the ComponentUnreachable that the expiry sent.
	unreachable map[string]bool
	// seen is when each key of names and unreachable was last written. A
	// key unseen for staleAfter is forgotten, so a fleet that rebuilds its
	// machines with new IDs does not grow the maps for ever.
	seen             map[string]time.Time
	renotifyInterval time.Duration
	// rewatchDelay is how long Start waits before it watches the registry
	// again after a watch ended.
	rewatchDelay time.Duration
	mu           sync.Mutex
}

// defaultRewatchDelay is how long Start waits to watch again after the
// registry watch ended, such as on a NATS reconnect.
const defaultRewatchDelay = time.Second

// staleAfter is how long a key with no heartbeat and no delete is kept in
// names and unreachable. A component that comes back after it gets no
// "heartbeat resumed" event.
const staleAfter = 24 * time.Hour

// nowFn is time.Now, for a test to replace.
var nowFn = time.Now

// NewWatcher creates a new Watcher that monitors registry KV entries and
// notifies on condition transitions. Active conditions are re-notified
// after renotifyInterval elapses (0 disables re-notification).
func NewWatcher(
	kv Registry,
	bucket string,
	notifier Notifier,
	logger *slog.Logger,
	renotifyInterval time.Duration,
) *Watcher {
	return &Watcher{
		kv:               kv,
		bucket:           bucket,
		notifier:         notifier,
		logger:           logger.With(slog.String("subsystem", "controller.notify")),
		prev:             make(map[string]map[types.ConditionType]*conditionState),
		names:            make(map[string]string),
		unreachable:      make(map[string]bool),
		seen:             make(map[string]time.Time),
		renotifyInterval: renotifyInterval,
		rewatchDelay:     defaultRewatchDelay,
	}
}

// Start watches the registry KV for changes. Blocks until ctx is cancelled.
// Returns nil when the context is cancelled, or an error if the first watch
// fails. A watch that ends before ctx, such as on a NATS reconnect, is
// logged and started again, so the notifications do not stop with no word.
// The state of each condition stays, so the values that a new watch sends
// first fire only what changed.
func (w *Watcher) Start(
	ctx context.Context,
) error {
	w.logger.Info("condition watcher started")

	updates, err := w.kv.KVWatch(ctx, w.bucket, ">")
	if err != nil {
		return err
	}

	for {
		w.consume(ctx, updates)
		if ctx.Err() != nil {
			return nil
		}
		w.logger.Warn("the registry watch ended, watching it again")

		updates = nil
		for updates == nil {
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(w.rewatchDelay):
			}
			updates, err = w.kv.KVWatch(ctx, w.bucket, ">")
			if err != nil {
				w.logger.Warn("watch the registry again",
					slog.String("error", err.Error()))
				updates = nil
			}
		}
	}
}

// consume handles each entry of updates until it closes or ctx ends.
func (w *Watcher) consume(
	ctx context.Context,
	updates <-chan jetstream.KeyValueEntry,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case entry, ok := <-updates:
			if !ok {
				return
			}
			w.handleEntry(ctx, entry)
		}
	}
}

// notify sends event, and logs a notifier that fails, so a lost
// notification is in the log.
func (w *Watcher) notify(
	ctx context.Context,
	event ConditionEvent,
) {
	if err := w.notifier.Notify(ctx, event); err != nil {
		w.logger.Warn("send a condition notification",
			slog.String("component_type", event.ComponentType),
			slog.String("hostname", event.Hostname),
			slog.String("condition", string(event.Condition)),
			slog.Bool("active", event.Active),
			slog.String("error", err.Error()))
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
	w.resolveUnreachable(ctx, key, componentType, displayName)

	conditions, ok := w.extractConditions(key, componentType, entry.Value())
	if !ok {
		return
	}

	w.detectTransitions(ctx, key, componentType, displayName, conditions)
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
// deleted, and cleans up internal state. The event names the component as
// its last heartbeat did, because a delete carries no value.
func (w *Watcher) handleDelete(
	ctx context.Context,
	key string,
	componentType string,
	identifier string,
) {
	w.mu.Lock()
	name, ok := w.names[key]
	if !ok {
		name = identifier
	}
	delete(w.prev, key)
	w.unreachable[key] = true
	w.touchLocked(key)
	w.mu.Unlock()

	w.notify(ctx, ConditionEvent{
		ComponentType: componentType,
		Hostname:      name,
		Condition:     ConditionComponentUnreachable,
		Active:        true,
		Reason:        "heartbeat expired",
		Timestamp:     time.Now(),
	})
}

// resolveUnreachable records name as the display name of key, and resolves
// the ComponentUnreachable of a key whose heartbeat expired and came back.
func (w *Watcher) resolveUnreachable(
	ctx context.Context,
	key string,
	componentType string,
	name string,
) {
	w.mu.Lock()
	w.names[key] = name
	was := w.unreachable[key]
	delete(w.unreachable, key)
	w.touchLocked(key)
	w.mu.Unlock()

	if was {
		w.notify(ctx, ConditionEvent{
			ComponentType: componentType,
			Hostname:      name,
			Condition:     ConditionComponentUnreachable,
			Active:        false,
			Reason:        "heartbeat resumed",
			Timestamp:     time.Now(),
		})
	}
}

// touchLocked records that key was written now, and forgets each key that
// was not written for staleAfter. The caller holds mu.
func (w *Watcher) touchLocked(
	key string,
) {
	now := nowFn()
	w.seen[key] = now
	for k, at := range w.seen {
		if now.Sub(at) > staleAfter {
			delete(w.seen, k)
			delete(w.names, k)
			delete(w.unreachable, k)
		}
	}
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
	ctx context.Context,
	key string,
	componentType string,
	hostname string,
	conditions []job.Condition,
) {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()

	current := make(map[types.ConditionType]bool)
	for _, c := range conditions {
		if c.Status {
			current[c.Type] = true
		}
	}

	prev := w.prev[key]
	if prev == nil {
		prev = make(map[types.ConditionType]*conditionState)
	}

	// New or re-notifiable conditions.
	for condType := range current {
		state := prev[condType]

		if state == nil {
			// New condition — fire.
			w.notify(ctx, ConditionEvent{
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
			w.notify(ctx, ConditionEvent{
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
			w.notify(ctx, ConditionEvent{
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

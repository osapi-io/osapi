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
	"time"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/retr0h/osapi/internal/job"
)

// ExportConditionState exposes the private conditionState type for testing.
type ExportConditionState = conditionState

// NewConditionState creates a conditionState for use in tests.
func NewConditionState(
	active bool,
	lastNotified time.Time,
) *conditionState {
	return &conditionState{
		active:       active,
		lastNotified: lastNotified,
	}
}

// WatcherGetPrev returns the internal prev state map for inspection in tests.
func WatcherGetPrev(
	w *Watcher,
) map[string]map[string]*conditionState {
	return w.prev
}

// WatcherSetPrev replaces the internal prev state map for test setup.
func WatcherSetPrev(
	w *Watcher,
	prev map[string]map[string]*conditionState,
) {
	w.prev = prev
}

// WatcherSetPrevEntry sets a single key in the prev state map for test setup.
func WatcherSetPrevEntry(
	w *Watcher,
	key string,
	entry map[string]*conditionState,
) {
	w.prev[key] = entry
}

// WatcherSetRenotifyInterval sets the renotify interval on a Watcher for testing.
func WatcherSetRenotifyInterval(
	w *Watcher,
	d time.Duration,
) {
	w.renotifyInterval = d
}

// WatcherSetKV sets the KV store on a Watcher for testing.
func WatcherSetKV(
	w *Watcher,
	kv jetstream.KeyValue,
) {
	w.kv = kv
}

// WatcherDetectTransitions exposes the private detectTransitions method for testing.
func WatcherDetectTransitions(
	w *Watcher,
	key string,
	componentType string,
	hostname string,
	conditions []job.Condition,
) {
	w.detectTransitions(key, componentType, hostname, conditions)
}

// WatcherHandleEntry exposes the private handleEntry method for testing.
func WatcherHandleEntry(
	ctx context.Context,
	w *Watcher,
	entry jetstream.KeyValueEntry,
) {
	w.handleEntry(ctx, entry)
}

// ExportResolveDisplayName exposes resolveDisplayName for testing.
func ExportResolveDisplayName(
	componentType string,
	identifier string,
	value []byte,
) string {
	return resolveDisplayName(componentType, identifier, value)
}

// ExportParseRegistryKey exposes parseRegistryKey for testing.
func ExportParseRegistryKey(
	key string,
) (string, string, bool) {
	return parseRegistryKey(key)
}

// WatcherHandleDelete exposes the private handleDelete method for testing.
func WatcherHandleDelete(
	ctx context.Context,
	w *Watcher,
	key string,
	componentType string,
	hostname string,
) {
	w.handleDelete(ctx, key, componentType, hostname)
}

// WatcherExtractConditions exposes the private extractConditions method for testing.
func WatcherExtractConditions(
	w *Watcher,
	key string,
	componentType string,
	value []byte,
) ([]job.Condition, bool) {
	return w.extractConditions(key, componentType, value)
}

// ParseRegistryKey exposes the private parseRegistryKey function for testing.
func ParseRegistryKey(
	key string,
) (componentType string, hostname string, ok bool) {
	return parseRegistryKey(key)
}

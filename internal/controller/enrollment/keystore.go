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

package enrollment

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/osapi-io/osapi/internal/pki"
)

// The three causes of a failed lookup stay distinct so an operator can tell
// "this agent has not re-enrolled yet" from "something is forging messages"
// from "the store could not be read". Collapsing them hides the difference
// at exactly the moment it matters.
var (
	// ErrAgentKeyNotFound means no key is stored for the machine ID. During
	// a rollout this is expected: no agent accepted before the store
	// existed has one.
	ErrAgentKeyNotFound = errors.New("no stored key for agent")

	// ErrNoPendingAgent means no enrollment request is waiting under the name
	// the caller gave. The API answers 404: there is nothing there to accept or
	// reject.
	ErrNoPendingAgent = errors.New("no pending agent")

	// ErrNoAcceptedAgent means no accepted agent has the hostname, or the
	// hostname and machine ID, the caller gave. The API answers 404.
	ErrNoAcceptedAgent = errors.New("no accepted agent")

	// ErrAmbiguousHostname means more than one accepted agent has the
	// hostname. Identity is the machine ID, so the caller names one. The API
	// answers 409.
	ErrAmbiguousHostname = errors.New("more than one accepted agent has this hostname")

	// ErrAgentKeyStoreUnavailable means the record could not be read or
	// decoded. It never resembles "no stored key", and is never treated as
	// verified.
	ErrAgentKeyStoreUnavailable = errors.New("agent key store unavailable")

	// ErrAgentKeySignatureMismatch means a key that is not this agent's
	// signed the message.
	ErrAgentKeySignatureMismatch = errors.New("agent signature mismatch")
)

// acceptedKey builds the KV key holding an accepted agent's record.
func acceptedKey(
	machineID string,
) string {
	return acceptedKVPrefix + machineID
}

// recordAgentKey stores an accepted agent's public key, keyed by machine ID,
// so it outlives the pending record deleted on acceptance. The fingerprint is
// recomputed from the key rather than copied from the request, because the
// request's fingerprint is self-reported.
//
// Called only from the acceptance path.
func (w *Watcher) recordAgentKey(
	ctx context.Context,
	pending PendingAgent,
) error {
	record := AcceptedAgent{
		MachineID:   pending.MachineID,
		Hostname:    pending.Hostname,
		PublicKey:   pending.PublicKey,
		Fingerprint: pki.FingerprintOf(pending.PublicKey),
		AcceptedAt:  nowFn(),
	}

	// A re-acceptance with a different key is a rotation. The outgoing key
	// stays acceptable for the configured grace period, so a message signed
	// moments before the rotation is not rejected for arriving late. The
	// expiry is stored as an instant rather than a duration: a restart cannot
	// then silently extend the window.
	//
	// Only read the previous record when a grace period is configured. With
	// none, a replacement takes effect at once and there is nothing to carry
	// forward.
	if w.rotationGrace > 0 {
		if previous, err := w.LookupAgentKey(ctx, pending.MachineID); err == nil &&
			!bytes.Equal(previous.PublicKey, record.PublicKey) {
			record.SupersededKey = previous.PublicKey
			record.SupersededUntil = nowFn().Add(w.rotationGrace)
		}
	}

	data, err := marshalFn(record)
	if err != nil {
		return fmt.Errorf(
			"marshal accepted agent %s: %w",
			pending.MachineID,
			err,
		)
	}

	key := acceptedKey(pending.MachineID)
	if _, err := w.enrollmentKV.Put(ctx, key, data); err != nil {
		return fmt.Errorf(
			"store accepted agent %s: %w",
			pending.MachineID,
			err,
		)
	}

	w.invalidateAgentKey(pending.MachineID)

	return nil
}

// LookupAgentKey returns the stored record for a machine ID. A missing
// record and an unreadable store are different answers: the first is
// ErrAgentKeyNotFound, the second ErrAgentKeyStoreUnavailable.
func (w *Watcher) LookupAgentKey(
	ctx context.Context,
	machineID string,
) (*AcceptedAgent, error) {
	if record, ok := w.cachedAgentKey(machineID); ok {
		return record, nil
	}

	entry, err := w.enrollmentKV.Get(ctx, acceptedKey(machineID))
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyNotFound) {
			return nil, fmt.Errorf("%w: %s", ErrAgentKeyNotFound, machineID)
		}

		return nil, errors.Join(
			ErrAgentKeyStoreUnavailable,
			fmt.Errorf("get accepted agent %s: %w", machineID, err),
		)
	}

	var record AcceptedAgent
	if err := unmarshalFn(entry.Value(), &record); err != nil {
		return nil, errors.Join(
			ErrAgentKeyStoreUnavailable,
			fmt.Errorf("unmarshal accepted agent %s: %w", machineID, err),
		)
	}

	w.cacheAgentKey(machineID, &record)

	return &record, nil
}

// RemoveAgentKey deletes an agent's stored record. After it returns, nothing
// signed by the removed key verifies.
func (w *Watcher) RemoveAgentKey(
	ctx context.Context,
	machineID string,
) error {
	if err := w.enrollmentKV.Delete(ctx, acceptedKey(machineID)); err != nil {
		return fmt.Errorf("remove accepted agent %s: %w", machineID, err)
	}

	w.invalidateAgentKey(machineID)

	return nil
}

// RemoveByHostname removes the stored key of the accepted agent with the
// hostname, and returns its machine ID. Two machines may share a hostname, so
// when more than one accepted agent has it, machineID names the one to remove;
// without it the call fails with ErrAmbiguousHostname rather than guess.
//
// Removing the key is the trust decision. The agent keeps the controller's key
// it was given at acceptance, so it does not enroll again by itself: an
// operator removes the controller key on the host and restarts the agent.
func (w *Watcher) RemoveByHostname(
	ctx context.Context,
	hostname string,
	machineID string,
) (string, error) {
	lister, err := w.enrollmentKV.ListKeys(ctx)
	if err != nil && !errors.Is(err, jetstream.ErrNoKeysFound) {
		return "", fmt.Errorf("list enrollment keys: %w", err)
	}

	var matches []string
	if err == nil {
		for key := range lister.Keys() {
			if !strings.HasPrefix(key, acceptedKVPrefix) {
				continue
			}

			entry, err := w.enrollmentKV.Get(ctx, key)
			if err != nil {
				return "", fmt.Errorf("get accepted agent %s: %w", key, err)
			}

			var record AcceptedAgent
			if err := unmarshalFn(entry.Value(), &record); err != nil {
				return "", fmt.Errorf("unmarshal accepted agent %s: %w", key, err)
			}

			if record.Hostname != hostname {
				continue
			}
			if machineID != "" && record.MachineID != machineID {
				continue
			}

			matches = append(matches, record.MachineID)
		}
	}

	switch {
	case len(matches) == 0:
		return "", fmt.Errorf("accepted agent with hostname %q: %w", hostname, ErrNoAcceptedAgent)
	case len(matches) > 1:
		slices.Sort(matches)

		return "", fmt.Errorf(
			"hostname %q, machine IDs %s: %w",
			hostname, strings.Join(matches, ", "), ErrAmbiguousHostname,
		)
	}

	if err := w.RemoveAgentKey(ctx, matches[0]); err != nil {
		return "", err
	}

	w.logger.Info(
		"removed accepted agent",
		slog.String("machine_id", matches[0]),
		slog.String("hostname", hostname),
	)

	return matches[0], nil
}

// cachedAgentKey returns a copy of the cached record for a machine ID, so a
// caller cannot mutate what the next lookup returns.
func (w *Watcher) cachedAgentKey(
	machineID string,
) (*AcceptedAgent, bool) {
	w.keyCacheMu.RLock()
	defer w.keyCacheMu.RUnlock()

	record, ok := w.keyCache[machineID]
	if !ok {
		return nil, false
	}

	cached := *record

	return &cached, true
}

// cacheAgentKey stores a copy of a record for later lookups. The cache is
// invalidated by acceptance and removal, never by elapsed time: a time-based
// entry would leave a removed agent verifying until it expired.
func (w *Watcher) cacheAgentKey(
	machineID string,
	record *AcceptedAgent,
) {
	w.keyCacheMu.Lock()
	defer w.keyCacheMu.Unlock()

	cached := *record
	w.keyCache[machineID] = &cached
}

// invalidateAgentKey drops a machine ID's cached record.
func (w *Watcher) invalidateAgentKey(
	machineID string,
) {
	w.keyCacheMu.Lock()
	defer w.keyCacheMu.Unlock()

	delete(w.keyCache, machineID)
}

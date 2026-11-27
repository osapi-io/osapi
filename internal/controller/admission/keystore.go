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

package admission

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
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

// recordAgentKey stores an accepted agent's public keys, keyed by machine ID,
// so it outlives the pending record deleted on acceptance. The fingerprint is
// recomputed from the key rather than copied from the request, because the
// request's fingerprint is self-reported.
//
// Called only from the acceptance path.
func (w *Watcher) recordAgentKey(
	ctx context.Context,
	pending PendingAgent,
	joinToken string,
) error {
	record := AcceptedAgent{
		JoinToken:     joinToken,
		JoinLabels:    pending.JoinLabels,
		NATSUserKey:   pending.NATSUserKey,
		MachineID:     pending.MachineID,
		Hostname:      pending.Hostname,
		PublicKey:     pending.PublicKey,
		EncryptionKey: pending.EncryptionKey,
		Fingerprint:   pki.FingerprintOfKeys(pending.PublicKey, pending.EncryptionKey),
		AcceptedAt:    nowFn(),
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
	if _, err := w.enrollmentKV.KVPut(ctx, w.enrollmentBucket, key, data); err != nil {
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
	gen := w.keyCacheGeneration()

	entry, err := w.enrollmentKV.KVGet(ctx, w.enrollmentBucket, acceptedKey(machineID))
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

	w.cacheAgentKey(machineID, &record, gen)

	return &record, nil
}

// RenameAgentKey records hostname as the name that an accepted machine
// answers as: after a hostname update through osapi, or a signed
// registration with a new name. The machine ID and the key stay the anchor.
// A name that a different admitted machine has, online or not, is refused.
func (w *Watcher) RenameAgentKey(
	ctx context.Context,
	machineID string,
	hostname string,
) error {
	record, err := w.LookupAgentKey(ctx, machineID)
	if err != nil {
		return err
	}
	// A different admitted machine keeps its name, also when it is offline.
	// Jobs and sealed secrets go to a hostname, so an agent that takes the
	// name of a different machine takes its work.
	if err := w.refuseHeldHostname(ctx, machineID, hostname); err != nil {
		return err
	}
	renamed := *record
	renamed.Hostname = hostname

	data, err := marshalFn(renamed)
	if err != nil {
		return fmt.Errorf("marshal accepted agent %s: %w", machineID, err)
	}
	if _, err := w.enrollmentKV.KVPut(
		ctx,
		w.enrollmentBucket,
		acceptedKey(machineID),
		data,
	); err != nil {
		return fmt.Errorf("rename accepted agent %s: %w", machineID, err)
	}

	w.invalidateAgentKey(machineID)

	return nil
}

// refuseHeldHostname is ErrHostnameHeld when an admitted machine other than
// machineID has hostname. A store that cannot be read is an error, so that
// no rename takes a name that it did not check.
func (w *Watcher) refuseHeldHostname(
	ctx context.Context,
	machineID string,
	hostname string,
) error {
	keys, err := w.enrollmentKV.KVKeys(ctx, w.enrollmentBucket)
	if err != nil {
		return fmt.Errorf("list admitted machines: %w", err)
	}
	for _, key := range keys {
		other, ok := strings.CutPrefix(key, acceptedKVPrefix)
		if !ok || other == machineID {
			continue
		}
		r, err := w.LookupAgentKey(ctx, other)
		if err != nil {
			return err
		}
		if r.Hostname == hostname {
			return fmt.Errorf("%w: %s is the hostname of %s", ErrHostnameHeld, hostname, other)
		}
	}

	return nil
}

// RemoveAgentKey deletes an agent's stored record. After it returns, nothing
// signed by the removed key verifies. With a revoker, the NATS user that the
// agent joined with is cut from the NATS server first: a revocation that is
// not recorded leaves the record, so the removal can run again.
func (w *Watcher) RemoveAgentKey(
	ctx context.Context,
	machineID string,
) error {
	if err := w.revokeNATSUser(ctx, machineID); err != nil {
		return err
	}
	if err := w.enrollmentKV.KVDelete(ctx, w.enrollmentBucket, acceptedKey(machineID)); err != nil {
		return fmt.Errorf("remove accepted agent %s: %w", machineID, err)
	}

	w.invalidateAgentKey(machineID)

	return nil
}

// revokeNATSUser revokes the NATS user of an admitted machine, when the
// watcher has a revoker and the machine joined with a user of its own. A
// machine that is not admitted has no user to revoke.
func (w *Watcher) revokeNATSUser(
	ctx context.Context,
	machineID string,
) error {
	if w.revoker == nil {
		return nil
	}
	record, err := w.LookupAgentKey(ctx, machineID)
	switch {
	case errors.Is(err, ErrAgentKeyNotFound):
		return nil
	case err != nil:
		return err
	case record.NATSUserKey == "":
		return nil
	}
	if err := w.revoker.Revoke(ctx, record.NATSUserKey); err != nil {
		return fmt.Errorf("revoke the NATS user of %s: %w", machineID, err)
	}
	w.logger.Info("revoked the NATS user of a machine",
		slog.String("machine_id", machineID),
		slog.String("hostname", record.Hostname),
	)

	return nil
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
	gen uint64,
) {
	w.keyCacheMu.Lock()
	defer w.keyCacheMu.Unlock()

	// A removal ran after this record was read: the record may be the one
	// that it removed.
	if w.keyCacheGen != gen {
		return
	}
	cached := *record
	w.keyCache[machineID] = &cached
}

// keyCacheGeneration is the count of removals from the cache so far.
func (w *Watcher) keyCacheGeneration() uint64 {
	w.keyCacheMu.RLock()
	defer w.keyCacheMu.RUnlock()

	return w.keyCacheGen
}

// invalidateAgentKey drops a machine ID's cached record.
func (w *Watcher) invalidateAgentKey(
	machineID string,
) {
	w.keyCacheMu.Lock()
	defer w.keyCacheMu.Unlock()

	delete(w.keyCache, machineID)
	w.keyCacheGen++
}

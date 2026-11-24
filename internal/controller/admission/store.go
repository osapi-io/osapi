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
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/osapi-io/osapi/internal/pki"
	"github.com/osapi-io/osapi/internal/telemetry/opspan"
)

// State is the state of an admission record. A machine has one at a time:
// the state of its latest record, by statePrecedence.
type State string

// The states of an admission record.
const (
	StatePending  State = "pending"
	StateAdmitted State = "admitted"
	StateRejected State = "rejected"
	StateRevoked  State = "revoked"
)

// rejectedKVPrefix and revokedKVPrefix are the key prefixes of the records
// that a decision against a machine leaves, so a machine that requests
// again is flagged with what was decided before.
const (
	rejectedKVPrefix = "rejected."
	revokedKVPrefix  = "revoked."
)

var (
	// ErrFingerprintMismatch is the answer to an admit whose fingerprint is
	// not the one the machine requested with: the operator is not looking
	// at the machine that asked.
	ErrFingerprintMismatch = errors.New("fingerprint does not match the request")

	// ErrNotFound is the answer to a get for a machine that osapi has no
	// record of.
	ErrNotFound = errors.New("no admission record")
)

// statePrecedence orders the records of one machine: a request waiting for
// a decision comes first, then the key in force, then the decisions
// against it.
var statePrecedence = map[State]int{
	StatePending:  0,
	StateAdmitted: 1,
	StateRejected: 2,
	StateRevoked:  3,
}

// Record is one machine's admission: its identity, its state, and when it
// got there. A pending record carries the state and the fingerprint of the
// machine's earlier record, when it had one, so a returning machine, or one
// whose key changed, is a decision, not a surprise.
type Record struct {
	MachineID   string `json:"machine_id"`
	Hostname    string `json:"hostname"`
	Fingerprint string `json:"fingerprint"`
	// EncryptionKey is the machine's public X25519 key, which the
	// fingerprint covers.
	EncryptionKey string    `json:"encryption_key,omitempty"`
	State         State     `json:"state"`
	RequestedAt   time.Time `json:"requested_at,omitempty"`
	AdmittedAt    time.Time `json:"admitted_at,omitempty"`
	// DecidedAt is when a machine was rejected or revoked.
	DecidedAt time.Time `json:"decided_at,omitempty"`
	// Reason is why a machine was rejected.
	Reason string `json:"reason,omitempty"`
	// Previous is the state of the machine's earlier record, on a pending
	// record; PreviousFingerprint is that record's fingerprint.
	Previous            State  `json:"previous,omitempty"`
	PreviousFingerprint string `json:"previous_fingerprint,omitempty"`
}

// List returns each machine's admission record, one per machine, in
// hostname order, in an admission.list span.
func (w *Watcher) List(
	ctx context.Context,
) ([]Record, error) {
	ctx, span := opspan.Start(ctx, tracerName, "admission.list")
	records, err := w.list(ctx)
	if err == nil {
		span.SetAttributes(attribute.Int("osapi.admission.count", len(records)))
	}

	return records, opspan.End(span, err)
}

// Get returns the admission record of a machine, or ErrNotFound, in an
// admission.get span.
func (w *Watcher) Get(
	ctx context.Context,
	machineID string,
) (Record, error) {
	ctx, span := opspan.Start(ctx, tracerName, "admission.get",
		attribute.String("osapi.agent.machine_id", machineID),
	)
	record, err := w.get(ctx, machineID)

	return record, opspan.End(span, err)
}

// Admit admits a pending machine, in an admission.admit span. fingerprint
// must be the one the machine requested with, which the operator reads on
// the machine itself: a mismatch is ErrFingerprintMismatch, and nothing is
// admitted. An earlier rejection or revocation of the machine is cleared.
func (w *Watcher) Admit(
	ctx context.Context,
	machineID string,
	fingerprint string,
) error {
	ctx, span := opspan.Start(ctx, tracerName, "admission.admit",
		attribute.String("osapi.agent.machine_id", machineID),
	)

	return opspan.End(span, w.admit(ctx, machineID, fingerprint))
}

// Reject rejects a pending machine with a reason, in an admission.reject
// span, and keeps a record of it.
func (w *Watcher) Reject(
	ctx context.Context,
	machineID string,
	reason string,
) error {
	ctx, span := opspan.Start(ctx, tracerName, "admission.reject",
		attribute.String("osapi.agent.machine_id", machineID),
	)

	return opspan.End(span, w.reject(ctx, machineID, reason))
}

// Revoke removes an admitted machine's key, in an admission.revoke span,
// and keeps a record of it. Nothing the machine signs verifies afterwards;
// to return, it requests admission again.
func (w *Watcher) Revoke(
	ctx context.Context,
	machineID string,
) error {
	ctx, span := opspan.Start(ctx, tracerName, "admission.revoke",
		attribute.String("osapi.agent.machine_id", machineID),
	)

	return opspan.End(span, w.revoke(ctx, machineID))
}

// list reads every record of the bucket and keeps, for each machine, the
// first by statePrecedence.
func (w *Watcher) list(
	ctx context.Context,
) ([]Record, error) {
	keys, err := w.enrollmentKV.KVKeys(ctx, w.enrollmentBucket)
	if err != nil {
		return nil, fmt.Errorf("list admission keys: %w", err)
	}

	byMachine := map[string][]Record{}
	for _, key := range keys {
		r, ok := w.read(ctx, key)
		if !ok {
			continue
		}
		byMachine[r.MachineID] = append(byMachine[r.MachineID], r)
	}

	records := make([]Record, 0, len(byMachine))
	for _, rs := range byMachine {
		records = append(records, merge(rs))
	}
	sort.Slice(records, func(a, b int) bool {
		if records[a].Hostname != records[b].Hostname {
			return records[a].Hostname < records[b].Hostname
		}

		return records[a].MachineID < records[b].MachineID
	})

	return records, nil
}

// get reads the records of one machine.
func (w *Watcher) get(
	ctx context.Context,
	machineID string,
) (Record, error) {
	var rs []Record
	for _, prefix := range []string{kvPrefix, acceptedKVPrefix, rejectedKVPrefix, revokedKVPrefix} {
		if r, ok := w.read(ctx, prefix+machineID); ok {
			rs = append(rs, r)
		}
	}
	if len(rs) == 0 {
		return Record{}, fmt.Errorf("%w: %s", ErrNotFound, machineID)
	}

	return merge(rs), nil
}

// merge is a machine's record from all of its records: the first by
// statePrecedence, and, when that one is pending, the state and the
// fingerprint of the next.
func merge(
	rs []Record,
) Record {
	sort.Slice(rs, func(a, b int) bool {
		return statePrecedence[rs[a].State] < statePrecedence[rs[b].State]
	})
	r := rs[0]
	if r.State == StatePending && len(rs) > 1 {
		r.Previous = rs[1].State
		r.PreviousFingerprint = rs[1].Fingerprint
	}

	return r
}

// read reads one key of the bucket as a record. A key of no known prefix,
// or a value that does not read, is no record.
func (w *Watcher) read(
	ctx context.Context,
	key string,
) (Record, bool) {
	var state State
	switch {
	case strings.HasPrefix(key, kvPrefix):
		state = StatePending
	case strings.HasPrefix(key, acceptedKVPrefix):
		state = StateAdmitted
	case strings.HasPrefix(key, rejectedKVPrefix):
		state = StateRejected
	case strings.HasPrefix(key, revokedKVPrefix):
		state = StateRevoked
	default:
		return Record{}, false
	}

	entry, err := w.enrollmentKV.KVGet(ctx, w.enrollmentBucket, key)
	if err != nil {
		return Record{}, false
	}

	switch state {
	case StatePending:
		var p PendingAgent
		if unmarshalFn(entry.Value(), &p) != nil {
			return Record{}, false
		}

		// The fingerprint of the keys the request carries, which are the
		// keys that admitting stores; the request's own fingerprint is what
		// the machine says of itself.
		return Record{
			MachineID:     p.MachineID,
			Hostname:      p.Hostname,
			Fingerprint:   pki.FingerprintOfKeys(p.PublicKey, p.EncryptionKey),
			EncryptionKey: p.EncryptionKey,
			State:         StatePending,
			RequestedAt:   p.RequestedAt,
		}, true
	case StateAdmitted:
		var a AcceptedAgent
		if unmarshalFn(entry.Value(), &a) != nil {
			return Record{}, false
		}

		return Record{
			MachineID:     a.MachineID,
			Hostname:      a.Hostname,
			Fingerprint:   a.Fingerprint,
			EncryptionKey: a.EncryptionKey,
			State:         StateAdmitted,
			AdmittedAt:    a.AcceptedAt,
		}, true
	default:
		var r Record
		if unmarshalFn(entry.Value(), &r) != nil {
			return Record{}, false
		}
		r.State = state

		return r, true
	}
}

// admit checks the fingerprint against the key of the request, admits the
// machine, and clears an earlier rejection or revocation.
func (w *Watcher) admit(
	ctx context.Context,
	machineID string,
	fingerprint string,
) error {
	r, ok := w.read(ctx, kvPrefix+machineID)
	if !ok {
		return fmt.Errorf("%w: %s", ErrNoPendingAgent, machineID)
	}
	if fingerprint == "" || fingerprint != r.Fingerprint {
		return fmt.Errorf("%w: %s", ErrFingerprintMismatch, machineID)
	}
	if err := w.AcceptAgent(ctx, machineID); err != nil {
		return err
	}
	for _, prefix := range []string{rejectedKVPrefix, revokedKVPrefix} {
		_ = w.enrollmentKV.KVDelete(ctx, w.enrollmentBucket, prefix+machineID)
	}

	return nil
}

// reject rejects the request and keeps a record of the decision.
func (w *Watcher) reject(
	ctx context.Context,
	machineID string,
	reason string,
) error {
	r, ok := w.read(ctx, kvPrefix+machineID)
	if !ok {
		return fmt.Errorf("%w: %s", ErrNoPendingAgent, machineID)
	}
	if err := w.RejectAgent(ctx, machineID, reason); err != nil {
		return err
	}

	return w.keep(ctx, rejectedKVPrefix, Record{
		MachineID:     machineID,
		Hostname:      r.Hostname,
		Fingerprint:   r.Fingerprint,
		EncryptionKey: r.EncryptionKey,
		RequestedAt:   r.RequestedAt,
		DecidedAt:     nowFn(),
		Reason:        reason,
	})
}

// revoke removes the admitted key and keeps a record of the decision.
func (w *Watcher) revoke(
	ctx context.Context,
	machineID string,
) error {
	r, ok := w.read(ctx, acceptedKVPrefix+machineID)
	if !ok {
		return fmt.Errorf("%w: %s", ErrNoAcceptedAgent, machineID)
	}
	if err := w.RemoveAgentKey(ctx, machineID); err != nil {
		return err
	}

	return w.keep(ctx, revokedKVPrefix, Record{
		MachineID:     machineID,
		Hostname:      r.Hostname,
		Fingerprint:   r.Fingerprint,
		EncryptionKey: r.EncryptionKey,
		AdmittedAt:    r.AdmittedAt,
		DecidedAt:     nowFn(),
	})
}

// keep stores the record of a decision under prefix.
func (w *Watcher) keep(
	ctx context.Context,
	prefix string,
	r Record,
) error {
	data, err := marshalFn(r)
	if err != nil {
		return fmt.Errorf(
			"marshal %s record %s: %w",
			strings.TrimSuffix(prefix, "."),
			r.MachineID,
			err,
		)
	}
	if _, err := w.enrollmentKV.KVPut(
		ctx,
		w.enrollmentBucket,
		prefix+r.MachineID,
		data,
	); err != nil {
		return fmt.Errorf(
			"store %s record %s: %w",
			strings.TrimSuffix(prefix, "."),
			r.MachineID,
			err,
		)
	}

	return nil
}

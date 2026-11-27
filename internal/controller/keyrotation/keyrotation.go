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

package keyrotation

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/osapi-io/osapi/internal/pki"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// interval is how often Run looks whether the agents have the next key.
var interval = 10 * time.Second

// Rotation rotates the controller's key.
type Rotation struct {
	cfg Config
	// mu keeps one Start or switch at a time in this process.
	mu sync.Mutex
}

// New is a Rotation with cfg.
func New(
	cfg Config,
) *Rotation {
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.RetiredKeyLifetime <= 0 {
		cfg.RetiredKeyLifetime = DefaultRetiredKeyLifetime
	}

	return &Rotation{cfg: cfg}
}

// read is the record and its revision, or the zero record and 0 when there
// is none.
func (r *Rotation) read(
	ctx context.Context,
) (pki.ControllerKeyRecord, uint64, error) {
	entry, err := r.cfg.KV.KVGet(ctx, r.cfg.Bucket, pki.ControllerKeyStateKey)
	if errors.Is(err, jetstream.ErrKeyNotFound) {
		return pki.ControllerKeyRecord{}, 0, nil
	}
	if err != nil {
		return pki.ControllerKeyRecord{}, 0, fmt.Errorf("read the controller key record: %w", err)
	}
	var rec pki.ControllerKeyRecord
	if err := json.Unmarshal(entry.Value(), &rec); err != nil {
		return pki.ControllerKeyRecord{}, 0, fmt.Errorf("the controller key record: %w", err)
	}

	return rec, entry.Revision(), nil
}

// write signs rec with priv and writes it in place of the record at
// revision, or as the first record at revision 0.
func (r *Rotation) write(
	ctx context.Context,
	rec pki.ControllerKeyRecord,
	priv ed25519.PrivateKey,
	revision uint64,
) error {
	rec.Sign(priv)
	// A record of byte slices, a number and times always marshals.
	data, _ := json.Marshal(rec)
	var err error
	if revision == 0 {
		_, err = r.cfg.KV.KVCreate(ctx, r.cfg.Bucket, pki.ControllerKeyStateKey, data)
	} else {
		_, err = r.cfg.KV.KVUpdate(ctx, r.cfg.Bucket, pki.ControllerKeyStateKey, data, revision)
	}
	if err != nil {
		return fmt.Errorf("write the controller key record: %w", err)
	}

	return nil
}

// trusted is the record when the controller's current key signed it. An
// agent can write the state bucket, so a record that the key did not sign
// is not the controller's: it is the zero record, as if there were none.
func (r *Rotation) trusted(
	rec pki.ControllerKeyRecord,
) pki.ControllerKeyRecord {
	if rec.SignedBy(r.cfg.Keys.PublicKey()) {
		return rec
	}

	return pki.ControllerKeyRecord{}
}

// Load takes the record at the start of the controller: the previous key
// after a switch keeps verifying the records that it signed, and a switch
// that stopped after its renames writes its record.
func (r *Rotation) Load(
	ctx context.Context,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	rec, revision, err := r.read(ctx)
	if err != nil {
		return err
	}
	pub := r.cfg.Keys.PublicKey()
	switch {
	case bytes.Equal(rec.Current, pub) && rec.SignedBy(pub):
		if len(rec.Previous) > 0 {
			r.cfg.Keys.SetPreviousControllerPublicKey(rec.Previous)
		}
		r.cfg.Keys.SetRetiredControllerKeys(rec.Retired, r.cfg.RetiredKeyLifetime)
	case rec.Rotating() && bytes.Equal(rec.Next, pub) && rec.SignedBy(rec.Current):
		// The keys moved, and the record of the switch was not written.
		r.cfg.Keys.SetPreviousControllerPublicKey(rec.Current)
		r.cfg.Keys.SetRetiredControllerKeys(r.retire(rec), r.cfg.RetiredKeyLifetime)

		return r.writeSwitch(ctx, rec, revision)
	}

	return nil
}

// Start starts a rotation: it makes the next key beside the current one,
// and writes the record that gives it to the agents. A rotation in progress
// is not started again: Start shows it.
func (r *Rotation) Start(
	ctx context.Context,
) (Status, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rec, revision, err := r.read(ctx)
	if err != nil {
		return Status{}, err
	}
	rec = r.trusted(rec)
	if !rec.Rotating() {
		next, err := r.cfg.Keys.Next()
		if err != nil {
			return Status{}, fmt.Errorf("make the next controller key: %w", err)
		}
		serial := rec.Serial + 1
		rec = pki.ControllerKeyRecord{
			Serial:     serial,
			Current:    r.cfg.Keys.PublicKey(),
			Next:       next.PublicKey(),
			Previous:   rec.Previous,
			Retired:    r.unexpired(rec.Retired),
			StartedAt:  r.cfg.Now().UTC(),
			SwitchedAt: rec.SwitchedAt,
		}
		if err := r.write(ctx, rec, r.cfg.Keys.PrivateKey(), revision); err != nil {
			return Status{}, err
		}
		r.cfg.Logger.Info("started the rotation of the controller's key",
			slog.String("next", pki.FingerprintOf(rec.Next)))
	}

	return r.status(ctx, rec)
}

// Status is the rotation now.
func (r *Rotation) Status(
	ctx context.Context,
) (Status, error) {
	rec, _, err := r.read(ctx)
	if err != nil {
		return Status{}, err
	}

	return r.status(ctx, r.trusted(rec))
}

// status is rec as the API shows it, with the agents that a switch waits for.
func (r *Rotation) status(
	ctx context.Context,
	rec pki.ControllerKeyRecord,
) (Status, error) {
	s := Status{Current: r.cfg.Keys.Fingerprint()}
	if !rec.Rotating() {
		return s, nil
	}
	s.Next = pki.FingerprintOf(rec.Next)
	s.StartedAt = rec.StartedAt
	s.SwitchBy = rec.StartedAt.Add(r.cfg.Grace)
	agents, err := r.waitingFor(ctx, s.Next)
	if err != nil {
		return Status{}, err
	}
	s.Agents = agents

	return s, nil
}

// waitingFor are the admitted agents that reported within the grace
// period, and whether each reports next. An admitted machine that has not
// reported for longer is cut off when the grace period ends in any case,
// so it does not hold the switch.
func (r *Rotation) waitingFor(
	ctx context.Context,
	next string,
) ([]Agent, error) {
	if r.cfg.Admitted == nil || r.cfg.Seen == nil {
		return nil, nil
	}
	records, err := r.cfg.Admitted.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list the admitted agents: %w", err)
	}
	infos, err := r.cfg.Seen.ListAgents(ctx)
	if err != nil {
		return nil, fmt.Errorf("list the agents: %w", err)
	}
	last, err := r.cfg.Seen.AgentsLastSeen(ctx)
	if err != nil {
		return nil, fmt.Errorf("read when the agents last reported: %w", err)
	}
	has := make(map[string]bool, len(infos))
	online := make(map[string]bool, len(infos))
	for _, a := range infos {
		online[a.MachineID] = true
		has[a.MachineID] = a.ControllerKeyNext == next || a.ControllerKey == next
	}
	now := r.cfg.Now()
	agents := []Agent{}
	for _, rec := range records {
		if rec.State != types.AdmissionAdmitted {
			continue
		}
		seen := last[rec.MachineID]
		if !online[rec.MachineID] && (seen.IsZero() || now.Sub(seen) > r.cfg.Grace) {
			continue
		}
		agents = append(agents, Agent{
			MachineID: rec.MachineID, Hostname: rec.Hostname, HasNext: has[rec.MachineID],
		})
	}

	return agents, nil
}

// Step switches to the next key when each agent that the switch waits for
// has it, or the grace period ended. It reports whether it switched.
func (r *Rotation) Step(
	ctx context.Context,
) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rec, revision, err := r.read(ctx)
	if err != nil {
		return false, err
	}
	rec = r.trusted(rec)
	if !rec.Rotating() {
		return false, nil
	}
	if r.cfg.Now().Before(rec.StartedAt.Add(r.cfg.Grace)) {
		agents, err := r.waitingFor(ctx, pki.FingerprintOf(rec.Next))
		if err != nil {
			return false, err
		}
		for _, a := range agents {
			if !a.HasNext {
				return false, nil
			}
		}
	}
	next, err := r.cfg.Keys.Next()
	if err != nil {
		return false, fmt.Errorf("read the next controller key: %w", err)
	}
	if !bytes.Equal(next.PublicKey(), rec.Next) {
		return false, fmt.Errorf(
			"the next controller key on the host is not the one of the record: "+
				"%s, and the record has %s",
			next.Fingerprint(),
			pki.FingerprintOf(rec.Next),
		)
	}
	old := r.cfg.Keys.PublicKey()
	if err := r.cfg.Keys.Promote(next); err != nil {
		return false, fmt.Errorf("switch to the next controller key: %w", err)
	}
	r.cfg.Keys.RotateControllerKey(r.cfg.Keys.PublicKey())
	r.cfg.Keys.SetPreviousControllerPublicKey(old)
	r.cfg.Keys.SetRetiredControllerKeys(r.retire(rec), r.cfg.RetiredKeyLifetime)
	if err := r.writeSwitch(ctx, rec, revision); err != nil {
		return false, err
	}

	return true, nil
}

// writeSwitch writes the record of the switch to rec's next key, which is
// the controller's key now, and which signs it.
func (r *Rotation) writeSwitch(
	ctx context.Context,
	rec pki.ControllerKeyRecord,
	revision uint64,
) error {
	done := pki.ControllerKeyRecord{
		Serial:     rec.Serial + 1,
		Current:    rec.Next,
		Previous:   rec.Current,
		Retired:    r.retire(rec),
		SwitchedAt: r.cfg.Now().UTC(),
	}
	if err := r.write(ctx, done, r.cfg.Keys.PrivateKey(), revision); err != nil {
		return err
	}
	r.cfg.Logger.Info("switched to the next controller key",
		slog.String("controller_key", pki.FingerprintOf(done.Current)),
		slog.String("previous", pki.FingerprintOf(done.Previous)))

	return nil
}

// Run switches the key when Step says so, every interval, until ctx is
// done. An error of one step is logged, and the next step tries again.
func (r *Rotation) Run(
	ctx context.Context,
) error {
	if err := r.Load(ctx); err != nil {
		r.cfg.Logger.Warn(
			"failed to load the controller key record",
			slog.String("error", err.Error()),
		)
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if _, err := r.Step(ctx); err != nil {
				r.cfg.Logger.Warn("the rotation of the controller key did not step",
					slog.String("error", err.Error()))
			}
		}
	}
}

// retire is the retired keys of the record that a switch from rec writes:
// rec's retired keys that have not expired, then its previous key, which the
// switch replaces. The previous key stopped signing at rec's switch, the
// one that made it the previous key.
func (r *Rotation) retire(
	rec pki.ControllerKeyRecord,
) []pki.RetiredKey {
	retired := r.unexpired(rec.Retired)
	if len(rec.Previous) > 0 {
		retired = append(retired, pki.RetiredKey{Key: rec.Previous, RetiredAt: rec.SwitchedAt})
	}

	return retired
}

// unexpired is each key of keys that still verifies a stored record now.
// The record drops an expired key: it verifies nothing.
func (r *Rotation) unexpired(
	keys []pki.RetiredKey,
) []pki.RetiredKey {
	now := r.cfg.Now()
	out := []pki.RetiredKey{}
	for _, k := range keys {
		if !k.Expired(now, r.cfg.RetiredKeyLifetime) {
			out = append(out, k)
		}
	}
	if len(out) == 0 {
		return nil
	}

	return out
}

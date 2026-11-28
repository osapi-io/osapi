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

package certrotation

import (
	"context"
	"crypto/ed25519"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/osfs"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/pki"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// hostFS is the filesystem of the host.
var hostFS avfs.VFS = osfs.NewWithNoIdm()

// interval is how often Run looks whether the agents trust the next
// certificate.
var interval = 10 * time.Second

// makeCert makes the next certificate. Tests swap it.
var makeCert = config.MakeSelfSigned

// Rotation rotates the NATS server's certificate.
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

	return &Rotation{cfg: cfg}
}

// NextFiles are the files of the next certificate of src, beside its own:
// nats.next.crt and nats.next.key for nats.crt.
func NextFiles(
	src pki.CertSource,
) (string, string) {
	return strings.TrimSuffix(src.CertFile, ".crt") + ".next.crt",
		strings.TrimSuffix(src.KeyFile, ".key") + ".next.key"
}

// read is the record and its revision, or the zero record and 0 when there
// is none, or the zero record and its revision when it does not unmarshal.
func (r *Rotation) read(
	ctx context.Context,
) (pki.CertRecord, uint64, error) {
	entry, err := r.cfg.KV.KVGet(ctx, r.cfg.Bucket, pki.NATSCertStateKey)
	if errors.Is(err, jetstream.ErrKeyNotFound) {
		return pki.CertRecord{}, 0, nil
	}
	if err != nil {
		return pki.CertRecord{}, 0, fmt.Errorf("read the certificate record: %w", err)
	}
	// An agent can write the state bucket, so a record that does not
	// unmarshal is not the controller's, as one that its key did not sign:
	// it is the zero record at its revision, which the next Start replaces.
	var rec pki.CertRecord
	if err := json.Unmarshal(entry.Value(), &rec); err != nil {
		r.cfg.Logger.Warn("the certificate record does not read, and the next "+
			"rotation replaces it", slog.String("error", err.Error()))

		return pki.CertRecord{}, entry.Revision(), nil
	}

	return rec, entry.Revision(), nil
}

// write signs rec with the controller's key and writes it in place of the
// record at revision, or as the first record at revision 0.
func (r *Rotation) write(
	ctx context.Context,
	rec pki.CertRecord,
	revision uint64,
) error {
	rec.Sign(r.cfg.Keys.PrivateKey())
	// A record of byte slices, a number and strings always marshals.
	data, _ := json.Marshal(rec)
	var err error
	if revision == 0 {
		_, err = r.cfg.KV.KVCreate(ctx, r.cfg.Bucket, pki.NATSCertStateKey, data)
	} else {
		_, err = r.cfg.KV.KVUpdate(ctx, r.cfg.Bucket, pki.NATSCertStateKey, data, revision)
	}
	if err != nil {
		return fmt.Errorf("write the certificate record: %w", err)
	}

	return nil
}

// trusted is the record when the controller's key signed it, or its previous
// key, which signed the records before a rotation of the controller's key.
// An agent can write the state bucket, so a record that neither signed is
// not the controller's: it is the zero record, as if there were none.
func (r *Rotation) trusted(
	rec pki.CertRecord,
) pki.CertRecord {
	verify := func(data []byte, signature []byte) bool {
		for _, key := range []ed25519.PublicKey{
			r.cfg.Keys.PublicKey(), r.cfg.Keys.PreviousControllerPublicKey(),
		} {
			if len(key) == ed25519.PublicKeySize && ed25519.Verify(key, data, signature) {
				return true
			}
		}

		return false
	}
	if rec.SignedWith(verify) {
		return rec
	}

	return pki.CertRecord{}
}

// current is the PEM certificate that the server serves now.
func (r *Rotation) current() ([]byte, error) {
	data, err := avfs.ReadFile(hostFS, r.cfg.Source.CertFile)
	if err != nil {
		return nil, fmt.Errorf(
			"read %s: rotate it on the host of the NATS server: %w",
			r.cfg.Source.CertFile,
			err,
		)
	}

	return data, nil
}

// Start starts a rotation: it makes the next certificate beside the current
// one, and writes the record that gives both to the agents. A rotation in
// progress is not started again: Start shows it.
func (r *Rotation) Start(
	ctx context.Context,
) (Status, error) {
	if !r.cfg.Source.SelfSigned {
		return Status{}, ErrOperatorCert
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	rec, revision, err := r.read(ctx)
	if err != nil {
		return Status{}, err
	}
	rec = r.trusted(rec)
	if !rec.Rotating() {
		current, err := r.current()
		if err != nil {
			return Status{}, err
		}
		certFile, keyFile := NextFiles(r.cfg.Source)
		if err := makeCert(certFile, keyFile); err != nil {
			return Status{}, fmt.Errorf("make the next certificate: %w", err)
		}
		next, err := avfs.ReadFile(hostFS, certFile)
		if err != nil {
			return Status{}, fmt.Errorf("read the next certificate: %w", err)
		}
		rec = pki.CertRecord{
			Serial:    rec.Serial + 1,
			Name:      r.cfg.Source.Name,
			Current:   current,
			Next:      next,
			StartedAt: r.cfg.Now().UTC().Format(time.RFC3339Nano),
		}
		if err := r.write(ctx, rec, revision); err != nil {
			return Status{}, err
		}
		r.cfg.Logger.Info("started the rotation of the NATS server's certificate",
			slog.String("next", pki.CertFingerprint(next)))
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
	rec = r.trusted(rec)
	if len(rec.Current) == 0 {
		current, err := r.current()
		if err != nil {
			return Status{}, err
		}
		rec.Current = current
	}

	return r.status(ctx, rec)
}

// status is rec as the API shows it, with the agents that a switch waits for.
func (r *Rotation) status(
	ctx context.Context,
	rec pki.CertRecord,
) (Status, error) {
	s := Status{Current: pki.CertFingerprint(rec.Current)}
	if !rec.Rotating() {
		return s, nil
	}
	s.Next = pki.CertFingerprint(rec.Next)
	// The record holds a time that write formatted.
	s.StartedAt, _ = time.Parse(time.RFC3339Nano, rec.StartedAt)
	s.SwitchBy = s.StartedAt.Add(r.cfg.Grace)
	agents, err := r.waitingFor(ctx, s.Next)
	if err != nil {
		return Status{}, err
	}
	s.Agents = agents

	return s, nil
}

// waitingFor are the admitted agents that reported within the grace
// period, and whether each trusts next. An admitted machine that has not
// reported for longer does not hold the switch.
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
		has[a.MachineID] = a.NATSCertNext == next
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

// Step switches to the next certificate when each agent that the switch
// waits for trusts it, or the grace period ended. It reports whether it
// switched.
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
	next := pki.CertFingerprint(rec.Next)
	start, _ := time.Parse(time.RFC3339Nano, rec.StartedAt)
	if r.cfg.Now().Before(start.Add(r.cfg.Grace)) {
		agents, err := r.waitingFor(ctx, next)
		if err != nil {
			return false, err
		}
		for _, a := range agents {
			if !a.HasNext {
				return false, nil
			}
		}
	}
	if err := r.promote(next); err != nil {
		return false, err
	}
	done := pki.CertRecord{Serial: rec.Serial + 1, Name: rec.Name, Current: rec.Next}
	if err := r.write(ctx, done, revision); err != nil {
		return false, err
	}
	r.cfg.Logger.Info("switched the NATS server to its next certificate",
		slog.String("fingerprint", next))

	return true, nil
}

// promote renames the next files into place: the key first, so that a
// handshake between the two renames finds a pair that does not load, and
// the server keeps the last one. A switch that stopped after its renames
// finds the next certificate in place, and only writes its record. A
// switch that stopped between them finds the next key in place, and
// renames only the certificate.
func (r *Rotation) promote(
	next string,
) error {
	certFile, keyFile := NextFiles(r.cfg.Source)
	data, err := avfs.ReadFile(hostFS, certFile)
	if errors.Is(err, os.ErrNotExist) {
		current, cerr := r.current()
		if cerr == nil && pki.CertFingerprint(current) == next {
			return nil
		}
	}
	if err != nil {
		return fmt.Errorf("read the next certificate: %w", err)
	}
	if pki.CertFingerprint(data) != next {
		return fmt.Errorf(
			"the next certificate on the host is not the one of the record: %s, "+
				"and the record has %s", pki.CertFingerprint(data), next)
	}
	if err := hostFS.Rename(keyFile, r.cfg.Source.KeyFile); err != nil {
		if !errors.Is(err, os.ErrNotExist) || !pairs(certFile, r.cfg.Source.KeyFile) {
			return fmt.Errorf("switch to the next certificate: %w", err)
		}
	}
	if err := hostFS.Rename(certFile, r.cfg.Source.CertFile); err != nil {
		return fmt.Errorf("switch to the next certificate: %w", err)
	}

	return nil
}

// Run switches the certificate when Step says so, every interval, until
// ctx is done. An error of one step is logged, and the next step tries
// again.
func (r *Rotation) Run(
	ctx context.Context,
) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if _, err := r.Step(ctx); err != nil {
				r.cfg.Logger.Warn("the rotation of the NATS certificate did not step",
					slog.String("error", err.Error()))
			}
		}
	}
}

// pairs is whether the certificate in certFile and the key in keyFile are
// a pair.
func pairs(
	certFile string,
	keyFile string,
) bool {
	_, err := tls.LoadX509KeyPair(certFile, keyFile)

	return err == nil
}

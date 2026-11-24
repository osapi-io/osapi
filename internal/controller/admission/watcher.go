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
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/osapi-io/osapi/internal/pki"
)

// marshalFn is the JSON marshal function (injectable for testing).
var marshalFn = json.Marshal

// unmarshalFn is the JSON unmarshal function (injectable for testing).
var unmarshalFn = json.Unmarshal

// nowFn returns the current time (injectable for testing).
var nowFn = time.Now

// kvPrefix is the key prefix for pending enrollment entries.
const kvPrefix = "enrollment."

// acceptedKVPrefix is the key prefix for accepted agents' stored keys. Both
// prefixes share the enrollment bucket, so every scan of pending entries
// must filter on kvPrefix rather than reading whatever keys it finds.
const acceptedKVPrefix = "accepted."

// Watcher monitors NATS for agent enrollment requests and manages
// pending agents in a JetStream KV bucket.
type Watcher struct {
	logger       *slog.Logger
	nc           NATSSubscriber
	enrollmentKV KVStore
	// enrollmentBucket is the name of the enrollment KV bucket.
	enrollmentBucket string
	pkiProvider      PKIProvider
	autoAdmit        bool
	namespace        string

	// keyCache holds accepted agents' records by machine ID. Verification
	// touches it for every response and heartbeat, so the lookup does not
	// hit the KV in steady state. Invalidated by acceptance and removal.
	keyCacheMu sync.RWMutex
	keyCache   map[string]*AcceptedAgent

	// rotationGrace is how long the key an acceptance replaces stays
	// acceptable. Zero means a replacement takes effect at once, which is
	// also the behaviour when nothing configures it.
	rotationGrace time.Duration
}

// SetRotationGracePeriod sets how long a replaced key keeps verifying after a
// rotation. Called once at startup from the controller's PKI configuration.
//
// Without it a rotation rejects everything signed in the moments before the new
// key was accepted, which is the reason operators turn verification off.
func (w *Watcher) SetRotationGracePeriod(
	d time.Duration,
) {
	w.rotationGrace = d
}

// NewWatcher creates a new enrollment Watcher.
func NewWatcher(
	logger *slog.Logger,
	nc NATSSubscriber,
	enrollmentKV KVStore,
	enrollmentBucket string,
	pkiProvider PKIProvider,
	autoAdmit bool,
	namespace string,
) *Watcher {
	return &Watcher{
		logger:           logger.With(slog.String("subsystem", "controller.pki")),
		nc:               nc,
		enrollmentKV:     enrollmentKV,
		enrollmentBucket: enrollmentBucket,
		pkiProvider:      pkiProvider,
		autoAdmit:        autoAdmit,
		namespace:        namespace,
		keyCache:         make(map[string]*AcceptedAgent),
	}
}

// Start subscribes to the enrollment request subject and blocks until
// the context is cancelled. Returns nil on clean shutdown.
func (w *Watcher) Start(
	ctx context.Context,
) error {
	subject := enrollSubject(w.namespace, pki.EnrollRequestSuffix)

	sub, err := w.nc.Subscribe(ctx, subject, func(ctx context.Context, msg *nats.Msg) {
		w.handleEnrollmentRequest(ctx, msg)
	})
	if err != nil {
		return fmt.Errorf("subscribe to enrollment requests: %w", err)
	}

	w.logger.Info(
		"enrollment watcher started",
		slog.String("subject", subject),
		slog.Bool("auto_admit", w.autoAdmit),
	)

	<-ctx.Done()

	if err := sub.Unsubscribe(); err != nil {
		w.logger.Warn(
			"failed to unsubscribe from enrollment requests",
			slog.String("error", err.Error()),
		)
	}

	return nil
}

// handleEnrollmentRequest processes an incoming enrollment request from
// an agent. It stores the request in KV and optionally admits it at once.
func (w *Watcher) handleEnrollmentRequest(
	ctx context.Context,
	msg *nats.Msg,
) {
	var req pki.EnrollmentRequest
	if err := unmarshalFn(msg.Data, &req); err != nil {
		w.logger.Warn(
			"failed to unmarshal enrollment request",
			slog.String("error", err.Error()),
		)
		return
	}

	// The fingerprint that admits the agent covers its encryption key. An
	// agent older than the controller sends none, and there is no
	// compatibility: it is upgraded and enrolls again.
	if !pki.ValidEncryptionKey(req.EncryptionKey) {
		w.logger.Warn(
			"enrollment request has no valid encryption key; upgrade the agent",
			slog.String("machine_id", req.MachineID),
			slog.String("hostname", req.Hostname),
		)
		return
	}

	w.logger.Info(
		"received enrollment request",
		slog.String("machine_id", req.MachineID),
		slog.String("hostname", req.Hostname),
		slog.String("fingerprint", req.Fingerprint),
	)

	pending := PendingAgent{
		MachineID:     req.MachineID,
		Hostname:      req.Hostname,
		PublicKey:     req.PublicKey,
		EncryptionKey: req.EncryptionKey,
		Fingerprint:   req.Fingerprint,
		RequestedAt:   nowFn(),
	}

	data, err := marshalFn(pending)
	if err != nil {
		w.logger.Warn(
			"failed to marshal pending agent",
			slog.String("machine_id", req.MachineID),
			slog.String("error", err.Error()),
		)
		return
	}

	key := kvPrefix + req.MachineID
	if _, err := w.enrollmentKV.KVPut(ctx, w.enrollmentBucket, key, data); err != nil {
		w.logger.Warn(
			"failed to store pending enrollment",
			slog.String("machine_id", req.MachineID),
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		return
	}

	w.logger.Info(
		"stored pending enrollment",
		slog.String("machine_id", req.MachineID),
		slog.String("key", key),
	)

	if w.autoAdmit {
		if err := w.AcceptAgent(ctx, req.MachineID); err != nil {
			w.logger.Warn(
				"auto-admit failed",
				slog.String("machine_id", req.MachineID),
				slog.String("error", err.Error()),
			)
		}
	}
}

// enrollSubject builds a namespaced NATS subject. When namespace is
// empty, the suffix is returned as-is.
func enrollSubject(
	namespace string,
	suffix string,
) string {
	if namespace == "" {
		return suffix
	}

	return namespace + "." + suffix
}

// tracerName names the spans of enrollment in a trace.
const tracerName = "osapi-admission"

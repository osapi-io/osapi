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
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/retr0h/osapi/internal/agent/pki"
)

// marshalFn is the JSON marshal function (injectable for testing).
var marshalFn = json.Marshal

// unmarshalFn is the JSON unmarshal function (injectable for testing).
var unmarshalFn = json.Unmarshal

// nowFn returns the current time (injectable for testing).
var nowFn = time.Now

// kvPrefix is the key prefix for pending enrollment entries.
const kvPrefix = "enrollment."

// Watcher monitors NATS for agent enrollment requests and manages
// pending agents in a JetStream KV bucket.
type Watcher struct {
	logger       *slog.Logger
	nc           NATSSubscriber
	enrollmentKV KVStore
	pkiProvider  PKIProvider
	autoAccept   bool
	namespace    string
}

// NewWatcher creates a new enrollment Watcher.
func NewWatcher(
	logger *slog.Logger,
	nc NATSSubscriber,
	enrollmentKV KVStore,
	pkiProvider PKIProvider,
	autoAccept bool,
	namespace string,
) *Watcher {
	return &Watcher{
		logger:       logger.With(slog.String("subsystem", "controller.pki")),
		nc:           nc,
		enrollmentKV: enrollmentKV,
		pkiProvider:  pkiProvider,
		autoAccept:   autoAccept,
		namespace:    namespace,
	}
}

// Start subscribes to the enrollment request subject and blocks until
// the context is cancelled. Returns nil on clean shutdown.
func (w *Watcher) Start(
	ctx context.Context,
) error {
	subject := enrollSubject(w.namespace, pki.EnrollRequestSuffix)

	sub, err := w.nc.Subscribe(subject, func(msg *nats.Msg) {
		w.handleEnrollmentRequest(ctx, msg)
	})
	if err != nil {
		return fmt.Errorf("subscribe to enrollment requests: %w", err)
	}

	w.logger.Info(
		"enrollment watcher started",
		slog.String("subject", subject),
		slog.Bool("auto_accept", w.autoAccept),
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
// an agent. It stores the request in KV and optionally auto-accepts.
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

	w.logger.Info(
		"received enrollment request",
		slog.String("machine_id", req.MachineID),
		slog.String("hostname", req.Hostname),
		slog.String("fingerprint", req.Fingerprint),
	)

	pending := PendingAgent{
		MachineID:   req.MachineID,
		Hostname:    req.Hostname,
		PublicKey:   req.PublicKey,
		Fingerprint: req.Fingerprint,
		RequestedAt: nowFn(),
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
	if _, err := w.enrollmentKV.Put(ctx, key, data); err != nil {
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

	if w.autoAccept {
		if err := w.AcceptAgent(ctx, req.MachineID); err != nil {
			w.logger.Warn(
				"auto-accept failed",
				slog.String("machine_id", req.MachineID),
				slog.String("error", err.Error()),
			)
		}
	}
}

// ListPending returns all pending agents from the enrollment KV bucket.
func (w *Watcher) ListPending(
	ctx context.Context,
) ([]PendingAgent, error) {
	lister, err := w.enrollmentKV.ListKeys(ctx)
	if err != nil {
		// jetstream.ErrNoKeysFound means the bucket is empty.
		if err == jetstream.ErrNoKeysFound {
			return nil, nil
		}
		return nil, fmt.Errorf("list enrollment keys: %w", err)
	}

	var pending []PendingAgent

	for key := range lister.Keys() {
		entry, err := w.enrollmentKV.Get(ctx, key)
		if err != nil {
			w.logger.Warn(
				"failed to get enrollment entry",
				slog.String("key", key),
				slog.String("error", err.Error()),
			)
			continue
		}

		var agent PendingAgent
		if err := unmarshalFn(entry.Value(), &agent); err != nil {
			w.logger.Warn(
				"failed to unmarshal enrollment entry",
				slog.String("key", key),
				slog.String("error", err.Error()),
			)
			continue
		}

		pending = append(pending, agent)
	}

	return pending, nil
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

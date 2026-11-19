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
	"fmt"
	"log/slog"

	"github.com/osapi-io/osapi/internal/pki"
)

// AcceptAgent accepts a pending agent by machine ID. It publishes an
// acceptance response containing the controller's public key to the
// agent's response subject and deletes the pending entry from KV.
func (w *Watcher) AcceptAgent(
	ctx context.Context,
	machineID string,
) error {
	key := kvPrefix + machineID

	entry, err := w.enrollmentKV.KVGet(ctx, w.enrollmentBucket, key)
	if err != nil {
		return fmt.Errorf("get pending agent %s: %w", machineID, err)
	}

	var pending PendingAgent
	if err := unmarshalFn(entry.Value(), &pending); err != nil {
		return fmt.Errorf("unmarshal pending agent %s: %w", machineID, err)
	}

	// Store the agent's key before anything is published: the store is the
	// authority for every later verification, and acceptance is the only
	// event allowed to write it. Deleting the pending entry below discards
	// the only other copy.
	if err := w.recordAgentKey(ctx, pending); err != nil {
		return err
	}

	resp := pki.EnrollmentResponse{
		Accepted:            true,
		ControllerPublicKey: w.pkiProvider.PublicKey(),
	}

	data, err := marshalFn(resp)
	if err != nil {
		return fmt.Errorf("marshal acceptance response: %w", err)
	}

	subject := enrollSubject(w.namespace, pki.EnrollResponsePrefix+"."+machineID)
	if err := w.nc.PublishCore(ctx, subject, data); err != nil {
		return fmt.Errorf("publish acceptance for %s: %w", machineID, err)
	}

	if err := w.enrollmentKV.KVDelete(ctx, w.enrollmentBucket, key); err != nil {
		return fmt.Errorf("delete pending agent %s: %w", machineID, err)
	}

	w.logger.Info(
		"accepted agent enrollment",
		slog.String("machine_id", machineID),
		slog.String("hostname", pending.Hostname),
		slog.String("fingerprint", pending.Fingerprint),
	)

	return nil
}

// RejectAgent rejects a pending agent by machine ID. It publishes a
// rejection response to the agent's response subject and deletes the
// pending entry from KV.
func (w *Watcher) RejectAgent(
	ctx context.Context,
	machineID string,
	reason string,
) error {
	key := kvPrefix + machineID

	entry, err := w.enrollmentKV.KVGet(ctx, w.enrollmentBucket, key)
	if err != nil {
		return fmt.Errorf("get pending agent %s: %w", machineID, err)
	}

	var pending PendingAgent
	if err := unmarshalFn(entry.Value(), &pending); err != nil {
		return fmt.Errorf("unmarshal pending agent %s: %w", machineID, err)
	}

	resp := pki.EnrollmentResponse{
		Accepted: false,
		Reason:   reason,
	}

	data, err := marshalFn(resp)
	if err != nil {
		return fmt.Errorf("marshal rejection response: %w", err)
	}

	subject := enrollSubject(w.namespace, pki.EnrollResponsePrefix+"."+machineID)
	if err := w.nc.PublishCore(ctx, subject, data); err != nil {
		return fmt.Errorf("publish rejection for %s: %w", machineID, err)
	}

	if err := w.enrollmentKV.KVDelete(ctx, w.enrollmentBucket, key); err != nil {
		return fmt.Errorf("delete pending agent %s: %w", machineID, err)
	}

	// A rejected agent must not keep a stored key from an earlier
	// acceptance, so nothing it signs verifies afterwards.
	if err := w.RemoveAgentKey(ctx, machineID); err != nil {
		return err
	}

	w.logger.Info(
		"rejected agent enrollment",
		slog.String("machine_id", machineID),
		slog.String("hostname", pending.Hostname),
		slog.String("reason", reason),
	)

	return nil
}

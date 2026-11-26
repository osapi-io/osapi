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
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/attribute"

	"github.com/osapi-io/osapi/internal/audit"
	"github.com/osapi-io/osapi/internal/pki"
	"github.com/osapi-io/osapi/internal/telemetry/opspan"
)

// ServeRotate answers each rotation of an agent's keys until ctx ends: an
// admitted machine whose current key, and whose new key, signed the request
// gets its admission moved to the new keys. The old signing key stays valid
// for the rotation grace period, as at a new admission, so a result that the
// agent signed just before it took the new keys is not refused.
func (w *Watcher) ServeRotate(
	ctx context.Context,
	listen JoinListener,
) error {
	subject := enrollSubject(w.namespace, pki.RotateRequestSuffix)
	sub, err := listen(subject, joinQueue, func(msg *nats.Msg) {
		ctx, span := opspan.Root(ctx, tracerName, "join.rotate")
		defer span.End()
		reply := w.rotateReply(ctx, msg.Data)
		if reply.Error != "" {
			span.SetAttributes(attribute.String("osapi.rotate.refused", reply.Error))
		}
		data, _ := json.Marshal(reply)
		if err := respond(msg, data); err != nil {
			w.logger.Warn("failed to reply to a rotation", slog.String("error", err.Error()))
		}
	})
	if err != nil {
		return fmt.Errorf("subscribe to rotations: %w", err)
	}
	w.logger.Info("key rotation over NATS started", slog.String("subject", subject))

	<-ctx.Done()
	_ = sub.Unsubscribe()

	return nil
}

// rotateReply is the reply to the rotation in data, which is audited, also
// when the controller refuses it.
func (w *Watcher) rotateReply(
	ctx context.Context,
	data []byte,
) pki.RotateReply {
	var req pki.RotateRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return pki.RotateReply{Error: "the rotation is not JSON"}
	}
	reply, code := w.rotateChecked(ctx, req)
	summary := "rotated its keys to " + reply.Fingerprint
	if reply.Error != "" {
		summary = "refused: " + reply.Error
	}
	w.record(ctx, audit.Change(ctx, audit.ActorAgent, req.MachineID, "agent.key.rotate",
		req.MachineID, code, summary))

	return reply
}

// rotateChecked is the reply to req, and its outcome as an HTTP status for
// the audit entry.
func (w *Watcher) rotateChecked(
	ctx context.Context,
	req pki.RotateRequest,
) (pki.RotateReply, int) {
	if d := nowFn().Sub(time.Unix(req.At, 0)); d > renewSkew || d < -renewSkew {
		return pki.RotateReply{Error: "the time of the rotation is not now"}, http.StatusBadRequest
	}
	if !pki.ValidEncryptionKey(req.EncryptionKey) {
		return pki.RotateReply{
			Error: "the new encryption key is not a P-256 key",
		}, http.StatusBadRequest
	}
	record, err := w.LookupAgentKey(ctx, req.MachineID)
	if err != nil {
		w.logger.Warn("refused a rotation of a machine that is not admitted",
			slog.String("machine_id", req.MachineID))

		return pki.RotateReply{Error: "the machine is not admitted"}, http.StatusForbidden
	}
	// The admission holds the new keys already: an earlier request moved it,
	// and the agent did not get the reply. The new key's signature is the
	// proof, since the old key no longer matches the admission.
	if bytes.Equal(record.PublicKey, req.PublicKey) && record.EncryptionKey == req.EncryptionKey {
		if !req.VerifyNew() {
			return pki.RotateReply{
				Error: "the rotation is not signed by the new key",
			}, http.StatusForbidden
		}

		return pki.RotateReply{Fingerprint: record.Fingerprint}, http.StatusOK
	}
	if !req.Verify(record.PublicKey) {
		return pki.RotateReply{
			Error: "the rotation is not signed by the admitted key and the new key",
		}, http.StatusForbidden
	}
	rotated := *record
	rotated.PublicKey = req.PublicKey
	rotated.EncryptionKey = req.EncryptionKey
	rotated.Fingerprint = pki.FingerprintOfKeys(req.PublicKey, req.EncryptionKey)
	rotated.SupersededKey, rotated.SupersededUntil = nil, time.Time{}
	if w.rotationGrace > 0 {
		rotated.SupersededKey = record.PublicKey
		rotated.SupersededUntil = nowFn().Add(w.rotationGrace)
	}
	data, err := marshalFn(rotated)
	if err != nil {
		return pki.RotateReply{
			Error: "the controller could not rotate: " + err.Error(),
		}, http.StatusInternalServerError
	}
	if _, err := w.enrollmentKV.KVPut(
		ctx,
		w.enrollmentBucket,
		acceptedKey(req.MachineID),
		data,
	); err != nil {
		return pki.RotateReply{
			Error: "the controller could not rotate: " + err.Error(),
		}, http.StatusInternalServerError
	}
	w.invalidateAgentKey(req.MachineID)

	return pki.RotateReply{Fingerprint: rotated.Fingerprint}, http.StatusOK
}

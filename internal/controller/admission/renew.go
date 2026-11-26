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
	"net/http"
	"time"

	"github.com/osapi-io/osapi/internal/audit"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/attribute"

	"github.com/osapi-io/osapi/internal/pki"
	"github.com/osapi-io/osapi/internal/telemetry/opspan"
)

// renewSkew is how far the time of a renewal can be from the controller's.
const renewSkew = 5 * time.Minute

// Mint is a JWT for the NATS user key of an admitted machine.
type Mint func(userKey string, machineID string) (string, error)

// ServeRenew answers each renewal of an agent's NATS user until ctx ends:
// an admitted machine whose key signed the request gets a new JWT from
// mint. A machine that is not admitted gets none, so a revoked agent is cut
// from NATS when its JWT expires.
func (w *Watcher) ServeRenew(
	ctx context.Context,
	listen JoinListener,
	mint Mint,
) error {
	subject := enrollSubject(w.namespace, pki.RenewRequestSuffix)
	sub, err := listen(subject, joinQueue, func(msg *nats.Msg) {
		ctx, span := opspan.Root(ctx, tracerName, "join.renew")
		defer span.End()
		reply := w.renewReply(ctx, msg.Data, mint)
		if reply.Error != "" {
			span.SetAttributes(attribute.String("osapi.renew.refused", reply.Error))
		}
		data, _ := json.Marshal(reply)
		if err := respond(msg, data); err != nil {
			w.logger.Warn("failed to reply to a renewal", slog.String("error", err.Error()))
		}
	})
	if err != nil {
		return fmt.Errorf("subscribe to renewals: %w", err)
	}
	w.logger.Info("renewal over NATS started", slog.String("subject", subject))

	<-ctx.Done()
	_ = sub.Unsubscribe()

	return nil
}

// renewReply is the reply to the renewal in data: a new JWT, or why not.
func (w *Watcher) renewReply(
	ctx context.Context,
	data []byte,
	mint Mint,
) pki.JoinReply {
	var req pki.RenewRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return pki.JoinReply{Error: "the renewal is not JSON"}
	}
	reply, code := w.renewChecked(ctx, req, mint)
	// Each renewal is a new NATS user for a machine, so each is audited:
	// a refusal too, since one that a machine keeps asking for is a sign.
	summary := "renewed its NATS user"
	if reply.Error != "" {
		summary = "refused: " + reply.Error
	}
	w.record(ctx, audit.Change(ctx, audit.ActorAgent, req.MachineID, "agent.nats.renew",
		req.MachineID, code, summary))

	return reply
}

// renewChecked is the reply to req, and its outcome as an HTTP status for
// the audit entry.
func (w *Watcher) renewChecked(
	ctx context.Context,
	req pki.RenewRequest,
	mint Mint,
) (pki.JoinReply, int) {
	if d := nowFn().Sub(time.Unix(req.At, 0)); d > renewSkew || d < -renewSkew {
		return pki.JoinReply{Error: "the time of the renewal is not now"}, http.StatusBadRequest
	}
	record, err := w.LookupAgentKey(ctx, req.MachineID)
	if err != nil {
		w.logger.Warn("refused a renewal of a machine that is not admitted",
			slog.String("machine_id", req.MachineID))

		return pki.JoinReply{Error: "the machine is not admitted"}, http.StatusForbidden
	}
	if !req.Verify(record.PublicKey) {
		return pki.JoinReply{
			Error: "the renewal is not signed by the admitted key",
		}, http.StatusForbidden
	}
	token, err := mint(req.NATSUserKey, req.MachineID)
	if err != nil {
		return pki.JoinReply{
			Error: "the controller could not renew: " + err.Error(),
		}, http.StatusInternalServerError
	}

	return pki.JoinReply{NATS: pki.JoinNATS{JWT: token}}, http.StatusOK
}

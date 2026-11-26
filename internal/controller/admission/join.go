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
	"crypto/ed25519"
	"errors"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/attribute"

	"github.com/osapi-io/osapi/internal/controller/jointoken"
	"github.com/osapi-io/osapi/internal/pki"
	"github.com/osapi-io/osapi/internal/telemetry/opspan"
	"github.com/osapi-io/osapi/pkg/sdk/types"
)

// ErrJoinSignature is a join request whose signature is not by the key that
// it asks the controller to admit.
var ErrJoinSignature = errors.New("the join request is not signed by its key")

// ErrNoJoinTokens is a controller with no store of join tokens.
var ErrNoJoinTokens = errors.New("this controller takes no join")

// JoinTokens takes a use of a join token: the store of join tokens.
type JoinTokens interface {
	Consume(ctx context.Context, token string) (jointoken.Token, error)
}

// SetJoinTokens gives the watcher the store of join tokens, which Join takes
// a use of. Called once at startup.
func (w *Watcher) SetJoinTokens(
	tokens JoinTokens,
) {
	w.joinTokens = tokens
}

// Join admits the machine of req at once, when req's token admits, in an
// admission.join span: the operator made the token, so the token is the
// decision that admit records. The admitted record names the token. It
// returns the controller's public key, which the agent keeps. A request that
// its key did not sign is ErrJoinSignature; a token that does not admit is
// one of jointoken's errors, and admits nothing.
func (w *Watcher) Join(
	ctx context.Context,
	req pki.JoinRequest,
) (ed25519.PublicKey, error) {
	key, _, err := w.joinLabels(ctx, req)

	return key, err
}

// joinLabels is Join, with the labels of the token that admitted the
// machine, which the agent takes.
func (w *Watcher) joinLabels(
	ctx context.Context,
	req pki.JoinRequest,
) (ed25519.PublicKey, map[string]string, error) {
	ctx, span := opspan.Start(ctx, tracerName, "admission.join",
		attribute.String("osapi.agent.machine_id", req.MachineID),
	)
	key, labels, err := w.join(ctx, req)

	return key, labels, opspan.End(span, err)
}

// join checks the signature, takes a use of the token, and admits.
func (w *Watcher) join(
	ctx context.Context,
	req pki.JoinRequest,
) (ed25519.PublicKey, map[string]string, error) {
	if w.joinTokens == nil {
		return nil, nil, ErrNoJoinTokens
	}
	if !req.Verify() {
		return nil, nil, ErrJoinSignature
	}
	// A join that admitted the machine, whose reply did not reach the
	// agent, is answered again: the same machine, key and token take no
	// second use of the token.
	if labels, ok := w.joinedBefore(ctx, req); ok {
		w.logger.Info(
			"answered a join again",
			slog.String("machine_id", req.MachineID),
			slog.String("hostname", req.Hostname),
		)

		return w.pkiProvider.PublicKey(), labels, nil
	}
	t, err := w.joinTokens.Consume(ctx, req.Token)
	if err != nil {
		return nil, nil, err
	}
	pending := PendingAgent{
		MachineID:     req.MachineID,
		Hostname:      req.Hostname,
		PublicKey:     req.PublicKey,
		EncryptionKey: req.EncryptionKey,
		Fingerprint:   pki.FingerprintOfKeys(req.PublicKey, req.EncryptionKey),
		RequestedAt:   nowFn(),
		JoinLabels:    t.Labels,
		NATSUserKey:   req.NATSUserKey,
	}
	if err := w.recordAgentKey(ctx, pending, t.ID); err != nil {
		return nil, nil, fmt.Errorf("admit %s: %w", req.MachineID, err)
	}
	w.indexNATSUserKey(ctx, req.NATSUserKey, req.MachineID)
	// A join decides as admit does: an earlier request, rejection or
	// revocation of the machine is cleared.
	for _, prefix := range []string{kvPrefix, rejectedKVPrefix, revokedKVPrefix} {
		_ = w.enrollmentKV.KVDelete(ctx, w.enrollmentBucket, prefix+req.MachineID)
	}
	w.event(ctx, req.MachineID, req.Hostname, "admitted", "admitted by join token "+t.ID)
	w.logger.Info(
		"admitted agent by join token",
		slog.String("machine_id", req.MachineID),
		slog.String("hostname", req.Hostname),
		slog.String("fingerprint", pending.Fingerprint),
		slog.String("join_token", t.ID),
	)

	return w.pkiProvider.PublicKey(), t.Labels, nil
}

// joinedBefore reports whether the token of req already admitted its
// machine with its keys, and the labels that the token gave it.
func (w *Watcher) joinedBefore(
	ctx context.Context,
	req pki.JoinRequest,
) (map[string]string, bool) {
	id, ok := jointoken.ID(req.Token)
	if !ok {
		return nil, false
	}
	r, err := w.Get(ctx, req.MachineID)
	if err != nil || r.State != types.AdmissionAdmitted || r.JoinToken != id ||
		r.Fingerprint != pki.FingerprintOfKeys(req.PublicKey, req.EncryptionKey) {
		return nil, false
	}

	return r.JoinLabels, true
}

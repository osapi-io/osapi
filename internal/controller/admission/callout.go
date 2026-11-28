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
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/nats-io/jwt/v2"
	"go.opentelemetry.io/otel/attribute"

	"github.com/osapi-io/osapi/internal/audit"
	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/controller/jointoken"
	"github.com/osapi-io/osapi/internal/natscallout"
	"github.com/osapi-io/osapi/internal/telemetry/opspan"
)

// natsUserKVPrefix keys the machine ID of each NATS user key that a join
// gave, so the auth callout finds the machine of a connection at once.
const natsUserKVPrefix = "natsuser."

// joinSessionTTL is the longest that a connection of a join token lasts:
// long enough for the join request and its reply.
const joinSessionTTL = 10 * time.Minute

// TokenChecker checks a join token without taking a use of it.
type TokenChecker interface {
	Check(ctx context.Context, token string) (jointoken.Token, error)
}

// CalloutRoles are what the auth callout gives each kind of connection.
type CalloutRoles struct {
	// Join is the permissions of a connection with a join token.
	Join bus.Permissions
	// Agent is the permissions of an admitted agent.
	Agent bus.Permissions
	// AgentTTL is how long an agent's connection lasts before the server
	// asks the callout again, or zero for as long as it stays connected.
	AgentTTL time.Duration
	// Tokens checks a join token.
	Tokens TokenChecker
}

// indexNATSUserKey records that key is the NATS user of machineID. A join
// with no key, such as one with static users, records nothing. The admitted
// record has the key too, and the callout checks it, so an entry that a
// later join replaced admits no one.
func (w *Watcher) indexNATSUserKey(
	ctx context.Context,
	key string,
	machineID string,
) {
	if key == "" {
		return
	}
	if _, err := w.enrollmentKV.KVPut(
		ctx, w.enrollmentBucket, natsUserKVPrefix+key, []byte(machineID),
	); err != nil {
		w.logger.Warn("failed to record the NATS user of a machine",
			slog.String("machine_id", machineID),
			slog.String("error", err.Error()),
		)
	}
}

// Callout is how the auth callout decides a connection: a join token that
// admits now gets the join role until it expires, at most joinSessionTTL; an
// admitted machine's own NATS user key, with the server's nonce signed with
// it, gets the agent role. Anything else is refused. Each decision is in an
// admission.callout span and the audit log.
func (w *Watcher) Callout(
	r CalloutRoles,
) natscallout.Decide {
	return func(ctx context.Context, req *jwt.AuthorizationRequest) (natscallout.Decision, error) {
		ctx, span := opspan.Start(ctx, tracerName, "admission.callout")
		d, actor, name, err := w.callout(ctx, r, req)
		span.SetAttributes(attribute.String("osapi.callout.actor", name))
		code, summary := http.StatusOK, "connected as "+d.Name
		if err != nil {
			code, summary = http.StatusForbidden, err.Error()
		}
		w.record(ctx, audit.Change(ctx, actor, name, "nats.connect", name, code, summary))

		return d, opspan.End(span, err)
	}
}

// callout is the decision of Callout, and who asked: the actor and name of
// its audit entry.
func (w *Watcher) callout(
	ctx context.Context,
	r CalloutRoles,
	req *jwt.AuthorizationRequest,
) (natscallout.Decision, audit.Actor, string, error) {
	o := req.ConnectOptions
	switch {
	case strings.HasPrefix(o.Token, "osapi_jt_"):
		id, _ := jointoken.ID(o.Token)
		t, err := r.Tokens.Check(ctx, o.Token)
		if errors.Is(err, jointoken.ErrInvalid) || errors.Is(err, jointoken.ErrExpired) ||
			errors.Is(err, jointoken.ErrUsed) {
			return natscallout.Decision{}, audit.ActorJoinToken, id,
				fmt.Errorf("%w: the join token does not admit: %w", natscallout.ErrRefused, err)
		}
		// A token that could not be read, such as when the bucket is down,
		// is not a refusal: the connection is undecided, and the cause goes
		// to the log, not to the NATS server.
		if err != nil {
			return natscallout.Decision{}, audit.ActorJoinToken, id, err
		}
		expires := nowFn().Add(joinSessionTTL)
		if t.ExpiresAt.Before(expires) {
			expires = t.ExpiresAt
		}

		return natscallout.Decision{
			Name: "join:" + t.ID, Permissions: r.Join, Expires: expires,
		}, audit.ActorJoinToken, t.ID, nil
	case o.Nkey != "":
		machineID, err := w.calloutAgent(ctx, req)
		if err != nil {
			return natscallout.Decision{}, audit.ActorAgent, o.Nkey, err
		}
		d := natscallout.Decision{Name: "agent:" + machineID, Permissions: r.Agent}
		if r.AgentTTL > 0 {
			d.Expires = nowFn().Add(r.AgentTTL)
		}

		return d, audit.ActorAgent, machineID, nil
	}

	return natscallout.Decision{}, audit.ActorAgent, "",
		fmt.Errorf("%w: no join token or NATS user key", natscallout.ErrRefused)
}

// calloutAgent is the admitted machine whose NATS user is the nkey of req,
// when the connection signed the server's nonce with it.
func (w *Watcher) calloutAgent(
	ctx context.Context,
	req *jwt.AuthorizationRequest,
) (string, error) {
	if err := natscallout.VerifyNkey(req); err != nil {
		return "", err
	}
	key := req.ConnectOptions.Nkey
	entry, err := w.enrollmentKV.KVGet(ctx, w.enrollmentBucket, natsUserKVPrefix+key)
	if err != nil {
		return "", fmt.Errorf(
			"%w: no machine joined with this NATS user key",
			natscallout.ErrRefused,
		)
	}
	machineID := string(entry.Value())
	record, err := w.LookupAgentKey(ctx, machineID)
	if err != nil {
		if errors.Is(err, ErrAgentKeyNotFound) {
			return "", fmt.Errorf(
				"%w: the machine %s is not admitted",
				natscallout.ErrRefused,
				machineID,
			)
		}

		return "", err
	}
	if record.NATSUserKey != key {
		return "", fmt.Errorf("%w: the machine %s joined again with another key",
			natscallout.ErrRefused, machineID)
	}

	return machineID, nil
}

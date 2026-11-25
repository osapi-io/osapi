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

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/trace"

	"github.com/osapi-io/osapi/internal/audit"
	"github.com/osapi-io/osapi/internal/controller/jointoken"
	"github.com/osapi-io/osapi/internal/pki"
)

// joinQueue is the queue group of the controllers' join subscriptions, so
// that one controller answers each join and takes one use of its token.
const joinQueue = "osapi-controllers"

// JoinListener subscribes handler to subject in queue: the QueueSubscribe of
// the controller's NATS connection.
type JoinListener func(
	subject string,
	queue string,
	handler nats.MsgHandler,
) (*nats.Subscription, error)

// respond sends data as the reply to msg. Tests swap it.
var respond = func(
	msg *nats.Msg,
	data []byte,
) error {
	return msg.Respond(data)
}

// AgentUser is the NATS user that a join gives the agent of req: one user
// for each agent, or in operator mode a JWT for the agent's own user key.
type AgentUser func(req pki.JoinRequest) (pki.JoinNATS, error)

// StaticUser is the AgentUser that gives each agent user.
func StaticUser(
	user pki.JoinNATS,
) AgentUser {
	return func(pki.JoinRequest) (pki.JoinNATS, error) { return user, nil }
}

// ServeJoin answers each join over NATS until ctx ends: it admits the
// machine as Join does, and replies with the controller's key and the NATS
// user that agent gives, which the agent connects with after the join.
func (w *Watcher) ServeJoin(
	ctx context.Context,
	listen JoinListener,
	agent AgentUser,
) error {
	subject := enrollSubject(w.namespace, pki.JoinRequestSuffix)
	sub, err := listen(subject, joinQueue, func(msg *nats.Msg) {
		w.respondJoin(ctx, msg, agent)
	})
	if err != nil {
		return fmt.Errorf("subscribe to join requests: %w", err)
	}
	w.logger.Info("join over NATS started", slog.String("subject", subject))

	<-ctx.Done()
	_ = sub.Unsubscribe()

	return nil
}

// JoinAudit is where a join over NATS is audited: the audit log of the
// controller. A *audit.StreamStore is one.
type JoinAudit interface {
	Write(ctx context.Context, entry audit.Entry) error
}

// SetJoinAudit gives the watcher the audit log, which each join over NATS
// writes an entry to. Without it, a join is only logged.
func (w *Watcher) SetJoinAudit(
	a JoinAudit,
) {
	w.joinAudit = a
}

// respondJoin replies to the join request of msg, and audits it.
func (w *Watcher) respondJoin(
	ctx context.Context,
	msg *nats.Msg,
	agent AgentUser,
) {
	start := nowFn()
	req, reply := w.joinReply(ctx, msg.Data, agent)
	data, _ := json.Marshal(reply)
	if err := respond(msg, data); err != nil {
		w.logger.Warn("failed to reply to a join", slog.String("error", err.Error()))
	}
	w.auditJoin(ctx, msg.Subject, req, reply, start)
}

// auditJoin writes the audit entry of a join over NATS: the token that
// admitted, or tried to, is the actor, and the machine is what it asked for.
// It holds no secret: the token's ID, never the token.
func (w *Watcher) auditJoin(
	ctx context.Context,
	subject string,
	req pki.JoinRequest,
	reply pki.JoinReply,
	start time.Time,
) {
	if w.joinAudit == nil {
		return
	}
	actor := "join-token"
	if id, ok := jointoken.ID(req.Token); ok {
		actor += ":" + id
	}
	code := http.StatusOK
	if reply.Error != "" {
		code = http.StatusForbidden
	}
	entry := audit.Entry{
		ID:           uuid.New().String(),
		Timestamp:    start,
		User:         actor,
		Method:       "NATS",
		Path:         subject,
		OperationID:  "agent.join",
		ResponseCode: code,
		DurationMs:   nowFn().Sub(start).Milliseconds(),
		RequestSummary: fmt.Sprintf("machine_id=%s hostname=%s",
			req.MachineID, req.Hostname),
	}
	if sc := trace.SpanContextFromContext(ctx); sc.HasTraceID() {
		entry.TraceID = sc.TraceID().String()
	}
	if err := w.joinAudit.Write(context.WithoutCancel(ctx), entry); err != nil {
		w.logger.Warn("failed to audit a join", slog.String("error", err.Error()))
	}
}

// joinReply is the request in data, and its reply: the controller's key and
// agent when the token admits, or why it did not.
func (w *Watcher) joinReply(
	ctx context.Context,
	data []byte,
	agent AgentUser,
) (pki.JoinRequest, pki.JoinReply) {
	var req pki.JoinRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return req, pki.JoinReply{Error: "the join request is not JSON"}
	}
	key, labels, err := w.joinLabels(ctx, req)
	if err != nil {
		w.logger.Warn(
			"refused a join over NATS",
			slog.String("machine_id", req.MachineID),
			slog.String("hostname", req.Hostname),
			slog.String("error", err.Error()),
		)

		return req, pki.JoinReply{Error: err.Error()}
	}

	user, err := agent(req)
	if err != nil {
		w.logger.Warn(
			"admitted a machine, but could not give it a NATS user",
			slog.String("machine_id", req.MachineID),
			slog.String("error", err.Error()),
		)

		return req, pki.JoinReply{
			Error: "the controller could not give a NATS user: " + err.Error(),
		}
	}

	return req, pki.JoinReply{ControllerPublicKey: key, NATS: user, Labels: labels}
}

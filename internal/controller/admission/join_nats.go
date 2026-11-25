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

	"github.com/nats-io/nats.go"

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

// ServeJoin answers each join over NATS until ctx ends: it admits the
// machine as Join does, and replies with the controller's key and agent, the
// NATS user that the agent connects with after the join.
func (w *Watcher) ServeJoin(
	ctx context.Context,
	listen JoinListener,
	agent pki.JoinNATS,
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

// respondJoin replies to the join request of msg.
func (w *Watcher) respondJoin(
	ctx context.Context,
	msg *nats.Msg,
	agent pki.JoinNATS,
) {
	data, _ := json.Marshal(w.joinReply(ctx, msg.Data, agent))
	if err := respond(msg, data); err != nil {
		w.logger.Warn("failed to reply to a join", slog.String("error", err.Error()))
	}
}

// joinReply is the reply to the join request data: the controller's key and
// agent when the token admits, or why it did not.
func (w *Watcher) joinReply(
	ctx context.Context,
	data []byte,
	agent pki.JoinNATS,
) pki.JoinReply {
	var req pki.JoinRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return pki.JoinReply{Error: "the join request is not JSON"}
	}
	key, err := w.Join(ctx, req)
	if err != nil {
		w.logger.Warn(
			"refused a join over NATS",
			slog.String("machine_id", req.MachineID),
			slog.String("hostname", req.Hostname),
			slog.String("error", err.Error()),
		)

		return pki.JoinReply{Error: err.Error()}
	}

	return pki.JoinReply{ControllerPublicKey: key, NATS: agent}
}

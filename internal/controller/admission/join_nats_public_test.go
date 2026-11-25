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

package admission_test

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/osapi-io/osapi/internal/audit"
	"go.opentelemetry.io/otel/trace"

	"github.com/nats-io/nats.go"

	"github.com/osapi-io/osapi/internal/controller/admission"
	"github.com/osapi-io/osapi/internal/controller/jointoken"
	"github.com/osapi-io/osapi/internal/pki"
)

// joinLog keeps the audit entries of the joins.
type joinLog struct {
	entries []audit.Entry
	err     error
}

func (l *joinLog) Write(
	_ context.Context,
	e audit.Entry,
) error {
	l.entries = append(l.entries, e)

	return l.err
}

// serveJoin runs ServeJoin with a listener that keeps the handler, sends
// data to it as a join request, and returns the reply.
func (s *StorePublicTestSuite) serveJoin(
	data []byte,
	respondErr error,
) pki.JoinReply {
	var (
		handler nats.MsgHandler
		subject string
		queue   string
		got     []byte
	)
	ready := make(chan struct{})
	listen := func(sub string, q string, h nats.MsgHandler) (*nats.Subscription, error) {
		subject, queue, handler = sub, q, h
		close(ready)

		return &nats.Subscription{}, nil
	}
	defer admission.SetRespond(func(_ *nats.Msg, d []byte) error {
		got = d

		return respondErr
	})()

	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{1},
		SpanID:  trace.SpanID{1},
	})
	ctx, cancel := context.WithCancel(trace.ContextWithSpanContext(context.Background(), sc))
	done := make(chan error, 1)
	go func() {
		done <- s.watcher.ServeJoin(ctx, listen, admission.StaticUser(pki.JoinNATS{Username: "agent", Password: "pw"}))
	}()
	<-ready
	handler(&nats.Msg{Subject: "osapi.join.request", Data: data})
	cancel()
	s.Require().NoError(<-done)
	s.Equal("osapi.join.request", subject)
	s.Equal("osapi-controllers", queue)

	var reply pki.JoinReply
	s.Require().NoError(json.Unmarshal(got, &reply))

	return reply
}

func (s *StorePublicTestSuite) TestServeJoin() {
	s.Run("a token that admits gets the controller's key and the agent user", func() {
		log := &joinLog{}
		s.watcher.SetAudit(log)
		defer s.watcher.SetAudit(nil)
		s.watcher.SetJoinTokens(fakeTokens{})
		data, _ := json.Marshal(joinRequest("osapi_jt_t1.secret"))

		reply := s.serveJoin(data, nil)

		s.Empty(reply.Error)
		s.Len(reply.ControllerPublicKey, 32)
		s.Equal(pki.JoinNATS{Username: "agent", Password: "pw"}, reply.NATS)
		s.Equal(
			map[string]string{"group": "web"},
			reply.Labels,
			"the agent gets the token's labels",
		)
		s.Require().Len(log.entries, 1)
		e := log.entries[0]
		s.Equal("t1", e.User, "the token's ID, never the token")
		s.Equal(audit.ActorJoinToken, e.Actor)
		s.Equal("agent.join", e.Action)
		s.Equal("m-1", e.Target)
		s.Equal("NATS", e.Method)
		s.Equal("osapi.join.request", e.Path)
		s.Equal("agent.join", e.OperationID)
		s.Equal(200, e.ResponseCode)
		s.Equal("machine_id=m-1 hostname=web-01", e.RequestSummary)
		s.NotContains(e.RequestSummary+e.User, "secret")
		s.Equal(trace.TraceID{1}.String(), e.TraceID)
	})

	s.Run("a token that does not admit gets why, and no user", func() {
		log := &joinLog{err: errors.New("audit down")}
		s.watcher.SetAudit(log)
		defer s.watcher.SetAudit(nil)
		s.watcher.SetJoinTokens(fakeTokens{err: jointoken.ErrExpired})
		data, _ := json.Marshal(joinRequest("osapi_jt_t1.secret"))

		reply := s.serveJoin(data, errors.New("no reply subject"))

		s.Equal(jointoken.ErrExpired.Error(), reply.Error)
		s.Require().Len(log.entries, 1)
		s.Equal(403, log.entries[0].ResponseCode)
		s.Empty(reply.NATS.Password)
		s.Empty(reply.ControllerPublicKey)
	})

	s.Run("a machine that is admitted but gets no user is told why", func() {
		s.watcher.SetJoinTokens(fakeTokens{})
		data, _ := json.Marshal(joinRequest("osapi_jt_t1.secret"))
		var handler nats.MsgHandler
		ready := make(chan struct{})
		var got []byte
		defer admission.SetRespond(func(_ *nats.Msg, d []byte) error {
			got = d

			return nil
		})()
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error, 1)
		go func() {
			done <- s.watcher.ServeJoin(ctx,
				func(_ string, _ string, h nats.MsgHandler) (*nats.Subscription, error) {
					handler = h
					close(ready)

					return &nats.Subscription{}, nil
				},
				func(pki.JoinRequest) (pki.JoinNATS, error) {
					return pki.JoinNATS{}, errors.New("no user key")
				})
		}()
		<-ready
		handler(&nats.Msg{Data: data})
		cancel()
		s.Require().NoError(<-done)
		var reply pki.JoinReply
		s.Require().NoError(json.Unmarshal(got, &reply))
		s.Equal("the controller could not give a NATS user: no user key", reply.Error)
		s.Empty(reply.ControllerPublicKey)
	})

	s.Run("a request that is not JSON gets why", func() {
		log := &joinLog{}
		s.watcher.SetAudit(log)
		defer s.watcher.SetAudit(nil)
		reply := s.serveJoin([]byte("{"), nil)
		s.Require().Len(log.entries, 1)
		s.Equal(audit.ActorJoinToken, log.entries[0].Actor)
		s.Empty(log.entries[0].User, "a request that is not JSON names no token")

		s.Equal("the join request is not JSON", reply.Error)
	})

	s.Run("a message with no reply subject is logged, not answered", func() {
		var handler nats.MsgHandler
		ready := make(chan struct{})
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error, 1)
		go func() {
			done <- s.watcher.ServeJoin(ctx, func(_ string, _ string, h nats.MsgHandler) (*nats.Subscription, error) {
				handler = h
				close(ready)

				return &nats.Subscription{}, nil
			}, admission.StaticUser(pki.JoinNATS{}))
		}()
		<-ready
		s.NotPanics(func() { handler(&nats.Msg{Data: []byte("{")}) })
		cancel()
		s.NoError(<-done)
	})

	s.Run("a subscription that fails is the error", func() {
		err := s.watcher.ServeJoin(
			context.Background(),
			func(string, string, nats.MsgHandler) (*nats.Subscription, error) {
				return nil, errors.New("no permission")
			},
			admission.StaticUser(pki.JoinNATS{}),
		)

		s.ErrorContains(err, "subscribe to join requests: no permission")
	})
}

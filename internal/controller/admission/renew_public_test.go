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
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"time"

	"github.com/osapi-io/osapi/internal/audit"

	"github.com/nats-io/nats.go"

	"github.com/osapi-io/osapi/internal/controller/admission"
	"github.com/osapi-io/osapi/internal/pki"
)

// renew sends data to ServeRenew with mint, and returns the reply.
func (s *StorePublicTestSuite) renew(
	data []byte,
	mint admission.Mint,
) pki.JoinReply {
	var (
		handler nats.MsgHandler
		got     []byte
	)
	ready := make(chan struct{})
	defer admission.SetRespond(func(_ *nats.Msg, d []byte) error {
		got = d

		return errors.New("no reply subject")
	})()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- s.watcher.ServeRenew(ctx, func(subject string, queue string, h nats.MsgHandler) (*nats.Subscription, error) {
			s.Equal("osapi.join.renew", subject)
			s.Equal("osapi-controllers", queue)
			handler = h
			close(ready)

			return &nats.Subscription{}, nil
		}, mint)
	}()
	<-ready
	handler(&nats.Msg{Data: data})
	cancel()
	s.Require().NoError(<-done)
	var reply pki.JoinReply
	s.Require().NoError(json.Unmarshal(got, &reply))

	return reply
}

func (s *StorePublicTestSuite) TestRenew() {
	now := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)
	admission.SetNowFn(func() time.Time { return now })
	defer admission.ResetNowFn()
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	s.put(admission.AcceptedKVPrefix()+"m-1", admission.AcceptedAgent{
		MachineID: "m-1", Hostname: "web-01", PublicKey: pub, Fingerprint: s.fp, AcceptedAt: now,
	})
	request := func(machineID string, at time.Time, key ed25519.PrivateKey) []byte {
		r := pki.RenewRequest{MachineID: machineID, NATSUserKey: "UA1", At: at.Unix()}
		r.Sign(key)
		data, _ := json.Marshal(r)

		return data
	}
	mint := func(userKey string, machineID string) (string, error) {
		return "jwt-for-" + userKey + "-" + machineID, nil
	}
	_, other, _ := ed25519.GenerateKey(rand.Reader)

	tests := []struct {
		name    string
		data    []byte
		mint    admission.Mint
		wantJWT string
		wantErr string
	}{
		{
			name:    "an admitted machine gets a new JWT",
			data:    request("m-1", now, priv),
			mint:    mint,
			wantJWT: "jwt-for-UA1-m-1",
		},
		{
			name:    "a request that is not JSON",
			data:    []byte("{"),
			mint:    mint,
			wantErr: "the renewal is not JSON",
		},
		{
			name:    "a request from long ago",
			data:    request("m-1", now.Add(-time.Hour), priv),
			mint:    mint,
			wantErr: "the time of the renewal is not now",
		},
		{
			name:    "a request from the future",
			data:    request("m-1", now.Add(time.Hour), priv),
			mint:    mint,
			wantErr: "the time of the renewal is not now",
		},
		{
			name:    "a machine that is not admitted",
			data:    request("m-x", now, priv),
			mint:    mint,
			wantErr: "the machine is not admitted",
		},
		{
			name:    "a request that another key signed",
			data:    request("m-1", now, other),
			mint:    mint,
			wantErr: "not signed by the admitted key",
		},
		{
			name:    "a JWT that cannot be minted",
			data:    request("m-1", now, priv),
			mint:    func(string, string) (string, error) { return "", errors.New("no key") },
			wantErr: "the controller could not renew: no key",
		},
	}
	log := &joinLog{}
	s.watcher.SetAudit(log)
	defer s.watcher.SetAudit(nil)
	for _, tt := range tests {
		reply := s.renew(tt.data, tt.mint)
		s.Equal(tt.wantJWT, reply.NATS.JWT, tt.name)
		s.Contains(reply.Error, tt.wantErr, tt.name)
	}

	// Each renewal that names a machine is audited, as the machine, with its
	// outcome: a refusal too. A request that is not JSON names none.
	codes := make([]int, 0, len(log.entries))
	for _, e := range log.entries {
		s.Equal(audit.ActorAgent, e.Actor)
		s.Equal("agent.nats.renew", e.Action)
		codes = append(codes, e.ResponseCode)
	}
	s.Equal([]int{200, 400, 400, 403, 403, 500}, codes)
	s.Equal("m-1", log.entries[0].Target)
	s.Equal("renewed its NATS user", log.entries[0].RequestSummary)

	err := s.watcher.ServeRenew(
		context.Background(),
		func(string, string, nats.MsgHandler) (*nats.Subscription, error) {
			return nil, errors.New("no permission")
		},
		mint,
	)
	s.ErrorContains(err, "subscribe to renewals: no permission")
}

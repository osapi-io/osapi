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
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/osapi-io/osapi/internal/audit"
	"github.com/osapi-io/osapi/internal/controller/admission"
	"github.com/osapi-io/osapi/internal/pki"
)

// rotate sends data to ServeRotate, and returns the reply.
func (s *StorePublicTestSuite) rotate(
	data []byte,
) pki.RotateReply {
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
		done <- s.watcher.ServeRotate(ctx, func(subject string, queue string, h nats.MsgHandler) (*nats.Subscription, error) {
			s.Equal("osapi.join.rotate", subject)
			s.Equal("osapi-controllers", queue)
			handler = h
			close(ready)

			return &nats.Subscription{}, nil
		})
	}()
	<-ready
	handler(&nats.Msg{Data: data})
	cancel()
	s.Require().NoError(<-done)
	var reply pki.RotateReply
	s.Require().NoError(json.Unmarshal(got, &reply))

	return reply
}

// encryptionKey is a new public P-256 key in pki.EncryptionKeyEncoding.
func encryptionKey() string {
	k, _ := ecdh.P256().GenerateKey(rand.Reader)

	return pki.EncodeEncryptionKey(k.PublicKey())
}

func (s *StorePublicTestSuite) TestRotate() {
	oldPub, oldPriv, _ := ed25519.GenerateKey(rand.Reader)
	newPub, newPriv, _ := ed25519.GenerateKey(rand.Reader)
	_, otherPriv, _ := ed25519.GenerateKey(rand.Reader)
	oldEkey, newEkey := encryptionKey(), encryptionKey()
	newFP := pki.FingerprintOfKeys(newPub, newEkey)

	request := func(edit func(r *pki.RotateRequest), current, next ed25519.PrivateKey) []byte {
		r := pki.RotateRequest{
			MachineID: "m-1", PublicKey: newPub, EncryptionKey: newEkey, At: s.at.Unix(),
		}
		if edit != nil {
			edit(&r)
		}
		r.Sign(current, next)
		data, _ := json.Marshal(r)

		return data
	}
	admitted := admission.AcceptedAgent{
		MachineID: "m-1", Hostname: "web-01", PublicKey: oldPub, EncryptionKey: oldEkey,
		Fingerprint: pki.FingerprintOfKeys(oldPub, oldEkey), AcceptedAt: s.at,
		NATSUserKey: "UA1",
	}

	tests := []struct {
		name     string
		setup    func()
		grace    time.Duration
		data     []byte
		wantFP   string
		wantErr  string
		wantCode int
		// wantKey is the signing key that the admission holds after, and
		// wantSuperseded the one it keeps for the grace period.
		wantKey        ed25519.PublicKey
		wantSuperseded ed25519.PublicKey
	}{
		{
			name:           "an admitted machine moves to its new keys, and keeps the old one for the grace period",
			grace:          time.Hour,
			data:           request(nil, oldPriv, newPriv),
			wantFP:         newFP,
			wantCode:       200,
			wantKey:        newPub,
			wantSuperseded: oldPub,
		},
		{
			name:     "with no grace period the old key stops at once",
			data:     request(nil, oldPriv, newPriv),
			wantFP:   newFP,
			wantCode: 200,
			wantKey:  newPub,
		},
		{
			name: "the same request again, after the admission moved, is the same answer",
			setup: func() {
				moved := admitted
				moved.PublicKey, moved.EncryptionKey, moved.Fingerprint = newPub, newEkey, newFP
				s.put(admission.AcceptedKVPrefix()+"m-1", moved)
			},
			data:     request(nil, oldPriv, newPriv),
			wantFP:   newFP,
			wantCode: 200,
			wantKey:  newPub,
		},
		{
			name: "a request for the admitted keys that the new key did not sign",
			setup: func() {
				moved := admitted
				moved.PublicKey, moved.EncryptionKey, moved.Fingerprint = newPub, newEkey, newFP
				s.put(admission.AcceptedKVPrefix()+"m-1", moved)
			},
			data:     request(nil, oldPriv, otherPriv),
			wantErr:  "the rotation is not signed by the new key",
			wantCode: 403,
			wantKey:  newPub,
		},
		{
			name:    "a request that is not JSON",
			data:    []byte("{"),
			wantErr: "the rotation is not JSON",
			wantKey: oldPub,
		},
		{
			name: "a request from long ago",
			data: request(func(r *pki.RotateRequest) {
				r.At = s.at.Add(-time.Hour).Unix()
			}, oldPriv, newPriv),
			wantErr:  "the time of the rotation is not now",
			wantCode: 400,
			wantKey:  oldPub,
		},
		{
			name: "a new encryption key that is not a P-256 key",
			data: request(func(r *pki.RotateRequest) {
				r.EncryptionKey = "no"
			}, oldPriv, newPriv),
			wantErr:  "the new encryption key is not a P-256 key",
			wantCode: 400,
			wantKey:  oldPub,
		},
		{
			name: "a machine that is not admitted, such as a revoked one",
			data: request(func(r *pki.RotateRequest) {
				r.MachineID = "m-x"
			}, oldPriv, newPriv),
			wantErr:  "the machine is not admitted",
			wantCode: 403,
			wantKey:  oldPub,
		},
		{
			name:     "a request that a key other than the admitted one signed",
			data:     request(nil, otherPriv, newPriv),
			wantErr:  "not signed by the admitted key and the new key",
			wantCode: 403,
			wantKey:  oldPub,
		},
		{
			name:     "a request that the new key did not sign",
			data:     request(nil, oldPriv, otherPriv),
			wantErr:  "not signed by the admitted key and the new key",
			wantCode: 403,
			wantKey:  oldPub,
		},
		{
			name: "the new record cannot be encoded",
			setup: func() {
				admission.SetMarshalFn(
					func(any) ([]byte, error) { return nil, errors.New("no json") },
				)
			},
			data:     request(nil, oldPriv, newPriv),
			wantErr:  "the controller could not rotate: no json",
			wantCode: 500,
			wantKey:  oldPub,
		},
		{
			name:     "the new record cannot be stored",
			setup:    func() { s.kv.failPut = admission.AcceptedKVPrefix() },
			data:     request(nil, oldPriv, newPriv),
			wantErr:  "the controller could not rotate: kv: full",
			wantCode: 500,
			wantKey:  oldPub,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.put(admission.AcceptedKVPrefix()+"m-1", admitted)
			s.watcher.SetRotationGracePeriod(tt.grace)
			log := &joinLog{}
			s.watcher.SetAudit(log)
			if tt.setup != nil {
				tt.setup()
			}
			reply := s.rotate(tt.data)
			admission.ResetMarshalFn()
			s.Equal(tt.wantFP, reply.Fingerprint)
			if tt.wantErr == "" {
				s.Empty(reply.Error)
			} else {
				s.Contains(reply.Error, tt.wantErr)
			}

			var after admission.AcceptedAgent
			s.Require().
				NoError(json.Unmarshal(s.kv.data[admission.AcceptedKVPrefix()+"m-1"], &after))
			s.Equal(tt.wantKey, after.PublicKey)
			s.Equal(tt.wantSuperseded, after.SupersededKey)
			s.Equal("UA1", after.NATSUserKey, "the rest of the admission stays")
			if tt.wantSuperseded != nil {
				s.Equal(s.at.Add(tt.grace), after.SupersededUntil)
			}

			if tt.wantCode == 0 {
				s.Empty(log.entries, "a request that is not JSON names no machine")

				return
			}
			s.Require().Len(log.entries, 1)
			e := log.entries[0]
			s.Equal(audit.ActorAgent, e.Actor)
			s.Equal("agent.key.rotate", e.Action)
			s.Equal(tt.wantCode, e.ResponseCode)
		})
	}
}

func (s *StorePublicTestSuite) TestRotateSubscribe() {
	s.Run("a subscription that NATS refuses", func() {
		err := s.watcher.ServeRotate(
			context.Background(),
			func(string, string, nats.MsgHandler) (*nats.Subscription, error) {
				return nil, errors.New("no permission")
			},
		)
		s.ErrorContains(err, "subscribe to rotations: no permission")
	})
}

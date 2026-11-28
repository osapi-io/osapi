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
	"encoding/base64"
	"errors"
	"time"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/audit"
	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/controller/admission"
	admissionMocks "github.com/osapi-io/osapi/internal/controller/admission/mocks"
	"github.com/osapi-io/osapi/internal/controller/jointoken"
	"github.com/osapi-io/osapi/internal/natscallout"
	"github.com/osapi-io/osapi/internal/pki"
)

// tokens is a TokenChecker that admits osapi_jt_t1.secret until expires,
// or fails each check with err.
func (s *StorePublicTestSuite) tokens(
	expires time.Time,
	err error,
) *admissionMocks.MockTokenChecker {
	m := admissionMocks.NewMockTokenChecker(s.ctrl)
	m.EXPECT().Check(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, token string) (jointoken.Token, error) {
			if err != nil {
				return jointoken.Token{}, err
			}
			if token != "osapi_jt_t1.secret" {
				return jointoken.Token{}, jointoken.ErrInvalid
			}

			return jointoken.Token{ID: "t1", ExpiresAt: expires}, nil
		}).
		AnyTimes()

	return m
}

// joinWithUserKey joins m-1 with the NATS user key of user.
func (s *StorePublicTestSuite) joinWithUserKey(
	user nkeys.KeyPair,
) {
	s.watcher.SetJoinTokens(fakeTokens{})
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	userPub, _ := user.PublicKey()
	r := pki.JoinRequest{
		Token:         "osapi_jt_t1.secret",
		MachineID:     "m-1",
		Hostname:      "web-01",
		PublicKey:     pub,
		EncryptionKey: "BH0Fq2u2_BBIGuYw6KRZYg4I1pT5DACIZUCrhDK5Y7Phpl1MRex_y4Y60InurvbYIiZDQ2MHIcmOXF8SfLlA5fs",
		NATSUserKey:   userPub,
	}
	r.Sign(priv)
	_, err := s.watcher.Join(context.Background(), r)
	s.Require().NoError(err)
}

// nkeyRequest is the authorization request of a connection that offers
// key and signs the server's nonce with signer.
func nkeyRequest(
	key nkeys.KeyPair,
	signer nkeys.KeyPair,
) *jwt.AuthorizationRequest {
	pub, _ := key.PublicKey()
	sig, _ := signer.Sign([]byte("nonce-1"))

	return &jwt.AuthorizationRequest{
		ClientInformation: jwt.ClientInformation{Nonce: "nonce-1"},
		ConnectOptions: jwt.ConnectOptions{
			Nkey: pub, SignedNonce: base64.RawURLEncoding.EncodeToString(sig),
		},
	}
}

func (s *StorePublicTestSuite) TestCallout() {
	join := bus.Permissions{Publish: []string{"osapi.join.request"}}
	agent := bus.Permissions{Publish: []string{"osapi.jobs.>"}}
	user, _ := nkeys.CreateUser()
	other, _ := nkeys.CreateUser()
	roles := func(expires time.Time, ttl time.Duration) admission.CalloutRoles {
		return admission.CalloutRoles{
			Join: join, Agent: agent, AgentTTL: ttl, Tokens: s.tokens(expires, nil),
		}
	}

	tests := []struct {
		name    string
		setup   func()
		roles   func() admission.CalloutRoles
		req     func() *jwt.AuthorizationRequest
		want    natscallout.Decision
		wantErr string
		refused bool
		actor   audit.Actor
	}{
		{
			name:  "a join token gets the join role, for ten minutes at most",
			roles: func() admission.CalloutRoles { return roles(s.at.Add(time.Hour), 0) },
			req: func() *jwt.AuthorizationRequest {
				return &jwt.AuthorizationRequest{
					ConnectOptions: jwt.ConnectOptions{Token: "osapi_jt_t1.secret"},
				}
			},
			want: natscallout.Decision{
				Name:        "join:t1",
				Permissions: join,
				Expires:     s.at.Add(10 * time.Minute),
			},
			actor: audit.ActorJoinToken,
		},
		{
			name:  "a join token that expires sooner ends the connection then",
			roles: func() admission.CalloutRoles { return roles(s.at.Add(time.Minute), 0) },
			req: func() *jwt.AuthorizationRequest {
				return &jwt.AuthorizationRequest{
					ConnectOptions: jwt.ConnectOptions{Token: "osapi_jt_t1.secret"},
				}
			},
			want: natscallout.Decision{
				Name:        "join:t1",
				Permissions: join,
				Expires:     s.at.Add(time.Minute),
			},
			actor: audit.ActorJoinToken,
		},
		{
			name:  "a join token that does not admit is refused",
			roles: func() admission.CalloutRoles { return roles(s.at.Add(time.Hour), 0) },
			req: func() *jwt.AuthorizationRequest {
				return &jwt.AuthorizationRequest{
					ConnectOptions: jwt.ConnectOptions{Token: "osapi_jt_t2.secret"},
				}
			},
			wantErr: "the join token does not admit", refused: true, actor: audit.ActorJoinToken,
		},
		{
			name: "a join token that cannot be read is undecided, not refused",
			roles: func() admission.CalloutRoles {
				return admission.CalloutRoles{
					Join:   join,
					Tokens: s.tokens(time.Time{}, errors.New("read the join token: kv: down")),
				}
			},
			req: func() *jwt.AuthorizationRequest {
				return &jwt.AuthorizationRequest{
					ConnectOptions: jwt.ConnectOptions{Token: "osapi_jt_t1.secret"},
				}
			},
			wantErr: "read the join token: kv: down", actor: audit.ActorJoinToken,
		},
		{
			name:  "an admitted machine's NATS user key gets the agent role",
			setup: func() { s.joinWithUserKey(user) },
			roles: func() admission.CalloutRoles { return roles(time.Time{}, time.Hour) },
			req:   func() *jwt.AuthorizationRequest { return nkeyRequest(user, user) },
			want: natscallout.Decision{
				Name:        "agent:m-1",
				Permissions: agent,
				Expires:     s.at.Add(time.Hour),
			},
			actor: audit.ActorAgent,
		},
		{
			name:  "with no agent_ttl the connection does not expire",
			setup: func() { s.joinWithUserKey(user) },
			roles: func() admission.CalloutRoles { return roles(time.Time{}, 0) },
			req:   func() *jwt.AuthorizationRequest { return nkeyRequest(user, user) },
			want:  natscallout.Decision{Name: "agent:m-1", Permissions: agent},
			actor: audit.ActorAgent,
		},
		{
			name:    "a nonce that another key signed is refused",
			setup:   func() { s.joinWithUserKey(user) },
			roles:   func() admission.CalloutRoles { return roles(time.Time{}, 0) },
			req:     func() *jwt.AuthorizationRequest { return nkeyRequest(user, other) },
			wantErr: "not signed by the nkey", refused: true, actor: audit.ActorAgent,
		},
		{
			name:    "a key that no machine joined with is refused",
			roles:   func() admission.CalloutRoles { return roles(time.Time{}, 0) },
			req:     func() *jwt.AuthorizationRequest { return nkeyRequest(other, other) },
			wantErr: "no machine joined with this NATS user key",
			refused: true,
			actor:   audit.ActorAgent,
		},
		{
			name: "a machine that is no longer admitted is refused",
			setup: func() {
				s.joinWithUserKey(user)
				s.Require().NoError(s.watcher.Revoke(context.Background(), "m-1"))
			},
			roles:   func() admission.CalloutRoles { return roles(time.Time{}, 0) },
			req:     func() *jwt.AuthorizationRequest { return nkeyRequest(user, user) },
			wantErr: "the machine m-1 is not admitted", refused: true, actor: audit.ActorAgent,
		},
		{
			name: "an old key of a machine that joined again is refused",
			setup: func() {
				s.joinWithUserKey(user)
				s.joinWithUserKey(other)
			},
			roles:   func() admission.CalloutRoles { return roles(time.Time{}, 0) },
			req:     func() *jwt.AuthorizationRequest { return nkeyRequest(user, user) },
			wantErr: "joined again with another key", refused: true, actor: audit.ActorAgent,
		},
		{
			name: "admission that cannot be read is an error, not a refusal",
			setup: func() {
				s.joinWithUserKey(user)
				s.kv.failGet = admission.AcceptedKVPrefix()
			},
			roles:   func() admission.CalloutRoles { return roles(time.Time{}, 0) },
			req:     func() *jwt.AuthorizationRequest { return nkeyRequest(user, user) },
			wantErr: "kv: down", actor: audit.ActorAgent,
		},
		{
			name:  "a connection with no token and no key is refused",
			roles: func() admission.CalloutRoles { return roles(time.Time{}, 0) },
			req: func() *jwt.AuthorizationRequest {
				return &jwt.AuthorizationRequest{ConnectOptions: jwt.ConnectOptions{Username: "x"}}
			},
			wantErr: "no join token or NATS user key", refused: true, actor: audit.ActorAgent,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.setup != nil {
				tt.setup()
			}
			log := &joinLog{}
			s.watcher.SetAudit(log)
			defer s.watcher.SetAudit(nil)

			got, err := s.watcher.Callout(tt.roles())(context.Background(), tt.req())

			s.Require().NotEmpty(log.entries, "each decision is audited")
			e := log.entries[len(log.entries)-1]
			s.Equal("nats.connect", e.Action)
			s.Equal(tt.actor, e.Actor)
			if tt.wantErr != "" {
				s.ErrorContains(err, tt.wantErr)
				s.Equal(tt.refused, errors.Is(err, natscallout.ErrRefused))
				s.Equal(403, e.ResponseCode)

				return
			}
			s.Require().NoError(err)
			s.Equal(tt.want, got)
			s.Equal(200, e.ResponseCode)
		})
	}
}

// TestJoinIndexFails: a join whose NATS user key cannot be recorded still
// admits; that agent's callout connections are refused until it joins again.
func (s *StorePublicTestSuite) TestJoinIndexFails() {
	s.Run("the index cannot be written", func() {
		s.kv.failPut = "natsuser."
		user, _ := nkeys.CreateUser()
		s.joinWithUserKey(user)
		_, err := s.watcher.LookupAgentKey(context.Background(), "m-1")
		s.NoError(err, "the machine is admitted")
	})
}

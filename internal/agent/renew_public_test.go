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
package agent_test

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"os"
	"time"

	"github.com/avfs/avfs"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"

	"github.com/osapi-io/osapi/internal/agent"
	"github.com/osapi-io/osapi/internal/pki"
)

// renewSetup gives the agent keys, a saved join of scheme with the fields
// of join, and nats.creds
// with a JWT of the user key that expires in expiresIn seconds (0 never).
// Encode sets the JWT's issue time to the clock, so the tests move the
// agent's clock ahead with SetRenewFns. It returns the account that signs,
// and the user key.
func (s *AgentPublicTestSuite) renewSetup(
	scheme string,
	join map[string]any,
	expiresIn int64,
) (*agent.Agent, nkeys.KeyPair, string, *pki.Manager) {
	s.appConfig.Agent.PKI.KeyDir = "/pki"
	s.Require().NoError(s.appFs.MkdirAll("/pki", 0o700))
	a := s.buildAgent()
	m := pki.New(s.appFs, "/pki", "agent", pki.WithEncryptionKey())
	s.Require().NoError(m.LoadOrGenerate())
	agent.SetAgentPKIManager(a, m)

	acct, _ := nkeys.CreateAccount()
	user, _ := nkeys.CreateUser()
	userKey, _ := user.PublicKey()
	seed, _ := user.Seed()
	uc := jwt.NewUserClaims(userKey)
	if expiresIn > 0 {
		uc.Expires = time.Now().Unix() + expiresIn
	}
	token, _ := uc.Encode(acct)
	creds, _ := jwt.FormatUserConfig(token, seed)
	s.Require().NoError(avfs.WriteFile(s.appFs, "/pki/nats.creds", creds, 0o600))
	fields := map[string]any{
		"scheme": scheme, "host": "hub", "port": 4222, "namespace": "osapi",
		"creds_file": "/pki/nats.creds",
	}
	maps.Copy(fields, join)
	saved, _ := json.Marshal(fields)
	s.Require().NoError(avfs.WriteFile(s.appFs, "/pki/"+agent.NATSJoinFile, saved, 0o600))

	return a, acct, userKey, m
}

func (s *AgentPublicTestSuite) TestRenewIfDue() {
	signed := func(
		acct nkeys.KeyPair,
		userKey string,
		mod func(*jwt.UserClaims),
	) ([]byte, error) {
		uc := jwt.NewUserClaims(userKey)
		uc.Name = "renewed"
		if mod != nil {
			mod(uc)
		}
		token, _ := uc.Encode(acct)

		return json.Marshal(pki.JoinReply{NATS: pki.JoinNATS{JWT: token}})
	}
	tests := []struct {
		name         string
		scheme       string
		join         map[string]any
		caDir        bool
		expiresIn    int64
		creds        []byte
		seedless     bool
		gone         bool
		reply        func(acct nkeys.KeyPair, userKey string) ([]byte, error)
		write        error
		wantCall     bool
		validateFunc func(
			a *agent.Agent,
			calls []renewCall,
			userKey string,
			m *pki.Manager,
			err error,
		)
	}{
		{
			name:      "a JWT with a third of its life left is renewed, for the same user key",
			scheme:    "tls",
			join:      map[string]any{"server_name": "nats.example.com"},
			expiresIn: 60,
			reply: func(acct nkeys.KeyPair, userKey string) ([]byte, error) {
				return signed(acct, userKey, nil)
			},
			wantCall: true,
			validateFunc: func(
				_ *agent.Agent,
				calls []renewCall,
				userKey string,
				m *pki.Manager,
				err error,
			) {
				s.Require().NoError(err)
				s.Require().Len(calls, 1)
				s.Equal("tls://hub:4222", calls[0].url)
				s.Equal("osapi.join.renew", calls[0].subject)
				s.NoError(calls[0].optsErr)
				s.Equal("nats.example.com", calls[0].opts.TLSConfig.ServerName)
				s.Equal(userKey, calls[0].req.NATSUserKey)
				s.Equal("m-1", calls[0].req.MachineID)
				s.True(calls[0].req.Verify(m.PublicKey()))
				creds, _ := avfs.ReadFile(s.appFs, "/pki/nats.creds")
				token, _ := jwt.ParseDecoratedJWT(creds)
				uc, _ := jwt.DecodeUserClaims(token)
				s.Equal("renewed", uc.Name)
			},
		},
		{
			name:      "a JWT with more than a third left is kept",
			scheme:    "nats",
			expiresIn: 3600,
			validateFunc: func(
				_ *agent.Agent,
				calls []renewCall,
				_ string,
				_ *pki.Manager,
				err error,
			) {
				s.NoError(err)
				s.Empty(calls)
			},
		},
		{
			name:   "a JWT that never expires is kept",
			scheme: "nats",
			validateFunc: func(
				_ *agent.Agent,
				calls []renewCall,
				_ string,
				_ *pki.Manager,
				err error,
			) {
				s.NoError(err)
				s.Empty(calls)
			},
		},
		{
			name:         "a saved CA that is not a certificate",
			scheme:       "tls",
			join:         map[string]any{"ca": "not a certificate"},
			expiresIn:    60,
			validateFunc: wantRenewErr(s, "nats tls"),
		},
		{
			name:         "a saved CA that cannot be kept",
			scheme:       "tls",
			join:         map[string]any{"ca": "not a certificate"},
			caDir:        true,
			expiresIn:    60,
			validateFunc: wantRenewErr(s, "nats-join-ca.pem"),
		},
		{
			name:      "a controller that refuses",
			scheme:    "nats",
			expiresIn: 60,
			reply: func(nkeys.KeyPair, string) ([]byte, error) {
				return json.Marshal(pki.JoinReply{Error: "the machine is not admitted"})
			},
			validateFunc: wantRenewErr(s, "the machine is not admitted"),
		},
		{
			name:      "a server that does not answer",
			scheme:    "nats",
			expiresIn: 60,
			reply: func(nkeys.KeyPair, string) ([]byte, error) {
				return nil, nats.ErrTimeout
			},
			validateFunc: wantRenewErr(s, "nats://hub:4222: nats: timeout"),
		},
		{
			name:      "a reply that is not JSON",
			scheme:    "nats",
			expiresIn: 60,
			reply: func(nkeys.KeyPair, string) ([]byte, error) {
				return []byte("{"), nil
			},
			validateFunc: wantRenewErr(s, "the reply of the renewal"),
		},
		{
			name:      "a reply with no JWT",
			scheme:    "nats",
			expiresIn: 60,
			reply: func(nkeys.KeyPair, string) ([]byte, error) {
				return json.Marshal(pki.JoinReply{})
			},
			validateFunc: wantRenewErr(s, "the JWT of the renewal"),
		},
		{
			name:      "a JWT for another key",
			scheme:    "nats",
			expiresIn: 60,
			reply: func(acct nkeys.KeyPair, _ string) ([]byte, error) {
				other, _ := nkeys.CreateUser()
				pub, _ := other.PublicKey()

				return signed(acct, pub, nil)
			},
			validateFunc: wantRenewErr(s, "the JWT of the renewal"),
		},
		{
			name:      "a JWT that has expired",
			scheme:    "nats",
			expiresIn: 60,
			reply: func(acct nkeys.KeyPair, userKey string) ([]byte, error) {
				return signed(acct, userKey, func(uc *jwt.UserClaims) {
					uc.Expires = time.Now().Add(-time.Hour).Unix()
				})
			},
			validateFunc: wantRenewErr(s, "the JWT of the renewal has expired"),
		},
		{
			name:         "creds that are not a file of creds",
			scheme:       "nats",
			expiresIn:    60,
			creds:        []byte("nothing"),
			validateFunc: wantRenewErr(s, "/pki/nats.creds"),
		},
		{
			name:      "creds whose JWT is not a user's",
			scheme:    "nats",
			expiresIn: 60,
			creds: []byte(
				"-----BEGIN NATS USER JWT-----\nnot.a.jwt\n------END NATS USER JWT------\n",
			),
			validateFunc: wantRenewErr(s, "/pki/nats.creds"),
		},
		{
			name:         "creds with a JWT and no seed",
			scheme:       "nats",
			expiresIn:    60,
			seedless:     true,
			validateFunc: wantRenewErr(s, "/pki/nats.creds"),
		},
		{
			name:         "creds that are gone",
			scheme:       "nats",
			expiresIn:    60,
			gone:         true,
			validateFunc: wantRenewErr(s, "/pki/nats.creds"),
		},
		{
			name:      "a renewal that cannot be kept",
			scheme:    "nats",
			expiresIn: 60,
			reply: func(acct nkeys.KeyPair, userKey string) ([]byte, error) {
				return signed(acct, userKey, nil)
			},
			write:        errors.New("disk full"),
			validateFunc: wantRenewErr(s, "disk full"),
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			// The agent's clock is 50 seconds ahead: a JWT of 60 seconds has
			// a sixth left, and one of an hour most of it. TearDownSubTest
			// puts back each function that the row swaps.
			agent.SetRenewFns(func() time.Time { return time.Now().Add(50 * time.Second) }, 0, nil)
			a, acct, userKey, m := s.renewSetup(tt.scheme, tt.join, tt.expiresIn)
			if tt.caDir {
				s.Require().NoError(s.appFs.RemoveAll("/pki/nats-join-ca.pem"))
				s.Require().NoError(s.appFs.MkdirAll("/pki/nats-join-ca.pem/x", 0o700))
			}
			creds := tt.creds
			if tt.seedless {
				uc := jwt.NewUserClaims(userKey)
				uc.Expires = time.Now().Unix() + 60
				token, _ := uc.Encode(acct)
				creds = []byte(
					"-----BEGIN NATS USER JWT-----\n" + token + "\n------END NATS USER JWT------\n",
				)
			}
			if creds != nil {
				s.Require().NoError(avfs.WriteFile(s.appFs, "/pki/nats.creds", creds, 0o600))
			}
			if tt.gone {
				s.Require().NoError(s.appFs.Remove("/pki/nats.creds"))
			}
			var calls []renewCall
			agent.SetJoinNATSFns(
				nil,
				func(_ context.Context, url string, opts []nats.Option, subject string, data []byte) ([]byte, error) {
					call := renewCall{url: url, subject: subject}
					for _, o := range opts {
						if err := o(&call.opts); err != nil {
							call.optsErr = err
						}
					}
					_ = json.Unmarshal(data, &call.req)
					calls = append(calls, call)
					if tt.reply == nil {
						return nil, errors.New("no reply")
					}

					return tt.reply(acct, userKey)
				},
			)
			if tt.write != nil {
				agent.SetRenewFns(
					nil,
					0,
					func(context.Context, avfs.VFS, string, []byte, os.FileMode) error {
						return tt.write
					},
				)
			}

			tt.validateFunc(a, calls, userKey, m, agent.ExportRenewIfDue(a, "m-1"))
		})
	}
}

// renewCall is one request of a renewal: its URL, its subject, the options
// applied to a nats.Options, and the request that it sent.
type renewCall struct {
	url     string
	subject string
	opts    nats.Options
	optsErr error
	req     pki.RenewRequest
}

// wantRenewErr is a validateFunc of TestRenewIfDue that wants an error with
// text.
func wantRenewErr(
	s *AgentPublicTestSuite,
	text string,
) func(*agent.Agent, []renewCall, string, *pki.Manager, error) {
	return func(
		_ *agent.Agent,
		_ []renewCall,
		_ string,
		_ *pki.Manager,
		err error,
	) {
		s.ErrorContains(err, text)
	}
}

func (s *AgentPublicTestSuite) TestStartRenewal() {
	tests := []struct {
		name  string
		setup func() (*agent.Agent, context.Context)
	}{
		{
			name: "an agent with no join renews nothing",
			setup: func() (*agent.Agent, context.Context) {
				s.appConfig.Agent.PKI.KeyDir = "/none"
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				return s.buildAgent(), ctx
			},
		},
		{
			name: "an agent with creds checks them until it stops, and logs a failure",
			setup: func() (*agent.Agent, context.Context) {
				a, _, _, _ := s.renewSetup("nats", nil, 60)
				calls := make(chan struct{}, 10)
				agent.SetJoinNATSFns(
					nil,
					func(context.Context, string, []nats.Option, string, []byte) ([]byte, error) {
						calls <- struct{}{}

						return nil, nats.ErrTimeout
					},
				)
				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					<-calls
					cancel()
				}()

				return a, ctx
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Check each 5 ms, with the agent's clock ahead so that
			// a JWT of 60 seconds is due. TearDownSubTest puts both back.
			agent.SetRenewFns(
				func() time.Time { return time.Now().Add(50 * time.Second) },
				5*time.Millisecond,
				nil,
			)
			a, ctx := tt.setup()

			agent.ExportStartRenewal(ctx, a)
		})
	}
}

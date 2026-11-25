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
	"os"
	"strings"
	"time"

	"github.com/avfs/avfs"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"

	"github.com/osapi-io/osapi/internal/agent"
	"github.com/osapi-io/osapi/internal/pki"
)

// renewSetup gives the agent keys, a saved join of scheme, and nats.creds
// with a JWT of the user key that expires in expiresIn seconds (0 never).
// Encode sets the JWT's issue time to the clock, so the tests move the
// agent's clock ahead with SetRenewFns. It returns the account that signs,
// and the user key.
func (s *AgentPublicTestSuite) renewSetup(
	scheme string,
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
	saved, _ := json.Marshal(map[string]any{
		"scheme": scheme, "host": "hub", "port": 4222, "namespace": "osapi",
		"creds_file": "/pki/nats.creds",
	})
	s.Require().NoError(avfs.WriteFile(s.appFs, "/pki/"+agent.NATSJoinFile, saved, 0o600))

	return a, acct, userKey, m
}

func (s *AgentPublicTestSuite) TestRenewIfDue() {
	// The agent's clock is 50 seconds ahead: a JWT of 60 seconds has a
	// sixth left, and one of an hour most of it.
	defer agent.SetRenewFns(func() time.Time { return time.Now().Add(50 * time.Second) }, 0, nil)()

	s.Run("a JWT with a third of its life left is renewed, for the same user key", func() {
		a, acct, userKey, m := s.renewSetup("tls", 60)
		defer agent.SetJoinNATSFns(
			nil,
			func(url string, _ []nats.Option, subject string, data []byte) ([]byte, error) {
				s.Equal("tls://hub:4222", url)
				s.Equal("osapi.join.renew", subject)
				var req pki.RenewRequest
				s.Require().NoError(json.Unmarshal(data, &req))
				s.Equal("m-1", req.MachineID)
				s.Equal(userKey, req.NATSUserKey)
				s.True(req.Verify(m.PublicKey()), "signed with the agent's key")
				next := jwt.NewUserClaims(userKey)
				next.Name = "renewed"
				token, _ := next.Encode(acct)

				return json.Marshal(pki.JoinReply{NATS: pki.JoinNATS{JWT: token}})
			},
		)()

		s.Require().NoError(agent.ExportRenewIfDue(a, "m-1"))
		creds, _ := avfs.ReadFile(s.appFs, "/pki/nats.creds")
		token, _ := jwt.ParseDecoratedJWT(creds)
		uc, _ := jwt.DecodeUserClaims(token)
		s.Equal("renewed", uc.Name)
	})

	noCall := func(string, []nats.Option, string, []byte) ([]byte, error) {
		s.Fail("no renewal is due")

		return nil, nil
	}
	s.Run("a JWT with more than a third left is kept", func() {
		a, _, _, _ := s.renewSetup("nats", 3600)
		defer agent.SetJoinNATSFns(nil, noCall)()
		s.NoError(agent.ExportRenewIfDue(a, "m-1"))
	})
	s.Run("a JWT that never expires is kept", func() {
		a, _, _, _ := s.renewSetup("nats", 0)
		defer agent.SetJoinNATSFns(nil, noCall)()
		s.NoError(agent.ExportRenewIfDue(a, "m-1"))
	})

	failures := []struct {
		name    string
		reply   func(acct nkeys.KeyPair) ([]byte, error)
		creds   []byte
		write   error
		wantErr string
	}{
		{name: "a controller that refuses", reply: func(nkeys.KeyPair) ([]byte, error) {
			return json.Marshal(pki.JoinReply{Error: "the machine is not admitted"})
		}, wantErr: "the machine is not admitted"},
		{name: "a server that does not answer", reply: func(nkeys.KeyPair) ([]byte, error) {
			return nil, nats.ErrTimeout
		}, wantErr: "nats://hub:4222: nats: timeout"},
		{name: "a reply that is not JSON", reply: func(nkeys.KeyPair) ([]byte, error) {
			return []byte("{"), nil
		}, wantErr: "the reply of the renewal"},
		{name: "a JWT for another key", reply: func(acct nkeys.KeyPair) ([]byte, error) {
			other, _ := nkeys.CreateUser()
			pub, _ := other.PublicKey()
			token, _ := jwt.NewUserClaims(pub).Encode(acct)

			return json.Marshal(pki.JoinReply{NATS: pki.JoinNATS{JWT: token}})
		}, wantErr: "the JWT of the renewal"},
		{
			name:    "creds that are not a file of creds",
			creds:   []byte("nothing"),
			wantErr: "/pki/nats.creds",
		},
		{
			name: "creds whose JWT is not a user's",
			creds: []byte(
				"-----BEGIN NATS USER JWT-----\nnot.a.jwt\n------END NATS USER JWT------\n",
			),
			wantErr: "/pki/nats.creds",
		},
		{
			name:    "a renewal that cannot be kept",
			write:   errors.New("disk full"),
			wantErr: "disk full",
		},
	}
	for _, tt := range failures {
		s.Run(tt.name, func() {
			a, acct, userKey, _ := s.renewSetup("nats", 60)
			if tt.creds != nil {
				s.Require().NoError(avfs.WriteFile(s.appFs, "/pki/nats.creds", tt.creds, 0o600))
			}
			reply := tt.reply
			if reply == nil {
				reply = func(acct nkeys.KeyPair) ([]byte, error) {
					token, _ := jwt.NewUserClaims(userKey).Encode(acct)

					return json.Marshal(pki.JoinReply{NATS: pki.JoinNATS{JWT: token}})
				}
			}
			defer agent.SetJoinNATSFns(
				nil,
				func(string, []nats.Option, string, []byte) ([]byte, error) {
					return reply(acct)
				},
			)()
			if tt.write != nil {
				defer agent.SetRenewFns(
					nil,
					0,
					func(context.Context, avfs.VFS, string, []byte, os.FileMode) error {
						return tt.write
					},
				)()
			}

			s.ErrorContains(agent.ExportRenewIfDue(a, "m-1"), tt.wantErr)
		})
	}

	s.Run("creds with a JWT and no seed", func() {
		a, acct, userKey, _ := s.renewSetup("nats", 60)
		uc := jwt.NewUserClaims(userKey)
		uc.Expires = time.Now().Unix() + 60
		token, _ := uc.Encode(acct)
		s.Require().NoError(avfs.WriteFile(
			s.appFs,
			"/pki/nats.creds",
			[]byte(
				"-----BEGIN NATS USER JWT-----\n"+token+"\n------END NATS USER JWT------\n",
			),
			0o600,
		))
		err := agent.ExportRenewIfDue(a, "m-1")
		s.Error(err)
		s.True(strings.Contains(err.Error(), "/pki/nats.creds"))
	})

	s.Run("creds that are gone", func() {
		a, _, _, _ := s.renewSetup("nats", 60)
		s.Require().NoError(s.appFs.Remove("/pki/nats.creds"))
		s.Error(agent.ExportRenewIfDue(a, "m-1"))
	})
}

func (s *AgentPublicTestSuite) TestStartRenewal() {
	defer agent.SetRenewFns(
		func() time.Time { return time.Now().Add(50 * time.Second) },
		5*time.Millisecond,
		nil,
	)()

	s.Run("an agent with no join renews nothing", func() {
		s.appConfig.Agent.PKI.KeyDir = "/none"
		a := s.buildAgent()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		agent.ExportStartRenewal(ctx, a)
	})

	s.Run("an agent with creds checks them until it stops, and logs a failure", func() {
		a, _, _, _ := s.renewSetup("nats", 60)
		calls := make(chan struct{}, 10)
		defer agent.SetJoinNATSFns(
			nil,
			func(string, []nats.Option, string, []byte) ([]byte, error) {
				calls <- struct{}{}

				return nil, nats.ErrTimeout
			},
		)()
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			<-calls
			cancel()
		}()
		agent.ExportStartRenewal(ctx, a)
	})
}

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
package natscallout_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/bus"
	"github.com/osapi-io/osapi/internal/natscallout"
)

type CalloutPublicTestSuite struct {
	suite.Suite
	srv    *server.Server
	issuer nkeys.KeyPair
	agent  nkeys.KeyPair
	cancel context.CancelFunc
}

func TestCalloutPublicTestSuite(t *testing.T) {
	suite.Run(t, new(CalloutPublicTestSuite))
}

// SetupSuite runs a NATS server whose auth_callout sends each connection
// but the callout user's to the service, as a NATS administrator sets one.
func (s *CalloutPublicTestSuite) SetupSuite() {
	s.issuer, _ = nkeys.CreateAccount()
	issuerPub, _ := s.issuer.PublicKey()
	s.agent, _ = nkeys.CreateUser()
	authAcc := server.NewAccount("AUTH")
	osapiAcc := server.NewAccount("OSAPI")
	opts := &server.Options{
		Host:   "127.0.0.1",
		Port:   -1,
		NoLog:  true,
		NoSigs: true,
		// The server offers a nonce, so an agent can sign it with its key.
		AlwaysEnableNonce: true,
		Accounts:          []*server.Account{authAcc, osapiAcc},
		Users:             []*server.User{{Username: "callout", Password: "pw", Account: authAcc}},
		AuthCallout: &server.AuthCallout{
			Issuer: issuerPub, Account: "AUTH", AuthUsers: []string{"callout"},
		},
	}
	srv, err := server.NewServer(opts)
	s.Require().NoError(err)
	go srv.Start()
	s.Require().True(srv.ReadyForConnections(5 * time.Second))
	s.srv = srv

	nc, err := nats.Connect(srv.ClientURL(), nats.UserInfo("callout", "pw"))
	s.Require().NoError(err)
	agentPub, _ := s.agent.PublicKey()
	svc := natscallout.Service{
		Issuer:  s.issuer,
		Account: "OSAPI",
		Logger:  slog.New(slog.DiscardHandler),
		Decide: func(_ context.Context, req *jwt.AuthorizationRequest) (natscallout.Decision, error) {
			switch {
			case req.ConnectOptions.Token == "osapi_jt_good":
				return natscallout.Decision{
					Name:        "join:good",
					Permissions: bus.Permissions{Publish: []string{"osapi.join.request"}},
					Expires:     time.Now().Add(time.Minute),
				}, nil
			case req.ConnectOptions.Nkey == agentPub:
				if err := natscallout.VerifyNkey(req); err != nil {
					return natscallout.Decision{}, err
				}

				return natscallout.Decision{
					Name: "agent:m-1",
					Permissions: bus.Permissions{
						Publish:   []string{"osapi.ok"},
						Subscribe: []string{"osapi.ok"},
					},
				}, nil
			}

			return natscallout.Decision{}, fmt.Errorf(
				"%w: no join token or admitted key",
				natscallout.ErrRefused,
			)
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	go func() { _ = svc.Serve(ctx, nc.Subscribe) }()
	time.Sleep(100 * time.Millisecond)
}

func (s *CalloutPublicTestSuite) TearDownSuite() {
	s.cancel()
	s.srv.Shutdown()
}

func (s *CalloutPublicTestSuite) TestConnections() {
	agentPub, _ := s.agent.PublicKey()
	other, _ := nkeys.CreateUser()
	otherPub, _ := other.PublicKey()
	sign := func(k nkeys.KeyPair) nats.SignatureHandler {
		return func(nonce []byte) ([]byte, error) { return k.Sign(nonce) }
	}
	tests := []struct {
		name    string
		opts    []nats.Option
		wantErr bool
		allowed string
		denied  string
	}{
		{
			name:    "a join token gets the join role",
			opts:    []nats.Option{nats.Token("osapi_jt_good")},
			allowed: "osapi.join.request",
			denied:  "osapi.jobs.x",
		},
		{
			name:    "an admitted agent's key gets the agent role",
			opts:    []nats.Option{nats.Nkey(agentPub, sign(s.agent))},
			allowed: "osapi.ok",
			denied:  "osapi.join.request",
		},
		{
			name:    "a key that signs for another key is refused",
			opts:    []nats.Option{nats.Nkey(agentPub, sign(other))},
			wantErr: true,
		},
		{
			name:    "a key that is not admitted is refused",
			opts:    []nats.Option{nats.Nkey(otherPub, sign(other))},
			wantErr: true,
		},
		{
			name:    "a wrong token is refused",
			opts:    []nats.Option{nats.Token("osapi_jt_bad")},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			violations := make(chan error, 4)
			opts := append(tt.opts, nats.MaxReconnects(0), nats.ErrorHandler(
				func(_ *nats.Conn, _ *nats.Subscription, err error) { violations <- err },
			))
			nc, err := nats.Connect(s.srv.ClientURL(), opts...)
			if tt.wantErr {
				s.Error(err)

				return
			}
			s.Require().NoError(err)
			defer nc.Close()
			s.Require().NoError(nc.Publish(tt.allowed, nil))
			s.Require().NoError(nc.Flush())
			s.Require().NoError(nc.Publish(tt.denied, nil))
			_ = nc.Flush()
			select {
			case err := <-violations:
				s.True(errors.Is(err, nats.ErrPermissionViolation) ||
					err != nil, "the role does not have %s", tt.denied)
			case <-time.After(2 * time.Second):
				s.Fail("the server let the role publish to " + tt.denied)
			}
		})
	}
}

func (s *CalloutPublicTestSuite) TestReplyRefusesWhatIsNotARequest() {
	svc := natscallout.Service{Issuer: s.issuer}
	_, err := svc.Reply(context.Background(), []byte("not a jwt"))
	s.ErrorContains(err, "the authorization request")
}

func (s *CalloutPublicTestSuite) TestVerifyNkey() {
	tests := []struct {
		name string
		opts jwt.ConnectOptions
		want string
	}{
		{name: "no nkey", want: "no nkey"},
		{
			name: "not a public key",
			opts: jwt.ConnectOptions{Nkey: "Ubad", SignedNonce: "x"},
			want: "not a public key",
		},
		{name: "a signature that does not decode", opts: jwt.ConnectOptions{Nkey: func() string {
			p, _ := s.agent.PublicKey()

			return p
		}(), SignedNonce: "!!!"}, want: "not signed"},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := natscallout.VerifyNkey(&jwt.AuthorizationRequest{ConnectOptions: tt.opts})
			s.ErrorIs(err, natscallout.ErrRefused)
			s.ErrorContains(err, tt.want)
		})
	}
}

// request is an authorization request from a server, for a user.
func request(
	s *CalloutPublicTestSuite,
) []byte {
	srvKey, _ := nkeys.CreateServer()
	srvPub, _ := srvKey.PublicKey()
	userKey, _ := nkeys.CreateUser()
	userPub, _ := userKey.PublicKey()
	rc := jwt.NewAuthorizationRequestClaims(userPub)
	rc.UserNkey = userPub
	rc.Server = jwt.ServerID{ID: srvPub}
	token, err := rc.Encode(srvKey)
	s.Require().NoError(err)

	return []byte(token)
}

func (s *CalloutPublicTestSuite) TestReply() {
	issuerPub, _ := s.issuer.PublicKey()
	publicOnly, _ := nkeys.FromPublicKey(issuerPub)
	allow := func(context.Context, *jwt.AuthorizationRequest) (natscallout.Decision, error) {
		return natscallout.Decision{
			Name: "controller", Permissions: bus.Permissions{Responses: true},
		}, nil
	}
	refuse := func(context.Context, *jwt.AuthorizationRequest) (natscallout.Decision, error) {
		return natscallout.Decision{}, natscallout.ErrRefused
	}

	s.Run("a role that answers requests may answer each once", func() {
		svc := natscallout.Service{Issuer: s.issuer, Account: "OSAPI", Decide: allow}
		out, err := svc.Reply(context.Background(), request(s))
		s.Require().NoError(err)
		resp, err := jwt.DecodeAuthorizationResponseClaims(string(out))
		s.Require().NoError(err)
		uc, err := jwt.DecodeUserClaims(resp.Jwt)
		s.Require().NoError(err)
		s.Require().NotNil(uc.Resp)
		s.Equal(1, uc.Resp.MaxMsgs)
		s.Equal("OSAPI", uc.Audience)
	})
	s.Run("a refusal sends its reason to the server", func() {
		svc := natscallout.Service{Issuer: s.issuer, Account: "OSAPI", Decide: refuse}
		out, err := svc.Reply(context.Background(), request(s))
		s.Require().NoError(err)
		resp, err := jwt.DecodeAuthorizationResponseClaims(string(out))
		s.Require().NoError(err)
		s.Equal("refused", resp.Error)
	})
	s.Run("an error that is not a refusal sends Undecided, not its cause", func() {
		svc := natscallout.Service{
			Issuer: s.issuer, Account: "OSAPI", Logger: slog.New(slog.DiscardHandler),
			Decide: func(context.Context, *jwt.AuthorizationRequest) (natscallout.Decision, error) {
				return natscallout.Decision{}, errors.New("kv get ENROLLMENT/accepted.m-1: timeout")
			},
		}
		out, err := svc.Reply(context.Background(), request(s))
		s.Require().NoError(err)
		resp, err := jwt.DecodeAuthorizationResponseClaims(string(out))
		s.Require().NoError(err)
		s.Equal(natscallout.Undecided, resp.Error)
		s.NotContains(resp.Error, "ENROLLMENT")
	})
	s.Run("a key that cannot sign cannot sign the user", func() {
		svc := natscallout.Service{Issuer: publicOnly, Decide: allow}
		_, err := svc.Reply(context.Background(), request(s))
		s.ErrorContains(err, "sign the user")
	})
	s.Run("a key that cannot sign cannot sign a refusal", func() {
		svc := natscallout.Service{Issuer: publicOnly, Decide: refuse}
		_, err := svc.Reply(context.Background(), request(s))
		s.ErrorContains(err, "sign the answer")
	})
}

func (s *CalloutPublicTestSuite) TestServe() {
	s.Run("a subscription that fails is the error", func() {
		svc := natscallout.Service{Issuer: s.issuer}
		err := svc.Serve(
			context.Background(),
			func(string, nats.MsgHandler) (*nats.Subscription, error) {
				return nil, errors.New("no permission")
			},
		)
		s.ErrorContains(err, "subscribe to $SYS.REQ.USER.AUTH: no permission")
	})
	s.Run("it decides a second request while the first one waits", func() {
		// A NATS subscription calls its handler for one message at a time.
		// Here the first decision waits until the second one has started,
		// so a Serve that decides in its handler never gets the second.
		started := make(chan struct{}, 2)
		both := make(chan struct{})
		decided := make(chan string, 2)
		svc := natscallout.Service{
			Issuer: s.issuer, Account: "OSAPI", Logger: slog.New(slog.DiscardHandler),
			Decide: func(context.Context, *jwt.AuthorizationRequest) (natscallout.Decision, error) {
				started <- struct{}{}
				select {
				case <-both:
					decided <- "both"
				case <-time.After(2 * time.Second):
					decided <- "alone"
				}

				return natscallout.Decision{}, natscallout.ErrRefused
			},
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		handlers := make(chan nats.MsgHandler, 1)
		go func() {
			_ = svc.Serve(ctx, func(_ string, h nats.MsgHandler) (*nats.Subscription, error) {
				handlers <- h

				return &nats.Subscription{}, nil
			})
		}()
		handler := <-handlers
		handler(&nats.Msg{Data: request(s)})
		handler(&nats.Msg{Data: request(s)})
		<-started
		<-started
		close(both)
		s.Equal("both", <-decided)
		s.Equal("both", <-decided)
	})
	s.Run("a request that waits for a slot when ctx ends has no answer", func() {
		defer natscallout.SetMaxInFlight(1)()
		release := make(chan struct{})
		calls := make(chan struct{}, 2)
		svc := natscallout.Service{
			Issuer: s.issuer, Account: "OSAPI",
			Decide: func(context.Context, *jwt.AuthorizationRequest) (natscallout.Decision, error) {
				calls <- struct{}{}
				<-release

				return natscallout.Decision{}, errors.New("no bucket")
			},
		}
		ctx, cancel := context.WithCancel(context.Background())
		handlers := make(chan nats.MsgHandler, 1)
		go func() {
			_ = svc.Serve(ctx, func(_ string, h nats.MsgHandler) (*nats.Subscription, error) {
				handlers <- h

				return &nats.Subscription{}, nil
			})
		}()
		handler := <-handlers
		handler(&nats.Msg{Data: request(s)})
		<-calls
		cancel()
		handler(&nats.Msg{Data: request(s)})
		close(release)
		select {
		case <-calls:
			s.Fail("the second request was decided after ctx ended")
		case <-time.After(100 * time.Millisecond):
		}
	})
	s.Run("a request that is not one is logged, and has no answer", func() {
		svc := natscallout.Service{Issuer: s.issuer, Logger: slog.New(slog.DiscardHandler)}
		ctx, cancel := context.WithCancel(context.Background())
		nc, err := nats.Connect(s.srv.ClientURL(), nats.UserInfo("callout", "pw"))
		s.Require().NoError(err)
		defer nc.Close()
		handlers := make(chan nats.MsgHandler, 1)
		done := make(chan error, 1)
		go func() {
			done <- svc.Serve(ctx, func(subj string, h nats.MsgHandler) (*nats.Subscription, error) {
				handlers <- h

				return nc.Subscribe(subj+".test", h)
			})
		}()
		handler := <-handlers
		handler(&nats.Msg{Data: []byte("not a request")})
		cancel()
		s.NoError(<-done)
	})
}

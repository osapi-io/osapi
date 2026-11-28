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

package natsrevoke_test

import (
	"context"
	"errors"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/osfs"
	"github.com/nats-io/jwt/v2"
	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nkeys"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/osapi-io/osapi/internal/config"
	jobmocks "github.com/osapi-io/osapi/internal/job/mocks"
	"github.com/osapi-io/osapi/internal/natsauth"
	"github.com/osapi-io/osapi/internal/natsrevoke"
	"github.com/osapi-io/osapi/internal/natsrevoke/mocks"
)

const bucket = "enrollment"

type NatsRevokePublicTestSuite struct {
	suite.Suite
	ctrl *gomock.Controller
	// fs is the host's own filesystem: the NATS server of a test reads the
	// resolver config and the account JWT by their paths.
	fs  avfs.VFS
	dir string
	now time.Time
}

func TestNatsRevokePublicTestSuite(t *testing.T) {
	suite.Run(t, new(NatsRevokePublicTestSuite))
}

func (s *NatsRevokePublicTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.fs = osfs.NewWithNoIdm()
	natsrevoke.SetHostFS(s.fs)
	s.dir = s.T().TempDir()
	s.now = time.Unix(1_800_000_000, 0)
	natsrevoke.SetNowFn(func() time.Time { return s.now })
}

func (s *NatsRevokePublicTestSuite) TearDownTest() {
	natsrevoke.ResetNowFn()
	natsrevoke.ResetHostFS()
}

// memKV is a KV of a map, for a test that needs the store to keep what it
// is given.
type memKV struct {
	mu   sync.Mutex
	m    map[string][]byte
	ctrl *gomock.Controller
}

func (s *NatsRevokePublicTestSuite) memKV() (*mocks.MockKV, *memKV) {
	store := &memKV{m: map[string][]byte{}, ctrl: s.ctrl}
	kv := mocks.NewMockKV(s.ctrl)
	kv.EXPECT().KVPut(gomock.Any(), bucket, gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _, k string, v []byte) (uint64, error) {
			store.mu.Lock()
			defer store.mu.Unlock()
			store.m[k] = v

			return 1, nil
		}).AnyTimes()
	kv.EXPECT().KVKeys(gomock.Any(), bucket, gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, _ ...string) ([]string, error) {
			store.mu.Lock()
			defer store.mu.Unlock()
			if len(store.m) == 0 {
				return nil, jetstream.ErrNoKeysFound
			}
			keys := make([]string, 0, len(store.m))
			for k := range store.m {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			return keys, nil
		}).AnyTimes()
	kv.EXPECT().KVGet(gomock.Any(), bucket, gomock.Any()).
		DoAndReturn(func(_ context.Context, _, k string) (jetstream.KeyValueEntry, error) {
			store.mu.Lock()
			defer store.mu.Unlock()

			return entry(s.ctrl, store.m[k]), nil
		}).AnyTimes()
	kv.EXPECT().KVDelete(gomock.Any(), bucket, gomock.Any()).
		DoAndReturn(func(_ context.Context, _, k string) error {
			store.mu.Lock()
			defer store.mu.Unlock()
			delete(store.m, k)

			return nil
		}).AnyTimes()

	return kv, store
}

// entry is a KV entry of value.
func entry(
	ctrl *gomock.Controller,
	value []byte,
) jetstream.KeyValueEntry {
	e := jobmocks.NewMockKeyValueEntry(ctrl)
	e.EXPECT().Value().Return(value).AnyTimes()

	return e
}

// setup writes an operator setup in s.dir and returns the operator signing
// key and the public key of the account.
func (s *NatsRevokePublicTestSuite) setup() (nkeys.KeyPair, string) {
	_, err := natsauth.Operator(
		context.Background(), s.fs, s.dir, "osapi", config.NATS{},
	)
	s.Require().NoError(err)
	seed, err := avfs.ReadFile(s.fs, filepath.Join(s.dir, natsauth.OperatorSigningKeyFile))
	s.Require().NoError(err)
	signer, err := nkeys.FromSeed(seed)
	s.Require().NoError(err)
	acct, err := avfs.ReadFile(s.fs, filepath.Join(s.dir, natsauth.AccountPubFile))
	s.Require().NoError(err)

	return signer, strings.TrimSpace(string(acct))
}

// startServer runs a NATS server on the resolver file of the setup in
// s.dir.
func (s *NatsRevokePublicTestSuite) startServer() *natsserver.Server {
	opts, err := natsserver.ProcessConfigFile(filepath.Join(s.dir, "resolver.conf"))
	s.Require().NoError(err)
	opts.Host, opts.Port, opts.NoLog, opts.NoSigs = "127.0.0.1", -1, true, true
	srv, err := natsserver.NewServer(opts)
	s.Require().NoError(err)
	go srv.Start()
	s.Require().True(srv.ReadyForConnections(5 * time.Second))
	s.T().Cleanup(srv.Shutdown)

	return srv
}

// userKey is the public key of the user in the creds file name of s.dir.
func (s *NatsRevokePublicTestSuite) userKey(
	name string,
) string {
	raw, err := avfs.ReadFile(s.fs, filepath.Join(s.dir, name))
	s.Require().NoError(err)
	token, err := jwt.ParseDecoratedJWT(raw)
	s.Require().NoError(err)
	uc, err := jwt.DecodeUserClaims(token)
	s.Require().NoError(err)

	return uc.Subject
}

// connect connects to srv as the user of the creds file name of s.dir.
func (s *NatsRevokePublicTestSuite) connect(
	srv *natsserver.Server,
	name string,
	opts ...nats.Option,
) (*nats.Conn, error) {
	return nats.Connect(srv.ClientURL(), append(opts,
		nats.UserCredentials(filepath.Join(s.dir, name)),
		nats.NoReconnect(),
	)...)
}

func (s *NatsRevokePublicTestSuite) TestRevokeCutsTheUserAtOnce() {
	signer, _ := s.setup()
	srv := s.startServer()
	kv, _ := s.memKV()

	closed := make(chan struct{})
	agent, err := s.connect(srv, "agent.creds",
		nats.ClosedHandler(func(*nats.Conn) { close(closed) }))
	s.Require().NoError(err, "the agent connects before its revocation")
	defer agent.Close()
	system, err := s.connect(srv, natsauth.SystemCredsFile)
	s.Require().NoError(err, "the controller's system user connects")
	defer system.Close()

	r := natsrevoke.New(natsrevoke.Config{
		KV:             kv,
		Bucket:         bucket,
		Conn:           natsrevoke.Requester{NC: system},
		Signer:         signer,
		AccountJWTFile: filepath.Join(s.dir, natsauth.AccountJWTFile),
		Logger:         slog.New(slog.DiscardHandler),
	})
	s.now = time.Now()
	s.Require().NoError(r.Revoke(context.Background(), s.userKey("agent.creds")))

	select {
	case <-closed:
	case <-time.After(5 * time.Second):
		s.Fail("the server did not close the revoked user's connection")
	}
	_, err = s.connect(srv, "agent.creds")
	s.Error(err, "the server refuses the revoked user")
	other, err := s.connect(srv, "controller.creds")
	s.Require().NoError(err, "a user that is not revoked still connects")
	other.Close()

	// A server that starts again has the account of its resolver file, so
	// the revoked user connects until the revoker sends the account again.
	srv.Shutdown()
	srv = s.startServer()
	again, err := s.connect(srv, "agent.creds")
	s.Require().NoError(err, "a new server has no revocations")
	defer again.Close()
	system2, err := s.connect(srv, natsauth.SystemCredsFile)
	s.Require().NoError(err)
	defer system2.Close()
	r2 := natsrevoke.New(natsrevoke.Config{
		KV:             kv,
		Bucket:         bucket,
		Conn:           natsrevoke.Requester{NC: system2},
		Signer:         signer,
		AccountJWTFile: filepath.Join(s.dir, natsauth.AccountJWTFile),
		Logger:         slog.New(slog.DiscardHandler),
	})
	s.Require().NoError(r2.Send(context.Background()))
	_, err = s.connect(srv, "agent.creds")
	s.Error(err, "the revocation is sent again from the store")
}

// reply is a NATS reply with data.
func reply(
	data string,
) *nats.Msg {
	return &nats.Msg{Data: []byte(data)}
}

const ok = `{"data":{"code":200,"message":"jwt updated"}}`

func (s *NatsRevokePublicTestSuite) TestSend() {
	user, _ := nkeys.CreateUser()
	userPub, _ := user.PublicKey()
	key := natsrevoke.KVPrefix() + userPub
	at := []byte("1800000000")
	errBoom := errors.New("boom")

	tests := []struct {
		name string
		// kv sets the store's expectations.
		kv func(kv *mocks.MockKV)
		// conn sets the expectations of the system user's connection.
		conn    func(c *mocks.MockConn)
		ttl     time.Duration
		signer  func(op nkeys.KeyPair) nkeys.KeyPair
		account func(path string)
		wantErr string
	}{
		{
			name: "no revocation and none sent sends nothing",
			kv: func(kv *mocks.MockKV) {
				kv.EXPECT().KVKeys(gomock.Any(), bucket, gomock.Any()).
					Return(nil, jetstream.ErrNoKeysFound)
			},
		},
		{
			name: "a revocation is sent with the account",
			kv: func(kv *mocks.MockKV) {
				kv.EXPECT().KVKeys(gomock.Any(), bucket, gomock.Any()).Return([]string{key}, nil)
				kv.EXPECT().KVGet(gomock.Any(), bucket, key).Return(entry(s.ctrl, at), nil)
			},
			conn: func(c *mocks.MockConn) {
				c.EXPECT().RequestWithContext(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, subj string, data []byte) (*nats.Msg, error) {
						ac, err := jwt.DecodeAccountClaims(string(data))
						s.Require().NoError(err)
						s.Equal(natsrevoke.ClaimsUpdateSubject(ac.Subject), subj)
						s.True(ac.Revocations.IsRevoked(userPub, time.Unix(1_800_000_000, 0)))

						return reply(ok), nil
					})
			},
		},
		{
			name: "a revocation older than the TTL is removed and not sent",
			ttl:  time.Hour,
			kv: func(kv *mocks.MockKV) {
				kv.EXPECT().KVKeys(gomock.Any(), bucket, gomock.Any()).Return([]string{key}, nil)
				kv.EXPECT().KVGet(gomock.Any(), bucket, key).
					Return(entry(s.ctrl, []byte("1799990000")), nil)
				kv.EXPECT().KVDelete(gomock.Any(), bucket, key).Return(nil)
			},
		},
		{
			name: "a list that fails",
			kv: func(kv *mocks.MockKV) {
				kv.EXPECT().KVKeys(gomock.Any(), bucket, gomock.Any()).Return(nil, errBoom)
			},
			wantErr: "list the revocations: boom",
		},
		{
			name: "a read that fails",
			kv: func(kv *mocks.MockKV) {
				kv.EXPECT().KVKeys(gomock.Any(), bucket, gomock.Any()).Return([]string{key}, nil)
				kv.EXPECT().KVGet(gomock.Any(), bucket, key).Return(nil, errBoom)
			},
			wantErr: "read " + key + ": boom",
		},
		{
			name: "a record that is not a time",
			kv: func(kv *mocks.MockKV) {
				kv.EXPECT().KVKeys(gomock.Any(), bucket, gomock.Any()).Return([]string{key}, nil)
				kv.EXPECT().KVGet(gomock.Any(), bucket, key).
					Return(entry(s.ctrl, []byte("soon")), nil)
			},
			wantErr: "invalid syntax",
		},
		{
			name: "a remove that fails",
			ttl:  time.Hour,
			kv: func(kv *mocks.MockKV) {
				kv.EXPECT().KVKeys(gomock.Any(), bucket, gomock.Any()).Return([]string{key}, nil)
				kv.EXPECT().KVGet(gomock.Any(), bucket, key).
					Return(entry(s.ctrl, []byte("1799990000")), nil)
				kv.EXPECT().KVDelete(gomock.Any(), bucket, key).Return(errBoom)
			},
			wantErr: "remove " + key + ": boom",
		},
		{
			name: "an account file that is not there",
			kv: func(kv *mocks.MockKV) {
				kv.EXPECT().KVKeys(gomock.Any(), bucket, gomock.Any()).Return([]string{key}, nil)
				kv.EXPECT().KVGet(gomock.Any(), bucket, key).Return(entry(s.ctrl, at), nil)
			},
			account: func(path string) { s.Require().NoError(s.fs.Remove(path)) },
			wantErr: "read the account",
		},
		{
			name: "an account file that is not a JWT",
			kv: func(kv *mocks.MockKV) {
				kv.EXPECT().KVKeys(gomock.Any(), bucket, gomock.Any()).Return([]string{key}, nil)
				kv.EXPECT().KVGet(gomock.Any(), bucket, key).Return(entry(s.ctrl, at), nil)
			},
			account: func(path string) { s.Require().NoError(avfs.WriteFile(s.fs, path, []byte("x"), 0o600)) },
			wantErr: "decode the account",
		},
		{
			name: "a signer that is not an operator key",
			kv: func(kv *mocks.MockKV) {
				kv.EXPECT().KVKeys(gomock.Any(), bucket, gomock.Any()).Return([]string{key}, nil)
				kv.EXPECT().KVGet(gomock.Any(), bucket, key).Return(entry(s.ctrl, at), nil)
			},
			signer: func(nkeys.KeyPair) nkeys.KeyPair {
				u, _ := nkeys.CreateUser()

				return u
			},
			wantErr: "sign the account",
		},
		{
			name: "a request that fails",
			kv: func(kv *mocks.MockKV) {
				kv.EXPECT().KVKeys(gomock.Any(), bucket, gomock.Any()).Return([]string{key}, nil)
				kv.EXPECT().KVGet(gomock.Any(), bucket, key).Return(entry(s.ctrl, at), nil)
			},
			conn: func(c *mocks.MockConn) {
				c.EXPECT().RequestWithContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errBoom)
			},
			wantErr: "send the account: boom",
		},
		{
			name: "a reply that is not JSON",
			kv: func(kv *mocks.MockKV) {
				kv.EXPECT().KVKeys(gomock.Any(), bucket, gomock.Any()).Return([]string{key}, nil)
				kv.EXPECT().KVGet(gomock.Any(), bucket, key).Return(entry(s.ctrl, at), nil)
			},
			conn: func(c *mocks.MockConn) {
				c.EXPECT().RequestWithContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(reply("nope"), nil)
			},
			wantErr: "the server's reply",
		},
		{
			name: "a reply with an error",
			kv: func(kv *mocks.MockKV) {
				kv.EXPECT().KVKeys(gomock.Any(), bucket, gomock.Any()).Return([]string{key}, nil)
				kv.EXPECT().KVGet(gomock.Any(), bucket, key).Return(entry(s.ctrl, at), nil)
			},
			conn: func(c *mocks.MockConn) {
				c.EXPECT().RequestWithContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(reply(`{"error":{"code":500,"description":"not trusted"}}`), nil)
			},
			wantErr: "the server refused the account: not trusted",
		},
		{
			name: "a reply that did not take the account",
			kv: func(kv *mocks.MockKV) {
				kv.EXPECT().KVKeys(gomock.Any(), bucket, gomock.Any()).Return([]string{key}, nil)
				kv.EXPECT().KVGet(gomock.Any(), bucket, key).Return(entry(s.ctrl, at), nil)
			},
			conn: func(c *mocks.MockConn) {
				c.EXPECT().RequestWithContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(reply(`{"data":{"code":304,"message":"jwt update skipped"}}`), nil)
			},
			wantErr: "the server did not take the account",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.dir = s.T().TempDir()
			signer, _ := s.setup()
			if tt.signer != nil {
				signer = tt.signer(signer)
			}
			path := filepath.Join(s.dir, natsauth.AccountJWTFile)
			if tt.account != nil {
				tt.account(path)
			}
			kv := mocks.NewMockKV(s.ctrl)
			tt.kv(kv)
			conn := mocks.NewMockConn(s.ctrl)
			if tt.conn != nil {
				tt.conn(conn)
			}
			r := natsrevoke.New(natsrevoke.Config{
				KV: kv, Bucket: bucket, Conn: conn, Signer: signer,
				AccountJWTFile: path, TTL: tt.ttl,
				Logger: slog.New(slog.DiscardHandler),
			})

			err := r.Send(context.Background())
			if tt.wantErr != "" {
				s.Require().Error(err)
				s.Contains(err.Error(), tt.wantErr)

				return
			}
			s.NoError(err)
		})
	}
}

func (s *NatsRevokePublicTestSuite) TestSendAgain() {
	s.Run("the same revocations send the same JWT, and none after some sends the account", func() {
		signer, _ := s.setup()
		user, _ := nkeys.CreateUser()
		userPub, _ := user.PublicKey()
		kv, store := s.memKV()
		store.m[natsrevoke.KVPrefix()+userPub] = []byte("1800000000")
		conn := mocks.NewMockConn(s.ctrl)
		var sent []string
		conn.EXPECT().RequestWithContext(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, _ string, data []byte) (*nats.Msg, error) {
				sent = append(sent, string(data))

				return reply(ok), nil
			}).Times(3)
		r := natsrevoke.New(natsrevoke.Config{
			KV: kv, Bucket: bucket, Conn: conn, Signer: signer,
			AccountJWTFile: filepath.Join(s.dir, natsauth.AccountJWTFile),
			TTL:            time.Hour,
			Logger:         slog.New(slog.DiscardHandler),
		})

		s.Require().NoError(r.Send(context.Background()))
		s.Require().NoError(r.Send(context.Background()))
		s.Equal(sent[0], sent[1], "the same account is the same JWT")

		// The revocation expires: the account is sent with none, so the
		// server stops refusing the user's later JWTs.
		s.now = s.now.Add(2 * time.Hour)
		s.Require().NoError(r.Send(context.Background()))
		ac, err := jwt.DecodeAccountClaims(sent[2])
		s.Require().NoError(err)
		s.Empty(ac.Revocations)
		s.Require().NoError(r.Send(context.Background()), "none left, and none to send again")
	})
}

func (s *NatsRevokePublicTestSuite) TestRevoke() {
	user, _ := nkeys.CreateUser()
	userPub, _ := user.PublicKey()

	tests := []struct {
		name    string
		userKey string
		kv      func(kv *mocks.MockKV)
		conn    func(c *mocks.MockConn)
		wantErr string
	}{
		{
			name:    "a key that is not a user key",
			userKey: "ABAD",
			wantErr: `"ABAD" is not a user public key (U...)`,
		},
		{
			name:    "a record that does not write",
			userKey: userPub,
			kv: func(kv *mocks.MockKV) {
				kv.EXPECT().
					KVPut(gomock.Any(), bucket, natsrevoke.KVPrefix()+userPub, []byte("1800000000")).
					Return(uint64(0), errors.New("boom"))
			},
			wantErr: "record the revocation of " + userPub + ": boom",
		},
		{
			name:    "a send that fails is not an error: Run sends it again",
			userKey: userPub,
			kv: func(kv *mocks.MockKV) {
				kv.EXPECT().
					KVPut(gomock.Any(), bucket, gomock.Any(), gomock.Any()).
					Return(uint64(1), nil)
				kv.EXPECT().
					KVKeys(gomock.Any(), bucket, gomock.Any()).
					Return(nil, errors.New("boom"))
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			kv := mocks.NewMockKV(s.ctrl)
			if tt.kv != nil {
				tt.kv(kv)
			}
			conn := mocks.NewMockConn(s.ctrl)
			r := natsrevoke.New(natsrevoke.Config{
				KV: kv, Bucket: bucket, Conn: conn,
				Logger: slog.New(slog.DiscardHandler),
			})

			err := r.Revoke(context.Background(), tt.userKey)
			if tt.wantErr != "" {
				s.Require().Error(err)
				s.Equal(tt.wantErr, err.Error())

				return
			}
			s.NoError(err)
		})
	}
}

func (s *NatsRevokePublicTestSuite) TestRun() {
	s.Run("sends at once and on each interval until ctx is done", func() {
		kv := mocks.NewMockKV(s.ctrl)
		calls := make(chan struct{}, 10)
		kv.EXPECT().KVKeys(gomock.Any(), bucket, gomock.Any()).
			DoAndReturn(func(context.Context, string, ...string) ([]string, error) {
				calls <- struct{}{}

				return nil, errors.New("boom")
			}).MinTimes(2)
		r := natsrevoke.New(natsrevoke.Config{
			KV: kv, Bucket: bucket, Interval: 10 * time.Millisecond,
			Logger: slog.New(slog.DiscardHandler),
		})
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error)
		go func() { done <- r.Run(ctx) }()
		<-calls
		<-calls
		cancel()
		s.NoError(<-done)
	})

	s.Run("a zero interval is a minute", func() {
		kv := mocks.NewMockKV(s.ctrl)
		kv.EXPECT().KVKeys(gomock.Any(), bucket, gomock.Any()).Return(nil, jetstream.ErrNoKeysFound)
		r := natsrevoke.New(natsrevoke.Config{
			KV: kv, Bucket: bucket, Logger: slog.New(slog.DiscardHandler),
		})
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		s.NoError(r.Run(ctx), "one send, then no tick before ctx is done")
	})
}

func (s *NatsRevokePublicTestSuite) TestRequester() {
	errBoom := errors.New("boom")

	tests := []struct {
		name     string
		deadline bool
		nc       func(nc *mocks.MockConnector)
		wantData string
		wantErr  string
	}{
		{
			name: "the first reply, with no deadline of its own",
			nc: func(nc *mocks.MockConnector) {
				var handler nats.MsgHandler
				nc.EXPECT().Subscribe(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ string, h nats.MsgHandler) (*nats.Subscription, error) {
						handler = h

						return &nats.Subscription{}, nil
					})
				nc.EXPECT().PublishMsg(gomock.Any()).DoAndReturn(func(m *nats.Msg) error {
					s.Equal("subj", m.Subject)
					s.NotEmpty(m.Reply)
					handler(reply("one"))
					handler(reply("two"))

					return nil
				})
			},
			wantData: "one",
		},
		{
			name: "a subscribe that fails",
			nc: func(nc *mocks.MockConnector) {
				nc.EXPECT().Subscribe(gomock.Any(), gomock.Any()).Return(nil, errBoom)
			},
			wantErr: "subscribe to the reply: boom",
		},
		{
			name: "a publish that fails",
			nc: func(nc *mocks.MockConnector) {
				nc.EXPECT().Subscribe(gomock.Any(), gomock.Any()).Return(&nats.Subscription{}, nil)
				nc.EXPECT().PublishMsg(gomock.Any()).Return(errBoom)
			},
			wantErr: "publish the request: boom",
		},
		{
			name:     "no reply before the deadline",
			deadline: true,
			nc: func(nc *mocks.MockConnector) {
				nc.EXPECT().Subscribe(gomock.Any(), gomock.Any()).Return(&nats.Subscription{}, nil)
				nc.EXPECT().PublishMsg(gomock.Any()).Return(nil)
			},
			wantErr: "wait for the reply: context deadline exceeded",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			nc := mocks.NewMockConnector(s.ctrl)
			tt.nc(nc)
			ctx := context.Background()
			if tt.deadline {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 10*time.Millisecond)
				defer cancel()
			}

			msg, err := natsrevoke.Requester{NC: nc}.RequestWithContext(ctx, "subj", []byte("x"))
			if tt.wantErr != "" {
				s.Require().Error(err)
				s.Equal(tt.wantErr, err.Error())

				return
			}
			s.Require().NoError(err)
			s.Equal(tt.wantData, string(msg.Data))
		})
	}
}

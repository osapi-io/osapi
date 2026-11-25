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

package join_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"math/big"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"

	natsserver "github.com/nats-io/nats-server/v2/server"
	natstest "github.com/nats-io/nats-server/v2/test"
	"github.com/osapi-io/osapi/internal/config"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/cli/commands/join"
	"github.com/osapi-io/osapi/internal/controller/api"
	"github.com/osapi-io/osapi/internal/controller/jointoken"
)

// entry is a value of the bucket.
type entry struct {
	jetstream.KeyValueEntry
	value []byte
}

func (e entry) Value() []byte    { return e.value }
func (e entry) Revision() uint64 { return 1 }

// bucket is the enrollment bucket in memory.
type bucket struct {
	data map[string][]byte
	err  error
}

func (b *bucket) KVGet(_ context.Context, _ string, key string) (jetstream.KeyValueEntry, error) {
	v, ok := b.data[key]
	if !ok {
		return nil, jetstream.ErrKeyNotFound
	}

	return entry{value: v}, nil
}

func (b *bucket) KVCreate(_ context.Context, _ string, key string, value []byte) (uint64, error) {
	b.data[key] = value

	return 1, b.err
}

func (b *bucket) KVUpdate(context.Context, string, string, []byte, uint64) (uint64, error) {
	return 2, nil
}

func (b *bucket) KVDelete(_ context.Context, _ string, key string) error {
	delete(b.data, key)

	return nil
}

func (b *bucket) KVKeys(context.Context, string, ...string) ([]string, error) {
	if b.err != nil {
		return nil, b.err
	}
	out := []string{}
	for k := range b.data {
		out = append(out, k)
	}

	return out, nil
}

type MethodsPublicTestSuite struct {
	suite.Suite

	kv      *bucket
	dialErr error
}

func (s *MethodsPublicTestSuite) SetupTest() {
	s.kv = &bucket{data: map[string][]byte{}}
	s.dialErr = nil
}

// deps is a controller that serves HTTPS with its self-signed certificate.
func (s *MethodsPublicTestSuite) deps() join.Deps {
	var cfg config.Config
	cfg.Controller.API.URL = "https://osapi.example.com:8080/"
	cfg.Controller.API.TLS.Enabled = true
	cfg.Controller.PKI.KeyDir = s.T().TempDir()
	_, err := api.TLSConfig(cfg.Controller.API.TLS, cfg.Controller.PKI.KeyDir)
	s.Require().NoError(err)
	cfg.NATS.Enrollment.Bucket = "enrollment"

	return join.Deps{
		Config: cfg,
		Now:    func() time.Time { return time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC) },
		Dial: func(context.Context) (jointoken.KV, func(), error) {
			if s.dialErr != nil {
				return nil, nil, s.dialErr
			}

			return s.kv, func() {}, nil
		},
	}
}

func (s *MethodsPublicTestSuite) TestCreate() {
	ctx := context.Background()
	in := join.CreateInput{
		Expires:   "1h",
		Uses:      2,
		Labels:    map[string]string{"group": "web"},
		Transport: "https",
	}

	d := s.deps()
	got, err := join.Create(ctx, d, in)
	s.Require().NoError(err)
	s.True(strings.HasPrefix(got.Token, "osapi_jt_"+got.ID+"."))
	u, err := url.Parse(got.Join)
	s.Require().NoError(err)
	s.Equal("osapi", u.Scheme)
	s.Equal("osapi.example.com:8080", u.Host)
	s.Equal(got.Token, u.Query().Get("token"))
	fp, err := api.CertFingerprint(d.Config.Controller.API.TLS, d.Config.Controller.PKI.KeyDir)
	s.Require().NoError(err)
	s.Equal(fp, u.Query().Get("fp"))
	s.Equal(2, got.Uses)

	noTLS := s.deps()
	noTLS.Config.Controller.API.TLS.Enabled = false
	_, err = join.Create(ctx, noTLS, in)
	s.ErrorContains(err, "controller.api.tls.enabled")

	noURL := s.deps()
	noURL.Config.Controller.API.URL = ""
	_, err = join.Create(ctx, noURL, in)
	s.ErrorContains(err, "controller.api.url")

	noCert := s.deps()
	noCert.Config.Controller.PKI.KeyDir = s.T().TempDir()
	_, err = join.Create(ctx, noCert, in)
	s.ErrorContains(err, "read the API's certificate")

	s.dialErr = errors.New("no nats")
	_, err = join.Create(ctx, s.deps(), in)
	s.ErrorContains(err, "no nats")
	s.dialErr = nil

	s.kv.err = errors.New("kv down")
	_, err = join.Create(ctx, s.deps(), in)
	s.ErrorContains(err, "store the join token")
}

func (s *MethodsPublicTestSuite) TestCreateNATS() {
	ctx := context.Background()
	in := join.CreateInput{Expires: "1h", Uses: 1, Transport: "nats"}
	natsDeps := func(url string) join.Deps {
		d := s.deps()
		d.Config.NATS.Server.URL = url
		d.Config.Controller.NATS.Namespace = "osapi"
		d.Config.NATS.Server.Auth.Users = []config.NATSServerUser{
			{Username: "join", Password: "jp", Role: "join"},
		}

		return d
	}

	s.Run("a nats:// server: the join user and the namespace, no fingerprint", func() {
		got, err := join.Create(ctx, natsDeps("nats://hub:4222"), in)
		s.Require().NoError(err)
		u, err := url.Parse(got.Join)
		s.Require().NoError(err)
		s.Equal("nats", u.Scheme)
		s.Equal("hub:4222", u.Host)
		s.Equal(got.Token, u.Query().Get("token"))
		s.Equal("join", u.Query().Get("user"))
		s.Equal("jp", u.Query().Get("pass"))
		s.Equal("osapi", u.Query().Get("ns"))
		s.Empty(u.Query().Get("fp"))
	})

	s.Run("a tls:// server: the fingerprint of its certificate", func() {
		defer join.SetNATSFingerprint(func(url string, auth []nats.Option) (string, error) {
			s.Equal("tls://hub:4222", url)
			s.Len(auth, 1, "it connects as the join user")

			return "SHA256:ab", nil
		})()
		got, err := join.Create(ctx, natsDeps("tls://hub:4222"), in)
		s.Require().NoError(err)
		u, _ := url.Parse(got.Join)
		s.Equal("SHA256:ab", u.Query().Get("fp"))
	})

	s.Run("a tls:// server that cannot be read makes no token", func() {
		defer join.SetNATSFingerprint(func(string, []nats.Option) (string, error) {
			return "", errors.New("refused")
		})()
		before, err := join.List(ctx, s.deps(), join.ListInput{})
		s.Require().NoError(err)
		_, err = join.Create(ctx, natsDeps("tls://hub:4222"), in)
		s.ErrorContains(err, "refused")
		after, err := join.List(ctx, s.deps(), join.ListInput{})
		s.Require().NoError(err)
		s.Len(after.Items, len(before.Items))
	})

	s.Run("operator mode: a bootstrap user that expires with the token", func() {
		d := natsDeps("nats://hub:4222")
		d.Config.Controller.Join.SigningKeyFile, d.Config.Controller.Join.Account = signingKey(
			s.T(),
		)
		got, err := join.Create(ctx, d, in)
		s.Require().NoError(err)
		u, _ := url.Parse(got.Join)
		s.Empty(u.Query().Get("user"), "no static user")
		creds, err := base64.RawURLEncoding.DecodeString(u.Query().Get("creds"))
		s.Require().NoError(err)
		userJWT, err := jwt.ParseDecoratedJWT(creds)
		s.Require().NoError(err)
		uc, err := jwt.DecodeUserClaims(userJWT)
		s.Require().NoError(err)
		s.Equal(jwt.StringList{"osapi.join.request"}, uc.Pub.Allow)
		s.Equal(got.ExpiresAt.Unix(), uc.Expires)
		_, err = jwt.ParseDecoratedNKey(creds)
		s.NoError(err, "the bootstrap user's seed is in the string")
	})

	s.Run("operator mode with a signing key that does not load makes no token", func() {
		d := natsDeps("nats://hub:4222")
		d.Config.Controller.Join.SigningKeyFile = "/no/such/key"
		_, err := join.Create(ctx, d, in)
		s.ErrorContains(err, "read the signing key")
	})

	s.Run("no address of NATS is the error", func() {
		_, err := join.Create(ctx, natsDeps(""), in)
		s.ErrorContains(err, "controller.join.url or nats.server.url")
	})

	s.Run("a real server that is not there is the error", func() {
		_, err := join.Create(ctx, natsDeps("tls://127.0.0.1:1"), in)
		s.ErrorContains(err, "read the certificate of the NATS server")
	})

	s.Run("a real server: its certificate", func() {
		srv := tlsServer(s.T())
		defer srv.Shutdown()
		d := natsDeps(strings.Replace(srv.ClientURL(), "nats://", "tls://", 1))
		d.Config.NATS.Server.Auth.Users = nil
		got, err := join.Create(ctx, d, in)
		s.Require().NoError(err)
		u, _ := url.Parse(got.Join)
		s.True(strings.HasPrefix(u.Query().Get("fp"), "SHA256:"))
	})
}

func (s *MethodsPublicTestSuite) TestListAndRevoke() {
	ctx := context.Background()
	d := s.deps()
	got, err := join.List(ctx, d, join.ListInput{})
	s.Require().NoError(err)
	s.Empty(got.Items)

	made, err := join.Create(ctx, d, join.CreateInput{Expires: "1h", Uses: 1, Transport: "https"})
	s.Require().NoError(err)
	got, err = join.List(ctx, d, join.ListInput{})
	s.Require().NoError(err)
	s.Require().Len(got.Items, 1)
	s.Equal(made.ID, got.Items[0].ID)

	r, err := join.Revoke(ctx, d, join.RevokeInput{ID: made.ID})
	s.Require().NoError(err)
	s.Equal(made.ID, r.ID)
	_, err = join.Revoke(ctx, d, join.RevokeInput{ID: made.ID})
	s.ErrorIs(err, jointoken.ErrInvalid)

	s.dialErr = errors.New("no nats")
	_, err = join.List(ctx, d, join.ListInput{})
	s.ErrorContains(err, "no nats")
	_, err = join.Revoke(ctx, d, join.RevokeInput{ID: "x"})
	s.ErrorContains(err, "no nats")
	s.dialErr = nil

	s.kv.err = errors.New("kv down")
	_, err = join.List(ctx, d, join.ListInput{})
	s.ErrorContains(err, "list the join tokens")
}

func TestMethodsPublicTestSuite(t *testing.T) {
	suite.Run(t, new(MethodsPublicTestSuite))
}

// tlsServer is a NATS server with TLS, on a free port, with a self-signed
// certificate.
func tlsServer(
	t *testing.T,
) *natsserver.Server {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotAfter:     time.Now().Add(time.Hour),
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	opts := natstest.DefaultTestOptions
	opts.Port = -1
	opts.TLS = true
	opts.TLSConfig = &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key}},
	}

	return natstest.RunServer(&opts)
}

// signingKey is a file with a new account signing key seed, and the public
// key of its account.
func signingKey(
	t *testing.T,
) (string, string) {
	t.Helper()
	acct, _ := nkeys.CreateAccount()
	pub, _ := acct.PublicKey()
	seed, _ := acct.Seed()
	file := filepath.Join(t.TempDir(), "signing.nk")
	if err := os.WriteFile(file, seed, 0o600); err != nil {
		t.Fatal(err)
	}

	return file, pub
}

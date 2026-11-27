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
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"math/big"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/nats-io/jwt/v2"
	natstest "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/agent"
	"github.com/osapi-io/osapi/internal/agent/identity"
	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/pki"
)

type NATSJoinPublicTestSuite struct {
	suite.Suite
	fs  avfs.VFS
	cfg config.Config
}

func TestNATSJoinPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(NATSJoinPublicTestSuite))
}

func (s *NATSJoinPublicTestSuite) SetupSubTest() {
	s.fs = memfs.New()
	s.Require().NoError(s.fs.MkdirAll("/pki", 0o700))
	s.cfg = config.Config{}
	s.cfg.Agent.PKI.KeyDir = "/pki"
	s.cfg.Agent.NATS.Namespace = "osapi"
	s.cfg.Agent.Join = "nats://hub:4222?token=osapi_jt_t1.secret&user=join&pass=jp"
}

// id is a host with machine ID m-1.
func id(avfs.VFS, string) (*identity.Identity, error) {
	return &identity.Identity{MachineID: "m-1", Hostname: "web-01"}, nil
}

// reply is a request func that checks the request and answers r.
func (s *NATSJoinPublicTestSuite) reply(
	r pki.JoinReply,
	wantSubject string,
) func(string, []nats.Option, string, []byte) ([]byte, error) {
	return func(_ string, _ []nats.Option, subject string, data []byte) ([]byte, error) {
		s.Equal(wantSubject, subject)
		var req pki.JoinRequest
		s.Require().NoError(json.Unmarshal(data, &req))
		s.True(req.Verify(), "the request is signed by its key")
		s.Equal("m-1", req.MachineID)

		return json.Marshal(r)
	}
}

var admitted = pki.JoinReply{
	ControllerPublicKey: make([]byte, 32),
	NATS:                pki.JoinNATS{Username: "agent", Password: "ap"},
	Labels:              map[string]string{"group": "web", "dc": "east"},
}

func (s *NATSJoinPublicTestSuite) TestJoinNATS() {
	s.Run("a join string that is not for NATS leaves the config", func() {
		s.cfg.Agent.Join = "osapi://ctl:8443?token=t&fp=SHA256:00"
		got, err := agent.ExportJoinNATS(context.Background(), slog.Default(), s.fs, s.cfg)
		s.Require().NoError(err)
		s.Equal(s.cfg, got)
	})

	s.Run("a join that admits keeps the user, and the next start reads it", func() {
		defer agent.SetJoinNATSFns(id, s.reply(admitted, "osapi.join.request"))()

		got, err := agent.ExportJoinNATS(context.Background(), slog.Default(), s.fs, s.cfg)

		s.Require().NoError(err)
		s.Equal("hub", got.Agent.NATS.Host)
		s.Equal(4222, got.Agent.NATS.Port)
		s.Equal(
			config.NATSAuth{Type: "user_pass", Username: "agent", Password: "ap"},
			got.Agent.NATS.Auth,
		)
		s.False(got.Agent.NATS.TLS.Enabled)
		s.Empty(got.Agent.Join, "the join is spent")
		s.Equal(
			map[string]string{"group": "web", "dc": "east"},
			got.Agent.Labels,
			"the token's labels",
		)
		info, err := s.fs.Stat(filepath.Join("/pki", agent.NATSJoinFile))
		s.Require().NoError(err)
		s.Equal("-rw-------", info.Mode().Perm().String())
		_, err = s.fs.Stat(filepath.Join("/pki", agent.EnrolledControllerKey))
		s.NoError(err, "the agent is admitted")

		defer agent.SetJoinNATSFns(
			nil,
			func(string, []nats.Option, string, []byte) ([]byte, error) {
				s.Fail("a second start does not join again")
				return nil, nil
			},
		)()
		again, err := agent.ExportJoinNATS(context.Background(), slog.Default(), s.fs, s.cfg)
		s.Require().NoError(err)
		s.Equal(got.Agent.NATS, again.Agent.NATS)
	})

	s.Run("a tls join pins the server, and trusts its certificate after", func() {
		cert := selfSigned(s.T())
		sum := sha256.Sum256(cert.Raw)
		s.cfg.Agent.Join = "tls://hub:4222?token=t&ns=prod&fp=SHA256:" + hex.EncodeToString(sum[:])
		defer agent.SetJoinNATSFns(
			id,
			func(_ string, opts []nats.Option, subject string, _ []byte) ([]byte, error) {
				s.Equal("prod.join.request", subject)
				o := nats.GetDefaultOptions()
				for _, opt := range opts {
					s.Require().NoError(opt(&o))
				}
				verify := o.TLSConfig.VerifyConnection
				s.Error(verify(tls.ConnectionState{}), "no certificate")
				s.Require().
					NoError(verify(tls.ConnectionState{PeerCertificates: []*x509.Certificate{cert}}))

				return json.Marshal(pki.JoinReply{ControllerPublicKey: make([]byte, 32)})
			},
		)()

		got, err := agent.ExportJoinNATS(context.Background(), slog.Default(), s.fs, s.cfg)

		s.Require().NoError(err)
		s.Equal("prod", got.Agent.NATS.Namespace)
		s.Equal("none", got.Agent.NATS.Auth.Type)
		s.True(got.Agent.NATS.TLS.Enabled)
		s.Equal("/pki/nats-join-ca.pem", got.Agent.NATS.TLS.CAFile)
	})

	s.Run("an empty ns in the string wins over the config's namespace", func() {
		cert := selfSigned(s.T())
		sum := sha256.Sum256(cert.Raw)
		s.cfg.Agent.NATS.Namespace = "osapi"
		defer func() { s.cfg.Agent.NATS.Namespace = "" }()
		s.cfg.Agent.Join = "tls://hub:4222?token=t&ns=&fp=SHA256:" + hex.EncodeToString(sum[:])
		defer agent.SetJoinNATSFns(
			id,
			func(_ string, _ []nats.Option, subject string, _ []byte) ([]byte, error) {
				s.Equal("join.request", subject)

				return json.Marshal(pki.JoinReply{ControllerPublicKey: make([]byte, 32)})
			},
		)()

		got, err := agent.ExportJoinNATS(context.Background(), slog.Default(), s.fs, s.cfg)

		s.Require().NoError(err)
		s.Empty(got.Agent.NATS.Namespace)
	})

	// joinTLS joins over tls with a server that shows chain, and returns the
	// agent's config after.
	joinTLS := func(chain []*x509.Certificate) config.Config {
		sum := sha256.Sum256(chain[0].Raw)
		s.cfg.Agent.Join = "tls://10.0.0.1:4222?token=t&fp=SHA256:" + hex.EncodeToString(sum[:])
		defer agent.SetJoinNATSFns(
			id,
			func(_ string, opts []nats.Option, _ string, _ []byte) ([]byte, error) {
				o := nats.GetDefaultOptions()
				for _, opt := range opts {
					s.Require().NoError(opt(&o))
				}
				s.Require().
					NoError(o.TLSConfig.VerifyConnection(tls.ConnectionState{PeerCertificates: chain}))

				return json.Marshal(pki.JoinReply{ControllerPublicKey: make([]byte, 32)})
			},
		)()
		got, err := agent.ExportJoinNATS(context.Background(), slog.Default(), s.fs, s.cfg)
		s.Require().NoError(err)

		return got
	}

	s.Run("a certificate that a trusted CA issued: the system's roots and its name", func() {
		ca, leaf := issued(s.T(), "hub.example.ts.net")
		roots := x509.NewCertPool()
		roots.AddCert(ca)
		defer agent.SetJoinRoots(roots)()

		// A server sends its chain: the leaf and the certificates above it.
		got := joinTLS([]*x509.Certificate{leaf, ca})

		s.True(got.Agent.NATS.TLS.Enabled)
		s.Empty(got.Agent.NATS.TLS.CAFile, "a renewal of the leaf changes nothing")
		s.Equal("hub.example.ts.net", got.Agent.NATS.TLS.ServerName,
			"the agent dials 10.0.0.1, and checks the name of the certificate")
		_, err := s.fs.Stat("/pki/nats-join-ca.pem")
		s.Error(err, "no CA is kept")
	})

	s.Run("a certificate that no trusted CA issued: the leaf it pinned", func() {
		defer agent.SetJoinRoots(x509.NewCertPool())()
		_, leaf := issued(s.T(), "hub.example.ts.net")

		got := joinTLS([]*x509.Certificate{leaf})

		s.Equal("/pki/nats-join-ca.pem", got.Agent.NATS.TLS.CAFile)
		s.Empty(got.Agent.NATS.TLS.ServerName)
	})

	s.Run("the agent's own server_name and ca_file win", func() {
		ca, leaf := issued(s.T(), "hub.example.ts.net")
		roots := x509.NewCertPool()
		roots.AddCert(ca)
		defer agent.SetJoinRoots(roots)()
		s.cfg.Agent.NATS.TLS = config.ClientTLS{ServerName: "nats.lab", CAFile: "/etc/osapi/ca.pem"}

		got := joinTLS([]*x509.Certificate{leaf})

		s.Equal("nats.lab", got.Agent.NATS.TLS.ServerName)
		s.Equal("/etc/osapi/ca.pem", got.Agent.NATS.TLS.CAFile)
	})

	failures := []struct {
		name    string
		join    string
		id      func(avfs.VFS, string) (*identity.Identity, error)
		req     func(string, []nats.Option, string, []byte) ([]byte, error)
		setup   func(fs avfs.VFS)
		wantErr string
	}{
		{
			name:    "a tls join with no fingerprint",
			join:    "tls://hub:4222?token=t",
			wantErr: "is not nats://",
		},
		{name: "a join with no port", join: "nats://hub?token=t", wantErr: "is not nats://"},
		{name: "a join with no token", join: "nats://hub:4222", wantErr: "is not nats://"},
		{
			name:    "creds that are not base64",
			join:    "nats://hub:4222?token=t&creds=%25%25",
			wantErr: "is not nats://",
		},
		{
			name: "a host with no machine ID",
			id: func(avfs.VFS, string) (*identity.Identity, error) {
				return nil, errors.New("no machine-id")
			},
			wantErr: "no machine-id",
		},
		{
			name:    "keys that cannot be made",
			setup:   func(fs avfs.VFS) { _ = fs.RemoveAll("/pki"); _ = avfs.WriteFile(fs, "/pki", nil, 0o600) },
			wantErr: "load or make the agent's keys",
		},
		{
			name: "a server that does not answer",
			req: func(string, []nats.Option, string, []byte) ([]byte, error) {
				return nil, nats.ErrTimeout
			},
			wantErr: "nats://hub:4222: nats: timeout",
		},
		{
			name: "a reply that is not JSON",
			req: func(string, []nats.Option, string, []byte) ([]byte, error) {
				return []byte("{"), nil
			},
			wantErr: "the reply of the join",
		},
		{
			name: "a token that does not admit",
			req: func(string, []nats.Option, string, []byte) ([]byte, error) {
				return json.Marshal(pki.JoinReply{Error: "join token has expired"})
			},
			wantErr: "join over NATS: join token has expired",
		},
		{
			name:    "a controller key that cannot be kept",
			setup:   func(fs avfs.VFS) { _ = fs.MkdirAll("/pki/"+agent.EnrolledControllerKey+"/x", 0o700) },
			wantErr: "join over NATS: ",
		},
		{
			name:    "a join that cannot be kept",
			setup:   func(fs avfs.VFS) { _ = fs.MkdirAll("/pki/"+agent.NATSJoinFile+"/x", 0o700) },
			wantErr: "join over NATS: ",
		},
	}
	for _, tt := range failures {
		s.Run(tt.name, func() {
			if tt.join != "" {
				s.cfg.Agent.Join = tt.join
			}
			if tt.setup != nil {
				tt.setup(s.fs)
			}
			i, r := tt.id, tt.req
			if i == nil {
				i = id
			}
			if r == nil {
				r = s.reply(admitted, "osapi.join.request")
			}
			defer agent.SetJoinNATSFns(i, r)()

			_, err := agent.ExportJoinNATS(context.Background(), slog.Default(), s.fs, s.cfg)

			s.ErrorContains(err, tt.wantErr)
		})
	}

	s.Run("operator mode: the bootstrap creds connect, and the agent keeps its own creds", func() {
		acct, _ := nkeys.CreateAccount()
		acctPub, _ := acct.PublicKey()
		boot, _ := nkeys.CreateUser()
		bootPub, _ := boot.PublicKey()
		bootSeed, _ := boot.Seed()
		bootJWT, _ := jwt.NewUserClaims(bootPub).Encode(acct)
		bootCreds, _ := jwt.FormatUserConfig(bootJWT, bootSeed)
		s.cfg.Agent.Join = "nats://hub:4222?token=t&creds=" + base64.RawURLEncoding.EncodeToString(
			bootCreds,
		)
		defer agent.SetJoinNATSFns(
			id,
			func(_ string, opts []nats.Option, _ string, data []byte) ([]byte, error) {
				o := nats.GetDefaultOptions()
				for _, opt := range opts {
					s.Require().NoError(opt(&o))
				}
				s.NotNil(o.UserJWT, "it connects with the bootstrap creds")
				var req pki.JoinRequest
				s.Require().NoError(json.Unmarshal(data, &req))
				s.True(strings.HasPrefix(req.NATSUserKey, "U"))
				uc := jwt.NewUserClaims(req.NATSUserKey)
				uc.IssuerAccount = acctPub
				agentJWT, _ := uc.Encode(acct)

				return json.Marshal(pki.JoinReply{
					ControllerPublicKey: make([]byte, 32),
					NATS:                pki.JoinNATS{JWT: agentJWT},
				})
			},
		)()

		got, err := agent.ExportJoinNATS(context.Background(), slog.Default(), s.fs, s.cfg)

		s.Require().NoError(err)
		s.Equal(config.NATSAuth{Type: "creds", CredsFile: "/pki/nats.creds"}, got.Agent.NATS.Auth)
		info, err := s.fs.Stat("/pki/nats.creds")
		s.Require().NoError(err)
		s.Equal("-rw-------", info.Mode().Perm().String())
		creds, _ := avfs.ReadFile(s.fs, "/pki/nats.creds")
		_, err = jwt.ParseDecoratedNKey(creds)
		s.NoError(err, "the agent's seed is in its creds")
	})

	s.Run("operator mode: a JWT that is not one is the error", func() {
		defer agent.SetJoinNATSFns(id, s.reply(pki.JoinReply{
			ControllerPublicKey: make([]byte, 32),
			NATS:                pki.JoinNATS{JWT: "not-a-jwt"},
		}, "osapi.join.request"))()

		_, err := agent.ExportJoinNATS(context.Background(), slog.Default(), s.fs, s.cfg)

		s.ErrorContains(err, "the JWT of the join")
	})

	s.Run("operator mode: creds that cannot be kept are the error", func() {
		s.Require().NoError(s.fs.MkdirAll("/pki/nats.creds/x", 0o700))
		acct, _ := nkeys.CreateAccount()
		defer agent.SetJoinNATSFns(
			id,
			func(_ string, _ []nats.Option, _ string, data []byte) ([]byte, error) {
				var req pki.JoinRequest
				s.Require().NoError(json.Unmarshal(data, &req))
				agentJWT, _ := jwt.NewUserClaims(req.NATSUserKey).Encode(acct)

				return json.Marshal(pki.JoinReply{
					ControllerPublicKey: make([]byte, 32),
					NATS:                pki.JoinNATS{JWT: agentJWT},
				})
			},
		)()

		_, err := agent.ExportJoinNATS(context.Background(), slog.Default(), s.fs, s.cfg)

		s.ErrorContains(err, "join over NATS: ")
		s.NotContains(err.Error(), "the JWT of the join")
	})

	s.Run("with auth callout the agent keeps its own user key and connects with it", func() {
		defer agent.SetJoinNATSFns(id, func(string, []nats.Option, string, []byte) ([]byte, error) {
			return json.Marshal(pki.JoinReply{
				ControllerPublicKey: make([]byte, 32),
				NATS:                pki.JoinNATS{Callout: true},
			})
		})()

		got, err := agent.ExportJoinNATS(context.Background(), slog.Default(), s.fs, s.cfg)

		s.Require().NoError(err)
		s.Equal("nkey", got.Agent.NATS.Auth.Type)
		s.Equal("/pki/nats-user.nk", got.Agent.NATS.Auth.NKeyFile)
		info, err := s.fs.Stat("/pki/nats-user.nk")
		s.Require().NoError(err)
		s.Equal("-rw-------", info.Mode().Perm().String(), "the seed is 0600")
	})

	s.Run("with auth callout a user key that cannot be kept is the error", func() {
		s.Require().NoError(s.fs.MkdirAll("/pki/nats-user.nk/x", 0o700))
		defer agent.SetJoinNATSFns(id, func(string, []nats.Option, string, []byte) ([]byte, error) {
			return json.Marshal(pki.JoinReply{
				ControllerPublicKey: make([]byte, 32),
				NATS:                pki.JoinNATS{Callout: true},
			})
		})()

		_, err := agent.ExportJoinNATS(context.Background(), slog.Default(), s.fs, s.cfg)

		s.Error(err)
	})

	s.Run("a user nkey that cannot be made is the error", func() {
		defer agent.SetNewNATSUser(func() ([]byte, string, error) {
			return nil, "", errors.New("no entropy")
		})()
		defer agent.SetJoinNATSFns(id, s.reply(admitted, "osapi.join.request"))()

		_, err := agent.ExportJoinNATS(context.Background(), slog.Default(), s.fs, s.cfg)

		s.ErrorContains(err, "no entropy")
	})

	s.Run("the agent's own labels win over the token's", func() {
		s.cfg.Agent.Labels = map[string]string{"group": "db"}
		defer agent.SetJoinNATSFns(id, s.reply(admitted, "osapi.join.request"))()

		got, err := agent.ExportJoinNATS(context.Background(), slog.Default(), s.fs, s.cfg)

		s.Require().NoError(err)
		s.Equal(map[string]string{"group": "db", "dc": "east"}, got.Agent.Labels)
	})

	s.Run("a saved join that is not JSON joins again", func() {
		s.Require().NoError(avfs.WriteFile(s.fs, "/pki/"+agent.NATSJoinFile, []byte("{"), 0o600))
		defer agent.SetJoinNATSFns(id, s.reply(admitted, "osapi.join.request"))()

		got, err := agent.ExportJoinNATS(context.Background(), slog.Default(), s.fs, s.cfg)

		s.Require().NoError(err)
		s.Equal("agent", got.Agent.NATS.Auth.Username)
	})

	s.Run("a join reaches a real NATS server", func() {
		srv := natstest.RunRandClientPortServer()
		defer srv.Shutdown()
		nc, err := nats.Connect(srv.ClientURL())
		s.Require().NoError(err)
		defer nc.Close()
		_, err = nc.Subscribe("osapi.join.request", func(m *nats.Msg) {
			data, _ := json.Marshal(admitted)
			_ = m.Respond(data)
		})
		s.Require().NoError(err)
		s.Require().NoError(nc.Flush())
		s.cfg.Agent.Join = srv.ClientURL() + "?token=t"
		defer agent.SetJoinNATSFns(id, nil)()

		got, err := agent.ExportJoinNATS(context.Background(), slog.Default(), s.fs, s.cfg)

		s.Require().NoError(err)
		s.Equal("agent", got.Agent.NATS.Auth.Username)
	})

	s.Run("a server where no controller answers is the error", func() {
		srv := natstest.RunRandClientPortServer()
		defer srv.Shutdown()
		s.cfg.Agent.Join = srv.ClientURL() + "?token=t"
		defer agent.SetJoinNATSFns(id, nil)()

		_, err := agent.ExportJoinNATS(context.Background(), slog.Default(), s.fs, s.cfg)

		s.ErrorContains(err, "no responders")
	})

	s.Run("a server that is not there is the error", func() {
		s.cfg.Agent.Join = "nats://127.0.0.1:1?token=t"
		defer agent.SetJoinNATSFns(id, nil)()

		_, err := agent.ExportJoinNATS(context.Background(), slog.Default(), s.fs, s.cfg)

		s.ErrorContains(err, "nats://127.0.0.1:1: ")
	})

	s.Run("a certificate that cannot be kept is the error", func() {
		s.cfg.Agent.Join = "tls://hub:4222?token=t&fp=SHA256:00"
		saved, _ := json.Marshal(
			map[string]any{"scheme": "tls", "host": "hub", "port": 4222, "ca": "x"},
		)
		s.Require().NoError(avfs.WriteFile(s.fs, "/pki/"+agent.NATSJoinFile, saved, 0o600))
		s.Require().NoError(s.fs.MkdirAll("/pki/nats-join-ca.pem/x", 0o700))

		_, err := agent.ExportJoinNATS(context.Background(), slog.Default(), s.fs, s.cfg)

		s.Error(err)
	})
}

// selfSigned is a self-signed certificate.
// issued is a CA and a leaf for name that it signed.
func issued(
	t *testing.T,
	name string,
) (*x509.Certificate, *x509.Certificate) {
	t.Helper()
	caPub, caPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	caTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(
			2,
		),
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTmpl, caTmpl, caPub, caPriv)
	if err != nil {
		t.Fatal(err)
	}
	ca, _ := x509.ParseCertificate(caDER)
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(
			3,
		),
		NotBefore:   time.Now().Add(-time.Hour),
		NotAfter:    time.Now().Add(time.Hour),
		DNSNames:    []string{name},
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, ca, pub, caPriv)
	if err != nil {
		t.Fatal(err)
	}
	leaf, _ := x509.ParseCertificate(der)

	return ca, leaf
}

func selfSigned(
	t *testing.T,
) *x509.Certificate {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(1), NotAfter: time.Now().Add(time.Hour)}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, pub, priv)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}

	return cert
}

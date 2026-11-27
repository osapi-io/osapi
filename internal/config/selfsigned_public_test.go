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
package config_test

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/config"
)

type SelfSignedPublicTestSuite struct {
	suite.Suite
}

func TestSelfSignedPublicTestSuite(t *testing.T) {
	suite.Run(t, new(SelfSignedPublicTestSuite))
}

func (s *SelfSignedPublicTestSuite) TearDownTest() {
	config.SelfSignedHostname = os.Hostname
	config.SelfSignedWrite = os.WriteFile
}

// cert is the certificate in file.
func (s *SelfSignedPublicTestSuite) cert(
	file string,
) *x509.Certificate {
	data, err := os.ReadFile(file)
	s.Require().NoError(err)
	block, _ := pem.Decode(data)
	c, err := x509.ParseCertificate(block.Bytes)
	s.Require().NoError(err)

	return c
}

// TestWithCurrentSelfSigned holds that a metrics certificate follows a rename
// of the host, and stays the same while the name does (osapi-io/osapi#1493).
func (s *SelfSignedPublicTestSuite) TestCertSources() {
	var c config.Config
	c.Controller.PKI.KeyDir = "/ctl"
	c.Controller.API.TLS = config.ServerTLS{CertFile: "/op/api.crt", KeyFile: "/op/api.key"}
	c.NATS.Server.KeyDir = "/nats"
	c.Agent.PKI.KeyDir = "/agent"

	got := config.CertSources(c)
	s.Require().Len(got, 5)
	s.Equal("api", got[0].Name)
	s.Equal("/op/api.crt", got[0].CertFile, "a cert_file is the operator's")
	s.False(got[0].SelfSigned)
	metricsCert, _ := config.SelfSignedFiles("/ctl", "controller-metrics")
	s.Equal(metricsCert, got[1].CertFile, "otherwise the self-signed one in the key directory")
	s.True(got[1].SelfSigned)

	agent := config.CertSourcesOf(c, "agent")
	s.Require().Len(agent, 1)
	s.Equal("agent-metrics", agent[0].Name)
	s.Len(config.CertSourcesOf(c, "controller", "nats"), 4)
}

func (s *SelfSignedPublicTestSuite) TestWithCurrentSelfSigned() {
	s.Run("a renamed host gets a certificate for its new name", func() {
		dir := s.T().TempDir()
		config.SelfSignedHostname = func() (string, error) { return "ubuntu-1", nil }
		first, err := config.ServerTLS{Enabled: true}.WithCurrentSelfSigned(dir, "agent-metrics")
		s.Require().NoError(err)
		before := s.cert(first.CertFile)

		again, err := config.ServerTLS{Enabled: true}.WithCurrentSelfSigned(dir, "agent-metrics")
		s.Require().NoError(err)
		s.Equal(before.Raw, s.cert(again.CertFile).Raw, "the same name keeps the certificate")

		config.SelfSignedHostname = func() (string, error) { return "web-01", nil }
		renamed, err := config.ServerTLS{Enabled: true}.WithCurrentSelfSigned(dir, "agent-metrics")
		s.Require().NoError(err)
		after := s.cert(renamed.CertFile)
		s.NotEqual(before.Raw, after.Raw)
		s.Contains(after.DNSNames, "web-01")
		s.NotContains(after.DNSNames, "ubuntu-1")
	})
	s.Run("a file that is not a certificate is made again", func() {
		dir := s.T().TempDir()
		config.SelfSignedHostname = func() (string, error) { return "web-01", nil }
		crt, _ := config.SelfSignedFiles(dir, "agent-metrics")
		s.Require().NoError(os.WriteFile(crt, []byte("nope"), 0o644))
		got, err := config.ServerTLS{Enabled: true}.WithCurrentSelfSigned(dir, "agent-metrics")
		s.Require().NoError(err)
		s.Contains(s.cert(got.CertFile).DNSNames, "web-01")
	})
	s.Run("a cert_file of the operator and TLS off are as they are", func() {
		in := config.ServerTLS{Enabled: true, CertFile: "/etc/x.crt", KeyFile: "/etc/x.key"}
		got, err := in.WithCurrentSelfSigned(s.T().TempDir(), "agent-metrics")
		s.Require().NoError(err)
		s.Equal(in, got)
		got, err = config.ServerTLS{}.WithCurrentSelfSigned(s.T().TempDir(), "agent-metrics")
		s.Require().NoError(err)
		s.Equal(config.ServerTLS{}, got)
	})
	s.Run("PEM that is not a certificate is made again", func() {
		dir := s.T().TempDir()
		config.SelfSignedHostname = func() (string, error) { return "web-01", nil }
		crt, _ := config.SelfSignedFiles(dir, "agent-metrics")
		s.Require().NoError(os.WriteFile(crt,
			[]byte("-----BEGIN CERTIFICATE-----\nAAAA\n-----END CERTIFICATE-----\n"), 0o644))
		got, err := config.ServerTLS{Enabled: true}.WithCurrentSelfSigned(dir, "agent-metrics")
		s.Require().NoError(err)
		s.Contains(s.cert(got.CertFile).DNSNames, "web-01")
	})
}

func (s *SelfSignedPublicTestSuite) TestMakeSelfSigned() {
	dir := s.T().TempDir()
	cert, key := filepath.Join(dir, "nats.next.crt"), filepath.Join(dir, "nats.next.key")
	s.Require().NoError(config.MakeSelfSigned(cert, key))
	pair, err := tls.LoadX509KeyPair(cert, key)
	s.Require().NoError(err, "the certificate and key are a pair")
	leaf, err := x509.ParseCertificate(pair.Certificate[0])
	s.Require().NoError(err)
	s.Contains(leaf.DNSNames, "localhost")

	s.Error(config.MakeSelfSigned(filepath.Join(cert, "x.crt"), filepath.Join(cert, "x.key")),
		"a directory that is a file")
}

func (s *SelfSignedPublicTestSuite) TestWithSelfSigned() {
	s.Run("TLS off is as it is", func() {
		got, err := config.ServerTLS{}.WithSelfSigned(s.T().TempDir(), "nats")
		s.Require().NoError(err)
		s.Equal(config.ServerTLS{}, got)
	})
	s.Run("a cert_file of the operator is kept", func() {
		in := config.ServerTLS{Enabled: true, CertFile: "/etc/x.crt", KeyFile: "/etc/x.key"}
		got, err := in.WithSelfSigned(s.T().TempDir(), "nats")
		s.Require().NoError(err)
		s.Equal(in, got)
	})
	s.Run("TLS on with no cert_file makes the certificate, for the host and each URL", func() {
		config.SelfSignedHostname = func() (string, error) { return "web-01", nil }
		dir := s.T().TempDir()
		got, err := config.ServerTLS{Enabled: true}.WithSelfSigned(
			dir, "nats", "tls://nats.example.com:4222", "https://10.0.0.5", "::bad",
		)
		s.Require().NoError(err)
		s.Equal(filepath.Join(dir, "nats.crt"), got.CertFile)
		s.Equal(filepath.Join(dir, "nats.key"), got.KeyFile)
		c := s.cert(got.CertFile)
		s.ElementsMatch([]string{"localhost", "web-01", "nats.example.com"}, c.DNSNames)
		s.Len(c.IPAddresses, 3, "127.0.0.1, ::1 and 10.0.0.5")
		info, err := os.Stat(got.KeyFile)
		s.Require().NoError(err)
		s.Equal(os.FileMode(0o600), info.Mode().Perm())

		// A second start serves the same certificate.
		again, err := config.ServerTLS{Enabled: true}.WithSelfSigned(dir, "nats")
		s.Require().NoError(err)
		s.Equal(c.Raw, s.cert(again.CertFile).Raw)
	})
	s.Run("a host with no name has the loopback names only", func() {
		config.SelfSignedHostname = func() (string, error) { return "", errors.New("no name") }
		got, err := config.ServerTLS{Enabled: true}.WithSelfSigned(s.T().TempDir(), "api")
		s.Require().NoError(err)
		s.Equal([]string{"localhost"}, s.cert(got.CertFile).DNSNames)
	})
	s.Run("a key directory that cannot be made is the error", func() {
		blocked := filepath.Join(s.T().TempDir(), "file")
		s.Require().NoError(os.WriteFile(blocked, nil, 0o600))
		_, err := config.ServerTLS{
			Enabled: true,
		}.WithSelfSigned(
			filepath.Join(blocked, "pki"),
			"api",
		)
		s.ErrorContains(err, "make the key directory")
	})
	s.Run("a key that cannot be written is the error", func() {
		config.SelfSignedWrite = func(string, []byte, os.FileMode) error { return errors.New("disk full") }
		_, err := config.ServerTLS{Enabled: true}.WithSelfSigned(s.T().TempDir(), "api")
		s.ErrorContains(err, "write the key of")
	})
	s.Run("a certificate that cannot be written is the error", func() {
		config.SelfSignedWrite = func(name string, data []byte, perm os.FileMode) error {
			if filepath.Ext(name) == ".crt" {
				return errors.New("disk full")
			}

			return os.WriteFile(name, data, perm)
		}
		_, err := config.ServerTLS{Enabled: true}.WithSelfSigned(s.T().TempDir(), "api")
		s.ErrorContains(err, "api.crt: disk full")
	})
}

// TestTrustLocalOperator: with osapi's NATS server in operator mode and no
// resolver_file, a client on this host with no auth of its own takes its
// role's credentials, and the controller the join's signing keys.
func (s *SelfSignedPublicTestSuite) TestTrustLocalOperator() {
	base := func() config.Config {
		var c config.Config
		c.NATS.Server.KeyDir = "/etc/osapi/pki"
		c.NATS.Server.Auth = config.NATSServerAuth{Type: "operator"}
		c.Controller.NATS = config.NATSConnection{
			Host: "localhost",
			Auth: config.NATSAuth{Type: "none"},
		}
		c.Agent.NATS = config.NATSConnection{Host: "127.0.0.1"}

		return c
	}
	s.Run("each local client takes its role's credentials", func() {
		got := config.TrustLocal(base())
		s.Equal(
			config.NATSAuth{Type: "creds", CredsFile: "/etc/osapi/pki/nats-auth/controller.creds"},
			got.Controller.NATS.Auth,
		)
		s.Equal(config.NATSAuth{Type: "creds", CredsFile: "/etc/osapi/pki/nats-auth/agent.creds"},
			got.Agent.NATS.Auth)
		s.True(got.Controller.Join.Scoped)
		s.Equal("/etc/osapi/pki/nats-auth/agent-signing.nk", got.Controller.Join.SigningKeyFile)
		s.Equal("/etc/osapi/pki/nats-auth/join-signing.nk", got.Controller.Join.JoinSigningKeyFile)
	})
	s.Run("an auth, a host elsewhere, signing keys or a callout of the config are kept", func() {
		c := base()
		c.Controller.NATS.Auth = config.NATSAuth{Type: "user_pass", Username: "ctl"}
		c.Agent.NATS.Host = "nats.example.com"
		c.Controller.Join.Callout.IssuerKeyFile = "/etc/osapi/issuer.nk"
		got := config.TrustLocal(c)
		s.Equal("user_pass", got.Controller.NATS.Auth.Type)
		s.Empty(got.Agent.NATS.Auth.Type)
		s.Empty(got.Controller.Join.SigningKeyFile)
	})
	s.Run("an agent with a join string is left to its join", func() {
		c := base()
		c.Agent.Join = "tls://nats.example.com:4222?token=t&fp=SHA256:aa"
		c.Agent.NATS.TLS.Enabled = true
		got := config.TrustLocal(c)
		s.Empty(got.Agent.NATS.Auth.Type, "no credentials of this host")
		s.Empty(got.Agent.NATS.TLS.CAFile, "no nats.crt of this host: the join says what it trusts")
		s.Equal("/etc/osapi/pki/nats-auth/controller.creds", got.Controller.NATS.Auth.CredsFile)
	})
	s.Run("a server with its own resolver_file is not self-made", func() {
		c := base()
		c.NATS.Server.Auth.ResolverFile = "/etc/nsc/resolver.conf"
		got := config.TrustLocal(c)
		s.Equal("none", got.Controller.NATS.Auth.Type)
	})
}

func (s *SelfSignedPublicTestSuite) TestTrustLocal() {
	base := func() config.Config {
		var c config.Config
		c.Controller.PKI.KeyDir = "/etc/osapi/pki"
		c.NATS.Server.KeyDir = "/var/osapi/pki"
		c.Controller.Client.URL = "https://localhost:8080"
		c.Controller.NATS = config.NATSConnection{
			Host: "localhost",
			TLS:  config.ClientTLS{Enabled: true},
		}
		c.Agent.NATS = config.NATSConnection{
			Host: "127.0.0.1",
			TLS:  config.ClientTLS{Enabled: true},
		}

		return c
	}
	tests := []struct {
		name   string
		change func(*config.Config)
		client string
		nats   string
		agent  string
	}{
		{
			name:   "each client on this host trusts the self-signed certificate",
			client: "/etc/osapi/pki/api.crt",
			nats:   "/var/osapi/pki/nats.crt",
			agent:  "/var/osapi/pki/nats.crt",
		},
		{name: "a server on another host needs ca_file", change: func(c *config.Config) {
			c.Controller.Client.URL = "https://controller.example.com:8080"
			c.Controller.NATS.Host = "nats.example.com"
			c.Agent.NATS.Host = "::1"
		}, agent: "/var/osapi/pki/nats.crt"},
		{name: "plain HTTP and NATS without TLS trust nothing", change: func(c *config.Config) {
			c.Controller.Client.URL = "http://localhost:8080"
			c.Controller.NATS.TLS.Enabled = false
			c.Agent.NATS.TLS.Enabled = false
		}},
		{name: "a ca_file that is set is kept", change: func(c *config.Config) {
			c.Controller.Client.TLS.CAFile = "/mine/ca.pem"
			c.Agent.NATS.TLS.CAFile = "/mine/ca.pem"
		}, client: "/mine/ca.pem", nats: "/var/osapi/pki/nats.crt", agent: "/mine/ca.pem"},
		{
			name: "a server with its own certificate is not self-signed",
			change: func(c *config.Config) {
				c.Controller.API.TLS.CertFile = "/etc/tls/api.crt"
				c.NATS.Server.TLS.CertFile = "/etc/tls/nats.crt"
			},
		},
		{name: "a URL that does not parse trusts nothing", change: func(c *config.Config) {
			c.Controller.Client.URL = "://"
		}, nats: "/var/osapi/pki/nats.crt", agent: "/var/osapi/pki/nats.crt"},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			c := base()
			if tt.change != nil {
				tt.change(&c)
			}
			got := config.TrustLocal(c)
			s.Equal(tt.client, got.Controller.Client.TLS.CAFile)
			s.Equal(tt.nats, got.Controller.NATS.TLS.CAFile)
			s.Equal(tt.agent, got.Agent.NATS.TLS.CAFile)
		})
	}
}

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
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/config"
	"github.com/osapi-io/osapi/internal/tlspolicy"

	"github.com/avfs/avfs"
)

type TLSPublicTestSuite struct {
	suite.Suite
}

func (s *TLSPublicTestSuite) TestClientConfig() {
	cert, key := writeCert(s.T())
	notPEM := filepath.Join(s.T().TempDir(), "not.pem")
	s.Require().NoError(avfs.WriteFile(testFS, notPEM, []byte("not a certificate"), 0o600))

	tests := []struct {
		name         string
		c            config.ClientTLS
		noSystem     bool
		validateFunc func(*tls.Config, error)
	}{
		{
			name: "when nothing is set, it trusts the system roots",
			validateFunc: func(got *tls.Config, err error) {
				s.Require().NoError(err)
				s.Nil(got.RootCAs)
				s.Equal(uint16(tls.VersionTLS12), got.MinVersion)
			},
		},
		{
			name: "when it has a CA, a client certificate and a server name, it uses each",
			c: config.ClientTLS{
				CAFile:     cert,
				CertFile:   cert,
				KeyFile:    key,
				ServerName: "api.lab",
			},
			validateFunc: func(got *tls.Config, err error) {
				s.Require().NoError(err)
				s.NotNil(got.RootCAs)
				s.Len(got.Certificates, 1)
				s.Equal("api.lab", got.ServerName)
			},
		},
		{
			name:     "when the system roots cannot be read, it trusts the CA alone",
			c:        config.ClientTLS{CAFile: cert},
			noSystem: true,
			validateFunc: func(got *tls.Config, err error) {
				s.Require().NoError(err)
				pemData, err := avfs.ReadFile(testFS, cert)
				s.Require().NoError(err)
				want := x509.NewCertPool()
				s.Require().True(want.AppendCertsFromPEM(pemData))
				s.True(want.Equal(got.RootCAs))
			},
		},
		{
			name: "when it skips the check, the config says so",
			c:    config.ClientTLS{InsecureSkipVerify: true},
			validateFunc: func(got *tls.Config, err error) {
				s.Require().NoError(err)
				s.True(got.InsecureSkipVerify)
			},
		},
		{
			name: "when the CA cannot be read, it is an error",
			c:    config.ClientTLS{CAFile: "/no/such/ca.crt"},
			validateFunc: func(_ *tls.Config, err error) {
				s.ErrorContains(err, "read ca_file")
			},
		},
		{
			name: "when the CA has no certificate, it is an error",
			c:    config.ClientTLS{CAFile: notPEM},
			validateFunc: func(_ *tls.Config, err error) {
				s.ErrorContains(err, "no PEM certificate")
			},
		},
		{
			name: "when the client certificate does not load, it is an error",
			c:    config.ClientTLS{CertFile: notPEM, KeyFile: notPEM},
			validateFunc: func(_ *tls.Config, err error) {
				s.ErrorContains(err, "load the client certificate")
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.noSystem {
				config.SetSystemCertPoolFn(func() (*x509.CertPool, error) {
					return nil, errors.New("no system roots")
				})
				defer config.ResetSystemCertPoolFn()
			}
			tc.validateFunc(tc.c.Config())
		})
	}
}

func (s *TLSPublicTestSuite) TestWarn() {
	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, nil))

	config.ClientTLS{}.Warn(log, "controller.client.tls")
	s.Empty(buf.String(), "a check is no warning")

	config.ClientTLS{InsecureSkipVerify: true}.Warn(log, "controller.client.tls")
	s.Contains(buf.String(), "controller.client.tls.insecure_skip_verify")
}

// handshake runs a TLS handshake between server and client over loopback
// TCP, and returns the client's error.
func handshake(
	server *tls.Config,
	client *tls.Config,
) error {
	ln, err := tls.Listen("tcp", "127.0.0.1:0", server)
	if err != nil {
		return err
	}
	defer func() { _ = ln.Close() }()
	done := make(chan struct{})
	go func() {
		defer close(done)
		c, err := ln.Accept()
		if err != nil {
			return
		}
		_ = c.SetDeadline(time.Now().Add(5 * time.Second))
		_ = c.(*tls.Conn).Handshake()
		_ = c.Close()
	}()
	d := &net.Dialer{Timeout: 5 * time.Second}
	c, err := tls.DialWithDialer(d, "tcp", ln.Addr().String(), client)
	if err == nil {
		_ = c.Close()
	}
	<-done

	return err
}

func (s *TLSPublicTestSuite) SetupTest() {
	config.CertCheckInterval = 0
}

func (s *TLSPublicTestSuite) TearDownTest() {
	config.CertCheckInterval = time.Second
}

// replace writes the certificate and key of from over to, a second later,
// so that a reader that compares modification times sees the change.
func (s *TLSPublicTestSuite) replace(
	fromCert string,
	fromKey string,
	toCert string,
	toKey string,
) {
	for _, f := range [][2]string{{fromCert, toCert}, {fromKey, toKey}} {
		data, err := avfs.ReadFile(testFS, f[0])
		s.Require().NoError(err)
		s.Require().NoError(avfs.WriteFile(testFS, f[1], data, 0o600))
		later := time.Now().Add(time.Second)
		s.Require().NoError(os.Chtimes(f[1], later, later))
	}
}

// leaf is the leaf of the certificate that server serves now.
func (s *TLSPublicTestSuite) leaf(
	server *tls.Config,
) []byte {
	got, err := server.GetCertificate(&tls.ClientHelloInfo{})
	s.Require().NoError(err)

	return got.Certificate[0]
}

func (s *TLSPublicTestSuite) TestServerConfig() {
	cert, key := writeCert(s.T())
	other, otherKey := writeCert(s.T())

	got, err := config.ServerTLS{}.Config()
	s.Require().NoError(err)
	s.Nil(got, "off serves plain HTTP")

	_, err = config.ServerTLS{Enabled: true}.Config()
	s.ErrorContains(err, "tls.enabled needs cert_file and key_file")

	_, err = config.ServerTLS{Enabled: true, CertFile: key, KeyFile: key}.Config()
	s.ErrorContains(err, "load the server certificate")

	got, err = config.ServerTLS{Enabled: true, CertFile: cert, KeyFile: key}.Config()
	s.Require().NoError(err)
	first := s.leaf(got)
	s.Equal(first, s.leaf(got), "unchanged files serve the same certificate")

	s.replace(other, otherKey, cert, key)
	second := s.leaf(got)
	s.NotEqual(first, second, "a replaced certificate is served at the next handshake")

	s.Require().NoError(avfs.WriteFile(testFS, cert, []byte("not a certificate"), 0o600))
	later := time.Now().Add(2 * time.Second)
	s.Require().NoError(os.Chtimes(cert, later, later))
	s.Equal(second, s.leaf(got), "a pair that does not load keeps the last one")

	s.Require().NoError(testFS.Remove(cert))
	s.Equal(second, s.leaf(got), "a missing file keeps the last one")

	// A copy that keeps the time of the file, such as cp -p, is read too:
	// the reloader compares the content.
	again, err := config.ServerTLS{Enabled: true, CertFile: other, KeyFile: otherKey}.Config()
	s.Require().NoError(err)
	before := s.leaf(again)
	third, thirdKey := writeCert(s.T())
	for _, f := range [][2]string{{third, other}, {thirdKey, otherKey}} {
		fi, err := testFS.Stat(f[1])
		s.Require().NoError(err)
		data, err := avfs.ReadFile(testFS, f[0])
		s.Require().NoError(err)
		s.Require().NoError(avfs.WriteFile(testFS, f[1], data, 0o600))
		s.Require().NoError(os.Chtimes(f[1], fi.ModTime(), fi.ModTime()))
	}
	s.NotEqual(before, s.leaf(again), "a copy with the same time is served")

	// Within the interval, a handshake reads no file.
	config.CertCheckInterval = time.Hour
	held, err := config.ServerTLS{Enabled: true, CertFile: other, KeyFile: otherKey}.Config()
	s.Require().NoError(err)
	heldLeaf := s.leaf(held)
	fourth, fourthKey := writeCert(s.T())
	s.replace(fourth, fourthKey, other, otherKey)
	s.Equal(heldLeaf, s.leaf(held), "a change within the interval waits for the next check")
	s.Equal(uint16(tls.VersionTLS12), got.MinVersion, "it serves what Negotiation says")
}

func (s *TLSPublicTestSuite) TestNegotiation() {
	s.Nil(config.ServerTLS{}.Negotiation(), "when TLS is off, it negotiates nothing")

	got := config.ServerTLS{Enabled: true}.Negotiation()
	s.True(
		tlspolicy.Approved(got, true),
		"when TLS is on, it negotiates only approved suites and groups",
	)
}

func (s *TLSPublicTestSuite) TestPolicy() {
	got, err := config.ClientTLS{Enabled: true}.Config()
	s.Require().NoError(err)
	s.True(tlspolicy.Approved(got, false), "a client offers only approved suites and groups")
	s.True(
		tlspolicy.Approved(config.SelfClient(&tls.Config{}), false),
		"so does a server's client of itself",
	)
}

func (s *TLSPublicTestSuite) TestDial() {
	got, err := config.ClientTLS{CAFile: "/no/such/ca.crt"}.Dial("hub.lab")
	s.Require().NoError(err, "off reads nothing")
	s.Nil(got)

	got, err = config.ClientTLS{Enabled: true, ServerName: "hub.lab"}.Dial("10.0.0.1")
	s.Require().NoError(err)
	s.Equal("hub.lab", got.ServerName)
	s.Nil(got.VerifyConnection, "with no ca_file, crypto/tls checks the system roots")

	_, err = config.ClientTLS{Enabled: true, CAFile: "/no/such/ca.crt"}.Dial("hub.lab")
	s.ErrorContains(err, "read ca_file")

	got, err = config.ClientTLS{Enabled: true, CAFile: "/no/such/ca.crt", InsecureSkipVerify: true}.
		Dial("hub.lab")
	s.ErrorContains(err, "read ca_file", "the file must load once, also with no check")
	s.Nil(got)

	cert, key := writeCertFor(s.T(), []string{"localhost"}, nil)
	next, nextKey := writeCertFor(s.T(), []string{"localhost"}, nil)
	server, err := config.ServerTLS{Enabled: true, CertFile: cert, KeyFile: key}.Config()
	s.Require().NoError(err)
	ca := filepath.Join(s.T().TempDir(), "ca.crt")
	current, err := avfs.ReadFile(testFS, cert)
	s.Require().NoError(err)
	s.Require().NoError(avfs.WriteFile(testFS, ca, current, 0o600))

	client, err := config.ClientTLS{Enabled: true, CAFile: ca}.Dial("localhost")
	s.Require().NoError(err)
	s.NoError(handshake(server, client), "the server's certificate is in ca_file")

	wrongName, err := config.ClientTLS{Enabled: true, CAFile: ca}.Dial("hub.lab")
	s.Require().NoError(err)
	s.Error(handshake(server, wrongName), "the certificate does not name the host")

	s.replace(next, nextKey, cert, key)
	s.Error(handshake(server, client), "ca_file does not have the new certificate yet")

	nextPEM, err := avfs.ReadFile(testFS, next)
	s.Require().NoError(err)
	s.Require().NoError(avfs.WriteFile(testFS, ca, append(current, nextPEM...), 0o600))
	s.NoError(handshake(server, client), "the same config trusts what ca_file gained")
	block, _ := pem.Decode(nextPEM)
	leaf, err := x509.ParseCertificate(block.Bytes)
	s.Require().NoError(err)
	s.NoError(client.VerifyConnection(tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{leaf, leaf},
	}), "a chain with an intermediate checks it too")

	s.Require().NoError(avfs.WriteFile(testFS, ca, []byte("not a certificate"), 0o600))
	s.ErrorContains(handshake(server, client), "no PEM certificate")

	s.Require().NoError(testFS.Remove(ca))
	s.ErrorContains(handshake(server, client), "read ca_file")

	s.ErrorContains(client.VerifyConnection(tls.ConnectionState{}), "no certificate")

	skip, err := config.ClientTLS{Enabled: true, CAFile: cert, InsecureSkipVerify: true}.
		Dial("localhost")
	s.Require().NoError(err)
	s.Nil(skip.VerifyConnection, "insecure_skip_verify checks nothing")
}

func (s *TLSPublicTestSuite) TestSelfClient() {
	s.Nil(config.SelfClient(nil), "no TLS is plain text")

	cert, key := writeCertFor(s.T(), []string{"hub.lab"}, nil)
	next, nextKey := writeCertFor(s.T(), nil, []net.IP{net.IPv4(10, 0, 0, 1)})
	server, err := config.ServerTLS{Enabled: true, CertFile: cert, KeyFile: key}.Config()
	s.Require().NoError(err)
	client := config.SelfClient(server)
	s.NoError(handshake(server, client), "it trusts the server's own certificate")

	s.replace(next, nextKey, cert, key)
	s.NoError(handshake(server, client), "and the one that the server loaded again")

	stranger, strangerKey := writeCert(s.T())
	other, err := config.ServerTLS{Enabled: true, CertFile: stranger, KeyFile: strangerKey}.Config()
	s.Require().NoError(err)
	s.ErrorContains(handshake(other, client), "not its own")
	s.ErrorContains(client.VerifyConnection(tls.ConnectionState{}), "not its own")

	failing := &tls.Config{
		GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
			return nil, os.ErrNotExist
		},
	}
	s.ErrorIs(config.SelfClient(failing).VerifyConnection(tls.ConnectionState{}), os.ErrNotExist)
}

func (s *TLSPublicTestSuite) TestAddress() {
	m := config.MetricsServer{Host: "0.0.0.0", Port: 9091}
	s.Equal("http://0.0.0.0:9091", m.Address())

	m.TLS.Enabled = true
	s.Equal("https://0.0.0.0:9091", m.Address())
}

func TestTLSPublicTestSuite(t *testing.T) {
	suite.Run(t, new(TLSPublicTestSuite))
}

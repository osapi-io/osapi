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
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi/internal/config"
)

type TLSPublicTestSuite struct {
	suite.Suite
}

func (s *TLSPublicTestSuite) TestClientConfig() {
	cert, key := writeCert(s.T())
	notPEM := filepath.Join(s.T().TempDir(), "not.pem")
	s.Require().NoError(os.WriteFile(notPEM, []byte("not a certificate"), 0o600))

	tests := []struct {
		name         string
		c            config.ClientTLS
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

func (s *TLSPublicTestSuite) TestServerConfig() {
	cert, key := writeCert(s.T())

	got, err := config.ServerTLS{}.Config()
	s.Require().NoError(err)
	s.Nil(got, "off serves plain HTTP")

	_, err = config.ServerTLS{Enabled: true}.Config()
	s.ErrorContains(err, "tls.enabled needs cert_file and key_file")

	_, err = config.ServerTLS{Enabled: true, CertFile: key, KeyFile: key}.Config()
	s.ErrorContains(err, "load the server certificate")

	got, err = config.ServerTLS{Enabled: true, CertFile: cert, KeyFile: key}.Config()
	s.Require().NoError(err)
	s.Len(got.Certificates, 1)
}

func (s *TLSPublicTestSuite) TestDial() {
	got, err := config.ClientTLS{CAFile: "/no/such/ca.crt"}.Dial()
	s.Require().NoError(err, "off reads nothing")
	s.Nil(got)

	got, err = config.ClientTLS{Enabled: true, ServerName: "hub.lab"}.Dial()
	s.Require().NoError(err)
	s.Equal("hub.lab", got.ServerName)
}

func (s *TLSPublicTestSuite) TestSelfClient() {
	s.Nil(config.SelfClient(nil), "no TLS is plain text")

	for _, tc := range []struct {
		name  string
		names []string
		ips   []net.IP
		want  string
	}{
		{name: "the first DNS name", names: []string{"hub.lab", "localhost"}, want: "hub.lab"},
		{name: "an IP when it has no DNS name", ips: []net.IP{net.IPv4(10, 0, 0, 1)}, want: "10.0.0.1"},
		{name: "localhost when it has neither", want: "localhost"},
	} {
		s.Run(tc.name, func() {
			cert, key := writeCertFor(s.T(), tc.names, tc.ips)
			server, err := config.ServerTLS{Enabled: true, CertFile: cert, KeyFile: key}.Config()
			s.Require().NoError(err)

			got := config.SelfClient(server)
			s.Equal(tc.want, got.ServerName)
			s.NotNil(got.RootCAs)
		})
	}
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
